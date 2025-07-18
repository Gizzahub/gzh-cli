// Package env provides an abstraction layer for environment variable access
package env

import (
	"os"
	"strings"
)

// Environment defines the interface for environment variable operations.
type Environment interface {
	// Get retrieves the value of the environment variable named by the key
	Get(key string) string

	// LookupEnv retrieves the value of the environment variable named by key
	// If the variable is present in the environment the value is returned and ok is true
	LookupEnv(key string) (string, bool)

	// Set sets the value of the environment variable named by the key
	Set(key, value string) error

	// Unset unsets a single environment variable
	Unset(key string) error

	// Expand replaces ${var} or $var in the string according to the values of environment variables
	Expand(s string) string

	// GetAll returns all environment variables as a map
	GetAll() map[string]string
}

// OSEnvironment implements Environment using the actual OS environment.
type OSEnvironment struct{}

// NewOSEnvironment creates a new OSEnvironment instance.
func NewOSEnvironment() Environment {
	return &OSEnvironment{}
}

// Get retrieves the value of the environment variable named by the key.
func (e *OSEnvironment) Get(key string) string {
	return os.Getenv(key)
}

// LookupEnv retrieves the value of the environment variable named by key.
func (e *OSEnvironment) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// Set sets the value of the environment variable named by the key.
func (e *OSEnvironment) Set(key, value string) error {
	return os.Setenv(key, value)
}

// Unset unsets a single environment variable.
func (e *OSEnvironment) Unset(key string) error {
	return os.Unsetenv(key)
}

// Expand replaces ${var} or $var in the string according to the values of environment variables.
func (e *OSEnvironment) Expand(s string) string {
	return os.ExpandEnv(s)
}

// GetAll returns all environment variables as a map.
func (e *OSEnvironment) GetAll() map[string]string {
	envs := make(map[string]string)

	for _, env := range os.Environ() {
		if idx := strings.Index(env, "="); idx != -1 {
			key := env[:idx]
			value := env[idx+1:]
			envs[key] = value
		}
	}

	return envs
}

// MockEnvironment implements Environment for testing purposes.
type MockEnvironment struct {
	vars map[string]string
}

// NewMockEnvironment creates a new MockEnvironment instance.
func NewMockEnvironment(initialVars map[string]string) Environment {
	vars := make(map[string]string)

	if initialVars != nil {
		for k, v := range initialVars {
			vars[k] = v
		}
	}

	return &MockEnvironment{vars: vars}
}

// Get retrieves the value of the environment variable named by the key.
func (e *MockEnvironment) Get(key string) string {
	return e.vars[key]
}

// LookupEnv retrieves the value of the environment variable named by key.
func (e *MockEnvironment) LookupEnv(key string) (string, bool) {
	value, exists := e.vars[key]
	return value, exists
}

// Set sets the value of the environment variable named by the key.
func (e *MockEnvironment) Set(key, value string) error {
	e.vars[key] = value
	return nil
}

// Unset unsets a single environment variable.
func (e *MockEnvironment) Unset(key string) error {
	delete(e.vars, key)
	return nil
}

// Expand replaces ${var} or $var in the string according to the values of environment variables.
func (e *MockEnvironment) Expand(s string) string {
	return os.Expand(s, func(key string) string {
		return e.vars[key]
	})
}

// GetAll returns all environment variables as a map.
func (e *MockEnvironment) GetAll() map[string]string {
	result := make(map[string]string)
	for k, v := range e.vars {
		result[k] = v
	}

	return result
}

// CommonEnvironmentKeys defines commonly used environment variable keys.
var CommonEnvironmentKeys = struct {
	GitHubToken   string
	GitLabToken   string
	GiteaToken    string
	GZHConfigPath string
	HomeDir       string
	User          string
	Username      string
}{
	GitHubToken:   "GITHUB_TOKEN",
	GitLabToken:   "GITLAB_TOKEN",
	GiteaToken:    "GITEA_TOKEN",
	GZHConfigPath: "GZH_CONFIG_PATH",
	HomeDir:       "HOME",
	User:          "USER",
	Username:      "USERNAME",
}
