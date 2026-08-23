package repository

import (
	"encoding/json"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu  sync.Mutex
	dir string
	reg *cdc_domain.Registry
}
type disk struct {
	Sources     []cdc_domain.Source     `json:"sources"`
	Pipelines   []cdc_domain.Pipeline   `json:"pipelines"`
	Checkpoints []cdc_domain.Checkpoint `json:"checkpoints"`
	DeadLetters []cdc_domain.DeadLetter `json:"dead_letters"`
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, reg: cdc_domain.NewRegistry()}
	if b, e := os.ReadFile(filepath.Join(dir, "state.json")); e == nil {
		var d disk
		if json.Unmarshal(b, &d) == nil {
			for _, x := range d.Sources {
				s.reg.PutSource(x)
			}
			for _, x := range d.Pipelines {
				s.reg.PutPipeline(x)
			}
			for _, x := range d.Checkpoints {
				s.reg.SetCheckpoint(x)
			}
			for _, x := range d.DeadLetters {
				s.reg.AddDeadLetter(x)
			}
		}
	}
	return s, nil
}
func (s *Store) Registry() *cdc_domain.Registry { return s.reg }
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := disk{Sources: s.reg.ListSources(), Pipelines: s.reg.ListPipelines()}
	for _, p := range d.Pipelines {
		if c, e := s.reg.GetCheckpoint(p.ID); e == nil {
			d.Checkpoints = append(d.Checkpoints, c)
		}
		d.DeadLetters = append(d.DeadLetters, s.reg.GetDeadLetters(p.ID)...)
	}
	b, e := json.MarshalIndent(d, "", "  ")
	if e != nil {
		return e
	}
	tmp := filepath.Join(s.dir, "state.json.tmp")
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, filepath.Join(s.dir, "state.json"))
}
