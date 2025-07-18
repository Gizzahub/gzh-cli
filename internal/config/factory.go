package config

import (
	"fmt"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// ServiceFactory creates configuration service instances.
type ServiceFactory interface {
	// CreateConfigService creates a new configuration service instance
	CreateConfigService(options *ConfigServiceOptions) (ConfigService, error)

	// CreateDefaultConfigService creates a configuration service with default options
	CreateDefaultConfigService() (ConfigService, error)

	// CreateConfigServiceWithEnvironment creates a configuration service with custom environment
	CreateConfigServiceWithEnvironment(environment env.Environment) (ConfigService, error)
}

// DefaultServiceFactory implements ServiceFactory.
type DefaultServiceFactory struct{}

// NewServiceFactory creates a new configuration service factory.
func NewServiceFactory() ServiceFactory {
	return &DefaultServiceFactory{}
}

// CreateConfigService creates a new configuration service instance.
func (f *DefaultServiceFactory) CreateConfigService(options *ConfigServiceOptions) (ConfigService, error) {
	if options == nil {
		return nil, fmt.Errorf("configuration service options cannot be nil")
	}

	service, err := NewConfigService(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration service: %w", err)
	}

	return service, nil
}

// CreateDefaultConfigService creates a configuration service with default options.
func (f *DefaultServiceFactory) CreateDefaultConfigService() (ConfigService, error) {
	options := DefaultConfigServiceOptions()
	return f.CreateConfigService(options)
}

// CreateConfigServiceWithEnvironment creates a configuration service with custom environment.
func (f *DefaultServiceFactory) CreateConfigServiceWithEnvironment(environment env.Environment) (ConfigService, error) {
	options := DefaultConfigServiceOptions()
	options.Environment = environment

	return f.CreateConfigService(options)
}

// Global factory instance for convenience.
var globalFactory ServiceFactory = NewServiceFactory()

// CreateConfigService creates a configuration service using the global factory.
func CreateConfigService(options *ConfigServiceOptions) (ConfigService, error) {
	return globalFactory.CreateConfigService(options)
}

// CreateDefaultConfigService creates a configuration service with default options using the global factory.
func CreateDefaultConfigService() (ConfigService, error) {
	return globalFactory.CreateDefaultConfigService()
}

// CreateConfigServiceWithEnvironment creates a configuration service with custom environment using the global factory.
func CreateConfigServiceWithEnvironment(environment env.Environment) (ConfigService, error) {
	return globalFactory.CreateConfigServiceWithEnvironment(environment)
}

// SetGlobalFactory sets the global factory instance (useful for testing).
func SetGlobalFactory(factory ServiceFactory) {
	globalFactory = factory
}
