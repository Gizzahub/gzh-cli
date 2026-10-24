//go:build gosecparity

// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package acceptedrisk

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gosecBinaryEnv names the pinned gosec binary. make security-gosec-parity sets
// it to the binary install-gosec just verified.
const gosecBinaryEnv = "GOSEC_BIN"

// gosecParityRule is the rule every fixture call site triggers. G304 fires on
// os.ReadFile with a caller-supplied path, so a site that gosec does not report
// as suppressed or unsuppressed means the fixture, not the matcher, is wrong.
const gosecParityRule = "G304"

// parityDirective is the registrable directive form, in the exact grammar
// directivePattern accepts.
const parityDirective = "\t//gosec:disable G304 -- AR-2026-001 the caller owns the path.\n"

var (
	gosecVersionPattern = regexp.MustCompile(`(?m)^GOSEC_VERSION[ \t]*:?=[ \t]*(\S+)[ \t]*$`)
	gosecModulePattern  = regexp.MustCompile(`(?m)^GOSEC_MODULE[ \t]*:?=[ \t]*(\S+)[ \t]*$`)
)

// gosecReportSuppression is one entry of an issue's suppressions in gosec's
// JSON report.
type gosecReportSuppression struct {
	Kind string `json:"kind"`
}

// paritySite is one call site in the generated fixture.
type paritySite struct {
	name string
	// tag is the blanket tag the comment is spelled with, or empty for the
	// directive and the uncommented control.
	tag string
	// directive marks the //gosec:disable site.
	directive    bool
	commentStart int
	callLine     int
}

// TestGosecParity runs the claims in gosecTagPlacementCases and
// gosecBlanketConfigCases through the gosec binary pinned by GOSEC_VERSION in
// .make/tools.mk, and fails when that binary suppresses a different set of call
// sites than the matcher predicts.
//
// It is behind the gosecparity build tag because it needs the binary. With the
// tag set, a missing binary or a binary built from a different version is a
// failure rather than a skip: a guard that skips when its input is absent
// reports a measurement nobody made.
func TestGosecParity(t *testing.T) {
	binary := pinnedGosecBinary(t)
	requireParityTagsCoverCases(t)

	dir := t.TempDir()
	sites := writeParityFixture(t, dir)

	scope := testScanScope(t)
	for _, configCase := range gosecBlanketConfigCases {
		t.Run(configCase.name, func(t *testing.T) {
			measured := runGosecOnFixture(t, binary, dir, configCase.config)

			predicted := predictSuppressedSites(t, dir, scope, sites, configCase.honored, configCase.directiveHonored)

			// Every site must be reported one way or the other; a site gosec
			// does not report at all proves nothing about suppression.
			for _, site := range sites {
				_, seen := measured[site.callLine]
				assert.True(t, seen, "gosec reported no %s at %s (line %d); the fixture no longer exercises the rule",
					gosecParityRule, site.name, site.callLine)
			}

			// Compare gosec with the production derivation, not only with the
			// hand-kept honored column: a tag gosec honored must be one the
			// gate derives from the same configuration.
			derived, err := gosecBlanketTokens([]byte(configCase.config))
			require.NoError(t, err)

			measuredSuppressed := make([]string, 0)
			for _, site := range sites {
				if measured[site.callLine] {
					measuredSuppressed = append(measuredSuppressed, site.name)
					if site.tag != "" {
						assert.Contains(t, derived, site.tag,
							"gosec honored %s at %s, but gosecBlanketTokens does not derive it", site.tag, site.name)
					}
				}
			}
			sort.Strings(measuredSuppressed)
			assert.Equal(t, predicted, measuredSuppressed,
				"the pinned gosec suppressed a different set of call sites than the acceptedrisk matcher predicts")
		})
	}
}

// pinnedGosecBinary resolves the binary and proves it is the pinned release.
func pinnedGosecBinary(t *testing.T) string {
	t.Helper()

	binary := os.Getenv(gosecBinaryEnv)
	require.NotEmpty(t, binary, "%s must name the pinned gosec binary; run make security-gosec-parity", gosecBinaryEnv)
	info, err := os.Stat(binary)
	require.NoError(t, err, "%s=%s", gosecBinaryEnv, binary)
	require.False(t, info.IsDir(), "%s=%s is a directory", gosecBinaryEnv, binary)

	contents, err := fs.ReadFile(os.DirFS(repositoryRoot), toolsMakefilePath)
	require.NoError(t, err)
	wantVersion := gosecVersionPattern.FindSubmatch(contents)
	require.NotNil(t, wantVersion, "GOSEC_VERSION is not declared in %s", toolsMakefilePath)
	wantModule := gosecModulePattern.FindSubmatch(contents)
	require.NotNil(t, wantModule, "GOSEC_MODULE is not declared in %s", toolsMakefilePath)

	// buildinfo reads the same module record `go version -m` prints, which is
	// what install-gosec verifies.
	built, err := buildinfo.ReadFile(binary)
	require.NoError(t, err, "read build information from %s", binary)
	require.Equal(t, string(wantModule[1]), built.Main.Path, "%s is not built from the pinned gosec module", binary)
	require.Equal(t, string(wantVersion[1]), built.Main.Version,
		"%s is not the gosec release pinned by GOSEC_VERSION in %s", binary, toolsMakefilePath)
	return binary
}

// parityTags is every blanket tag the fixture spells, so each tag is measured
// both where it is live and where it is inert. It is a literal rather than
// derived from the want column: a derivation that drops a tag, with the table
// edited to match, would otherwise remove that tag's sites from the fixture and
// leave nothing to contradict it.
var parityTags = []string{"#false", "#nosec", "#skipme", "#true"}

// requireParityTagsCoverCases keeps the literal tag list from falling behind
// the configuration cases.
func requireParityTagsCoverCases(t *testing.T) {
	t.Helper()
	for _, configCase := range gosecBlanketConfigCases {
		for _, tag := range slices.Concat(configCase.want, configCase.honored) {
			require.Contains(t, parityTags, tag, "parityTags must spell %s for %q", tag, configCase.name)
		}
	}
}

// writeParityFixture writes a standalone module whose call sites differ only by
// the comment above them, and returns where each site sits.
//
// The fixture is generated rather than committed as .go source because the
// formatters would rewrite it: gofumpt inserts a space after "//", which turns
// the no-space placement cases into different cases.
func writeParityFixture(t *testing.T, dir string) []paritySite {
	t.Helper()

	var source strings.Builder
	source.WriteString("package parity\n\nimport \"os\"\n")
	line := 3

	var sites []paritySite
	addSite := func(site paritySite, comment string) {
		source.WriteString("\n")
		line++
		fmt.Fprintf(&source, "func site%d(p string) ([]byte, error) {\n", len(sites))
		line++
		site.commentStart = line + 1
		source.WriteString(comment)
		line += strings.Count(comment, "\n")
		source.WriteString("\treturn os.ReadFile(p)\n}\n")
		site.callLine = line + 1
		line += 2
		sites = append(sites, site)
	}

	addSite(paritySite{name: "control without a comment"}, "")
	addSite(paritySite{name: "registrable directive", directive: true}, parityDirective)
	addSite(paritySite{name: "directive with initial tab", directive: true}, "\t//gosec:disable\tG304 -- AR-2026-001 the caller owns the path.\n")
	addSite(paritySite{name: "directive with space then tab", directive: true}, "\t//gosec:disable \tG304 -- AR-2026-001 the caller owns the path.\n")
	addSite(paritySite{name: "directive with tab-separated fields", directive: true}, "\t//gosec:disable G304\t--\tAR-2026-001\tthe caller owns the path.\n")
	for _, tag := range parityTags {
		for _, placement := range gosecTagPlacementCases {
			comment := strings.ReplaceAll(placement.comment, "#NOSEC", strings.ToUpper(tag))
			comment = strings.ReplaceAll(comment, "#nosec", tag)
			addSite(paritySite{name: tag + " " + placement.name, tag: tag}, comment)
		}
	}

	lines := strings.Split(source.String(), "\n")
	for _, site := range sites {
		require.Contains(t, lines[site.callLine-1], "os.ReadFile", "fixture line bookkeeping for %s", site.name)
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module parity\n\ngo 1.22\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "parity.go"), []byte(source.String()), 0o600))
	return sites
}

// runGosecOnFixture runs the pinned binary on the fixture under one
// configuration and returns, for each line gosec reported the parity rule on,
// whether it reported the finding as suppressed.
func runGosecOnFixture(t *testing.T, binary, dir, config string) map[int]bool {
	t.Helper()

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "gosec.json")
	reportPath := filepath.Join(configDir, "report.json")
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary,
		"-conf="+configPath,
		"-include="+gosecParityRule,
		"-track-suppressions",
		"-no-fail",
		"-fmt=json",
		"-out="+reportPath,
		"./...",
	)
	command.Dir = dir
	// The fixture is its own module outside any workspace.
	command.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=")
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	require.NoError(t, command.Run(), "gosec failed:\n%s", output.String())

	reportContents, err := os.ReadFile(reportPath)
	require.NoError(t, err)

	var report struct {
		Issues []struct {
			RuleID       string                   `json:"rule_id"`
			File         string                   `json:"file"`
			Line         string                   `json:"line"`
			Suppressions []gosecReportSuppression `json:"suppressions"`
		} `json:"Issues"`
		GolangErrors map[string]json.RawMessage `json:"Golang errors"`
	}
	require.NoError(t, json.Unmarshal(reportContents, &report), "decode gosec report")
	require.Empty(t, report.GolangErrors, "gosec could not load the fixture")

	measured := make(map[int]bool)
	for _, issue := range report.Issues {
		if issue.RuleID != gosecParityRule || filepath.Base(issue.File) != "parity.go" {
			continue
		}
		reportedLine, convErr := strconv.Atoi(issue.Line)
		require.NoError(t, convErr, "gosec issue line %q", issue.Line)
		// Only in-source suppressions are comments; an external one would come
		// from configuration the fixture does not set.
		measured[reportedLine] = slices.ContainsFunc(issue.Suppressions, func(s gosecReportSuppression) bool {
			return s.Kind == "inSource"
		})
	}
	return measured
}

// predictSuppressedSites asks the matcher which sites the pinned gosec should
// suppress when it honors exactly the given blanket tags.
func predictSuppressedSites(
	t *testing.T,
	dir string,
	scope scanScope,
	sites []paritySite,
	honored blanketTokens,
	directiveHonored bool,
) []string {
	t.Helper()

	// scanSuppressions requires at least one tag, so with none honored this
	// helper cannot ask the matcher about the directive. No case needs that
	// today; fail rather than predict it by hand.
	require.False(t, len(honored) == 0 && directiveHonored,
		"predictSuppressedSites cannot model an honored directive with no honored blanket tag")

	var found []suppression
	if len(honored) > 0 {
		var err error
		found, err = scanSuppressions(os.DirFS(dir), scope, honored)
		require.NoError(t, err)
	}

	predicted := make([]string, 0)
	for _, site := range sites {
		for _, current := range found {
			if current.Line < site.commentStart || current.Line >= site.callLine {
				continue
			}
			if current.Blanket || (directiveHonored && current.Rule == gosecParityRule && current.RiskID != "") {
				predicted = append(predicted, site.name)
				break
			}
		}
	}
	sort.Strings(predicted)
	return predicted
}
