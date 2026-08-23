package replaycontrol

import (
	"context"
	"time"
)

type Step func(context.Context) error

func RetryReplay(ctx context.Context, attempts int, delay time.Duration, step Step) error {
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if last = step(ctx); last == nil {
			return nil
		}
		if attempt+1 == attempts {
			break
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
	return last
}
