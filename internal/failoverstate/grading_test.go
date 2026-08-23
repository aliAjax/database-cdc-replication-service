package failoverstate

import (
	"context"
	"testing"
)

func TestRecoveryTransitionsAreAllowed(t *testing.T) {
	if err := Transition(StateDegraded, StateRecovering); err != nil {
		t.Fatal(err)
	}
	if err := Transition(StateRecovering, StateStreaming); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidMoveKeepsState(t *testing.T) {
	machine := NewMachine(StateStreaming)
	if err := machine.Move(StateRecovering); err == nil {
		t.Fatal("illegal move accepted")
	}
	if got := machine.State(); got != StateStreaming {
		t.Fatalf("state=%s", got)
	}
}

func TestWorkerPublishesTerminalState(t *testing.T) {
	machine := NewMachine(StateDegraded)
	worker := Worker{Machine: machine, Recover: func(context.Context) error { return nil }}
	if err := worker.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := machine.State(); got != StateStreaming {
		t.Fatalf("state=%s", got)
	}
}

func TestRecoveringRemainsActive(t *testing.T) {
	got := Active(map[string]State{"p1": StateRecovering})
	if len(got) != 1 || got[0] != "p1" {
		t.Fatalf("active=%v", got)
	}
}
