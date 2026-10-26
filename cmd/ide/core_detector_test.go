// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package ide

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/idecore"
)

// 공개 facade와 타입 별칭이 공유 backend 계약을 유지하는지 확인한다.
func TestDetectorFacade(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var detector *idecore.IDEDetector = NewIDEDetector()
	var contract idecore.IDEDetectorInterface = detector
	require.NotNil(t, contract)
	ides := []IDE{{Name: "Fixture Editor", Aliases: []string{"fixture"}}}
	require.Equal(t, &ides[0], detector.FindIDEByAlias(ides, "FIXTURE"))
}
