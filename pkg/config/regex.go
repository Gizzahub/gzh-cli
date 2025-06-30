package config

import "regexp"

// CompileRegex compiles and validates a regex pattern
func CompileRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}
