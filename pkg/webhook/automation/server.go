package automation

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

	"github.com/google/go-github/v66/github"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Server handles incoming webhook events
type Server struct {
	engine       *Engine
	webhookPath  string
	secret       string
	logger       *zap.Logger
	httpServer   *http.Server
	eventChannel chan *Event
	workers      int
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Port         int
	WebhookPath  string
	Secret       string
	Workers      int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewServer creates a new webhook server
func NewServer(engine *Engine, config ServerConfig, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config.Workers == 0 {
		config.Workers = 10
	}

	if config.WebhookPath == "" {
		config.WebhookPath = "/webhook"
	}

	if config.ReadTimeout == 0 {
		config.ReadTimeout = 10 * time.Second
	}

	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}

	s := &Server{
		engine:       engine,
		webhookPath:  config.WebhookPath,
		secret:       config.Secret,
		logger:       logger,
		eventChannel: make(chan *Event, 100),
		workers:      config.Workers,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(config.WebhookPath, s.handleWebhook)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return s
}

// Start starts the webhook server
func (s *Server) Start(ctx context.Context) error {
	// Start event processing workers
	for i := 0; i < s.workers; i++ {
		go s.eventWorker(ctx, i)
	}

	s.logger.Info("Starting webhook server",
		zap.String("address", s.httpServer.Addr),
		zap.String("path", s.webhookPath),
		zap.Int("workers", s.workers))

	// Start HTTP server
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Shutdown gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("Shutting down webhook server")
	return s.httpServer.Shutdown(shutdownCtx)
}

// handleWebhook processes incoming webhook requests
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature if secret is configured
	if s.secret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if !s.verifySignature(body, signature) {
			s.logger.Warn("Invalid webhook signature")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse event
	event, err := s.parseEvent(r.Header, body)
	if err != nil {
		s.logger.Error("Failed to parse event", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Queue event for processing
	select {
	case s.eventChannel <- event:
		s.logger.Debug("Event queued for processing",
			zap.String("id", event.ID),
			zap.String("type", event.Type))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"accepted","event_id":"%s"}`, event.ID)
	default:
		s.logger.Warn("Event queue full, dropping event",
			zap.String("id", event.ID),
			zap.String("type", event.Type))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleMetrics returns server metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.engine.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"events_processed":  metrics.EventsProcessed,
		"rules_evaluated":   metrics.RulesEvaluated,
		"actions_executed":  metrics.ActionsExecuted,
		"errors":            metrics.Errors,
		"avg_processing_ms": float64(metrics.ProcessingTime.Milliseconds()) / float64(metrics.EventsProcessed),
		"queue_size":        len(s.eventChannel),
	})
}

// verifySignature verifies the webhook signature
func (s *Server) verifySignature(payload []byte, signature string) bool {
	if signature == "" {
		return false
	}

	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha256" {
		return false
	}

	expectedMAC, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(payload)
	computedMAC := mac.Sum(nil)

	return hmac.Equal(computedMAC, expectedMAC)
}

// parseEvent parses a webhook event from the request
func (s *Server) parseEvent(headers http.Header, body []byte) (*Event, error) {
	eventType := headers.Get("X-GitHub-Event")
	if eventType == "" {
		return nil, fmt.Errorf("missing X-GitHub-Event header")
	}

	deliveryID := headers.Get("X-GitHub-Delivery")
	if deliveryID == "" {
		deliveryID = uuid.New().String()
	}

	event := &Event{
		ID:         deliveryID,
		Type:       eventType,
		ReceivedAt: time.Now(),
		Headers:    make(map[string]string),
		RawPayload: json.RawMessage(body),
	}

	// Copy relevant headers
	for key, values := range headers {
		if strings.HasPrefix(key, "X-GitHub-") && len(values) > 0 {
			event.Headers[key] = values[0]
		}
	}

	// Parse the payload based on event type
	switch eventType {
	case "push":
		var pushEvent github.PushEvent
		if err := json.Unmarshal(body, &pushEvent); err != nil {
			return nil, fmt.Errorf("failed to parse push event: %w", err)
		}
		// Convert PushEventRepository to Repository
		if pushEvent.Repo != nil {
			repo := &github.Repository{
				ID:       pushEvent.Repo.ID,
				Name:     pushEvent.Repo.Name,
				FullName: pushEvent.Repo.FullName,
				Private:  pushEvent.Repo.Private,
				Owner: &github.User{
					Login: pushEvent.Repo.Owner.Login,
					ID:    pushEvent.Repo.Owner.ID,
					Type:  pushEvent.Repo.Owner.Type,
				},
				DefaultBranch: pushEvent.Repo.DefaultBranch,
				Language:      pushEvent.Repo.Language,
			}
			event.Repository = repo
		}
		event.Sender = pushEvent.Sender
		event.Payload = &pushEvent

	case "pull_request":
		var prEvent github.PullRequestEvent
		if err := json.Unmarshal(body, &prEvent); err != nil {
			return nil, fmt.Errorf("failed to parse pull_request event: %w", err)
		}
		event.Action = prEvent.GetAction()
		event.Repository = prEvent.Repo
		event.Sender = prEvent.Sender
		event.Payload = &prEvent

	case "issues":
		var issueEvent github.IssuesEvent
		if err := json.Unmarshal(body, &issueEvent); err != nil {
			return nil, fmt.Errorf("failed to parse issues event: %w", err)
		}
		event.Action = issueEvent.GetAction()
		event.Repository = issueEvent.Repo
		event.Sender = issueEvent.Sender
		event.Payload = &issueEvent

	case "issue_comment":
		var commentEvent github.IssueCommentEvent
		if err := json.Unmarshal(body, &commentEvent); err != nil {
			return nil, fmt.Errorf("failed to parse issue_comment event: %w", err)
		}
		event.Action = commentEvent.GetAction()
		event.Repository = commentEvent.Repo
		event.Sender = commentEvent.Sender
		event.Payload = &commentEvent

	case "workflow_run":
		var workflowEvent github.WorkflowRunEvent
		if err := json.Unmarshal(body, &workflowEvent); err != nil {
			return nil, fmt.Errorf("failed to parse workflow_run event: %w", err)
		}
		event.Action = workflowEvent.GetAction()
		event.Repository = workflowEvent.Repo
		event.Sender = workflowEvent.Sender
		event.Payload = &workflowEvent

	case "release":
		var releaseEvent github.ReleaseEvent
		if err := json.Unmarshal(body, &releaseEvent); err != nil {
			return nil, fmt.Errorf("failed to parse release event: %w", err)
		}
		event.Action = releaseEvent.GetAction()
		event.Repository = releaseEvent.Repo
		event.Sender = releaseEvent.Sender
		event.Payload = &releaseEvent

	default:
		// For unknown event types, store the raw JSON
		var genericPayload map[string]interface{}
		if err := json.Unmarshal(body, &genericPayload); err != nil {
			return nil, fmt.Errorf("failed to parse event payload: %w", err)
		}
		event.Payload = genericPayload

		// Try to extract common fields
		if repo, ok := genericPayload["repository"].(map[string]interface{}); ok {
			repoJSON, _ := json.Marshal(repo)
			var repository github.Repository
			if err := json.Unmarshal(repoJSON, &repository); err == nil {
				event.Repository = &repository
			}
		}

		if sender, ok := genericPayload["sender"].(map[string]interface{}); ok {
			senderJSON, _ := json.Marshal(sender)
			var user github.User
			if err := json.Unmarshal(senderJSON, &user); err == nil {
				event.Sender = &user
			}
		}

		if action, ok := genericPayload["action"].(string); ok {
			event.Action = action
		}
	}

	return event, nil
}

// eventWorker processes events from the queue
func (s *Server) eventWorker(ctx context.Context, workerID int) {
	s.logger.Info("Starting event worker", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping event worker", zap.Int("worker_id", workerID))
			return
		case event := <-s.eventChannel:
			s.processEvent(ctx, event)
		}
	}
}

// processEvent processes a single event
func (s *Server) processEvent(ctx context.Context, event *Event) {
	s.logger.Debug("Processing event",
		zap.String("id", event.ID),
		zap.String("type", event.Type),
		zap.String("action", event.Action))

	if err := s.engine.ProcessEvent(ctx, event); err != nil {
		s.logger.Error("Failed to process event",
			zap.String("id", event.ID),
			zap.String("type", event.Type),
			zap.Error(err))
	}
}
