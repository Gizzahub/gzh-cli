// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-dev-env/pkg/environment"
)

type reportingFailureService struct {
	switchErr   error
	rollbackErr error
	calls       []string
}

func (s *reportingFailureService) Name() string { return "ssh" }
func (s *reportingFailureService) GetCurrentState(context.Context) (any, error) {
	s.calls = append(s.calls, "state")
	return &environment.SSHConfig{Config: "previous"}, nil
}

func (s *reportingFailureService) Switch(context.Context, any) error {
	s.calls = append(s.calls, "switch")
	return s.switchErr
}

func (s *reportingFailureService) Rollback(context.Context, any) error {
	s.calls = append(s.calls, "rollback")
	return s.rollbackErr
}

func TestSwitchReportingRealManagerFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	t.Setenv("APPDATA", filepath.Join(home, "config"))
	t.Setenv("PATH", t.TempDir())
	original := errors.New("fixture switch failed")
	rollback := errors.New("fixture rollback failed")
	service := &reportingFailureService{switchErr: original, rollbackErr: rollback}
	manager := environment.NewEnvironmentSwitcher()
	manager.Register(service)
	result, switchErr := manager.SwitchEnvironment(context.Background(), &environment.Environment{Name: "fixture", Services: map[string]environment.ServiceConfig{"ssh": {SSH: &environment.SSHConfig{Config: "next"}}}}, environment.SwitchOptions{RollbackOnError: true})
	require.ErrorIs(t, switchErr, original)
	require.NotNil(t, result)
	require.True(t, result.RollbackPerformed)
	require.Equal(t, []string{"state", "switch", "rollback"}, service.calls)
	output, err := captureSSHSaveOutput(t, func() error { return (&switchAllOptions{}).finishSwitch("fixture", result, switchErr) })
	require.ErrorIs(t, err, original)
	require.Contains(t, err.Error(), "environment switch failed:")
	require.Contains(t, output, original.Error())
	require.Contains(t, output, rollback.Error())
	require.Contains(t, output, "Rollback: Attempted")
	require.NotContains(t, output, "Successfully switched")
}

func TestSwitchReportingResultBoundaries(t *testing.T) {
	original := errors.New("no result failure")
	for _, tc := range []struct {
		name    string
		result  *environment.SwitchResult
		cause   error
		want    string
		success bool
	}{
		{name: "success", result: &environment.SwitchResult{Success: true}, success: true},
		{name: "false success", result: &environment.SwitchResult{}, want: "completed with errors"},
		{name: "nil failure", cause: original, want: original.Error()},
		{name: "nil nil", want: "returned no result"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := captureSSHSaveOutput(t, func() error { return (&switchAllOptions{}).finishSwitch("fixture", tc.result, tc.cause) })
			if tc.success {
				require.NoError(t, err)
				require.Contains(t, output, "Successfully switched to environment: fixture")
			} else {
				require.ErrorContains(t, err, tc.want)
				require.NotContains(t, output, "Successfully switched")
			}
			if tc.cause != nil {
				require.ErrorIs(t, err, tc.cause)
			}
		})
	}
}

func TestSwitchReportingHelpContract(t *testing.T) {
	cmd := newSwitchAllCmd()
	require.NotContains(t, cmd.Short, "atomically")
	require.NotContains(t, cmd.Long, "all are rolled back")
	require.Contains(t, cmd.Long, "rollback is attempted")
	require.Contains(t, cmd.Long, "Rollback can also fail")
}

func TestSwitchReportingRunDisplaysOwnerFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	t.Setenv("APPDATA", filepath.Join(home, "config"))
	t.Setenv("PATH", t.TempDir())
	active := filepath.Join(home, ".ssh", "config")
	require.NoError(t, os.MkdirAll(filepath.Dir(active), 0o700))
	require.NoError(t, os.WriteFile(active, []byte("user-owned SSH configuration"), 0o600))
	target := filepath.Join(home, "target")
	require.NoError(t, os.WriteFile(target, []byte("target SSH configuration"), 0o600))
	// JSON 문자열은 YAML 경로에서도 역슬래시와 공백을 안전하게 표현한다.
	quoted, err := json.Marshal(target)
	require.NoError(t, err)
	config := filepath.Join(home, "environment.yaml")
	require.NoError(t, os.WriteFile(config, []byte("name: fixture\nservices:\n  ssh:\n    ssh:\n      config: "+string(quoted)+"\n"), 0o600))
	opts := &switchAllOptions{fromFile: config, force: true, timeout: time.Second}
	output, err := captureSSHSaveOutput(t, func() error { return opts.run(context.Background()) })
	require.ErrorContains(t, err, "environment switch failed:")
	require.ErrorContains(t, err, "refusing to overwrite")
	require.NotNil(t, errors.Unwrap(err))
	require.Contains(t, output, "Switch Results:")
	require.Contains(t, output, "Success: false")
	require.Contains(t, output, "refusing to overwrite")
	require.NotContains(t, output, "Successfully switched")
	data, readErr := os.ReadFile(active)
	require.NoError(t, readErr)
	require.Equal(t, "user-owned SSH configuration", string(data))
	info, statErr := os.Lstat(active)
	require.NoError(t, statErr)
	require.True(t, info.Mode().IsRegular())
}
