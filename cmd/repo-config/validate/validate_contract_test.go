// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package validate

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func validationCommandOutput(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var output bytes.Buffer
	cmd := NewCmd()
	cmd.SetOut(&output)
	cmd.SetErr(io.Discard)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

func TestValidationContractRealLoader(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		valid         bool
	}{
		{name: "valid", content: "version: '1'\norganization: org\n", valid: true},
		{name: "syntax", content: "templates: [invalid"},
		{name: "version", content: "organization: org\n"},
		{name: "organization", content: "version: '1'\n"},
		{name: "null", content: "version: '1'\norganization: org\ntemplates:\n  absent: null\n"},
		{name: "baseNull", content: "version: '1'\norganization: org\ntemplates:\n  derived: {base: absent}\n  absent: null\n"},
		{name: "baseMissing", content: "version: '1'\norganization: org\ntemplates:\n  derived: {base: missing}\n"},
		{name: "circular", content: "version: '1'\norganization: org\ntemplates:\n  first: {base: second}\n  second: {base: first}\n"},
		{name: "pattern", content: "version: '1'\norganization: org\nrepositories:\n  patterns: [{template: any}]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(file, []byte(tc.content), 0o600))
			for _, format := range []string{"table", "json", "yaml"} {
				out, err := validationCommandOutput(t, "--config-file", file, "--format", format, "--verbose", "--token", "unused")
				if tc.valid {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
				var summary ValidationSummary
				switch format {
				case "json":
					require.NoError(t, json.Unmarshal([]byte(out), &summary))
				case "yaml":
					require.NoError(t, yaml.Unmarshal([]byte(out), &summary))
				default:
					require.Contains(t, out, "Local configuration loader")
					continue
				}
				require.Equal(t, tc.valid, summary.Valid)
				require.Len(t, summary.Checks, 1)
				require.Zero(t, summary.Warnings)
				require.Zero(t, summary.Checks[0].Line)
				require.NotContains(t, out, "enterprise")
				require.NotContains(t, out, "GitHub API Compatibility")
			}
			after, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, tc.content, string(after))
		})
	}
}

func TestValidationContractMissingAndStrict(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := validationCommandOutput(t)
	require.ErrorContains(t, err, "no configuration file")
	missing := filepath.Join(t.TempDir(), "missing")
	out, err := validationCommandOutput(t, "--config-file", missing, "--format", "json")
	require.ErrorIs(t, err, os.ErrNotExist)
	var summary ValidationSummary
	require.NoError(t, json.Unmarshal([]byte(out), &summary))
	require.False(t, summary.Valid)
	out, err = validationCommandOutput(t, "--strict", "--config-file", missing)
	require.ErrorContains(t, err, "unsupported")
	require.NotErrorIs(t, err, os.ErrNotExist)
	require.Empty(t, out)
	summary, err = performValidation(missing, true)
	require.ErrorContains(t, err, "unsupported")
	require.Empty(t, summary.Checks)
	out, err = validationCommandOutput(t, "--format", "invalid", "--config-file", missing)
	require.ErrorContains(t, err, "unsupported format")
	require.Empty(t, out)
}

func TestValidationContractDiscoveryOrder(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	names := []string{".gzh/repo-config.yaml", "repo-config.yaml", "gzh-repo-config.yaml"}
	for _, name := range names {
		require.NoError(t, os.MkdirAll(filepath.Dir(name), 0o700))
		require.NoError(t, os.WriteFile(name, []byte("version: '1'\norganization: org\n"), 0o600))
	}
	for _, name := range names {
		out, err := validationCommandOutput(t, "--format", "json")
		require.NoError(t, err)
		var summary ValidationSummary
		require.NoError(t, json.Unmarshal([]byte(out), &summary))
		require.Equal(t, name, summary.File)
		require.NoError(t, os.Remove(name))
	}
}

type validationFailWriter struct{ err error }

func (w validationFailWriter) Write([]byte) (int, error) { return 0, w.err }

func TestValidationContractRenderErrorPriority(t *testing.T) {
	sentinel := errors.New("controlled writer failure")
	missing := filepath.Join(t.TempDir(), "missing")
	for _, format := range []string{"table", "json", "yaml"} {
		err := runValidateCommandTo(validationFailWriter{err: sentinel}, GlobalFlags{ConfigFile: missing}, false, format)
		require.ErrorIs(t, err, sentinel)
		require.NotErrorIs(t, err, os.ErrNotExist)
	}
	// 기존 비공개 진입점도 strict 요청에서 출력이나 파일 접근 없이 실패한다.
	require.ErrorContains(t, runValidateCommand(GlobalFlags{ConfigFile: missing}, true, "table"), "unsupported")
}
