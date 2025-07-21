// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package testlib

import "os"

const trueString = "true"

func IsCI() bool {
	ci := os.Getenv("CI")
	githubActions := os.Getenv("GITHUB_ACTIONS")

	return ci == trueString || githubActions == trueString
}

func IsLocal() bool {
	return os.Getenv("IS_LOCAL") == trueString
}
