package decoderregistry

import (
	"context"
	"errors"
	"reflect"
)

var ErrProviderUnavailable = errors.New("decoder provider unavailable")

func ValidateProvider(ctx context.Context, provider Provider, driver string) error {
	if nilProvider(provider) {
		return ErrProviderUnavailable
	}
	return provider.Validate(ctx, driver)
}

func nilProvider(provider Provider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
