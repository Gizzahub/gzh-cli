// Copyright (c) 2026 Gizzahub
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
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type accountFixtureCall struct {
	Args      []string
	Output    string
	Exit      int
	WriteFile string
}

// 이름이 gcloud인 소유 test binary만 재실행하며 실제 SDK나 계정은 사용하지 않는다.
func runAccountBehaviorFixture() int {
	data, err := os.ReadFile(os.Getenv("GZ_ACCOUNT_CASES"))
	if err != nil {
		return 90
	}
	var cases []accountFixtureCall
	if json.Unmarshal(data, &cases) != nil {
		return 91
	}
	f, err := os.OpenFile(os.Getenv("GZ_ACCOUNT_TRACE"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return 92
	}
	err = json.NewEncoder(f).Encode(ownerInvocation{Name: "gcloud", Args: os.Args[1:]})
	if errors.Join(err, f.Close()) != nil {
		return 93
	}
	actual, err := json.Marshal(os.Args[1:])
	if err != nil {
		return 94
	}
	for _, c := range cases {
		expected, marshalErr := json.Marshal(c.Args)
		if marshalErr != nil {
			return 95
		}
		if string(expected) != string(actual) {
			continue
		}
		if c.WriteFile != "" {
			if os.WriteFile(c.WriteFile, []byte("synthetic-key-fixture"), 0o644) != nil {
				return 96
			}
		}
		if _, err := os.Stdout.WriteString(c.Output); err != nil {
			return 97
		}
		return c.Exit
	}
	return 98
}

func accountBehaviorFixture(t *testing.T) (string, func(...accountFixtureCall)) {
	t.Helper()
	home := t.TempDir()
	bin := filepath.Join(home, "bin")
	require.NoError(t, os.Mkdir(bin, 0o700))
	exe, err := os.Executable()
	require.NoError(t, err)
	name := "gcloud"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	input, err := os.Open(exe)
	require.NoError(t, err)
	output, err := os.OpenFile(filepath.Join(bin, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	require.NoError(t, err)
	_, copyErr := io.Copy(output, input)
	require.NoError(t, errors.Join(copyErr, input.Close(), output.Close()))
	setTestHome(t, home)
	t.Setenv("PATH", bin)
	t.Setenv("PATHEXT", ".EXE")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	t.Setenv("CLOUDSDK_CONFIG", filepath.Join(home, "config"))
	t.Setenv("GZ_ACCOUNT_FIXTURE", "native-v1")
	trace, cases := filepath.Join(home, "trace"), filepath.Join(home, "cases.json")
	t.Setenv("GZ_ACCOUNT_TRACE", trace)
	t.Setenv("GZ_ACCOUNT_CASES", cases)
	return trace, func(calls ...accountFixtureCall) {
		data, err := json.Marshal(calls)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(cases, data, 0o600))
		require.NoError(t, os.WriteFile(trace, nil, 0o600))
	}
}

func captureAccountBehaviorOutput(t *testing.T, action func() error) (string, error) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "output")
	require.NoError(t, err)
	previous := os.Stdout
	os.Stdout = f
	defer func() { os.Stdout = previous; _ = f.Close() }()
	actionErr := action()
	os.Stdout = previous
	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)
	data, err := io.ReadAll(f)
	require.NoError(t, err)
	return string(data), actionErr
}

func TestAccountBehaviorLoadsMetadataAndKeyCounts(t *testing.T) {
	trace, set := accountBehaviorFixture(t)
	set(
		accountFixtureCall{Args: []string{"config", "get-value", "project"}, Output: "  fixture-project \n"},
		accountFixtureCall{Args: []string{"iam", "service-accounts", "list", "--project", "fixture-project", "--format=json"}, Output: `[{"email":"one@example.invalid","name":"projects/p/serviceAccounts/one","displayName":"One","projectId":"fixture-project","uniqueId":"17","description":"fixture","oauth2ClientId":"client"},{"email":"two@example.invalid","disabled":true}]`},
		accountFixtureCall{Args: []string{"auth", "list", "--filter=status:ACTIVE", "--format=value(account)"}, Output: " one@example.invalid\n"},
		accountFixtureCall{Args: []string{"iam", "service-accounts", "keys", "list", "--iam-account", "one@example.invalid", "--project", "fixture-project", "--format=json"}, Output: `[{"name":"first"},{"name":"second"}]`},
		accountFixtureCall{Args: []string{"iam", "service-accounts", "keys", "list", "--iam-account", "two@example.invalid", "--project", "fixture-project", "--format=json"}, Output: `[]`},
	)
	m, err := NewGCPServiceAccountManager(t.Context(), "")
	require.NoError(t, err)
	require.Equal(t, "fixture-project", m.projectID)
	require.Len(t, m.serviceAccounts, 2)
	one, two := m.serviceAccounts["one@example.invalid"], m.serviceAccounts["two@example.invalid"]
	require.Equal(t, "One", one.DisplayName)
	require.Equal(t, "17", one.UniqueID)
	require.Equal(t, "fixture", one.Description)
	require.Equal(t, 2, one.KeyCount)
	require.True(t, one.IsActive)
	require.False(t, two.IsActive)
	require.True(t, two.Disabled)
	observed := ownerTrace(t, trace)
	require.Len(t, observed, 5)
	require.Equal(t, []string{"config", "get-value", "project"}, observed[0].Args)
	text, err := captureAccountBehaviorOutput(t, func() error { return m.listServiceAccounts("json") })
	require.NoError(t, err)
	var accounts []GCPServiceAccount
	require.NoError(t, json.Unmarshal([]byte(text), &accounts))
	require.Len(t, accounts, 2)
	_, err = captureAccountBehaviorOutput(t, func() error { return m.listServiceAccounts("table") })
	require.NoError(t, err)
	require.ErrorContains(t, m.listServiceAccounts("unsupported"), "unsupported output format")
	require.ErrorContains(t, m.showServiceAccount("missing", "json"), "not found")
	text, err = captureAccountBehaviorOutput(t, func() error { return m.showServiceAccount(one.Email, "json") })
	require.NoError(t, err)
	var shown GCPServiceAccount
	require.NoError(t, json.Unmarshal([]byte(text), &shown))
	require.Equal(t, *one, shown)
	require.Len(t, ownerTrace(t, trace), 5)
}

func TestAccountBehaviorFailsBeforeMutationOnQueryErrors(t *testing.T) {
	trace, set := accountBehaviorFixture(t)
	set(accountFixtureCall{Args: []string{"config", "get-value", "project"}, Exit: 23})
	_, err := NewGCPServiceAccountManager(t.Context(), "")
	var exit *exec.ExitError
	require.ErrorAs(t, err, &exit)
	require.Equal(t, 23, exit.ExitCode())
	require.Len(t, ownerTrace(t, trace), 1)
	listArgs := []string{"iam", "service-accounts", "list", "--project", "p", "--format=json"}
	set(accountFixtureCall{Args: listArgs, Exit: 23})
	_, err = NewGCPServiceAccountManager(t.Context(), "p")
	require.ErrorContains(t, err, "failed to list service accounts")
	require.ErrorAs(t, err, &exit)
	set(accountFixtureCall{Args: listArgs, Output: "{"})
	_, err = NewGCPServiceAccountManager(t.Context(), "p")
	require.ErrorContains(t, err, "failed to parse service accounts JSON")
	require.Len(t, ownerTrace(t, trace), 1)
	set()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = NewGCPServiceAccountManager(ctx, "p")
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, ownerTrace(t, trace))
}

func TestAccountBehaviorMutationArgumentsAndActivationState(t *testing.T) {
	trace, set := accountBehaviorFixture(t)
	m := &GCPServiceAccountManager{projectID: "p", ctx: t.Context(), serviceAccounts: map[string]*GCPServiceAccount{}}
	create := []string{"iam", "service-accounts", "create", "robot", "--project", "p", "--display-name", "Robot name", "--description", "fixture description"}
	set(accountFixtureCall{Args: create}, accountFixtureCall{Args: []string{"iam", "service-accounts", "list", "--project", "p", "--format=json"}, Output: "[]"}, accountFixtureCall{Args: []string{"auth", "list", "--filter=status:ACTIVE", "--format=value(account)"}})
	_, err := captureAccountBehaviorOutput(t, func() error { return m.createServiceAccount("robot", "Robot name", "fixture description") })
	require.NoError(t, err)
	require.Equal(t, create, ownerTrace(t, trace)[0].Args)
	one, two := &GCPServiceAccount{Email: "one", IsActive: true}, &GCPServiceAccount{Email: "two"}
	m.serviceAccounts = map[string]*GCPServiceAccount{"one": one, "two": two}
	activate := []string{"auth", "activate-service-account", "two", "--key-file", "owned-key.json"}
	set(accountFixtureCall{Args: activate})
	_, err = captureAccountBehaviorOutput(t, func() error { return m.activateServiceAccount("two", "owned-key.json") })
	require.NoError(t, err)
	require.False(t, one.IsActive)
	require.True(t, two.IsActive)
	require.Equal(t, activate, ownerTrace(t, trace)[0].Args)
	set(accountFixtureCall{Args: []string{"config", "set", "account", "one"}, Exit: 23})
	err = m.activateServiceAccount("one", "")
	require.ErrorContains(t, err, "failed to set active account")
	require.False(t, one.IsActive)
	require.True(t, two.IsActive)
	remove := []string{"iam", "service-accounts", "delete", "one", "--project", "p", "--quiet"}
	set(accountFixtureCall{Args: remove})
	_, err = captureAccountBehaviorOutput(t, func() error { return m.deleteServiceAccount("one", true) })
	require.NoError(t, err)
	require.Equal(t, remove, ownerTrace(t, trace)[0].Args)
}

func TestAccountBehaviorKeyCommandsUseOwnedFilesAndPrivateMode(t *testing.T) {
	trace, set := accountBehaviorFixture(t)
	m := &GCPServiceAccountManager{projectID: "p", ctx: t.Context(), serviceAccounts: map[string]*GCPServiceAccount{}}
	path := filepath.Join(t.TempDir(), "key.json")
	create := []string{"iam", "service-accounts", "keys", "create", path, "--iam-account", "one", "--project", "p", "--key-file-type", "json"}
	set(accountFixtureCall{Args: create, WriteFile: path})
	_, err := captureAccountBehaviorOutput(t, func() error { return m.createServiceAccountKey("one", "json", path) })
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assertPrivateMode(t, info, 0o600)
	require.Equal(t, create, ownerTrace(t, trace)[0].Args)
	remove := []string{"iam", "service-accounts", "keys", "delete", "key-id", "--iam-account", "one", "--project", "p", "--quiet"}
	set(accountFixtureCall{Args: remove, Exit: 23})
	require.ErrorContains(t, m.deleteServiceAccountKey("one", "key-id", true), "failed to delete service account key")
	set(accountFixtureCall{Args: remove})
	_, err = captureAccountBehaviorOutput(t, func() error { return m.deleteServiceAccountKey("one", "key-id", true) })
	require.NoError(t, err)
	require.Equal(t, remove, ownerTrace(t, trace)[0].Args)
	list := []string{"iam", "service-accounts", "keys", "list", "--iam-account", "one", "--project", "p", "--format=json"}
	set(accountFixtureCall{Args: list, Output: `[{"name":"projects/p/serviceAccounts/one/keys/key-id","privateKeyType":"TYPE_JSON","keyAlgorithm":"RSA","keyOrigin":"USER","validAfterTime":"2026-01-01","validBeforeTime":"2027-01-01"}]`})
	text, err := captureAccountBehaviorOutput(t, func() error { return m.listServiceAccountKeys("one") })
	require.NoError(t, err)
	require.Contains(t, text, "key-id")
	set(accountFixtureCall{Args: list, Output: "{"})
	require.ErrorContains(t, m.listServiceAccountKeys("one"), "failed to parse keys JSON")
}

func TestAccountBehaviorMutationFailuresKeepCachedState(t *testing.T) {
	trace, set := accountBehaviorFixture(t)
	active := &GCPServiceAccount{Email: "one", IsActive: true}
	m := &GCPServiceAccountManager{projectID: "p", ctx: t.Context(), serviceAccounts: map[string]*GCPServiceAccount{"one": active}}
	path := filepath.Join(t.TempDir(), "key.json")
	cases := []struct {
		name string
		args []string
		run  func() error
	}{
		{"create", []string{"iam", "service-accounts", "create", "robot", "--project", "p"}, func() error { return m.createServiceAccount("robot", "", "") }},
		{"delete", []string{"iam", "service-accounts", "delete", "one", "--project", "p", "--quiet"}, func() error { return m.deleteServiceAccount("one", true) }},
		{"create key", []string{"iam", "service-accounts", "keys", "create", path, "--iam-account", "one", "--project", "p", "--key-file-type", "json"}, func() error { return m.createServiceAccountKey("one", "json", path) }},
		{"activate key", []string{"auth", "activate-service-account", "other", "--key-file", path}, func() error { return m.activateServiceAccount("other", path) }},
		{"list keys", []string{"iam", "service-accounts", "keys", "list", "--iam-account", "one", "--project", "p", "--format=json"}, func() error { return m.listServiceAccountKeys("one") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set(accountFixtureCall{Args: tc.args, Exit: 23})
			var exit *exec.ExitError
			require.ErrorAs(t, tc.run(), &exit)
			require.Equal(t, 23, exit.ExitCode())
			require.Equal(t, []ownerInvocation{{Name: "gcloud", Args: tc.args}}, ownerTrace(t, trace))
			require.True(t, active.IsActive)
			_, err := os.Stat(path)
			require.True(t, os.IsNotExist(err))
		})
	}
}
