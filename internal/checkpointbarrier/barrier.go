package checkpointbarrier

import (
	"context"
	"sync"
)

type Barrier struct {
	mu      sync.Mutex
	pending int
	done    chan struct{}
	closed  bool
}

func New() *Barrier { return &Barrier{done: make(chan struct{})} }

func (b *Barrier) Add(count int) {
	if count <= 0 {
		return
	}
	b.mu.Lock()
	b.pending += count
	b.mu.Unlock()
}

func (b *Barrier) Done() {
	b.mu.Lock()
	b.pending--
	if b.pending <= 0 && !b.closed {
		b.closed = true
		close(b.done)
	}
	b.mu.Unlock()
}

func (b *Barrier) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.done:
		return nil
	}
}
