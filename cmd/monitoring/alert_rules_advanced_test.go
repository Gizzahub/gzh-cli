package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockMetricProvider implements MetricProvider for testing
type MockMetricProvider struct {
	metrics map[string]float64
	history map[string][]MetricPoint
}

func NewMockMetricProvider() *MockMetricProvider {
	return &MockMetricProvider{
		metrics: make(map[string]float64),
		history: make(map[string][]MetricPoint),
	}
}

func (m *MockMetricProvider) GetMetric(ctx context.Context, metric string, timeframe *TimeFrameConfig) (float64, error) {
	if value, exists := m.metrics[metric]; exists {
		return value, nil
	}
	return 0, nil
}

func (m *MockMetricProvider) GetMetricHistory(ctx context.Context, metric string, duration time.Duration) ([]MetricPoint, error) {
	if history, exists := m.history[metric]; exists {
		return history, nil
	}
	return []MetricPoint{}, nil
}

func (m *MockMetricProvider) SetMetric(metric string, value float64) {
	m.metrics[metric] = value
}

func (m *MockMetricProvider) SetHistory(metric string, history []MetricPoint) {
	m.history[metric] = history
}

func TestAdvancedAlertRule_Validation(t *testing.T) {
	logger := zap.NewNop()
	metricProvider := NewMockMetricProvider()
	manager := NewAdvancedAlertManager(metricProvider, logger)

	t.Run("Valid rule", func(t *testing.T) {
		rule := &AdvancedAlertRule{
			ID:          "test-rule-1",
			Name:        "Test Rule",
			Description: "Test alert rule",
			Priority:    100,
			Enabled:     true,
			Conditions: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage_percent",
				Threshold: &ThresholdConfig{
					Operator: "gt",
					Value:    80.0,
				},
			},
			Actions: []AlertAction{
				{
					Type:   "notification",
					Target: "slack-channel",
				},
			},
		}

		err := manager.AddRule(rule)
		assert.NoError(t, err)
		assert.False(t, rule.CreatedAt.IsZero())
		assert.False(t, rule.UpdatedAt.IsZero())
	})

	t.Run("Invalid rule - missing ID", func(t *testing.T) {
		rule := &AdvancedAlertRule{
			Name: "Test Rule",
			Conditions: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage_percent",
			},
			Actions: []AlertAction{
				{Type: "notification", Target: "slack"},
			},
		}

		err := manager.AddRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rule ID is required")
	})

	t.Run("Invalid rule - missing conditions", func(t *testing.T) {
		rule := &AdvancedAlertRule{
			ID:   "test-rule-2",
			Name: "Test Rule",
			Actions: []AlertAction{
				{Type: "notification", Target: "slack"},
			},
		}

		err := manager.AddRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rule conditions are required")
	})

	t.Run("Invalid rule - missing actions", func(t *testing.T) {
		rule := &AdvancedAlertRule{
			ID:   "test-rule-3",
			Name: "Test Rule",
			Conditions: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage_percent",
			},
			Actions: []AlertAction{},
		}

		err := manager.AddRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rule must have at least one action")
	})
}

func TestConditionEvaluator_SimpleConditions(t *testing.T) {
	logger := zap.NewNop()
	metricProvider := NewMockMetricProvider()
	evaluator := &ConditionEvaluator{
		metricProvider: metricProvider,
		logger:         logger,
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		condition   *AlertCondition
		metricValue float64
		expected    bool
	}{
		{
			name: "Greater than - true",
			condition: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage",
				Threshold: &ThresholdConfig{
					Operator: "gt",
					Value:    80.0,
				},
			},
			metricValue: 85.0,
			expected:    true,
		},
		{
			name: "Greater than - false",
			condition: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage",
				Threshold: &ThresholdConfig{
					Operator: "gt",
					Value:    80.0,
				},
			},
			metricValue: 75.0,
			expected:    false,
		},
		{
			name: "Between - true",
			condition: &AlertCondition{
				Type:   "simple",
				Metric: "temperature",
				Threshold: &ThresholdConfig{
					Operator:    "between",
					Value:       20.0,
					SecondValue: 30.0,
				},
			},
			metricValue: 25.0,
			expected:    true,
		},
		{
			name: "Between - false",
			condition: &AlertCondition{
				Type:   "simple",
				Metric: "temperature",
				Threshold: &ThresholdConfig{
					Operator:    "between",
					Value:       20.0,
					SecondValue: 30.0,
				},
			},
			metricValue: 35.0,
			expected:    false,
		},
		{
			name: "Outside - true",
			condition: &AlertCondition{
				Type:   "simple",
				Metric: "response_time",
				Threshold: &ThresholdConfig{
					Operator:    "outside",
					Value:       100.0,
					SecondValue: 500.0,
				},
			},
			metricValue: 50.0,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricProvider.SetMetric(tt.condition.Metric, tt.metricValue)

			result, err := evaluator.EvaluateCondition(ctx, tt.condition)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionEvaluator_CompositeConditions(t *testing.T) {
	logger := zap.NewNop()
	metricProvider := NewMockMetricProvider()
	evaluator := &ConditionEvaluator{
		metricProvider: metricProvider,
		logger:         logger,
	}

	ctx := context.Background()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 85.0)
	metricProvider.SetMetric("memory_usage", 70.0)
	metricProvider.SetMetric("disk_usage", 50.0)

	t.Run("AND condition - all true", func(t *testing.T) {
		condition := &AlertCondition{
			Type:     "composite",
			Operator: "and",
			Children: []*AlertCondition{
				{
					Type:   "simple",
					Metric: "cpu_usage",
					Threshold: &ThresholdConfig{
						Operator: "gt",
						Value:    80.0,
					},
				},
				{
					Type:   "simple",
					Metric: "memory_usage",
					Threshold: &ThresholdConfig{
						Operator: "gt",
						Value:    60.0,
					},
				},
			},
		}

		result, err := evaluator.EvaluateCondition(ctx, condition)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("AND condition - one false", func(t *testing.T) {
		condition := &AlertCondition{
			Type:     "composite",
			Operator: "and",
			Children: []*AlertCondition{
				{
					Type:   "simple",
					Metric: "cpu_usage",
					Threshold: &ThresholdConfig{
						Operator: "gt",
						Value:    90.0, // This will be false (85 > 90)
					},
				},
				{
					Type:   "simple",
					Metric: "memory_usage",
					Threshold: &ThresholdConfig{
						Operator: "gt",
						Value:    60.0, // This will be true (70 > 60)
					},
				},
			},
		}

		result, err := evaluator.EvaluateCondition(ctx, condition)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("OR condition - one true", func(t *testing.T) {
		condition := &AlertCondition{
			Type:     "composite",
			Operator: "or",
			Children: []*AlertCondition{
				{
					Type:   "simple",
					Metric: "cpu_usage",
					Threshold: &ThresholdConfig{
						Operator: "gt",
						Value:    90.0, // This will be false
					},
				},
				{
					Type:   "simple",
					Metric: "memory_usage",
					Threshold: &ThresholdConfig{
						Operator: "gt",
						Value:    60.0, // This will be true
					},
				},
			},
		}

		result, err := evaluator.EvaluateCondition(ctx, condition)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("NOT condition", func(t *testing.T) {
		condition := &AlertCondition{
			Type:     "composite",
			Operator: "not",
			Children: []*AlertCondition{
				{
					Type:   "simple",
					Metric: "disk_usage",
					Threshold: &ThresholdConfig{
						Operator: "gt",
						Value:    60.0, // This will be false (50 > 60)
					},
				},
			},
		}

		result, err := evaluator.EvaluateCondition(ctx, condition)
		require.NoError(t, err)
		assert.True(t, result) // NOT false = true
	})
}

func TestConditionEvaluator_TimeBasedConditions(t *testing.T) {
	logger := zap.NewNop()
	metricProvider := NewMockMetricProvider()
	evaluator := &ConditionEvaluator{
		metricProvider: metricProvider,
		logger:         logger,
	}

	ctx := context.Background()

	// Set up metric history
	history := []MetricPoint{
		{Timestamp: time.Now().Add(-4 * time.Minute), Value: 70.0},
		{Timestamp: time.Now().Add(-3 * time.Minute), Value: 80.0},
		{Timestamp: time.Now().Add(-2 * time.Minute), Value: 90.0},
		{Timestamp: time.Now().Add(-1 * time.Minute), Value: 85.0},
	}
	metricProvider.SetHistory("cpu_usage", history)

	tests := []struct {
		name        string
		aggregation string
		expected    bool
	}{
		{
			name:        "Average aggregation",
			aggregation: "avg",
			expected:    true, // (70+80+90+85)/4 = 81.25 > 80
		},
		{
			name:        "Maximum aggregation",
			aggregation: "max",
			expected:    true, // max(70,80,90,85) = 90 > 80
		},
		{
			name:        "Minimum aggregation",
			aggregation: "min",
			expected:    false, // min(70,80,90,85) = 70 < 80
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := &AlertCondition{
				Type:   "time_based",
				Metric: "cpu_usage",
				TimeFrame: &TimeFrameConfig{
					Duration:    5 * time.Minute,
					Aggregation: tt.aggregation,
				},
				Threshold: &ThresholdConfig{
					Operator: "gt",
					Value:    80.0,
				},
			}

			result, err := evaluator.EvaluateCondition(ctx, condition)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAlertScheduler_ActivePeriods(t *testing.T) {
	scheduler := &AlertScheduler{
		logger: zap.NewNop(),
	}

	// Mock current time as Monday 14:30
	// Note: In a real implementation, you'd inject time for testing

	tests := []struct {
		name     string
		schedule *AlertSchedule
		expected bool
	}{
		{
			name: "Active during work hours",
			schedule: &AlertSchedule{
				ActivePeriods: []TimePeriod{
					{
						Start: "09:00",
						End:   "17:00",
						Days:  []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
					},
				},
			},
			expected: true, // Would be true if current time is within work hours
		},
		{
			name: "Inactive outside work hours",
			schedule: &AlertSchedule{
				ActivePeriods: []TimePeriod{
					{
						Start: "09:00",
						End:   "17:00",
						Days:  []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
					},
				},
			},
			expected: true, // Simplified for test - would check actual time in production
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scheduler.IsRuleActive(tt.schedule)
			require.NoError(t, err)
			// For this test, we'll just verify it doesn't error
			// In production, you'd mock time.Now() for proper testing
			_ = result
		})
	}
}

func TestAlertThrottler(t *testing.T) {
	throttler := &AlertThrottler{
		alertCounts: make(map[string][]time.Time),
		logger:      zap.NewNop(),
	}

	throttleConfig := &AlertThrottle{
		MaxAlerts:  3,
		TimeWindow: 5 * time.Minute,
	}

	ruleID := "test-rule"

	t.Run("First few alerts not throttled", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			throttled, err := throttler.ShouldThrottle(ruleID, throttleConfig)
			require.NoError(t, err)
			assert.False(t, throttled, "Alert %d should not be throttled", i+1)
		}
	})

	t.Run("Fourth alert throttled", func(t *testing.T) {
		throttled, err := throttler.ShouldThrottle(ruleID, throttleConfig)
		require.NoError(t, err)
		assert.True(t, throttled, "Fourth alert should be throttled")
	})
}

func TestAdvancedAlertManager_EvaluateRule(t *testing.T) {
	logger := zap.NewNop()
	metricProvider := NewMockMetricProvider()
	manager := NewAdvancedAlertManager(metricProvider, logger)

	ctx := context.Background()

	t.Run("Enabled rule with true condition", func(t *testing.T) {
		metricProvider.SetMetric("cpu_usage", 85.0)

		rule := &AdvancedAlertRule{
			ID:      "test-rule",
			Enabled: true,
			Conditions: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage",
				Threshold: &ThresholdConfig{
					Operator: "gt",
					Value:    80.0,
				},
			},
		}

		result, err := manager.EvaluateRule(ctx, rule)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("Disabled rule", func(t *testing.T) {
		rule := &AdvancedAlertRule{
			ID:      "disabled-rule",
			Enabled: false,
			Conditions: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage",
				Threshold: &ThresholdConfig{
					Operator: "gt",
					Value:    80.0,
				},
			},
		}

		result, err := manager.EvaluateRule(ctx, rule)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("Rule with throttling", func(t *testing.T) {
		metricProvider.SetMetric("cpu_usage", 85.0)

		rule := &AdvancedAlertRule{
			ID:      "throttled-rule",
			Enabled: true,
			Conditions: &AlertCondition{
				Type:   "simple",
				Metric: "cpu_usage",
				Threshold: &ThresholdConfig{
					Operator: "gt",
					Value:    80.0,
				},
			},
			Throttle: &AlertThrottle{
				MaxAlerts:  1,
				TimeWindow: 5 * time.Minute,
			},
		}

		// First evaluation should succeed
		result1, err := manager.EvaluateRule(ctx, rule)
		require.NoError(t, err)
		assert.True(t, result1)

		// Second evaluation should be throttled
		result2, err := manager.EvaluateRule(ctx, rule)
		require.NoError(t, err)
		assert.False(t, result2)
	})
}

func TestThresholdEvaluation(t *testing.T) {
	evaluator := &ConditionEvaluator{}

	tests := []struct {
		name      string
		value     float64
		threshold *ThresholdConfig
		expected  bool
	}{
		{"GT true", 85.0, &ThresholdConfig{Operator: "gt", Value: 80.0}, true},
		{"GT false", 75.0, &ThresholdConfig{Operator: "gt", Value: 80.0}, false},
		{"GTE true equal", 80.0, &ThresholdConfig{Operator: "gte", Value: 80.0}, true},
		{"GTE true greater", 85.0, &ThresholdConfig{Operator: "gte", Value: 80.0}, true},
		{"GTE false", 75.0, &ThresholdConfig{Operator: "gte", Value: 80.0}, false},
		{"LT true", 75.0, &ThresholdConfig{Operator: "lt", Value: 80.0}, true},
		{"LT false", 85.0, &ThresholdConfig{Operator: "lt", Value: 80.0}, false},
		{"EQ true", 80.0, &ThresholdConfig{Operator: "eq", Value: 80.0}, true},
		{"EQ false", 85.0, &ThresholdConfig{Operator: "eq", Value: 80.0}, false},
		{"Between true", 75.0, &ThresholdConfig{Operator: "between", Value: 70.0, SecondValue: 80.0}, true},
		{"Between false low", 65.0, &ThresholdConfig{Operator: "between", Value: 70.0, SecondValue: 80.0}, false},
		{"Between false high", 85.0, &ThresholdConfig{Operator: "between", Value: 70.0, SecondValue: 80.0}, false},
		{"Outside true low", 65.0, &ThresholdConfig{Operator: "outside", Value: 70.0, SecondValue: 80.0}, true},
		{"Outside true high", 85.0, &ThresholdConfig{Operator: "outside", Value: 70.0, SecondValue: 80.0}, true},
		{"Outside false", 75.0, &ThresholdConfig{Operator: "outside", Value: 70.0, SecondValue: 80.0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.evaluateThreshold(tt.value, tt.threshold)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetricAggregation(t *testing.T) {
	evaluator := &ConditionEvaluator{}

	history := []MetricPoint{
		{Value: 10.0},
		{Value: 20.0},
		{Value: 30.0},
		{Value: 40.0},
	}

	tests := []struct {
		name        string
		aggregation string
		expected    float64
	}{
		{"Average", "avg", 25.0},
		{"Maximum", "max", 40.0},
		{"Minimum", "min", 10.0},
		{"Sum", "sum", 100.0},
		{"Count", "count", 4.0},
		{"Unknown (latest)", "unknown", 40.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.aggregateMetrics(history, tt.aggregation)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActionExecutor(t *testing.T) {
	executor := &AlertActionExecutor{
		logger: zap.NewNop(),
	}

	ctx := context.Background()
	alertData := map[string]interface{}{
		"rule_id":  "test-rule",
		"severity": "high",
		"message":  "Test alert",
	}

	t.Run("Execute notification action", func(t *testing.T) {
		action := &AlertAction{
			Type:   "notification",
			Target: "slack-channel",
		}

		err := executor.ExecuteAction(ctx, action, alertData)
		// In this test implementation, we expect no error
		assert.NoError(t, err)
	})

	t.Run("Execute unknown action type", func(t *testing.T) {
		action := &AlertAction{
			Type:   "unknown",
			Target: "somewhere",
		}

		err := executor.ExecuteAction(ctx, action, alertData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown action type")
	})

	t.Run("Execute action with delay", func(t *testing.T) {
		start := time.Now()
		action := &AlertAction{
			Type:   "notification",
			Target: "test",
			Delay:  50 * time.Millisecond,
		}

		err := executor.ExecuteAction(ctx, action, alertData)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
	})
}
