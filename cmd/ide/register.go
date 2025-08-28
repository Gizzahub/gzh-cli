package ide

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type ideCmdProvider struct {
	appCtx *app.AppContext
}

func (p ideCmdProvider) Command() *cobra.Command {
	return NewIDECmd(context.Background(), p.appCtx)
}

func RegisterIDECmd(appCtx *app.AppContext) {
	registry.Register(ideCmdProvider{appCtx: appCtx})
}
