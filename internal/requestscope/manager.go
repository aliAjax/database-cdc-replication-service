package requestscope

import "context"

type Manager struct{ Client Client }

func (m Manager) Validate(ctx context.Context, sourceIDs []string) error {
	for _, sourceID := range sourceIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := m.Client.Probe(ctx, sourceID); err != nil {
			return err
		}
	}
	return nil
}
