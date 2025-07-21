// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import "regexp"

// CompileRegex compiles and validates a regex pattern.
func CompileRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}
