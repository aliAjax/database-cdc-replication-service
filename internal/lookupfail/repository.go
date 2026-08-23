package lookupfail

import (
	"errors"
	"fmt"
)

var ErrStreamMissing = errors.New("stream missing")

type Repository struct{ streams map[string]string }

func NewRepository(streams map[string]string) *Repository { return &Repository{streams: streams} }

func (r *Repository) Load(id string) (string, error) {
	value, ok := r.streams[id]
	if !ok {
		return "", fmt.Errorf("load stream %q: %w", id, ErrStreamMissing)
	}
	return value, nil
}
