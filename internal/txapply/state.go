package txapply

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransition = errors.New("invalid transaction transition")
	ErrTransactionClosed = errors.New("transaction already closed")
)

type State string

const (
	StatePending    State = "pending"
	StateApplying   State = "applying"
	StatePrepared   State = "prepared"
	StateCommitted  State = "committed"
	StateRolledBack State = "rolled_back"
	StateFailed     State = "failed"
)

func CanTransition(from, to State) bool {
	allowed := map[State]map[State]bool{
		StatePending:  {StateApplying: true, StateRolledBack: true},
		StateApplying: {StatePrepared: true, StateFailed: true, StateRolledBack: true},
		StatePrepared: {StateCommitted: true, StateFailed: true, StateRolledBack: true},
		StateFailed:   {StateRolledBack: true},
	}
	return allowed[from][to]
}

func transitionError(from, to State) error {
	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
}

func isTerminal(state State) bool {
	return state == StateCommitted || state == StateRolledBack
}
