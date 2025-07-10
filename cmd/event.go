package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

// NewEventCmd creates a new event command
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

func runEventServer(cmd *cobra.Command, args []string, host string, port int, secret string) error {
	ctx := context.Background()

	logger := getLogger()
	logger.Info("Starting GitHub webhook server", "host", host, "port", port)

	// Create storage implementation (would be real implementation)
	storage := &mockEventStorage{}

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

		metrics := processor.GetMetrics()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	logger.Info("Webhook server started", "address", addr)
	fmt.Printf("GitHub webhook server listening on %s\n", addr)
	fmt.Printf("Webhook endpoint: http://%s/webhook\n", addr)
	fmt.Printf("Health check: http://%s/health\n", addr)
	fmt.Printf("Metrics: http://%s/metrics\n", addr)

	return srv.ListenAndServe()
}

func runEventList(cmd *cobra.Command, args []string, org, repo, eventType, action, sender, since, until string, limit, offset int, outputFormat string) error {
	ctx := context.Background()
	logger := getLogger()

	// Create storage implementation
	storage := &mockEventStorage{}

	// Build event filter
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

	// Parse time filters
	if since != "" || until != "" {
		timeRange := &github.TimeRange{}
		if since != "" {
			sinceTime, err := time.Parse(time.RFC3339, since)
			if err != nil {
				return fmt.Errorf("invalid since time format: %w", err)
			}
			timeRange.Start = sinceTime
		}
		if until != "" {
			untilTime, err := time.Parse(time.RFC3339, until)
			if err != nil {
				return fmt.Errorf("invalid until time format: %w", err)
			}
			timeRange.End = untilTime
		}
		filter.TimeRange = timeRange
	}

	// Mock events for demonstration
	events := []*github.GitHubEvent{
		{
			ID:           "event-1",
			Type:         "push",
			Action:       "created",
			Organization: "testorg",
			Repository:   "testrepo",
			Sender:       "user1",
			Timestamp:    time.Now().Add(-1 * time.Hour),
		},
		{
			ID:           "event-2",
			Type:         "pull_request",
			Action:       "opened",
			Organization: "testorg",
			Repository:   "testrepo",
			Sender:       "user2",
			Timestamp:    time.Now().Add(-30 * time.Minute),
		},
	}

	// Output events
	switch outputFormat {
	case "json":
		return outputJSON(events)
	case "yaml":
		return outputYAML(events)
	default:
		return outputEventTable(events)
	}
}

func runEventGet(cmd *cobra.Command, args []string, outputFormat string) error {
	eventID := args[0]
	ctx := context.Background()
	logger := getLogger()

	// Create storage implementation
	storage := &mockEventStorage{}

	// Mock event for demonstration
	event := &github.GitHubEvent{
		ID:           eventID,
		Type:         "push",
		Action:       "created",
		Organization: "testorg",
		Repository:   "testrepo",
		Sender:       "testuser",
		Timestamp:    time.Now(),
		Payload: map[string]interface{}{
			"ref": "refs/heads/main",
			"commits": []interface{}{
				map[string]interface{}{
					"id":      "abc123",
					"message": "Test commit",
					"author": map[string]interface{}{
						"name":  "Test User",
						"email": "test@example.com",
					},
				},
			},
		},
		Headers: map[string]string{
			"X-GitHub-Event":    "push",
			"X-GitHub-Delivery": eventID,
		},
		Signature: "sha256=example-signature",
	}

	switch outputFormat {
	case "yaml":
		return outputYAML(event)
	default:
		return outputJSON(event)
	}
}

func runEventMetrics(cmd *cobra.Command, args []string, outputFormat string) error {
	ctx := context.Background()
	logger := getLogger()

	// Create storage implementation
	storage := &mockEventStorage{}
	processor := github.NewEventProcessor(storage, logger)

	// Mock metrics for demonstration
	metrics := &github.EventMetrics{
		TotalEventsReceived:  1250,
		TotalEventsProcessed: 1248,
		TotalEventsFailed:    2,
		EventsByType: map[string]int64{
			"push":         450,
			"pull_request": 320,
			"issues":       200,
			"release":      150,
			"workflow_run": 130,
		},
		EventsByOrganization: map[string]int64{
			"org1": 600,
			"org2": 400,
			"org3": 250,
		},
		AverageProcessingTime: 125 * time.Millisecond,
		LastEventAt:           time.Now().Add(-5 * time.Minute),
		HandlersStatus: map[string]string{
			"push":         "active",
			"pull_request": "active",
			"issues":       "active",
		},
	}

	switch outputFormat {
	case "json":
		return outputJSON(metrics)
	case "yaml":
		return outputYAML(metrics)
	default:
		return outputMetricsTable(metrics)
	}
}

func runEventTest(cmd *cobra.Command, args []string, eventType, action, payload string, port int) error {
	logger := getLogger()

	// Default test payload
	testPayload := map[string]interface{}{
		"action": action,
		"repository": map[string]interface{}{
			"name": "test-repo",
			"owner": map[string]interface{}{
				"login": "test-org",
			},
		},
		"sender": map[string]interface{}{
			"login": "test-user",
		},
	}

	// Load custom payload if specified
	if payload != "" {
		file, err := os.Open(payload)
		if err != nil {
			return fmt.Errorf("failed to open payload file: %w", err)
		}
		defer file.Close()

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

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(jsonPayload))
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
	defer resp.Body.Close()

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

// Output helper functions
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
	fmt.Println("GitHub Event Processing Metrics")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total Events Received:  %d\n", metrics.TotalEventsReceived)
	fmt.Printf("Total Events Processed: %d\n", metrics.TotalEventsProcessed)
	fmt.Printf("Total Events Failed:    %d\n", metrics.TotalEventsFailed)
	fmt.Printf("Average Processing Time: %v\n", metrics.AverageProcessingTime)
	fmt.Printf("Last Event At:          %s\n", metrics.LastEventAt.Format(time.RFC3339))

	fmt.Println("\nEvents by Type:")
	fmt.Println(strings.Repeat("-", 30))
	for eventType, count := range metrics.EventsByType {
		fmt.Printf("  %-15s %d\n", eventType, count)
	}

	fmt.Println("\nEvents by Organization:")
	fmt.Println(strings.Repeat("-", 30))
	for org, count := range metrics.EventsByOrganization {
		fmt.Printf("  %-15s %d\n", org, count)
	}

	fmt.Println("\nHandler Status:")
	fmt.Println(strings.Repeat("-", 30))
	for handler, status := range metrics.HandlersStatus {
		fmt.Printf("  %-15s %s\n", handler, status)
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Mock storage implementation for CLI demonstration
type mockEventStorage struct{}

func (m *mockEventStorage) StoreEvent(ctx context.Context, event *github.GitHubEvent) error {
	return nil
}

func (m *mockEventStorage) GetEvent(ctx context.Context, eventID string) (*github.GitHubEvent, error) {
	return nil, nil
}

func (m *mockEventStorage) ListEvents(ctx context.Context, filter *github.EventFilter, limit, offset int) ([]*github.GitHubEvent, error) {
	return []*github.GitHubEvent{}, nil
}

func (m *mockEventStorage) DeleteEvent(ctx context.Context, eventID string) error {
	return nil
}

func (m *mockEventStorage) CountEvents(ctx context.Context, filter *github.EventFilter) (int, error) {
	return 0, nil
}
