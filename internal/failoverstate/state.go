package failoverstate

import "fmt"

type State string

const (
	StateStreaming  State = "streaming"
	StateDegraded   State = "degraded"
	StateRecovering State = "recovering"
	StateFailed     State = "failed"
)

func Transition(current, next State) error {
	allowed := map[State]map[State]bool{
		StateStreaming:  {StateDegraded: true, StateFailed: true},
		StateDegraded:   {StateFailed: true},
		StateRecovering: {StateFailed: true},
		StateFailed:     {},
	}
	if !allowed[current][next] {
		return fmt.Errorf("invalid failover transition %s -> %s", current, next)
	}
	return nil
}
