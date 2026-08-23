package snapshotlease

import "sync"

type Snapshot struct {
	SourceID string
	Version  uint64
	Tables   []string
	Labels   map[string]string
}

func cloneSnapshot(in Snapshot) Snapshot {
	out := in
	out.Tables = append([]string(nil), in.Tables...)
	out.Labels = make(map[string]string, len(in.Labels))
	for key, value := range in.Labels {
		out.Labels[key] = value
	}
	return out
}

type Store struct {
	mu    sync.RWMutex
	items map[string]Snapshot
}

func NewStore() *Store { return &Store{items: make(map[string]Snapshot)} }

func (s *Store) Put(snapshot Snapshot) {
	s.mu.Lock()
	s.items[snapshot.SourceID] = cloneSnapshot(snapshot)
	s.mu.Unlock()
}

func (s *Store) Get(sourceID string) (Snapshot, bool) {
	s.mu.RLock()
	snapshot, ok := s.items[sourceID]
	s.mu.RUnlock()
	if !ok {
		return Snapshot{}, false
	}
	return cloneSnapshot(snapshot), true
}
