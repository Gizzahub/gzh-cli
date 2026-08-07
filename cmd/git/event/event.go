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
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

const (
	outputFormatJSON  = "json"
	outputFormatYAML  = "yaml"
	outputFormatTable = "table"
)

// defaultEventStorage is process-local storage shared by server/list/get/metrics
// within the same process. It is intentionally in-memory only.
var (
	defaultStorageMu sync.RWMutex
	defaultStorage   github.EventStorage = github.NewMemoryEventStorage()
)

// SetDefaultEventStorage replaces the package-level storage (tests / DI).
func SetDefaultEventStorage(storage github.EventStorage) {
	defaultStorageMu.Lock()
	defer defaultStorageMu.Unlock()
	if storage == nil {
		defaultStorage = github.NewMemoryEventStorage()
		return
	}
	defaultStorage = storage
}

// DefaultEventStorage returns the package-level event storage.
func DefaultEventStorage() github.EventStorage {
	defaultStorageMu.RLock()
	defer defaultStorageMu.RUnlock()
	return defaultStorage
}

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
- Monitoring event processing metrics

Note: event storage is process-local (in-memory). list/get/metrics only see
events received by a server running in the same process.`,
	}

	eventServerCmd := &cobra.Command{
		Use:   "server",
		Short: "Start GitHub webhook server",
		Long: `Start a webhook server to receive and process GitHub events.

The server listens for incoming webhook requests from GitHub and processes them
according to registered event handlers and policies. Received events are stored
in process-local memory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEventServer(cmd, args, eventServerHost, eventServerPort, eventServerSecret)
		},
	}

	eventListCmd := &cobra.Command{
		Use:   "list",
		Short: "List stored GitHub events",
		Long: `List GitHub events stored in the system with optional filtering.

Supports filtering by organization, repository, event type, action, sender,
and time range to help find specific events.

Events are process-local (in-memory); an empty result means nothing has been
stored in this process yet.`,
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
associated handler results. Returns a clear error when the event is not found.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEventGet(cmd, args, eventOutputFormat)
		},
	}

	eventMetricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show event processing metrics",
		Long: `Display metrics derived from process-local stored events including:
- Total events received and stored
- Events by type and organization
- Last event timestamp

Zeros mean no events are stored in this process (not fabricated sample data).`,
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

func runEventServer(_ *cobra.Command, _ []string, host string, port int, secret string) error {
	logger := getLogger()
	logger.Info("Starting GitHub webhook server", "host", host, "port", port)
	logger.Info("using process-local in-memory event storage")

	storage := DefaultEventStorage()
	processor := github.NewEventProcessor(storage, logger)
	server := github.NewEventWebhookServer(processor, secret, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", server.HandleWebhook)
	mux.HandleFunc("/health", server.GetHealthCheck)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Prefer storage-backed counts; fall back to processor timing metrics.
		var metrics any
		if mem, ok := storage.(*github.MemoryEventStorage); ok {
			metrics = mem.AggregateMetrics(r.Context())
		} else {
			metrics = processor.GetMetrics()
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		}
	})

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
	fmt.Printf("Note: events are stored in process-local memory only\n")

	return srv.ListenAndServe()
}

func runEventList(_ *cobra.Command, _ []string, org, repo, eventType, action, sender, since, until string, limit, offset int, outputFormat string) error {
	ctx := context.Background()
	storage := DefaultEventStorage()

	filter, err := buildEventFilter(org, repo, eventType, action, sender, since, until)
	if err != nil {
		return err
	}

	events, err := storage.ListEvents(ctx, filter, limit, offset)
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}

	if len(events) == 0 {
		fmt.Println("No events stored in process-local memory")
		fmt.Println("(start `gz git event server` in this process and receive webhooks first)")
		return nil
	}

	switch outputFormat {
	case outputFormatJSON:
		return outputJSON(events)
	case outputFormatYAML:
		return outputYAML(events)
	default:
		return outputEventTable(events)
	}
}

func runEventGet(_ *cobra.Command, args []string, outputFormat string) error {
	eventID := args[0]
	ctx := context.Background()
	storage := DefaultEventStorage()

	event, err := storage.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}

	switch outputFormat {
	case outputFormatYAML:
		return outputYAML(event)
	default:
		return outputJSON(event)
	}
}

func runEventMetrics(_ *cobra.Command, _ []string, outputFormat string) error {
	ctx := context.Background()
	storage := DefaultEventStorage()

	var metrics *github.EventMetrics
	if mem, ok := storage.(*github.MemoryEventStorage); ok {
		metrics = mem.AggregateMetrics(ctx)
	} else {
		// Best-effort for non-memory backends: count only.
		count, err := storage.CountEvents(ctx, nil)
		if err != nil {
			return fmt.Errorf("count events: %w", err)
		}
		metrics = &github.EventMetrics{
			TotalEventsReceived:  int64(count),
			TotalEventsProcessed: int64(count),
			EventsByType:         map[string]int64{},
			EventsByOrganization: map[string]int64{},
			HandlersStatus:       map[string]string{},
		}
	}

	switch outputFormat {
	case outputFormatJSON:
		return outputJSON(metrics)
	case outputFormatYAML:
		return outputYAML(metrics)
	default:
		return outputMetricsTable(metrics)
	}
}

func buildEventFilter(org, repo, eventType, action, sender, since, until string) (*github.EventFilter, error) {
	filter := &github.EventFilter{
		Organization: org,
		Repository:   repo,
		Sender:       sender,
	}

	if eventType != "" {
		filter.EventTypes = []github.EventType{github.EventType(eventType)}
	}
	if action != "" {
		filter.Actions = []github.EventAction{github.EventAction(action)}
	}

	var timeRange *github.TimeRange
	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			return nil, fmt.Errorf("invalid since time format: %w", err)
		}
		timeRange = &github.TimeRange{Start: t}
	}
	if until != "" {
		t, err := time.Parse(time.RFC3339, until)
		if err != nil {
			return nil, fmt.Errorf("invalid until time format: %w", err)
		}
		if timeRange == nil {
			timeRange = &github.TimeRange{}
		}
		timeRange.End = t
	}
	filter.TimeRange = timeRange

	// If no criteria set, return nil filter for "match all".
	if org == "" && repo == "" && eventType == "" && action == "" && sender == "" && timeRange == nil {
		return nil, nil
	}
	return filter, nil
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

// Output helper functions.
func outputEventTable(events []*github.GitHubEvent) error {
	fmt.Printf("%-20s %-15s %-12s %-15s %-15s %-20s\n",
		"EVENT ID", "TYPE", "ACTION", "ORGANIZATION", "REPOSITORY", "TIMESTAMP")
	fmt.Println(strings.Repeat("-", 100))

	for _, event := range events {
		timestamp := event.Timestamp.Format("2006-01-02 15:04:05")
		fmt.Printf("%-20s %-15s %-12s %-15s %-15s %-20s\n",
			truncate(event.ID, 20),
			truncate(event.Type, 15),
			truncate(event.Action, 12),
			truncate(event.Organization, 15),
			truncate(event.Repository, 15),
			timestamp)
	}

	fmt.Printf("\nTotal: %d events\n", len(events))
	return nil
}

func outputMetricsTable(metrics *github.EventMetrics) error {
	fmt.Println("GitHub Event Processing Metrics (process-local storage)")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total Events Received:  %d\n", metrics.TotalEventsReceived)
	fmt.Printf("Total Events Processed: %d\n", metrics.TotalEventsProcessed)
	fmt.Printf("Total Events Failed:    %d\n", metrics.TotalEventsFailed)
	if metrics.AverageProcessingTime > 0 {
		fmt.Printf("Average Processing Time: %v\n", metrics.AverageProcessingTime)
	}
	if !metrics.LastEventAt.IsZero() {
		fmt.Printf("Last Event At:          %s\n", metrics.LastEventAt.Format(time.RFC3339))
	} else {
		fmt.Printf("Last Event At:          (none)\n")
	}

	fmt.Println("\nEvents by Type:")
	fmt.Println(strings.Repeat("-", 30))
	if len(metrics.EventsByType) == 0 {
		fmt.Println("  (none)")
	}
	for eventType, count := range metrics.EventsByType {
		fmt.Printf("  %-15s %d\n", eventType, count)
	}

	fmt.Println("\nEvents by Organization:")
	fmt.Println(strings.Repeat("-", 30))
	if len(metrics.EventsByOrganization) == 0 {
		fmt.Println("  (none)")
	}
	for org, count := range metrics.EventsByOrganization {
		fmt.Printf("  %-15s %d\n", org, count)
	}

	if len(metrics.HandlersStatus) > 0 {
		fmt.Println("\nHandler Status:")
		fmt.Println(strings.Repeat("-", 30))
		for handler, status := range metrics.HandlersStatus {
			fmt.Printf("  %-15s %s\n", handler, status)
		}
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
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

// outputJSON marshals the data to JSON and prints it.
func outputJSON(data any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputYAML marshals the data to YAML and prints it.
func outputYAML(data any) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	fmt.Println(string(yamlData))
	return nil
}
