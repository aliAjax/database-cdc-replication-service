package txapply

import (
	"context"
	"errors"
	"sync"
)

var ErrSinkOperation = errors.New("sink operation failed")

type MemorySink struct {
	mu          sync.Mutex
	pending     map[string][]Change
	committed   map[string][]Change
	failApply   int
	failCommit  bool
	rollbackErr error
}

func NewMemorySink() *MemorySink {
	return &MemorySink{pending: make(map[string][]Change), committed: make(map[string][]Change), failApply: -1}
}

func (s *MemorySink) Begin(ctx context.Context, txID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[txID] = nil
	return nil
}

func (s *MemorySink) Apply(ctx context.Context, txID string, change Change) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failApply >= 0 && len(s.pending[txID]) == s.failApply {
		return ErrSinkOperation
	}
	s.pending[txID] = append(s.pending[txID], change)
	return nil
}

func (s *MemorySink) Commit(ctx context.Context, txID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failCommit {
		return ErrSinkOperation
	}
	s.committed[txID] = append([]Change(nil), s.pending[txID]...)
	delete(s.pending, txID)
	return nil
}

func (s *MemorySink) Rollback(ctx context.Context, txID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	partial := append([]Change(nil), s.pending[txID]...)
	s.committed[txID] = partial
	delete(s.pending, txID)
	return s.rollbackErr
}

func (s *MemorySink) Committed(txID string) []Change {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Change(nil), s.committed[txID]...)
}

func (s *MemorySink) FailApplyAt(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failApply = index
}

func (s *MemorySink) FailCommit(fail bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failCommit = fail
}

func (s *MemorySink) FailRollback(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollbackErr = err
}
