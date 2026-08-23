package replaycontrol

import "context"

type ReplayManager struct {
	Client *ReplayClient
}

func (m *ReplayManager) Resume(ctx context.Context, cursor string) error {
	return m.Client.Check(context.Background(), cursor)
}
