// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package dashboard

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// GlobalFlags represents global flags for all repo-config commands.
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// addGlobalFlags adds common flags to the command.
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVar(&flags.Organization, "org", "", "GitHub organization name")
	cmd.Flags().StringVar(&flags.ConfigFile, "config-file", "", "Configuration file path")
	cmd.Flags().StringVar(&flags.Token, "token", "", "GitHub token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "Request timeout")
}

// NewCmd creates the dashboard subcommand.
func NewCmd() *cobra.Command {
	var (
		flags       GlobalFlags
		port        int
		autoRefresh bool
		refreshRate int
	)

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Repository compliance dashboard (unsupported)",
		Long: `The repository compliance dashboard is not implemented.
Execution returns an unsupported error without starting a web server.
Existing flags are retained for compatibility.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboardCommand(flags, port, autoRefresh, refreshRate)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add dashboard-specific flags
	cmd.Flags().IntVar(&port, "port", 8080, "Dashboard server port")
	cmd.Flags().BoolVar(&autoRefresh, "auto-refresh", false, "Enable automatic refresh")
	cmd.Flags().IntVar(&refreshRate, "refresh-rate", 60, "Auto refresh rate in seconds")

	return cmd
}

// runDashboardCommand rejects the unimplemented dashboard before any side effects.
func runDashboardCommand(_ GlobalFlags, _ int, _ bool, _ int) error {
	return fmt.Errorf("repository compliance dashboard is unsupported: no dashboard backend is implemented")
}

// handleDashboardHome reports the unsupported dashboard without fabricated HTML.
func handleDashboardHome(w http.ResponseWriter, _ *http.Request, _ string, _ bool, _ int) {
	http.Error(w, "repository compliance dashboard is unsupported", http.StatusNotImplemented)
}

// handleRepositoriesAPI reports the unsupported dashboard repository endpoint.
func handleRepositoriesAPI(w http.ResponseWriter, _ *http.Request, _, _ string) {
	http.Error(w, "dashboard repository reports are unsupported", http.StatusNotImplemented)
}

// handleComplianceAPI reports the unsupported dashboard compliance endpoint.
func handleComplianceAPI(w http.ResponseWriter, _ *http.Request, _, _ string) {
	http.Error(w, "dashboard compliance reports are unsupported", http.StatusNotImplemented)
}

// handleStaticAssets has no assets to serve.
func handleStaticAssets(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
