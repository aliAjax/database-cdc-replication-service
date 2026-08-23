package cursorresume

import (
	"context"
	"errors"
)

func ResumeWithRetry(ctx context.Context, attempts int, resume func(context.Context) error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = resume(ctx)
		if last == nil {
			return nil
		}
		if errors.Is(last, ErrInvalidCursor) {
			return last
		}
	}
	return last
}
