package snapshotter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"sort"
	"sync"
	"time"
)

type Chunk struct {
	Table      string
	Start, End string
	Rows       int
	Checksum   string
	Completed  bool
}
type Plan struct {
	SourceID  string
	Chunks    []Chunk
	CreatedAt time.Time
}
type Provider interface {
	Tables(context.Context, string) ([]cdc_domain.TableSelection, error)
	ReadChunk(context.Context, string, Chunk) ([]map[string]any, error)
}
type Snapshotter struct {
	provider Provider
	workers  int
	mu       sync.Mutex
	done     map[string]Chunk
}

func New(p Provider, workers int) *Snapshotter {
	if workers < 1 {
		workers = 1
	}
	return &Snapshotter{provider: p, workers: workers, done: map[string]Chunk{}}
}
func (s *Snapshotter) Plan(ctx context.Context, source string) (Plan, error) {
	tables, e := s.provider.Tables(ctx, source)
	if e != nil {
		return Plan{}, e
	}
	out := Plan{SourceID: source, CreatedAt: time.Now().UTC()}
	for _, t := range tables {
		out.Chunks = append(out.Chunks, Chunk{Table: t.Schema + "." + t.Table, Start: "0", End: "max"})
	}
	sort.Slice(out.Chunks, func(i, j int) bool { return out.Chunks[i].Table < out.Chunks[j].Table })
	return out, nil
}
func (s *Snapshotter) Run(ctx context.Context, plan Plan, emit func(cdc_domain.ChangeEvent) error) error {
	jobs := make(chan Chunk)
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				rows, err := s.provider.ReadChunk(ctx, plan.SourceID, c)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					continue
				}
				h := sha256.New()
				for _, row := range rows {
					fmt.Fprintf(h, "%v", row)
					e := cdc_domain.NewEvent("snapshot", cdc_domain.Position{Protocol: "snapshot", File: c.Table, Offset: uint64(len(rows))}, c.Table, cdc_domain.OpInsert, nil, row)
					if er := emit(e); er != nil {
						select {
						case errCh <- er:
						default:
						}
						return
					}
				}
				c.Rows = len(rows)
				c.Checksum = fmt.Sprintf("%x", h.Sum(nil))
				c.Completed = true
				s.mu.Lock()
				s.done[c.Table] = c
				s.mu.Unlock()
			}
		}()
	}
	for _, c := range plan.Chunks {
		select {
		case jobs <- c:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
func (s *Snapshotter) Completed() []Chunk {
	s.mu.Lock()
	done := s.done
	s.mu.Unlock()
	o := make([]Chunk, 0, len(done))
	for _, c := range done {
		o = append(o, c)
	}
	return o
}
