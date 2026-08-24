package snapshandoff

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrArchiveWrite = errors.New("snapshot archive write failed")

type Archive interface {
	Write(context.Context, Manifest) error
	Delete(context.Context, Manifest) error
}

type MemoryArchive struct {
	mu        sync.Mutex
	items     map[string]Manifest
	writeErr  error
	deleteErr error
}

func NewMemoryArchive() *MemoryArchive { return &MemoryArchive{items: make(map[string]Manifest)} }

func (a *MemoryArchive) Write(ctx context.Context, manifest Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.writeErr != nil {
		return fmt.Errorf("archive write: %w", a.writeErr)
	}
	a.items[manifest.Stream] = manifest.Clone()
	return nil
}

func (a *MemoryArchive) Delete(ctx context.Context, manifest Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.deleteErr != nil {
		return a.deleteErr
	}
	delete(a.items, manifest.Stream)
	return nil
}

func (a *MemoryArchive) Has(stream string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.items[stream]
	return ok
}
func (a *MemoryArchive) FailWrite(err error)  { a.mu.Lock(); defer a.mu.Unlock(); a.writeErr = err }
func (a *MemoryArchive) FailDelete(err error) { a.mu.Lock(); defer a.mu.Unlock(); a.deleteErr = err }
