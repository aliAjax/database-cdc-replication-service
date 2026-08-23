package requestscope

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBudgetReachesProbe(t *testing.T) {
	err := WithBudget(context.Background(), time.Millisecond, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return errors.New("budget did not arrive")
		}
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}

func TestClientUsesCurrentContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := Client{Fetch: func(got context.Context, _ string) error { return got.Err() }}
	if !errors.Is(client.Probe(ctx, "s1"), context.Canceled) {
		t.Fatal("client ignored current request context")
	}
}

func TestRetryStopsOnAbortSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	worker := Worker{Step: func(context.Context, int) error {
		calls++
		cancel()
		return errors.New("retry")
	}}
	if !errors.Is(worker.Retry(ctx, 5), context.Canceled) || calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestSecondValidationUsesOwnDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	manager := Manager{Client: Client{Fetch: func(got context.Context, _ string) error { return got.Err() }}}
	if !errors.Is(manager.Validate(ctx, []string{"s1"}), context.Canceled) {
		t.Fatal("manager replaced canceled request context")
	}
}
