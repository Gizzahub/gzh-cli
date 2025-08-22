// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import "github.com/spf13/cobra"

// NewGitRepoCmd 어댑터: 기존 테스트 호환용 래퍼
// 기존 코드와 테스트에서 사용하던 시그니처를 유지한다
func NewGitRepoCmd() *cobra.Command {
	return NewCmd()
}
