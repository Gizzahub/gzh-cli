import React, { useState } from 'react';
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  IconButton,
  Chip,
  LinearProgress,
  Paper,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Divider,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
} from '@mui/material';
import {
  Memory as MemoryIcon,
  Speed as SpeedIcon,
  Assignment as TaskIcon,
  Warning as WarningIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  PlayArrow as PlayIcon,
  Stop as StopIcon,
  Refresh as RefreshIcon,
  Clear as ClearIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
} from '@mui/icons-material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts';
import { useWebSocket } from '../contexts/WebSocketContext';
import { formatDistanceToNow } from 'date-fns';

const Dashboard = () => {
  const {
    systemStatus,
    metrics,
    tasks,
    alerts,
    logs,
    clearLogs,
    isConnected,
  } = useWebSocket();
  
  const [logsExpanded, setLogsExpanded] = useState(false);
  const [metricsHistory, setMetricsHistory] = useState([]);

  // Mock data for charts (replace with real data from WebSocket)
  const chartData = [
    { time: '10:00', cpu: 20, memory: 45, network: 12 },
    { time: '10:05', cpu: 35, memory: 52, network: 18 },
    { time: '10:10', cpu: 28, memory: 48, network: 22 },
    { time: '10:15', cpu: 42, memory: 61, network: 15 },
    { time: '10:20', cpu: 38, memory: 55, network: 28 },
    { time: '10:25', cpu: 45, memory: 63, network: 32 },
  ];

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const getStatusColor = (status) => {
    switch (status?.toLowerCase()) {
      case 'healthy':
        return 'success';
      case 'warning':
        return 'warning';
      case 'critical':
        return 'error';
      default:
        return 'default';
    }
  };

  const getTaskStatusIcon = (status) => {
    switch (status) {
      case 'running':
        return <PlayIcon color="primary" />;
      case 'completed':
        return <CheckCircleIcon color="success" />;
      case 'failed':
        return <ErrorIcon color="error" />;
      default:
        return <TaskIcon color="action" />;
    }
  };

  const getAlertSeverityColor = (severity) => {
    switch (severity) {
      case 'critical':
        return 'error';
      case 'high':
        return 'warning';
      case 'medium':
        return 'info';
      case 'low':
        return 'success';
      default:
        return 'default';
    }
  };

  const getLogLevelColor = (level) => {
    switch (level) {
      case 'error':
        return '#f44336';
      case 'warning':
        return '#ff9800';
      case 'success':
        return '#4caf50';
      case 'info':
      default:
        return '#2196f3';
    }
  };

  return (
    <Box>
      {/* Header */}
      <Box mb={3}>
        <Typography variant="h4" gutterBottom>
          System Dashboard
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Real-time monitoring and system overview
        </Typography>
      </Box>

      {/* Status Cards */}
      <Grid container spacing={3} mb={3}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="text.secondary" gutterBottom>
                    System Status
                  </Typography>
                  <Chip
                    label={systemStatus?.status || 'Unknown'}
                    color={getStatusColor(systemStatus?.status)}
                    size="small"
                  />
                  <Typography variant="caption" display="block" mt={1}>
                    Uptime: {systemStatus?.uptime || 'N/A'}
                  </Typography>
                </Box>
                <CheckCircleIcon color={getStatusColor(systemStatus?.status)} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="text.secondary" gutterBottom>
                    Active Tasks
                  </Typography>
                  <Typography variant="h5">
                    {systemStatus?.active_tasks || 0}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    Running processes
                  </Typography>
                </Box>
                <TaskIcon color="primary" />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="text.secondary" gutterBottom>
                    Memory Usage
                  </Typography>
                  <Typography variant="h5">
                    {systemStatus ? formatBytes(systemStatus.memory_usage) : 'N/A'}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    System memory
                  </Typography>
                </Box>
                <MemoryIcon color="secondary" />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="text.secondary" gutterBottom>
                    CPU Usage
                  </Typography>
                  <Typography variant="h5">
                    {systemStatus?.cpu_usage?.toFixed(1) || '0'}%
                  </Typography>
                  <LinearProgress
                    variant="determinate"
                    value={systemStatus?.cpu_usage || 0}
                    sx={{ mt: 1 }}
                  />
                </Box>
                <SpeedIcon color="info" />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Charts Row */}
      <Grid container spacing={3} mb={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                System Metrics
              </Typography>
              <Box height={300}>
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Line type="monotone" dataKey="cpu" stroke="#8884d8" name="CPU %" />
                    <Line type="monotone" dataKey="memory" stroke="#82ca9d" name="Memory %" />
                  </LineChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Network I/O
              </Typography>
              <Box height={300}>
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Area type="monotone" dataKey="network" stroke="#ffc658" fill="#ffc658" name="Network KB/s" />
                  </AreaChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Tasks and Alerts Row */}
      <Grid container spacing={3} mb={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                <Typography variant="h6">
                  Active Tasks ({tasks.length})
                </Typography>
                <IconButton size="small">
                  <RefreshIcon />
                </IconButton>
              </Box>
              
              <Box maxHeight={300} overflow="auto">
                {tasks.length > 0 ? (
                  <List dense>
                    {tasks.slice(0, 5).map((task, index) => (
                      <React.Fragment key={task.id || index}>
                        <ListItem
                          secondaryAction={
                            <Chip
                              label={task.status}
                              size="small"
                              color={task.status === 'running' ? 'primary' : 'default'}
                            />
                          }
                        >
                          <ListItemIcon>
                            {getTaskStatusIcon(task.status)}
                          </ListItemIcon>
                          <ListItemText
                            primary={task.name || `Task ${index + 1}`}
                            secondary={
                              <Box>
                                <LinearProgress
                                  variant="determinate"
                                  value={task.progress || 0}
                                  sx={{ mt: 0.5, mb: 0.5 }}
                                />
                                <Typography variant="caption">
                                  Progress: {task.progress || 0}%
                                </Typography>
                              </Box>
                            }
                          />
                        </ListItem>
                        {index < Math.min(tasks.length, 5) - 1 && <Divider />}
                      </React.Fragment>
                    ))}
                  </List>
                ) : (
                  <Typography color="text.secondary" textAlign="center" py={2}>
                    No active tasks
                  </Typography>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                <Typography variant="h6">
                  Recent Alerts ({alerts.length})
                </Typography>
                <IconButton size="small">
                  <RefreshIcon />
                </IconButton>
              </Box>
              
              <Box maxHeight={300} overflow="auto">
                {alerts.length > 0 ? (
                  <List dense>
                    {alerts.slice(0, 5).map((alert, index) => (
                      <React.Fragment key={alert.id || index}>
                        <ListItem>
                          <ListItemIcon>
                            <WarningIcon color={getAlertSeverityColor(alert.severity)} />
                          </ListItemIcon>
                          <ListItemText
                            primary={alert.rule_name || alert.message || 'Alert'}
                            secondary={
                              <Box>
                                <Chip
                                  label={alert.severity}
                                  size="small"
                                  color={getAlertSeverityColor(alert.severity)}
                                  sx={{ mr: 1 }}
                                />
                                <Typography variant="caption">
                                  {alert.message || 'No description'}
                                </Typography>
                              </Box>
                            }
                          />
                        </ListItem>
                        {index < Math.min(alerts.length, 5) - 1 && <Divider />}
                      </React.Fragment>
                    ))}
                  </List>
                ) : (
                  <Typography color="text.secondary" textAlign="center" py={2}>
                    No recent alerts
                  </Typography>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* WebSocket Logs */}
      <Card>
        <CardContent>
          <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
            <Typography variant="h6">
              WebSocket Activity ({logs.length})
            </Typography>
            <Box>
              <IconButton onClick={clearLogs} size="small">
                <ClearIcon />
              </IconButton>
              <IconButton 
                onClick={() => setLogsExpanded(!logsExpanded)} 
                size="small"
              >
                {logsExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
              </IconButton>
            </Box>
          </Box>

          {!isConnected && (
            <Alert severity="warning" sx={{ mb: 2 }}>
              WebSocket connection is not active. Real-time updates are unavailable.
            </Alert>
          )}

          <Paper
            elevation={0}
            sx={{
              bgcolor: '#263238',
              color: 'white',
              p: 2,
              height: logsExpanded ? 400 : 200,
              overflow: 'auto',
              fontFamily: 'monospace',
              fontSize: '0.875rem',
            }}
          >
            {logs.length > 0 ? (
              logs.slice(-50).map((log) => (
                <Box
                  key={log.id}
                  sx={{
                    mb: 0.5,
                    color: getLogLevelColor(log.level),
                  }}
                >
                  [{log.timestamp.toLocaleTimeString()}] {log.message}
                </Box>
              ))
            ) : (
              <Box color="text.secondary">
                No logs available. WebSocket events will appear here.
              </Box>
            )}
          </Paper>
        </CardContent>
      </Card>
    </Box>
  );
};

export default Dashboard;