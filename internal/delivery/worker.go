package delivery

import (
	"context"
	"errors"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	Captured   atomic.Uint64
	Delivered  atomic.Uint64
	Retried    atomic.Uint64
	Duplicates atomic.Uint64
	Poison     atomic.Uint64
	Failed     atomic.Uint64
}

func (m *Metrics) Snapshot() map[string]uint64 {
	return map[string]uint64{
		"captured":   m.Captured.Load(),
		"delivered":  m.Delivered.Load(),
		"retried":    m.Retried.Load(),
		"duplicates": m.Duplicates.Load(),
		"poison":     m.Poison.Load(),
		"failed":     m.Failed.Load(),
	}
}

type Batch struct {
	Events  []cdc_domain.ChangeEvent
	Created time.Time
	Bytes   int
}

func NewBatch() Batch {
	return Batch{Events: make([]cdc_domain.ChangeEvent, 0), Created: time.Now().UTC()}
}

func (b *Batch) Add(e cdc_domain.ChangeEvent) {
	b.Events = append(b.Events, e)
	b.Bytes += len(e.ID) + len(e.Table) + len(e.TxID)
}

func (b Batch) Empty() bool { return len(b.Events) == 0 }

type Queue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []cdc_domain.ChangeEvent
	closed bool
	limit  int
}

func NewQueue(limit int) *Queue {
	if limit < 1 {
		limit = 1
	}
	q := &Queue{limit: limit, items: make([]cdc_domain.ChangeEvent, 0, limit)}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Push(ctx context.Context, e cdc_domain.ChangeEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) >= q.limit && !q.closed {
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				close(done)
			case <-time.After(10 * time.Millisecond):
			}
		}()
		q.cond.Wait()
		select {
		case <-done:
			return ctx.Err()
		default:
		}
	}
	if q.closed {
		return errors.New("queue closed")
	}
	q.items = append(q.items, e)
	q.cond.Signal()
	return nil
}

func (q *Queue) Pop(ctx context.Context) (cdc_domain.ChangeEvent, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.closed {
		if ctx.Err() != nil {
			return cdc_domain.ChangeEvent{}, ctx.Err()
		}
		q.cond.Wait()
	}
	if len(q.items) == 0 && q.closed {
		return cdc_domain.ChangeEvent{}, ioEOF{}
	}
	e := q.items[0]
	q.items = q.items[1:]
	q.cond.Broadcast()
	return e, nil
}

func (q *Queue) Close() {
	q.mu.Lock()
	q.closed = true
	q.cond.Broadcast()
	q.mu.Unlock()
}

func (q *Queue) Len() int { q.mu.Lock(); defer q.mu.Unlock(); return len(q.items) }

type ioEOF struct{}

func (ioEOF) Error() string { return "queue drained" }

type Worker struct {
	Queue      *Queue
	Sink       Sink
	Metrics    *Metrics
	BatchSize  int
	FlushEvery time.Duration
	Retry      RetryPolicy
	DeadLetter func(cdc_domain.DeadLetter)
}

func (w *Worker) Run(ctx context.Context) error {
	if w.Queue == nil || w.Sink == nil {
		return errors.New("worker queue and sink are required")
	}
	if w.BatchSize < 1 {
		w.BatchSize = 100
	}
	if w.FlushEvery <= 0 {
		w.FlushEvery = time.Second
	}
	ticker := time.NewTicker(w.FlushEvery)
	defer ticker.Stop()
	batch := NewBatch()
	flush := func() error {
		if batch.Empty() {
			return nil
		}
		err := ApplyWithRetry(ctx, w.Sink, batch.Events, w.Retry)
		if err != nil {
			if w.Metrics != nil {
				w.Metrics.Failed.Add(uint64(len(batch.Events)))
			}
			if w.DeadLetter != nil {
				for _, e := range batch.Events {
					w.DeadLetter(cdc_domain.DeadLetter{ID: e.ID, Event: e, Reason: err.Error(), Attempts: w.Retry.MaxAttempts, CreatedAt: time.Now().UTC()})
				}
			}
			batch = NewBatch()
			return err
		}
		if w.Metrics != nil {
			w.Metrics.Delivered.Add(uint64(len(batch.Events)))
		}
		batch = NewBatch()
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			_ = flush()
			return ctx.Err()
		case <-ticker.C:
			if err := flush(); err != nil {
				return err
			}
		default:
			e, err := w.Queue.Pop(ctx)
			if err != nil {
				if _, ok := err.(ioEOF); ok {
					return flush()
				}
				return err
			}
			batch.Add(e)
			if w.Metrics != nil {
				w.Metrics.Captured.Add(1)
			}
			if len(batch.Events) >= w.BatchSize {
				if err := flush(); err != nil {
					return err
				}
			}
		}
	}
}

type Supervisor struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	workers []*Worker
	running bool
}

func NewSupervisor() *Supervisor { return &Supervisor{} }
func (s *Supervisor) Start(parent context.Context, workers ...*Worker) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return errors.New("supervisor already running")
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	s.workers = workers
	s.running = true
	for _, w := range workers {
		go func(x *Worker) { _ = x.Run(ctx) }(w)
	}
	return nil
}
func (s *Supervisor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.running = false
}
func (s *Supervisor) Running() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.running }
func EnsureBatch(e []cdc_domain.ChangeEvent, max int) [][]cdc_domain.ChangeEvent {
	if max < 1 {
		max = 1
	}
	out := make([][]cdc_domain.ChangeEvent, 0)
	for len(e) > 0 {
		n := max
		if n > len(e) {
			n = len(e)
		}
		out = append(out, e[:n])
		e = e[n:]
	}
	return out
}
func ValidateBatch(events []cdc_domain.ChangeEvent) error {
	if len(events) == 0 {
		return errors.New("empty batch")
	}
	seen := map[string]bool{}
	for _, e := range events {
		if e.ID == "" {
			return errors.New("event id missing")
		}
		if seen[e.ID] {
			return fmt.Errorf("duplicate event %s", e.ID)
		}
		seen[e.ID] = true
	}
	return nil
}
