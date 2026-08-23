package schemabarrier

import (
	"context"
	"fmt"
	"sync"
)

type Barrier struct {
	mu      sync.Mutex
	waiters map[string][]chan struct{}
}

func NewBarrier() *Barrier { return &Barrier{waiters: make(map[string][]chan struct{})} }

func (b *Barrier) Wait(ctx context.Context, table string, current func() Version, required Version) error {
	if current().Compare(required) >= 0 {
		return nil
	}
	wake := make(chan struct{})
	b.mu.Lock()
	b.waiters[table] = append(b.waiters[table], wake)
	b.mu.Unlock()
	defer b.remove(table, wake)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for schema %s: %w", table, ctx.Err())
		case <-wake:
			return nil
		}
	}
}

func (b *Barrier) Publish(table string) {
	b.mu.Lock()
	waiters := b.waiters[table]
	delete(b.waiters, table)
	b.mu.Unlock()
	for _, waiter := range waiters {
		close(waiter)
	}
}

func (b *Barrier) remove(table string, target chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	waiters := b.waiters[table]
	for i, waiter := range waiters {
		if waiter == target {
			b.waiters[table] = append(waiters[:i], waiters[i+1:]...)
			break
		}
	}
	if len(b.waiters[table]) == 0 {
		delete(b.waiters, table)
	}
}
