package netenv

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type netEnvCmdProvider struct {
	appCtx *app.AppContext
}

func (p netEnvCmdProvider) Command() *cobra.Command {
	return NewNetEnvCmd(context.Background(), p.appCtx)
}

func RegisterNetEnvCmd(appCtx *app.AppContext) {
	registry.Register(netEnvCmdProvider{appCtx: appCtx})
}
