package cursorresume

import "context"

type CheckpointStore interface {
	Load(context.Context, string) (Cursor, error)
	Save(context.Context, Cursor) error
}

type Resumer struct {
	store CheckpointStore
}

func NewResumer(store CheckpointStore) *Resumer {
	return &Resumer{store: store}
}

func (r *Resumer) Resume(ctx context.Context, token string) (Cursor, error) {
	next, err := DecodeCursor(token)
	if err != nil {
		return Cursor{}, err
	}
	current, err := r.store.Load(ctx, next.Stream)
	if err != nil {
		return Cursor{}, err
	}
	if err := r.store.Save(ctx, next); err != nil {
		return Cursor{}, err
	}
	if err := ValidateAdvance(current, next); err != nil {
		return Cursor{}, err
	}
	return next, nil
}
