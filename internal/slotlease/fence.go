package slotlease

import (
	"context"
	"fmt"
	"time"
)

type Fence struct {
	store Store
	now   func() time.Time
}

func NewFence(store Store, now func() time.Time) *Fence {
	if now == nil {
		now = time.Now
	}
	return &Fence{store: store, now: now}
}

func (f *Fence) Authorize(ctx context.Context, presented Lease) error {
	current, ok := f.store.Get(ctx, presented.Slot)
	if !ok {
		return fmt.Errorf("authorize write: %w", ErrLeaseExpired)
	}
	if current.Owner != presented.Owner {
		return fmt.Errorf("authorize write: %w", ErrLeaseOwner)
	}
	return nil
}
