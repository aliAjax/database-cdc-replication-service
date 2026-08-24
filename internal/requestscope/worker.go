package requestscope

import "context"

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
	}
	return last
}
