package slotlease

import (
	"context"
	"sync"
	"time"
)

type Store interface {
	Acquire(context.Context, string, string, time.Duration) (Lease, error)
	Renew(context.Context, Lease, time.Duration) (Lease, error)
	Release(context.Context, Lease) error
	Get(context.Context, string) (Lease, bool)
}

type MemoryStore struct {
	mu     sync.Mutex
	now    func() time.Time
	leases map[string]Lease
}

func NewMemoryStore(now func() time.Time) *MemoryStore {
	if now == nil {
		now = time.Now
	}
	return &MemoryStore{now: now, leases: make(map[string]Lease)}
}

func (s *MemoryStore) Acquire(ctx context.Context, slot, owner string, ttl time.Duration) (Lease, error) {
	if err := ctx.Err(); err != nil {
		return Lease{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	current := s.leases[slot]
	if current.Active(now) && current.Owner != owner {
		return Lease{}, ErrLeaseHeld
	}
	lease := Lease{Slot: slot, Owner: owner, Generation: current.Generation + 1, ExpiresAt: now.Add(ttl)}
	s.leases[slot] = lease
	return lease, nil
}

func (s *MemoryStore) Renew(ctx context.Context, lease Lease, ttl time.Duration) (Lease, error) {
	if err := ctx.Err(); err != nil {
		return Lease{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	current, ok := s.leases[lease.Slot]
	if !ok || !current.Active(now) {
		return Lease{}, ErrLeaseExpired
	}
	if current.Owner != lease.Owner {
		return Lease{}, ErrLeaseOwner
	}
	current.ExpiresAt = now.Add(ttl)
	s.leases[lease.Slot] = current
	return current, nil
}

func (s *MemoryStore) Release(ctx context.Context, lease Lease) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.leases[lease.Slot]
	if !ok {
		return nil
	}
	if current.Owner != lease.Owner || current.Generation != lease.Generation {
		return ErrLeaseOwner
	}
	delete(s.leases, lease.Slot)
	return nil
}

func (s *MemoryStore) Get(ctx context.Context, slot string) (Lease, bool) {
	if ctx.Err() != nil {
		return Lease{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	lease, ok := s.leases[slot]
	return lease, ok
}
