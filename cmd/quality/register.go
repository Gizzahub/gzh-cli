package quality

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type qualityCmdProvider struct {
	appCtx *app.AppContext
}

func (p qualityCmdProvider) Command() *cobra.Command {
	return NewQualityCmd(p.appCtx)
}

func RegisterQualityCmd(appCtx *app.AppContext) {
	registry.Register(qualityCmdProvider{appCtx: appCtx})
}
