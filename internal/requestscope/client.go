package requestscope

import "context"

type Client struct {
	Fetch func(context.Context, string) error
}

func (c Client) Probe(ctx context.Context, sourceID string) error {
	return c.Fetch(ctx, sourceID)
}
