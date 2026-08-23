package backfillmanifest

import "sync"

type Cache struct {
	mu    sync.RWMutex
	items map[string]Manifest
}

func NewCache() *Cache {
	return &Cache{items: make(map[string]Manifest)}
}

func (c *Cache) Set(manifest Manifest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[manifest.ID] = cloneManifest(manifest)
}

func (c *Cache) Load(id string) (Manifest, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	manifest, ok := c.items[id]
	if !ok {
		return Manifest{}, false
	}
	return cloneManifest(manifest), true
}
