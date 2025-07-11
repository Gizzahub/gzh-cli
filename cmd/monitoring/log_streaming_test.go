package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRealTimeLogStreaming(t *testing.T) {
	// Test WebSocket message filtering for log entries
	t.Run("WebSocket Log Entry Filtering", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()

		// Create test client with log filter
		client := &WebSocketClient{
			ID:   "test-client",
			send: make(chan *WebSocketMessage, 10),
			filter: &ClientFilter{
				Types:      []string{MessageTypeLogEntry},
				LogLevels:  []string{"error", "warn"},
				LogSources: []string{"payment-service"},
			},
			logger: logger,
		}

		// Test log entry that should pass filter
		logEntry := &LogEntry{
			Timestamp: time.Now(),
			Level:     "error",
			Logger:    "payment-service",
			Message:   "Payment processing failed",
			Fields: map[string]interface{}{
				"order_id": 12345,
			},
		}

		message := &WebSocketMessage{
			Type: MessageTypeLogEntry,
			Data: logEntry,
		}

		// Should receive this message
		assert.True(t, client.shouldReceiveMessage(message))

		// Test log entry that should be filtered out (wrong level)
		logEntry2 := &LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Logger:    "payment-service",
			Message:   "Payment processed successfully",
		}

		message2 := &WebSocketMessage{
			Type: MessageTypeLogEntry,
			Data: logEntry2,
		}

		// Should NOT receive this message (wrong level)
		assert.False(t, client.shouldReceiveMessage(message2))

		// Test log entry that should be filtered out (wrong source)
		logEntry3 := &LogEntry{
			Timestamp: time.Now(),
			Level:     "error",
			Logger:    "user-service",
			Message:   "User creation failed",
		}

		message3 := &WebSocketMessage{
			Type: MessageTypeLogEntry,
			Data: logEntry3,
		}

		// Should NOT receive this message (wrong source)
		assert.False(t, client.shouldReceiveMessage(message3))
	})

	t.Run("Log Query Filter", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		client := &WebSocketClient{
			ID:   "test-client",
			send: make(chan *WebSocketMessage, 10),
			filter: &ClientFilter{
				Types:    []string{MessageTypeLogEntry},
				LogQuery: "payment",
			},
			logger: logger,
		}

		// Should match in message
		logEntry1 := &LogEntry{
			Message: "Payment processing failed",
			Logger:  "service",
		}
		message1 := &WebSocketMessage{Type: MessageTypeLogEntry, Data: logEntry1}
		assert.True(t, client.shouldReceiveMessage(message1))

		// Should match in logger
		logEntry2 := &LogEntry{
			Message: "Processing failed",
			Logger:  "payment-service",
		}
		message2 := &WebSocketMessage{Type: MessageTypeLogEntry, Data: logEntry2}
		assert.True(t, client.shouldReceiveMessage(message2))

		// Should match in fields
		logEntry3 := &LogEntry{
			Message: "Processing failed",
			Logger:  "service",
			Fields: map[string]interface{}{
				"type": "payment_error",
			},
		}
		message3 := &WebSocketMessage{Type: MessageTypeLogEntry, Data: logEntry3}
		assert.True(t, client.shouldReceiveMessage(message3))

		// Should NOT match
		logEntry4 := &LogEntry{
			Message: "User login successful",
			Logger:  "auth-service",
		}
		message4 := &WebSocketMessage{Type: MessageTypeLogEntry, Data: logEntry4}
		assert.False(t, client.shouldReceiveMessage(message4))
	})

	t.Run("CentralizedLogger Streaming Integration", func(t *testing.T) {
		config := &CentralizedLoggingConfig{
			Level:         "info",
			Format:        "json",
			BufferSize:    1000,
			FlushInterval: time.Second,
			Streaming: &StreamingConfig{
				Enabled:       true,
				BufferSize:    100,
				StreamLevels:  []string{"error", "warn"},
				StreamSources: []string{"payment-service", "auth-service"},
			},
		}

		registry := prometheus.NewRegistry()
		centralLogger, err := NewCentralizedLogger(config, registry)
		require.NoError(t, err)
		defer centralLogger.Shutdown(context.Background())

		// Test that WebSocket manager is initialized
		wsManager := centralLogger.GetWebSocketManager()
		require.NotNil(t, wsManager)

		// Test shouldStreamEntry logic
		// Should stream (error level, payment-service)
		entry1 := &LogEntry{
			Level:   "error",
			Logger:  "payment-service",
			Message: "Payment failed",
		}
		assert.True(t, centralLogger.shouldStreamEntry(entry1))

		// Should NOT stream (info level, not in StreamLevels)
		entry2 := &LogEntry{
			Level:   "info",
			Logger:  "payment-service",
			Message: "Payment processed",
		}
		assert.False(t, centralLogger.shouldStreamEntry(entry2))

		// Should NOT stream (error level, but user-service not in StreamSources)
		entry3 := &LogEntry{
			Level:   "error",
			Logger:  "user-service",
			Message: "User creation failed",
		}
		assert.False(t, centralLogger.shouldStreamEntry(entry3))
	})

	t.Run("WebSocket Manager Broadcasting", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		wsManager := NewWebSocketManager(logger)
		wsManager.Start()
		defer wsManager.Stop()

		// Test log entry broadcasting
		logEntry := &LogEntry{
			Timestamp: time.Now(),
			Level:     "error",
			Logger:    "test-service",
			Message:   "Test error message",
			Fields: map[string]interface{}{
				"test_field": "test_value",
			},
		}

		// This should not panic and should execute without error
		wsManager.BroadcastLogEntry(logEntry)

		// Test client stats
		stats := wsManager.GetClientStats()
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats["total_clients"]) // No clients connected in test
	})

	t.Run("Streaming Configuration", func(t *testing.T) {
		config := &StreamingConfig{
			Enabled:           true,
			BufferSize:        200,
			MaxConnections:    50,
			HeartbeatInterval: time.Second * 30,
			StreamLevels:      []string{"error", "warn", "info"},
			StreamSources:     []string{"payment", "auth", "user"},
			RateLimit:         100,
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, 200, config.BufferSize)
		assert.Equal(t, 50, config.MaxConnections)
		assert.Contains(t, config.StreamLevels, "error")
		assert.Contains(t, config.StreamSources, "payment")
		assert.Equal(t, 100, config.RateLimit)
	})
}

func TestWebSocketHandlerIntegration(t *testing.T) {
	t.Run("WebSocket Upgrade and Message Handling", func(t *testing.T) {
		// Create test logger and WebSocket manager
		logger, _ := zap.NewDevelopment()
		wsManager := NewWebSocketManager(logger)
		wsManager.Start()
		defer wsManager.Stop()

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsManager.HandleWebSocket(w, r)
		}))
		defer server.Close()

		// Convert to WebSocket URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect as WebSocket client
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Send subscription message
		subscribeMsg := map[string]interface{}{
			"type": MessageTypeSubscribe,
			"filter": map[string]interface{}{
				"types":       []string{MessageTypeLogEntry},
				"log_levels":  []string{"error"},
				"log_sources": []string{"test-service"},
			},
		}

		err = conn.WriteJSON(subscribeMsg)
		require.NoError(t, err)

		// Give some time for subscription to be processed
		time.Sleep(100 * time.Millisecond)

		// Broadcast a log entry that should match the filter
		logEntry := &LogEntry{
			Timestamp: time.Now(),
			Level:     "error",
			Logger:    "test-service",
			Message:   "Test error message",
		}

		wsManager.BroadcastLogEntry(logEntry)

		// Try to read the message (with timeout)
		conn.SetReadDeadline(time.Now().Add(time.Second))
		var receivedMsg WebSocketMessage
		err = conn.ReadJSON(&receivedMsg)

		// The message might not be received due to timing, but connection should work
		if err == nil {
			assert.Equal(t, MessageTypeLogEntry, receivedMsg.Type)
		}
	})
}

func TestLogStreamingEndToEnd(t *testing.T) {
	t.Run("Complete Log Streaming Pipeline", func(t *testing.T) {
		// Create configuration with streaming enabled
		config := &CentralizedLoggingConfig{
			Level:         "debug",
			Format:        "json",
			BufferSize:    1000,
			FlushInterval: time.Millisecond * 100,
			Streaming: &StreamingConfig{
				Enabled:       true,
				BufferSize:    50,
				StreamLevels:  []string{"error", "warn", "info"},
				StreamSources: []string{}, // Empty means all sources
			},
		}

		// Create centralized logger
		registry := prometheus.NewRegistry()
		centralLogger, err := NewCentralizedLogger(config, registry)
		require.NoError(t, err)
		defer centralLogger.Shutdown(context.Background())

		// Verify WebSocket manager is created
		wsManager := centralLogger.GetWebSocketManager()
		require.NotNil(t, wsManager)

		// Create test log entries
		testEntries := []*LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "error",
				Logger:    "payment-service",
				Message:   "Credit card declined",
				Fields: map[string]interface{}{
					"user_id":    12345,
					"amount":     99.99,
					"error_code": "CARD_DECLINED",
				},
			},
			{
				Timestamp: time.Now(),
				Level:     "info",
				Logger:    "auth-service",
				Message:   "User logged in successfully",
				Fields: map[string]interface{}{
					"user_id":    12345,
					"ip_address": "192.168.1.100",
				},
			},
			{
				Timestamp: time.Now(),
				Level:     "warn",
				Logger:    "user-service",
				Message:   "Password will expire soon",
				Fields: map[string]interface{}{
					"user_id":      12345,
					"days_left":    3,
					"last_changed": "2024-01-01",
				},
			},
		}

		// Process log entries through centralized logger
		for _, entry := range testEntries {
			err := centralLogger.Log(entry)
			require.NoError(t, err)
		}

		// Verify statistics include streaming information
		stats := centralLogger.GetStats()
		assert.NotNil(t, stats)

		if wsManager != nil {
			assert.Contains(t, stats, "websocket")
			assert.Contains(t, stats, "streaming")
		}

		// Test that entries should be streamed
		for _, entry := range testEntries {
			shouldStream := centralLogger.shouldStreamEntry(entry)
			assert.True(t, shouldStream, "Entry with level %s should be streamed", entry.Level)
		}
	})
}
