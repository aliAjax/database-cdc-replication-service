package slotlease

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Keeper struct {
	store Store
	ttl   time.Duration
	tick  time.Duration
	mu    sync.Mutex
	lease Lease
}

func NewKeeper(store Store, ttl, tick time.Duration) *Keeper {
	return &Keeper{store: store, ttl: ttl, tick: tick}
}

func (k *Keeper) Current() Lease {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.lease
}

func (k *Keeper) Run(ctx context.Context, slot, owner string, ready chan<- Lease) error {
	lease, err := k.store.Acquire(ctx, slot, owner, k.ttl)
	if err != nil {
		return err
	}
	k.set(lease)
	if ready != nil {
		select {
		case ready <- lease:
		case <-ctx.Done():
			return k.release(lease)
		}
	}
	ticker := time.NewTicker(k.tick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return k.release(lease)
		case <-ticker.C:
			renewed, renewErr := k.store.Renew(ctx, lease, k.ttl)
			if renewErr != nil {
				return errors.Join(renewErr, k.release(lease))
			}
			lease = renewed
			k.set(renewed)
		}
	}
}

func (k *Keeper) release(lease Lease) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), k.tick)
	defer cancel()
	err := k.store.Release(cleanupCtx, lease)
	if err == nil || errors.Is(err, ErrLeaseOwner) {
		k.set(Lease{})
	}
	return err
}

func (k *Keeper) set(lease Lease) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.lease = lease
}
