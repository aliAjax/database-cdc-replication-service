package checkpointbarrier

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBarrierWaitsForRegisteredCommits(t *testing.T) {
	barrier := New()
	barrier.Add(1)
	barrier.Add(1)
	barrier.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if !errors.Is(barrier.Wait(ctx), context.DeadlineExceeded) {
		t.Fatal("barrier completed before both registrations")
	}
	barrier.Done()
}

func TestCoordinatorCollectsLateErrors(t *testing.T) {
	release := make(chan struct{})
	lateErr := errors.New("late commit failed")
	commits := []func(context.Context) error{
		func(context.Context) error { return nil },
		func(context.Context) error { <-release; return lateErr },
	}
	done := make(chan []error, 1)
	go func() { done <- (Coordinator{}).CommitAll(context.Background(), commits) }()
	select {
	case <-done:
		t.Fatal("coordinator returned before late commit")
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	got := <-done
	if len(got) != 1 || !errors.Is(got[0], lateErr) {
		t.Fatalf("errors=%v", got)
	}
}

func TestPublisherIsMonotonic(t *testing.T) {
	publisher := NewPublisher()
	_ = publisher.Publish(context.Background(), "p1", 20)
	_ = publisher.Publish(context.Background(), "p1", 10)
	if got := publisher.Position("p1"); got != 20 {
		t.Fatalf("position=%d", got)
	}
}

func TestCollectorStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	input := make(chan Update)
	done := make(chan error, 1)
	go func() { done <- Collect(ctx, input, NewPublisher()) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err=%v", err)
		}
	case <-time.After(100 * time.Millisecond):
		close(input)
		t.Fatal("collector ignored cancellation")
	}
}
