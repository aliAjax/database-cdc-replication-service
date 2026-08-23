package checkpointbarrier

import "context"

type Update struct {
	PipelineID string
	Position   uint64
}

func Collect(ctx context.Context, input <-chan Update, publisher *Publisher) error {
	for update := range input {
		if err := publisher.Publish(context.Background(), update.PipelineID, update.Position); err != nil {
			return err
		}
	}
	return nil
}
