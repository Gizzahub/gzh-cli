package cmd

import (
	"context"

	"github.com/gizzahub/gzh-manager-go/internal/event"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

// EventDependencies holds all dependencies for event commands.
type EventDependencies struct {
	Logger  github.Logger
	Storage github.EventStorage
}

// EventCommandFactory creates event commands with injected dependencies.
type EventCommandFactory struct {
	deps *EventDependencies
}

// NewEventCommandFactory creates a new event command factory.
func NewEventCommandFactory(deps *EventDependencies) *EventCommandFactory {
	// Provide defaults if not specified
	if deps.Logger == nil {
		deps.Logger = event.NewLoggerAdapter()
	}

	if deps.Storage == nil {
		deps.Storage = event.NewMockStorage()
	}

	return &EventCommandFactory{
		deps: deps,
	}
}

// NewEventCmd creates a new event command with dependency injection.
func (f *EventCommandFactory) NewEventCmd() *cobra.Command {
	// Command flags
	var (
		eventServerPort   int
		eventServerSecret string
		eventServerHost   string
	)

	eventCmd := &cobra.Command{
		Use:   "event",
		Short: "GitHub event management and webhook server",
		Long: `Manage GitHub events, run webhook servers, and monitor event processing.
	
This command provides comprehensive event management capabilities including:
- Running webhook servers to receive GitHub events
- Querying and filtering stored events
- Managing event handlers and processors
- Monitoring event processing metrics`,
	}

	eventServerCmd := &cobra.Command{
		Use:   "server",
		Short: "Start GitHub webhook server",
		Long: `Start a webhook server to receive and process GitHub events.
	
The server listens for incoming webhook requests from GitHub and processes them
according to registered event handlers and policies.`,
		RunE: func(_ *cobra.Command, args []string) error {
			// Use injected dependencies
			server := event.NewServer(
				eventServerHost,
				eventServerPort,
				eventServerSecret,
				f.deps.Storage,
				f.deps.Logger,
			)
			return server.Start(context.Background())
		},
	}

	// Add other commands with similar dependency injection...

	// Add flags to server command
	eventServerCmd.Flags().IntVarP(&eventServerPort, "port", "p", 8089, "Server port")
	eventServerCmd.Flags().StringVarP(&eventServerSecret, "secret", "s", "", "Webhook secret")
	eventServerCmd.Flags().StringVar(&eventServerHost, "host", "0.0.0.0", "Server host")

	// Add subcommands
	eventCmd.AddCommand(eventServerCmd)

	return eventCmd
}
