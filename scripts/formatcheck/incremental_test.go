package formatcheck

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIncrementalFormatterRecipes(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	recipe := filepath.ToSlash(filepath.Join(workingDirectory, "..", "..", ".make", "quality.mk"))
	executable, err := os.Executable()
	require.NoError(t, err)
	helper, err := os.ReadFile(executable)
	require.NoError(t, err)
	fixture := t.TempDir()
	tools := []string{"gofumpt", "goimports", "gci"}
	variables := []string{}
	for _, tool := range tools {
		name := tool
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		path := filepath.Join(fixture, name)
		require.NoError(t, os.WriteFile(path, helper, 0o700))
		variables = append(variables, strings.ToUpper(tool)+"="+filepath.ToSlash(path))
	}
	require.NoError(t, os.WriteFile(filepath.Join(fixture, "Makefile"), []byte("include "+recipe+"\ninstall-format-tools:\n\t@:\n"), 0o600))
	files := []string{"first.go", "second.go"}
	for _, file := range files {
		require.NoError(t, os.WriteFile(filepath.Join(fixture, file), []byte("package fixture\n"), 0o600))
	}
	// fmt-diff가 실제 dirty tracked 파일을 선택하도록 독립 Git 저장소를 만든다.
	runFixtureCommand(t, fixture, nil, "git", "init", "--quiet")
	runFixtureCommand(t, fixture, nil, "git", "add", "--", files[0], files[1])
	runFixtureCommand(t, fixture, nil, "git", "-c", "user.name=Formatter fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "fixture")
	for _, file := range files {
		require.NoError(t, os.WriteFile(filepath.Join(fixture, file), []byte("package fixture\n\n// dirty tracked fixture\n"), 0o600))
	}
	trace := filepath.Join(fixture, "invocations.log")
	for _, target := range []string{"format-file", "fmt-diff"} {
		successMessages := []string{"formatted successfully", "All files processed"}
		if target == "fmt-diff" {
			successMessages = []string{"Changed files formatted"}
		}
		for _, selected := range tools {
			for _, mode := range []string{"clean", "error"} {
				t.Run(target+"/"+selected+"/"+mode, func(t *testing.T) {
					require.NoError(t, os.WriteFile(trace, nil, 0o600))
					arguments := []string{"--no-print-directory", "-f", "Makefile", target}
					if target == "format-file" {
						arguments = append(arguments, files...)
					}
					arguments = append(arguments, variables...)
					output, runErr := fixtureCommand(t, fixture, []string{"FORMAT_CHECK_HELPER=1", "FORMAT_CHECK_TRACE=" + trace, "FORMAT_CHECK_SELECTED=" + selected, "FORMAT_CHECK_MODE=" + mode}, "make", arguments...)
					actual, readErr := os.ReadFile(trace)
					require.NoError(t, readErr)
					var expected []string
					for _, file := range files {
						for _, tool := range tools {
							expected = append(expected, tool+":"+file)
							if mode == "error" && tool == selected {
								break
							}
						}
						if mode == "error" {
							break
						}
					}
					if mode == "error" {
						require.Error(t, runErr, "%s", output)
						require.Contains(t, string(output), selected+" failed for first.go (exit 23)")
						require.Contains(t, string(output), "formatter diagnostic")
						for _, message := range successMessages {
							require.NotContains(t, string(output), message)
						}
					} else {
						require.NoError(t, runErr, "%s", output)
						for _, message := range successMessages {
							require.Contains(t, string(output), message)
						}
					}
					require.Equal(t, expected, strings.Split(strings.TrimSpace(string(actual)), "\n"))
				})
			}
		}
	}
}

func runFixtureCommand(t *testing.T, directory string, environment []string, executable string, arguments ...string) {
	t.Helper()
	output, err := fixtureCommand(t, directory, environment, executable, arguments...)
	require.NoError(t, err, "%s", output)
}

func fixtureCommand(t *testing.T, directory string, environment []string, executable string, arguments ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, arguments...)
	command.Dir = directory
	command.Stdin = strings.NewReader("")
	// 상위 make와 사용자 Git 설정이 fixture의 선택·커밋에 개입하지 않게 한다.
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		if strings.HasPrefix(key, "GIT_") || strings.HasPrefix(key, "FORMAT_CHECK_") {
			continue
		}
		switch key {
		case "MAKEFLAGS", "MFLAGS", "MAKELEVEL", "MAKEFILES", "GNUMAKEFLAGS":
			continue
		}
		command.Env = append(command.Env, variable)
	}
	command.Env = append(command.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_COUNT=2", "GIT_CONFIG_KEY_0=core.hooksPath", "GIT_CONFIG_VALUE_0="+filepath.Join(directory, "no-hooks"), "GIT_CONFIG_KEY_1=commit.gpgsign", "GIT_CONFIG_VALUE_1=false")
	command.Env = append(command.Env, environment...)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return output, fmt.Errorf("fixture command timed out: %w", ctx.Err())
	}
	return output, err
}
