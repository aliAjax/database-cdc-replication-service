package decoderregistry

import (
	"context"
	"fmt"
)

type Provider interface {
	Validate(context.Context, string) error
}

type StaticProvider struct {
	Allowed map[string]bool
}

func (p *StaticProvider) Validate(_ context.Context, driver string) error {
	if !p.Allowed[driver] {
		return fmt.Errorf("decoder driver %q is unavailable", driver)
	}
	return nil
}

func NewProvider(enabled bool, drivers []string) Provider {
	if !enabled {
		var provider *StaticProvider
		return provider
	}
	allowed := make(map[string]bool, len(drivers))
	for _, driver := range drivers {
		allowed[driver] = true
	}
	return &StaticProvider{Allowed: allowed}
}
