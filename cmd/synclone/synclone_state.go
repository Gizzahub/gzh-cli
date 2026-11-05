// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/app"
	synclonepkg "github.com/gizzahub/gzh-cli/pkg/synclone"
)

func newSyncCloneStateCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage bulk clone operation states",
		Long: `Manage bulk clone operation states including listing, showing details, and cleaning up saved states.

States are automatically created when using resumable clone operations and allow you to resume interrupted operations.
Without --target, commands read ~/.gzh/state. With --target, they read only <target>/.gzh/state.
The two directories are not combined, and resume keeps writing under the clone target.`,
	}
	cmd.PersistentFlags().String("target", "", "Clone directory whose <target>/.gzh/state is read. Omit to read ~/.gzh/state only")

	cmd.AddCommand(newBulkCloneStateListCmd())
	cmd.AddCommand(newBulkCloneStateShowCmd())
	cmd.AddCommand(newBulkCloneStateCleanCmd())

	return cmd
}

func newBulkCloneStateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved clone states",
		RunE:  runStateList,
	}

	return cmd
}

func newBulkCloneStateShowCmd() *cobra.Command {
	var provider, organization string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show details of a specific clone state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateShow(cmd, provider, organization)
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Provider (github, gitlab)")
	cmd.Flags().StringVarP(&organization, "organization", "o", "", "Organization/group name")
	if err := cmd.MarkFlagRequired("provider"); err != nil {
		// Error marking flag required is unlikely in practice
		panic(err)
	}
	if err := cmd.MarkFlagRequired("organization"); err != nil {
		// Error marking flag required is unlikely in practice
		panic(err)
	}

	return cmd
}

func newBulkCloneStateCleanCmd() *cobra.Command {
	var (
		provider, organization string
		all                    bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up saved clone states",
		Long:  `Clean up saved clone states. Use --all to clean all states, or specify --provider and --organization to clean a specific state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateClean(cmd, provider, organization, all)
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Provider (github, gitlab)")
	cmd.Flags().StringVarP(&organization, "organization", "o", "", "Organization/group name")
	cmd.Flags().BoolVar(&all, "all", false, "Clean all saved states")

	// Either --all or both --provider and --organization must be specified
	cmd.MarkFlagsOneRequired("all", "provider")
	cmd.MarkFlagsRequiredTogether("provider", "organization")

	return cmd
}

func stateManagerFromCommand(cmd *cobra.Command) (*synclonepkg.StateManager, error) {
	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return nil, fmt.Errorf("failed to read --target: %w", err)
	}
	if target == "" {
		return synclonepkg.NewStateManager(""), nil
	}

	return synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state")), nil
}

func runStateList(cmd *cobra.Command, args []string) error {
	stateManager, err := stateManagerFromCommand(cmd)
	if err != nil {
		return err
	}

	states, err := stateManager.ListStates()
	if err != nil {
		return fmt.Errorf("failed to list states: %w", err)
	}

	if len(states) == 0 {
		fmt.Println("No saved clone states found")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "PROVIDER\tORGANIZATION\tSTATUS\tPROGRESS\tLAST UPDATED\tTARGET PATH"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "--------\t------------\t------\t--------\t------------\t-----------"); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	for _, state := range states {
		progress := fmt.Sprintf("%.1f%%", state.GetProgressPercent())
		lastUpdated := state.LastUpdated.Format("2006-01-02 15:04:05")

		if _, err := fmt.Fprintf(
			w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			state.Provider,
			state.Organization,
			state.Status,
			progress,
			lastUpdated,
			state.TargetPath,
		); err != nil {
			return fmt.Errorf("failed to write state row: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	return nil
}

func runStateShow(cmd *cobra.Command, provider, organization string) error {
	stateManager, err := stateManagerFromCommand(cmd)
	if err != nil {
		return err
	}

	state, err := stateManager.LoadState(provider, organization)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Display state details
	fmt.Printf("Clone State Details\n")
	fmt.Printf("===================\n\n")
	fmt.Printf("Provider:      %s\n", state.Provider)
	fmt.Printf("Organization:  %s\n", state.Organization)
	fmt.Printf("Status:        %s\n", state.Status)
	fmt.Printf("Target Path:   %s\n", state.TargetPath)
	fmt.Printf("Strategy:      %s\n", state.Strategy)
	fmt.Printf("Started:       %s\n", state.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Last Updated:  %s\n", state.LastUpdated.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration:      %s\n", state.LastUpdated.Sub(state.StartTime).Round(time.Second))

	// Progress information
	completed, failed, pending := state.GetProgress()

	fmt.Printf("\nProgress\n")
	fmt.Printf("--------\n")
	fmt.Printf("Total Repositories: %d\n", state.TotalRepositories)
	fmt.Printf("Completed:          %d\n", completed)
	fmt.Printf("Failed:             %d\n", failed)
	fmt.Printf("Pending:            %d\n", pending)
	fmt.Printf("Progress:           %.1f%%\n", state.GetProgressPercent())

	// Configuration
	fmt.Printf("\nConfiguration\n")
	fmt.Printf("-------------\n")
	fmt.Printf("Parallel Workers: %d\n", state.Parallel)
	fmt.Printf("Max Retries:      %d\n", state.MaxRetries)

	// Failed repositories
	if len(state.FailedRepos) > 0 {
		fmt.Printf("\nFailed Repositories\n")
		fmt.Printf("-------------------\n")

		for _, failed := range state.FailedRepos {
			fmt.Printf("• %s: %s (attempts: %d)\n", failed.Name, failed.Error, failed.Attempts)
		}
	}

	// Pending repositories (show first 10)
	if len(state.PendingRepos) > 0 {
		fmt.Printf("\nPending Repositories\n")
		fmt.Printf("--------------------\n")

		limit := min(len(state.PendingRepos), 10)

		for i := range limit {
			fmt.Printf("• %s\n", state.PendingRepos[i])
		}

		if len(state.PendingRepos) > 10 {
			fmt.Printf("... and %d more\n", len(state.PendingRepos)-10)
		}
	}

	return nil
}

func runStateClean(cmd *cobra.Command, provider, organization string, all bool) error {
	stateManager, err := stateManagerFromCommand(cmd)
	if err != nil {
		return err
	}

	if all {
		// 목록 검증이 실패하면 이름도 찍지 않고 파일도 지우지 않는다.
		states, err := stateManager.ListStates()
		if err != nil {
			return fmt.Errorf("failed to verify states: %w", err)
		}
		if len(states) == 0 {
			fmt.Println("No saved states to clean")
			return nil
		}

		removed, err := stateManager.DeleteAllStates()
		if err != nil {
			return fmt.Errorf("failed to clean states after deleting %d state(s): %w", removed, err)
		}
		if removed != len(states) {
			return fmt.Errorf("verified %d state(s) but deleted %d", len(states), removed)
		}
		for _, state := range states {
			fmt.Printf("Cleaned %s/%s\n", state.Provider, state.Organization)
		}
		fmt.Printf("Cleaned %d state(s)\n", removed)
	} else {
		// Clean specific state
		if !stateManager.HasState(provider, organization) {
			fmt.Printf("No saved state found for %s/%s\n", provider, organization)
			return nil
		}

		if err := stateManager.DeleteState(provider, organization); err != nil {
			return fmt.Errorf("failed to delete state: %w", err)
		}

		fmt.Printf("Deleted state for %s/%s\n", provider, organization)
	}

	return nil
}
