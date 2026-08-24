package dispatchlease

import "fmt"

type Worker struct{ pool *Pool }

func NewWorker(pool *Pool) *Worker { return &Worker{pool: pool} }
func (w *Worker) Process(items []string, failAt int) error {
	for i, item := range items {
		if err := w.processOne(item, i == failAt); err != nil {
			return err
		}
	}
	return nil
}
func (w *Worker) processOne(item string, fail bool) error {
	lease, err := w.pool.Acquire()
	if err != nil {
		return err
	}
	defer w.pool.Release(lease)
	if fail {
		return fmt.Errorf("process %s failed", item)
	}
	return nil
}
