// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func verify(ctx context.Context, o options, run commandRunner) (map[string][]byte, releaseManifest, error) {
	var m releaseManifest
	if err := validIdentity(o.Tag, o.SHA); err != nil {
		return nil, m, err
	}
	files, err := readStage(o.Dir)
	if err != nil {
		return nil, m, err
	}
	if err := decodeStrict(files[manifestName], &m); err != nil {
		return nil, m, fmt.Errorf("manifest: %w", err)
	}
	if m.Schema != 1 || m.Tag != o.Tag || m.SHA != o.SHA || m.Snapshot || m.Version != strings.TrimPrefix(o.Tag, "v") {
		return nil, m, fmt.Errorf("manifest stable identity mismatch; snapshot cannot publish")
	}
	if len(files) != len(m.Assets)+3 {
		return nil, m, fmt.Errorf("stage exact inventory mismatch")
	}
	listed := make(map[string][]byte, len(m.Assets)+1)
	for _, a := range m.Assets {
		if a.Name == manifestName || a.Name == checksumName || a.Name == bundleName {
			return nil, m, fmt.Errorf("reserved asset name")
		}
		listed[a.Name] = files[a.Name]
	}
	listed[manifestName] = files[manifestName]
	if !bytes.Equal(makeChecksums(listed), files[checksumName]) {
		return nil, m, fmt.Errorf("checksum inventory/hash mismatch")
	}
	if len(files[bundleName]) == 0 || len(files[bundleName]) > maxJSON {
		return nil, m, fmt.Errorf("missing/bounded signature bundle required")
	}
	docs, err := sourceDocuments(ctx, o, run)
	if err != nil {
		return nil, m, err
	}
	if err := validatePayload(m, files, docs, m.Tag); err != nil {
		return nil, m, err
	}
	if err := verifyCosignVersion(ctx, o, run); err != nil {
		return nil, m, err
	}
	if err := verifySignatureSnapshot(ctx, o, files, run); err != nil {
		return nil, m, err
	}
	// 원본 재검사는 일관성 관측이며, 서명 대상은 위의 private bytes다.
	current, err := readStage(o.Dir)
	if err != nil {
		return nil, m, err
	}
	if !sameStage(files, current) {
		return nil, m, fmt.Errorf("stage changed during signature verification")
	}
	return files, m, nil
}

func sameStage(files, current map[string][]byte) bool {
	if len(current) != len(files) {
		return false
	}
	for name, b := range files {
		if !bytes.Equal(b, current[name]) {
			return false
		}
	}
	return true
}

func verifyCosignVersion(ctx context.Context, o options, run commandRunner) error {
	if !filepath.IsAbs(o.Cosign) {
		return fmt.Errorf("absolute cosign v3.1.3 executable required")
	}
	if _, err := regularBytes(o.Cosign, maxFile); err != nil {
		return fmt.Errorf("cosign executable: %w", err)
	}
	version, err := run(ctx, o.Cosign, []string{"version", "--json"}, cleanEnvironment(), o.Repo)
	if err != nil {
		return fmt.Errorf("cosign version: %w", err)
	}
	// 정규식 부분 일치 대신 실제 GitVersion 필드의 고정 버전을 확인한다.
	var gitVersion string
	if err := decodeVersion(version, &gitVersion); err != nil || gitVersion != "v3.1.3" {
		return fmt.Errorf("cosign v3.1.3 required")
	}
	return nil
}

func verifySignatureSnapshot(ctx context.Context, o options, files map[string][]byte, run commandRunner) (err error) {
	// mutable 원본 경로 대신 게시할 메모리와 같은 bytes만 별도 0700 공간에 고정한다.
	private, err := os.MkdirTemp("", "gz-release-verify-")
	if err != nil {
		return fmt.Errorf("create private verification directory: %w", err)
	}
	defer func() {
		var cleanup error
		for _, name := range []string{checksumName, bundleName} {
			// Windows의 read-only 속성도 소유한 두 파일에만 해제한다.
			if e := os.Chmod(filepath.Join(private, name), 0o600); e != nil && !errors.Is(e, os.ErrNotExist) {
				cleanup = errors.Join(cleanup, e)
			}
			if e := os.Remove(filepath.Join(private, name)); e != nil && !errors.Is(e, os.ErrNotExist) {
				cleanup = errors.Join(cleanup, e)
			}
		}
		cleanup = errors.Join(cleanup, os.Remove(private))
		if cleanup != nil {
			err = errors.Join(err, fmt.Errorf("private verification cleanup failed: %w", cleanup))
		}
	}()
	for _, name := range []string{checksumName, bundleName} {
		filename := filepath.Join(private, name)
		f, e := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if e != nil {
			return e
		}
		_, writeErr := f.Write(files[name])
		closeErr := f.Close()
		if e := errors.Join(writeErr, closeErr); e != nil {
			return e
		}
		if e := os.Chmod(filename, 0o400); e != nil {
			return e
		}
	}
	checksums := filepath.Join(private, checksumName)
	args := []string{
		"verify-blob", "--bundle", filepath.Join(private, bundleName),
		"--certificate-identity", "https://github.com/Gizzahub/gzh-cli/.github/workflows/release.yml@refs/tags/" + o.Tag,
		"--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
		"--certificate-github-workflow-sha", o.SHA,
		"--certificate-github-workflow-repository", "Gizzahub/gzh-cli",
		"--certificate-github-workflow-ref", "refs/tags/" + o.Tag,
		"--certificate-github-workflow-trigger", "push", checksums,
	}
	if _, err := run(ctx, o.Cosign, args, cleanEnvironment(), o.Repo); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}
