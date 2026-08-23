package cursorresume

import (
	"context"
)

func ResumeWithRetry(ctx context.Context, attempts int, resume func(context.Context) error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for i := 0; i < attempts; i++ {
		last = resume(ctx)
		if last == nil {
			return nil
		}
	}
	return last
}
