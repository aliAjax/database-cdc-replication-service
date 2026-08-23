package replaycontrol

import "context"

type Fetcher func(context.Context, string) error

type ReplayClient struct {
	Fetch Fetcher
}

func (c *ReplayClient) Check(ctx context.Context, cursor string) error {
	return c.Fetch(context.Background(), cursor)
}
