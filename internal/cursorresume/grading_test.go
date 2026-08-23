package cursorresume

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestCursorResumeDecodePreservesInvalidSentinel(t *testing.T) {
	for _, token := range []string{"%%%", "e2JhZA"} {
		_, err := DecodeCursor(token)
		if !errors.Is(err, ErrInvalidCursor) {
			t.Fatalf("token %q lost invalid cursor sentinel: %v", token, err)
		}
	}
}

func TestCursorResumeWrappedInvalidStopsRetry(t *testing.T) {
	calls := 0
	err := ResumeWithRetry(context.Background(), 4, func(context.Context) error {
		calls++
		return fmt.Errorf("resume rejected: %w", ErrInvalidCursor)
	})
	if !errors.Is(err, ErrInvalidCursor) || calls != 1 {
		t.Fatalf("wrapped invalid cursor retried: calls=%d err=%v", calls, err)
	}
}

type failingCheckpointStore struct {
	loadErr error
	saveErr error
}

func (s *failingCheckpointStore) Load(context.Context, string) (Cursor, error) {
	return Cursor{Stream: "orders", Epoch: 1, Offset: 1}, s.loadErr
}

func (s *failingCheckpointStore) Save(context.Context, Cursor) error { return s.saveErr }

func TestCursorResumeStoreErrorsPreserveCause(t *testing.T) {
	token, err := EncodeCursor(Cursor{Stream: "orders", Epoch: 1, Offset: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewResumer(nil).Resume(context.Background(), token); !errors.Is(err, ErrCheckpointStoreUnavailable) {
		t.Fatalf("nil store was not classified: %v", err)
	}
	store := &failingCheckpointStore{loadErr: context.Canceled}
	if _, err := NewResumer(store).Resume(context.Background(), token); !errors.Is(err, context.Canceled) {
		t.Fatalf("load cancellation lost: %v", err)
	}
	store.loadErr = nil
	store.saveErr = context.DeadlineExceeded
	if _, err := NewResumer(store).Resume(context.Background(), token); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("save deadline lost: %v", err)
	}
}

func TestCursorResumeAdvanceErrorsPreserveSentinel(t *testing.T) {
	cases := []struct {
		current Cursor
		next    Cursor
	}{
		{Cursor{Stream: "orders"}, Cursor{Stream: "users"}},
		{Cursor{Stream: "orders", Epoch: 2}, Cursor{Stream: "orders", Epoch: 1}},
		{Cursor{Stream: "orders", Epoch: 2, Offset: 9}, Cursor{Stream: "orders", Epoch: 2, Offset: 8}},
	}
	for _, tc := range cases {
		if err := ValidateAdvance(tc.current, tc.next); !errors.Is(err, ErrInvalidCursor) {
			t.Fatalf("advance error lost invalid cursor sentinel: %v", err)
		}
	}
}
