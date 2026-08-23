package snapshandoff

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrManifestMissing = errors.New("snapshot manifest missing")
	ErrManifestStale   = errors.New("snapshot manifest is stale")
)

type Manifest struct {
	Stream string
	Epoch  uint64
	Files  []string
	Meta   map[string]string
}

func (m Manifest) Clone() Manifest {
	clone := Manifest{Stream: m.Stream, Epoch: m.Epoch, Files: append([]string(nil), m.Files...), Meta: make(map[string]string, len(m.Meta))}
	for k, v := range m.Meta {
		clone.Meta[k] = v
	}
	return clone
}

type Catalog struct {
	mu        sync.Mutex
	manifests map[string]Manifest
}

func NewCatalog() *Catalog { return &Catalog{manifests: make(map[string]Manifest)} }

func (c *Catalog) Put(ctx context.Context, manifest Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	current, ok := c.manifests[manifest.Stream]
	if ok && manifest.Epoch <= current.Epoch {
		return fmt.Errorf("%w: %s %d <= %d", ErrManifestStale, manifest.Stream, manifest.Epoch, current.Epoch)
	}
	c.manifests[manifest.Stream] = manifest.Clone()
	return nil
}

func (c *Catalog) Get(ctx context.Context, stream string) (Manifest, error) {
	if err := ctx.Err(); err != nil {
		return Manifest{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	manifest, ok := c.manifests[stream]
	if !ok {
		return Manifest{}, ErrManifestMissing
	}
	return manifest.Clone(), nil
}
