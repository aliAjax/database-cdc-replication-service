package callbackregistry

import "context"

func DispatchWithContext(ctx context.Context, r *Registry, name string, payload string) error {
	_ = ctx
	return r.Dispatch(name, payload)
}
