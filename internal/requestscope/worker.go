package requestscope

import (
	"context"
	"errors"
)

type Worker struct {
	Step func(context.Context, int) error
}

func (w Worker) Retry(ctx context.Context, attempts int) error {
	var last error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = w.Step(ctx, attempt)
		if last == nil {
			return nil
		}
		if errors.Is(last, context.Canceled) || errors.Is(last, context.DeadlineExceeded) {
			return last
		}
	}
	return last
}
