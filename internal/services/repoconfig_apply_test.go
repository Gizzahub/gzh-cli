// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/env"
	"github.com/gizzahub/gzh-cli/pkg/github"
)

// 환경 조회를 시도하면 즉시 실패하므로 실제 credential fallback은 발생하지 않는다.
type forbiddenApplyEnvironment struct {
	env.Environment
	t *testing.T
}

func (e forbiddenApplyEnvironment) Get(string) string {
	e.t.Fatal("unsupported apply must fail before credential lookup")
	return ""
}

func TestRepositoryApplyUnsupportedBeforeCredentials(t *testing.T) {
	service := &RepoConfigService{environment: forbiddenApplyEnvironment{t: t}}
	for _, dryRun := range []bool{false, true} {
		for _, force := range []bool{false, true} {
			for _, interactive := range []bool{false, true} {
				err := service.ApplyConfiguration(context.Background(), ApplyOptions{Organization: "fixture", ConfigFile: "must-not-read.yaml", DryRun: dryRun, Force: force, Interactive: interactive})
				require.ErrorContains(t, err, "not supported")
			}
		}
	}
}

func TestRepositoryApplyAndPlanningNeverCallHTTP(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	service := NewRepoConfigServiceWithClient(github.NewRepoConfigClientWithBaseURL("fixture-not-secret", server.URL))
	ctx := context.Background()
	require.ErrorContains(t, service.ApplyConfiguration(ctx, ApplyOptions{Organization: "fixture", Token: "fixture-not-secret"}), "not supported")
	changes, err := service.GetConfigurationChanges(ctx, "fixture", ".*", "security")
	require.ErrorContains(t, err, "not supported")
	require.Nil(t, changes)
	changes, err = service.getConfigurationChanges(ctx, "fixture", "[", "security")
	require.ErrorContains(t, err, "not supported")
	require.Nil(t, changes)
	require.ErrorContains(t, service.applyConfigurationChange(ctx, ConfigurationChange{Repository: "fixture", Setting: "private"}), "not supported")
	for _, interactive := range []bool{false, true} {
		require.ErrorContains(t, service.applyChanges(ctx, nil, interactive), "not supported")
		require.ErrorContains(t, service.applyChanges(ctx, []ConfigurationChange{{Repository: "fixture"}}, interactive), "not supported")
	}
	require.Zero(t, calls.Load(), "unsupported entrypoints must never query or mutate remote repositories")
	require.Equal(t, 2, service.GetAffectedRepoCount([]ConfigurationChange{{Repository: "a"}, {Repository: "a"}, {Repository: "b"}}))
}
