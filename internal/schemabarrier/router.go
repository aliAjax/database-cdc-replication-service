package schemabarrier

import (
	"context"
	"fmt"
)

type Event struct {
	Table   string
	Version Version
	Payload []byte
}

type Handler interface {
	Apply(context.Context, Event) error
}

type Router struct {
	registry *Registry
	barrier  *Barrier
	handler  Handler
}

func NewRouter(registry *Registry, barrier *Barrier, handler Handler) *Router {
	return &Router{registry: registry, barrier: barrier, handler: handler}
}

func (r *Router) Route(ctx context.Context, event Event) error {
	current, err := r.registry.Current(ctx, event.Table)
	if err != nil {
		return err
	}
	if event.Version.Compare(current) < 0 {
		return staleError(event.Version, current)
	}
	if event.Version.Compare(current) > 0 {
		err = r.barrier.Wait(ctx, event.Table, func() Version {
			version, _ := r.registry.Current(context.WithoutCancel(ctx), event.Table)
			return version
		}, event.Version)
		if err != nil {
			return err
		}
	}
	if err := r.handler.Apply(ctx, event); err != nil {
		return fmt.Errorf("apply %s event: %w", event.Table, err)
	}
	return nil
}
