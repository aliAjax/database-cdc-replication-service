package replaycontrol

import (
	"context"
	"time"
)

func WithReplayBudget(parent context.Context, timeout time.Duration, next func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	_ = ctx
	return next(context.Background())
}
