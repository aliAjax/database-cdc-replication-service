package checkpointbarrier

import "context"

type Coordinator struct{}

func (Coordinator) CommitAll(ctx context.Context, commits []func(context.Context) error) []error {
	barrier := New()
	barrier.Add(len(commits))
	errs := make(chan error, len(commits))
	for _, commit := range commits {
		commit := commit
		go func() {
			defer barrier.Done()
			if err := commit(ctx); err != nil {
				errs <- err
			}
		}()
	}
	_ = barrier.Wait(ctx)
	close(errs)
	out := make([]error, 0, len(errs))
	for err := range errs {
		out = append(out, err)
	}
	return out
}
