// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestHostKeyCallbackKnownUnknownAndChanged(t *testing.T) {
	address := "server.example:22"
	remote := &net.TCPAddr{IP: net.ParseIP("192.0.2.10"), Port: 22}
	knownKey := newTestHostKey(t)
	changedKey := newTestHostKey(t)

	t.Run("known key", func(t *testing.T) {
		path := writeKnownHosts(t, knownhosts.Line([]string{address}, knownKey)+"\n")
		callback, err := newHostKeyCallback(path, false, nil)
		require.NoError(t, err)
		require.NoError(t, callback(address, remote, knownKey))
	})

	t.Run("unknown key is rejected by default", func(t *testing.T) {
		path := writeKnownHosts(t, "")
		callback, err := newHostKeyCallback(path, false, nil)
		require.NoError(t, err)
		require.ErrorContains(t, callback(address, remote, knownKey), "key is unknown")
	})

	t.Run("changed key is rejected even with accept new", func(t *testing.T) {
		line := knownhosts.Line([]string{address}, knownKey) + "\n"
		path := writeKnownHosts(t, line)
		callback, err := newHostKeyCallback(path, true, nil)
		require.NoError(t, err)
		require.ErrorContains(t, callback(address, remote, changedKey), "key mismatch")
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, line, string(content))
	})

	t.Run("revoked key is rejected even with accept new", func(t *testing.T) {
		line := "@revoked " + knownhosts.Line([]string{address}, knownKey) + "\n"
		path := writeKnownHosts(t, line)
		callback, err := newHostKeyCallback(path, true, nil)
		require.NoError(t, err)
		require.ErrorContains(t, callback(address, remote, knownKey), "revoked")
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, line, string(content))
	})
}

func TestHostKeyCallbackAcceptsAndRecordsUnknownKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "known_hosts")
	address := "server.example:2222"
	remote := &net.TCPAddr{IP: net.ParseIP("192.0.2.11"), Port: 2222}
	key := newTestHostKey(t)
	var output bytes.Buffer

	callback, err := newHostKeyCallback(path, true, &output)
	require.NoError(t, err)
	require.NoError(t, callback(address, remote, key))
	require.Contains(t, output.String(), knownhosts.Normalize(address))
	require.Contains(t, output.String(), ssh.FingerprintSHA256(key))

	strictCallback, err := knownhosts.New(path)
	require.NoError(t, err)
	require.NoError(t, strictCallback(address, remote, key))

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		parentInfo, err := os.Stat(filepath.Dir(path))
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o700), parentInfo.Mode().Perm())
	}
}

func TestPrepareKnownHostsFileRejectsUnsafePaths(t *testing.T) {
	t.Run("missing strict file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "known_hosts")
		err := prepareKnownHostsFile(path, false)
		require.ErrorContains(t, err, "does not exist")
		_, statErr := os.Stat(path)
		require.ErrorIs(t, statErr, os.ErrNotExist)
	})

	t.Run("malformed file", func(t *testing.T) {
		path := writeKnownHosts(t, "this is not a known_hosts entry\n")
		require.ErrorContains(t, prepareKnownHostsFile(path, false), "parse known_hosts")
		require.ErrorContains(t, prepareKnownHostsFile(path, true), "parse known_hosts")
	})

	t.Run("directory", func(t *testing.T) {
		path := t.TempDir()
		require.ErrorContains(t, prepareKnownHostsFile(path, false), "not a regular file")
		require.ErrorContains(t, prepareKnownHostsFile(path, true), "not a regular file")
	})

	t.Run("symlink", func(t *testing.T) {
		target := writeKnownHosts(t, "")
		path := filepath.Join(t.TempDir(), "known_hosts")
		if err := os.Symlink(target, path); err != nil {
			t.Skipf("symlink is unavailable: %v", err)
		}
		require.ErrorContains(t, prepareKnownHostsFile(path, false), "is a symlink")
		require.ErrorContains(t, prepareKnownHostsFile(path, true), "is a symlink")
	})

	if runtime.GOOS != "windows" {
		t.Run("existing permissions are not widened", func(t *testing.T) {
			path := writeKnownHosts(t, "")
			require.NoError(t, os.Chmod(path, 0o400))
			defer func() { require.NoError(t, os.Chmod(path, 0o600)) }()

			require.NoError(t, prepareKnownHostsFile(path, false))
			callback, err := newHostKeyCallback(path, true, nil)
			require.NoError(t, err)
			err = callback("readonly.example:22", &net.TCPAddr{}, newTestHostKey(t))
			require.Error(t, err)
			info, err := os.Stat(path)
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0o400), info.Mode().Perm())
		})

		t.Run("existing writable permissions survive append", func(t *testing.T) {
			path := writeKnownHosts(t, "")
			callback, err := newHostKeyCallback(path, true, nil)
			require.NoError(t, err)
			require.NoError(t, callback("writable.example:22", &net.TCPAddr{}, newTestHostKey(t)))
			info, err := os.Stat(path)
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		})

		t.Run("unreadable file", func(t *testing.T) {
			path := writeKnownHosts(t, "")
			require.NoError(t, os.Chmod(path, 0o000))
			defer func() { require.NoError(t, os.Chmod(path, 0o600)) }()
			require.ErrorContains(t, prepareKnownHostsFile(path, false), "read known_hosts")
			require.ErrorContains(t, prepareKnownHostsFile(path, true), "read known_hosts")
		})
	}
}

func TestResolveKnownHostsPathMakesRelativePathAbsolute(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)

	path, err := resolveKnownHostsPath(filepath.Join("fixtures", "known_hosts"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workingDirectory, "fixtures", "known_hosts"), path)
}

func TestHostKeyCallbackCanonicalAddresses(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{name: "hostname default port", address: "server.example:22", want: "server.example"},
		{name: "hostname custom port", address: "server.example:2222", want: "[server.example]:2222"},
		{name: "IPv4 default port", address: "192.0.2.12:22", want: "192.0.2.12"},
		{name: "IPv6 default port", address: "[2001:db8::12]:22", want: "2001:db8::12"},
		{name: "IPv6 custom port", address: "[2001:db8::12]:2222", want: "[2001:db8::12]:2222"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeKnownHosts(t, "")
			callback, err := newHostKeyCallback(path, true, nil)
			require.NoError(t, err)
			require.NoError(t, callback(tt.address, &net.TCPAddr{}, newTestHostKey(t)))
			content, err := os.ReadFile(path)
			require.NoError(t, err)
			fields := strings.Fields(string(content))
			require.NotEmpty(t, fields)
			require.Equal(t, tt.want, fields[0])
		})
	}
}

func TestHostKeyCallbackSerializesParallelAcceptNew(t *testing.T) {
	path := writeKnownHosts(t, "")
	callback, err := newHostKeyCallback(path, true, nil)
	require.NoError(t, err)

	type host struct {
		address string
		remote  net.Addr
		key     ssh.PublicKey
	}
	hosts := []host{
		{address: "one.example:22", remote: &net.TCPAddr{IP: net.ParseIP("192.0.2.21"), Port: 22}, key: newTestHostKey(t)},
		{address: "two.example:2222", remote: &net.TCPAddr{IP: net.ParseIP("192.0.2.22"), Port: 2222}, key: newTestHostKey(t)},
	}

	var waitGroup sync.WaitGroup
	errorsByHost := make(chan error, len(hosts))
	for _, item := range hosts {
		waitGroup.Add(1)
		go func(item host) {
			defer waitGroup.Done()
			errorsByHost <- callback(item.address, item.remote, item.key)
		}(item)
	}
	waitGroup.Wait()
	close(errorsByHost)
	for err := range errorsByHost {
		require.NoError(t, err)
	}

	strictCallback, err := knownhosts.New(path)
	require.NoError(t, err)
	for _, item := range hosts {
		require.NoError(t, strictCallback(item.address, item.remote, item.key))
	}
}

func newTestHostKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key, err := ssh.NewPublicKey(publicKey)
	require.NoError(t, err)

	return key
}

func writeKnownHosts(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "known_hosts")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
