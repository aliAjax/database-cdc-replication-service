package snapshandoff

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrHandoffBusy  = errors.New("snapshot handoff already running")
	ErrHandoffState = errors.New("snapshot handoff state invalid")
	ErrHandoffUnavailable = errors.New("snapshot handoff unavailable")
)

type HandoffState string

const (
	StatePrepared  HandoffState = "prepared"
	StateArchived  HandoffState = "archived"
	StatePublished HandoffState = "published"
	StateAborted   HandoffState = "aborted"
)

type Handoff struct {
	mu    sync.Mutex
	state map[string]HandoffState
	cat   *Catalog
	arch  Archive
}

func NewHandoff(cat *Catalog, arch Archive) *Handoff {
	return &Handoff{state: make(map[string]HandoffState), cat: cat, arch: arch}
}

func (h *Handoff) State(stream string) HandoffState {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.state[stream]
}

func (h *Handoff) Prepare(ctx context.Context, manifest Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.mu.Lock()
	if current := h.state[manifest.Stream]; current != "" && current != StateAborted {
		h.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrHandoffBusy, current)
	}
	h.state[manifest.Stream] = StatePrepared
	h.mu.Unlock()
	return h.cat.Put(ctx, manifest)
}

func (h *Handoff) Archive(ctx context.Context, manifest Manifest) error {
	if err := h.arch.Write(ctx, manifest); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.state[manifest.Stream] != StatePrepared {
		return ErrHandoffState
	}
	h.state[manifest.Stream] = StateArchived
	return nil
}

func (h *Handoff) Publish(ctx context.Context, manifest Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.state[manifest.Stream] != StateArchived {
		return ErrHandoffState
	}
	h.state[manifest.Stream] = StatePublished
	return nil
}

func (h *Handoff) Abort(stream string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state[stream] = StateAborted
}
