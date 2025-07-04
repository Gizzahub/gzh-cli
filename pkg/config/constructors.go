package config

import (
	"context"
	"fmt"
	"io"
	"time"
)

// FileSystemInterface for dependency injection
type FileSystemInterface interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm int) error
	Exists(path string) bool
	Stat(path string) (FileInfo, error)
	MkdirAll(path string, perm int) error
}

// FileInfo interface for file information
type FileInfo interface {
	IsDir() bool
	ModTime() time.Time
	Size() int64
}

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// ConfigLoaderImpl implements the ConfigLoader interface
type ConfigLoaderImpl struct {
	fileSystem  FileSystemInterface
	parser      ConfigParser
	validator   ConfigValidator
	logger      Logger
	searchPaths []string
}

// ConfigLoaderConfig holds configuration for the config loader
type ConfigLoaderConfig struct {
	SearchPaths    []string
	EnableCache    bool
	CacheTTL       time.Duration
	ValidateOnLoad bool
}

// DefaultConfigLoaderConfig returns default configuration
func DefaultConfigLoaderConfig() *ConfigLoaderConfig {
	return &ConfigLoaderConfig{
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

// NewConfigLoader creates a new config loader with dependencies
func NewConfigLoader(
	config *ConfigLoaderConfig,
	fileSystem FileSystemInterface,
	parser ConfigParser,
	validator ConfigValidator,
	logger Logger,
) ConfigLoader {
	if config == nil {
		config = DefaultConfigLoaderConfig()
	}

	return &ConfigLoaderImpl{
		fileSystem:  fileSystem,
		parser:      parser,
		validator:   validator,
		logger:      logger,
		searchPaths: config.SearchPaths,
	}
}

// LoadConfig implements ConfigLoader interface
func (l *ConfigLoaderImpl) LoadConfig(ctx context.Context) (*Config, error) {
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

// LoadConfigFromFile implements ConfigLoader interface
func (l *ConfigLoaderImpl) LoadConfigFromFile(ctx context.Context, filename string) (*Config, error) {
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

// LoadConfigFromReader implements ConfigLoader interface
func (l *ConfigLoaderImpl) LoadConfigFromReader(ctx context.Context, reader io.Reader) (*Config, error) {
	l.logger.Debug("Loading configuration from reader")

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return l.parser.ParseConfig(ctx, data)
}

// GetSearchPaths implements ConfigLoader interface
func (l *ConfigLoaderImpl) GetSearchPaths() []string {
	return l.searchPaths
}

// SetSearchPaths implements ConfigLoader interface
func (l *ConfigLoaderImpl) SetSearchPaths(paths []string) {
	l.searchPaths = paths
}

// expandPath expands environment variables and home directory
func (l *ConfigLoaderImpl) expandPath(path string) string {
	// Implementation would expand ~ and environment variables
	return path
}

// ConfigValidatorImpl implements the ConfigValidator interface
type ConfigValidatorImpl struct {
	schemaValidator SchemaValidator
	logger          Logger
}

// NewConfigValidator creates a new config validator with dependencies
func NewConfigValidator(schemaValidator SchemaValidator, logger Logger) ConfigValidator {
	return &ConfigValidatorImpl{
		schemaValidator: schemaValidator,
		logger:          logger,
	}
}

// ValidateConfig implements ConfigValidator interface
func (v *ConfigValidatorImpl) ValidateConfig(ctx context.Context, config *Config) error {
	v.logger.Debug("Validating configuration")

	return v.schemaValidator.ValidateStructure(ctx, config)
}

// ValidateConfigFile implements ConfigValidator interface
func (v *ConfigValidatorImpl) ValidateConfigFile(ctx context.Context, filename string) error {
	v.logger.Debug("Validating configuration file", "file", filename)

	// Implementation would load and validate file
	return nil
}

// GetValidationErrors implements ConfigValidator interface
func (v *ConfigValidatorImpl) GetValidationErrors(ctx context.Context, config *Config) []ValidationError {
	v.logger.Debug("Getting validation errors")

	// Implementation would return detailed validation errors
	return nil
}

// IsValid implements ConfigValidator interface
func (v *ConfigValidatorImpl) IsValid(ctx context.Context, config *Config) bool {
	return v.ValidateConfig(ctx, config) == nil
}

// ConfigParserImpl implements the ConfigParser interface
type ConfigParserImpl struct {
	logger Logger
}

// NewConfigParser creates a new config parser with dependencies
func NewConfigParser(logger Logger) ConfigParser {
	return &ConfigParserImpl{
		logger: logger,
	}
}

// ParseConfig implements ConfigParser interface
func (p *ConfigParserImpl) ParseConfig(ctx context.Context, data []byte) (*Config, error) {
	p.logger.Debug("Parsing configuration data")

	// Implementation would parse YAML/JSON data
	return &Config{}, nil
}

// ParseConfigWithFormat implements ConfigParser interface
func (p *ConfigParserImpl) ParseConfigWithFormat(ctx context.Context, data []byte, format string) (*Config, error) {
	p.logger.Debug("Parsing configuration with format", "format", format)

	// Implementation would parse based on format
	return &Config{}, nil
}

// GetSupportedFormats implements ConfigParser interface
func (p *ConfigParserImpl) GetSupportedFormats() []string {
	return []string{"yaml", "yml", "json"}
}

// IsFormatSupported implements ConfigParser interface
func (p *ConfigParserImpl) IsFormatSupported(format string) bool {
	for _, supported := range p.GetSupportedFormats() {
		if format == supported {
			return true
		}
	}
	return false
}

// ProviderManagerImpl implements the ProviderManager interface
type ProviderManagerImpl struct {
	config *Config
	logger Logger
}

// NewProviderManager creates a new provider manager with dependencies
func NewProviderManager(config *Config, logger Logger) ProviderManager {
	return &ProviderManagerImpl{
		config: config,
		logger: logger,
	}
}

// GetProviders implements ProviderManager interface
func (m *ProviderManagerImpl) GetProviders(ctx context.Context) (map[string]Provider, error) {
	m.logger.Debug("Getting all providers")

	return m.config.Providers, nil
}

// GetProvider implements ProviderManager interface
func (m *ProviderManagerImpl) GetProvider(ctx context.Context, name string) (*Provider, error) {
	m.logger.Debug("Getting provider", "name", name)

	provider, exists := m.config.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return &provider, nil
}

// CreateProviderCloner implements ProviderManager interface
func (m *ProviderManagerImpl) CreateProviderCloner(ctx context.Context, providerName, token string) (ProviderCloner, error) {
	m.logger.Debug("Creating provider cloner", "provider", providerName)

	return CreateProviderCloner(providerName, token)
}

// ValidateProvider implements ProviderManager interface
func (m *ProviderManagerImpl) ValidateProvider(ctx context.Context, provider *Provider) error {
	m.logger.Debug("Validating provider", "name", provider.Name)

	// Implementation would validate provider configuration
	return nil
}

// GetSupportedProviders implements ProviderManager interface
func (m *ProviderManagerImpl) GetSupportedProviders() []string {
	return []string{"github", "gitlab", "gitea"}
}

// ConfigServiceImpl implements the unified ConfigService interface
type ConfigServiceImpl struct {
	ConfigLoader
	ConfigValidator
	ConfigParser
	SchemaValidator
	ProviderManager
	DirectoryResolver
	FilterService
	IntegrationService
}

// ConfigServiceConfig holds configuration for the config service
type ConfigServiceConfig struct {
	Loader        *ConfigLoaderConfig
	CacheSize     int
	EnableMetrics bool
}

// DefaultConfigServiceConfig returns default configuration
func DefaultConfigServiceConfig() *ConfigServiceConfig {
	return &ConfigServiceConfig{
		Loader:        DefaultConfigLoaderConfig(),
		CacheSize:     100,
		EnableMetrics: true,
	}
}

// NewConfigService creates a new config service with all dependencies
func NewConfigService(
	config *ConfigServiceConfig,
	fileSystem FileSystemInterface,
	logger Logger,
) ConfigService {
	if config == nil {
		config = DefaultConfigServiceConfig()
	}

	parser := NewConfigParser(logger)

	// Schema validator would be created with its own dependencies
	var schemaValidator SchemaValidator

	validator := NewConfigValidator(schemaValidator, logger)
	loader := NewConfigLoader(config.Loader, fileSystem, parser, validator, logger)

	// Load the actual config
	configData, err := loader.LoadConfig(context.Background())
	if err != nil {
		logger.Warn("Failed to load config during service creation", "error", err)
		configData = &Config{} // Use empty config as fallback
	}

	providerManager := NewProviderManager(configData, logger)

	// Other services would be created similarly
	var directoryResolver DirectoryResolver
	var filterService FilterService
	var integrationService IntegrationService

	return &ConfigServiceImpl{
		ConfigLoader:       loader,
		ConfigValidator:    validator,
		ConfigParser:       parser,
		SchemaValidator:    schemaValidator,
		ProviderManager:    providerManager,
		DirectoryResolver:  directoryResolver,
		FilterService:      filterService,
		IntegrationService: integrationService,
	}
}

// ServiceDependencies holds all the dependencies needed for config services
type ServiceDependencies struct {
	FileSystem FileSystemInterface
	Logger     Logger
}

// NewServiceDependencies creates a default set of service dependencies
func NewServiceDependencies(fileSystem FileSystemInterface, logger Logger) *ServiceDependencies {
	return &ServiceDependencies{
		FileSystem: fileSystem,
		Logger:     logger,
	}
}
