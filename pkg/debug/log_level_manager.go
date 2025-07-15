package debug

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogLevelRule represents a rule for conditional logging that allows
// dynamic adjustment of log levels based on various conditions.
//
// Rules can be configured to respond to:
//   - Log message content and metadata
//   - System performance metrics (CPU, memory)
//   - Module or component context
//   - Time-based conditions
//   - Frequency patterns
//
// Each rule has a priority level for conflict resolution and tracks
// usage statistics for monitoring and optimization.
type LogLevelRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Conditions  []LogCondition `json:"conditions"`
	Actions     []LogAction    `json:"actions"`
	Priority    int            `json:"priority"` // Higher priority rules are evaluated first
	Created     time.Time      `json:"created"`
	LastApplied *time.Time     `json:"last_applied,omitempty"`
	ApplyCount  int64          `json:"apply_count"`
}

// LogCondition represents a condition that must be met for a log rule to apply.
// Conditions support various types of evaluations including field comparisons,
// pattern matching, and system state checks.
//
// Supported condition types:
//   - "level": Log severity level
//   - "module": Module or component name
//   - "field": Custom field in log entry
//   - "message": Log message content
//   - "time": Time-based conditions
//   - "frequency": Log frequency patterns
//   - "cpu": CPU usage percentage
//   - "memory": Memory usage percentage
//
// Operators include: eq, ne, gt, lt, gte, lte, contains, regex, in
type LogCondition struct {
	Type     string         `json:"type"`     // "level", "module", "field", "message", "time", "frequency", "cpu", "memory"
	Field    string         `json:"field"`    // Field name for "field" type
	Operator string         `json:"operator"` // "eq", "ne", "gt", "lt", "gte", "lte", "contains", "regex", "in"
	Value    interface{}    `json:"value"`    // Expected value
	Regex    *regexp.Regexp `json:"-"`        // Compiled regex for "regex" operator
}

// LogAction represents an action to execute when rule conditions are satisfied.
// Actions can modify logging behavior dynamically, including level changes,
// sampling adjustments, and routing decisions.
//
// Supported action types:
//   - "set_level": Change log level for target scope
//   - "sample": Apply sampling rate to reduce log volume
//   - "drop": Suppress log entries matching conditions
//   - "redirect": Route logs to different outputs
//   - "throttle": Limit log rate to prevent flooding
//
// Actions can be temporary (with duration) or permanent until rule removal.
type LogAction struct {
	Type     string         `json:"type"`               // "set_level", "sample", "drop", "redirect", "throttle"
	Target   string         `json:"target"`             // Module name or "global"
	Value    interface{}    `json:"value"`              // Action parameter
	Duration *time.Duration `json:"duration,omitempty"` // Duration for temporary actions
}

// LogLevelProfile represents a predefined set of log level configurations
// for different operational scenarios (development, testing, production).
//
// Profiles provide:
//   - Global and module-specific log levels
//   - Pre-configured rule sets for common patterns
//   - Sampling configurations optimized for the environment
//   - Easy switching between operational modes
//
// Example profiles:
//   - "development": Verbose logging with minimal sampling
//   - "testing": Structured logging with error focus
//   - "production": Optimized logging with aggressive sampling
type LogLevelProfile struct {
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	GlobalLevel  RFC5424Severity            `json:"global_level"`
	ModuleLevels map[string]RFC5424Severity `json:"module_levels"`
	Rules        []LogLevelRule             `json:"rules"`
	Sampling     SamplingConfig             `json:"sampling"`
	Created      time.Time                  `json:"created"`
}

// SamplingConfig represents sampling configuration
type SamplingConfig struct {
	Enabled          bool                        `json:"enabled"`
	DefaultRate      float64                     `json:"default_rate"`
	LevelRates       map[RFC5424Severity]float64 `json:"level_rates"`
	ModuleRates      map[string]float64          `json:"module_rates"`
	AdaptiveEnabled  bool                        `json:"adaptive_enabled"`
	CPUThresholds    map[string]float64          `json:"cpu_thresholds"`    // CPU % -> sample rate (string keys for JSON)
	MemoryThresholds map[string]float64          `json:"memory_thresholds"` // Memory % -> sample rate (string keys for JSON)
}

// SystemMetrics represents current system performance metrics
type SystemMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	LogRate     float64   `json:"log_rate"` // logs per second
	Timestamp   time.Time `json:"timestamp"`
}

// LogLevelManager provides advanced log level management with dynamic
// rule evaluation, profile switching, and performance-aware sampling.
//
// Key features:
//   - Dynamic rule evaluation based on conditions
//   - Profile-based configuration management
//   - HTTP API for runtime control and monitoring
//   - Signal-based configuration reloading
//   - Adaptive sampling based on system performance
//   - Performance metrics collection and analysis
//
// The manager integrates with StructuredLogger to provide real-time
// log level adjustments and optimization for different operational scenarios.
//
// Example usage:
//
//	manager, err := NewLogLevelManager(config, logger)
//	if err != nil {
//	    return err
//	}
//	defer manager.Shutdown()
//
//	// Apply production profile
//	manager.ApplyProfile("production")
//
//	// Start HTTP API
//	manager.StartHTTPServer(8080)
type LogLevelManager struct {
	logger *StructuredLogger
	mutex  sync.RWMutex

	// Configuration
	profiles       map[string]*LogLevelProfile
	rules          []LogLevelRule
	currentProfile *LogLevelProfile

	// Dynamic control
	httpServer *http.Server
	signalChan chan os.Signal

	// Performance monitoring
	metrics      SystemMetrics
	metricsMutex sync.RWMutex

	// Adaptive sampling
	adaptiveEnabled bool
	lastCPUCheck    time.Time
	logCounter      int64
	lastLogCount    int64

	// Rule evaluation cache
	ruleCache    map[string]bool
	cacheMutex   sync.RWMutex
	cacheTimeout time.Duration
}

// LogLevelManagerConfig holds log level manager configuration
type LogLevelManagerConfig struct {
	EnableHTTPControl   bool          `json:"enable_http_control"`
	HTTPPort            int           `json:"http_port"`
	EnableSignalControl bool          `json:"enable_signal_control"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
	CacheTimeout        time.Duration `json:"cache_timeout"`
	DefaultProfile      string        `json:"default_profile"`
}

// DefaultLogLevelManagerConfig returns default configuration
func DefaultLogLevelManagerConfig() *LogLevelManagerConfig {
	return &LogLevelManagerConfig{
		EnableHTTPControl:   true,
		HTTPPort:            8080,
		EnableSignalControl: true,
		MetricsInterval:     time.Second * 30,
		CacheTimeout:        time.Minute * 5,
		DefaultProfile:      "development",
	}
}

// NewLogLevelManager creates a new log level manager
func NewLogLevelManager(logger *StructuredLogger, config *LogLevelManagerConfig) (*LogLevelManager, error) {
	if config == nil {
		config = DefaultLogLevelManagerConfig()
	}

	manager := &LogLevelManager{
		logger:          logger,
		profiles:        make(map[string]*LogLevelProfile),
		rules:           make([]LogLevelRule, 0),
		ruleCache:       make(map[string]bool),
		cacheTimeout:    config.CacheTimeout,
		adaptiveEnabled: false,
	}

	// Initialize built-in profiles
	manager.initBuiltinProfiles()

	// Set default profile
	if profile, exists := manager.profiles[config.DefaultProfile]; exists {
		manager.ApplyProfile(config.DefaultProfile)
		manager.currentProfile = profile
	}

	// Start HTTP server if enabled
	if config.EnableHTTPControl {
		if err := manager.startHTTPServer(config.HTTPPort); err != nil {
			return nil, fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}

	// Setup signal handling if enabled
	if config.EnableSignalControl {
		manager.setupSignalHandling()
	}

	// Start metrics collection if interval is positive
	if config.MetricsInterval > 0 {
		go manager.collectMetrics(config.MetricsInterval)
	}

	return manager, nil
}

// initBuiltinProfiles initializes predefined log level profiles
func (lm *LogLevelManager) initBuiltinProfiles() {
	// Development Profile
	lm.profiles["development"] = &LogLevelProfile{
		Name:        "development",
		Description: "Development environment with debug logging",
		GlobalLevel: SeverityDebug,
		ModuleLevels: map[string]RFC5424Severity{
			"auth":     SeverityDebug,
			"api":      SeverityDebug,
			"database": SeverityInfo,
		},
		Rules: []LogLevelRule{
			{
				ID:          "dev-auth-debug",
				Name:        "Auth Debug Logging",
				Description: "Enable debug logging for authentication modules",
				Enabled:     true,
				Priority:    10,
				Conditions: []LogCondition{
					{Type: "module", Operator: "contains", Value: "auth"},
				},
				Actions: []LogAction{
					{Type: "set_level", Target: "module", Value: SeverityDebug},
				},
				Created: time.Now(),
			},
		},
		Sampling: SamplingConfig{
			Enabled:         false,
			DefaultRate:     1.0,
			AdaptiveEnabled: false,
		},
		Created: time.Now(),
	}

	// Production Profile
	lm.profiles["production"] = &LogLevelProfile{
		Name:        "production",
		Description: "Production environment with optimized logging",
		GlobalLevel: SeverityWarning,
		ModuleLevels: map[string]RFC5424Severity{
			"auth":     SeverityInfo,
			"api":      SeverityWarning,
			"database": SeverityError,
		},
		Rules: []LogLevelRule{
			{
				ID:          "prod-cpu-throttle",
				Name:        "CPU-based Log Throttling",
				Description: "Reduce logging when CPU usage is high",
				Enabled:     true,
				Priority:    100,
				Conditions: []LogCondition{
					{Type: "cpu", Operator: "gt", Value: 80.0},
				},
				Actions: []LogAction{
					{Type: "sample", Target: "global", Value: 0.1}, // 10% sampling
					{Type: "set_level", Target: "global", Value: SeverityError},
				},
				Created: time.Now(),
			},
		},
		Sampling: SamplingConfig{
			Enabled:     true,
			DefaultRate: 0.5,
			LevelRates: map[RFC5424Severity]float64{
				SeverityEmergency: 1.0,
				SeverityAlert:     1.0,
				SeverityCritical:  1.0,
				SeverityError:     0.8,
				SeverityWarning:   0.3,
				SeverityInfo:      0.1,
				SeverityDebug:     0.01,
			},
			AdaptiveEnabled: true,
			CPUThresholds: map[string]float64{
				"50.0": 1.0,
				"70.0": 0.5,
				"80.0": 0.2,
				"90.0": 0.05,
			},
			MemoryThresholds: map[string]float64{
				"60.0": 1.0,
				"75.0": 0.5,
				"85.0": 0.2,
				"95.0": 0.05,
			},
		},
		Created: time.Now(),
	}

	// Testing Profile
	lm.profiles["testing"] = &LogLevelProfile{
		Name:        "testing",
		Description: "Testing environment with comprehensive logging",
		GlobalLevel: SeverityInfo,
		ModuleLevels: map[string]RFC5424Severity{
			"test":     SeverityDebug,
			"auth":     SeverityInfo,
			"api":      SeverityInfo,
			"database": SeverityWarning,
		},
		Rules: []LogLevelRule{
			{
				ID:          "test-error-verbose",
				Name:        "Verbose Error Logging",
				Description: "Enable verbose logging for error conditions in tests",
				Enabled:     true,
				Priority:    50,
				Conditions: []LogCondition{
					{Type: "level", Operator: "lte", Value: SeverityError},
					{Type: "module", Operator: "contains", Value: "test"},
				},
				Actions: []LogAction{
					{Type: "set_level", Target: "module", Value: SeverityDebug},
				},
				Created: time.Now(),
			},
		},
		Sampling: SamplingConfig{
			Enabled:         false,
			DefaultRate:     1.0,
			AdaptiveEnabled: false,
		},
		Created: time.Now(),
	}
}

// ApplyProfile applies a predefined log level profile
func (lm *LogLevelManager) ApplyProfile(profileName string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	profile, exists := lm.profiles[profileName]
	if !exists {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	// Apply global level
	lm.logger.SetLevel(profile.GlobalLevel)

	// Apply module levels
	for module, level := range profile.ModuleLevels {
		lm.logger.SetModuleLevel(module, level)
	}

	// Apply rules
	lm.rules = append(lm.rules[:0], profile.Rules...)

	// Clear rule cache
	lm.cacheMutex.Lock()
	lm.ruleCache = make(map[string]bool)
	lm.cacheMutex.Unlock()

	lm.currentProfile = profile

	lm.logger.InfoLevel(context.Background(), "Applied log level profile",
		map[string]interface{}{
			"profile":      profileName,
			"global_level": profile.GlobalLevel,
			"module_count": len(profile.ModuleLevels),
			"rule_count":   len(profile.Rules),
		})

	return nil
}

// AddRule adds a new log level rule
func (lm *LogLevelManager) AddRule(rule LogLevelRule) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Validate rule
	if err := lm.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	// Compile regex patterns
	for i := range rule.Conditions {
		if rule.Conditions[i].Operator == "regex" {
			if pattern, ok := rule.Conditions[i].Value.(string); ok {
				regex, err := regexp.Compile(pattern)
				if err != nil {
					return fmt.Errorf("invalid regex pattern: %w", err)
				}
				rule.Conditions[i].Regex = regex
			}
		}
	}

	rule.Created = time.Now()
	lm.rules = append(lm.rules, rule)

	// Clear rule cache
	lm.cacheMutex.Lock()
	lm.ruleCache = make(map[string]bool)
	lm.cacheMutex.Unlock()

	return nil
}

// RemoveRule removes a rule by ID
func (lm *LogLevelManager) RemoveRule(ruleID string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for i, rule := range lm.rules {
		if rule.ID == ruleID {
			lm.rules = append(lm.rules[:i], lm.rules[i+1:]...)

			// Clear rule cache
			lm.cacheMutex.Lock()
			lm.ruleCache = make(map[string]bool)
			lm.cacheMutex.Unlock()

			return nil
		}
	}

	return fmt.Errorf("rule with ID '%s' not found", ruleID)
}

// EvaluateRules evaluates all rules against the current log entry context
func (lm *LogLevelManager) EvaluateRules(ctx context.Context, level RFC5424Severity, module string, fields map[string]interface{}) bool {
	lm.mutex.RLock()
	rules := make([]LogLevelRule, len(lm.rules))
	copy(rules, lm.rules)
	lm.mutex.RUnlock()

	// Create cache key
	cacheKey := fmt.Sprintf("%d:%s:%v", level, module, fields)

	// Check cache first
	lm.cacheMutex.RLock()
	if result, exists := lm.ruleCache[cacheKey]; exists {
		lm.cacheMutex.RUnlock()
		return result
	}
	lm.cacheMutex.RUnlock()

	// Evaluate rules
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		conditionsMatch := lm.evaluateRuleConditions(rule, level, module, fields)
		if conditionsMatch {
			lm.applyRuleActions(rule, module)

			// Update rule statistics
			lm.mutex.Lock()
			for i := range lm.rules {
				if lm.rules[i].ID == rule.ID {
					lm.rules[i].ApplyCount++
					now := time.Now()
					lm.rules[i].LastApplied = &now
					break
				}
			}
			lm.mutex.Unlock()

			// Cache result
			lm.cacheMutex.Lock()
			lm.ruleCache[cacheKey] = true
			lm.cacheMutex.Unlock()

			return true
		}
	}

	// Cache negative result
	lm.cacheMutex.Lock()
	lm.ruleCache[cacheKey] = false
	lm.cacheMutex.Unlock()

	return false
}

// evaluateRuleConditions evaluates all conditions for a rule
func (lm *LogLevelManager) evaluateRuleConditions(rule LogLevelRule, level RFC5424Severity, module string, fields map[string]interface{}) bool {
	for _, condition := range rule.Conditions {
		if !lm.evaluateCondition(condition, level, module, fields) {
			return false // All conditions must be true (AND logic)
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (lm *LogLevelManager) evaluateCondition(condition LogCondition, level RFC5424Severity, module string, fields map[string]interface{}) bool {
	switch condition.Type {
	case "level":
		// For RFC5424 severity levels, we need to handle the inverted priority logic
		// Lower numbers = higher priority, so we need to invert comparison operators
		var expectedLevel int
		if conditionLevel, ok := condition.Value.(RFC5424Severity); ok {
			expectedLevel = int(conditionLevel)
		} else if v, ok := condition.Value.(int); ok {
			expectedLevel = v
		} else {
			return false
		}

		actualLevel := int(level)

		// Invert the comparison operators for severity levels
		switch condition.Operator {
		case "eq":
			return actualLevel == expectedLevel
		case "ne":
			return actualLevel != expectedLevel
		case "gt":
			return actualLevel < expectedLevel // Inverted: higher priority = lower number
		case "lt":
			return actualLevel > expectedLevel // Inverted: lower priority = higher number
		case "gte":
			return actualLevel <= expectedLevel // Inverted: same or higher priority = same or lower number
		case "lte":
			return actualLevel >= expectedLevel // Inverted: same or lower priority = same or higher number
		default:
			return false
		}
	case "module":
		return lm.compareValues(module, condition.Operator, condition.Value)
	case "field":
		if value, exists := fields[condition.Field]; exists {
			return lm.compareValues(value, condition.Operator, condition.Value)
		}
		return false
	case "cpu":
		lm.metricsMutex.RLock()
		cpuUsage := lm.metrics.CPUUsage
		lm.metricsMutex.RUnlock()
		return lm.compareValues(cpuUsage, condition.Operator, condition.Value)
	case "memory":
		lm.metricsMutex.RLock()
		memUsage := lm.metrics.MemoryUsage
		lm.metricsMutex.RUnlock()
		return lm.compareValues(memUsage, condition.Operator, condition.Value)
	default:
		return false
	}
}

// compareValues compares two values using the specified operator
func (lm *LogLevelManager) compareValues(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "eq", "ne":
		// Try numeric comparison first for mixed numeric types
		if lm.isNumeric(actual) && lm.isNumeric(expected) {
			actualFloat, actualOk := lm.toFloat64(actual)
			expectedFloat, expectedOk := lm.toFloat64(expected)
			if actualOk && expectedOk {
				if operator == "eq" {
					return actualFloat == expectedFloat
				} else {
					return actualFloat != expectedFloat
				}
			}
		}
		// Fall back to direct comparison
		if operator == "eq" {
			return actual == expected
		} else {
			return actual != expected
		}
	case "contains":
		if actualStr, ok := actual.(string); ok {
			if expectedStr, ok := expected.(string); ok {
				return strings.Contains(actualStr, expectedStr)
			}
		}
		return false
	case "regex":
		if actualStr, ok := actual.(string); ok {
			if pattern, ok := expected.(string); ok {
				matched, _ := regexp.MatchString(pattern, actualStr)
				return matched
			}
		}
		return false
	case "gt", "lt", "gte", "lte":
		return lm.compareNumbers(actual, operator, expected)
	default:
		return false
	}
}

// isNumeric checks if a value is numeric
func (lm *LogLevelManager) isNumeric(value interface{}) bool {
	_, ok := lm.toFloat64(value)
	return ok
}

// compareNumbers compares numeric values
func (lm *LogLevelManager) compareNumbers(actual interface{}, operator string, expected interface{}) bool {
	actualFloat, actualOk := lm.toFloat64(actual)
	expectedFloat, expectedOk := lm.toFloat64(expected)

	if !actualOk || !expectedOk {
		return false
	}

	// Regular numeric comparison
	switch operator {
	case "gt":
		return actualFloat > expectedFloat
	case "lt":
		return actualFloat < expectedFloat
	case "gte":
		return actualFloat >= expectedFloat
	case "lte":
		return actualFloat <= expectedFloat
	default:
		return false
	}
}

// toFloat64 converts interface{} to float64
func (lm *LogLevelManager) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

// parseStringToFloat64 converts string to float64 (for adaptive sampling thresholds)
func (lm *LogLevelManager) parseStringToFloat64(value string) (float64, bool) {
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f, true
	}
	return 0, false
}

// applyRuleActions applies all actions for a rule
func (lm *LogLevelManager) applyRuleActions(rule LogLevelRule, module string) {
	for _, action := range rule.Actions {
		switch action.Type {
		case "set_level":
			if level, ok := action.Value.(RFC5424Severity); ok {
				if action.Target == "global" {
					lm.logger.SetLevel(level)
				} else if action.Target == "module" {
					lm.logger.SetModuleLevel(module, level)
				}
			}
		case "sample":
			// Sampling is handled in the logger's shouldSample method
		case "drop":
			// Log dropping is handled by returning false from rule evaluation
		}
	}
}

// validateRule validates a log level rule
func (lm *LogLevelManager) validateRule(rule LogLevelRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if len(rule.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}

	if len(rule.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	// Validate conditions
	for _, condition := range rule.Conditions {
		if condition.Type == "" {
			return fmt.Errorf("condition type is required")
		}
		if condition.Operator == "" {
			return fmt.Errorf("condition operator is required")
		}
	}

	// Validate actions
	for _, action := range rule.Actions {
		if action.Type == "" {
			return fmt.Errorf("action type is required")
		}
	}

	return nil
}

// collectMetrics collects system performance metrics
func (lm *LogLevelManager) collectMetrics(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Simple CPU usage estimation (not precise, but functional for this purpose)
		cpuUsage := lm.estimateCPUUsage()

		// Calculate memory usage percentage
		memUsage := float64(memStats.Alloc) / float64(memStats.Sys) * 100

		// Calculate log rate
		logRate := lm.calculateLogRate()

		lm.metricsMutex.Lock()
		lm.metrics = SystemMetrics{
			CPUUsage:    cpuUsage,
			MemoryUsage: memUsage,
			LogRate:     logRate,
			Timestamp:   time.Now(),
		}
		lm.metricsMutex.Unlock()

		// Apply adaptive sampling if enabled
		if lm.adaptiveEnabled {
			lm.applyAdaptiveSampling()
		}
	}
}

// estimateCPUUsage provides a simple CPU usage estimation
func (lm *LogLevelManager) estimateCPUUsage() float64 {
	// This is a simplified estimation - in production, you'd want to use
	// more sophisticated CPU monitoring
	return float64(runtime.NumGoroutine()) / float64(runtime.GOMAXPROCS(0)) * 10
}

// calculateLogRate calculates the current logging rate
func (lm *LogLevelManager) calculateLogRate() float64 {
	currentCount := lm.logCounter
	if lm.lastLogCount == 0 {
		lm.lastLogCount = currentCount
		return 0
	}

	rate := float64(currentCount-lm.lastLogCount) / 30.0 // logs per second (30s interval)
	lm.lastLogCount = currentCount
	return rate
}

// applyAdaptiveSampling adjusts sampling rates based on system metrics
func (lm *LogLevelManager) applyAdaptiveSampling() {
	if lm.currentProfile == nil || !lm.currentProfile.Sampling.AdaptiveEnabled {
		return
	}

	lm.metricsMutex.RLock()
	cpuUsage := lm.metrics.CPUUsage
	memUsage := lm.metrics.MemoryUsage
	lm.metricsMutex.RUnlock()

	// Find appropriate sampling rate based on CPU usage
	var sampleRate float64 = 1.0
	for thresholdStr, rate := range lm.currentProfile.Sampling.CPUThresholds {
		if threshold, ok := lm.parseStringToFloat64(thresholdStr); ok && cpuUsage >= threshold {
			sampleRate = rate
		}
	}

	// Adjust based on memory usage as well
	for thresholdStr, rate := range lm.currentProfile.Sampling.MemoryThresholds {
		if threshold, ok := lm.parseStringToFloat64(thresholdStr); ok && memUsage >= threshold && rate < sampleRate {
			sampleRate = rate
		}
	}

	// Apply the calculated sample rate
	// This would need to be integrated with the logger's sampling configuration
	lm.logger.InfoLevel(context.Background(), "Applied adaptive sampling",
		map[string]interface{}{
			"cpu_usage":    cpuUsage,
			"memory_usage": memUsage,
			"sample_rate":  sampleRate,
		})
}

// GetCurrentProfile returns the currently active profile
func (lm *LogLevelManager) GetCurrentProfile() *LogLevelProfile {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	return lm.currentProfile
}

// GetProfiles returns all available profiles
func (lm *LogLevelManager) GetProfiles() map[string]*LogLevelProfile {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	profiles := make(map[string]*LogLevelProfile)
	for name, profile := range lm.profiles {
		profiles[name] = profile
	}
	return profiles
}

// GetRules returns all active rules
func (lm *LogLevelManager) GetRules() []LogLevelRule {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	rules := make([]LogLevelRule, len(lm.rules))
	copy(rules, lm.rules)
	return rules
}

// GetMetrics returns current system metrics
func (lm *LogLevelManager) GetMetrics() SystemMetrics {
	lm.metricsMutex.RLock()
	defer lm.metricsMutex.RUnlock()
	return lm.metrics
}

// Close shuts down the log level manager
func (lm *LogLevelManager) Close() error {
	if lm.httpServer != nil {
		return lm.httpServer.Shutdown(context.Background())
	}

	if lm.signalChan != nil {
		signal.Stop(lm.signalChan)
		close(lm.signalChan)
	}

	return nil
}
