package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Sink interface {
	ApplyBatch(context.Context, []cdc_domain.ChangeEvent) error
	Name() string
	Close() error
}
type MemorySink struct {
	mu     sync.Mutex
	events []cdc_domain.ChangeEvent
	seen   map[string]struct{}
}

func NewMemorySink() *MemorySink   { return &MemorySink{seen: map[string]struct{}{}} }
func (s *MemorySink) Name() string { return "memory" }
func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) ApplyBatch(ctx context.Context, events []cdc_domain.ChangeEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if _, ok := s.seen[e.IdempotencyKey()]; ok {
			continue
		}
		s.seen[e.IdempotencyKey()] = struct{}{}
		s.events = append(s.events, e)
	}
	return nil
}
func (s *MemorySink) Events() []cdc_domain.ChangeEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]cdc_domain.ChangeEvent(nil), s.events...)
}

type FileSink struct {
	mu   sync.Mutex
	path string
	file *os.File
	seen map[string]struct{}
}

func NewFileSink(path string) (*FileSink, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, err
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return nil, e
	}
	return &FileSink{path: path, file: f, seen: map[string]struct{}{}}, nil
}
func (s *FileSink) Name() string { return "file" }
func (s *FileSink) Close() error { s.mu.Lock(); defer s.mu.Unlock(); return s.file.Close() }
func (s *FileSink) ApplyBatch(ctx context.Context, events []cdc_domain.ChangeEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	enc := json.NewEncoder(s.file)
	for _, e := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		k := e.IdempotencyKey()
		if _, ok := s.seen[k]; ok {
			continue
		}
		if err := enc.Encode(e); err != nil {
			return err
		}
		s.seen[k] = struct{}{}
	}
	return s.file.Sync()
}

type RetryPolicy struct {
	MaxAttempts int
	Initial     time.Duration
	Max         time.Duration
}

func (p RetryPolicy) backoff(n int) time.Duration {
	d := p.Initial
	for i := 1; i < n; i++ {
		d *= 2
		if d >= p.Max {
			return p.Max
		}
	}
	return d
}
func ApplyWithRetry(ctx context.Context, s Sink, events []cdc_domain.ChangeEvent, p RetryPolicy) error {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}
	if p.Initial <= 0 {
		p.Initial = 10 * time.Millisecond
	}
	if p.Max <= 0 {
		p.Max = time.Second
	}
	var err error
	for i := 1; i <= p.MaxAttempts; i++ {
		if err = s.ApplyBatch(ctx, events); err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		if i < p.MaxAttempts {
			t := time.NewTimer(p.backoff(i))
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
			}
		}
	}
	return fmt.Errorf("sink %s failed after %d attempts: %w", s.Name(), p.MaxAttempts, err)
}

type Router struct{ Sinks map[string]Sink }

func (r Router) Deliver(ctx context.Context, route map[string][]cdc_domain.ChangeEvent, p RetryPolicy) map[string]error {
	out := map[string]error{}
	for n, events := range route {
		if s, ok := r.Sinks[n]; ok {
			out[n] = ApplyWithRetry(ctx, s, events, p)
		} else {
			out[n] = fmt.Errorf("sink %s unavailable", n)
		}
	}
	return out
}
