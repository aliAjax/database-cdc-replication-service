package lookupfail

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestRepositoryPreservesMissingCause(t *testing.T) {
	_, err := NewRepository(nil).Load("absent")
	if !errors.Is(err, ErrStreamMissing) {
		t.Fatalf("missing cause lost: %v", err)
	}
}

func TestClassifyWrappedMissing(t *testing.T) {
	err := fmt.Errorf("service lookup: %w", ErrStreamMissing)
	if got := Classify(err); got != KindMissing {
		t.Fatalf("kind=%s", got)
	}
}

func TestMissingMapsToNotFound(t *testing.T) {
	if got := StatusFor(fmt.Errorf("wrapped: %w", ErrStreamMissing)); got != http.StatusNotFound {
		t.Fatalf("status=%d", got)
	}
}

func TestMissingDoesNotRetry(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), 5, func() error {
		calls++
		return fmt.Errorf("repository: %w", ErrStreamMissing)
	})
	if !errors.Is(err, ErrStreamMissing) || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}
