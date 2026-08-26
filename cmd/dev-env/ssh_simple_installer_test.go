// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestBuildSimpleSSHArgsHostKeyPolicy(t *testing.T) {
	tests := []struct {
		name      string
		acceptNew bool
		wantMode  string
	}{
		{name: "strict by default", wantMode: "StrictHostKeyChecking=yes"},
		{name: "explicit accept new", acceptNew: true, wantMode: "StrictHostKeyChecking=accept-new"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := buildSimpleSSHArgs(
				"server.example",
				"admin",
				"echo ready",
				"/isolated/known hosts",
				tt.acceptNew,
			)
			require.NoError(t, err)

			require.Contains(t, args, tt.wantMode)
			require.Contains(t, args, `UserKnownHostsFile="/isolated/known hosts"`)
			require.Contains(t, args, "GlobalKnownHostsFile="+os.DevNull)
			require.Contains(t, args, "KnownHostsCommand=none")
			require.Contains(t, args, "VerifyHostKeyDNS=no")
			require.Contains(t, args, "UpdateHostKeys=no")
			require.Contains(t, args, "CheckHostIP=no")
			require.Contains(t, args, "HashKnownHosts=no")
			require.Contains(t, args, `HostKeyAlias="server.example"`)
			require.Contains(t, args, "CanonicalizeHostname=no")
			require.NotContains(t, args, "StrictHostKeyChecking=no")
			require.Equal(t, []string{"admin@server.example", "echo ready"}, args[len(args)-2:])
		})
	}
}

func TestBuildSimpleSSHArgsEscapesKnownHostsTokens(t *testing.T) {
	args, err := buildSimpleSSHArgs("server.example", "admin", "true", `/tmp/known hosts%h"quoted`, false)
	require.NoError(t, err)
	require.Contains(t, args, `UserKnownHostsFile="/tmp/known hosts%%h\"quoted"`)
}

func TestBuildSimpleSSHArgsEscapesHostIdentityTokens(t *testing.T) {
	args, err := buildSimpleSSHArgs(`server%h "alias"`, "admin", "true", "/tmp/known_hosts", false)
	require.NoError(t, err)
	require.Contains(t, args, `HostKeyAlias="server%%h \"alias\""`)
}

func TestSerializeOpenSSHPathRejectsExpansionAndControls(t *testing.T) {
	_, err := serializeOpenSSHPath("/tmp/${HOME}/known_hosts")
	require.ErrorContains(t, err, "environment expansion")
	_, err = serializeOpenSSHPath("/tmp/known\nhosts")
	require.ErrorContains(t, err, "control character")
}

func TestSerializeOpenSSHPathMatchesSSHConfigParser(t *testing.T) {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		t.Skipf("system ssh is unavailable: %v", err)
	}

	path := filepath.Join(t.TempDir(), `known hosts%h"quoted\part`)
	serialized, err := serializeOpenSSHPath(path)
	require.NoError(t, err)
	output, err := exec.CommandContext(t.Context(), sshPath, "-G", "-o", "UserKnownHostsFile="+serialized, "example.invalid").CombinedOutput()
	require.NoError(t, err, string(output))

	want := "userknownhostsfile " + filepath.ToSlash(path)
	lines := strings.Split(string(output), "\n")
	require.Contains(t, lines, want)
}

func TestSimpleInstallerPreflightStopsExecution(t *testing.T) {
	publicKeyPath := filepath.Join(t.TempDir(), "id_ed25519.pub")
	require.NoError(t, os.WriteFile(publicKeyPath, []byte("ssh-ed25519 test-key\n"), 0o600))
	knownHostsPath := filepath.Join(t.TempDir(), "known_hosts")
	require.NoError(t, os.WriteFile(knownHostsPath, []byte("malformed\n"), 0o600))

	installer := NewSimpleSSHInstaller()
	err := installer.InstallPublicKeySimpleWithOptions(t.Context(), "server.example", "admin", publicKeyPath, HostKeyOptions{
		KnownHostsPath: knownHostsPath,
	})
	require.ErrorContains(t, err, "parse known_hosts")
}

func TestSimpleInstallerReturnsSSHFailureWithoutFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix executable fixture")
	}

	publicKeyPath := filepath.Join(t.TempDir(), "id_ed25519.pub")
	require.NoError(t, os.WriteFile(publicKeyPath, []byte("ssh-ed25519 test-key\n"), 0o600))
	knownHostsPath := filepath.Join(t.TempDir(), "known_hosts")
	binDir := t.TempDir()
	sshFixture := filepath.Join(binDir, "ssh")
	require.NoError(t, os.WriteFile(sshFixture, []byte("#!/bin/sh\necho unsupported accept-new >&2\nexit 255\n"), 0o700))
	t.Setenv("PATH", binDir)
	installer := NewSimpleSSHInstaller()

	err := installer.InstallPublicKeySimpleWithOptions(t.Context(), "server.example", "admin", publicKeyPath, HostKeyOptions{
		KnownHostsPath:   knownHostsPath,
		AcceptNewHostKey: true,
	})
	require.Error(t, err)
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	require.Equal(t, 255, exitError.ExitCode())
}

func TestInstallKeyCommandsExposeHostKeyPolicyFlags(t *testing.T) {
	command := &EnhancedSSHCommand{}
	for _, cmd := range []struct {
		name string
		cmd  interface{ Lookup(string) *pflag.Flag }
	}{
		{name: "advanced", cmd: command.CreateInstallKeyCommand().Flags()},
		{name: "simple", cmd: command.CreateInstallKeySimpleCommand().Flags()},
	} {
		t.Run(cmd.name, func(t *testing.T) {
			require.NotNil(t, cmd.cmd.Lookup("known-hosts"))
			require.NotNil(t, cmd.cmd.Lookup("accept-new-host-key"))
		})
	}
}
