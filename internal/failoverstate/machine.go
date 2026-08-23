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
	current := m.state
	m.state = next
	if err := Transition(current, next); err != nil {
		return err
	}
	return nil
}

func (m *Machine) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}
