package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestWebSocketHub(t *testing.T) {
	logger := zaptest.NewLogger(t)
	hub := NewWebSocketHub(logger)

	// Start hub
	go hub.Run()
	defer hub.Stop()

	// Give hub time to start
	time.Sleep(10 * time.Millisecond)

	t.Run("BroadcastMessage", func(t *testing.T) {
		// Test broadcasting
		hub.BroadcastMessage(MessageTypeSystemStatus, map[string]interface{}{
			"status": "healthy",
			"uptime": "1h",
		})

		// Since we don't have clients connected, just verify no panic
		assert.NotPanics(t, func() {
			hub.BroadcastMessage(MessageTypeMetrics, map[string]interface{}{
				"cpu": 50.0,
				"mem": 1024,
			})
		})
	})

	t.Run("ClientRegistration", func(t *testing.T) {
		// Create mock client
		client := &WebSocketClient{
			ID:          "test-client",
			send:        make(chan *WebSocketMessage, 1),
			isConnected: true,
			logger:      logger,
			lastPing:    time.Now(),
		}

		// Register client
		hub.register <- client

		// Give time for registration
		time.Sleep(10 * time.Millisecond)

		// Check client count
		hub.mu.RLock()
		count := len(hub.clients)
		hub.mu.RUnlock()
		assert.Equal(t, 1, count)

		// Unregister client
		hub.unregister <- client

		// Give time for unregistration
		time.Sleep(10 * time.Millisecond)

		// Check client count
		hub.mu.RLock()
		count = len(hub.clients)
		hub.mu.RUnlock()
		assert.Equal(t, 0, count)
	})
}

func TestWebSocketClient_MessageFiltering(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name     string
		filter   *ClientFilter
		message  *WebSocketMessage
		expected bool
	}{
		{
			name:   "No filter - receive all",
			filter: nil,
			message: &WebSocketMessage{
				Type: MessageTypeSystemStatus,
				Data: map[string]interface{}{"status": "healthy"},
			},
			expected: true,
		},
		{
			name: "Type filter - match",
			filter: &ClientFilter{
				Types: []string{MessageTypeSystemStatus, MessageTypeMetrics},
			},
			message: &WebSocketMessage{
				Type: MessageTypeSystemStatus,
				Data: map[string]interface{}{"status": "healthy"},
			},
			expected: true,
		},
		{
			name: "Type filter - no match",
			filter: &ClientFilter{
				Types: []string{MessageTypeAlert},
			},
			message: &WebSocketMessage{
				Type: MessageTypeSystemStatus,
				Data: map[string]interface{}{"status": "healthy"},
			},
			expected: false,
		},
		{
			name: "Task ID filter - match",
			filter: &ClientFilter{
				TaskIDs: []string{"task-1", "task-2"},
			},
			message: &WebSocketMessage{
				Type: MessageTypeTaskUpdate,
				Data: map[string]interface{}{"task_id": "task-1"},
			},
			expected: true,
		},
		{
			name: "Task ID filter - no match",
			filter: &ClientFilter{
				TaskIDs: []string{"task-1", "task-2"},
			},
			message: &WebSocketMessage{
				Type: MessageTypeTaskUpdate,
				Data: map[string]interface{}{"task_id": "task-3"},
			},
			expected: false,
		},
		{
			name: "Severity filter - match",
			filter: &ClientFilter{
				Severity: []string{"critical", "high"},
			},
			message: &WebSocketMessage{
				Type: MessageTypeAlert,
				Data: map[string]interface{}{"severity": "critical"},
			},
			expected: true,
		},
		{
			name: "Combined filter",
			filter: &ClientFilter{
				Types:    []string{MessageTypeAlert},
				Severity: []string{"critical"},
			},
			message: &WebSocketMessage{
				Type: MessageTypeAlert,
				Data: map[string]interface{}{"severity": "critical"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &WebSocketClient{
				filter: tt.filter,
				logger: logger,
			}

			result := client.shouldReceiveMessage(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebSocketManager_Integration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewWebSocketManager(logger)

	// Start manager
	manager.Start()
	defer manager.Stop()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(manager.HandleWebSocket))
	defer server.Close()

	// Convert http to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Run("ClientConnection", func(t *testing.T) {
		// Connect client
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer ws.Close()

		// Wait for connection
		time.Sleep(100 * time.Millisecond)

		// Check connected clients
		count := manager.GetConnectedClients()
		assert.Equal(t, 1, count)

		// Send subscribe message
		subscribeMsg := map[string]interface{}{
			"type": MessageTypeSubscribe,
			"filter": map[string]interface{}{
				"types": []string{MessageTypeSystemStatus},
			},
		}
		err = ws.WriteJSON(subscribeMsg)
		require.NoError(t, err)

		// Broadcast a message
		manager.BroadcastSystemStatus(map[string]interface{}{
			"status": "healthy",
			"uptime": "1h",
		})

		// Read message
		var msg WebSocketMessage
		err = ws.ReadJSON(&msg)
		require.NoError(t, err)

		// Could be initial state or broadcast message
		assert.Contains(t, []string{MessageTypeInitialState, MessageTypeSystemStatus}, msg.Type)
	})

	t.Run("MultipleClients", func(t *testing.T) {
		// Connect multiple clients
		clients := make([]*websocket.Conn, 3)
		for i := 0; i < 3; i++ {
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			clients[i] = ws
		}

		// Wait for connections
		time.Sleep(100 * time.Millisecond)

		// Check connected clients
		count := manager.GetConnectedClients()
		assert.GreaterOrEqual(t, count, 3)

		// Clean up
		for _, ws := range clients {
			ws.Close()
		}
	})

	t.Run("ClientStats", func(t *testing.T) {
		stats := manager.GetClientStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "total_clients")
		assert.Contains(t, stats, "clients")
	})
}

func TestWebSocketMessage_JSON(t *testing.T) {
	msg := &WebSocketMessage{
		ID:        "test-123",
		Type:      MessageTypeSystemStatus,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status": "healthy",
			"uptime": "1h",
		},
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	// Marshal
	data, err := json.Marshal(msg)
	require.NoError(t, err)

	// Unmarshal
	var decoded WebSocketMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, msg.ID, decoded.ID)
	assert.Equal(t, msg.Type, decoded.Type)
	assert.NotNil(t, decoded.Data)
}

func BenchmarkWebSocketHub_Broadcast(b *testing.B) {
	logger := zaptest.NewLogger(b)
	hub := NewWebSocketHub(logger)

	// Start hub
	go hub.Run()
	defer hub.Stop()

	// Add some clients
	for i := 0; i < 100; i++ {
		client := &WebSocketClient{
			ID:          fmt.Sprintf("client-%d", i),
			send:        make(chan *WebSocketMessage, 256),
			isConnected: true,
			logger:      logger,
			lastPing:    time.Now(),
		}
		hub.register <- client
	}

	// Give time for registration
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hub.BroadcastMessage(MessageTypeMetrics, map[string]interface{}{
			"cpu":    float64(i % 100),
			"memory": i * 1024,
		})
	}
}
