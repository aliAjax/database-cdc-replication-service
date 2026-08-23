package streamfanout

import "context"

type Producer struct{}

func (Producer) Send(ctx context.Context, values []int) (<-chan int, <-chan error) {
	out := make(chan int)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		for _, value := range values {
			select {
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			case out <- value:
			}
		}
	}()
	return out, errs
}
