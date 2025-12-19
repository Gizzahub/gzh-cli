package gitsync

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type gitSyncCmdProvider struct {
	appCtx *app.AppContext
}

func (p gitSyncCmdProvider) Command() *cobra.Command {
	return NewGitSyncCmd(p.appCtx)
}

func (p gitSyncCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "git-sync",
		Category:     registry.CategoryConfig,
		Version:      "0.1.0",
		Priority:     31,
		Experimental: false,
		Dependencies: []string{"git"},
		Tags:         []string{"sync", "clone", "repos", "git"},
		Lifecycle:    registry.LifecycleStable,
	}
}

// RegisterGitSyncCmd는 `git-sync` 커맨드를 registry에 등록합니다.
func RegisterGitSyncCmd(appCtx *app.AppContext) {
	registry.Register(gitSyncCmdProvider{appCtx: appCtx})
}
