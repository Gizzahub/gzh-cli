package cloud

import (
	"context"
	"fmt"
)

// ProviderType represents supported cloud provider types
type ProviderType string

const (
	// ProviderTypeAWS represents Amazon Web Services
	ProviderTypeAWS ProviderType = "aws"

	// ProviderTypeGCP represents Google Cloud Platform
	ProviderTypeGCP ProviderType = "gcp"

	// ProviderTypeAzure represents Microsoft Azure
	ProviderTypeAzure ProviderType = "azure"
)

// ProviderFactory is a factory function for creating providers
type ProviderFactory func(ctx context.Context, config ProviderConfig) (Provider, error)

// Registry holds registered provider factories
type Registry struct {
	factories map[ProviderType]ProviderFactory
}

// globalRegistry is the global provider registry
var globalRegistry = &Registry{
	factories: make(map[ProviderType]ProviderFactory),
}

// Register registers a provider factory
func Register(providerType ProviderType, factory ProviderFactory) {
	globalRegistry.factories[providerType] = factory
}

// NewProvider creates a new provider instance
func NewProvider(ctx context.Context, config ProviderConfig) (Provider, error) {
	providerType := ProviderType(config.Type)
	factory, exists := globalRegistry.factories[providerType]
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}

	return factory(ctx, config)
}

// GetSupportedProviders returns list of supported provider types
func GetSupportedProviders() []ProviderType {
	providers := make([]ProviderType, 0, len(globalRegistry.factories))
	for p := range globalRegistry.factories {
		providers = append(providers, p)
	}
	return providers
}

// IsProviderSupported checks if a provider type is supported
func IsProviderSupported(providerType string) bool {
	_, exists := globalRegistry.factories[ProviderType(providerType)]
	return exists
}

// GetRegisteredProviders returns list of registered provider type names
func GetRegisteredProviders() []string {
	providers := make([]string, 0, len(globalRegistry.factories))
	for p := range globalRegistry.factories {
		providers = append(providers, string(p))
	}
	return providers
}
