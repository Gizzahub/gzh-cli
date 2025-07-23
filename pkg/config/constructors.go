// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"io"
	"time"
)

// FileSystemInterface for dependency injection.
type FileSystemInterface interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm int) error
	Exists(path string) bool
	Stat(path string) (FileInfo, error)
	MkdirAll(path string, perm int) error
}

// FileInfo interface for file information.
type FileInfo interface {
	IsDir() bool
	ModTime() time.Time
	Size() int64
}

// Logger interface for dependency injection.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// configLoaderImpl implements the Loader interface.
type configLoaderImpl struct {
	fileSystem  FileSystemInterface
	parser      Parser
	validator   Validator
	logger      Logger
	searchPaths []string
}

// LoaderConfig holds configuration for the config loader.
type LoaderConfig struct {
	SearchPaths    []string
	EnableCache    bool
	CacheTTL       time.Duration
	ValidateOnLoad bool
}

// DefaultLoaderConfig returns default configuration.
func DefaultLoaderConfig() *LoaderConfig {
	return &LoaderConfig{
		SearchPaths: []string{
			"./gzh.yaml",
			"./gzh.yml",
			"~/.config/gzh.yaml",
			"~/.config/gzh.yml",
			"~/.config/gzh-manager/gzh.yaml",
			"~/.config/gzh-manager/gzh.yml",
			"/etc/gzh-manager/gzh.yaml",
			"/etc/gzh-manager/gzh.yml",
		},
		EnableCache:    true,
		CacheTTL:       5 * time.Minute,
		ValidateOnLoad: true,
	}
}

// NewLoader creates a new config loader with dependencies.
func NewLoader(
	config *LoaderConfig,
	fileSystem FileSystemInterface,
	parser Parser,
	validator Validator,
	logger Logger,
) Loader {
	if config == nil {
		config = DefaultLoaderConfig()
	}

	return &configLoaderImpl{
		fileSystem:  fileSystem,
		parser:      parser,
		validator:   validator,
		logger:      logger,
		searchPaths: config.SearchPaths,
	}
}

// LoadConfig implements Loader interface.
func (l *configLoaderImpl) LoadConfig(ctx context.Context) (*Config, error) {
	l.logger.Debug("Loading configuration from search paths")

	for _, path := range l.searchPaths {
		expandedPath := l.expandPath(path)
		if l.fileSystem.Exists(expandedPath) {
			l.logger.Debug("Found config file", "path", expandedPath)
			return l.LoadConfigFromFile(ctx, expandedPath)
		}
	}

	return nil, fmt.Errorf("no configuration file found in search paths")
}

// LoadConfigFromFile implements Loader interface.
func (l *configLoaderImpl) LoadConfigFromFile(ctx context.Context, filename string) (*Config, error) {
	l.logger.Debug("Loading configuration from file", "file", filename)

	data, err := l.fileSystem.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	config, err := l.parser.ParseConfig(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	if err := l.validator.ValidateConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("config validation failed for %s: %w", filename, err)
	}

	return config, nil
}

// LoadConfigFromReader implements Loader interface.
func (l *configLoaderImpl) LoadConfigFromReader(ctx context.Context, reader io.Reader) (*Config, error) {
	l.logger.Debug("Loading configuration from reader")

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return l.parser.ParseConfig(ctx, data)
}

// GetSearchPaths implements Loader interface.
func (l *configLoaderImpl) GetSearchPaths() []string {
	return l.searchPaths
}

// SetSearchPaths implements Loader interface.
func (l *configLoaderImpl) SetSearchPaths(paths []string) {
	l.searchPaths = paths
}

// expandPath expands environment variables and home directory.
func (l *configLoaderImpl) expandPath(path string) string {
	// Implementation would expand ~ and environment variables
	return path
}

// configValidatorImpl implements the Validator interface.
type configValidatorImpl struct {
	schemaValidator SchemaValidator
	logger          Logger
}

// NewValidator creates a new config validator with dependencies.
func NewValidator(schemaValidator SchemaValidator, logger Logger) Validator {
	return &configValidatorImpl{
		schemaValidator: schemaValidator,
		logger:          logger,
	}
}

// ValidateConfig implements Validator interface.
func (v *configValidatorImpl) ValidateConfig(ctx context.Context, config *Config) error {
	v.logger.Debug("Validating configuration")

	return v.schemaValidator.ValidateStructure(ctx, config)
}

// ValidateConfigFile implements Validator interface.
func (v *configValidatorImpl) ValidateConfigFile(_ context.Context, filename string) error {
	v.logger.Debug("Validating configuration file", "file", filename)

	// Implementation would load and validate file
	return nil
}

// GetValidationErrors implements Validator interface.
func (v *configValidatorImpl) GetValidationErrors(_ context.Context, _ *Config) []ValidationError {
	v.logger.Debug("Getting validation errors")

	// Implementation would return detailed validation errors
	return nil
}

// IsValid implements Validator interface.
func (v *configValidatorImpl) IsValid(ctx context.Context, config *Config) bool {
	return v.ValidateConfig(ctx, config) == nil
}

// configParserImpl implements the Parser interface.
type configParserImpl struct {
	logger Logger
}

// NewParser creates a new config parser with dependencies.
func NewParser(logger Logger) Parser {
	return &configParserImpl{
		logger: logger,
	}
}

// ParseConfig implements Parser interface.
func (p *configParserImpl) ParseConfig(_ context.Context, _ []byte) (*Config, error) {
	p.logger.Debug("Parsing configuration data")

	// Implementation would parse YAML/JSON data
	return &Config{}, nil
}

// ParseConfigWithFormat implements Parser interface.
func (p *configParserImpl) ParseConfigWithFormat(_ context.Context, _ []byte, format string) (*Config, error) {
	p.logger.Debug("Parsing configuration with format", "format", format)

	// Implementation would parse based on format
	return &Config{}, nil
}

// GetSupportedFormats implements Parser interface.
func (p *configParserImpl) GetSupportedFormats() []string {
	return []string{"yaml", "yml", "json"}
}

// IsFormatSupported implements Parser interface.
func (p *configParserImpl) IsFormatSupported(format string) bool {
	for _, supported := range p.GetSupportedFormats() {
		if format == supported {
			return true
		}
	}

	return false
}

// providerManagerImpl implements the ProviderManager interface.
type providerManagerImpl struct {
	config *Config
	logger Logger
}

// NewProviderManager creates a new provider manager with dependencies.
func NewProviderManager(config *Config, logger Logger) ProviderManager {
	return &providerManagerImpl{
		config: config,
		logger: logger,
	}
}

// GetProviders implements ProviderManager interface.
func (m *providerManagerImpl) GetProviders(_ context.Context) (map[string]Provider, error) {
	m.logger.Debug("Getting all providers")

	return m.config.Providers, nil
}

// GetProvider implements ProviderManager interface.
func (m *providerManagerImpl) GetProvider(_ context.Context, name string) (*Provider, error) {
	m.logger.Debug("Getting provider", "name", name)

	provider, exists := m.config.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return &provider, nil
}

// CreateProviderCloner implements ProviderManager interface.
func (m *providerManagerImpl) CreateProviderCloner(ctx context.Context, providerName, token string) (ProviderCloner, error) {
	m.logger.Debug("Creating provider cloner", "provider", providerName)

	// Use factory pattern for provider creation
	factory := NewProviderFactory(nil, m.logger)

	return factory.CreateCloner(ctx, providerName, token)
}

// ValidateProvider implements ProviderManager interface.
func (m *providerManagerImpl) ValidateProvider(_ context.Context, _ *Provider) error {
	m.logger.Debug("Validating provider")

	// Implementation would validate provider configuration
	return nil
}

// GetSupportedProviders implements ProviderManager interface.
func (m *providerManagerImpl) GetSupportedProviders() []string {
	return []string{"github", "gitlab", "gitea"}
}

// configServiceImpl implements the unified Service interface.
type configServiceImpl struct {
	Loader
	Validator
	Parser
	SchemaValidator
	ProviderManager
	DirectoryResolverInterface
	FilterService
	IntegrationService
}

// ServiceConfig holds configuration for the config service.
type ServiceConfig struct {
	Loader        *LoaderConfig
	CacheSize     int
	EnableMetrics bool
}

// DefaultServiceConfig returns default configuration.
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		Loader:        DefaultLoaderConfig(),
		CacheSize:     100,
		EnableMetrics: true,
	}
}

// NewService creates a new config service with all dependencies.
func NewService(
	config *ServiceConfig,
	fileSystem FileSystemInterface,
	logger Logger,
) Service {
	if config == nil {
		config = DefaultServiceConfig()
	}

	parser := NewParser(logger)

	// Schema validator would be created with its own dependencies
	var schemaValidator SchemaValidator

	validator := NewValidator(schemaValidator, logger)
	loader := NewLoader(config.Loader, fileSystem, parser, validator, logger)

	// Load the actual config
	configData, err := loader.LoadConfig(context.Background())
	if err != nil {
		logger.Warn("Failed to load config during service creation", "error", err)

		configData = &Config{} // Use empty config as fallback
	}

	providerManager := NewProviderManager(configData, logger)

	// Other services would be created similarly
	var (
		directoryResolver  DirectoryResolverInterface
		filterService      FilterService
		integrationService IntegrationService
	)

	return &configServiceImpl{
		Loader:                     loader,
		Validator:                  validator,
		Parser:                     parser,
		SchemaValidator:            schemaValidator,
		ProviderManager:            providerManager,
		DirectoryResolverInterface: directoryResolver,
		FilterService:              filterService,
		IntegrationService:         integrationService,
	}
}

// ServiceDependencies holds all the dependencies needed for config services.
type ServiceDependencies struct {
	FileSystem FileSystemInterface
	Logger     Logger
}

// NewServiceDependencies creates a default set of service dependencies.
func NewServiceDependencies(fileSystem FileSystemInterface, logger Logger) *ServiceDependencies {
	return &ServiceDependencies{
		FileSystem: fileSystem,
		Logger:     logger,
	}
}
