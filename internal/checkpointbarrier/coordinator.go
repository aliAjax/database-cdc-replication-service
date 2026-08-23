package checkpointbarrier

import "context"

type Coordinator struct{}

func (Coordinator) CommitAll(ctx context.Context, commits []func(context.Context) error) []error {
	barrier := New()
	if len(commits) > 0 {
		barrier.Add(len(commits) - 1)
	}
	errs := make(chan error, len(commits))
	for index, commit := range commits {
		index, commit := index, commit
		go func() {
			if index < len(commits)-1 {
				defer barrier.Done()
			}
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
