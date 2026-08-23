package batchresources

import (
	"errors"
	"testing"
)

type fakeFile struct {
	open  *int
	limit int
}

func (f *fakeFile) Write([]byte) error {
	if *f.open > f.limit {
		return errors.New("too many open files")
	}
	return nil
}
func (f *fakeFile) Close() error { *f.open--; return nil }

func TestFilesClosePerIteration(t *testing.T) {
	open := 4
	files := []File{
		&fakeFile{open: &open, limit: 4},
		&fakeFile{open: &open, limit: 3},
		&fakeFile{open: &open, limit: 2},
		&fakeFile{open: &open, limit: 1},
	}
	if err := WriteFiles(files, []byte("segment")); err != nil {
		t.Fatal(err)
	}
	if open != 0 {
		t.Fatalf("open=%d", open)
	}
}

type fakeTx struct {
	committed, rolled      bool
	commitErr, rollbackErr error
}

func (f *fakeTx) Commit() error   { f.committed = true; return f.commitErr }
func (f *fakeTx) Rollback() error { f.rolled = true; return f.rollbackErr }

func TestTransactionKeepsOperationError(t *testing.T) {
	operationErr := errors.New("sink write failed")
	commitErr := errors.New("commit failed")
	tx := &fakeTx{commitErr: commitErr}
	err := Finish(tx, func() error { return operationErr })
	if !errors.Is(err, operationErr) || !tx.rolled || tx.committed {
		t.Fatalf("err=%v rolled=%v committed=%v", err, tx.rolled, tx.committed)
	}
}

type fakeLease struct {
	released bool
	err      error
}

func (f *fakeLease) Release() error { f.released = true; return f.err }

func TestLeaseAlwaysReleases(t *testing.T) {
	lease := &fakeLease{}
	operationErr := errors.New("apply failed")
	err := WithLease(lease, func() error { return operationErr })
	if !lease.released || !errors.Is(err, operationErr) {
		t.Fatalf("released=%v err=%v", lease.released, err)
	}
}

type fakeAudit struct{ flushErr error }

func (f fakeAudit) Append(string) error { return errors.New("append failed") }
func (f fakeAudit) Flush() error        { return f.flushErr }

func TestAuditFlushErrorIsJoined(t *testing.T) {
	flushErr := errors.New("flush failed")
	err := Record(fakeAudit{flushErr: flushErr}, []string{"entry"})
	if err == nil || !errors.Is(err, flushErr) {
		t.Fatalf("err=%v", err)
	}
}
