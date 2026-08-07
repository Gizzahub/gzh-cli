// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

// NewEventCmd creates a new event command.
func NewEventCmd() *cobra.Command {
	// Command flags - declare as local variables
	var (
		eventServerPort   int
		eventServerSecret string
		eventServerHost   string
		eventFilterOrg    string
		eventFilterRepo   string
		eventFilterType   string
		eventFilterAction string
		eventFilterSender string
		eventFilterSince  string
		eventFilterUntil  string
		eventListLimit    int
		eventListOffset   int
		eventOutputFormat string
		eventTestType     string
		eventTestAction   string
		eventTestPayload  string
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEventServer(cmd, args, eventServerHost, eventServerPort, eventServerSecret)
		},
	}

	eventListCmd := &cobra.Command{
		Use:   "list",
		Short: "List stored GitHub events",
		Long: `List GitHub events stored in the system with optional filtering.

Supports filtering by organization, repository, event type, action, sender,
and time range to help find specific events.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEventList(cmd, args, eventFilterOrg, eventFilterRepo, eventFilterType,
				eventFilterAction, eventFilterSender, eventFilterSince, eventFilterUntil,
				eventListLimit, eventListOffset, eventOutputFormat)
		},
	}

	eventGetCmd := &cobra.Command{
		Use:   "get [event-id]",
		Short: "Get details of a specific event",
		Long: `Retrieve detailed information about a specific GitHub event by its ID.

Shows the complete event payload, headers, processing status, and any
associated handler results.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEventGet(cmd, args, eventOutputFormat)
		},
	}

	eventMetricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show event processing metrics",
		Long: `Display comprehensive metrics about event processing including:
- Total events received and processed
- Events by type and organization
- Average processing time
- Handler status and performance`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEventMetrics(cmd, args, eventOutputFormat)
		},
	}

	eventTestCmd := &cobra.Command{
		Use:   "test",
		Short: "Test webhook endpoint",
		Long: `Send a test webhook event to verify server functionality.

Useful for testing webhook configuration and event processing
without waiting for actual GitHub events.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEventTest(cmd, args, eventTestType, eventTestAction, eventTestPayload, eventServerPort)
		},
	}

	eventCmd.AddCommand(eventServerCmd)
	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventGetCmd)
	eventCmd.AddCommand(eventMetricsCmd)
	eventCmd.AddCommand(eventTestCmd)

	// Server command flags
	eventServerCmd.Flags().IntVarP(&eventServerPort, "port", "p", 8080, "Port to listen on")
	eventServerCmd.Flags().StringVarP(&eventServerSecret, "secret", "s", "", "Webhook secret for signature validation")
	eventServerCmd.Flags().StringVar(&eventServerHost, "host", "0.0.0.0", "Host to bind to")

	// List command flags
	eventListCmd.Flags().StringVar(&eventFilterOrg, "org", "", "Filter by organization")
	eventListCmd.Flags().StringVar(&eventFilterRepo, "repo", "", "Filter by repository")
	eventListCmd.Flags().StringVar(&eventFilterType, "type", "", "Filter by event type")
	eventListCmd.Flags().StringVar(&eventFilterAction, "action", "", "Filter by event action")
	eventListCmd.Flags().StringVar(&eventFilterSender, "sender", "", "Filter by sender")
	eventListCmd.Flags().StringVar(&eventFilterSince, "since", "", "Filter events since (RFC3339 format)")
	eventListCmd.Flags().StringVar(&eventFilterUntil, "until", "", "Filter events until (RFC3339 format)")
	eventListCmd.Flags().IntVar(&eventListLimit, "limit", 50, "Maximum number of events to return")
	eventListCmd.Flags().IntVar(&eventListOffset, "offset", 0, "Number of events to skip")

	// Output format flags
	eventListCmd.Flags().StringVarP(&eventOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	eventGetCmd.Flags().StringVarP(&eventOutputFormat, "output", "o", "json", "Output format (json, yaml)")
	eventMetricsCmd.Flags().StringVarP(&eventOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")

	// Test command flags
	eventTestCmd.Flags().StringVar(&eventTestType, "type", "push", "Event type to test")
	eventTestCmd.Flags().StringVar(&eventTestAction, "action", "created", "Event action to test")
	eventTestCmd.Flags().StringVar(&eventTestPayload, "payload", "", "JSON payload file to send")

	return eventCmd
}

// errEventStorageNotImplemented is returned by list/get/metrics until real storage exists.
var errEventStorageNotImplemented = fmt.Errorf("event storage is not implemented")

func runEventServer(_ *cobra.Command, _ []string, host string, port int, secret string) error {
	logger := getLogger()
	logger.Info("Starting GitHub webhook server", "host", host, "port", port)
	logger.Warn("event storage is a no-op: received events are not persisted; list/get/metrics will fail")

	// Storage is intentionally a no-op: events are processed but not retained.
	storage := &noopEventStorage{}

	// Create event processor
	processor := github.NewEventProcessor(storage, logger)

	// Create webhook server
	server := github.NewEventWebhookServer(processor, secret, logger)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", server.HandleWebhook)
	mux.HandleFunc("/health", server.GetHealthCheck)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Processor metrics cover in-process handling only (not durable storage).
		metrics := processor.GetMetrics()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		}
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	logger.Info("Webhook server started", "address", addr)
	fmt.Printf("GitHub webhook server listening on %s\n", addr)
	fmt.Printf("Webhook endpoint: http://%s/webhook\n", addr)
	fmt.Printf("Health check: http://%s/health\n", addr)
	fmt.Printf("Metrics: http://%s/metrics\n", addr)
	fmt.Printf("Warning: event storage is not implemented; events are not persisted\n")

	return srv.ListenAndServe()
}

func runEventList(_ *cobra.Command, _ []string, _, _, _, _, _, _, _ string, _, _ int, _ string) error {
	return errEventStorageNotImplemented
}

func runEventGet(_ *cobra.Command, _ []string, _ string) error {
	return errEventStorageNotImplemented
}

func runEventMetrics(_ *cobra.Command, _ []string, _ string) error {
	return errEventStorageNotImplemented
}

func runEventTest(cmd *cobra.Command, _ []string, eventType, action, payload string, port int) error {
	logger := getLogger()

	// Default test payload
	testPayload := map[string]any{
		"action": action,
		"repository": map[string]any{
			"name": "test-repo",
			"owner": map[string]any{
				"login": "test-org",
			},
		},
		"sender": map[string]any{
			"login": "test-user",
		},
	}

	// Load custom payload if specified
	if payload != "" {
		// Validate payload file path to prevent directory traversal
		if !filepath.IsAbs(payload) {
			return fmt.Errorf("payload file path must be absolute: %s", payload)
		}
		file, err := os.Open(filepath.Clean(payload))
		if err != nil {
			return fmt.Errorf("failed to open payload file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				// Log error but don't override main error
				fmt.Printf("Warning: failed to close file: %v\n", err)
			}
		}()

		if err := json.NewDecoder(file).Decode(&testPayload); err != nil {
			return fmt.Errorf("failed to parse payload JSON: %w", err)
		}
	}

	// Send test webhook
	webhookURL := fmt.Sprintf("http://localhost:%d/webhook", port)

	logger.Info("Sending test webhook", "url", webhookURL, "type", eventType, "action", action)

	jsonPayload, err := json.Marshal(testPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(cmd.Context(), "POST", webhookURL, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", eventType)
	req.Header.Set("X-GitHub-Delivery", fmt.Sprintf("test-%d", time.Now().Unix()))

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send test webhook: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't override main error
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}()

	fmt.Printf("Test webhook sent successfully\n")
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Event Type: %s\n", eventType)
	fmt.Printf("Action: %s\n", action)

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Webhook processed successfully\n")
	} else {
		fmt.Printf("❌ Webhook processing failed\n")
	}

	return nil
}

// getLogger returns a logger that implements the github.Logger interface.
func getLogger() github.Logger {
	return &simpleLogger{}
}

// simpleLogger implements the github.Logger interface.
type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, args ...any) {
	log.Printf("[DEBUG] %s", formatMessage(msg, args...))
}

func (l *simpleLogger) Info(msg string, args ...any) {
	log.Printf("[INFO] %s", formatMessage(msg, args...))
}

func (l *simpleLogger) Warn(msg string, args ...any) {
	log.Printf("[WARN] %s", formatMessage(msg, args...))
}

func (l *simpleLogger) Error(msg string, args ...any) {
	log.Printf("[ERROR] %s", formatMessage(msg, args...))
}

// formatMessage formats a message with key-value pairs.
func formatMessage(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}

	return fmt.Sprintf("%s %v", msg, args)
}

// noopEventStorage accepts events but does not persist them.
// list/get/metrics CLI commands fail-fast until a real backend exists.
type noopEventStorage struct{}

func (m *noopEventStorage) StoreEvent(_ context.Context, _ *github.GitHubEvent) error {
	return nil
}

func (m *noopEventStorage) GetEvent(_ context.Context, eventID string) (*github.GitHubEvent, error) {
	return nil, fmt.Errorf("%w: %s", errEventStorageNotImplemented, eventID)
}

func (m *noopEventStorage) ListEvents(_ context.Context, _ *github.EventFilter, _, _ int) ([]*github.GitHubEvent, error) {
	return nil, errEventStorageNotImplemented
}

func (m *noopEventStorage) DeleteEvent(_ context.Context, _ string) error {
	return errEventStorageNotImplemented
}

func (m *noopEventStorage) CountEvents(_ context.Context, _ *github.EventFilter) (int, error) {
	return 0, errEventStorageNotImplemented
}
