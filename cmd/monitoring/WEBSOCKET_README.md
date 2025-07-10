# WebSocket Real-Time Updates for GZH Monitoring

This document describes the WebSocket implementation for real-time monitoring updates in the GZH Manager monitoring system.

## Overview

The monitoring system now includes WebSocket support for real-time updates, allowing clients to receive live system metrics, task updates, and alerts without polling.

## Architecture

### Components

1. **WebSocketHub** (`websocket.go`)
   - Manages all WebSocket client connections
   - Broadcasts messages to connected clients
   - Supports message filtering per client

2. **WebSocketClient** (`websocket.go`)
   - Represents individual WebSocket connections
   - Handles bidirectional communication
   - Supports custom message filters

3. **WebSocketManager** (`websocket.go`)
   - Integrates WebSocket functionality with the monitoring server
   - Sends periodic system updates
   - Manages lifecycle of WebSocket components

4. **Dashboard** (`monitoring.go` - embedded HTML)
   - Simple real-time dashboard
   - Displays system metrics with live updates
   - Auto-reconnects on connection loss

## Usage

### Starting the Monitoring Server

```bash
# Start the monitoring server with WebSocket support
gz monitoring server

# With custom port
gz monitoring server -p 9090

# With debug mode
gz monitoring server -d
```

The server will display:
- Dashboard URL: `http://localhost:8080/dashboard`
- WebSocket endpoint: `ws://localhost:8080/ws`
- API endpoints: `http://localhost:8080/api/v1/*`

### WebSocket Protocol

#### Connection

Connect to the WebSocket endpoint at `/ws`. The server will automatically send an initial state message upon connection.

#### Message Types

##### Client to Server

1. **Subscribe**
```json
{
  "type": "subscribe",
  "filter": {
    "types": ["system_status", "task_update", "alert"],
    "task_ids": ["task-123", "task-456"],
    "alert_level": "high"
  }
}
```

2. **Ping**
```json
{
  "type": "ping"
}
```

3. **Unsubscribe**
```json
{
  "type": "unsubscribe"
}
```

##### Server to Client

1. **Initial State**
```json
{
  "type": "initial_state",
  "timestamp": "2024-01-10T10:00:00Z",
  "data": {
    "system_status": { ... },
    "metrics_summary": { ... },
    "alert_stats": { ... },
    "active_tasks": [ ... ]
  }
}
```

2. **System Status**
```json
{
  "type": "system_status",
  "timestamp": "2024-01-10T10:00:00Z",
  "data": {
    "status": "healthy",
    "uptime": "2h30m",
    "active_tasks": 5,
    "total_requests": 1234,
    "memory_usage": 536870912,
    "cpu_usage": 25.5,
    "disk_usage": 45.0,
    "network_io": {
      "bytes_in": 1024000,
      "bytes_out": 2048000
    }
  }
}
```

3. **Metrics Update**
```json
{
  "type": "metrics_update",
  "timestamp": "2024-01-10T10:00:00Z",
  "data": {
    "active_tasks": 5,
    "total_requests": 1234,
    "error_rate": 0.5,
    "avg_response_time": "150ms",
    "memory_usage_mb": 512.5,
    "cpu_usage": 25.5,
    "uptime": "2h30m"
  }
}
```

4. **Task Update**
```json
{
  "type": "task_update",
  "timestamp": "2024-01-10T10:00:00Z",
  "data": {
    "id": "task-123",
    "name": "Bulk Clone GitHub",
    "status": "running",
    "progress": 75,
    "start_time": "2024-01-10T09:30:00Z",
    "details": {
      "processed": 150,
      "total": 200,
      "errors": 2
    }
  }
}
```

5. **Alert**
```json
{
  "type": "alert",
  "timestamp": "2024-01-10T10:00:00Z",
  "data": {
    "id": "alert-789",
    "rule_id": "cpu-high",
    "rule_name": "High CPU Usage",
    "status": "firing",
    "severity": "critical",
    "message": "CPU usage above 90% for 5 minutes",
    "value": 92.5,
    "threshold": 90.0,
    "starts_at": "2024-01-10T09:55:00Z"
  }
}
```

6. **Pong**
```json
{
  "type": "pong",
  "timestamp": "2024-01-10T10:00:00Z",
  "data": {
    "client_id": "client-123456789"
  }
}
```

### Client Filtering

Clients can subscribe to specific types of updates to reduce bandwidth:

- **Message Types**: Filter by message type (e.g., only receive alerts)
- **Task IDs**: Only receive updates for specific tasks
- **Alert Level**: Only receive alerts above a certain severity level

## Implementation Example

### JavaScript Client

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

// Handle connection open
ws.onopen = function() {
    console.log('Connected to monitoring server');
    
    // Subscribe to specific updates
    ws.send(JSON.stringify({
        type: 'subscribe',
        filter: {
            types: ['system_status', 'alert'],
            alert_level: 'high'
        }
    }));
};

// Handle messages
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    switch(message.type) {
        case 'system_status':
            updateSystemMetrics(message.data);
            break;
        case 'alert':
            showAlert(message.data);
            break;
    }
};

// Handle errors
ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};

// Handle connection close
ws.onclose = function() {
    console.log('Disconnected from monitoring server');
    // Implement reconnection logic
};

// Keep connection alive
setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }));
    }
}, 30000);
```

### Go Client

```go
import (
    "github.com/gorilla/websocket"
)

// Connect to monitoring WebSocket
conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
if err != nil {
    log.Fatal("dial:", err)
}
defer conn.Close()

// Subscribe to updates
subscribe := map[string]interface{}{
    "type": "subscribe",
    "filter": map[string]interface{}{
        "types": []string{"all"},
    },
}
if err := conn.WriteJSON(subscribe); err != nil {
    log.Fatal("subscribe:", err)
}

// Read messages
for {
    var msg map[string]interface{}
    err := conn.ReadJSON(&msg)
    if err != nil {
        log.Println("read:", err)
        break
    }
    
    // Process message based on type
    switch msg["type"] {
    case "system_status":
        handleSystemStatus(msg["data"])
    case "alert":
        handleAlert(msg["data"])
    }
}
```

## Features

1. **Auto-reconnection**: Clients automatically reconnect on connection loss
2. **Message Filtering**: Reduce bandwidth by subscribing to specific updates
3. **Real-time Updates**: System metrics updated every 5 seconds
4. **Alert Notifications**: Immediate alert notifications based on rules
5. **Task Progress**: Live task progress updates
6. **Connection Health**: Ping/pong mechanism to detect stale connections

## Testing

Run the WebSocket tests:

```bash
# Run all monitoring tests
go test ./cmd/monitoring -v

# Run only WebSocket tests
go test ./cmd/monitoring -v -run TestWebSocket

# Run with race detection
go test ./cmd/monitoring -race

# Benchmark WebSocket broadcast performance
go test ./cmd/monitoring -bench=BenchmarkWebSocketBroadcast
```

## Performance Considerations

1. **Message Buffer**: Each client has a 256-message buffer to handle bursts
2. **Broadcast Channel**: Hub has a 100-message broadcast buffer
3. **Concurrent Broadcasting**: Messages are sent to clients concurrently
4. **Connection Limits**: Consider implementing connection limits for production
5. **Message Size**: Keep messages small for better performance

## Security Considerations

1. **Origin Validation**: Currently allows all origins (update for production)
2. **Authentication**: Implement token-based authentication for production
3. **Rate Limiting**: Add rate limiting to prevent abuse
4. **Message Validation**: Validate all incoming messages
5. **TLS/SSL**: Use wss:// in production environments

## Future Enhancements

1. **Message History**: Store recent messages for replay on reconnection
2. **Client Groups**: Support for client groups/rooms
3. **Binary Messages**: Support for binary data transmission
4. **Compression**: Enable per-message compression
5. **Metrics**: Add WebSocket-specific metrics (connections, messages/sec, etc.)