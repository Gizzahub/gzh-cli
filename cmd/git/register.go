package git

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type gitCmdProvider struct {
	appCtx *app.AppContext
}

func (p gitCmdProvider) Command() *cobra.Command {
	return NewGitCmd(p.appCtx)
}

func (p gitCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "git",
		Category:     registry.CategoryGit,
		Version:      "1.0.0",
		Priority:     10,
		Experimental: false,
		Dependencies: []string{"git"},
		Tags:         []string{"git", "repository", "vcs", "clone", "pull"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterGitCmd(appCtx *app.AppContext) {
	registry.Register(gitCmdProvider{appCtx: appCtx})
}
