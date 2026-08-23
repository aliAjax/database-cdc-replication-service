package failoverstate

import "sync"

type Machine struct {
	mu    sync.Mutex
	state State
}

func NewMachine(initial State) *Machine { return &Machine{state: initial} }

func (m *Machine) Move(next State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := Transition(m.state, next); err != nil {
		return err
	}
	m.state = next
	return nil
}

func (m *Machine) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}
