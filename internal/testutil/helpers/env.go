package helpers

import (
	"os"
	"testing"
)

// SetEnv sets an environment variable and returns a cleanup function.
func SetEnv(t *testing.T, key, value string) func() {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)

	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env var %s: %v", key, err)
	}

	return func() {
		if existed {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	}
}

// SetEnvs sets multiple environment variables and returns a cleanup function.
func SetEnvs(t *testing.T, envs map[string]string) func() {
	t.Helper()

	cleanups := make([]func(), 0, len(envs))

	for key, value := range envs {
		cleanup := SetEnv(t, key, value)
		cleanups = append(cleanups, cleanup)
	}

	return func() {
		// Execute cleanups in reverse order
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}
}

// UnsetEnv temporarily unsets an environment variable.
func UnsetEnv(t *testing.T, key string) func() {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset env var %s: %v", key, err)
	}

	return func() {
		if existed {
			os.Setenv(key, oldValue)
		}
	}
}

// RequireEnv skips the test if the environment variable is not set.
func RequireEnv(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		t.Skipf("skipping test: %s environment variable not set", key)
	}

	return value
}

// RequireAnyEnv skips the test if none of the environment variables are set.
func RequireAnyEnv(t *testing.T, keys ...string) (string, string) {
	t.Helper()

	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return key, value
		}
	}

	t.Skipf("skipping test: none of %v environment variables are set", keys)

	return "", ""
}
