package replaycontrol

import "context"

type Fetcher func(context.Context, string) error

type ReplayClient struct {
	Fetch Fetcher
}

func (c *ReplayClient) Check(ctx context.Context, cursor string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.Fetch(ctx, cursor)
}
