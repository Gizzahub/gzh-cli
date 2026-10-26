// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func artifactPath(dir, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || strings.Contains(name, "\\") {
		return "", fmt.Errorf("invalid artifact path")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	// GoReleaser는 작업 디렉터리 기준 dist/... 경로를 기록한다.
	candidate, err := filepath.Abs(filepath.FromSlash(name))
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(candidate, abs+string(filepath.Separator)) {
		candidate = filepath.Join(abs, filepath.FromSlash(name))
	}
	rel, err := filepath.Rel(abs, candidate)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == "." {
		return "", fmt.Errorf("artifact escapes dist")
	}
	return candidate, nil
}

func prepare(ctx context.Context, o options, run commandRunner) error {
	meta, err := readMetadata(o)
	if err != nil {
		return err
	}
	if o.Snapshot && o.Tag == "" {
		o.Tag = meta.Tag
	}
	docs, err := sourceDocuments(ctx, o, run)
	if err != nil {
		return err
	}
	assets, files, err := collectArtifacts(o.Dir)
	if err != nil {
		return err
	}
	body, err := releaseBody(o, docs)
	if err != nil {
		return err
	}
	files["release-notes.md"] = body
	assets = append(assets, asset{Name: "release-notes.md", Kind: "notes", SHA256: digest(body), Size: int64(len(body))})
	sort.Slice(assets, func(i, j int) bool { return assets[i].Name < assets[j].Name })
	m := releaseManifest{Schema: 1, Tag: o.Tag, Version: meta.Version, SHA: o.SHA, Snapshot: o.Snapshot, Assets: assets}
	if err := validatePayload(m, files, docs, meta.Tag); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	files[manifestName] = append(b, '\n')
	files[checksumName] = makeChecksums(files)
	return writeStage(filepath.Join(o.Dir, "publish"), files)
}

func readMetadata(o options) (metadata, error) {
	var meta metadata
	if o.Snapshot {
		if !commitSHA.MatchString(o.SHA) {
			return meta, fmt.Errorf("expected full lowercase source SHA")
		}
	} else if err := validIdentity(o.Tag, o.SHA); err != nil {
		return meta, err
	}
	b, err := regularBytes(filepath.Join(o.Dir, "metadata.json"), maxJSON)
	if err != nil {
		return meta, fmt.Errorf("metadata: %w", err)
	}
	if err := json.Unmarshal(b, &meta); err != nil {
		return meta, err
	}
	if meta.Commit != o.SHA || meta.Project != "gzh-cli" || meta.Version == "" {
		return meta, fmt.Errorf("metadata source/project/version mismatch")
	}
	if !o.Snapshot && (meta.Tag != o.Tag || meta.Version != strings.TrimPrefix(o.Tag, "v")) {
		return meta, fmt.Errorf("metadata stable tag/version mismatch")
	}
	if o.Snapshot && !strings.Contains(strings.ToUpper(meta.Version), "SNAPSHOT") {
		return meta, fmt.Errorf("snapshot version must be explicitly marked")
	}
	return meta, nil
}

func collectArtifacts(dir string) ([]asset, map[string][]byte, error) {
	b, err := regularBytes(filepath.Join(dir, "artifacts.json"), maxJSON)
	if err != nil {
		return nil, nil, err
	}
	var artifacts []artifact
	if err := json.Unmarshal(b, &artifacts); err != nil {
		return nil, nil, err
	}
	if len(artifacts) > 128 {
		return nil, nil, fmt.Errorf("too many artifacts")
	}
	var assets []asset
	files := make(map[string][]byte)
	var stagedBytes int64
	for _, a := range artifacts {
		kind := ""
		switch a.Type {
		case "Archive":
			kind = "archive"
		case "SBOM":
			kind = "sbom"
		case "Homebrew Cask":
			kind = "cask"
		default:
			continue
		}
		if !validName(a.Name) {
			return nil, nil, fmt.Errorf("invalid artifact name")
		}
		filename, err := artifactPath(dir, a.Path)
		if err != nil {
			return nil, nil, err
		}
		b, err := regularBytes(filename, maxFile)
		if err != nil {
			return nil, nil, fmt.Errorf("artifact %s: %w", a.Name, err)
		}
		stagedBytes += int64(len(b))
		if stagedBytes > maxTotal {
			return nil, nil, fmt.Errorf("artifacts exceed total size limit")
		}
		name := a.Name
		if kind == "cask" {
			name = "gz.rb"
		}
		if _, exists := files[name]; exists {
			return nil, nil, fmt.Errorf("duplicate artifact %s", name)
		}
		files[name] = b
		item := asset{Name: name, Kind: kind, SHA256: digest(b), Size: int64(len(b))}
		if kind == "archive" {
			item.Platform = &platform{OS: a.OS, Arch: a.Arch, ARM: a.ARM}
		}
		if kind == "sbom" {
			item.Archive = strings.TrimSuffix(name, ".sbom.json")
		}
		assets = append(assets, item)
	}
	return assets, files, nil
}

func releaseBody(o options, docs map[string][]byte) ([]byte, error) {
	var body []byte
	if o.Snapshot {
		body = append([]byte("# SNAPSHOT — NOT FOR PUBLICATION\n\n"), docs["CHANGELOG.md"]...)
	} else {
		var err error
		body, err = regularBytes(filepath.Join(o.Dir, "CHANGELOG.md"), maxJSON)
		if err != nil {
			return nil, fmt.Errorf("generated release body required: %w", err)
		}
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("empty release body")
	}
	return body, nil
}

func writeStage(stage string, files map[string][]byte) error {
	if err := os.Mkdir(stage, 0o755); err != nil {
		return fmt.Errorf("fresh stage required (partial stages retained): %w", err)
	}
	for name, b := range files {
		f, err := os.OpenFile(filepath.Join(stage, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return err
		}
		_, writeErr := f.Write(b)
		closeErr := f.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

type payloadInventory struct {
	archives           map[string]asset
	platformSet, sboms map[string]bool
	casks, notes       int
}

func validatePayload(m releaseManifest, files, docs map[string][]byte, caskTag string) error {
	seen := make(map[string]bool)
	inventory := payloadInventory{archives: make(map[string]asset), platformSet: make(map[string]bool), sboms: make(map[string]bool)}
	var total int64
	for _, a := range m.Assets {
		b, ok := files[a.Name]
		if !validName(a.Name) || seen[a.Name] || !ok || a.Size <= 0 || a.Size != int64(len(b)) || a.Size > maxFile || !fileSHA.MatchString(a.SHA256) || digest(b) != a.SHA256 {
			return fmt.Errorf("invalid/duplicate asset %s", a.Name)
		}
		seen[a.Name] = true
		total += a.Size
		if total > maxTotal {
			return fmt.Errorf("assets exceed total size limit")
		}
		if err := inventory.add(a, b, docs); err != nil {
			return err
		}
	}
	if err := inventory.complete(); err != nil {
		return err
	}
	return validateCask(files["gz.rb"], m.Version, caskTag, inventory.archives)
}

func (inventory *payloadInventory) add(a asset, b []byte, docs map[string][]byte) error {
	switch a.Kind {
	case "archive":
		if a.Platform == nil || a.Archive != "" || inventory.platformSet[a.Platform.key()] {
			return fmt.Errorf("invalid archive platform")
		}
		inventory.platformSet[a.Platform.key()] = true
		inventory.archives[a.Name] = a
		if err := validateArchive(a, b, docs); err != nil {
			return fmt.Errorf("archive %s: %w", a.Name, err)
		}
	case "sbom":
		if a.Platform != nil || a.Name != a.Archive+".sbom.json" || inventory.sboms[a.Archive] || !json.Valid(b) {
			return fmt.Errorf("invalid SBOM binding")
		}
		inventory.sboms[a.Archive] = true
	case "cask":
		if a.Name != "gz.rb" || a.Platform != nil || a.Archive != "" {
			return fmt.Errorf("invalid cask")
		}
		inventory.casks++
	case "notes":
		if a.Name != "release-notes.md" || a.Platform != nil || a.Archive != "" || len(bytes.TrimSpace(b)) == 0 {
			return fmt.Errorf("invalid notes")
		}
		inventory.notes++
	default:
		return fmt.Errorf("unexpected asset kind")
	}
	return nil
}

func (inventory *payloadInventory) complete() error {
	if len(inventory.archives) != len(platforms) || len(inventory.sboms) != len(inventory.archives) || inventory.casks != 1 || inventory.notes != 1 {
		return fmt.Errorf("incomplete archive/SBOM/cask/notes inventory")
	}
	for _, p := range platforms {
		if !inventory.platformSet[p.key()] {
			return fmt.Errorf("missing expected platform %s", p.key())
		}
	}
	for n := range inventory.archives {
		if !inventory.sboms[n] {
			return fmt.Errorf("missing SBOM for %s", n)
		}
	}
	return nil
}

type archiveValidator struct {
	seen     map[string]bool
	expanded int64
	docs     map[string][]byte
	binary   string
}

func validateArchive(a asset, b []byte, docs map[string][]byte) error {
	v := archiveValidator{seen: make(map[string]bool), docs: docs, binary: "gz"}
	var err error
	if a.Platform.OS == "windows" {
		v.binary = "gz.exe"
		err = validateZIPArchive(a.Name, b, &v)
	} else {
		err = validateTarArchive(a.Name, b, &v)
	}
	if err != nil {
		return err
	}
	for _, name := range []string{v.binary, "LICENSE", "README.md", "CHANGELOG.md", "manpages/gzh-manager.1.gz", "completions/gzh-manager.bash", "completions/gzh-manager.zsh", "completions/gzh-manager.fish"} {
		if !v.seen[name] {
			return fmt.Errorf("missing archive payload %s", name)
		}
	}
	return nil
}

func (v *archiveValidator) checkMember(name string, size int64, reader io.Reader, directory bool) error {
	clean := strings.TrimSuffix(name, "/")
	if clean == "" || path.Clean(clean) != clean || strings.ContainsAny(clean, "\\\r\n") || strings.HasPrefix(clean, "/") || clean == ".." || strings.HasPrefix(clean, "../") || v.seen[clean] {
		return fmt.Errorf("unsafe/duplicate archive path")
	}
	v.seen[clean] = true
	if directory {
		if clean != "manpages" && clean != "completions" {
			return fmt.Errorf("unexpected archive directory")
		}
		return nil
	}
	if size < 0 || size > maxFile {
		return fmt.Errorf("unbounded archive member")
	}
	v.expanded += size
	if v.expanded > maxTotal {
		return fmt.Errorf("expanded archive exceeds limit")
	}
	expected, document := v.docs[clean]
	allowed := document || clean == v.binary || clean == "manpages/gzh-manager.1.gz" || clean == "completions/gzh-manager.bash" || clean == "completions/gzh-manager.zsh" || clean == "completions/gzh-manager.fish"
	if !allowed || size == 0 {
		return fmt.Errorf("unexpected/empty archive payload %s", clean)
	}
	return validateMemberBytes(reader, size, expected, document, clean)
}

func validateMemberBytes(reader io.Reader, size int64, expected []byte, document bool, name string) error {
	if document {
		data, err := io.ReadAll(io.LimitReader(reader, maxFile+1))
		if err != nil {
			return err
		}
		if !bytes.Equal(data, expected) {
			return fmt.Errorf("source document mismatch %s", name)
		}
		return nil
	}
	n, err := io.Copy(io.Discard, io.LimitReader(reader, size+1))
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("archive member size mismatch")
	}
	return nil
}

func validateZIPArchive(name string, b []byte, v *archiveValidator) error {
	if !strings.HasSuffix(name, ".zip") {
		return fmt.Errorf("windows archive must be zip")
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}
	if len(z.File) > 32 {
		return fmt.Errorf("too many archive members")
	}
	for _, f := range z.File {
		if err := validateZIPMember(f, v); err != nil {
			return err
		}
	}
	return nil
}

func validateZIPMember(f *zip.File, v *archiveValidator) error {
	if !f.Mode().IsRegular() && !f.FileInfo().IsDir() {
		return fmt.Errorf("archive link/device rejected")
	}
	// uint64 크기를 signed로 바꾸기 전에 실제 허용 상한과 비교한다.
	if f.UncompressedSize64 > uint64(maxFile) {
		return fmt.Errorf("unbounded archive member")
	}
	r, err := f.Open()
	if err != nil {
		return err
	}
	err = v.checkMember(f.Name, int64(f.UncompressedSize64), r, f.FileInfo().IsDir())
	return errors.Join(err, r.Close())
}

func validateTarArchive(name string, b []byte, v *archiveValidator) error {
	if !strings.HasSuffix(name, ".tar.gz") {
		return fmt.Errorf("unix archive must be tar.gz")
	}
	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer gz.Close()
	t := tar.NewReader(gz)
	for count := 0; ; count++ {
		f, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if count >= 32 || (f.Typeflag != tar.TypeReg && f.Typeflag != tar.TypeRegA && f.Typeflag != tar.TypeDir) {
			return fmt.Errorf("archive link/device/member count rejected")
		}
		if err := v.checkMember(f.Name, f.Size, t, f.Typeflag == tar.TypeDir); err != nil {
			return err
		}
	}
	trailing, err := io.ReadAll(io.LimitReader(gz, (1<<20)+1))
	if err != nil {
		return err
	}
	if len(trailing) > 1<<20 || len(bytes.Trim(trailing, "\x00")) != 0 {
		return fmt.Errorf("unexpected archive trailing payload")
	}
	return nil
}

var caskPairs = regexp.MustCompile(`sha256 "([0-9a-f]{64})"\s+url "([^"\r\n]+)"`)

func validateCask(b []byte, version, tag string, archives map[string]asset) error {
	text := string(b)
	if !strings.Contains(text, `cask "gz" do`) || !strings.Contains(text, `version "`+version+`"`) || !strings.Contains(text, `binary "gz"`) {
		return fmt.Errorf("cask identity mismatch")
	}
	text = strings.ReplaceAll(text, "#{version}", version)
	pairs := caskPairs.FindAllStringSubmatch(text, -1)
	if len(pairs) != 4 {
		return fmt.Errorf("expected four cask URL/hash pairs")
	}
	seen := make(map[string]bool)
	for _, pair := range pairs {
		matched := false
		for name, a := range archives {
			if a.Platform.ARM != "" || a.Platform.OS == "windows" {
				continue
			}
			url := "https://github.com/Gizzahub/gzh-cli/releases/download/" + tag + "/" + name
			if pair[1] == a.SHA256 && pair[2] == url && !seen[name] {
				seen[name] = true
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("cask URL/hash not bound to archives")
		}
	}
	return nil
}
