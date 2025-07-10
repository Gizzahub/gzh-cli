# Monitoring WebSocket Real-time Updates Guide

This guide describes the WebSocket real-time update feature in the GZH Manager monitoring system, enabling live updates for system status, metrics, tasks, and alerts.

## Overview

The WebSocket implementation provides:
- **Real-time updates** without polling
- **Message filtering** for efficient bandwidth usage
- **Auto-reconnection** for reliability
- **Multiple client support** with individual subscriptions
- **Embedded dashboard** for immediate visualization

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Web Client    │────▶│ WebSocket Hub   │◀────│ Monitoring      │
│   (Browser)     │     │                 │     │ Server          │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                        │
         │                       ▼                        ▼
         │              ┌─────────────────┐     ┌─────────────────┐
         └──────────────│ WebSocket       │     │ Metrics         │
                        │ Manager         │     │ Collector       │
                        └─────────────────┘     └─────────────────┘
```

## WebSocket Endpoint

- **URL**: `ws://localhost:8080/ws` (or `wss://` for HTTPS)
- **Protocol**: Standard WebSocket (RFC 6455)
- **Message Format**: JSON

## Message Types

### Client to Server Messages

#### 1. Subscribe
Subscribe to specific message types:
```json
{
  "type": "subscribe",
  "filter": {
    "types": ["system_status", "metrics_update"],
    "task_ids": ["task-123", "task-456"],
    "severity": ["critical", "high"],
    "components": ["api", "worker"]
  }
}
```

#### 2. Unsubscribe
Clear all filters and receive all messages:
```json
{
  "type": "unsubscribe"
}
```

#### 3. Pong
Response to server ping:
```json
{
  "type": "pong"
}
```

### Server to Client Messages

#### 1. Initial State
Sent immediately after connection:
```json
{
  "id": "msg-123",
  "type": "initial_state",
  "timestamp": "2024-01-10T10:00:00Z",
  "data": {
    "connected_at": "2024-01-10T10:00:00Z",
    "server_time": "2024-01-10T10:00:00Z",
    "version": "1.0.0"
  }
}
```

#### 2. System Status
System health and status updates:
```json
{
  "id": "msg-124",
  "type": "system_status",
  "timestamp": "2024-01-10T10:00:05Z",
  "data": {
    "status": "healthy",
    "uptime": "2h15m30s",
    "active_tasks": 5,
    "total_requests": 1024,
    "memory_usage": 134217728,
    "cpu_usage": 45.5
  }
}
```

#### 3. Metrics Update
Detailed metrics information:
```json
{
  "id": "msg-125",
  "type": "metrics_update",
  "timestamp": "2024-01-10T10:00:10Z",
  "data": {
    "active_tasks": 5,
    "memory_usage_mb": 128.5,
    "cpu_usage": 45.5,
    "total_requests": 1024,
    "request_rate": 10.5,
    "error_rate": 0.1
  }
}
```

#### 4. Task Update
Task progress and status updates:
```json
{
  "id": "msg-126",
  "type": "task_update",
  "timestamp": "2024-01-10T10:00:15Z",
  "data": {
    "task_id": "task-123",
    "name": "Bulk Clone Repositories",
    "status": "running",
    "progress": 75,
    "current": 150,
    "total": 200,
    "message": "Cloning repository 150 of 200"
  }
}
```

#### 5. Alert
Real-time alert notifications:
```json
{
  "id": "msg-127",
  "type": "alert",
  "timestamp": "2024-01-10T10:00:20Z",
  "data": {
    "id": "alert-456",
    "name": "High Memory Usage",
    "description": "Memory usage exceeded 90% threshold",
    "severity": "high",
    "status": "active",
    "created_at": "2024-01-10T10:00:20Z"
  }
}
```

#### 6. Ping
Keep-alive message:
```json
{
  "id": "msg-128",
  "type": "ping",
  "timestamp": "2024-01-10T10:00:30Z"
}
```

## Usage Examples

### JavaScript Client Example

```javascript
class MonitoringWebSocket {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 10;
    this.reconnectDelay = 1000;
  }

  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      
      // Subscribe to specific message types
      this.subscribe({
        types: ['system_status', 'metrics_update', 'alert'],
        severity: ['critical', 'high']
      });
    };
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };
    
    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.reconnect();
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }
  
  reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      console.log(`Reconnecting in ${delay}ms...`);
      setTimeout(() => this.connect(), delay);
    }
  }
  
  subscribe(filter) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'subscribe',
        filter: filter
      }));
    }
  }
  
  handleMessage(message) {
    switch (message.type) {
      case 'system_status':
        this.updateSystemStatus(message.data);
        break;
      case 'metrics_update':
        this.updateMetrics(message.data);
        break;
      case 'task_update':
        this.updateTask(message.data);
        break;
      case 'alert':
        this.handleAlert(message.data);
        break;
      case 'ping':
        this.ws.send(JSON.stringify({ type: 'pong' }));
        break;
    }
  }
  
  updateSystemStatus(status) {
    console.log('System status:', status);
    // Update UI with system status
  }
  
  updateMetrics(metrics) {
    console.log('Metrics update:', metrics);
    // Update charts and gauges
  }
  
  updateTask(task) {
    console.log('Task update:', task);
    // Update task progress bars
  }
  
  handleAlert(alert) {
    console.log('Alert:', alert);
    // Show notification
  }
}

// Usage
const monitoring = new MonitoringWebSocket('ws://localhost:8080/ws');
monitoring.connect();
```

### Go Client Example

```go
package main

import (
    "encoding/json"
    "log"
    "time"
    
    "github.com/gorilla/websocket"
)

type MonitoringClient struct {
    url    string
    conn   *websocket.Conn
    done   chan struct{}
}

func NewMonitoringClient(url string) *MonitoringClient {
    return &MonitoringClient{
        url:  url,
        done: make(chan struct{}),
    }
}

func (c *MonitoringClient) Connect() error {
    conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
    if err != nil {
        return err
    }
    c.conn = conn
    
    // Subscribe to updates
    subscribe := map[string]interface{}{
        "type": "subscribe",
        "filter": map[string]interface{}{
            "types": []string{"system_status", "alert"},
        },
    }
    
    if err := c.conn.WriteJSON(subscribe); err != nil {
        return err
    }
    
    // Start reading messages
    go c.readMessages()
    
    return nil
}

func (c *MonitoringClient) readMessages() {
    defer close(c.done)
    
    for {
        var msg map[string]interface{}
        err := c.conn.ReadJSON(&msg)
        if err != nil {
            log.Println("Read error:", err)
            return
        }
        
        msgType, _ := msg["type"].(string)
        switch msgType {
        case "system_status":
            c.handleSystemStatus(msg["data"])
        case "alert":
            c.handleAlert(msg["data"])
        case "ping":
            c.conn.WriteJSON(map[string]string{"type": "pong"})
        }
    }
}

func (c *MonitoringClient) Close() {
    c.conn.Close()
    <-c.done
}
```

## Dashboard Integration

The monitoring server includes an embedded dashboard at `http://localhost:8080/dashboard` that automatically connects to the WebSocket endpoint and displays real-time updates.

### Features:
- **Connection Status**: Visual indicator showing WebSocket connection state
- **System Metrics**: Real-time CPU, memory, and task counters
- **Auto-reconnect**: Automatic reconnection on connection loss
- **Responsive Design**: Works on desktop and mobile devices

## Configuration

### Server Configuration

```go
// Start monitoring server with WebSocket support
server := NewMonitoringServer(&ServerConfig{
    Host:  "localhost",
    Port:  8080,
    Debug: true,
})

// The WebSocket manager is automatically initialized
// Updates are sent every 5 seconds by default
```

### Client Configuration Options

```javascript
const config = {
  reconnect: true,
  reconnectDelay: 1000,      // Initial delay in ms
  maxReconnectDelay: 30000,  // Maximum delay in ms
  reconnectAttempts: 10,     // Maximum attempts
  heartbeatInterval: 30000,  // Ping interval in ms
};
```

## Performance Considerations

### Message Filtering
Use filters to reduce bandwidth and processing:
- Subscribe only to needed message types
- Filter by task IDs when monitoring specific tasks
- Filter alerts by severity

### Connection Management
- Clients automatically reconnect on disconnection
- Exponential backoff prevents connection flooding
- Ping/pong mechanism detects stale connections

### Resource Usage
- Each client connection uses ~10KB memory
- Messages are buffered (256 messages per client)
- Broadcast operations are optimized for many clients

## Security Considerations

### Authentication
Currently, WebSocket connections are not authenticated. In production:
- Implement token-based authentication
- Validate origin headers
- Use WSS (WebSocket Secure) over HTTPS

### Example with Authentication:
```javascript
const ws = new WebSocket('wss://example.com/ws', {
  headers: {
    'Authorization': 'Bearer ' + token
  }
});
```

## Troubleshooting

### Connection Issues
1. **Check server is running**: `gz monitoring server`
2. **Verify WebSocket endpoint**: `ws://localhost:8080/ws`
3. **Check browser console** for errors
4. **Verify firewall** allows WebSocket connections

### Message Not Received
1. **Check subscription filter**: Ensure you're subscribed to the message type
2. **Verify message format**: Check server logs for broadcast errors
3. **Monitor network**: Use browser DevTools to inspect WebSocket frames

### High CPU/Memory Usage
1. **Reduce update frequency**: Adjust periodic update interval
2. **Optimize filters**: Subscribe only to needed messages
3. **Limit concurrent clients**: Configure maximum connections

## API Reference

### WebSocketManager Methods

```go
// Start the WebSocket manager
Start()

// Stop the WebSocket manager
Stop()

// Handle WebSocket upgrade requests
HandleWebSocket(w http.ResponseWriter, r *http.Request)

// Broadcast messages
BroadcastSystemStatus(status interface{})
BroadcastMetrics(metrics interface{})
BroadcastTaskUpdate(taskUpdate interface{})
BroadcastAlert(alert interface{})

// Get statistics
GetConnectedClients() int
GetClientStats() map[string]interface{}
```

## Best Practices

1. **Use Message Filtering**: Subscribe only to needed message types
2. **Handle Reconnection**: Implement automatic reconnection logic
3. **Process Messages Asynchronously**: Don't block the message handler
4. **Monitor Connection Health**: Implement ping/pong handling
5. **Graceful Shutdown**: Clean up connections properly
6. **Error Handling**: Log and handle all WebSocket errors

## Future Enhancements

Planned improvements for the WebSocket system:
1. **Authentication & Authorization**: Secure WebSocket connections
2. **Message History**: Replay recent messages on reconnection
3. **Compression**: Reduce bandwidth with message compression
4. **Clustering**: Support multiple server instances
5. **Rate Limiting**: Prevent message flooding
6. **Custom Events**: User-defined event types and handlers