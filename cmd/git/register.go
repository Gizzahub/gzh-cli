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

func RegisterGitCmd(appCtx *app.AppContext) {
	registry.Register(gitCmdProvider{appCtx: appCtx})
}
