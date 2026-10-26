// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTemplateContractDirectUnsupported(t *testing.T) {
	for _, name := range []string{"security", "microservice", "unknown"} {
		for _, strict := range []bool{false, true} {
			finish := captureTemplateStdout(t)
			err := runTemplateValidateCommand(name, strict)
			output := finish()
			require.ErrorContains(t, err, "not supported")
			require.Empty(t, output)
		}
		finish := captureTemplateStdout(t)
		err := runTemplateShowCommand(name, formatJSON)
		output := finish()
		require.Error(t, err)
		require.Empty(t, output)
		if name == "unknown" {
			require.ErrorContains(t, err, "not found")
		} else {
			require.ErrorContains(t, err, "JSON template output is not supported")
		}
	}
}

// 내장 catalog와 YAML 예제는 원격 검증 결과가 아닌 정적 자료로 유지한다.
func TestTemplateContractMachinePayloads(t *testing.T) {
	finish := captureTemplateStdout(t)
	err := runTemplateListCommand(formatJSON)
	output := finish()
	require.NoError(t, err)
	var catalog struct {
		Templates []ConfigTemplate `json:"templates"`
		Count     int              `json:"count"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &catalog))
	require.Equal(t, 5, catalog.Count)
	require.Len(t, catalog.Templates, 5)
	names := []string{}
	for _, entry := range catalog.Templates {
		names = append(names, entry.Name)
	}
	require.Equal(t, []string{"security", "microservice", "frontend", "library", "opensource"}, names)
	for _, name := range []string{"security", "microservice"} {
		finish = captureTemplateStdout(t)
		err = runTemplateShowCommand(name, formatYAML)
		output = finish()
		require.NoError(t, err)
		var example struct {
			Name     string `yaml:"name"`
			Settings struct {
				Visibility string `yaml:"visibility"`
				Security   struct {
					BranchProtection map[string]struct {
						RequiredReviews int `yaml:"required_reviews"`
					} `yaml:"branch_protection"`
				} `yaml:"security"`
			} `yaml:"settings"`
		}
		require.NoError(t, yaml.Unmarshal([]byte(output), &example))
		require.Equal(t, name, example.Name)
		require.Equal(t, "private", example.Settings.Visibility)
		require.Contains(t, example.Settings.Security.BranchProtection, "main")
		require.Equal(t, 2, example.Settings.Security.BranchProtection["main"].RequiredReviews)
	}
	finish = captureTemplateStdout(t)
	err = runTemplateListCommand(formatTable)
	output = finish()
	require.NoError(t, err)
	require.Contains(t, output, "Available Repository Configuration Templates")
	require.Contains(t, output, "Total templates: 5")
	require.Empty(t, getTemplateConfiguration("frontend"), "catalog entry does not invent a YAML sample")
	require.ErrorContains(t, runTemplateShowCommand("frontend", formatYAML), "not found")
}

func TestTemplateContractWriterFailureVisible(t *testing.T) {
	original := os.Stdout
	closed, err := os.CreateTemp(t.TempDir(), "closed-writer")
	require.NoError(t, err)
	require.NoError(t, closed.Close())
	os.Stdout = closed
	t.Cleanup(func() { os.Stdout = original })
	require.ErrorIs(t, runTemplateListCommand(formatJSON), os.ErrClosed)
	require.ErrorIs(t, runTemplateShowCommand("security", formatYAML), os.ErrClosed)
}

// 병렬 실행하지 않으며 fmt 출력도 pipe로 수집하고 조기 실패 시 복원한다.
func captureTemplateStdout(t *testing.T) func() string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	type captured struct {
		data []byte
		err  error
	}
	done := make(chan captured, 1)
	go func() { data, readErr := io.ReadAll(reader); done <- captured{data, readErr} }()
	finished := false
	var output string
	finish := func() string {
		if finished {
			return output
		}
		finished = true
		os.Stdout = original
		closeErr := writer.Close()
		result := <-done
		readerErr := reader.Close()
		output = string(result.data)
		require.NoError(t, closeErr)
		require.NoError(t, result.err)
		require.NoError(t, readerErr)
		return output
	}
	t.Cleanup(func() { finish() })
	return finish
}

func TestTemplateContractCobraPreserved(t *testing.T) {
	for _, args := range [][]string{{"validate", "security"}, {"validate", "security", "--strict"}, {"validate", "unknown"}, {"show", "security", "--format", "json"}, {"show", "unknown", "--format", "json"}, {"show"}, {"validate"}} {
		command := newTemplateCmd()
		command.SilenceErrors, command.SilenceUsage = true, true
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		command.SetArgs(args)
		finish := captureTemplateStdout(t)
		err := command.Execute()
		output := finish()
		require.Error(t, err)
		require.Empty(t, output)
	}
	command := newTemplateCmd()
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"--help"})
	require.NoError(t, command.Execute())
	show, _, err := command.Find([]string{"show"})
	require.NoError(t, err)
	require.Equal(t, "yaml", show.Flags().Lookup("format").DefValue)
	validate, _, err := command.Find([]string{"validate"})
	require.NoError(t, err)
	require.Equal(t, "false", validate.Flags().Lookup("strict").DefValue)
	command = newTemplateCmd()
	command.SetArgs([]string{"list", "--format", "json"})
	finish := captureTemplateStdout(t)
	err = command.Execute()
	output := finish()
	require.NoError(t, err)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &parsed))
	require.Equal(t, float64(5), parsed["count"])
}
