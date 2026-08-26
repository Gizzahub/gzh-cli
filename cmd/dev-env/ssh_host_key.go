// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyOptions controls SSH server identity verification.
type HostKeyOptions struct {
	KnownHostsPath   string
	AcceptNewHostKey bool
}

var knownHostsAppendMu sync.Mutex

func resolveKnownHostsPath(path string) (string, error) {
	if path != "" {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("resolve known_hosts %q: %w", path, err)
		}

		return filepath.Clean(absolutePath), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home for known_hosts: %w", err)
	}

	return filepath.Join(homeDir, ".ssh", "known_hosts"), nil
}

func newHostKeyCallback(path string, acceptNew bool, output io.Writer) (ssh.HostKeyCallback, error) {
	if err := prepareKnownHostsFile(path, acceptNew); err != nil {
		return nil, err
	}

	// knownhosts.New accepts only a path. The rooted preflight rejects static
	// symlinks and non-regular files, but a hostile local replacement after this
	// point remains an operating-system-level TOCTOU boundary.
	verify, err := knownhosts.New(path)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts %q: %w", path, err)
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := verify(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if !acceptNew || !errors.As(err, &keyErr) || len(keyErr.Want) != 0 {
			return fmt.Errorf("verify SSH host key for %s: %w", knownhosts.Normalize(hostname), err)
		}

		return appendNewHostKey(path, hostname, remote, key, output)
	}, nil
}

func prepareKnownHostsFile(path string, acceptNew bool) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if !acceptNew {
			return fmt.Errorf("known_hosts %q does not exist; verify the host fingerprint and retry with --accept-new-host-key", path)
		}

		parent := filepath.Dir(path)
		if err := os.MkdirAll(parent, 0o700); err != nil {
			return fmt.Errorf("create known_hosts directory %q: %w", parent, err)
		}
		file, createErr := openFileInParent(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if createErr != nil {
			return fmt.Errorf("create known_hosts %q: %w", path, createErr)
		}
		if closeErr := file.Close(); closeErr != nil {
			return fmt.Errorf("close known_hosts %q: %w", path, closeErr)
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return fmt.Errorf("inspect known_hosts %q: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("known_hosts %q is a symlink; replace it with a regular file", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("known_hosts %q is not a regular file", path)
	}

	file, err := openFileInParent(path, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("read known_hosts %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close known_hosts %q after validation: %w", path, err)
	}
	if _, err := knownhosts.New(path); err != nil {
		return fmt.Errorf("parse known_hosts %q; repair or replace the file: %w", path, err)
	}

	return nil
}

func appendNewHostKey(path, hostname string, remote net.Addr, key ssh.PublicKey, output io.Writer) error {
	knownHostsAppendMu.Lock()
	defer knownHostsAppendMu.Unlock()

	lockPath := path + ".lock"
	lockFile, err := openFileInParent(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("lock known_hosts %q: %w", path, err)
	}
	if err := lockFile.Close(); err != nil {
		_ = os.Remove(lockPath)
		return fmt.Errorf("close known_hosts lock %q: %w", lockPath, err)
	}
	defer func() {
		_ = os.Remove(lockPath)
	}()

	verify, err := knownhosts.New(path)
	if err != nil {
		return fmt.Errorf("reload known_hosts %q: %w", path, err)
	}
	if err := verify(hostname, remote, key); err == nil {
		return nil
	} else {
		var keyErr *knownhosts.KeyError
		if !errors.As(err, &keyErr) || len(keyErr.Want) != 0 {
			return fmt.Errorf("recheck SSH host key for %s: %w", knownhosts.Normalize(hostname), err)
		}
	}

	file, err := openFileInParent(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open known_hosts %q for append: %w", path, err)
	}
	line := knownhosts.Line([]string{hostname}, key) + "\n"
	if _, err := file.WriteString(line); err != nil {
		_ = file.Close()
		return fmt.Errorf("append SSH host key to %q: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync known_hosts %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close known_hosts %q: %w", path, err)
	}

	if output != nil {
		_, _ = fmt.Fprintf(output, "WARNING: accepting a new SSH host key uses trust on first use (TOFU); verify this fingerprint out of band: %s (%s)\n",
			knownhosts.Normalize(hostname), ssh.FingerprintSHA256(key))
	}

	return nil
}

func openFileInParent(path string, flag int, perm os.FileMode) (*os.File, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	file, openErr := root.OpenFile(filepath.Base(path), flag, perm)
	closeErr := root.Close()
	if openErr != nil {
		return nil, openErr
	}
	if closeErr != nil {
		_ = file.Close()
		return nil, closeErr
	}

	return file, nil
}
