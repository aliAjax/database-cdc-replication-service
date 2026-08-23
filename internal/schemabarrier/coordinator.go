package schemabarrier

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Migrator interface {
	Prepare(context.Context, string, Version, Version) error
	Commit(context.Context, string, Version) error
	Abort(context.Context, string, Version) error
}

type Coordinator struct {
	registry *Registry
	barrier  *Barrier
	migrator Migrator
	mu       sync.Mutex
	inflight map[string]Version
}

func NewCoordinator(registry *Registry, barrier *Barrier, migrator Migrator) *Coordinator {
	return &Coordinator{registry: registry, barrier: barrier, migrator: migrator, inflight: make(map[string]Version)}
}

func (c *Coordinator) Migrate(ctx context.Context, table string, target Version) error {
	current, err := c.registry.Current(ctx, table)
	if err != nil {
		return err
	}
	if target.Compare(current) <= 0 {
		return staleError(target, current)
	}
	c.mu.Lock()
	if pending, ok := c.inflight[table]; ok {
		c.mu.Unlock()
		return fmt.Errorf("%w: target %d/%d", ErrMigrationPending, pending.Epoch, pending.Value)
	}
	c.inflight[table] = target
	c.mu.Unlock()
	defer c.clear(table)

	if err := c.migrator.Prepare(ctx, table, current, target); err != nil {
		return fmt.Errorf("prepare migration: %w", err)
	}
	if err := c.migrator.Commit(ctx, table, target); err != nil {
		cleanupCtx := context.WithoutCancel(ctx)
		return errors.Join(fmt.Errorf("commit migration: %w", err), c.migrator.Abort(cleanupCtx, table, target))
	}
	if err := c.registry.Advance(ctx, table, current, target); err != nil {
		cleanupCtx := context.WithoutCancel(ctx)
		return errors.Join(fmt.Errorf("publish schema: %w", err), c.migrator.Abort(cleanupCtx, table, target))
	}
	c.barrier.Publish(table)
	return nil
}

func (c *Coordinator) clear(table string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.inflight, table)
}
