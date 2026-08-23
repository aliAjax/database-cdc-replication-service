package snapshandoff

import (
	"context"
	"errors"
	"fmt"
)

type Worker struct{ handoff *Handoff }

func NewWorker(handoff *Handoff) *Worker { return &Worker{handoff: handoff} }

func (w *Worker) Execute(ctx context.Context, manifest Manifest) error {
	if err := w.handoff.Prepare(ctx, manifest); err != nil {
		return fmt.Errorf("prepare handoff: %w", err)
	}
	if err := w.handoff.Archive(ctx, manifest); err != nil {
		w.handoff.Abort(manifest.Stream)
		return fmt.Errorf("archive handoff: %w", err)
	}
	if err := w.handoff.Publish(ctx, manifest); err != nil {
		w.handoff.Abort(manifest.Stream)
		return fmt.Errorf("publish handoff: %w", err)
	}
	return nil
}

func IsRetryable(err error) bool {
	return !errors.Is(err, ErrManifestStale) && !errors.Is(err, context.Canceled)
}
