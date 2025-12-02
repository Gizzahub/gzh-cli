// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

func TestListOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		opts        *ListOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid options",
			opts: &ListOptions{
				Visibility: "all",
				Sort:       "name",
				Order:      "asc",
				Format:     "table",
			},
			expectError: false,
		},
		{
			name: "invalid visibility",
			opts: &ListOptions{
				Visibility: "invalid",
				Sort:       "name",
				Order:      "asc",
				Format:     "table",
			},
			expectError: true,
			errorMsg:    "invalid visibility",
		},
		{
			name: "invalid sort field",
			opts: &ListOptions{
				Visibility: "all",
				Sort:       "invalid",
				Order:      "asc",
				Format:     "table",
			},
			expectError: true,
			errorMsg:    "invalid sort field",
		},
		{
			name: "invalid order",
			opts: &ListOptions{
				Visibility: "all",
				Sort:       "name",
				Order:      "invalid",
				Format:     "table",
			},
			expectError: true,
			errorMsg:    "invalid sort order",
		},
		{
			name: "invalid format",
			opts: &ListOptions{
				Visibility: "all",
				Sort:       "name",
				Order:      "asc",
				Format:     "invalid",
			},
			expectError: true,
			errorMsg:    "invalid output format",
		},
		{
			name: "negative min-stars",
			opts: &ListOptions{
				Visibility: "all",
				Sort:       "name",
				Order:      "asc",
				Format:     "table",
				MinStars:   -1,
			},
			expectError: true,
			errorMsg:    "min-stars cannot be negative",
		},
		{
			name: "max-stars less than min-stars",
			opts: &ListOptions{
				Visibility: "all",
				Sort:       "name",
				Order:      "asc",
				Format:     "table",
				MinStars:   100,
				MaxStars:   50,
			},
			expectError: true,
			errorMsg:    "max-stars must be greater than min-stars",
		},
		{
			name: "valid updated-since date",
			opts: &ListOptions{
				Visibility:   "all",
				Sort:         "name",
				Order:        "asc",
				Format:       "table",
				UpdatedSince: "2024-01-15",
			},
			expectError: false,
		},
		{
			name: "invalid updated-since date",
			opts: &ListOptions{
				Visibility:   "all",
				Sort:         "name",
				Order:        "asc",
				Format:       "table",
				UpdatedSince: "invalid-date",
			},
			expectError: true,
			errorMsg:    "invalid updated-since date format",
		},
		{
			name: "valid ISO 8601 date with time",
			opts: &ListOptions{
				Visibility:   "all",
				Sort:         "name",
				Order:        "asc",
				Format:       "table",
				UpdatedSince: "2024-01-15T10:30:00",
			},
			expectError: false,
		},
		{
			name: "valid RFC3339 date",
			opts: &ListOptions{
				Visibility:   "all",
				Sort:         "name",
				Order:        "asc",
				Format:       "table",
				UpdatedSince: "2024-01-15T10:30:00Z",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListOptions_applyFilters(t *testing.T) {
	now := time.Now()
	oldDate := now.AddDate(0, -6, 0)     // 6 months ago
	newDate := now.AddDate(0, -1, 0)     // 1 month ago
	veryOldDate := now.AddDate(-1, 0, 0) // 1 year ago

	repos := []provider.Repository{
		{
			Name:      "repo-alpha",
			Stars:     100,
			UpdatedAt: newDate,
		},
		{
			Name:      "repo-beta",
			Stars:     50,
			UpdatedAt: oldDate,
		},
		{
			Name:      "test-gamma",
			Stars:     200,
			UpdatedAt: newDate,
		},
		{
			Name:      "repo-delta",
			Stars:     10,
			UpdatedAt: veryOldDate,
		},
	}

	t.Run("filter by name pattern", func(t *testing.T) {
		opts := &ListOptions{Match: "^repo-"}
		filtered := opts.applyFilters(repos)
		assert.Len(t, filtered, 3)
		for _, r := range filtered {
			assert.Contains(t, r.Name, "repo-")
		}
	})

	t.Run("filter by min stars", func(t *testing.T) {
		opts := &ListOptions{MinStars: 100}
		filtered := opts.applyFilters(repos)
		assert.Len(t, filtered, 2)
		for _, r := range filtered {
			assert.GreaterOrEqual(t, r.Stars, 100)
		}
	})

	t.Run("filter by max stars", func(t *testing.T) {
		opts := &ListOptions{MaxStars: 100}
		filtered := opts.applyFilters(repos)
		assert.Len(t, filtered, 3)
		for _, r := range filtered {
			assert.LessOrEqual(t, r.Stars, 100)
		}
	})

	t.Run("filter by star range", func(t *testing.T) {
		opts := &ListOptions{MinStars: 50, MaxStars: 150}
		filtered := opts.applyFilters(repos)
		assert.Len(t, filtered, 2)
		for _, r := range filtered {
			assert.GreaterOrEqual(t, r.Stars, 50)
			assert.LessOrEqual(t, r.Stars, 150)
		}
	})

	t.Run("filter by updated-since", func(t *testing.T) {
		// 3개월 전 날짜로 필터링
		threeMonthsAgo := now.AddDate(0, -3, 0)
		opts := &ListOptions{UpdatedSince: threeMonthsAgo.Format("2006-01-02")}
		filtered := opts.applyFilters(repos)
		// newDate(1개월 전)인 것만 남아야 함
		assert.Len(t, filtered, 2)
	})

	t.Run("combined filters", func(t *testing.T) {
		opts := &ListOptions{
			Match:    "^repo-",
			MinStars: 50,
		}
		filtered := opts.applyFilters(repos)
		assert.Len(t, filtered, 2)
	})

	t.Run("no filters", func(t *testing.T) {
		opts := &ListOptions{}
		filtered := opts.applyFilters(repos)
		assert.Len(t, filtered, len(repos))
	})
}

func TestListOptions_applySorting(t *testing.T) {
	now := time.Now()
	repos := []provider.Repository{
		{Name: "charlie", Stars: 100, CreatedAt: now.AddDate(0, -2, 0), UpdatedAt: now.AddDate(0, 0, -5), Forks: 30},
		{Name: "alpha", Stars: 50, CreatedAt: now.AddDate(0, -1, 0), UpdatedAt: now.AddDate(0, 0, -1), Forks: 10},
		{Name: "bravo", Stars: 200, CreatedAt: now.AddDate(0, -3, 0), UpdatedAt: now.AddDate(0, 0, -10), Forks: 20},
	}

	t.Run("sort by name asc", func(t *testing.T) {
		opts := &ListOptions{Sort: "name", Order: "asc"}
		sorted := opts.applySorting(repos)
		assert.Equal(t, "alpha", sorted[0].Name)
		assert.Equal(t, "bravo", sorted[1].Name)
		assert.Equal(t, "charlie", sorted[2].Name)
	})

	t.Run("sort by name desc", func(t *testing.T) {
		opts := &ListOptions{Sort: "name", Order: "desc"}
		sorted := opts.applySorting(repos)
		assert.Equal(t, "charlie", sorted[0].Name)
		assert.Equal(t, "bravo", sorted[1].Name)
		assert.Equal(t, "alpha", sorted[2].Name)
	})

	t.Run("sort by stars asc", func(t *testing.T) {
		opts := &ListOptions{Sort: "stars", Order: "asc"}
		sorted := opts.applySorting(repos)
		assert.Equal(t, 50, sorted[0].Stars)
		assert.Equal(t, 100, sorted[1].Stars)
		assert.Equal(t, 200, sorted[2].Stars)
	})

	t.Run("sort by stars desc", func(t *testing.T) {
		opts := &ListOptions{Sort: "stars", Order: "desc"}
		sorted := opts.applySorting(repos)
		assert.Equal(t, 200, sorted[0].Stars)
		assert.Equal(t, 100, sorted[1].Stars)
		assert.Equal(t, 50, sorted[2].Stars)
	})

	t.Run("sort by forks asc", func(t *testing.T) {
		opts := &ListOptions{Sort: "forks", Order: "asc"}
		sorted := opts.applySorting(repos)
		assert.Equal(t, 10, sorted[0].Forks)
		assert.Equal(t, 20, sorted[1].Forks)
		assert.Equal(t, 30, sorted[2].Forks)
	})

	t.Run("sort by updated desc", func(t *testing.T) {
		opts := &ListOptions{Sort: "updated", Order: "desc"}
		sorted := opts.applySorting(repos)
		// 가장 최근 업데이트된 것이 첫 번째
		assert.Equal(t, "alpha", sorted[0].Name)
	})

	t.Run("sort by created asc", func(t *testing.T) {
		opts := &ListOptions{Sort: "created", Order: "asc"}
		sorted := opts.applySorting(repos)
		// 가장 오래 전에 생성된 것이 첫 번째
		assert.Equal(t, "bravo", sorted[0].Name)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		slice := []string{"apple", "banana", "cherry"}
		assert.True(t, contains(slice, "banana"))
		assert.False(t, contains(slice, "grape"))
	})

	t.Run("truncateString", func(t *testing.T) {
		assert.Equal(t, "hello", truncateString("hello", 10))
		assert.Equal(t, "hello...", truncateString("hello world", 8))
		assert.Equal(t, "he", truncateString("hello", 2))
	})

	t.Run("formatTime", func(t *testing.T) {
		now := time.Now()
		assert.NotEmpty(t, formatTime(now))
		assert.Empty(t, formatTime(time.Time{}))
	})
}
