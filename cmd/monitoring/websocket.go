package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocket message types
const (
	MessageTypeSystemStatus = "system_status"
	MessageTypeMetrics      = "metrics_update"
	MessageTypeTaskUpdate   = "task_update"
	MessageTypeAlert        = "alert"
	MessageTypePing         = "ping"
	MessageTypePong         = "pong"
	MessageTypeSubscribe    = "subscribe"
	MessageTypeUnsubscribe  = "unsubscribe"
	MessageTypeInitialState = "initial_state"
	// Log streaming message types
	MessageTypeLogStream    = "log_stream"
	MessageTypeLogEntry     = "log_entry"
	MessageTypeLogSubscribe = "log_subscribe"
	MessageTypeLogFilter    = "log_filter"
)

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ClientFilter defines what types of messages a client wants to receive
type ClientFilter struct {
	Types      []string `json:"types,omitempty"`      // Message types to receive
	TaskIDs    []string `json:"task_ids,omitempty"`   // Specific task IDs to monitor
	Severity   []string `json:"severity,omitempty"`   // Alert severity levels
	Components []string `json:"components,omitempty"` // Specific components to monitor
	// Log streaming filters
	LogLevels  []string          `json:"log_levels,omitempty"`  // debug, info, warn, error
	LogSources []string          `json:"log_sources,omitempty"` // specific loggers/services
	LogFields  map[string]string `json:"log_fields,omitempty"`  // field-based filtering
	LogQuery   string            `json:"log_query,omitempty"`   // search query for logs
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID          string
	conn        *websocket.Conn
	send        chan *WebSocketMessage
	filter      *ClientFilter
	user        *User // Authenticated user
	mu          sync.RWMutex
	isConnected bool
	logger      *zap.Logger
	lastPing    time.Time
}

// WebSocketHub manages all WebSocket connections
type WebSocketHub struct {
	clients    map[string]*WebSocketClient
	broadcast  chan *WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
	logger     *zap.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(logger *zap.Logger) *WebSocketHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketHub{
		clients:    make(map[string]*WebSocketClient),
		broadcast:  make(chan *WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info("WebSocket hub shutting down")
			h.closeAllClients()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			h.logger.Info("WebSocket client registered",
				zap.String("client_id", client.ID),
				zap.Int("total_clients", len(h.clients)))

			// Send initial state to new client
			h.sendInitialState(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
				h.mu.Unlock()
				h.logger.Info("WebSocket client unregistered",
					zap.String("client_id", client.ID),
					zap.Int("total_clients", len(h.clients)))
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := make([]*WebSocketClient, 0, len(h.clients))
			for _, client := range h.clients {
				clients = append(clients, client)
			}
			h.mu.RUnlock()

			// Send to all clients that match the filter
			for _, client := range clients {
				if client.shouldReceiveMessage(message) {
					select {
					case client.send <- message:
					default:
						// Client's send channel is full, close it
						h.logger.Warn("Client send channel full, closing connection",
							zap.String("client_id", client.ID))
						go func(c *WebSocketClient) {
							h.unregister <- c
						}(client)
					}
				}
			}

		case <-ticker.C:
			// Ping all clients to check connection health
			h.pingAllClients()
		}
	}
}

// Stop gracefully shuts down the hub
func (h *WebSocketHub) Stop() {
	h.cancel()
}

// BroadcastMessage sends a message to all connected clients
func (h *WebSocketHub) BroadcastMessage(msgType string, data interface{}) {
	message := &WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case h.broadcast <- message:
	case <-time.After(time.Second):
		h.logger.Warn("Failed to broadcast message, channel full")
	}
}

// sendInitialState sends the current system state to a newly connected client
func (h *WebSocketHub) sendInitialState(client *WebSocketClient) {
	// This would gather current system state
	initialState := map[string]interface{}{
		"connected_at": time.Now(),
		"server_time":  time.Now(),
		"version":      "1.0.0",
		// Add more initial state data as needed
	}

	message := &WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      MessageTypeInitialState,
		Timestamp: time.Now(),
		Data:      initialState,
	}

	select {
	case client.send <- message:
	default:
		h.logger.Warn("Failed to send initial state to client",
			zap.String("client_id", client.ID))
	}
}

// pingAllClients sends ping messages to check connection health
func (h *WebSocketHub) pingAllClients() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	pingMessage := &WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      MessageTypePing,
		Timestamp: time.Now(),
	}

	for _, client := range h.clients {
		select {
		case client.send <- pingMessage:
			client.lastPing = time.Now()
		default:
			h.logger.Warn("Failed to ping client",
				zap.String("client_id", client.ID))
		}
	}
}

// closeAllClients closes all client connections
func (h *WebSocketHub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, client := range h.clients {
		close(client.send)
		client.conn.Close()
	}
	h.clients = make(map[string]*WebSocketClient)
}

// shouldReceiveMessage checks if a client should receive a specific message
func (c *WebSocketClient) shouldReceiveMessage(msg *WebSocketMessage) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// If no filter is set, receive all messages
	if c.filter == nil {
		return true
	}

	// Check message type filter
	if len(c.filter.Types) > 0 {
		typeMatch := false
		for _, t := range c.filter.Types {
			if t == msg.Type {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}

	// Check task ID filter for task updates
	if msg.Type == MessageTypeTaskUpdate && len(c.filter.TaskIDs) > 0 {
		if taskData, ok := msg.Data.(map[string]interface{}); ok {
			if taskID, ok := taskData["task_id"].(string); ok {
				taskMatch := false
				for _, id := range c.filter.TaskIDs {
					if id == taskID {
						taskMatch = true
						break
					}
				}
				if !taskMatch {
					return false
				}
			}
		}
	}

	// Check severity filter for alerts
	if msg.Type == MessageTypeAlert && len(c.filter.Severity) > 0 {
		if alertData, ok := msg.Data.(map[string]interface{}); ok {
			if severity, ok := alertData["severity"].(string); ok {
				severityMatch := false
				for _, s := range c.filter.Severity {
					if s == severity {
						severityMatch = true
						break
					}
				}
				if !severityMatch {
					return false
				}
			}
		}
	}

	// Check log entry filters
	if msg.Type == MessageTypeLogEntry {
		logEntry, ok := msg.Data.(*LogEntry)
		if !ok {
			return false
		}

		// Check log level filter
		if len(c.filter.LogLevels) > 0 {
			levelMatch := false
			for _, level := range c.filter.LogLevels {
				if level == logEntry.Level {
					levelMatch = true
					break
				}
			}
			if !levelMatch {
				return false
			}
		}

		// Check log source (logger) filter
		if len(c.filter.LogSources) > 0 {
			sourceMatch := false
			for _, source := range c.filter.LogSources {
				if source == logEntry.Logger {
					sourceMatch = true
					break
				}
			}
			if !sourceMatch {
				return false
			}
		}

		// Check log fields filter
		if len(c.filter.LogFields) > 0 {
			for filterKey, filterValue := range c.filter.LogFields {
				if logValue, exists := logEntry.Fields[filterKey]; exists {
					if fmt.Sprintf("%v", logValue) != filterValue {
						return false
					}
				} else {
					return false
				}
			}
		}

		// Check log query filter (simple contains check)
		if c.filter.LogQuery != "" {
			queryMatch := false
			// Search in message
			if strings.Contains(strings.ToLower(logEntry.Message), strings.ToLower(c.filter.LogQuery)) {
				queryMatch = true
			}
			// Search in logger
			if strings.Contains(strings.ToLower(logEntry.Logger), strings.ToLower(c.filter.LogQuery)) {
				queryMatch = true
			}
			// Search in field values
			for _, value := range logEntry.Fields {
				if valueStr := fmt.Sprintf("%v", value); strings.Contains(strings.ToLower(valueStr), strings.ToLower(c.filter.LogQuery)) {
					queryMatch = true
					break
				}
			}
			if !queryMatch {
				return false
			}
		}
	}

	return true
}

// readPump handles incoming messages from the client
func (c *WebSocketClient) readPump(hub *WebSocketHub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
		c.isConnected = false
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message map[string]interface{}
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error",
					zap.String("client_id", c.ID),
					zap.Error(err))
			}
			break
		}

		// Handle client messages
		if msgType, ok := message["type"].(string); ok {
			switch msgType {
			case MessageTypeSubscribe:
				c.handleSubscribe(message)
			case MessageTypeUnsubscribe:
				c.handleUnsubscribe()
			case MessageTypePong:
				// Client responded to ping
				c.lastPing = time.Now()
			}
		}
	}
}

// writePump handles outgoing messages to the client
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.isConnected = false
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				c.logger.Error("WebSocket write error",
					zap.String("client_id", c.ID),
					zap.Error(err))
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe updates the client's filter based on subscription message
func (c *WebSocketClient) handleSubscribe(message map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if filterData, ok := message["filter"].(map[string]interface{}); ok {
		filter := &ClientFilter{}

		// Parse filter types
		if types, ok := filterData["types"].([]interface{}); ok {
			filter.Types = make([]string, 0, len(types))
			for _, t := range types {
				if str, ok := t.(string); ok {
					filter.Types = append(filter.Types, str)
				}
			}
		}

		// Parse task IDs
		if taskIDs, ok := filterData["task_ids"].([]interface{}); ok {
			filter.TaskIDs = make([]string, 0, len(taskIDs))
			for _, id := range taskIDs {
				if str, ok := id.(string); ok {
					filter.TaskIDs = append(filter.TaskIDs, str)
				}
			}
		}

		// Parse severity levels
		if severity, ok := filterData["severity"].([]interface{}); ok {
			filter.Severity = make([]string, 0, len(severity))
			for _, s := range severity {
				if str, ok := s.(string); ok {
					filter.Severity = append(filter.Severity, str)
				}
			}
		}

		// Parse log levels
		if logLevels, ok := filterData["log_levels"].([]interface{}); ok {
			filter.LogLevels = make([]string, 0, len(logLevels))
			for _, l := range logLevels {
				if str, ok := l.(string); ok {
					filter.LogLevels = append(filter.LogLevels, str)
				}
			}
		}

		// Parse log sources
		if logSources, ok := filterData["log_sources"].([]interface{}); ok {
			filter.LogSources = make([]string, 0, len(logSources))
			for _, s := range logSources {
				if str, ok := s.(string); ok {
					filter.LogSources = append(filter.LogSources, str)
				}
			}
		}

		// Parse log fields
		if logFields, ok := filterData["log_fields"].(map[string]interface{}); ok {
			filter.LogFields = make(map[string]string)
			for k, v := range logFields {
				if str, ok := v.(string); ok {
					filter.LogFields[k] = str
				}
			}
		}

		// Parse log query
		if logQuery, ok := filterData["log_query"].(string); ok {
			filter.LogQuery = logQuery
		}

		c.filter = filter
		c.logger.Info("Client filter updated",
			zap.String("client_id", c.ID),
			zap.Any("filter", filter))
	}
}

// handleUnsubscribe clears the client's filter
func (c *WebSocketClient) handleUnsubscribe() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.filter = nil
	c.logger.Info("Client filter cleared",
		zap.String("client_id", c.ID))
}

// WebSocketManager integrates WebSocket with the monitoring server
type WebSocketManager struct {
	hub      *WebSocketHub
	upgrader websocket.Upgrader
	logger   *zap.Logger
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(logger *zap.Logger) *WebSocketManager {
	return &WebSocketManager{
		hub: NewWebSocketHub(logger),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: logger,
	}
}

// Start starts the WebSocket manager
func (m *WebSocketManager) Start() {
	go m.hub.Run()
}

// Stop stops the WebSocket manager
func (m *WebSocketManager) Stop() {
	m.hub.Stop()
}

// HandleWebSocket handles WebSocket upgrade requests (without authentication)
func (m *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}

	client := &WebSocketClient{
		ID:          uuid.New().String(),
		conn:        conn,
		send:        make(chan *WebSocketMessage, 256),
		isConnected: true,
		logger:      m.logger,
		lastPing:    time.Now(),
	}

	m.hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump(m.hub)
}

// HandleAuthenticatedWebSocket handles WebSocket upgrade requests with authentication
func (m *WebSocketManager) HandleAuthenticatedWebSocket(w http.ResponseWriter, r *http.Request, user *User) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}

	client := &WebSocketClient{
		ID:          uuid.New().String(),
		conn:        conn,
		send:        make(chan *WebSocketMessage, 256),
		user:        user,
		isConnected: true,
		logger:      m.logger,
		lastPing:    time.Now(),
	}

	m.hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump(m.hub)
}

// BroadcastSystemStatus broadcasts system status update
func (m *WebSocketManager) BroadcastSystemStatus(status interface{}) {
	m.hub.BroadcastMessage(MessageTypeSystemStatus, status)
}

// BroadcastMetrics broadcasts metrics update
func (m *WebSocketManager) BroadcastMetrics(metrics interface{}) {
	m.hub.BroadcastMessage(MessageTypeMetrics, metrics)
}

// BroadcastTaskUpdate broadcasts task update
func (m *WebSocketManager) BroadcastTaskUpdate(taskUpdate interface{}) {
	m.hub.BroadcastMessage(MessageTypeTaskUpdate, taskUpdate)
}

// BroadcastAlert broadcasts alert
func (m *WebSocketManager) BroadcastAlert(alert interface{}) {
	m.hub.BroadcastMessage(MessageTypeAlert, alert)
}

// BroadcastLogEntry broadcasts a log entry to subscribed clients
func (m *WebSocketManager) BroadcastLogEntry(logEntry *LogEntry) {
	// Create metadata for the log entry
	metadata := map[string]interface{}{
		"source": "centralized_logging",
		"type":   "log_entry",
	}

	message := &WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      MessageTypeLogEntry,
		Timestamp: time.Now(),
		Data:      logEntry,
		Metadata:  metadata,
	}

	select {
	case m.hub.broadcast <- message:
	case <-time.After(time.Millisecond * 100):
		// Don't block log processing if WebSocket is slow
		m.logger.Debug("WebSocket broadcast timeout for log entry")
	}
}

// BroadcastLogStream broadcasts log stream start/stop events
func (m *WebSocketManager) BroadcastLogStream(streamEvent interface{}) {
	m.hub.BroadcastMessage(MessageTypeLogStream, streamEvent)
}

// GetConnectedClients returns the number of connected clients
func (m *WebSocketManager) GetConnectedClients() int {
	m.hub.mu.RLock()
	defer m.hub.mu.RUnlock()
	return len(m.hub.clients)
}

// GetClientStats returns statistics about connected clients
func (m *WebSocketManager) GetClientStats() map[string]interface{} {
	m.hub.mu.RLock()
	defer m.hub.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients": len(m.hub.clients),
		"clients":       make([]map[string]interface{}, 0, len(m.hub.clients)),
	}

	for _, client := range m.hub.clients {
		clientInfo := map[string]interface{}{
			"id":        client.ID,
			"connected": client.isConnected,
			"last_ping": client.lastPing,
		}
		if client.filter != nil {
			clientInfo["filter"] = client.filter
		}
		stats["clients"] = append(stats["clients"].([]map[string]interface{}), clientInfo)
	}

	return stats
}
