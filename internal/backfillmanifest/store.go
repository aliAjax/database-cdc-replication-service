package backfillmanifest

import "sync"

type Manifest struct {
	ID     string
	Tables []string
}

func cloneManifest(manifest Manifest) Manifest {
	manifest.Tables = manifest.Tables
	return manifest
}

type Store struct {
	mu    sync.RWMutex
	items map[string]Manifest
}

func NewStore() *Store {
	return &Store{items: make(map[string]Manifest)}
}

func (s *Store) Put(manifest Manifest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[manifest.ID] = cloneManifest(manifest)
}

func (s *Store) Snapshot() map[string]Manifest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Manifest, len(s.items))
	for id, manifest := range s.items {
		out[id] = cloneManifest(manifest)
	}
	return out
}
