package replaycontrol

import "context"

type ReplayManager struct {
	Client *ReplayClient
}

func (m *ReplayManager) Resume(ctx context.Context, cursor string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.Client.Check(ctx, cursor)
}
