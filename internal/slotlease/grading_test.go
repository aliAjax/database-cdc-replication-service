package slotlease

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSlotLeaseExpiryAllowsTakeover(t *testing.T) {
	now := time.Unix(100, 0)
	store := NewMemoryStore(func() time.Time { return now })
	if _, err := store.Acquire(context.Background(), "orders_slot", "capture-a", time.Second); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Second)
	lease, err := store.Acquire(context.Background(), "orders_slot", "capture-b", time.Second)
	if err != nil {
		t.Fatalf("expired lease could not be replaced: %v", err)
	}
	if lease.Owner != "capture-b" {
		t.Fatalf("owner = %q, want capture-b", lease.Owner)
	}
}

func TestSlotLeaseStaleHandleCannotRenew(t *testing.T) {
	now := time.Unix(200, 0)
	store := NewMemoryStore(func() time.Time { return now })
	oldLease, err := store.Acquire(context.Background(), "events_slot", "capture-a", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	current, err := store.Acquire(context.Background(), "events_slot", "capture-a", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if current.Generation <= oldLease.Generation {
		t.Fatal("reacquire did not advance generation")
	}
	if _, err := store.Renew(context.Background(), oldLease, time.Minute); !errors.Is(err, ErrLeaseOwner) {
		t.Fatalf("stale Renew error = %v, want %v", err, ErrLeaseOwner)
	}
}

func TestSlotLeaseCancellationReleasesOwnership(t *testing.T) {
	store := NewMemoryStore(nil)
	keeper := NewKeeper(store, time.Second, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan Lease, 1)
	done := make(chan error, 1)
	go func() { done <- keeper.Run(ctx, "audit_slot", "capture-c", ready) }()
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("keeper did not acquire lease")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("keeper cancellation returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("keeper did not stop after cancellation")
	}
	if _, ok := store.Get(context.Background(), "audit_slot"); ok {
		t.Fatal("canceled keeper left lease in store")
	}
	if lease := keeper.Current(); lease.Owner != "" {
		t.Fatalf("keeper retained current owner %q", lease.Owner)
	}
}

func TestSlotLeaseFenceRejectsStaleGeneration(t *testing.T) {
	now := time.Unix(300, 0)
	store := NewMemoryStore(func() time.Time { return now })
	oldLease, err := store.Acquire(context.Background(), "schema_slot", "capture-d", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	current, err := store.Acquire(context.Background(), "schema_slot", "capture-d", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	fence := NewFence(store, func() time.Time { return now })
	if err := fence.Authorize(context.Background(), oldLease); !errors.Is(err, ErrLeaseOwner) {
		t.Fatalf("stale fence error = %v, want %v", err, ErrLeaseOwner)
	}
	now = now.Add(2 * time.Minute)
	if err := fence.Authorize(context.Background(), current); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("expired fence error = %v, want %v", err, ErrLeaseExpired)
	}
}
