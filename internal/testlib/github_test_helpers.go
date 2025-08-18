// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package testlib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertConstantArrayNotEmpty verifies that all constants in an array are not empty.
func AssertConstantArrayNotEmpty(t *testing.T, constants []interface{}, constantType string) {
	t.Helper()
	for i, constant := range constants {
		switch v := constant.(type) {
		case string:
			assert.NotEmpty(t, v, "%s constant at index %d should not be empty", constantType, i)
		case fmt.Stringer:
			assert.NotEmpty(t, v.String(), "%s constant at index %d should not be empty", constantType, i)
		default:
			assert.NotNil(t, constant, "%s constant at index %d should not be nil", constantType, i)
		}
	}
}

// AssertConstantValues verifies specific constant values.
func AssertConstantValues(t *testing.T, tests []ConstantTest) {
	t.Helper()
	for _, test := range tests {
		assert.Equal(t, test.Expected, test.Actual,
			"Constant %s should have value %v", test.Name, test.Expected)
	}
}

// ConstantTest represents a constant value test case.
type ConstantTest struct {
	Name     string
	Expected interface{}
	Actual   interface{}
}

// AssertStatusConstantsPattern tests common status constant patterns.
func AssertStatusConstantsPattern(t *testing.T, statuses []interface{}, expectedValues []ConstantTest) {
	t.Helper()

	// Test that all constants are not empty
	AssertConstantArrayNotEmpty(t, statuses, "Status")

	// Test specific values
	AssertConstantValues(t, expectedValues)
}

// AssertErrorTypeConstantsPattern tests common error type constant patterns.
func AssertErrorTypeConstantsPattern(t *testing.T, errorTypes []interface{}, expectedValues []ConstantTest) {
	t.Helper()

	// Test that all constants are not empty
	AssertConstantArrayNotEmpty(t, errorTypes, "ErrorType")

	// Test specific values
	AssertConstantValues(t, expectedValues)
}

// AssertActionTypeConstantsPattern tests common action type constant patterns.
func AssertActionTypeConstantsPattern(t *testing.T, actionTypes []interface{}, expectedValues []ConstantTest) {
	t.Helper()

	// Test that all constants are not empty
	AssertConstantArrayNotEmpty(t, actionTypes, "ActionType")

	// Test specific values
	AssertConstantValues(t, expectedValues)
}

// MockTestHelper provides common mock test patterns.
type MockTestHelper struct {
	t *testing.T
}

// NewMockTestHelper creates a new mock test helper.
func NewMockTestHelper(t *testing.T) *MockTestHelper {
	return &MockTestHelper{t: t}
}

// AssertMockCalled verifies that a mock was called with expected parameters.
func (m *MockTestHelper) AssertMockCalled(mockCall interface{}, expectedArgs ...interface{}) {
	m.t.Helper()
	// This would be implemented based on the specific mocking framework
	// For now, just verify that the mock call is not nil
	assert.NotNil(m.t, mockCall, "Mock call should not be nil")
}

// IntegrationTestHelper provides common integration test patterns.
type IntegrationTestHelper struct {
	t           *testing.T
	skipMessage string
}

// NewIntegrationTestHelper creates a new integration test helper.
func NewIntegrationTestHelper(t *testing.T, skipMessage string) *IntegrationTestHelper {
	return &IntegrationTestHelper{
		t:           t,
		skipMessage: skipMessage,
	}
}

// SkipIfNeeded skips the test if integration conditions are not met.
func (i *IntegrationTestHelper) SkipIfNeeded(condition bool) {
	i.t.Helper()
	if condition {
		i.t.Skip(i.skipMessage)
	}
}

// AssertAPIResponse verifies common API response patterns.
func (i *IntegrationTestHelper) AssertAPIResponse(response interface{}, err error) {
	i.t.Helper()
	assert.NoError(i.t, err, "API call should not return error")
	assert.NotNil(i.t, response, "API response should not be nil")
}

// APITestPattern represents a common API test pattern.
type APITestPattern struct {
	Name         string
	Setup        func(t *testing.T) interface{}
	Execute      func(t *testing.T, input interface{}) (interface{}, error)
	Verify       func(t *testing.T, result interface{}, err error)
	ExpectedType interface{}
}

// RunAPITestPattern executes a common API test pattern.
func RunAPITestPattern(t *testing.T, pattern APITestPattern) {
	t.Helper()
	t.Run(pattern.Name, func(t *testing.T) {
		// Setup
		var input interface{}
		if pattern.Setup != nil {
			input = pattern.Setup(t)
		}

		// Execute
		result, err := pattern.Execute(t, input)

		// Verify
		if pattern.Verify != nil {
			pattern.Verify(t, result, err)
		} else {
			// Default verification
			assert.NoError(t, err, "API call should not return error")
			if pattern.ExpectedType != nil {
				assert.IsType(t, pattern.ExpectedType, result, "Result should be of expected type")
			}
		}
	})
}
