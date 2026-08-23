package workers

import (
	"context"
	"errors"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"github.com/example/cdc-replication/internal/delivery"
	"github.com/example/cdc-replication/internal/transform"
	"github.com/example/cdc-replication/internal/wal_parser"
	"sync"
	"time"
)

type CheckpointStore interface {
	Load(string) (cdc_domain.Checkpoint, error)
	Commit(context.Context, cdc_domain.Checkpoint) error
}
type MemoryCheckpoint struct {
	mu   sync.Mutex
	data map[string]cdc_domain.Checkpoint
}

func NewMemoryCheckpoint() *MemoryCheckpoint {
	return &MemoryCheckpoint{data: map[string]cdc_domain.Checkpoint{}}
}
func (m *MemoryCheckpoint) Load(id string) (cdc_domain.Checkpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.data[id]
	if !ok {
		return c, cdc_domain.ErrNotFound
	}
	return c, nil
}
func (m *MemoryCheckpoint) Commit(ctx context.Context, c cdc_domain.Checkpoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.data[c.PipelineID]; ok && c.Position.Compare(old.Position) < 0 {
		return errors.New("checkpoint regression")
	}
	m.data[c.PipelineID] = c
	return nil
}

type Capture struct {
	SourceID    string
	PipelineID  string
	Simulator   *wal_parser.Simulator
	Queue       *delivery.Queue
	Checkpoints CheckpointStore
	Metrics     *delivery.Metrics
}

func (c *Capture) Run(ctx context.Context) error {
	if c.Simulator == nil || c.Queue == nil {
		return errors.New("capture dependencies missing")
	}
	frames := make([]wal_parser.Frame, 0)
	err := wal_parser.Stream(ctx, c.Simulator, func(f wal_parser.Frame) error {
		frames = append(frames, f)
		if f.Type == wal_parser.FrameCommit {
			events, e := wal_parser.Assemble(frames)
			if e != nil {
				return e
			}
			for _, event := range events {
				if err := c.Queue.Push(ctx, event); err != nil {
					return err
				}
			}
			frames = frames[:0]
		}
		return nil
	})
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

type Processor struct {
	PipelineID  string
	Queue       *delivery.Queue
	Mapping     transform.Mapping
	Sink        delivery.Sink
	Checkpoints CheckpointStore
	Metrics     *delivery.Metrics
	BatchSize   int
}

func (p *Processor) Run(ctx context.Context) error {
	if p.Queue == nil || p.Sink == nil {
		return errors.New("processor dependencies missing")
	}
	events := make([]cdc_domain.ChangeEvent, 0, p.BatchSize)
	for {
		e, err := p.Queue.Pop(ctx)
		if err != nil {
			return nil
		}
		res := p.Mapping.Apply(e)
		if res.Dropped {
			continue
		}
		if len(res.Errors) > 0 {
			return fmt.Errorf("transform event %s: %s", e.ID, res.Errors[0])
		}
		events = append(events, res.Event)
		if len(events) < p.BatchSize {
			continue
		}
		if err := delivery.ApplyWithRetry(ctx, p.Sink, events, delivery.RetryPolicy{MaxAttempts: 3, Initial: 20 * time.Millisecond, Max: time.Second}); err != nil {
			return err
		}
		last := events[len(events)-1]
		if p.Checkpoints != nil {
			if err := p.Checkpoints.Commit(ctx, cdc_domain.Checkpoint{PipelineID: p.PipelineID, Position: last.Position, TxID: last.TxID, CommittedAt: time.Now().UTC(), EventCount: uint64(len(events))}); err != nil {
				return err
			}
		}
		if p.Metrics != nil {
			p.Metrics.Delivered.Add(uint64(len(events)))
		}
		events = events[:0]
	}
}

type Runtime struct {
	Capture   *Capture
	Processor *Processor
	cancel    context.CancelFunc
	done      chan error
}

func (r *Runtime) Start(parent context.Context) error {
	if r.Capture == nil || r.Processor == nil {
		return errors.New("runtime not configured")
	}
	ctx, cancel := context.WithCancel(parent)
	r.cancel = cancel
	r.done = make(chan error, 2)
	go func() { r.done <- r.Capture.Run(ctx) }()
	go func() { r.done <- r.Processor.Run(ctx) }()
	return nil
}
func (r *Runtime) Stop() error {
	if r.cancel != nil {
		r.cancel()
	}
	if r.done == nil {
		return nil
	}
	select {
	case e := <-r.done:
		return e
	case <-time.After(time.Second):
		return nil
	}
}
func (r *Runtime) Wait() error {
	if r.done == nil {
		return errors.New("runtime not started")
	}
	return <-r.done
}
