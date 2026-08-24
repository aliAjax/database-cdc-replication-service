package transactionwindow

import "sync"

type History struct {
	mu      sync.RWMutex
	windows [][]Change
}

func (h *History) Append(changes []Change) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.windows = append(h.windows, CloneAll(changes))
}

func (h *History) Latest() []Change {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.windows) == 0 {
		return nil
	}
	return CloneAll(h.windows[len(h.windows)-1])
}
