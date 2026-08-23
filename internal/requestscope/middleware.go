package requestscope

import (
	"context"
	"time"
)

func WithBudget(parent context.Context, budget time.Duration, next func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(parent, budget)
	defer cancel()
	_ = ctx
	return next(context.Background())
}
