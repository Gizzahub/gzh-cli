// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

type syncCloneGiteaOptions struct {
	targetPath string
	orgName    string
	strategy   string
}

func defaultSyncCloneGiteaOptions() *syncCloneGiteaOptions {
	return &syncCloneGiteaOptions{
		strategy: "reset",
	}
}

func newSyncCloneGiteaCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	o := defaultSyncCloneGiteaOptions()

	cmd := &cobra.Command{
		Use:   "gitea",
		Short: "Clone repositories from a Gitea organization",
		Args:  cobra.NoArgs,
		RunE:  o.run,
	}

	cmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "targetPath")
	cmd.Flags().StringVarP(&o.orgName, "orgName", "o", o.orgName, "orgName")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")

	return cmd
}

func (o *syncCloneGiteaOptions) run(_ *cobra.Command, args []string) error {
	if o.targetPath == "" || o.orgName == "" {
		return fmt.Errorf("both targetPath and orgName must be specified")
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// err := gitea_org.RefreshAll(o.targetPath, o.orgName, o.strategy)
	// if err != nil {
	//	return err
	//}

	return nil
}
