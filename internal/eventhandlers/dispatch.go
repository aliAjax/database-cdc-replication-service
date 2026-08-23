package eventhandlers

import "context"

func DispatchWithContext(ctx context.Context, r *Registry, name string, payload string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Dispatch(name, payload)
}
