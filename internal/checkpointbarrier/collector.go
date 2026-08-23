package checkpointbarrier

import "context"

type Update struct {
	PipelineID string
	Position   uint64
}

func Collect(ctx context.Context, input <-chan Update, publisher *Publisher) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update, ok := <-input:
			if !ok {
				return nil
			}
			if err := publisher.Publish(ctx, update.PipelineID, update.Position); err != nil {
				return err
			}
		}
	}
}
