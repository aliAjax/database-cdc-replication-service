package decoderregistry

import (
	"context"
	"errors"
	"testing"
)

func TestDecoderRegistryDefaultCatalogIsWritable(t *testing.T) {
	catalog := DefaultCatalog()
	if catalog.Drivers == nil {
		t.Fatal("default catalog has a nil driver map")
	}
	catalog.Drivers["postgres"] = DriverConfig{Endpoint: "db:5432", Enabled: true}
}

func TestDecoderRegistryDisabledProviderIsNil(t *testing.T) {
	provider := NewProvider(false, []string{"postgres"})
	if provider != nil {
		t.Fatalf("disabled provider is a non-nil interface: %T", provider)
	}
}

func TestDecoderRegistryZeroValueRegisterIsSafe(t *testing.T) {
	var registry Registry
	registry.Register("postgres", DriverConfig{Endpoint: "db:5432", Enabled: true})
	if _, ok := registry.Snapshot()["postgres"]; !ok {
		t.Fatal("zero-value registry did not retain the registered driver")
	}
}

func TestDecoderRegistryTypedNilProviderIsRejected(t *testing.T) {
	var static *StaticProvider
	var provider Provider = static
	err := ValidateProvider(context.Background(), provider, "postgres")
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("typed-nil provider error = %v, want ErrProviderUnavailable", err)
	}
}
