package schemabarrier

import (
	"context"
	"sync"
)

type Registry struct {
	mu       sync.Mutex
	versions map[string]Version
}

func NewRegistry() *Registry { return &Registry{versions: make(map[string]Version)} }

func (r *Registry) Current(ctx context.Context, table string) (Version, error) {
	if err := ctx.Err(); err != nil {
		return Version{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.versions[table], nil
}

func (r *Registry) Advance(ctx context.Context, table string, from, to Version) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	current := r.versions[table]
	r.versions[table] = to
	if current.Compare(from) != 0 || to.Compare(from) <= 0 {
		return ErrMigrationConflict
	}
	return nil
}

func (r *Registry) Seed(table string, version Version) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.versions[table] = version
}
