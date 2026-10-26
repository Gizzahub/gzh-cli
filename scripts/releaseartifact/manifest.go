// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	maxFile      = 256 << 20
	maxTotal     = 1 << 30
	maxJSON      = 8 << 20
	manifestName = "release-manifest.json"
	checksumName = "checksums.txt"
	bundleName   = "checksums.txt.sigstore.json"
)

var (
	stableTag = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	commitSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)
	fileSHA   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	assetName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	githubTag = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
)

type options struct {
	Dir, Tag, SHA, Repo, Cosign string
	Snapshot                    bool
}

type platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	ARM  string `json:"arm,omitempty"`
}

func (p platform) key() string { return p.OS + "/" + p.Arch + "/" + p.ARM }

var platforms = []platform{
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "linux", Arch: "arm", ARM: "6"},
	{OS: "linux", Arch: "arm", ARM: "7"},
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "windows", Arch: "amd64"},
}

type asset struct {
	Name     string    `json:"name"`
	Kind     string    `json:"kind"`
	SHA256   string    `json:"sha256"`
	Size     int64     `json:"size"`
	Platform *platform `json:"platform,omitempty"`
	Archive  string    `json:"archive,omitempty"`
}

type releaseManifest struct {
	Schema   int     `json:"schema"`
	Tag      string  `json:"tag"`
	Version  string  `json:"version"`
	SHA      string  `json:"sha"`
	Snapshot bool    `json:"snapshot"`
	Assets   []asset `json:"assets"`
}

type metadata struct {
	Project string `json:"project_name"`
	Tag     string `json:"tag"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// GoReleaser v2.19의 공개 artifacts.json 필드만 읽는다.
type artifact struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	OS   string `json:"goos"`
	Arch string `json:"goarch"`
	ARM  string `json:"goarm"`
}

type commandRunner func(context.Context, string, []string, []string, string) ([]byte, error)

type limitedBuffer struct{ buffer bytes.Buffer }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.buffer.Len()+len(p) > 64<<10 {
		return 0, fmt.Errorf("command output exceeds limit")
	}
	return b.buffer.Write(p)
}

func execute(ctx context.Context, executable string, args, env []string, cwd string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir, cmd.Env, cmd.WaitDelay = cwd, env, time.Second
	var output limitedBuffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}
	return output.buffer.Bytes(), nil
}

func cleanEnvironment() []string {
	var env []string
	for _, v := range os.Environ() {
		k, _, _ := strings.Cut(v, "=")
		if strings.HasPrefix(k, "COSIGN_") || strings.HasPrefix(k, "SIGSTORE_") || strings.HasPrefix(k, "GIT_") || k == "GITHUB_TOKEN" || k == "GH_TOKEN" || k == "HOMEBREW_TAP_TOKEN" {
			continue
		}
		env = append(env, v)
	}
	env = append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
	return env
}

func digest(b []byte) string { d := sha256.Sum256(b); return hex.EncodeToString(d[:]) }

func validIdentity(tag, sha string) error {
	if !stableTag.MatchString(tag) || !commitSHA.MatchString(sha) {
		return fmt.Errorf("expected stable tag and full lowercase source SHA")
	}
	return nil
}

func validName(name string) bool {
	return assetName.MatchString(name) && name != "." && name != ".." && path.Base(name) == name
}

// 경로의 모든 구성요소에서 symlink를 거부한다. 파일은 크기를 확인한 뒤 제한해 읽는다.
func regularBytes(filename string, limit int64) ([]byte, error) {
	abs, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	volume := filepath.VolumeName(abs)
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(abs, volume), string(filepath.Separator)), string(filepath.Separator))
	current := volume + string(filepath.Separator)
	for _, part := range parts {
		current = filepath.Join(current, part)
		st, e := os.Lstat(current)
		if e != nil {
			return nil, e
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symlink path rejected")
		}
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > limit {
		return nil, fmt.Errorf("not a bounded regular file")
	}
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("file exceeds limit")
	}
	return b, nil
}

func decodeStrict(b []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func sourceDocuments(ctx context.Context, o options, run commandRunner) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for _, name := range []string{"LICENSE", "README.md", "CHANGELOG.md"} {
		b, err := run(ctx, "git", []string{"show", o.SHA + ":" + name}, cleanEnvironment(), o.Repo)
		if err != nil {
			return nil, fmt.Errorf("source document %s: %w", name, err)
		}
		result[name] = b
	}
	return result, nil
}

func makeChecksums(files map[string][]byte) []byte {
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		fmt.Fprintf(&b, "%s  %s\n", digest(files[n]), n)
	}
	return []byte(b.String())
}

func readStage(dir string) (map[string][]byte, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte)
	var total int64
	for _, entry := range entries {
		if !validName(entry.Name()) {
			return nil, fmt.Errorf("invalid stage filename")
		}
		b, err := regularBytes(filepath.Join(dir, entry.Name()), maxFile)
		if err != nil {
			return nil, err
		}
		total += int64(len(b))
		if total > maxTotal {
			return nil, fmt.Errorf("stage exceeds total size limit")
		}
		files[entry.Name()] = b
	}
	return files, nil
}
