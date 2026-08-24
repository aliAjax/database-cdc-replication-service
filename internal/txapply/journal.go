package txapply

import (
	"context"
	"sync"
)

type JournalEntry struct {
	TxID  string
	State State
	Count int
}

type Journal interface {
	Append(context.Context, JournalEntry) error
}

type MemoryJournal struct {
	mu      sync.Mutex
	entries []JournalEntry
	err     error
}

func (j *MemoryJournal) Append(ctx context.Context, entry JournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.err != nil {
		return j.err
	}
	j.entries = append(j.entries, entry)
	return nil
}

func (j *MemoryJournal) Entries() []JournalEntry {
	j.mu.Lock()
	defer j.mu.Unlock()
	return append([]JournalEntry(nil), j.entries...)
}

func (j *MemoryJournal) Fail(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.err = err
}
