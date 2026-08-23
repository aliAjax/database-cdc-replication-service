package streamfanout

import "context"

type Consumer struct{}

func (Consumer) Drain(ctx context.Context, values <-chan int) ([]int, error) {
	out := make([]int, 0)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case value, ok := <-values:
			if !ok {
				return out, nil
			}
			out = append(out, value)
		}
	}
}
