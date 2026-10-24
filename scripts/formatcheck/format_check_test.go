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

func TestMain(m *testing.M) {
	// 테스트 실행 파일을 formatter 대역으로 재사용해 OS별 셸 스크립트를 피한다.
	if os.Getenv("FORMAT_CHECK_HELPER") == "1" {
		tool := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		fmt.Fprintln(os.Stderr, "invoked:"+tool)
		if tool == os.Getenv("FORMAT_CHECK_SELECTED") {
			switch os.Getenv("FORMAT_CHECK_MODE") {
			case "error":
				fmt.Fprintln(os.Stderr, "formatter diagnostic")
				os.Exit(23)
			case "difference":
				fmt.Println("unformatted.go")
			}
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestFormatCheckRecipe(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	recipe := filepath.ToSlash(filepath.Join(workingDirectory, "..", "..", ".make", "quality.mk"))
	executable, err := os.Executable()
	require.NoError(t, err)
	helper, err := os.ReadFile(executable)
	require.NoError(t, err)
	fixture := t.TempDir()
	tools := []string{"gofumpt", "gci", "goimports"}
	arguments := []string{"--no-print-directory", "-f", "Makefile", "format-check"}
	for _, tool := range tools {
		name := tool
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		path := filepath.Join(fixture, name)
		require.NoError(t, os.WriteFile(path, helper, 0o700))
		arguments = append(arguments, strings.ToUpper(tool)+"="+filepath.ToSlash(path))
	}
	// 실제 recipe를 읽고 무관한 도구 설치 전제만 대체한다.
	require.NoError(t, os.WriteFile(filepath.Join(fixture, "Makefile"), []byte("include "+recipe+"\ninstall-format-tools:\n\t@:\n"), 0o600))
	sentinel := filepath.Join(fixture, "unchanged.go")
	const contents = "package unchanged\n"
	require.NoError(t, os.WriteFile(sentinel, []byte(contents), 0o600))

	for _, tool := range tools {
		for _, mode := range []string{"clean", "error", "difference"} {
			t.Run(tool+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, "make", arguments...)
				command.Dir = fixture
				command.Stdin = strings.NewReader("")
				// 상위 make의 jobserver·추가 makefile 설정을 fixture에 전달하지 않는다.
				for _, variable := range os.Environ() {
					key, _, _ := strings.Cut(variable, "=")
					switch key {
					case "MAKEFLAGS", "MFLAGS", "MAKELEVEL", "MAKEFILES", "GNUMAKEFLAGS":
						continue
					}
					command.Env = append(command.Env, variable)
				}
				command.Env = append(command.Env, "FORMAT_CHECK_HELPER=1", "FORMAT_CHECK_SELECTED="+tool, "FORMAT_CHECK_MODE="+mode)
				output, runErr := command.CombinedOutput()
				require.NoError(t, ctx.Err(), "%s", output)
				if mode == "clean" {
					require.NoError(t, runErr, "%s", output)
					require.Contains(t, string(output), "Go files are properly formatted")
					for _, invoked := range tools {
						require.Contains(t, string(output), "invoked:"+invoked)
					}
				} else {
					require.Error(t, runErr, "%s", output)
					require.NotContains(t, string(output), "Go files are properly formatted")
					if mode == "error" {
						require.Contains(t, string(output), tool+" failed (exit 23)")
						require.Contains(t, string(output), "formatter diagnostic")
						for index, invoked := range tools {
							if invoked == tool {
								for _, later := range tools[index+1:] {
									require.NotContains(t, string(output), "invoked:"+later)
								}
							}
						}
					} else {
						require.Contains(t, string(output), "unformatted.go")
					}
				}
				actual, readErr := os.ReadFile(sentinel)
				require.NoError(t, readErr)
				require.Equal(t, contents, string(actual))
			})
		}
	}
}
