package streamfanout

import (
	"context"
	"sync"
)

type Coordinator struct{}

func (Coordinator) Run(ctx context.Context, jobs []func(context.Context) error) []error {
	errs := make(chan error, len(jobs))
	var group sync.WaitGroup
	group.Add(len(jobs))
	for _, job := range jobs {
		job := job
		go func() {
			defer group.Done()
			if err := job(ctx); err != nil {
				errs <- err
			}
		}()
	}
	group.Wait()
	close(errs)
	return CollectErrors(errs)
}
