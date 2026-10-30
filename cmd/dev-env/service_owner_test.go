// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-dev-env/pkg/environment"
)

type ownerInvocation struct {
	Name string
	Args []string
}

func TestMain(m *testing.M) {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	if os.Getenv("GZ_ACCOUNT_FIXTURE") == "native-v1" && name == "gcloud" {
		os.Exit(runAccountBehaviorFixture())
	}
	if os.Getenv("GZ_MAIN_AWS_FIXTURE") == "native-v1" && name == "aws" {
		os.Exit(runAWSOwnerFixture())
	}
	if os.Getenv("GZ_SERVICE_OWNER_HELPER") == "native-v1" && (name == "gcloud" || name == "az" || name == "docker" || name == "kubectl") {
		data, err := json.Marshal(ownerInvocation{Name: name, Args: os.Args[1:]})
		if err != nil {
			os.Exit(90)
		}
		f, err := os.OpenFile(os.Getenv("GZ_SERVICE_OWNER_TRACE"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			os.Exit(91)
		}
		_, writeErr := f.Write(append(data, '\n'))
		closeErr := f.Close()
		if writeErr != nil || closeErr != nil {
			os.Exit(92)
		}
		key := strings.Join(os.Args[1:], "|")
		if key == os.Getenv("GZ_SERVICE_OWNER_FAIL") {
			os.Exit(23)
		}
		if os.Getenv("GZ_SERVICE_OWNER_BLANK") != "1" {
			_, _ = os.Stdout.WriteString("  value \n")
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func ownerFixture(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	bin := filepath.Join(home, "bin")
	require.NoError(t, os.Mkdir(bin, 0o700))
	exe, err := os.Executable()
	require.NoError(t, err)
	for _, name := range []string{"gcloud", "az", "docker", "kubectl"} {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		input, err := os.Open(exe)
		require.NoError(t, err)
		output, err := os.OpenFile(filepath.Join(bin, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
		require.NoError(t, err)
		_, copyErr := io.Copy(output, input)
		require.NoError(t, errors.Join(copyErr, output.Close(), input.Close()))
	}
	for _, key := range []string{"HOME", "USERPROFILE", "XDG_CONFIG_HOME", "APPDATA", "CLOUDSDK_CONFIG", "AZURE_CONFIG_DIR", "DOCKER_CONFIG", "KUBECONFIG"} {
		t.Setenv(key, filepath.Join(home, key))
	}
	t.Setenv("PATH", bin)
	t.Setenv("PATHEXT", ".EXE")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("GZ_SERVICE_OWNER_HELPER", "native-v1")
	trace := filepath.Join(home, "trace")
	t.Setenv("GZ_SERVICE_OWNER_TRACE", trace)
	return trace, home
}

func ownerTrace(t *testing.T, path string) []ownerInvocation {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	var result []ownerInvocation
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var item ownerInvocation
		require.NoError(t, json.Unmarshal([]byte(line), &item))
		result = append(result, item)
	}
	return result
}

func TestServiceOwnerAdapters(t *testing.T) {
	trace, _ := ownerFixture(t)
	cases := []struct {
		name, exe     string
		adapter       environment.ServiceSwitcher
		config, blank any
		queries, sets [][]string
	}{
		{"gcp", "gcloud", &GCPSwitcher{}, &environment.GCPConfig{Project: "value", Account: "value", Region: "value"}, &environment.GCPConfig{}, [][]string{{"config", "get-value", "project"}, {"config", "get-value", "account"}, {"config", "get-value", "compute/region"}}, [][]string{{"config", "set", "project", "value"}, {"config", "set", "account", "value"}, {"config", "set", "compute/region", "value"}}},
		{"azure", "az", &AzureSwitcher{}, &environment.AzureConfig{Subscription: "value", Tenant: "value"}, &environment.AzureConfig{}, [][]string{{"account", "show", "--query", "id", "-o", "tsv"}, {"account", "show", "--query", "tenantId", "-o", "tsv"}}, [][]string{{"account", "set", "--subscription", "value"}}},
		{"docker", "docker", &DockerSwitcher{}, &environment.DockerConfig{Context: "value"}, &environment.DockerConfig{}, [][]string{{"context", "show"}}, [][]string{{"context", "use", "value"}}},
		{"kubernetes", "kubectl", &KubernetesSwitcher{}, &environment.KubernetesConfig{Context: "value", Namespace: "value"}, &environment.KubernetesConfig{}, [][]string{{"config", "current-context"}, {"config", "view", "--minify", "--output", "jsonpath={..namespace}"}}, [][]string{{"config", "use-context", "value"}, {"config", "set-context", "--current", "--namespace", "value"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reset := func() { require.NoError(t, os.WriteFile(trace, nil, 0o600)) }
			expected := func(args [][]string) []ownerInvocation {
				items := make([]ownerInvocation, 0, len(args))
				for _, a := range args {
					items = append(items, ownerInvocation{Name: tc.exe, Args: a})
				}
				return items
			}
			require.Equal(t, tc.name, tc.adapter.Name())
			t.Run("manager query failure prevents mutation and journal", func(t *testing.T) {
				t.Setenv("GZ_SERVICE_OWNER_FAIL", strings.Join(tc.queries[0], "|"))
				reset()
				manager := environment.NewEnvironmentSwitcher()
				manager.Register(tc.adapter)
				service := environment.ServiceConfig{}
				switch config := tc.config.(type) {
				case *environment.GCPConfig:
					service.GCP = config
				case *environment.AzureConfig:
					service.Azure = config
				case *environment.DockerConfig:
					service.Docker = config
				case *environment.KubernetesConfig:
					service.Kubernetes = config
				}
				result, err := manager.SwitchEnvironment(context.Background(), &environment.Environment{Name: "fixture", Services: map[string]environment.ServiceConfig{tc.name: service}}, environment.SwitchOptions{RollbackOnError: true})
				require.Error(t, err)
				require.NotNil(t, result)
				require.False(t, result.Success)
				require.Equal(t, expected(tc.queries[:1]), ownerTrace(t, trace))
				journal, err := environment.DefaultJournalPath()
				require.NoError(t, err)
				_, err = os.Stat(journal)
				require.True(t, os.IsNotExist(err))
			})
			t.Run("missing executable", func(t *testing.T) {
				t.Setenv("PATH", t.TempDir())
				reset()
				state, err := tc.adapter.GetCurrentState(context.Background())
				require.Nil(t, state)
				require.Error(t, err)
				require.Empty(t, ownerTrace(t, trace))
			})

			reset()
			state, err := tc.adapter.GetCurrentState(context.Background())
			require.NoError(t, err)
			require.Equal(t, tc.config, state)
			require.Equal(t, expected(tc.queries), ownerTrace(t, trace))
			for _, call := range []func(context.Context, any) error{tc.adapter.Switch, tc.adapter.Rollback} {
				reset()
				require.NoError(t, call(context.Background(), tc.config))
				require.Equal(t, expected(tc.sets), ownerTrace(t, trace))
				reset()
				require.NoError(t, call(context.Background(), tc.blank))
				require.Empty(t, ownerTrace(t, trace))
				require.Error(t, call(context.Background(), nil))
				require.Error(t, call(context.Background(), "wrong"))
				require.Error(t, call(context.Background(), reflect.Zero(reflect.TypeOf(tc.blank)).Interface()))
				t.Setenv("GZ_SERVICE_OWNER_FAIL", strings.Join(tc.sets[0], "|"))
				reset()
				err = call(context.Background(), tc.config)
				var exit *exec.ExitError
				require.ErrorAs(t, err, &exit)
				require.Equal(t, 23, exit.ExitCode())
				require.Equal(t, expected(tc.sets[:1]), ownerTrace(t, trace))
				t.Setenv("GZ_SERVICE_OWNER_FAIL", "")
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				reset()
				require.ErrorIs(t, call(ctx, tc.config), context.Canceled)
				require.Empty(t, ownerTrace(t, trace))
			}
			t.Setenv("GZ_SERVICE_OWNER_BLANK", "1")
			reset()
			state, err = tc.adapter.GetCurrentState(context.Background())
			require.NoError(t, err)
			require.Equal(t, tc.blank, state)
			t.Setenv("GZ_SERVICE_OWNER_BLANK", "")
			for index, query := range tc.queries {
				t.Setenv("GZ_SERVICE_OWNER_FAIL", strings.Join(query, "|"))
				reset()
				state, err = tc.adapter.GetCurrentState(context.Background())
				require.Nil(t, state)
				var exit *exec.ExitError
				require.ErrorAs(t, err, &exit)
				require.Equal(t, 23, exit.ExitCode())
				require.Equal(t, expected(tc.queries[:index+1]), ownerTrace(t, trace))
			}
			t.Setenv("GZ_SERVICE_OWNER_FAIL", "")
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			reset()
			state, err = tc.adapter.GetCurrentState(ctx)
			require.Nil(t, state)
			require.ErrorIs(t, err, context.Canceled)
			require.Empty(t, ownerTrace(t, trace))
		})
	}
}
