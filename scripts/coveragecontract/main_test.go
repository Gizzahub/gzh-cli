// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestCohortMatrixContainsTwentyFourUniqueProfiles(t *testing.T) {
	// 세 OS/아키텍처와 여덟 태그 조합을 모두 독립 프로파일로 요구한다.
	all := expectedCohorts()
	if len(all) != 24 {
		t.Fatalf("cohort count = %d, want 24", len(all))
	}
	seen := map[string]bool{}
	for _, c := range all {
		k := c.Name + "__" + tagKey(c.Tags)
		if seen[k] {
			t.Fatalf("duplicate cohort %s", k)
		}
		seen[k] = true
	}
}

func TestCoverageThresholdRejectsBelowAndAcceptsExact80(t *testing.T) {
	if err := checkCoverage(7999, 10000); err == nil {
		t.Fatal("79.99% unexpectedly passed")
	}
	if err := checkCoverage(80, 100); err != nil {
		t.Fatalf("exactly 80%% rejected: %v", err)
	}
	if err := checkCoverage(0, 0); err == nil {
		t.Fatal("empty profile unexpectedly passed")
	}
}

func TestCompareBlocksRequiresExactInstrumentedCoordinates(t *testing.T) {
	selected := map[string]bool{"pkg/core.go": true}
	known := map[string]bool{"pkg/core.go": true, "scripts/tool.go": true}
	core := map[string]bool{"pkg/core.go": true}
	expected := []block{{File: "pkg/core.go", StartLine: 2, StartCol: 1, EndLine: 2, EndCol: 7, NumStmt: 2}}
	for name, actual := range map[string][]block{
		"missing":             nil,
		"coordinate mismatch": {{File: "pkg/core.go", StartLine: 3, StartCol: 1, EndLine: 3, EndCol: 7, NumStmt: 2}},
		"statement mismatch":  {{File: "pkg/core.go", StartLine: 2, StartCol: 1, EndLine: 2, EndCol: 7, NumStmt: 1}},
		"unknown file":        {{File: "pkg/extra.go", StartLine: 2, StartCol: 1, EndLine: 2, EndCol: 7, NumStmt: 2}},
	} {
		if _, _, err := compareBlocks(selected, known, core, expected, actual); err == nil {
			t.Errorf("%s profile accepted", name)
		}
	}
	actual := []block{{File: "pkg/core.go", StartLine: 2, StartCol: 1, EndLine: 2, EndCol: 7, NumStmt: 2, Count: 1}}
	actual = append(actual, block{File: "scripts/tool.go", StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 2, NumStmt: 1, Count: 0})
	covered, total, err := compareBlocks(selected, known, core, expected, actual)
	if err != nil {
		t.Fatal(err)
	}
	if covered != 2 || total != 2 {
		t.Fatalf("got %d/%d", covered, total)
	}
}

func TestParseProfileRejectsMissingMalformedAndDuplicateBlocks(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"missing mode": "pkg/core.go:1.1,1.2 1 0\n",
		"malformed":    "mode: count\nnot a profile row\n",
		"duplicate":    "mode: count\npkg/core.go:1.1,1.2 1 0\npkg/core.go:1.1,1.2 1 0\n",
		"empty":        "mode: count\n",
	}
	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(dir, name+".out")
			if err := os.WriteFile(p, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := parseProfile(p, "example.com/mod"); err == nil {
				t.Fatal("invalid profile accepted")
			}
		})
	}
}

func TestParseProfileNormalizesGoModulePaths(t *testing.T) {
	p := filepath.Join(t.TempDir(), "coverage.out")
	if err := os.WriteFile(p, []byte("mode: count\nexample.com/mod/pkg/core.go:1.1,1.2 1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	blocks, err := parseProfile(p, "example.com/mod")
	if err != nil {
		t.Fatal(err)
	}
	if blocks[0].File != "pkg/core.go" {
		t.Fatalf("normalized source = %q", blocks[0].File)
	}
}

func TestProfileModesPreserveExistingWholeProfileModes(t *testing.T) {
	for _, mode := range []string{"set", "count", "atomic"} {
		p := filepath.Join(t.TempDir(), mode+".out")
		if err := os.WriteFile(p, []byte("mode: "+mode+"\npkg/core.go:1.1,1.2 1 0\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := profileMode(p)
		if err != nil {
			t.Fatal(err)
		}
		if got != mode {
			t.Fatalf("mode=%q, want %q", got, mode)
		}
		if _, err = parseProfile(p, "example.com/mod"); err != nil {
			t.Fatalf("mode %s rejected: %v", mode, err)
		}
	}
}

func TestProfileParserKeepsNumStmtAndCountInTheirGoCoverColumns(t *testing.T) {
	for _, mode := range []string{"set", "count", "atomic"} {
		t.Run(mode, func(t *testing.T) {
			count := "3"
			if mode == "set" {
				count = "1"
			}
			path := filepath.Join(t.TempDir(), "coverage.out")
			contents := "mode: " + mode + "\npkg/core.go:1.1,1.2 7 " + count + "\n"
			if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			blocks, err := parseProfile(path, "example.com/mod")
			if err != nil {
				t.Fatal(err)
			}
			if got := blocks[0]; got.NumStmt != 7 || got.Count != mustAtoi(t, count) {
				t.Fatalf("parsed block = %+v, want NumStmt=7 Count=%s", got, count)
			}
		})
	}
	path := filepath.Join(t.TempDir(), "set.out")
	if err := os.WriteFile(path, []byte("mode: set\npkg/core.go:1.1,1.2 7 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := parseProfile(path, "example.com/mod"); err == nil {
		t.Fatal("set profile with Count=2 accepted")
	}
}

func mustAtoi(t *testing.T, value string) int {
	t.Helper()
	parsed, err := strconv.Atoi(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestProfileParserAllowsInstrumentedZeroStatementUnits(t *testing.T) {
	p := filepath.Join(t.TempDir(), "zero.out")
	if err := os.WriteFile(p, []byte("mode: set\npkg/core.go:1.1,1.2 0 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	blocks, err := parseProfile(p, "example.com/mod")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].NumStmt != 0 {
		t.Fatalf("zero NumStmt unit was not preserved: %+v", blocks)
	}
	if err = checkCoverage(0, 0); err == nil {
		t.Fatal("zero denominator unexpectedly passed")
	}
}

func TestParseInstrumentedSourceUsesGoCoverPositionAndStatementArrays(t *testing.T) {
	path := filepath.Join(t.TempDir(), "instrumented.go")
	source := "package fixture\nvar GZCoverage = struct { Count [2]uint32; Pos [6]uint32; NumStmt [2]uint16 }{ Pos: [6]uint32{2, 3, 0x00040002, 7, 8, 0x00030001}, NumStmt: [2]uint16{2, 1} }\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	blocks, err := parseInstrumentedSource(path, "pkg/core.go")
	if err != nil {
		t.Fatal(err)
	}
	want := []block{{File: "pkg/core.go", StartLine: 2, StartCol: 2, EndLine: 3, EndCol: 4, NumStmt: 2}, {File: "pkg/core.go", StartLine: 7, StartCol: 1, EndLine: 8, EndCol: 3, NumStmt: 1}}
	if len(blocks) != len(want) {
		t.Fatalf("blocks=%v", blocks)
	}
	for i := range want {
		if blocks[i] != want[i] {
			t.Errorf("block[%d]=%+v want %+v", i, blocks[i], want[i])
		}
	}
}

func TestPinnedGoCoverToolInstrumentsSourceWithoutExecutingIt(t *testing.T) {
	goBin, version, goHash, coverHash, err := coverTool()
	if err != nil {
		t.Fatal(err)
	}
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	if goHash == "" || coverHash == "" || goBin != filepath.Join(runtime.GOROOT(), "bin", "go"+exe) {
		t.Fatalf("cover tool did not resolve from runtime GOROOT: %s", goBin)
	}
	root := t.TempDir()
	source := filepath.Join(root, "fixture.go")
	code := "package fixture\nvar initialized = 1\nfunc Branch(input bool) int { if input { return initialized }; return 0 }\n"
	if err = os.WriteFile(source, []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	blocks, err := runCoverInstrumenter(goBin, version, root, tmpDir, "fixture.go", "count")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("pinned Go cover tool produced no blocks for executable source")
	}
	for _, b := range blocks {
		if b.File != "fixture.go" || b.NumStmt < 1 {
			t.Fatalf("invalid instrumented block: %+v", b)
		}
	}
}

func TestProfileDirectoryRejectsMissingAndUnknownArtifacts(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "profiles"), 0o700); err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(dir, "profiles")
	if err := validateProfileDirectory(dir); err == nil {
		t.Fatal("missing matrix accepted")
	}
	for _, c := range expectedCohorts() {
		key := c.Name + "__" + tagKey(c.Tags)
		for _, suffix := range []string{".out", ".expected", ".expected.provenance.json", ".provenance.json"} {
			if err := os.WriteFile(filepath.Join(dir, key+suffix), nil, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := validateProfileDirectory(dir); err != nil {
		t.Fatalf("complete filename matrix rejected: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "unknown.out"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateProfileDirectory(dir); err == nil {
		t.Fatal("unknown artifact accepted")
	}
}

func TestZeroBlockFilesReportsInstrumenterZeroBlocksExplicitly(t *testing.T) {
	got := zeroBlockFiles([]string{"pkg/init.go", "pkg/live.go"}, map[string][]block{"pkg/live.go": {{File: "pkg/live.go", StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 2, NumStmt: 1}}})
	if !sameStrings(got, []string{"pkg/init.go"}) {
		t.Fatalf("zero-block set=%v", got)
	}
}

func TestManifestRejectsForgedCohortSelection(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test source path unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := makeManifest([]string{"-root", root, "-out", manifestPath}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err = json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if err = validateManifest(root, m); err != nil {
		t.Fatalf("generated manifest rejected: %v", err)
	}
	m.Cohorts[0].Files = append([]string(nil), m.Cohorts[0].Files[1:]...)
	if err = validateManifest(root, m); err == nil {
		t.Fatal("cohort with an omitted selected source was accepted")
	}
}

func TestAutocrlfKeepsCoverageHashInputsByteStable(t *testing.T) {
	_, testSource, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test source path unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(testSource), "..", ".."))
	coreFiles, err := readInventory(filepath.Join(root, filepath.FromSlash(coreInventoryPath)))
	if err != nil {
		t.Fatal(err)
	}
	attributes, err := os.ReadFile(filepath.Join(root, ".gitattributes"))
	if err != nil {
		t.Fatal(err)
	}
	fixture := t.TempDir()
	inputs := map[string][]byte{}
	for _, path := range []string{
		coreFiles[0], coreInventoryPath, exclusionInventoryPath,
		"scripts/coveragecontract/main.go", "scripts/coveragecontract/main_test.go",
		"go.mod", "go.sum", ".goreleaser.yml", ".make/test.mk",
	} {
		data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if readErr != nil {
			t.Fatalf("read hash input %s: %v", path, readErr)
		}
		inputs[path] = data
	}
	if err := os.WriteFile(filepath.Join(fixture, ".gitattributes"), attributes, 0o644); err != nil {
		t.Fatal(err)
	}
	for path, data := range inputs {
		fullPath := filepath.Join(fixture, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.name", "coveragecontract test"},
		{"config", "user.email", "coveragecontract@example.invalid"},
		{"config", "core.autocrlf", "true"},
	} {
		if _, err := gitOutput(fixture, args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	if _, err := gitOutput(fixture, "add", "--", "."); err != nil {
		t.Fatalf("stage fixture: %v", err)
	}
	if _, err := gitOutput(fixture, "commit", "-qm", "fixture"); err != nil {
		t.Fatalf("commit fixture: %v", err)
	}
	for path, want := range inputs {
		fullPath := filepath.Join(fixture, filepath.FromSlash(path))
		if err := os.WriteFile(fullPath, []byte("modified\r\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := gitOutput(fixture, "checkout", "--", path); err != nil {
			t.Fatalf("checkout %s: %v", path, err)
		}
		got, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("core.autocrlf=true changed bound bytes for %s", path)
		}
		blob, err := gitOutput(fixture, "rev-parse", "HEAD:"+path)
		if err != nil {
			t.Fatal(err)
		}
		workingHash, err := gitOutput(fixture, "hash-object", path)
		if err != nil {
			t.Fatal(err)
		}
		if workingHash != blob {
			t.Fatalf("working tree bytes for %s do not hash to the committed blob", path)
		}
		gotSum, wantSum := sha256.Sum256(got), sha256.Sum256(want)
		if gotSum != wantSum {
			t.Fatalf("SHA256 changed after checkout for %s", path)
		}
	}
	if err := validateInventories(fixture); err != nil {
		t.Fatalf("coverage inventory rejected after autocrlf checkout: %v", err)
	}
}

func TestForgedReferenceAndMatchingSidecarCannotOmitInstrumentedBlocks(t *testing.T) {
	dir := t.TempDir()
	expectedPath := filepath.Join(dir, "expected.out")
	data := []byte("mode: set\npkg/core.go:1.1,1.2 1 0\n")
	if err := os.WriteFile(expectedPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	manifestBytes := []byte("manifest bytes")
	m := manifest{SourceCommit: "source", GoVersion: runtime.Version()}
	c := selectedCohort{Name: "darwin-arm64", GOOS: "darwin", GOARCH: "arm64", Files: []string{"pkg/core.go"}}
	manifestSum := sha256.Sum256(manifestBytes)
	expectedSum, err := fileSHA(expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	toolHash := strings.Repeat("a", 64)
	ref := referenceProvenance{ManifestSHA256: hex.EncodeToString(manifestSum[:]), SourceCommit: m.SourceCommit, GoVersion: m.GoVersion, ToolGOOS: "darwin", ToolGOARCH: "arm64", GoBinarySHA256: toolHash, CoverSHA256: toolHash, Cohort: c, ReferenceSHA256: expectedSum, Mode: "set", Instrumenter: "go tool cover " + m.GoVersion, InstrumenterSHA256: toolHash, ZeroBlockReason: "go tool cover generated no statement blocks for the selected source file"}
	encoded, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(expectedPath+".provenance.json", encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	blocks, err := parseProfile(expectedPath, "example.com/mod")
	if err != nil {
		t.Fatal(err)
	}
	if err = validateReferenceProvenance(expectedPath, manifestBytes, m, c, "set", blocks, blocks, nil); err != nil {
		t.Fatalf("canonical reference profile rejected: %v", err)
	}
	canonical := append([]block(nil), blocks...)
	canonical = append(canonical, block{File: "pkg/core.go", StartLine: 2, StartCol: 1, EndLine: 3, EndCol: 2, NumStmt: 4})
	if err = validateReferenceProvenance(expectedPath, manifestBytes, m, c, "set", blocks, canonical, nil); err == nil {
		t.Fatal("forged expected map with matching sidecar omitted a standard-instrumented block")
	}
}

func TestReleaseBindingRejectsDirtyPreparationManifests(t *testing.T) {
	if err := requireCleanReleaseSnapshot(t.TempDir(), manifest{Dirty: true}); err == nil {
		t.Fatal("dirty source preparation was accepted as release qualification")
	}
}

func TestBuildContextSelectsExplicitTagsAndPlatforms(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"base.go":         "package sample\n",
		"external.go":     "//go:build devenv_external\n\npackage sample\n",
		"base_windows.go": "package sample\n",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx := releaseBuildContext("windows", "amd64", []string{"devenv_external"})
	for name, want := range map[string]bool{"base.go": true, "external.go": true, "base_windows.go": true} {
		got, err := ctx.MatchFile(dir, name)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("MatchFile(%s)=%v want %v", name, got, want)
		}
	}
	ctx = releaseBuildContext("windows", "amd64", nil)
	got, err := ctx.MatchFile(dir, "external.go")
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Fatal("tagged source selected without its tag")
	}
}

func TestReleaseBuildContextProjectsArchitectureTags(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		ctx := releaseBuildContext("linux", arch, nil)
		for _, tag := range ctx.ToolTags {
			if strings.HasPrefix(tag, "amd64.") && arch != "amd64" {
				t.Fatalf("%s context inherited amd64 feature tag %s", arch, tag)
			}
			if strings.HasPrefix(tag, "arm64.") && arch != "arm64" {
				t.Fatalf("%s context inherited arm64 feature tag %s", arch, tag)
			}
		}
		want := arch + ".v1"
		if arch == "arm64" {
			want = "arm64.v8.0"
		}
		found := false
		for _, tag := range ctx.ToolTags {
			found = found || tag == want
		}
		if !found {
			t.Errorf("%s context missing release default feature tag %s: %v", arch, want, ctx.ToolTags)
		}
	}
}

func TestCohortUnionRejectsCoreSourceOutsideAllReleaseSelectors(t *testing.T) {
	cohorts := []selectedCohort{{Files: []string{"cmd/main.go"}}, {Files: []string{"internal/a.go"}}}
	if err := validateCohortUnion([]string{"cmd/main.go", "internal/a.go", "internal/constraint-only.go"}, cohorts); err == nil {
		t.Fatal("core source outside every release selector was accepted")
	}
	if err := validateCohortUnion([]string{"cmd/main.go", "internal/a.go"}, cohorts); err != nil {
		t.Fatalf("complete cohort union rejected: %v", err)
	}
}

func TestNativeProvenanceSupportsDistinctHostBinariesUnderSharedManifest(t *testing.T) {
	manifestBytes := []byte(`{"source_commit":"same","go_version":"go1.27.3","source_selector_sha256":"shared"}`)
	hosts := []cohort{{"darwin-arm64", "darwin", "arm64"}, {"linux-amd64", "linux", "amd64"}, {"windows-amd64", "windows", "amd64"}}
	for i, host := range hosts {
		t.Run(host.Name, func(t *testing.T) {
			dir := t.TempDir()
			profile := filepath.Join(dir, "actual.out")
			expected := filepath.Join(dir, "expected.out")
			if err := os.WriteFile(profile, []byte("mode: set\npkg/core.go:1.1,1.2 1 1\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(expected, []byte("mode: set\npkg/core.go:1.1,1.2 1 0\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			manifestSum := sha256.Sum256(manifestBytes)
			goHash := strings.Repeat(string("0123456789abcdef"[i]), 64)
			coverHash := strings.Repeat(string("abcdef0123456789"[i]), 64)
			c := selectedCohort{Name: host.Name, GOOS: host.GOOS, GOARCH: host.GOARCH}
			ref := referenceProvenance{ManifestSHA256: hex.EncodeToString(manifestSum[:]), SourceCommit: "same", GoVersion: "go1.27.3", ToolGOOS: host.GOOS, ToolGOARCH: host.GOARCH, GoBinarySHA256: goHash, CoverSHA256: coverHash, Cohort: c, ReferenceSHA256: mustSHA256(t, expected), Mode: "set", Instrumenter: "go tool cover go1.27.3", InstrumenterSHA256: coverHash}
			writeJSON(t, expected+".provenance.json", ref)
			if err := validateNativeReferenceTool(expected, host.GOOS, host.GOARCH, goHash, coverHash); err != nil {
				t.Fatalf("native expected-map tool identity rejected: %v", err)
			}
			if err := validateNativeReferenceTool(expected, host.GOOS, host.GOARCH, strings.Repeat("f", 64), coverHash); err == nil {
				t.Fatal("expected-map profile from another Go binary accepted at bind time")
			}
			p := provenance{ManifestSHA256: hex.EncodeToString(manifestSum[:]), SourceCommit: "same", GoVersion: "go1.27.3", ToolGOOS: host.GOOS, ToolGOARCH: host.GOARCH, GoBinarySHA256: goHash, CoverSHA256: coverHash, Cohort: c, ProfileSHA256: mustSHA256(t, profile), ExpectedSHA256: mustSHA256(t, expected), ExpectedProvenanceSHA256: mustSHA256(t, expected+".provenance.json"), Mode: "set", ExpectedMode: "set"}
			prov := filepath.Join(dir, "bind.json")
			writeJSON(t, prov, p)
			m := manifest{SourceCommit: "same", GoVersion: "go1.27.3"}
			if err := verifyProvenance(prov, m, manifestBytes, c, profile, expected); err != nil {
				t.Fatalf("host-specific tool hashes rejected: %v", err)
			}
			p.ToolGOOS = "darwin"
			if host.GOOS == "darwin" {
				p.ToolGOOS = "linux"
			}
			writeJSON(t, prov, p)
			if err := verifyProvenance(prov, m, manifestBytes, c, profile, expected); err == nil {
				t.Fatal("mismatched native host accepted")
			}
			p.ToolGOOS = host.GOOS
			p.SourceCommit = "different-source"
			writeJSON(t, prov, p)
			if err := verifyProvenance(prov, m, manifestBytes, c, profile, expected); err == nil {
				t.Fatal("mismatched source provenance accepted")
			}
			p.SourceCommit = "same"
			p.CoverSHA256 = strings.Repeat("0", 64)
			writeJSON(t, prov, p)
			if err := verifyProvenance(prov, m, manifestBytes, c, profile, expected); err == nil {
				t.Fatal("mismatched native tool provenance accepted")
			}
		})
	}
}

func mustSHA256(t *testing.T, path string) string {
	t.Helper()
	value, err := fileSHA(path)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestInventoryRejectsDuplicateOrTraversalEntries(t *testing.T) {
	for _, contents := range []string{"../escape.go\n", "same.go\nsame.go\n"} {
		p := filepath.Join(t.TempDir(), "inventory.txt")
		if err := os.WriteFile(p, []byte(strings.TrimSpace(contents)), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := readInventory(p); err == nil {
			t.Fatalf("invalid inventory accepted: %q", contents)
		}
	}
}
