package event

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// Server represents a GitHub webhook server.
type Server struct {
	processor github.EventProcessor
	logger    github.Logger
	host      string
	port      int
	secret    string
}

// NewServer creates a new event server.
func NewServer(host string, port int, secret string, storage github.EventStorage, logger github.Logger) *Server {
	processor := github.NewEventProcessor(storage, logger)

	return &Server{
		processor: processor,
		logger:    logger,
		host:      host,
		port:      port,
		secret:    secret,
	}
}

// Start starts the webhook server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting GitHub webhook server", "host", s.host, "port", s.port)

	// Create webhook server
	webhookServer := github.NewEventWebhookServer(s.processor, s.secret, s.logger)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookServer.HandleWebhook)
	mux.HandleFunc("/health", webhookServer.GetHealthCheck)
	mux.HandleFunc("/metrics", s.handleMetrics)

	// Start server
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.logger.Info("Webhook server started", "address", addr)
	fmt.Printf("GitHub webhook server listening on %s\n", addr)
	fmt.Printf("Webhook endpoint: http://%s/webhook\n", addr)
	fmt.Printf("Health check: http://%s/health\n", addr)
	fmt.Printf("Metrics: http://%s/metrics\n", addr)

	return srv.ListenAndServe()
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.processor.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
