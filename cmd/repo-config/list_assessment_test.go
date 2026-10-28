// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/services"
)

func TestRepositoryAssessmentTable(t *testing.T) {
	for _, repos := range [][]services.RepositoryInfo{nil, {{Name: "assigned", Compliant: true}, {Name: "unassigned", Compliant: false}}} {
		previous := os.Stdout
		reader, writer, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = writer
		printTableFormat(repos, false)
		os.Stdout = previous
		require.NoError(t, writer.Close())
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		text := string(data)
		require.Contains(t, text, "Compliance: not assessed")
		require.NotContains(t, text, "NaN")
		require.NotContains(t, text, "✓")
		require.NotContains(t, text, "✗")
		require.NotContains(t, text, "100.0%")
		if len(repos) > 0 {
			require.Contains(t, text, "NOT_ASSESSED")
		}
	}
}
