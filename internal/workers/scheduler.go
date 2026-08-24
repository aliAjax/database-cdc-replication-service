package workers

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Job struct {
	ID       string
	Run      func(context.Context) error
	Interval time.Duration
	Once     bool
}
type Scheduler struct {
	mu      sync.Mutex
	jobs    map[string]Job
	cancel  context.CancelFunc
	running bool
}

func NewScheduler() *Scheduler { return &Scheduler{jobs: map[string]Job{}} }
func (s *Scheduler) Add(j Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j.ID == "" || j.Run == nil {
		return errors.New("job id and function required")
	}
	if _, ok := s.jobs[j.ID]; ok {
		return errors.New("job already exists")
	}
	if j.Interval <= 0 {
		j.Interval = time.Minute
	}
	s.jobs[j.ID] = j
	return nil
}
func (s *Scheduler) Remove(id string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.jobs, id) }
func (s *Scheduler) Start(parent context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("scheduler running")
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	s.running = true
	jobs := make([]Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	s.mu.Unlock()
	for _, j := range jobs {
		go s.loop(ctx, j)
	}
	return nil
}
func (s *Scheduler) loop(ctx context.Context, j Job) {
	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()
	for {
		if err := j.Run(ctx); err != nil && ctx.Err() != nil {
			return
		}
		if j.Once {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.running = false
}
