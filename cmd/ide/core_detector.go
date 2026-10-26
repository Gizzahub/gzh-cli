// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package ide

import "github.com/gizzahub/gzh-cli/internal/idecore"

// 공개 타입은 기존 호출자가 같은 backend를 사용하도록 유지한다.
type (
	IDE         = idecore.IDE
	IDECache    = idecore.IDECache
	IDEDetector = idecore.IDEDetector
)

// NewIDEDetector는 공유 backend의 기본 감지기를 반환한다.
func NewIDEDetector() *IDEDetector {
	return idecore.NewIDEDetector()
}
