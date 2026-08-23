package dispatchlease

import (
	"errors"
	"testing"
)

func TestDispatchLeaseReleaseReturnsLease(t *testing.T) {
	p := NewPool(1)
	l, err := p.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	p.Release(l)
	p.Release(Lease{})
	if _, err := p.Acquire(); err != nil {
		t.Fatalf("released lease unavailable: %v", err)
	}
}

func TestDispatchLeaseWorkerFailureReleasesLease(t *testing.T) {
	p := NewPool(1)
	if err := NewWorker(p).Process([]string{"bad"}, 0); err == nil {
		t.Fatal("expected processing failure")
	}
	if _, err := p.Acquire(); err != nil {
		t.Fatalf("worker leaked lease: %v", err)
	}
}

func TestDispatchLeaseFinishRollsBackJoinedFailure(t *testing.T) {
	tx := &Transaction{committed: true}
	op := errors.New("write failed")
	err := FinalizeTransaction(tx, op, nil)
	if !errors.Is(err, op) || !errors.Is(err, ErrRollback) || tx.committed {
		t.Fatalf("operation lost or transaction committed: %v", err)
	}
}

func TestDispatchLeaseAuditPreservesFailures(t *testing.T) {
	a := &Audit{}
	op := errors.New("audit write failed")
	err := a.Flush(op)
	if !errors.Is(err, op) || !errors.Is(err, ErrAuditFlush) || a.Flushed() {
		t.Fatalf("flush errors lost: %v", err)
	}
}
