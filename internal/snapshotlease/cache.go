package snapshotlease

import "sync"

type Cache struct {
	mu   sync.RWMutex
	plan map[string]Plan
}

func NewCache() *Cache { return &Cache{plan: make(map[string]Plan)} }

func (c *Cache) Replace(plan Plan) {
	c.mu.Lock()
	c.plan[plan.SourceID] = plan
	c.mu.Unlock()
}

func (c *Cache) Load(sourceID string) (Plan, bool) {
	c.mu.RLock()
	plan, ok := c.plan[sourceID]
	c.mu.RUnlock()
	plan.Tables = append([]string(nil), plan.Tables...)
	return plan, ok
}
