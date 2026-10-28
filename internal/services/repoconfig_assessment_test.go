// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli/pkg/config"
	"github.com/gizzahub/gzh-cli/pkg/github"
)

func TestRepositoryComplianceIsNotAssessed(t *testing.T) {
	service := &RepoConfigService{}
	for _, template := range []string{"", "fixture-template"} {
		cfg := &config.RepoConfig{Repositories: &config.RepoTargets{Default: &config.RepoDefaultConfig{Template: template}}}
		info := service.createRepositoryInfo(&github.Repository{Name: "fixture"}, cfg)
		require.Equal(t, "not_assessed", info.ComplianceAssessment)
		// 호환 필드의 타입과 기존 template 판정을 유지하되 감사 결과로 취급하지 않는다.
		require.Equal(t, template != "", info.Compliant)
		require.Zero(t, info.Issues)
		raw, err := json.Marshal(info)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(raw, &payload))
		require.Equal(t, "not_assessed", payload["compliance_assessment"])
		require.IsType(t, true, payload["compliant"])
		require.Equal(t, float64(0), payload["issues"])
		raw, err = yaml.Marshal(info)
		require.NoError(t, err)
		require.Contains(t, string(raw), "compliance_assessment: not_assessed")
	}
}
