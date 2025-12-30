// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"github.com/spf13/cobra"

	repopkg "github.com/gizzahub/gzh-cli/cmd/git/repo"
)

// NewGitRepoCmd 어댑터: 기존 공개 API 유지, 내부는 하위 패키지로 위임
func NewGitRepoCmd() *cobra.Command {
	return repopkg.NewCmd()
}
