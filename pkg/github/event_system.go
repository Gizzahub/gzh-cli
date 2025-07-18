package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubEvent represents a GitHub webhook event.
type GitHubEvent struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Action       string                 `json:"action,omitempty"`
	Organization string                 `json:"organization"`
	Repository   string                 `json:"repository"`
	Sender       string                 `json:"sender"`
	Timestamp    time.Time              `json:"timestamp"`
	Payload      map[string]interface{} `json:"payload"`
	Headers      map[string]string      `json:"headers"`
	Signature    string                 `json:"signature"`
}

// EventType defines the type of GitHub events.
type EventType string

const (
	EventTypePush              EventType = "push"
	EventTypePullRequest       EventType = "pull_request"
	EventTypeIssues            EventType = "issues"
	EventTypeRepository        EventType = "repository"
	EventTypeRelease           EventType = "release"
	EventTypeCreate            EventType = "create"
	EventTypeDelete            EventType = "delete"
	EventTypeWorkflowRun       EventType = "workflow_run"
	EventTypeDeployment        EventType = "deployment"
	EventTypeMember            EventType = "member"
	EventTypeTeam              EventType = "team"
	EventTypeOrganization      EventType = "organization"
	EventTypeInstallation      EventType = "installation"
	EventTypeInstallationRepos EventType = "installation_repositories"
)

// EventAction defines specific actions within events.
type EventAction string

const (
	ActionOpened      EventAction = "opened"
	ActionClosed      EventAction = "closed"
	ActionSynchronize EventAction = "synchronize"
	ActionCreated     EventAction = "created"
	ActionDeleted     EventAction = "deleted"
	ActionEdited      EventAction = "edited"
	ActionCompleted   EventAction = "completed"
	ActionRequested   EventAction = "requested"
	ActionSubmitted   EventAction = "submitted"
	ActionPublished   EventAction = "published"
	ActionAdded       EventAction = "added"
	ActionRemoved     EventAction = "removed"
)

// EventProcessor defines the interface for processing GitHub events.
type EventProcessor interface {
	ProcessEvent(ctx context.Context, event *GitHubEvent) error
	ValidateSignature(payload []byte, signature, secret string) bool
	ParseWebhookEvent(r *http.Request) (*GitHubEvent, error)
	RegisterEventHandler(eventType EventType, handler EventHandler) error
	UnregisterEventHandler(eventType EventType) error
	GetMetrics() *EventMetrics
	ValidateEvent(ctx context.Context, event *GitHubEvent) error
	FilterEvent(ctx context.Context, event *GitHubEvent, filter *EventFilter) (bool, error)
}

// EventHandler defines the interface for handling specific event types.
type EventHandler interface {
	HandleEvent(ctx context.Context, event *GitHubEvent) error
	GetSupportedActions() []EventAction
	GetPriority() int // Higher number = higher priority
}

// EventFilter defines criteria for filtering events.
type EventFilter struct {
	Organization  string        `json:"organization,omitempty"`
	Repository    string        `json:"repository,omitempty"`
	EventTypes    []EventType   `json:"event_types,omitempty"`
	Actions       []EventAction `json:"actions,omitempty"`
	Sender        string        `json:"sender,omitempty"`
	BranchPattern string        `json:"branch_pattern,omitempty"`
	FilePattern   string        `json:"file_pattern,omitempty"`
	TimeRange     *TimeRange    `json:"time_range,omitempty"`
}

// TimeRange defines a time range for event filtering.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// EventProcessingResult represents the result of event processing.
type EventProcessingResult struct {
	EventID     string    `json:"event_id"`
	Success     bool      `json:"success"`
	HandlerName string    `json:"handler_name"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
	Duration    string    `json:"duration"`
	Actions     []string  `json:"actions,omitempty"`
}

// EventStorage defines the interface for storing events.
type EventStorage interface {
	StoreEvent(ctx context.Context, event *GitHubEvent) error
	GetEvent(ctx context.Context, eventID string) (*GitHubEvent, error)
	ListEvents(ctx context.Context, filter *EventFilter, limit, offset int) ([]*GitHubEvent, error)
	DeleteEvent(ctx context.Context, eventID string) error
	CountEvents(ctx context.Context, filter *EventFilter) (int, error)
}

// EventMetrics provides metrics for event processing.
type EventMetrics struct {
	TotalEventsReceived   int64             `json:"total_events_received"`
	TotalEventsProcessed  int64             `json:"total_events_processed"`
	TotalEventsFailed     int64             `json:"total_events_failed"`
	EventsByType          map[string]int64  `json:"events_by_type"`
	EventsByOrganization  map[string]int64  `json:"events_by_organization"`
	AverageProcessingTime time.Duration     `json:"average_processing_time"`
	LastEventAt           time.Time         `json:"last_event_at"`
	HandlersStatus        map[string]string `json:"handlers_status"`
}

// eventProcessorImpl implements the EventProcessor interface.
type eventProcessorImpl struct {
	handlers map[EventType][]EventHandler
	storage  EventStorage
	logger   Logger
	metrics  *EventMetrics
}

// NewEventProcessor creates a new event processor.
func NewEventProcessor(storage EventStorage, logger Logger) EventProcessor {
	return &eventProcessorImpl{
		handlers: make(map[EventType][]EventHandler),
		storage:  storage,
		logger:   logger,
		metrics: &EventMetrics{
			EventsByType:         make(map[string]int64),
			EventsByOrganization: make(map[string]int64),
			HandlersStatus:       make(map[string]string),
		},
	}
}

// ProcessEvent processes a GitHub event.
func (e *eventProcessorImpl) ProcessEvent(ctx context.Context, event *GitHubEvent) error {
	e.logger.Info("Processing GitHub event", "event_id", event.ID, "type", event.Type, "action", event.Action)

	startTime := time.Now()

	defer func() {
		e.updateMetrics(event, time.Since(startTime))
	}()

	// Store the event
	if err := e.storage.StoreEvent(ctx, event); err != nil {
		e.logger.Error("Failed to store event", "event_id", event.ID, "error", err)
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Get handlers for this event type
	handlers, exists := e.handlers[EventType(event.Type)]
	if !exists || len(handlers) == 0 {
		e.logger.Debug("No handlers registered for event type", "type", event.Type)
		return nil
	}

	// Process event with all registered handlers
	var lastErr error

	handledCount := 0

	for _, handler := range handlers {
		// Check if handler supports this action
		if !e.handlerSupportsAction(handler, EventAction(event.Action)) {
			continue
		}

		if err := handler.HandleEvent(ctx, event); err != nil {
			e.logger.Error("Handler failed to process event",
				"handler", fmt.Sprintf("%T", handler),
				"event_id", event.ID,
				"error", err)
			lastErr = err
		} else {
			handledCount++
		}
	}

	if handledCount == 0 && lastErr != nil {
		return fmt.Errorf("all handlers failed to process event: %w", lastErr)
	}

	e.logger.Info("Event processed successfully",
		"event_id", event.ID,
		"handlers_count", handledCount)

	return nil
}

// ValidateSignature validates the GitHub webhook signature.
func (e *eventProcessorImpl) ValidateSignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		e.logger.Warn("No secret configured for signature validation")
		return true // Allow if no secret is configured
	}

	if signature == "" {
		e.logger.Warn("No signature provided in request")
		return false
	}

	// Remove 'sha256=' prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	valid := hmac.Equal([]byte(signature), []byte(expectedSignature))

	if !valid {
		e.logger.Warn("Invalid webhook signature",
			"provided", signature[:8]+"...",
			"expected", expectedSignature[:8]+"...")
	}

	return valid
}

// ParseWebhookEvent parses a GitHub webhook request into an event.
func (e *eventProcessorImpl) ParseWebhookEvent(r *http.Request) (*GitHubEvent, error) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Get event metadata from headers
	eventType := r.Header.Get("X-GitHub-Event")
	eventID := r.Header.Get("X-GitHub-Delivery")
	signature := r.Header.Get("X-Hub-Signature-256")

	if eventType == "" {
		return nil, fmt.Errorf("missing X-GitHub-Event header")
	}

	if eventID == "" {
		return nil, fmt.Errorf("missing X-GitHub-Delivery header")
	}

	// Parse the JSON payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JSON payload: %w", err)
	}

	// Extract common fields
	event := &GitHubEvent{
		ID:        eventID,
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
		Headers:   make(map[string]string),
		Signature: signature,
	}

	// Copy headers
	for name, values := range r.Header {
		if len(values) > 0 {
			event.Headers[name] = values[0]
		}
	}

	// Extract standard fields from payload
	e.extractStandardFields(event, payload)

	return event, nil
}

// RegisterEventHandler registers an event handler for a specific event type.
func (e *eventProcessorImpl) RegisterEventHandler(eventType EventType, handler EventHandler) error {
	e.logger.Info("Registering event handler", "event_type", eventType, "handler", fmt.Sprintf("%T", handler))

	if _, exists := e.handlers[eventType]; !exists {
		e.handlers[eventType] = make([]EventHandler, 0)
	}

	e.handlers[eventType] = append(e.handlers[eventType], handler)

	// Sort handlers by priority (highest first)
	e.sortHandlersByPriority(eventType)

	e.metrics.HandlersStatus[string(eventType)] = "active"

	return nil
}

// UnregisterEventHandler removes an event handler for a specific event type.
func (e *eventProcessorImpl) UnregisterEventHandler(eventType EventType) error {
	e.logger.Info("Unregistering event handlers", "event_type", eventType)

	delete(e.handlers, eventType)
	delete(e.metrics.HandlersStatus, string(eventType))

	return nil
}

// Helper methods

func (e *eventProcessorImpl) extractStandardFields(event *GitHubEvent, payload map[string]interface{}) {
	// Extract action
	if action, ok := payload["action"].(string); ok {
		event.Action = action
	}

	// Extract organization
	if org, ok := payload["organization"].(map[string]interface{}); ok {
		if orgLogin, ok := org["login"].(string); ok {
			event.Organization = orgLogin
		}
	}

	// Extract repository
	if repo, ok := payload["repository"].(map[string]interface{}); ok {
		if repoName, ok := repo["name"].(string); ok {
			event.Repository = repoName
		}
		// If organization not found in organization field, try repository owner
		if event.Organization == "" {
			if owner, ok := repo["owner"].(map[string]interface{}); ok {
				if ownerLogin, ok := owner["login"].(string); ok {
					event.Organization = ownerLogin
				}
			}
		}
	}

	// Extract sender
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if senderLogin, ok := sender["login"].(string); ok {
			event.Sender = senderLogin
		}
	}
}

func (e *eventProcessorImpl) handlerSupportsAction(handler EventHandler, action EventAction) bool {
	supportedActions := handler.GetSupportedActions()
	if len(supportedActions) == 0 {
		return true // Handler supports all actions
	}

	for _, supportedAction := range supportedActions {
		if supportedAction == action {
			return true
		}
	}

	return false
}

func (e *eventProcessorImpl) sortHandlersByPriority(eventType EventType) {
	handlers := e.handlers[eventType]

	// Simple bubble sort by priority (highest first)
	for i := 0; i < len(handlers)-1; i++ {
		for j := 0; j < len(handlers)-i-1; j++ {
			if handlers[j].GetPriority() < handlers[j+1].GetPriority() {
				handlers[j], handlers[j+1] = handlers[j+1], handlers[j]
			}
		}
	}
}

func (e *eventProcessorImpl) updateMetrics(event *GitHubEvent, duration time.Duration) {
	e.metrics.TotalEventsReceived++
	e.metrics.TotalEventsProcessed++
	e.metrics.EventsByType[event.Type]++
	e.metrics.EventsByOrganization[event.Organization]++
	e.metrics.LastEventAt = event.Timestamp

	// Update average processing time
	if e.metrics.AverageProcessingTime == 0 {
		e.metrics.AverageProcessingTime = duration
	} else {
		e.metrics.AverageProcessingTime = (e.metrics.AverageProcessingTime + duration) / 2
	}
}

// GetMetrics returns current event processing metrics.
func (e *eventProcessorImpl) GetMetrics() *EventMetrics {
	return e.metrics
}

// ValidateEvent validates a GitHub event.
func (e *eventProcessorImpl) ValidateEvent(ctx context.Context, event *GitHubEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	if event.ID == "" {
		return fmt.Errorf("event ID is empty")
	}

	if event.Type == "" {
		return fmt.Errorf("event type is empty")
	}

	return nil
}

// FilterEvent filters a GitHub event based on the provided filter.
func (e *eventProcessorImpl) FilterEvent(ctx context.Context, event *GitHubEvent, filter *EventFilter) (bool, error) {
	if filter == nil {
		return true, nil
	}

	// Check organization filter
	if filter.Organization != "" && event.Organization != filter.Organization {
		return false, nil
	}

	// Check repository filter
	if filter.Repository != "" && event.Repository != filter.Repository {
		return false, nil
	}

	// Check event type filter
	if len(filter.EventTypes) > 0 {
		found := false

		for _, eventType := range filter.EventTypes {
			if string(eventType) == event.Type {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	// Check action filter
	if len(filter.Actions) > 0 && event.Action != "" {
		found := false

		for _, action := range filter.Actions {
			if string(action) == event.Action {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	// Check sender filter
	if filter.Sender != "" && event.Sender != filter.Sender {
		return false, nil
	}

	// Check time range filter
	if filter.TimeRange != nil {
		if event.Timestamp.Before(filter.TimeRange.Start) || event.Timestamp.After(filter.TimeRange.End) {
			return false, nil
		}
	}

	return true, nil
}

// EventWebhookServer provides HTTP server functionality for receiving GitHub webhooks.
type EventWebhookServer struct {
	processor EventProcessor
	secret    string
	logger    Logger
}

// NewEventWebhookServer creates a new webhook server.
func NewEventWebhookServer(processor EventProcessor, secret string, logger Logger) *EventWebhookServer {
	return &EventWebhookServer{
		processor: processor,
		secret:    secret,
		logger:    logger,
	}
}

// HandleWebhook handles incoming GitHub webhook requests.
func (s *EventWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the webhook event
	event, err := s.processor.ParseWebhookEvent(r)
	if err != nil {
		s.logger.Error("Failed to parse webhook event", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	// Validate signature if secret is configured
	if s.secret != "" {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(body))) // Reset body for parsing

		if !s.processor.ValidateSignature(body, event.Signature, s.secret) {
			s.logger.Warn("Invalid webhook signature", "event_id", event.ID)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}
	}

	// Process the event
	ctx := r.Context()
	if err := s.processor.ProcessEvent(ctx, event); err != nil {
		s.logger.Error("Failed to process event", "event_id", event.ID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":   "success",
		"event_id": event.ID,
		"message":  "Event processed successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// GetHealthCheck provides a health check endpoint.
func (s *EventWebhookServer) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "github-event-processor",
	}
	json.NewEncoder(w).Encode(response)
}
