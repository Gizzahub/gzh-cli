import React, { createContext, useContext, useEffect, useState, useRef } from 'react';
import { useAuth } from './AuthContext';

const WebSocketContext = createContext();

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

export const WebSocketProvider = ({ children }) => {
  const { token, user } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [systemStatus, setSystemStatus] = useState(null);
  const [metrics, setMetrics] = useState({});
  const [tasks, setTasks] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [logs, setLogs] = useState([]);
  
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;

  const addLog = (level, message) => {
    const logEntry = {
      id: Date.now(),
      timestamp: new Date(),
      level,
      message,
    };
    setLogs(prev => [...prev.slice(-99), logEntry]); // Keep last 100 logs
  };

  const connectWebSocket = () => {
    if (!token || !user) {
      return;
    }

    // Cleanup existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?token=${token}`;

    addLog('info', `Connecting to WebSocket...`);

    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onopen = () => {
      setIsConnected(true);
      reconnectAttempts.current = 0;
      addLog('success', 'WebSocket connected successfully');

      // Subscribe to all updates
      wsRef.current.send(JSON.stringify({
        type: 'subscribe',
        filter: {
          types: ['all']
        }
      }));
    };

    wsRef.current.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleMessage(data);
      } catch (error) {
        addLog('error', `Failed to parse message: ${error.message}`);
      }
    };

    wsRef.current.onclose = () => {
      setIsConnected(false);
      addLog('warning', 'WebSocket connection closed');

      // Attempt to reconnect if we haven't exceeded max attempts
      if (reconnectAttempts.current < maxReconnectAttempts) {
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000);
        addLog('info', `Reconnecting in ${delay / 1000} seconds...`);
        
        reconnectTimeoutRef.current = setTimeout(() => {
          reconnectAttempts.current++;
          connectWebSocket();
        }, delay);
      } else {
        addLog('error', 'Max reconnection attempts reached');
      }
    };

    wsRef.current.onerror = (error) => {
      addLog('error', 'WebSocket error occurred');
      console.error('WebSocket error:', error);
    };
  };

  const handleMessage = (data) => {
    switch (data.type) {
      case 'initial_state':
        if (data.data.system_status) {
          setSystemStatus(data.data.system_status);
        }
        if (data.data.metrics_summary) {
          setMetrics(data.data.metrics_summary);
        }
        if (data.data.active_tasks) {
          setTasks(data.data.active_tasks);
        }
        addLog('info', 'Received initial state');
        break;
        
      case 'system_status':
        setSystemStatus(data.data);
        break;
        
      case 'metrics_update':
        setMetrics(prev => ({ ...prev, ...data.data }));
        break;
        
      case 'task_update':
        setTasks(prev => {
          const existing = prev.find(t => t.id === data.data.id);
          if (existing) {
            return prev.map(t => t.id === data.data.id ? { ...t, ...data.data } : t);
          } else {
            return [...prev, data.data];
          }
        });
        addLog('info', `Task ${data.data.name} updated: ${data.data.status}`);
        break;
        
      case 'alert':
        setAlerts(prev => [data.data, ...prev.slice(0, 49)]); // Keep last 50 alerts
        addLog(data.data.severity === 'critical' ? 'error' : 'warning', 
               `Alert: ${data.data.message}`);
        break;
        
      case 'pong':
        addLog('info', 'Received pong response');
        break;
        
      default:
        addLog('warning', `Unknown message type: ${data.type}`);
    }
  };

  const sendMessage = (message) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
      return true;
    }
    addLog('error', 'Cannot send message: WebSocket not connected');
    return false;
  };

  const ping = () => {
    return sendMessage({ type: 'ping' });
  };

  const subscribe = (filter) => {
    return sendMessage({
      type: 'subscribe',
      filter
    });
  };

  const clearLogs = () => {
    setLogs([]);
    addLog('info', 'Logs cleared');
  };

  useEffect(() => {
    if (token && user) {
      connectWebSocket();
    }

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [token, user]);

  // Send ping every 30 seconds to keep connection alive
  useEffect(() => {
    const pingInterval = setInterval(() => {
      if (isConnected) {
        ping();
      }
    }, 30000);

    return () => clearInterval(pingInterval);
  }, [isConnected]);

  const value = {
    isConnected,
    systemStatus,
    metrics,
    tasks,
    alerts,
    logs,
    sendMessage,
    ping,
    subscribe,
    clearLogs,
    reconnect: connectWebSocket,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
};