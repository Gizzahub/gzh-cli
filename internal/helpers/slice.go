// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package helpers

// Contains checks if a slice contains a specific element.
func Contains(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}

	return false
}

// Difference returns the elements in 'a' that are not in 'b'.
func Difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}

	var diff []string

	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}

	return diff
}

// Unique returns unique elements from a slice.
func Unique(items []string) []string {
	seen := make(map[string]struct{})

	var result []string

	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// Filter returns elements that match the predicate function.
func Filter(items []string, predicate func(string) bool) []string {
	var result []string

	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}
