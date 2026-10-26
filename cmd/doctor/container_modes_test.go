// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/cli"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

// 모든 미지원 조합은 실제 Cobra와 직접 진입점에서 외부 조회와 저장 전에 실패한다.
func TestContainerUnsupportedModes(t *testing.T) {
	for mask := 1; mask < 8; mask++ {
		for _, existing := range []bool{false, true} {
			for _, cobraPath := range []bool{false, true} {
				output := filepath.Join(t.TempDir(), "diagnostics.json")
				original := []byte("existing user bytes\n")
				if existing {
					require.NoError(t, os.WriteFile(output, original, 0o600))
				}
				calls := 0
				deps := containerDependencies{
					available: func(context.Context) bool { calls++; return true },
					collect: func(context.Context, containerOptions, logger.CommonLogger) (*ContainerDiagnostics, error) {
						calls++
						return &ContainerDiagnostics{}, nil
					},
				}
				opts := containerOptions{watchMode: mask&1 != 0, serverMode: mask&2 != 0, generateReport: mask&4 != 0, outputFile: output}
				var err error
				if cobraPath {
					command := newContainerCmdWithDependencies(deps)
					command.SetOut(io.Discard)
					command.SetErr(io.Discard)
					command.SilenceErrors, command.SilenceUsage = true, true
					args := []string{"--output", output}
					if opts.watchMode {
						args = append(args, "--watch")
					}
					if opts.serverMode {
						args = append(args, "--server")
					}
					if opts.generateReport {
						args = append(args, "--generate-report")
					}
					command.SetArgs(args)
					err = command.Execute()
				} else {
					err = runContainerDiagnosticsWithDependencies(context.Background(), &cli.CommonFlags{Format: "table"}, opts, deps)
				}
				require.ErrorContains(t, err, "not supported")
				require.Zero(t, calls)
				if existing {
					data, readErr := os.ReadFile(output)
					require.NoError(t, readErr)
					require.Equal(t, original, data)
				} else {
					_, statErr := os.Stat(output)
					require.ErrorIs(t, statErr, os.ErrNotExist)
				}
			}
		}
	}
	// 일반 직접 진입점도 Docker 유무와 관계없이 같은 모드를 거부한다.
	require.ErrorContains(t, runContainerDiagnostics(context.Background(), nil, containerOptions{watchMode: true}), "--watch")
}

// 비활성 모드의 설정값은 정상 단발 진단과 실제 파일 저장을 막지 않는다.
func TestContainerOneShotWithInactiveModeSettings(t *testing.T) {
	available, collected := 0, 0
	output := filepath.Join(t.TempDir(), "diagnostics.json")
	fixture := &ContainerDiagnostics{Recommendations: []string{"fixture recommendation"}}
	deps := containerDependencies{
		available: func(context.Context) bool { available++; return true },
		collect: func(_ context.Context, opts containerOptions, _ logger.CommonLogger) (*ContainerDiagnostics, error) {
			collected++
			require.False(t, opts.watchMode)
			require.False(t, opts.serverMode)
			require.False(t, opts.generateReport)
			require.Equal(t, 37*time.Second, opts.watchInterval)
			require.Equal(t, 12345, opts.serverPort)
			return fixture, nil
		},
	}
	command := newContainerCmdWithDependencies(deps)
	require.Equal(t, "false", command.Flags().Lookup("watch").DefValue)
	require.Equal(t, "false", command.Flags().Lookup("server").DefValue)
	require.Equal(t, "false", command.Flags().Lookup("generate-report").DefValue)
	command.SetArgs([]string{"--watch-interval", "37s", "--server-port", "12345", "--output", output})
	require.NoError(t, command.Execute())
	require.Equal(t, 1, available)
	require.Equal(t, 1, collected)
	data, err := os.ReadFile(output)
	require.NoError(t, err)
	var saved ContainerDiagnostics
	require.NoError(t, json.Unmarshal(data, &saved))
	require.Equal(t, fixture.Recommendations, saved.Recommendations)
}
