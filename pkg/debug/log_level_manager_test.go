package debug

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLogLevelManagerConfig(t *testing.T) {
	config := DefaultLogLevelManagerConfig()

	assert.True(t, config.EnableHTTPControl)
	assert.Equal(t, 8080, config.HTTPPort)
	assert.True(t, config.EnableSignalControl)
	assert.Equal(t, time.Second*30, config.MetricsInterval)
	assert.Equal(t, time.Minute*5, config.CacheTimeout)
	assert.Equal(t, "development", config.DefaultProfile)
}

func TestNewLogLevelManager(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false, // Disable to avoid port conflicts
		EnableSignalControl: false, // Disable for testing
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
		DefaultProfile:      "testing",
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	assert.NotNil(t, manager)
	assert.Equal(t, "testing", manager.GetCurrentProfile().Name)
	assert.Equal(t, SeverityInfo, logger.GetLevel())
}

func TestBuiltinProfiles(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		DefaultProfile:      "development",
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	profiles := manager.GetProfiles()

	// Check that all built-in profiles exist
	builtinProfiles := []string{"development", "production", "testing"}
	for _, name := range builtinProfiles {
		profile, exists := profiles[name]
		assert.True(t, exists, "Profile %s should exist", name)
		assert.Equal(t, name, profile.Name)
		assert.NotEmpty(t, profile.Description)
	}

	// Test development profile specifics
	devProfile := profiles["development"]
	assert.Equal(t, SeverityDebug, devProfile.GlobalLevel)
	assert.False(t, devProfile.Sampling.Enabled)
	assert.Len(t, devProfile.Rules, 1)

	// Test production profile specifics
	prodProfile := profiles["production"]
	assert.Equal(t, SeverityWarning, prodProfile.GlobalLevel)
	assert.True(t, prodProfile.Sampling.Enabled)
	assert.True(t, prodProfile.Sampling.AdaptiveEnabled)
}

func TestApplyProfile(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		DefaultProfile:      "development",
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	// Apply production profile
	err = manager.ApplyProfile("production")
	assert.NoError(t, err)
	assert.Equal(t, SeverityWarning, logger.GetLevel())
	assert.Equal(t, "production", manager.GetCurrentProfile().Name)

	// Try to apply non-existent profile
	err = manager.ApplyProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLogLevelRule(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	// Create a test rule
	rule := LogLevelRule{
		ID:          "test-rule",
		Name:        "Test Rule",
		Description: "A test rule for unit testing",
		Enabled:     true,
		Priority:    10,
		Conditions: []LogCondition{
			{Type: "module", Operator: "eq", Value: "test"},
			{Type: "level", Operator: "gte", Value: SeverityError},
		},
		Actions: []LogAction{
			{Type: "set_level", Target: "module", Value: SeverityDebug},
		},
	}

	// Add the rule
	err = manager.AddRule(rule)
	assert.NoError(t, err)

	// Verify rule was added
	rules := manager.GetRules()
	assert.Len(t, rules, 1)
	assert.Equal(t, "test-rule", rules[0].ID)

	// Remove the rule
	err = manager.RemoveRule("test-rule")
	assert.NoError(t, err)

	// Verify rule was removed
	rules = manager.GetRules()
	assert.Len(t, rules, 0)
}

func TestRuleValidation(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	tests := []struct {
		name        string
		rule        LogLevelRule
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid rule",
			rule: LogLevelRule{
				ID:   "valid",
				Name: "Valid Rule",
				Conditions: []LogCondition{
					{Type: "level", Operator: "eq", Value: SeverityInfo},
				},
				Actions: []LogAction{
					{Type: "set_level", Target: "global", Value: SeverityDebug},
				},
			},
			expectError: false,
		},
		{
			name: "missing ID",
			rule: LogLevelRule{
				Name: "No ID Rule",
				Conditions: []LogCondition{
					{Type: "level", Operator: "eq", Value: SeverityInfo},
				},
				Actions: []LogAction{
					{Type: "set_level", Target: "global", Value: SeverityDebug},
				},
			},
			expectError: true,
			errorMsg:    "rule ID is required",
		},
		{
			name: "missing conditions",
			rule: LogLevelRule{
				ID:   "no-conditions",
				Name: "No Conditions Rule",
				Actions: []LogAction{
					{Type: "set_level", Target: "global", Value: SeverityDebug},
				},
			},
			expectError: true,
			errorMsg:    "at least one condition is required",
		},
		{
			name: "missing actions",
			rule: LogLevelRule{
				ID:   "no-actions",
				Name: "No Actions Rule",
				Conditions: []LogCondition{
					{Type: "level", Operator: "eq", Value: SeverityInfo},
				},
			},
			expectError: true,
			errorMsg:    "at least one action is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddRule(tt.rule)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				// Clean up
				manager.RemoveRule(tt.rule.ID)
			}
		})
	}
}

func TestRuleEvaluation(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	// Add test rules
	rules := []LogLevelRule{
		{
			ID:       "module-rule",
			Name:     "Module Rule",
			Enabled:  true,
			Priority: 10,
			Conditions: []LogCondition{
				{Type: "module", Operator: "eq", Value: "auth"},
			},
			Actions: []LogAction{
				{Type: "set_level", Target: "module", Value: SeverityDebug},
			},
		},
		{
			ID:       "level-rule",
			Name:     "Level Rule",
			Enabled:  true,
			Priority: 20,
			Conditions: []LogCondition{
				{Type: "level", Operator: "gte", Value: SeverityError},
			},
			Actions: []LogAction{
				{Type: "set_level", Target: "global", Value: SeverityInfo},
			},
		},
		{
			ID:       "disabled-rule",
			Name:     "Disabled Rule",
			Enabled:  false,
			Priority: 30,
			Conditions: []LogCondition{
				{Type: "module", Operator: "eq", Value: "disabled"},
			},
			Actions: []LogAction{
				{Type: "set_level", Target: "global", Value: SeverityDebug},
			},
		},
	}

	// Clear existing rules first to avoid interference
	manager.mutex.Lock()
	manager.rules = []LogLevelRule{}
	manager.mutex.Unlock()

	for _, rule := range rules {
		err := manager.AddRule(rule)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		level    RFC5424Severity
		module   string
		fields   map[string]interface{}
		expected bool
	}{
		{
			name:     "matches module rule",
			level:    SeverityInfo,
			module:   "auth",
			fields:   nil,
			expected: true,
		},
		{
			name:     "matches level rule",
			level:    SeverityError,
			module:   "api",
			fields:   nil,
			expected: true,
		},
		{
			name:     "no match",
			level:    SeverityInfo,
			module:   "api",
			fields:   nil,
			expected: false,
		},
		{
			name:     "disabled rule should not match",
			level:    SeverityInfo,
			module:   "disabled",
			fields:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.EvaluateRules(context.Background(), tt.level, tt.module, tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionEvaluation(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	tests := []struct {
		name      string
		condition LogCondition
		level     RFC5424Severity
		module    string
		fields    map[string]interface{}
		expected  bool
	}{
		{
			name:      "level equality",
			condition: LogCondition{Type: "level", Operator: "eq", Value: SeverityInfo},
			level:     SeverityInfo,
			expected:  true,
		},
		{
			name:      "level greater than",
			condition: LogCondition{Type: "level", Operator: "gt", Value: SeverityInfo},
			level:     SeverityWarning,
			expected:  true, // Warning (4) has higher priority than Info (6)
		},
		{
			name:      "module contains",
			condition: LogCondition{Type: "module", Operator: "contains", Value: "auth"},
			module:    "authentication",
			expected:  true,
		},
		{
			name:      "module exact match",
			condition: LogCondition{Type: "module", Operator: "eq", Value: "auth"},
			module:    "authentication",
			expected:  false,
		},
		{
			name:      "field exists and matches",
			condition: LogCondition{Type: "field", Field: "user_id", Operator: "eq", Value: "123"},
			fields:    map[string]interface{}{"user_id": "123"},
			expected:  true,
		},
		{
			name:      "field missing",
			condition: LogCondition{Type: "field", Field: "user_id", Operator: "eq", Value: "123"},
			fields:    map[string]interface{}{"other_field": "value"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.evaluateCondition(tt.condition, tt.level, tt.module, tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTTPEndpoints(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false, // We'll test endpoints directly
		EnableSignalControl: false,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	// Test health check
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	manager.handleHealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var health map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &health)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", health["status"])

	// Test get profiles
	req = httptest.NewRequest("GET", "/api/v1/logging/profiles", nil)
	w = httptest.NewRecorder()
	manager.handleGetProfiles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var profiles map[string]*LogLevelProfile
	err = json.Unmarshal(w.Body.Bytes(), &profiles)
	assert.NoError(t, err)
	assert.Contains(t, profiles, "development")
	assert.Contains(t, profiles, "production")
	assert.Contains(t, profiles, "testing")

	// Test get current level
	req = httptest.NewRequest("GET", "/api/v1/logging/level", nil)
	w = httptest.NewRecorder()
	manager.handleGetLevel(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var levelResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &levelResp)
	assert.NoError(t, err)
	assert.Contains(t, levelResp, "level")
	assert.Contains(t, levelResp, "name")
	assert.Contains(t, levelResp, "numeric")

	// Test set level
	setLevelReq := map[string]string{"level": "error"}
	reqBody, _ := json.Marshal(setLevelReq)
	req = httptest.NewRequest("POST", "/api/v1/logging/level", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	manager.handleSetLevel(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, SeverityError, logger.GetLevel())

	// Test metrics endpoint
	req = httptest.NewRequest("GET", "/api/v1/logging/metrics", nil)
	w = httptest.NewRecorder()
	manager.handleGetMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var metrics SystemMetrics
	err = json.Unmarshal(w.Body.Bytes(), &metrics)
	assert.NoError(t, err)
}

func TestCompareValues(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	tests := []struct {
		name     string
		actual   interface{}
		operator string
		expected interface{}
		result   bool
	}{
		{"string equality", "test", "eq", "test", true},
		{"string inequality", "test", "ne", "other", true},
		{"string contains", "testing", "contains", "test", true},
		{"string not contains", "testing", "contains", "xyz", false},
		{"number greater than", 10, "gt", 5, true},
		{"number less than", 5, "lt", 10, true},
		{"number greater than equal", 10, "gte", 10, true},
		{"number less than equal", 5, "lte", 5, true},
		{"float comparison", 10.5, "gt", 10.0, true},
		{"mixed number types", int64(10), "eq", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.compareValues(tt.actual, tt.operator, tt.expected)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestToFloat64(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	tests := []struct {
		name     string
		value    interface{}
		expected float64
		ok       bool
	}{
		{"int", 10, 10.0, true},
		{"int64", int64(10), 10.0, true},
		{"float64", 10.5, 10.5, true},
		{"float32", float32(10.5), 10.5, true},
		{"string", "10", 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := manager.toFloat64(tt.value)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRuleCache(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		CacheTimeout:        time.Millisecond * 100,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	// Add a test rule
	rule := LogLevelRule{
		ID:       "cache-test",
		Name:     "Cache Test Rule",
		Enabled:  true,
		Priority: 10,
		Conditions: []LogCondition{
			{Type: "module", Operator: "eq", Value: "cache-test"},
		},
		Actions: []LogAction{
			{Type: "set_level", Target: "module", Value: SeverityDebug},
		},
	}

	err = manager.AddRule(rule)
	require.NoError(t, err)

	ctx := context.Background()
	level := SeverityInfo
	module := "cache-test"
	fields := map[string]interface{}{}

	// First evaluation should add to cache
	result1 := manager.EvaluateRules(ctx, level, module, fields)
	assert.True(t, result1)

	// Second evaluation should use cache
	result2 := manager.EvaluateRules(ctx, level, module, fields)
	assert.True(t, result2)

	// Verify cache has entry
	cacheKey := "6:cache-test:map[]"
	manager.cacheMutex.RLock()
	_, exists := manager.ruleCache[cacheKey]
	manager.cacheMutex.RUnlock()
	assert.True(t, exists)
}

func TestMetricsCollection(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 50,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(t, err)
	defer manager.Close()

	// Wait for metrics collection
	time.Sleep(time.Millisecond * 100)

	metrics := manager.GetMetrics()
	assert.True(t, metrics.Timestamp.After(time.Time{}))
	assert.GreaterOrEqual(t, metrics.CPUUsage, 0.0)
	assert.GreaterOrEqual(t, metrics.MemoryUsage, 0.0)
	assert.GreaterOrEqual(t, metrics.LogRate, 0.0)
}

func BenchmarkRuleEvaluation(b *testing.B) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(b, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(b, err)
	defer manager.Close()

	// Add multiple rules
	for i := 0; i < 10; i++ {
		rule := LogLevelRule{
			ID:       fmt.Sprintf("bench-rule-%d", i),
			Name:     fmt.Sprintf("Benchmark Rule %d", i),
			Enabled:  true,
			Priority: i,
			Conditions: []LogCondition{
				{Type: "module", Operator: "contains", Value: fmt.Sprintf("bench-%d", i)},
			},
			Actions: []LogAction{
				{Type: "set_level", Target: "module", Value: SeverityDebug},
			},
		}
		err = manager.AddRule(rule)
		require.NoError(b, err)
	}

	ctx := context.Background()
	level := SeverityInfo
	module := "bench-5"
	fields := map[string]interface{}{"test": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.EvaluateRules(ctx, level, module, fields)
	}
}

func BenchmarkConditionEvaluation(b *testing.B) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(b, err)
	defer logger.Close()

	config := &LogLevelManagerConfig{
		EnableHTTPControl:   false,
		EnableSignalControl: false,
		MetricsInterval:     time.Millisecond * 100,
		CacheTimeout:        time.Second,
	}

	manager, err := NewLogLevelManager(logger, config)
	require.NoError(b, err)
	defer manager.Close()

	condition := LogCondition{
		Type:     "module",
		Operator: "contains",
		Value:    "test",
	}

	level := SeverityInfo
	module := "test-module"
	fields := map[string]interface{}{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.evaluateCondition(condition, level, module, fields)
	}
}
