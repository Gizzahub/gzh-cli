package synclone

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type syncCloneCmdProvider struct{}

func (syncCloneCmdProvider) Command() *cobra.Command {
	return NewSyncCloneCmd(context.Background())
}

func init() {
	registry.Register(syncCloneCmdProvider{})
}
