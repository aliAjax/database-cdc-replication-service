package callbackregistry

import (
	"errors"
	"fmt"
)

var ErrUnknownEvent = errors.New("unknown event")
var ErrHandlerUnavailable = errors.New("handler unavailable")

type Handler func(string) error
type Registry struct{ handlers map[string]Handler }

func NewRegistry() *Registry { return &Registry{handlers: map[string]Handler{}} }
func (r *Registry) Register(name string, h Handler) error {
	if h == nil {
		return ErrHandlerUnavailable
	}
	r.handlers[name] = h
	return nil
}
func (r *Registry) Dispatch(name, payload string) error {
	h, ok := r.handlers[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownEvent, name)
	}
	if h == nil {
		return ErrHandlerUnavailable
	}
	return h(payload)
}
