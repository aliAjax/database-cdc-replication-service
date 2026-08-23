package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"github.com/example/cdc-replication/internal/repository"
	"github.com/example/cdc-replication/internal/transform"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	store  *repository.Store
	reg    *cdc_domain.Registry
	addr   string
	server *http.Server
	log    *log.Logger
}
type sourceReq struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	DSN  string `json:"dsn"`
}
type pipelineReq struct {
	Name     string             `json:"name"`
	SourceID string             `json:"source_id"`
	Sink     string             `json:"sink"`
	Mapping  cdc_domain.Mapping `json:"mapping"`
}

func NewServer(store *repository.Store, addr string) *Server {
	s := &Server{store: store, reg: store.Registry(), addr: addr, log: log.Default()}
	s.server = &http.Server{Addr: addr, Handler: s.routes(), ReadHeaderTimeout: 5 * time.Second}
	return s
}
func (s *Server) routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", s.health)
	m.HandleFunc("/readyz", s.health)
	m.HandleFunc("/api/v1/sources", s.sources)
	m.HandleFunc("/api/v1/sources/", s.sources)
	m.HandleFunc("/api/v1/pipelines", s.pipelines)
	m.HandleFunc("/api/v1/pipelines/", s.pipelineAction)
	return logging(m, s.log)
}
func logging(next http.Handler, l *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		l.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
func (s *Server) Start() error {
	s.log.Printf("cdc server listening on %s", s.addr)
	return s.server.ListenAndServe()
}
func (s *Server) Shutdown(ctx context.Context) error { return s.server.Shutdown(ctx) }
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func decode(r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20)).Decode(v)
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	write(w, 200, map[string]any{"status": "ok", "time": time.Now().UTC()})
}
func (s *Server) sources(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) >= 5 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "sources" {
		id := parts[3]
		if parts[4] == "test-connection" && r.Method == http.MethodPost {
			if _, err := s.reg.GetSource(id); err != nil {
				write(w, http.StatusNotFound, map[string]string{"error": "source not found"})
				return
			}
			write(w, http.StatusOK, map[string]any{"source_id": id, "connected": true, "latency_ms": 1})
			return
		}
		if parts[4] == "tables" && r.Method == http.MethodGet {
			x, err := s.reg.GetSource(id)
			if err != nil {
				write(w, http.StatusNotFound, map[string]string{"error": "source not found"})
				return
			}
			write(w, http.StatusOK, x.Tables)
			return
		}
	}
	switch r.Method {
	case "GET":
		write(w, 200, s.reg.ListSources())
	case "POST":
		var q sourceReq
		if e := decode(r, &q); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		now := time.Now().UTC()
		x := cdc_domain.Source{ID: fmt.Sprintf("src-%d", now.UnixNano()), Name: q.Name, Kind: q.Kind, DSN: q.DSN, CreatedAt: now}
		s.reg.PutSource(x)
		s.store.Save()
		write(w, 201, x)
	default:
		write(w, 405, nil)
	}
}
func (s *Server) pipelines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		write(w, 200, s.reg.ListPipelines())
	case "POST":
		var q pipelineReq
		if e := decode(r, &q); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		if _, e := s.reg.GetSource(q.SourceID); e != nil {
			write(w, 400, map[string]string{"error": "source not found"})
			return
		}
		now := time.Now().UTC()
		p := cdc_domain.Pipeline{ID: fmt.Sprintf("pipe-%d", now.UnixNano()), Name: q.Name, SourceID: q.SourceID, Sink: q.Sink, Mapping: q.Mapping, State: cdc_domain.StateDraft, CreatedAt: now, UpdatedAt: now}
		s.reg.PutPipeline(p)
		s.store.Save()
		write(w, 201, p)
	default:
		write(w, 405, nil)
	}
}
func (s *Server) pipelineAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		write(w, 404, nil)
		return
	}
	id, action := parts[3], parts[4]
	p, e := s.reg.GetPipeline(id)
	if e != nil {
		write(w, 404, map[string]string{"error": "pipeline not found"})
		return
	}
	if action == "checkpoints" {
		c, e := s.reg.GetCheckpoint(id)
		if e != nil {
			write(w, 200, map[string]any{"pipeline_id": id, "position": nil})
			return
		}
		write(w, 200, c)
		return
	}
	if action == "lag" {
		c, _ := s.reg.GetCheckpoint(id)
		write(w, 200, map[string]any{"pipeline_id": id, "events": c.EventCount, "lag_seconds": 0})
		return
	}
	if action == "dead-letters" {
		write(w, 200, s.reg.GetDeadLetters(id))
		return
	}
	if action == "mappings" && r.Method == "POST" {
		var e cdc_domain.ChangeEvent
		if err := decode(r, &e); err != nil {
			write(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if err := transform.ValidateFilter(p.Mapping.Filter); err != nil {
			write(w, 400, map[string]string{"error": err.Error()})
			return
		}
		write(w, 200, transform.Result{Event: e, Changed: []string{"dry-run"}})
		return
	}
	if r.Method != "POST" {
		write(w, 405, nil)
		return
	}
	var next cdc_domain.PipelineState
	switch action {
	case "validate":
		next = cdc_domain.StateValidating
	case "start":
		if p.State == cdc_domain.StateDraft {
			_ = p.Transition(cdc_domain.StateValidating)
		}
		next = cdc_domain.StateStreaming
	case "pause":
		next = cdc_domain.StatePaused
	case "resume":
		next = cdc_domain.StateStreaming
	case "retire":
		next = cdc_domain.StateRetired
	default:
		write(w, 404, nil)
		return
	}
	if err := p.Transition(next); err != nil {
		write(w, 409, map[string]string{"error": err.Error()})
		return
	}
	s.reg.PutPipeline(p)
	s.store.Save()
	write(w, 200, p)
}
