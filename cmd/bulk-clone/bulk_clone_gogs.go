package bulk_clone

import (
	"fmt"

	"github.com/spf13/cobra"
)

type bulkCloneGogsOptions struct {
	targetPath string
	orgName    string
	strategy   string
}

func defaultBulkCloneGogsOptions() *bulkCloneGogsOptions {
	return &bulkCloneGogsOptions{
		strategy: "reset",
	}
}

func newBulkCloneGogsCmd() *cobra.Command {
	o := defaultBulkCloneGogsOptions()

	cmd := &cobra.Command{
		Use:   "gogs",
		Short: "Clone repositories from a Gogs organization",
		Args:  cobra.NoArgs,
		RunE:  o.run,
	}

	cmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "targetPath")
	cmd.Flags().StringVarP(&o.orgName, "orgName", "o", o.orgName, "orgName")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")

	cmd.MarkFlagRequired("targetPath")
	cmd.MarkFlagRequired("orgName")

	return cmd
}

func (o *bulkCloneGogsOptions) run(_ *cobra.Command, args []string) error {
	if o.targetPath == "" || o.orgName == "" {
		return fmt.Errorf("both targetPath and orgName must be specified")
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// TODO: Implement when gogs package is ready
	//err := gogs.RefreshAll(o.targetPath, o.orgName, o.strategy)
	//if err != nil {
	//	return err
	//}

	return nil
}
