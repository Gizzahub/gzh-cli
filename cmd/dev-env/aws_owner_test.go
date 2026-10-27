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
	"strings"
	"testing"

	"github.com/gizzahub/gzh-cli-dev-env/pkg/environment"
)

var (
	_ environment.ServiceSwitcher   = (&AWSSwitcher{})
	_ environment.StateCapturer     = (&AWSSwitcher{})
	_ environment.StateApplier      = (&AWSSwitcher{})
	_ environment.RecoveryValidator = (&AWSSwitcher{})
)

func awsOwnerCheck(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func awsOwnerIsolated(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	for _, key := range []string{"HOME", "USERPROFILE", "XDG_CONFIG_HOME", "APPDATA", "LOCALAPPDATA"} {
		t.Setenv(key, home)
	}
	t.Setenv("PATH", t.TempDir())
	for _, key := range []string{"AWS_PROFILE", "AWS_DEFAULT_PROFILE", "AWS_REGION", "AWS_DEFAULT_REGION", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_WEB_IDENTITY_TOKEN_FILE", "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "AWS_CONTAINER_CREDENTIALS_FULL_URI", "AWS_CONTAINER_AUTHORIZATION_TOKEN"} {
		t.Setenv(key, "")
	}
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(home, "aws-config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(home, "credentials"))
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	config, err := os.UserConfigDir()
	awsOwnerCheck(t, err)
	return filepath.Join(config, "gzh-dev-env", "aws-active-profile")
}

func awsOwnerSeed(t *testing.T, path, raw string) {
	t.Helper()
	awsOwnerCheck(t, os.MkdirAll(filepath.Dir(path), 0o700))
	awsOwnerCheck(t, os.WriteFile(path, []byte(raw), 0o640))
}

func awsOwnerRestored(t *testing.T, path, kind, raw string, mode os.FileMode) {
	t.Helper()
	data, err := os.ReadFile(path)
	if kind == "absent" {
		if !os.IsNotExist(err) {
			t.Fatalf("expected absence, raw=%q err=%v", data, err)
		}
		return
	}
	awsOwnerCheck(t, err)
	info, err := os.Stat(path)
	awsOwnerCheck(t, err)
	if string(data) != raw || info.Mode().Perm() != mode {
		t.Fatalf("restoration raw=%q mode=%v", data, info.Mode().Perm())
	}
}

func TestMainAWSOwnerCapabilitiesAndJournal(t *testing.T) {
	for _, kind := range []string{"absent", "empty", "raw"} {
		t.Run(kind, func(t *testing.T) {
			path := awsOwnerIsolated(t)
			raw := ""
			var mode os.FileMode
			if kind == "raw" {
				raw = "B"
			}
			if kind != "absent" {
				awsOwnerSeed(t, path, raw)
				info, err := os.Stat(path)
				awsOwnerCheck(t, err)
				mode = info.Mode().Perm()
			}
			owner := (&AWSSwitcher{})
			value, err := owner.CaptureState(context.Background(), &environment.AWSConfig{Profile: "C"})
			awsOwnerCheck(t, err)
			state, ok := value.(*environment.AWSRecoveryState)
			if !ok {
				t.Fatalf("state type %T", value)
			}
			awsOwnerCheck(t, owner.ValidateRecoveryState(state))
			if state.ProfileBefore.Present != (kind != "absent") || state.ProfileBefore.Raw != raw || state.ProfileBefore.Mode != uint32(mode) {
				t.Fatalf("capture metadata %+v", state.ProfileBefore)
			}
			journal, err := environment.DefaultJournalPath()
			awsOwnerCheck(t, err)
			awsOwnerCheck(t, environment.WriteJournal(journal, false, "consumer", map[string]any{"aws": state}))
			rawJournal, readErr := os.ReadFile(journal)
			awsOwnerCheck(t, readErr)
			if !strings.Contains(string(rawJournal), `"version": "aws-target-v2"`) {
				t.Fatalf("not string-v2 journal: %s", rawJournal)
			}
			loaded, err := environment.LoadJournal(journal)
			awsOwnerCheck(t, err)
			if loaded == nil || loaded.Version != 2 {
				t.Fatalf("journal=%+v", loaded)
			}
			var decoded environment.AWSRecoveryState
			awsOwnerCheck(t, json.Unmarshal(loaded.Services["aws"].Data, &decoded))
			awsOwnerCheck(t, owner.ValidateRecoveryState(&decoded))
			awsOwnerCheck(t, owner.ApplyState(context.Background(), state))
			active, err := os.ReadFile(path)
			awsOwnerCheck(t, err)
			if string(active) != "C\n" {
				t.Fatalf("active %q", active)
			}
			awsOwnerCheck(t, owner.Rollback(context.Background(), &decoded))
			awsOwnerRestored(t, path, kind, raw, mode)
		})
	}
}

func TestMainAWSOwnerB1AccountAndCancellationBeforeWrite(t *testing.T) {
	for _, name := range []string{"profile-conflict", "region-conflict", "default-region-conflict", "account-unsupported", "pre-cancel"} {
		t.Run(name, func(t *testing.T) {
			path := awsOwnerIsolated(t)
			awsOwnerSeed(t, path, "B")
			request := &environment.AWSConfig{Profile: "C"}
			ctx := context.Background()
			switch name {
			case "profile-conflict":
				t.Setenv("AWS_PROFILE", "A")
			case "region-conflict":
				t.Setenv("AWS_REGION", "other")
				request.Region = "new"
			case "default-region-conflict":
				t.Setenv("AWS_DEFAULT_REGION", "other")
				request.Region = "new"
			case "account-unsupported":
				request.AccountID = "unsupported"
			case "pre-cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			state, err := (&AWSSwitcher{}).CaptureState(ctx, request)
			if err == nil || state != nil {
				t.Fatalf("expected no plan, got %T %v", state, err)
			}
			if name == "pre-cancel" && !errors.Is(err, context.Canceled) {
				t.Fatalf("cancel identity lost: %v", err)
			}
			manager := environment.NewEnvironmentSwitcher()
			manager.Register(&AWSSwitcher{})
			_, managerErr := manager.SwitchEnvironment(ctx, &environment.Environment{Name: "blocked", Services: map[string]environment.ServiceConfig{"aws": {AWS: request}}}, environment.SwitchOptions{RollbackOnError: true})
			if managerErr == nil {
				t.Fatal("manager accepted unsupported or canceled request")
			}
			if name == "pre-cancel" {
				if !errors.Is(managerErr, context.Canceled) {
					t.Fatalf("manager lost cancellation: %v", managerErr)
				}
			} else {
				want := "conflicts"
				if name == "account-unsupported" {
					want = "unsupported"
				}
				if !strings.Contains(managerErr.Error(), want) {
					t.Fatalf("wrong manager failure: %v", managerErr)
				}
			}
			journal, journalErr := environment.DefaultJournalPath()
			awsOwnerCheck(t, journalErr)
			_, journalErr = os.Stat(journal)
			if !os.IsNotExist(journalErr) {
				t.Fatalf("journal created before request rejection: %v", journalErr)
			}
			info, err := os.Stat(path)
			awsOwnerCheck(t, err)
			awsOwnerRestored(t, path, "raw", "B", info.Mode().Perm())
			_, err = os.Stat(os.Getenv("AWS_CONFIG_FILE"))
			if !os.IsNotExist(err) {
				t.Fatalf("unexpected config mutation: %v", err)
			}
		})
	}
}

type awsOwnerLaterFailure struct {
	path       string
	sawApplied bool
	failure    error
}

func (f *awsOwnerLaterFailure) Name() string { return "docker" }
func (f *awsOwnerLaterFailure) GetCurrentState(context.Context) (any, error) {
	return &environment.DockerConfig{Context: "old"}, nil
}

func (f *awsOwnerLaterFailure) Switch(context.Context, any) error {
	data, err := os.ReadFile(f.path)
	f.sawApplied = err == nil && string(data) == "C\n"
	return f.failure
}
func (f *awsOwnerLaterFailure) Rollback(context.Context, any) error { return nil }
func TestMainAWSOwnerManagerLaterFailureAndReplay(t *testing.T) {
	path := awsOwnerIsolated(t)
	awsOwnerSeed(t, path, "B")
	info, err := os.Stat(path)
	awsOwnerCheck(t, err)
	mode := info.Mode().Perm()
	failure := errors.New("controlled later service failure")
	later := &awsOwnerLaterFailure{path: path, failure: failure}
	manager := environment.NewEnvironmentSwitcher()
	manager.Register((&AWSSwitcher{}))
	manager.Register(later)
	result, err := manager.SwitchEnvironment(context.Background(), &environment.Environment{Name: "consumer", Services: map[string]environment.ServiceConfig{"aws": {AWS: &environment.AWSConfig{Profile: "C"}}, "docker": {Docker: &environment.DockerConfig{Context: "new"}}}, Dependencies: []string{"aws -> docker"}}, environment.SwitchOptions{RollbackOnError: true})
	if !errors.Is(err, failure) || result == nil || result.Success || !result.RollbackPerformed || !later.sawApplied {
		t.Fatalf("result=%+v err=%v applied=%v", result, err, later.sawApplied)
	}
	awsOwnerRestored(t, path, "raw", "B", mode)
	journal, err := environment.DefaultJournalPath()
	awsOwnerCheck(t, err)
	loaded, err := environment.LoadJournal(journal)
	awsOwnerCheck(t, err)
	if loaded == nil || loaded.Version != 2 {
		t.Fatalf("retained journal=%+v", loaded)
	}
	awsOwnerCheck(t, manager.RecoverFromJournal(context.Background(), journal))
	awsOwnerRestored(t, path, "raw", "B", mode)
	_, err = os.Stat(journal)
	if !os.IsNotExist(err) {
		t.Fatalf("journal not cleared: %v", err)
	}
}

// runAWSOwnerFixture 는 고정된 소유 fixture만 실행하고 실제 AWS를 호출하지 않는다.
func runAWSOwnerFixture() int {
	args := os.Args[1:]
	trace, err := os.OpenFile(os.Getenv("GZ_MAIN_AWS_TRACE"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return 90
	}
	err = json.NewEncoder(trace).Encode(args)
	if errors.Join(err, trace.Close()) != nil {
		return 91
	}
	if os.Getenv("GZ_MAIN_AWS_MODE") == "query-fail" {
		return 23
	}
	if strings.Join(args, "|") == "configure|set|region|new|--profile|C" {
		if err := os.WriteFile(os.Getenv("AWS_CONFIG_FILE"), []byte("[profile C]\nregion = new\noutput = json\n"), 0o600); err != nil {
			return 92
		}
		return 0
	}
	if strings.Join(args, "|") == "configure|get|region|--profile|C" {
		_, err := os.Stdout.WriteString("  new\n")
		if err != nil {
			return 93
		}
		return 0
	}
	return 94
}

func TestMainAWSOwnerNativeQueryAndRegion(t *testing.T) {
	path := awsOwnerIsolated(t)
	home := filepath.Dir(filepath.Dir(path))
	bin := t.TempDir()
	exe, err := os.Executable()
	awsOwnerCheck(t, err)
	name := "aws"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	input, err := os.Open(exe)
	awsOwnerCheck(t, err)
	output, err := os.OpenFile(filepath.Join(bin, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	awsOwnerCheck(t, err)
	_, copyErr := io.Copy(output, input)
	awsOwnerCheck(t, errors.Join(copyErr, output.Close(), input.Close()))
	t.Setenv("PATH", bin)
	t.Setenv("PATHEXT", ".EXE")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	t.Setenv("GZ_MAIN_AWS_FIXTURE", "native-v1")
	trace := filepath.Join(home, "native-trace")
	t.Setenv("GZ_MAIN_AWS_TRACE", trace)
	config := os.Getenv("AWS_CONFIG_FILE")
	awsOwnerCheck(t, os.WriteFile(config, []byte("[profile C]\nregion = old\noutput = json\n"), 0o600))
	owner := &AWSSwitcher{}
	plan, err := owner.CaptureState(context.Background(), &environment.AWSConfig{Profile: "C", Region: "new"})
	awsOwnerCheck(t, err)
	state, ok := plan.(*environment.AWSRecoveryState)
	if !ok {
		t.Fatalf("unexpected plan type %T", plan)
	}
	if state.TargetProfile != "C" || state.RegionBefore.Value != "old" {
		t.Fatalf("wrong target plan %+v", state)
	}
	awsOwnerCheck(t, owner.ApplyState(context.Background(), plan))
	snapshot, err := owner.GetCurrentState(context.Background())
	awsOwnerCheck(t, err)
	got, ok := snapshot.(*environment.AWSConfig)
	if !ok {
		t.Fatalf("unexpected snapshot type %T", snapshot)
	}
	if got.Profile != "C" || got.Region != "new" {
		t.Fatalf("query result %+v", got)
	}
	awsOwnerCheck(t, owner.Rollback(context.Background(), plan))
	data, err := os.ReadFile(config)
	awsOwnerCheck(t, err)
	if !strings.Contains(string(data), "region = old") {
		t.Fatalf("target not restored: %s", data)
	}
	calls, err := os.ReadFile(trace)
	awsOwnerCheck(t, err)
	want := `["configure","set","region","new","--profile","C"]` + "\n" + `["configure","get","region","--profile","C"]` + "\n"
	if string(calls) != want {
		t.Fatalf("wrong native argv: %s", calls)
	}
	awsOwnerSeed(t, path, "C")
	t.Setenv("GZ_MAIN_AWS_MODE", "query-fail")
	snapshot, err = owner.GetCurrentState(context.Background())
	var exit *exec.ExitError
	if snapshot != nil || !errors.As(err, &exit) || exit.ExitCode() != 23 {
		t.Fatalf("query failure hidden: %v %v", snapshot, err)
	}
}

type awsOwnerRollbackObserver struct{ calls int }

func (f *awsOwnerRollbackObserver) Name() string { return "docker" }
func (f *awsOwnerRollbackObserver) GetCurrentState(context.Context) (any, error) {
	return &environment.DockerConfig{}, nil
}
func (f *awsOwnerRollbackObserver) Switch(context.Context, any) error   { return nil }
func (f *awsOwnerRollbackObserver) Rollback(context.Context, any) error { f.calls++; return nil }
func TestMainAWSOwnerMixedJournalPreflight(t *testing.T) {
	for _, kind := range []string{"legacy", "malformed"} {
		t.Run(kind, func(t *testing.T) {
			path := awsOwnerIsolated(t)
			awsOwnerSeed(t, path, "B")
			journal, err := environment.DefaultJournalPath()
			awsOwnerCheck(t, err)
			version := any("aws-target-v2")
			entryType := "AWSRecoveryState"
			if kind == "legacy" {
				version = 1
				entryType = "AWSConfig"
			}
			raw, err := json.Marshal(map[string]any{"version": version, "environment": "fixture", "services": map[string]any{"aws": map[string]any{"type": entryType, "data": map[string]any{}}, "docker": map[string]any{"type": "DockerConfig", "data": map[string]any{"Context": "old"}}}})
			awsOwnerCheck(t, err)
			awsOwnerCheck(t, os.WriteFile(journal, raw, 0o600))
			observer := &awsOwnerRollbackObserver{}
			manager := environment.NewEnvironmentSwitcher()
			manager.Register(&AWSSwitcher{})
			manager.Register(observer)
			if err := manager.RecoverFromJournal(context.Background(), journal); err == nil {
				t.Fatal("unsupported journal accepted")
			}
			if observer.calls != 0 {
				t.Fatal("partial rollback occurred before journal rejection")
			}
			data, err := os.ReadFile(path)
			awsOwnerCheck(t, err)
			if string(data) != "B" {
				t.Fatalf("profile changed: %q", data)
			}
			after, err := os.ReadFile(journal)
			awsOwnerCheck(t, err)
			if string(after) != string(raw) {
				t.Fatal("failed journal changed")
			}
		})
	}
}
