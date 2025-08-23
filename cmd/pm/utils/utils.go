// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package utils provides common utility functions for PM packages
package utils

import "strings"

// ParseCSVList parses a comma-separated string into a slice of strings.
func ParseCSVList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
