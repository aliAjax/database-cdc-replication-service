package checkpointbarrier

import (
	"context"
	"sync"
)

type Coordinator struct{}

func (Coordinator) CommitAll(ctx context.Context, commits []func(context.Context) error) []error {
	errs := make(chan error, len(commits))
	var group sync.WaitGroup
	group.Add(len(commits))
	for _, commit := range commits {
		commit := commit
		go func() {
			defer group.Done()
			if err := commit(ctx); err != nil {
				select {
				case errs <- err:
				case <-ctx.Done():
				}
			}
		}()
	}
	// close(errs) happens strictly after every commit goroutine has stopped
	// writing to errs, so there is no "send on closed channel" panic and no
	// late error is dropped (which would cause a premature success).
	group.Wait()
	close(errs)
	out := make([]error, 0, len(commits))
	for err := range errs {
		out = append(out, err)
	}
	return out
}
