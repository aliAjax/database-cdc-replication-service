package failoverstate

import "context"

type Worker struct {
	Machine *Machine
	Recover func(context.Context) error
}

func (w Worker) Run(ctx context.Context) error {
	if err := w.Machine.Move(StateRecovering); err != nil {
		return err
	}
	if err := w.Recover(ctx); err != nil {
		_ = w.Machine.Move(StateFailed)
		return err
	}
	if err := w.Machine.Move(StateStreaming); err != nil {
		return err
	}
	return nil
}
