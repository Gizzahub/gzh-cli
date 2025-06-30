package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitTarget_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		target   GitTarget
		expected GitTarget
	}{
		{
			name:   "empty target gets defaults",
			target: GitTarget{Name: "test"},
			expected: GitTarget{
				Name:       "test",
				Visibility: VisibilityAll,
				Strategy:   StrategyReset,
			},
		},
		{
			name: "existing values not overwritten",
			target: GitTarget{
				Name:       "test",
				Visibility: VisibilityPublic,
				Strategy:   StrategyPull,
			},
			expected: GitTarget{
				Name:       "test",
				Visibility: VisibilityPublic,
				Strategy:   StrategyPull,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.SetDefaults()
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name:    "missing version",
			config:  Config{},
			wantErr: ErrMissingVersion,
		},
		{
			name: "valid config",
			config: Config{
				Version: "1.0.0",
				Providers: map[string]Provider{
					"github": {
						Token: "${GITHUB_TOKEN}",
						Orgs: []GitTarget{
							{Name: "test-org"},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitTarget_Validate(t *testing.T) {
	tests := []struct {
		name    string
		target  GitTarget
		wantErr error
	}{
		{
			name:    "missing name",
			target:  GitTarget{},
			wantErr: ErrMissingName,
		},
		{
			name: "invalid visibility",
			target: GitTarget{
				Name:       "test",
				Visibility: "invalid",
			},
			wantErr: ErrInvalidVisibility,
		},
		{
			name: "invalid strategy",
			target: GitTarget{
				Name:     "test",
				Strategy: "invalid",
			},
			wantErr: ErrInvalidStrategy,
		},
		{
			name: "invalid regex",
			target: GitTarget{
				Name:  "test",
				Match: "[invalid",
			},
			wantErr: ErrInvalidRegex,
		},
		{
			name: "valid target",
			target: GitTarget{
				Name:       "test",
				Visibility: VisibilityPublic,
				Strategy:   StrategyPull,
				Match:      "^test-.*",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}