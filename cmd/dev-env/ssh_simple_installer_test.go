// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"os"
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
			args := buildSimpleSSHArgs(
				"server.example",
				"admin",
				"echo ready",
				"/isolated/known hosts",
				tt.acceptNew,
			)

			require.Contains(t, args, tt.wantMode)
			require.Contains(t, args, "UserKnownHostsFile=/isolated/known hosts")
			require.Contains(t, args, "GlobalKnownHostsFile="+os.DevNull)
			require.Contains(t, args, "KnownHostsCommand=none")
			require.Contains(t, args, "VerifyHostKeyDNS=no")
			require.Contains(t, args, "UpdateHostKeys=no")
			require.Contains(t, args, "CheckHostIP=no")
			require.Contains(t, args, "HashKnownHosts=no")
			require.Contains(t, args, "HostKeyAlias=server.example")
			require.Contains(t, args, "CanonicalizeHostname=no")
			require.NotContains(t, args, "StrictHostKeyChecking=no")
			require.Equal(t, []string{"admin@server.example", "echo ready"}, args[len(args)-2:])
		})
	}
}

func TestBuildSimpleSSHArgsEscapesKnownHostsTokens(t *testing.T) {
	args := buildSimpleSSHArgs("server.example", "admin", "true", "/tmp/known%h", false)
	require.Contains(t, args, "UserKnownHostsFile=/tmp/known%%h")
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
