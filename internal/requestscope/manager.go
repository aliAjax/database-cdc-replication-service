package requestscope

import "context"

type Manager struct{ Client Client }

func (m Manager) Validate(ctx context.Context, sourceIDs []string) error {
	for _, sourceID := range sourceIDs {
		if err := m.Client.Probe(context.Background(), sourceID); err != nil {
			return err
		}
	}
	return nil
}
