package decoderregistry

import (
	"context"
	"errors"
)

var ErrProviderUnavailable = errors.New("decoder provider unavailable")

func ValidateProvider(ctx context.Context, provider Provider, driver string) error {
	if provider == nil {
		return ErrProviderUnavailable
	}
	return provider.Validate(ctx, driver)
}
