package testlib

import "os"

const trueString = "true"

// IsCI returns true if running in a CI environment.
func IsCI() bool {
	ci := os.Getenv("CI")
	githubActions := os.Getenv("GITHUB_ACTIONS")

	return ci == trueString || githubActions == trueString
}

// IsLocal returns true if running in a local development environment.
func IsLocal() bool {
	return os.Getenv("IS_LOCAL") == trueString
}
