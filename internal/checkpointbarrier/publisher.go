package checkpointbarrier

import (
	"context"
	"sync"
)

type Publisher struct {
	mu        sync.Mutex
	positions map[string]uint64
}

func NewPublisher() *Publisher { return &Publisher{positions: make(map[string]uint64)} }

func (p *Publisher) Publish(ctx context.Context, pipelineID string, position uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.Lock()
	// Checkpoints must only move forward. Shards commit concurrently and may
	// arrive out of order; an older position must never overwrite a newer one,
	// otherwise the recorded checkpoint regresses.
	if position < p.positions[pipelineID] {
		p.mu.Unlock()
		return nil
	}
	p.positions[pipelineID] = position
	p.mu.Unlock()
	return nil
}

func (p *Publisher) Position(pipelineID string) uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.positions[pipelineID]
}
