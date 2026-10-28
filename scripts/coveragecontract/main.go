// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Command coveragecontract는 관리 대상 핵심 파일을 고르고 Go cover 프로파일을 검증한다.
package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	threshold                = 80
	coreInventorySHA256      = "fba46ce9c7dfa2168508a30865c98d1d595e700244429c4b1f94aab6db918808"
	exclusionInventorySHA256 = "a515728437337c813971e56bf4ab5e453028b361d06c8b35821931e348cabde0"
	coreInventoryPath        = "scripts/coveragecontract/core-files.txt"
	exclusionInventoryPath   = "scripts/coveragecontract/reviewed-exclusions.txt"
	devenvExternalTag        = "devenv_external"
	netenvExternalTag        = "netenv_external"
	pmExternalTag            = "pm_external"
)

var cohorts = []cohort{
	{"darwin-arm64", "darwin", "arm64"},
	{"linux-amd64", "linux", "amd64"},
	{"windows-amd64", "windows", "amd64"},
}

var tagSets = [][]string{
	{},
	{devenvExternalTag},
	{netenvExternalTag},
	{pmExternalTag},
	{devenvExternalTag, netenvExternalTag},
	{devenvExternalTag, pmExternalTag},
	{netenvExternalTag, pmExternalTag},
	{devenvExternalTag, netenvExternalTag, pmExternalTag},
}

type (
	cohort         struct{ Name, GOOS, GOARCH string }
	selectedCohort struct {
		Name       string   `json:"name"`
		GOOS       string   `json:"goos"`
		GOARCH     string   `json:"goarch"`
		CGOEnabled bool     `json:"cgo_enabled"`
		Tags       []string `json:"tags"`
		Files      []string `json:"files"`
	}
)

type manifest struct {
	Schema         int               `json:"schema"`
	SourceCommit   string            `json:"source_commit"`
	SourceTree     string            `json:"source_tree"`
	Dirty          bool              `json:"dirty_worktree"`
	ModulePath     string            `json:"module_path"`
	GoVersion      string            `json:"go_version"`
	SelectorSHA256 string            `json:"source_selector_sha256"`
	Files          map[string]string `json:"source_sha256"`
	BuildInputs    map[string]string `json:"go_build_input_sha256"`
	ContractFiles  map[string]string `json:"contract_sha256"`
	Cohorts        []selectedCohort  `json:"cohorts"`
}
type provenance struct {
	ManifestSHA256           string         `json:"manifest_sha256"`
	SourceCommit             string         `json:"source_commit"`
	GoVersion                string         `json:"go_version"`
	ToolGOOS                 string         `json:"tool_goos"`
	ToolGOARCH               string         `json:"tool_goarch"`
	GoBinarySHA256           string         `json:"go_binary_sha256"`
	CoverSHA256              string         `json:"cover_tool_sha256"`
	Cohort                   selectedCohort `json:"cohort"`
	ProfileSHA256            string         `json:"profile_sha256"`
	ExpectedSHA256           string         `json:"expected_sha256"`
	ExpectedProvenanceSHA256 string         `json:"expected_provenance_sha256"`
	Mode                     string         `json:"mode"`
	ExpectedMode             string         `json:"expected_mode"`
}
type referenceProvenance struct {
	ManifestSHA256     string         `json:"manifest_sha256"`
	SourceCommit       string         `json:"source_commit"`
	GoVersion          string         `json:"go_version"`
	ToolGOOS           string         `json:"tool_goos"`
	ToolGOARCH         string         `json:"tool_goarch"`
	GoBinarySHA256     string         `json:"go_binary_sha256"`
	CoverSHA256        string         `json:"cover_tool_sha256"`
	Cohort             selectedCohort `json:"cohort"`
	ReferenceSHA256    string         `json:"reference_sha256"`
	Mode               string         `json:"mode"`
	Instrumenter       string         `json:"instrumenter"`
	InstrumenterSHA256 string         `json:"instrumenter_sha256"`
	ZeroBlockFiles     []string       `json:"zero_instrumented_block_files"`
	ZeroBlockReason    string         `json:"zero_block_reason"`
}
type block struct {
	File                                                 string
	StartLine, StartCol, EndLine, EndCol, NumStmt, Count int
}

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: coveragecontract manifest|verify")
	}
	var err error
	switch os.Args[1] {
	case "manifest":
		err = makeManifest(os.Args[2:])
	case "bind":
		err = bindProfiles(os.Args[2:])
	case "instrument":
		err = instrumentProfiles(os.Args[2:])
	case "verify":
		err = verifyProfiles(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fatalf("%v", err)
	}
}

func instrumentProfiles(args []string) error {
	fs := flag.NewFlagSet("instrument", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "tmp/coveragecontract/coverage-contract.json", "selector manifest")
	outDir := fs.String("out", "tmp/coveragecontract/profiles", "expected profile directory")
	root := fs.String("root", ".", "repository root")
	mode := fs.String("mode", "set", "Go cover mode: set, count, or atomic")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *mode != "set" && *mode != "count" && *mode != "atomic" {
		return fmt.Errorf("unsupported cover mode %q", *mode)
	}
	data, err := os.ReadFile(*manifestPath)
	if err != nil {
		return err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	abs, err := filepath.Abs(*root)
	if err != nil {
		return err
	}
	if err := validateManifest(abs, m); err != nil {
		return err
	}
	goBin, version, goHash, coverHash, err := coverTool()
	if err != nil {
		return err
	}
	if version != m.GoVersion {
		return fmt.Errorf("cover tool %s does not match manifest Go %s", version, m.GoVersion)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(*outDir, ".instrument-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpDir)
	// 각 소스는 한 번 계측하고 OS와 빌드 태그에 따라 계측 맵을 코호트에 묶는다.
	blocksByFile, err := instrumentFiles(goBin, version, abs, tmpDir, mapKeys(m.Files), *mode)
	if err != nil {
		return err
	}
	manifestSum := sha256.Sum256(data)
	manifestDigest := hex.EncodeToString(manifestSum[:])
	for _, c := range m.Cohorts {
		if c.CGOEnabled {
			return fmt.Errorf("cohort %s unexpectedly enables cgo", c.Name)
		}
		key := c.Name + "__" + tagKey(c.Tags)
		path := filepath.Join(*outDir, key+".expected")
		if err := writeExpectedProfile(path, *mode, c.Files, blocksByFile); err != nil {
			return err
		}
		zeroFiles := zeroBlockFiles(c.Files, blocksByFile)
		if len(zeroFiles) == len(c.Files) {
			return fmt.Errorf("cohort %s has zero instrumented statements; no valid coverage denominator", key)
		}
		digest, err := fileSHA(path)
		if err != nil {
			return err
		}
		p := referenceProvenance{ManifestSHA256: manifestDigest, SourceCommit: m.SourceCommit, GoVersion: m.GoVersion, ToolGOOS: runtime.GOOS, ToolGOARCH: runtime.GOARCH, GoBinarySHA256: goHash, CoverSHA256: coverHash, Cohort: c, ReferenceSHA256: digest, Mode: *mode, Instrumenter: "go tool cover " + version, InstrumenterSHA256: coverHash, ZeroBlockFiles: zeroFiles, ZeroBlockReason: "go tool cover generated no statement blocks for the selected source file"}
		b, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path+".provenance.json", append(b, '\n'), 0o644); err != nil {
			return err
		}
	}
	return verifySourceSnapshot(abs, m)
}

func runCoverInstrumenter(goBin, version, root, tmpDir, source, mode string) ([]block, error) {
	input := filepath.Join(root, filepath.FromSlash(source))
	sum := sha256.Sum256([]byte(source))
	output := filepath.Join(tmpDir, hex.EncodeToString(sum[:])+".go")
	defer os.Remove(output)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, goBin, "tool", "cover", "-mode="+mode, "-var=GZCoverage", "-o", output, input)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("go tool cover %s %s: %w: %s", version, source, err, strings.TrimSpace(string(out)))
	}
	return parseInstrumentedSource(output, source)
}

func instrumentFiles(goBin, version, root, tmpDir string, files []string, mode string) (map[string][]block, error) {
	out := make(map[string][]block, len(files))
	for _, p := range files {
		blocks, err := runCoverInstrumenter(goBin, version, root, tmpDir, p, mode)
		if err != nil {
			return nil, err
		}
		out[p] = blocks
	}
	return out, nil
}

func mapKeys(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for p := range values {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func blocksForFiles(files []string, byFile map[string][]block) []block {
	var out []block
	for _, p := range files {
		out = append(out, byFile[p]...)
	}
	return out
}

func coverTool() (string, string, string, string, error) {
	root := runtime.GOROOT()
	if root == "" {
		return "", "", "", "", errors.New("runtime has no GOROOT; cannot resolve the pinned Go toolchain")
	}
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	goBin := filepath.Join(root, "bin", "go"+exe)
	coverBin := filepath.Join(root, "pkg", "tool", runtime.GOOS+"_"+runtime.GOARCH, "cover"+exe)
	if err := regularAbsoluteFile(goBin); err != nil {
		return "", "", "", "", err
	}
	if err := regularAbsoluteFile(coverBin); err != nil {
		return "", "", "", "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	goVersionCmd := exec.CommandContext(ctx, goBin, "version")
	goVersionOut, err := goVersionCmd.CombinedOutput()
	if err != nil {
		return "", "", "", "", fmt.Errorf("query GOROOT Go version: %w: %s", err, strings.TrimSpace(string(goVersionOut)))
	}
	if !strings.Contains(string(goVersionOut), runtime.Version()) {
		return "", "", "", "", fmt.Errorf("GOROOT Go binary reports %q, runtime is %q", strings.TrimSpace(string(goVersionOut)), runtime.Version())
	}
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, goBin, "tool", "cover", "-V")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", "", "", fmt.Errorf("query GOROOT go tool cover version: %w: %s", err, strings.TrimSpace(string(out)))
	}
	version := strings.TrimSpace(string(out))
	if !strings.HasSuffix(version, runtime.Version()) {
		return "", "", "", "", fmt.Errorf("cover tool version %q does not match runtime %q", version, runtime.Version())
	}
	goHash, err := fileSHA(goBin)
	if err != nil {
		return "", "", "", "", err
	}
	coverHash, err := fileSHA(coverBin)
	if err != nil {
		return "", "", "", "", err
	}
	return goBin, runtime.Version(), goHash, coverHash, nil
}

func regularAbsoluteFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("pinned tool binary %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("pinned tool path is not a regular file: %s", path)
	}
	return nil
}

func parseInstrumentedSource(path, source string) ([]block, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse Go cover output for %s: %w", source, err)
	}
	target, err := coverageMapLiteral(file, source)
	if err != nil {
		return nil, err
	}
	positions, statements, err := coverageArrays(target, source)
	if err != nil {
		return nil, err
	}
	return instrumentedBlocks(source, positions, statements)
}

func coverageMapLiteral(file *ast.File, source string) (*ast.CompositeLit, error) {
	var target *ast.CompositeLit
	var invalid bool
	ast.Inspect(file, func(node ast.Node) bool {
		v, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, name := range v.Names {
			if name.Name != "GZCoverage" || i >= len(v.Values) {
				continue
			}
			target, ok = v.Values[i].(*ast.CompositeLit)
			invalid = !ok
			return false
		}
		return true
	})
	if invalid {
		return nil, fmt.Errorf("go tool cover GZCoverage value for %s is not a composite literal", source)
	}
	if target == nil {
		return nil, fmt.Errorf("go tool cover output for %s has no GZCoverage map", source)
	}
	return target, nil
}

func coverageArrays(target *ast.CompositeLit, source string) ([]uint64, []uint64, error) {
	var positions, statements []uint64
	for _, elt := range target.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		lit, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}
		var err error
		switch key.Name {
		case "Pos":
			positions, err = uintValues(lit)
		case "NumStmt":
			statements, err = uintValues(lit)
		default:
			continue
		}
		if err != nil {
			return nil, nil, fmt.Errorf("read %s coverage map: %w", source, err)
		}
	}
	if len(positions)%3 != 0 || len(positions)/3 != len(statements) {
		return nil, nil, fmt.Errorf("go tool cover output for %s has inconsistent position/NumStmt arrays", source)
	}
	return positions, statements, nil
}

func instrumentedBlocks(source string, positions, statements []uint64) ([]block, error) {
	out := make([]block, 0, len(statements))
	for i, n := range statements {
		packed := positions[i*3+2]
		startLine, err := uintToInt(positions[i*3])
		if err != nil {
			return nil, fmt.Errorf("invalid start line in %s: %w", source, err)
		}
		endLine, err := uintToInt(positions[i*3+1])
		if err != nil {
			return nil, fmt.Errorf("invalid end line in %s: %w", source, err)
		}
		statements, err := uintToInt(n)
		if err != nil {
			return nil, fmt.Errorf("invalid statement count in %s: %w", source, err)
		}
		b := block{File: filepath.ToSlash(source), StartLine: startLine, StartCol: int(packed & 0xffff), EndLine: endLine, EndCol: int((packed >> 16) & 0xffff), NumStmt: statements}
		out = append(out, b)
	}
	return out, nil
}

func uintToInt(n uint64) (int, error) {
	if n > uint64(^uint(0)>>1) {
		return 0, fmt.Errorf("%d exceeds platform int range", n)
	}
	return int(n), nil
}

func uintValues(lit *ast.CompositeLit) ([]uint64, error) {
	out := make([]uint64, 0, len(lit.Elts))
	for _, elt := range lit.Elts {
		v, ok := elt.(*ast.BasicLit)
		if !ok {
			return nil, fmt.Errorf("unexpected value %T", elt)
		}
		n, err := strconv.ParseUint(v.Value, 0, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func writeExpectedProfile(path, mode string, files []string, byFile map[string][]block) error {
	var out strings.Builder
	fmt.Fprintf(&out, "mode: %s\n", mode)
	for _, file := range files {
		for _, b := range byFile[file] {
			fmt.Fprintf(&out, "%s:%d.%d,%d.%d %d 0\n", b.File, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmt)
		}
	}
	return os.WriteFile(path, []byte(out.String()), 0o644)
}

func zeroBlockFiles(files []string, byFile map[string][]block) []string {
	var out []string
	for _, p := range files {
		if len(byFile[p]) == 0 {
			out = append(out, p)
		}
	}
	return out
}

func sameStrings(a, b []string) bool {
	return slices.Equal(a, b)
}

func bindProfiles(args []string) error {
	fs := flag.NewFlagSet("bind", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "tmp/coveragecontract/coverage-contract.json", "selector manifest")
	cohortName := fs.String("cohort", "", "cohort identifier, e.g. linux-amd64")
	tagsArg := fs.String("tags", "", "comma-separated build tags")
	profilePath := fs.String("profile", "", "measured count profile")
	expectedPath := fs.String("expected", "", "same-toolchain instrumented reference profile")
	out := fs.String("out", "", "provenance output")
	root := fs.String("root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *cohortName == "" || *profilePath == "" || *expectedPath == "" || *out == "" {
		return errors.New("bind requires cohort, profile, expected and out")
	}
	data, err := os.ReadFile(*manifestPath)
	if err != nil {
		return err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m.GoVersion != runtime.Version() {
		return fmt.Errorf("profile toolchain %s does not match manifest %s", runtime.Version(), m.GoVersion)
	}
	if *cohortName != runtime.GOOS+"-"+runtime.GOARCH {
		return fmt.Errorf("actual profile must be bound on native cohort %s; runtime is %s-%s", *cohortName, runtime.GOOS, runtime.GOARCH)
	}
	if os.Getenv("CGO_ENABLED") != "0" {
		return errors.New("actual profile must use CGO_ENABLED=0, matching the release environment")
	}
	abs, err := filepath.Abs(*root)
	if err != nil {
		return err
	}
	if err := validateManifest(abs, m); err != nil {
		return err
	}
	if err := requireCleanReleaseSnapshot(abs, m); err != nil {
		return err
	}
	tags := parseTags(*tagsArg)
	var cohortEntry *selectedCohort
	for i := range m.Cohorts {
		c := &m.Cohorts[i]
		if c.Name == *cohortName && tagKey(c.Tags) == tagKey(tags) {
			cohortEntry = c
			break
		}
	}
	if cohortEntry == nil {
		return fmt.Errorf("unknown cohort %s with tags %v", *cohortName, tags)
	}
	if cohortEntry.CGOEnabled {
		return errors.New("release profile contract requires CGO_ENABLED=0")
	}
	actualSum, expectedSum, expectedProvenanceSum, actualMode, goHash, coverHash, err := bindCohortEvidence(abs, data, m, *cohortEntry, *profilePath, *expectedPath)
	if err != nil {
		return err
	}
	manifestSum := sha256.Sum256(data)
	p := provenance{ManifestSHA256: hex.EncodeToString(manifestSum[:]), SourceCommit: m.SourceCommit, GoVersion: runtime.Version(), ToolGOOS: runtime.GOOS, ToolGOARCH: runtime.GOARCH, GoBinarySHA256: goHash, CoverSHA256: coverHash, Cohort: *cohortEntry, ProfileSHA256: actualSum, ExpectedSHA256: expectedSum, ExpectedProvenanceSHA256: expectedProvenanceSum, Mode: actualMode, ExpectedMode: actualMode}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := verifySourceSnapshot(abs, m); err != nil {
		return err
	}
	if err := requireCleanReleaseSnapshot(abs, m); err != nil {
		return err
	}
	return os.WriteFile(*out, append(b, '\n'), 0o644)
}

func bindCohortEvidence(root string, manifestBytes []byte, m manifest, c selectedCohort, profilePath, expectedPath string) (actualSum, expectedSum, expectedProvenanceSum, mode, goHash, coverHash string, err error) {
	expected, err := parseProfile(expectedPath, m.ModulePath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	actual, err := parseProfile(profilePath, m.ModulePath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	mode, err = profileMode(profilePath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	expectedMode, err := profileMode(expectedPath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	if mode != expectedMode {
		err = fmt.Errorf("coverage mode differs between actual (%s) and instrumented reference (%s)", mode, expectedMode)
		return "", "", "", "", "", "", err
	}
	goBin, toolVersion, goHash, coverHash, err := coverTool()
	if err != nil {
		return "", "", "", "", "", "", err
	}
	if toolVersion != m.GoVersion {
		err = errors.New("resolved Go/cover version differs from the manifest")
		return "", "", "", "", "", "", err
	}
	if runtime.GOOS != c.GOOS || runtime.GOARCH != c.GOARCH {
		err = fmt.Errorf("native profile host %s-%s does not match cohort %s", runtime.GOOS, runtime.GOARCH, c.Name)
		return "", "", "", "", "", "", err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(expectedPath), ".bind-source-map-")
	if err != nil {
		return "", "", "", "", "", "", err
	}
	defer os.Remove(tmpDir)
	canonical, zeroFiles, err := instrumentCohort(goBin, toolVersion, root, tmpDir, &c, mode)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	if err := validateReferenceProvenance(expectedPath, manifestBytes, m, c, mode, expected, canonical, zeroFiles); err != nil {
		return "", "", "", "", "", "", err
	}
	if err := validateNativeReferenceTool(expectedPath, runtime.GOOS, runtime.GOARCH, goHash, coverHash); err != nil {
		return "", "", "", "", "", "", err
	}
	if _, _, err = compareCoreProfile(m, c, expected, actual); err != nil {
		return "", "", "", "", "", "", err
	}
	actualSum, err = fileSHA(profilePath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	expectedSum, err = fileSHA(expectedPath)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	expectedProvenanceSum, err = fileSHA(expectedPath + ".provenance.json")
	return actualSum, expectedSum, expectedProvenanceSum, mode, goHash, coverHash, err
}

func compareCoreProfile(m manifest, c selectedCohort, expected, actual []block) (int64, int64, error) {
	selected := make(map[string]bool, len(c.Files))
	for _, p := range c.Files {
		selected[p] = true
	}
	known := make(map[string]bool, len(m.BuildInputs))
	for p := range m.BuildInputs {
		known[p] = true
	}
	core := make(map[string]bool, len(m.Files))
	for p := range m.Files {
		core[p] = true
	}
	covered, total, err := compareBlocks(selected, known, core, expected, actual)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot bind incomplete/mismatched profiles: %w", err)
	}
	return covered, total, nil
}

func validateNativeReferenceTool(expectedPath, goos, goarch, goHash, coverHash string) error {
	data, err := os.ReadFile(expectedPath + ".provenance.json")
	if err != nil {
		return err
	}
	var ref referenceProvenance
	if err := json.Unmarshal(data, &ref); err != nil {
		return err
	}
	if ref.ToolGOOS != goos || ref.ToolGOARCH != goarch || ref.GoBinarySHA256 != goHash || ref.CoverSHA256 != coverHash {
		return errors.New("expected source map was generated by a different native Go/cover binary")
	}
	return nil
}

func validateReferenceProvenance(expectedPath string, manifestBytes []byte, m manifest, c selectedCohort, mode string, blocks, canonical []block, zeroFiles []string) error {
	refPath := expectedPath + ".provenance.json"
	refData, err := os.ReadFile(refPath)
	if err != nil {
		return fmt.Errorf("read Go cover instrumenter provenance: %w", err)
	}
	var ref referenceProvenance
	if err := json.Unmarshal(refData, &ref); err != nil {
		return fmt.Errorf("decode Go cover instrumenter provenance: %w", err)
	}
	manifestDigest := sha256.Sum256(manifestBytes)
	refDigest, err := fileSHA(expectedPath)
	if err != nil {
		return err
	}
	if !referenceMatchesSelector(ref, hex.EncodeToString(manifestDigest[:]), m, c) || !referenceMatchesToolAndArtifact(ref, mode, refDigest) {
		return errors.New("expected profile is not bound to this manifest and pinned Go cover instrumenter")
	}
	if ref.ZeroBlockReason != "go tool cover generated no statement blocks for the selected source file" {
		return errors.New("zero-block explanation is missing or inaccurate")
	}
	if !sameStrings(zeroFiles, ref.ZeroBlockFiles) {
		return errors.New("expected profile zero-block file list does not match source re-instrumentation")
	}
	for _, b := range blocks {
		if b.Count != 0 {
			return errors.New("expected source map must have zero counters")
		}
	}
	selected := map[string]bool{}
	for _, p := range c.Files {
		selected[p] = true
	}
	if _, _, err = compareBlocks(selected, selected, selected, canonical, blocks); err != nil {
		return fmt.Errorf("expected profile differs from current source instrumentation: %w", err)
	}
	return nil
}

func referenceMatchesSelector(ref referenceProvenance, manifestDigest string, m manifest, c selectedCohort) bool {
	return ref.ManifestSHA256 == manifestDigest && ref.SourceCommit == m.SourceCommit && ref.GoVersion == m.GoVersion &&
		ref.Cohort.Name == c.Name && ref.Cohort.GOOS == c.GOOS && ref.Cohort.GOARCH == c.GOARCH &&
		ref.Cohort.CGOEnabled == c.CGOEnabled && tagKey(ref.Cohort.Tags) == tagKey(c.Tags)
}

func referenceMatchesToolAndArtifact(ref referenceProvenance, mode, digest string) bool {
	return ref.ToolGOOS != "" && ref.ToolGOARCH != "" && validSHA256(ref.GoBinarySHA256) && validSHA256(ref.CoverSHA256) &&
		ref.ReferenceSHA256 == digest && ref.Mode == mode && ref.Instrumenter == "go tool cover "+ref.GoVersion &&
		ref.InstrumenterSHA256 == ref.CoverSHA256
}

func instrumentCohort(goBin, version, root, tmpDir string, c *selectedCohort, mode string) ([]block, []string, error) {
	byFile, err := instrumentFiles(goBin, version, root, tmpDir, c.Files, mode)
	if err != nil {
		return nil, nil, err
	}
	zero := zeroBlockFiles(c.Files, byFile)
	if len(zero) == len(c.Files) {
		return nil, nil, fmt.Errorf("cohort %s has zero instrumented statements; no valid coverage denominator", c.Name)
	}
	return blocksForFiles(c.Files, byFile), zero, nil
}

func makeManifest(args []string) error {
	fs := flag.NewFlagSet("manifest", flag.ContinueOnError)
	root := fs.String("root", ".", "repository root")
	out := fs.String("out", "tmp/coveragecontract/coverage-contract.json", "manifest output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	abs, err := filepath.Abs(*root)
	if err != nil {
		return err
	}
	if err := validateInventories(abs); err != nil {
		return err
	}
	paths, err := readInventory(filepath.Join(abs, filepath.FromSlash(coreInventoryPath)))
	if err != nil {
		return err
	}
	excluded, err := readPathFile(filepath.Join(abs, filepath.FromSlash(exclusionInventoryPath)))
	if err != nil {
		return err
	}
	m, err := buildManifest(abs, paths, excluded)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(*out, append(data, '\n'), 0o644)
}

func buildManifest(root string, paths, excluded []string) (manifest, error) {
	m := manifest{Schema: 1, Files: map[string]string{}, BuildInputs: map[string]string{}, ContractFiles: map[string]string{}}
	if err := manifestSourceMetadata(root, &m); err != nil {
		return manifest{}, err
	}
	files, err := hashPaths(root, paths)
	if err != nil {
		return manifest{}, err
	}
	m.Files = files
	contractPaths := []string{"scripts/coveragecontract/main.go", coreInventoryPath, exclusionInventoryPath}
	contractHashes, err := hashPaths(root, contractPaths)
	if err != nil {
		return manifest{}, err
	}
	m.ContractFiles = contractHashes
	inputs, err := goSourcePaths(root)
	if err != nil {
		return manifest{}, err
	}
	m.BuildInputs, err = hashPaths(root, inputs)
	if err != nil {
		return manifest{}, err
	}
	contractHashes, err = hashPaths(root, []string{"go.mod", "go.sum", ".goreleaser.yml", ".make/test.mk"})
	if err != nil {
		return manifest{}, err
	}
	for p, digest := range contractHashes {
		m.ContractFiles[p] = digest
	}
	if err := validateProductionSourceClassification(root, paths, excluded); err != nil {
		return manifest{}, err
	}
	m.Cohorts, err = selectReleaseCohorts(root, paths)
	if err != nil {
		return manifest{}, err
	}
	m.SelectorSHA256, err = selectorDigest(m)
	if err != nil {
		return manifest{}, err
	}
	return m, nil
}

func manifestSourceMetadata(root string, m *manifest) error {
	var err error
	m.SourceCommit, err = gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	m.SourceTree, err = gitOutput(root, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return err
	}
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return err
	}
	m.Dirty = strings.TrimSpace(status) != ""
	m.GoVersion = runtime.Version()
	_, version, _, _, err := coverTool()
	if err != nil {
		return err
	}
	if version != m.GoVersion {
		return fmt.Errorf("resolved toolchain %s does not match runtime %s", version, m.GoVersion)
	}
	m.ModulePath, err = modulePath(filepath.Join(root, "go.mod"))
	return err
}

func hashPaths(root string, paths []string) (map[string]string, error) {
	out := make(map[string]string, len(paths))
	for _, path := range paths {
		if err := validateRegularPath(root, path); err != nil {
			return nil, err
		}
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return nil, fmt.Errorf("hash source %s: %w", path, err)
		}
		sum := sha256.Sum256(contents)
		out[path] = hex.EncodeToString(sum[:])
	}
	return out, nil
}

func goSourcePaths(root string) ([]string, error) {
	tracked, err := gitOutput(root, "ls-files", "--", "*.go")
	if err != nil {
		return nil, err
	}
	untracked, err := gitOutput(root, "ls-files", "--others", "--exclude-standard", "--", "*.go")
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, list := range []string{tracked, untracked} {
		for _, path := range strings.Split(list, "\n") {
			if path != "" {
				set[path] = true
			}
		}
	}
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func validateProductionSourceClassification(root string, paths, excluded []string) error {
	coreSet := make(map[string]bool, len(paths)+len(excluded)+2)
	for _, path := range paths {
		coreSet[path] = true
	}
	for _, path := range excluded {
		if coreSet[path] {
			return fmt.Errorf("source is both included and excluded: %s", path)
		}
		coreSet[path] = true
	}
	// 계약 검증기는 검토된 개발 도구이며 제품 코드가 아니다.
	coreSet["scripts/coveragecontract/main.go"] = true
	coreSet["scripts/coveragecontract/main_test.go"] = true
	tracked, err := gitOutput(root, "ls-files", "--", "*.go")
	if err != nil {
		return err
	}
	untracked, err := gitOutput(root, "ls-files", "--others", "--exclude-standard", "--", "*.go")
	if err != nil {
		return err
	}
	if err := checkProductionSources(strings.Split(strings.TrimSpace(tracked), "\n"), coreSet, false); err != nil {
		return err
	}
	return checkProductionSources(strings.Split(strings.TrimSpace(untracked), "\n"), coreSet, true)
}

func checkProductionSources(paths []string, coreSet map[string]bool, untracked bool) error {
	for _, path := range paths {
		if path == "" || strings.HasSuffix(path, "_test.go") {
			continue
		}
		if !coreSet[path] {
			if untracked {
				return fmt.Errorf("unclassified untracked production Go source %s", path)
			}
			return fmt.Errorf("unclassified production Go source %s; update reviewed core inventory explicitly", path)
		}
	}
	return nil
}

func selectReleaseCohorts(root string, paths []string) ([]selectedCohort, error) {
	out := make([]selectedCohort, 0, len(cohorts)*len(tagSets))
	for _, target := range cohorts {
		for _, tags := range tagSets {
			selected, err := selectCohortFiles(root, paths, target, tags)
			if err != nil {
				return nil, err
			}
			out = append(out, selected)
		}
	}
	return out, nil
}

func selectCohortFiles(root string, paths []string, target cohort, tags []string) (selectedCohort, error) {
	selected := selectedCohort{Name: target.Name, GOOS: target.GOOS, GOARCH: target.GOARCH, CGOEnabled: false, Tags: append([]string(nil), tags...)}
	ctx := releaseBuildContext(target.GOOS, target.GOARCH, tags)
	for _, path := range paths {
		match, err := ctx.MatchFile(root, path)
		if err != nil {
			return selectedCohort{}, fmt.Errorf("select %s for %s/%s %v: %w", path, target.GOOS, target.GOARCH, tags, err)
		}
		if match {
			selected.Files = append(selected.Files, path)
		}
	}
	if len(selected.Files) == 0 {
		return selectedCohort{}, fmt.Errorf("empty core cohort %s %v", target.Name, tags)
	}
	return selected, nil
}

func verifyProfiles(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "tmp/coveragecontract/coverage-contract.json", "source selector manifest")
	profilesDir := fs.String("profiles", "tmp/coveragecontract/profiles", "directory with <cohort>__<tagkey>.out and .expected profiles")
	root := fs.String("root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	data, err := os.ReadFile(*manifestPath)
	if err != nil {
		return err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	if m.GoVersion != runtime.Version() {
		return fmt.Errorf("profile toolchain %s does not match manifest %s", runtime.Version(), m.GoVersion)
	}
	abs, err := filepath.Abs(*root)
	if err != nil {
		return err
	}
	if err := validateManifest(abs, m); err != nil {
		return err
	}
	if err := requireCleanReleaseSnapshot(abs, m); err != nil {
		return err
	}
	if err := validateProfileDirectory(*profilesDir); err != nil {
		return err
	}
	goBin, toolVersion, _, _, err := coverTool()
	if err != nil {
		return err
	}
	if toolVersion != m.GoVersion {
		return errors.New("resolved Go/cover version differs from the manifest")
	}
	tmpDir, err := os.MkdirTemp(*profilesDir, ".verify-source-map-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpDir)
	blocksByMode := map[string]map[string][]block{}
	manifestSum := sha256.Sum256(data)
	manifestDigest := hex.EncodeToString(manifestSum[:])
	seen := map[string]bool{}
	rows := make([]string, 0, len(m.Cohorts))
	for _, c := range m.Cohorts {
		key := c.Name + "__" + tagKey(c.Tags)
		if seen[key] {
			return fmt.Errorf("duplicate cohort %s", key)
		}
		seen[key] = true
		covered, total, err := verifyCohortProfile(abs, *profilesDir, tmpDir, goBin, toolVersion, m, data, c, blocksByMode)
		if err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		rows = append(rows, fmt.Sprintf("%s %d/%d (%s%%)", key, covered, total, ratio(covered, total)))
	}

	if len(seen) != 24 {
		return fmt.Errorf("found %d of 24 cohort profiles", len(seen))
	}
	if err := verifySourceSnapshot(abs, m); err != nil {
		return err
	}
	if err := requireCleanReleaseSnapshot(abs, m); err != nil {
		return err
	}
	sort.Strings(rows)
	fmt.Printf("verified manifest sha256=%s\n", manifestDigest)
	for _, r := range rows {
		fmt.Println(r)
	}
	return nil
}

func verifyCohortProfile(root, profilesDir, tmpDir, goBin, toolVersion string, m manifest, manifestBytes []byte, c selectedCohort, blocksByMode map[string]map[string][]block) (int64, int64, error) {
	key := c.Name + "__" + tagKey(c.Tags)
	selected := make(map[string]bool, len(c.Files))
	for _, path := range c.Files {
		if selected[path] {
			return 0, 0, fmt.Errorf("duplicate selected file %s", path)
		}
		selected[path] = true
	}
	actualPath := filepath.Join(profilesDir, key+".out")
	expectedPath := filepath.Join(profilesDir, key+".expected")
	if err := verifyProvenance(filepath.Join(profilesDir, key+".provenance.json"), m, manifestBytes, c, actualPath, expectedPath); err != nil {
		return 0, 0, err
	}
	actual, err := parseProfile(actualPath, m.ModulePath)
	if err != nil {
		return 0, 0, err
	}
	expected, err := parseProfile(expectedPath, m.ModulePath)
	if err != nil {
		return 0, 0, err
	}
	actualMode, err := profileMode(actualPath)
	if err != nil {
		return 0, 0, err
	}
	expectedMode, err := profileMode(expectedPath)
	if err != nil {
		return 0, 0, err
	}
	if actualMode != expectedMode {
		return 0, 0, fmt.Errorf("coverage mode mismatch: actual=%s expected=%s", actualMode, expectedMode)
	}
	byFile, ok := blocksByMode[expectedMode]
	if !ok {
		byFile, err = instrumentFiles(goBin, toolVersion, root, tmpDir, mapKeys(m.Files), expectedMode)
		if err != nil {
			return 0, 0, err
		}
		blocksByMode[expectedMode] = byFile
	}
	canonical := blocksForFiles(c.Files, byFile)
	zeroFiles := zeroBlockFiles(c.Files, byFile)
	if err := validateReferenceProvenance(expectedPath, manifestBytes, m, c, expectedMode, expected, canonical, zeroFiles); err != nil {
		return 0, 0, err
	}
	covered, total, err := compareCoreProfile(m, c, expected, actual)
	if err != nil {
		return 0, 0, err
	}
	if err := checkCoverage(covered, total); err != nil {
		return 0, 0, fmt.Errorf("coverage %d/%d (%s%%): %w", covered, total, ratio(covered, total), err)
	}
	return covered, total, nil
}

func verifySourceSnapshot(root string, m manifest) error {
	if len(m.ContractFiles) != 7 || len(m.BuildInputs) == 0 {
		return errors.New("manifest contract source hashes are incomplete")
	}
	for p, want := range m.Files {
		if err := validateRegularPath(root, p); err != nil {
			return err
		}
		b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if e != nil {
			return fmt.Errorf("source %s: %w", p, e)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			return fmt.Errorf("source changed since profile manifest: %s", p)
		}
	}
	for p, want := range m.ContractFiles {
		if err := validateRegularPath(root, p); err != nil {
			return err
		}
		b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if e != nil {
			return fmt.Errorf("contract source %s: %w", p, e)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			return fmt.Errorf("coverage contract changed since manifest: %s", p)
		}
	}
	for p, want := range m.BuildInputs {
		if err := validateRegularPath(root, p); err != nil {
			return err
		}
		b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if e != nil {
			return fmt.Errorf("go build input %s: %w", p, e)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			return fmt.Errorf("go build input changed since manifest: %s", p)
		}
	}
	return nil
}

func validateManifest(root string, m manifest) error {
	if m.Schema != 1 || len(m.Cohorts) != 24 || len(m.Files) != 367 {
		return errors.New("manifest schema/cohort/source inventory is incomplete")
	}
	if err := validateInventories(root); err != nil {
		return err
	}
	paths, err := readInventory(filepath.Join(root, "scripts", "coveragecontract", "core-files.txt"))
	if err != nil {
		return err
	}
	if err := validateManifestInventory(root, paths, m); err != nil {
		return err
	}
	if err := validateManifestCheckout(root, m); err != nil {
		return err
	}
	if err := validateManifestCohorts(root, paths, m); err != nil {
		return err
	}
	selector, err := selectorDigest(m)
	if err != nil {
		return err
	}
	if selector != m.SelectorSHA256 {
		return errors.New("source selector identity mismatch")
	}
	return nil
}

func validateManifestInventory(root string, paths []string, m manifest) error {
	remaining := make(map[string]bool, len(paths))
	for _, p := range paths {
		remaining[p] = true
	}
	for p := range m.Files {
		if !remaining[p] {
			return fmt.Errorf("manifest has unreviewed core source %s", p)
		}
		delete(remaining, p)
		if _, ok := m.BuildInputs[p]; !ok {
			return fmt.Errorf("core source %s missing from go build inputs", p)
		}
	}
	if len(remaining) > 0 {
		return errors.New("manifest omits reviewed core source files")
	}
	if err := validateBuildInputSet(root, m.BuildInputs); err != nil {
		return err
	}
	expected := []string{"scripts/coveragecontract/main.go", coreInventoryPath, exclusionInventoryPath, "go.mod", "go.sum", ".goreleaser.yml", ".make/test.mk"}
	if len(m.ContractFiles) != len(expected) {
		return errors.New("manifest contract file list is incomplete")
	}
	for _, p := range expected {
		if _, ok := m.ContractFiles[p]; !ok {
			return fmt.Errorf("contract file missing from manifest: %s", p)
		}
	}
	return nil
}

func validateBuildInputSet(root string, buildInputs map[string]string) error {
	tracked, err := gitOutput(root, "ls-files", "--", "*.go")
	if err != nil {
		return err
	}
	untracked, err := gitOutput(root, "ls-files", "--others", "--exclude-standard", "--", "*.go")
	if err != nil {
		return err
	}
	inputs := map[string]bool{}
	for _, list := range []string{tracked, untracked} {
		for _, p := range strings.Split(list, "\n") {
			if p != "" {
				inputs[p] = true
			}
		}
	}
	if len(inputs) != len(buildInputs) {
		return errors.New("tracked/untracked go build input list changed since manifest")
	}
	for p := range inputs {
		if _, ok := buildInputs[p]; !ok {
			return fmt.Errorf("go build input missing from manifest: %s", p)
		}
	}
	return nil
}

func validateManifestCheckout(root string, m manifest) error {
	if m.SourceCommit == "" || m.SourceTree == "" {
		return errors.New("manifest commit/tree metadata is missing")
	}
	_, version, _, _, err := coverTool()
	if err != nil {
		return err
	}
	if m.GoVersion != version || m.GoVersion != runtime.Version() {
		return errors.New("manifest Go version mismatch")
	}
	head, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	tree, err := gitOutput(root, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return err
	}
	if head != m.SourceCommit || tree != m.SourceTree {
		return errors.New("manifest commit/tree does not match current checkout")
	}
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return err
	}
	if m.Dirty != (strings.TrimSpace(status) != "") {
		return errors.New("manifest dirty-worktree state does not match current checkout")
	}
	return verifySourceSnapshot(root, m)
}

func validateManifestCohorts(root string, paths []string, m manifest) error {
	want := make(map[string]selectedCohort, len(m.Cohorts))
	for _, c := range expectedCohorts() {
		want[c.Name+"__"+tagKey(c.Tags)] = c
	}
	seen := map[string]bool{}
	for _, c := range m.Cohorts {
		if err := validateCohortSelection(root, paths, c, want, seen); err != nil {
			return err
		}
	}
	if len(seen) != 24 {
		return fmt.Errorf("manifest has %d of 24 cohorts", len(seen))
	}
	return validateCohortUnion(mapKeys(m.Files), m.Cohorts)
}

func validateCohortSelection(root string, paths []string, c selectedCohort, want map[string]selectedCohort, seen map[string]bool) error {
	key := c.Name + "__" + tagKey(c.Tags)
	base, ok := want[key]
	if !ok || seen[key] {
		return fmt.Errorf("unknown or duplicate cohort %s", key)
	}
	seen[key] = true
	if c.GOOS != base.GOOS || c.GOARCH != base.GOARCH || c.CGOEnabled != base.CGOEnabled {
		return fmt.Errorf("cohort platform/CGO metadata mismatch for %s", key)
	}
	ctx := releaseBuildContext(c.GOOS, c.GOARCH, c.Tags)
	selected := make([]string, 0, len(paths))
	for _, p := range paths {
		match, err := ctx.MatchFile(root, p)
		if err != nil {
			return fmt.Errorf("reselect %s: %w", p, err)
		}
		if match {
			selected = append(selected, p)
		}
	}
	if !sameStrings(selected, c.Files) {
		return fmt.Errorf("selected source list mismatch for %s", key)
	}
	return nil
}

func selectorDigest(m manifest) (string, error) {
	// 호스트별 실행 파일 지문은 공유 선택자 식별자에 포함하지 않는다.
	m.SelectorSHA256 = ""
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validateCohortUnion(files []string, cohorts []selectedCohort) error {
	want := map[string]bool{}
	for _, p := range files {
		want[p] = true
	}
	got := map[string]bool{}
	for _, c := range cohorts {
		for _, p := range c.Files {
			got[p] = true
		}
	}
	for p := range want {
		if !got[p] {
			return fmt.Errorf("reviewed core source %s is not selected by any release cohort", p)
		}
	}
	for p := range got {
		if !want[p] {
			return fmt.Errorf("cohort selects unreviewed core source %s", p)
		}
	}
	return nil
}

func releaseBuildContext(goos, goarch string, tags []string) build.Context {
	ctx := build.Default
	ctx.GOOS = goos
	ctx.GOARCH = goarch
	ctx.CgoEnabled = false
	ctx.BuildTags = append([]string(nil), tags...)
	// 릴리스 기본 GOAMD64/GOARM64를 고정하며 preparing host의 feature tag를 복사하지 않는다.
	ctx.ToolTags = []string{"gc"}
	switch goarch {
	case "amd64":
		ctx.ToolTags = append(ctx.ToolTags, "amd64.v1")
	case "arm64":
		ctx.ToolTags = append(ctx.ToolTags, "arm64.v8.0")
	}
	return ctx
}

func requireCleanReleaseSnapshot(root string, m manifest) error {
	if m.Dirty {
		return errors.New("release coverage binding rejects a dirty source manifest; dirty preparation is not release qualification")
	}
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) != "" {
		return errors.New("release coverage binding requires a clean source worktree")
	}
	return nil
}

func validateInventories(root string) error {
	for path, want := range map[string]string{coreInventoryPath: coreInventorySHA256, exclusionInventoryPath: exclusionInventorySHA256} {
		if err := validateRegularPath(root, path); err != nil {
			return err
		}
		got, err := fileSHA(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("reviewed source inventory changed: %s", path)
		}
	}
	exclusions, err := readPathFile(filepath.Join(root, "scripts", "coveragecontract", "reviewed-exclusions.txt"))
	if err != nil {
		return err
	}
	if len(exclusions) != 56 {
		return fmt.Errorf("reviewed source exclusion inventory has %d entries, want 56", len(exclusions))
	}
	if err := validateRegularPath(root, ".goreleaser.yml"); err != nil {
		return err
	}
	config, err := os.ReadFile(filepath.Join(root, ".goreleaser.yml"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(config), "CGO_ENABLED=0") {
		return errors.New("release build no longer declares CGO_ENABLED=0")
	}
	return nil
}

func validateRegularPath(root, p string) error {
	if p == "" || filepath.IsAbs(p) || filepath.ToSlash(filepath.Clean(filepath.FromSlash(p))) != p {
		return fmt.Errorf("unsafe relative source path %q", p)
	}
	current := root
	parts := strings.Split(p, "/")
	for i, part := range parts {
		if part == ".." || part == "." || part == "" {
			return fmt.Errorf("unsafe source path component %q", p)
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("source path %s: %w", p, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source path %s traverses a symlink", p)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("source path parent is not a directory: %s", current)
		}
		if i == len(parts)-1 && !info.Mode().IsRegular() {
			return fmt.Errorf("source is not a regular file: %s", p)
		}
	}
	return nil
}

func verifyProvenance(path string, m manifest, manifestBytes []byte, c selectedCohort, profile, expected string) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return fmt.Errorf("provenance %s: %w", path, e)
	}
	var p provenance
	if e = json.Unmarshal(b, &p); e != nil {
		return fmt.Errorf("decode provenance: %w", e)
	}
	sum := sha256.Sum256(manifestBytes)
	manifestDigest := hex.EncodeToString(sum[:])
	if !profileMatchesManifest(p, m, c, manifestDigest) {
		return errors.New("provenance does not match manifest, toolchain, or cohort")
	}
	if !validNativeToolIdentity(p.ToolGOOS, p.ToolGOARCH, p.GoBinarySHA256, p.CoverSHA256, c) {
		return errors.New("profile provenance lacks matching native tool identity")
	}
	refBytes, e := os.ReadFile(expected + ".provenance.json")
	if e != nil {
		return e
	}
	var ref referenceProvenance
	if err := json.Unmarshal(refBytes, &ref); err != nil {
		return err
	}
	if !sameProfileReferenceTool(p, ref, manifestDigest) {
		return errors.New("profile and expected-map native tool provenance differ")
	}
	actualMode, e := profileMode(profile)
	if e != nil {
		return e
	}
	expectedMode, e := profileMode(expected)
	if e != nil {
		return e
	}
	if actualMode != expectedMode || p.Mode != actualMode || p.ExpectedMode != expectedMode {
		return errors.New("provenance coverage mode does not match profile inputs")
	}
	return validateProfileArtifactDigests(p, profile, expected)
}

func profileMatchesManifest(p provenance, m manifest, c selectedCohort, digest string) bool {
	return p.ManifestSHA256 == digest && p.SourceCommit == m.SourceCommit && p.GoVersion == m.GoVersion &&
		p.Cohort.Name == c.Name && p.Cohort.GOOS == c.GOOS && p.Cohort.GOARCH == c.GOARCH &&
		p.Cohort.CGOEnabled == c.CGOEnabled && tagKey(p.Cohort.Tags) == tagKey(c.Tags)
}

func validNativeToolIdentity(goos, goarch, goHash, coverHash string, c selectedCohort) bool {
	return goos == c.GOOS && goarch == c.GOARCH && validSHA256(goHash) && validSHA256(coverHash)
}

func sameProfileReferenceTool(p provenance, ref referenceProvenance, manifestDigest string) bool {
	return p.ToolGOOS == ref.ToolGOOS && p.ToolGOARCH == ref.ToolGOARCH &&
		p.GoBinarySHA256 == ref.GoBinarySHA256 && p.CoverSHA256 == ref.CoverSHA256 &&
		ref.ManifestSHA256 == manifestDigest
}

func validateProfileArtifactDigests(p provenance, profile, expected string) error {
	actualSum, err := fileSHA(profile)
	if err != nil {
		return err
	}
	expectedSum, err := fileSHA(expected)
	if err != nil {
		return err
	}
	if p.ProfileSHA256 != actualSum || p.ExpectedSHA256 != expectedSum {
		return errors.New("profile digest does not match provenance")
	}
	expectedProvenanceSum, err := fileSHA(expected + ".provenance.json")
	if err != nil {
		return err
	}
	if p.ExpectedProvenanceSHA256 != expectedProvenanceSum {
		return errors.New("go cover instrumenter provenance digest mismatch")
	}
	return nil
}

func validateProfileDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read coverage profile directory: %w", err)
	}
	want := map[string]bool{}
	for _, c := range expectedCohorts() {
		key := c.Name + "__" + tagKey(c.Tags)
		for _, suffix := range []string{".out", ".expected", ".expected.provenance.json", ".provenance.json"} {
			want[key+suffix] = true
		}
	}
	if len(entries) != len(want) {
		return fmt.Errorf("profile directory has %d entries; need exactly %d (4 for each of 24 cohorts)", len(entries), len(want))
	}
	for _, entry := range entries {
		if !want[entry.Name()] {
			return fmt.Errorf("unknown profile artifact %s", entry.Name())
		}
		info, err := os.Lstat(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("profile artifact is not a regular file: %s", entry.Name())
		}
		delete(want, entry.Name())
	}
	if len(want) > 0 {
		return fmt.Errorf("missing profile artifact %s", firstKey(want))
	}
	return nil
}

func firstKey(m map[string]bool) string {
	for k := range m {
		return k
	}
	return ""
}

func fileSHA(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func parseTags(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	sort.Strings(parts)
	return parts
}

func compareBlocks(selected, known, core map[string]bool, expected, actual []block) (int64, int64, error) {
	expectedMap := map[string]block{}
	actualMap := map[string]block{}
	for _, b := range expected {
		if !selected[b.File] {
			return 0, 0, fmt.Errorf("expected profile has unknown/unselected source %s", b.File)
		}
		k := blockKey(b)
		if _, ok := expectedMap[k]; ok {
			return 0, 0, fmt.Errorf("duplicate expected block %s", k)
		}
		expectedMap[k] = b
	}
	for _, b := range actual {
		if !known[b.File] {
			return 0, 0, fmt.Errorf("actual profile has unknown source %s", b.File)
		}
		if !selected[b.File] {
			if core[b.File] {
				return 0, 0, fmt.Errorf("actual profile has core source not selected for this cohort: %s", b.File)
			}
			continue
		}
		k := blockKey(b)
		if _, ok := actualMap[k]; ok {
			return 0, 0, fmt.Errorf("duplicate actual block %s", k)
		}
		actualMap[k] = b
	}
	var hit, total int64
	for k, e := range expectedMap {
		a, ok := actualMap[k]
		if !ok {
			return 0, 0, fmt.Errorf("missing instrumented block %s", k)
		}
		if a.NumStmt != e.NumStmt {
			return 0, 0, fmt.Errorf("NumStmt mismatch for %s: expected %d got %d", k, e.NumStmt, a.NumStmt)
		}
		total += int64(a.NumStmt)
		if a.Count > 0 {
			hit += int64(a.NumStmt)
		}
	}
	if len(actualMap) != len(expectedMap) {
		return 0, 0, fmt.Errorf("actual profile has %d blocks; expected %d", len(actualMap), len(expectedMap))
	}
	return hit, total, nil
}

func parseProfile(path, module string) ([]block, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("profile %s: %w", path, err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 4096), 1024*1024)
	if !s.Scan() {
		return nil, fmt.Errorf("profile %s is empty", path)
	}
	mode := strings.TrimPrefix(s.Text(), "mode: ")
	if mode != "set" && mode != "count" && mode != "atomic" {
		return nil, fmt.Errorf("profile %s has unsupported mode %q", path, mode)
	}
	seen := map[string]bool{}
	out := make([]block, 0)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		parsed, err := parseProfileRow(line, module, mode, seen)
		if err != nil {
			return nil, err
		}
		out = append(out, parsed)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("profile %s has no blocks", path)
	}
	return out, nil
}

func parseProfileRow(line, module, mode string, seen map[string]bool) (block, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return block{}, fmt.Errorf("malformed profile row %q", line)
	}
	parts := strings.SplitN(fields[0], ":", 2)
	if len(parts) != 2 {
		return block{}, fmt.Errorf("malformed profile coordinate %q", fields[0])
	}
	rangeParts := strings.Split(parts[1], ",")
	if len(rangeParts) != 2 {
		return block{}, fmt.Errorf("malformed profile range %q", fields[0])
	}
	start, end := strings.Split(rangeParts[0], "."), strings.Split(rangeParts[1], ".")
	if len(start) != 2 || len(end) != 2 {
		return block{}, fmt.Errorf("malformed profile range %q", fields[0])
	}
	values, err := parseProfileIntegers(start, end, fields[1], fields[2])
	if err != nil {
		return block{}, fmt.Errorf("malformed profile row %q: %w", line, err)
	}
	if mode == "set" && values[5] > 1 {
		return block{}, fmt.Errorf("set-mode count must be 0 or 1 in %q", line)
	}
	file := strings.TrimPrefix(filepath.ToSlash(parts[0]), module+"/")
	parsed := block{File: file, StartLine: values[0], StartCol: values[1], EndLine: values[2], EndCol: values[3], NumStmt: values[4], Count: values[5]}
	if parsed.NumStmt < 0 {
		return block{}, fmt.Errorf("invalid NumStmt in %q", line)
	}
	key := blockKey(parsed)
	if seen[key] {
		return block{}, fmt.Errorf("duplicate block %s", key)
	}
	seen[key] = true
	return parsed, nil
}

func parseProfileIntegers(start, end []string, statements, count string) ([]int, error) {
	values := make([]int, 6)
	for i, raw := range []string{start[0], start[1], end[0], end[1], statements, count} {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return nil, fmt.Errorf("invalid integer %q", raw)
		}
		values[i] = value
	}
	return values, nil
}

func profileMode(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return "", fmt.Errorf("profile %s is empty", path)
	}
	line := s.Text()
	if !strings.HasPrefix(line, "mode: ") {
		return "", fmt.Errorf("profile %s missing mode header", path)
	}
	mode := strings.TrimPrefix(line, "mode: ")
	if mode != "set" && mode != "count" && mode != "atomic" {
		return "", fmt.Errorf("profile %s has unsupported mode %q", path, mode)
	}
	return mode, nil
}

func blockKey(b block) string {
	return fmt.Sprintf("%s:%d.%d,%d.%d", b.File, b.StartLine, b.StartCol, b.EndLine, b.EndCol)
}

func tagKey(tags []string) string {
	if len(tags) == 0 {
		return "default"
	}
	return strings.Join(tags, "+")
}

func ratio(a, b int64) string {
	if b <= 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d.%02d", a*100/b, (a*10000/b)%100)
}

func checkCoverage(covered, total int64) error {
	if total <= 0 {
		return errors.New("profile has no statements")
	}
	if covered < 0 || covered > total {
		return errors.New("covered statement count is invalid")
	}
	if covered*100 < total*threshold {
		return errors.New("below 80%")
	}
	return nil
}

func expectedCohorts() []selectedCohort {
	out := make([]selectedCohort, 0, len(cohorts)*len(tagSets))
	for _, c := range cohorts {
		for _, tags := range tagSets {
			out = append(out, selectedCohort{Name: c.Name, GOOS: c.GOOS, GOARCH: c.GOARCH, CGOEnabled: false, Tags: tags})
		}
	}
	return out
}

func modulePath(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "module ")), "\"`"), nil
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", errors.New("module directive missing")
}

func readInventory(path string) ([]string, error) {
	if info, err := os.Lstat(path); err != nil {
		return nil, err
	} else if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("inventory is not a regular file: %s", path)
	}
	out, e := readPathFile(path)
	if e != nil {
		return nil, e
	}
	if len(out) != 367 {
		return nil, fmt.Errorf("reviewed core inventory has %d entries, want 367", len(out))
	}
	return out, nil
}

func readPathFile(path string) ([]string, error) {
	if info, err := os.Lstat(path); err != nil {
		return nil, err
	} else if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("inventory is not a regular file: %s", path)
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	var out []string
	seen := map[string]bool{}
	for s.Scan() {
		p := strings.TrimSpace(s.Text())
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		unsafe := filepath.IsAbs(p) || filepath.ToSlash(filepath.Clean(filepath.FromSlash(p))) != p || seen[p]
		for _, part := range strings.Split(p, "/") {
			if part == ".." || part == "." || part == "" {
				unsafe = true
			}
		}
		if unsafe {
			return nil, fmt.Errorf("invalid/duplicate inventory path %q", p)
		}
		seen[p] = true
		out = append(out, p)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func gitOutput(root string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	out, e := cmd.Output()
	return strings.TrimSpace(string(out)), e
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "coveragecontract: "+format+"\n", args...)
	os.Exit(1)
}
