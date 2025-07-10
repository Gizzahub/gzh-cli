import React, { useState, useEffect } from 'react';
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  Chip,
  IconButton,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
  Tooltip,
  LinearProgress,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Computer as ComputerIcon,
  CloudQueue as CloudIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';

const InstanceManager = () => {
  const { user } = useAuth();
  const [instances, setInstances] = useState([]);
  const [clusterStatus, setClusterStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [newInstance, setNewInstance] = useState({ host: '', port: 8080 });

  useEffect(() => {
    fetchInstances();
    fetchClusterStatus();
    
    // Refresh data every 30 seconds
    const interval = setInterval(() => {
      fetchInstances();
      fetchClusterStatus();
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  const fetchInstances = async () => {
    try {
      const response = await api.get('/instances');
      setInstances(response.data.instances || []);
      setError(null);
    } catch (err) {
      setError('Failed to fetch instances');
      console.error('Error fetching instances:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchClusterStatus = async () => {
    try {
      const response = await api.get('/instances/cluster/status');
      setClusterStatus(response.data);
    } catch (err) {
      console.error('Error fetching cluster status:', err);
    }
  };

  const handleRefresh = () => {
    setLoading(true);
    fetchInstances();
    fetchClusterStatus();
  };

  const handleAddInstance = async () => {
    try {
      await api.post('/instances/discover', newInstance);
      setAddDialogOpen(false);
      setNewInstance({ host: '', port: 8080 });
      fetchInstances();
      fetchClusterStatus();
    } catch (err) {
      setError('Failed to add instance: ' + (err.response?.data?.error || err.message));
    }
  };

  const handleRemoveInstance = async (instanceId) => {
    try {
      await api.delete(`/instances/${instanceId}`);
      fetchInstances();
      fetchClusterStatus();
    } catch (err) {
      setError('Failed to remove instance: ' + (err.response?.data?.error || err.message));
    }
  };

  const getStatusColor = (status) => {
    switch (status?.toLowerCase()) {
      case 'running':
        return 'success';
      case 'unhealthy':
        return 'error';
      case 'warning':
        return 'warning';
      default:
        return 'default';
    }
  };

  const getStatusIcon = (status) => {
    switch (status?.toLowerCase()) {
      case 'running':
        return <CheckCircleIcon color="success" />;
      case 'unhealthy':
        return <ErrorIcon color="error" />;
      case 'warning':
        return <WarningIcon color="warning" />;
      default:
        return <ComputerIcon color="action" />;
    }
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const formatUptime = (uptime) => {
    if (!uptime) return 'N/A';
    // Simple uptime parsing - assumes format like "1h2m3s"
    return uptime;
  };

  if (loading && instances.length === 0) {
    return <LinearProgress />;
  }

  return (
    <Box>
      {/* Header */}
      <Box mb={3} display="flex" justifyContent="space-between" alignItems="center">
        <Box>
          <Typography variant="h4" gutterBottom>
            Instance Manager
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Manage and monitor multiple monitoring instances
          </Typography>
        </Box>
        <Box>
          <IconButton onClick={handleRefresh} disabled={loading}>
            <RefreshIcon />
          </IconButton>
          {user?.role === 'admin' && (
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => setAddDialogOpen(true)}
              sx={{ ml: 1 }}
            >
              Add Instance
            </Button>
          )}
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {/* Cluster Status Cards */}
      {clusterStatus && (
        <Grid container spacing={3} mb={3}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="text.secondary" gutterBottom>
                      Total Instances
                    </Typography>
                    <Typography variant="h5">
                      {clusterStatus.total_instances}
                    </Typography>
                  </Box>
                  <CloudIcon color="primary" />
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
                      Running
                    </Typography>
                    <Typography variant="h5" color="success.main">
                      {clusterStatus.running_instances}
                    </Typography>
                  </Box>
                  <CheckCircleIcon color="success" />
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
                      Unhealthy
                    </Typography>
                    <Typography variant="h5" color="error.main">
                      {clusterStatus.unhealthy_instances}
                    </Typography>
                  </Box>
                  <ErrorIcon color="error" />
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
                      Health Rate
                    </Typography>
                    <Typography variant="h5">
                      {clusterStatus.total_instances > 0 
                        ? Math.round((clusterStatus.running_instances / clusterStatus.total_instances) * 100)
                        : 0}%
                    </Typography>
                  </Box>
                  <ComputerIcon color="info" />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {/* Instances Table */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Monitoring Instances
          </Typography>
          
          <TableContainer component={Paper} elevation={0}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Status</TableCell>
                  <TableCell>Instance</TableCell>
                  <TableCell>Type</TableCell>
                  <TableCell>Environment</TableCell>
                  <TableCell>CPU</TableCell>
                  <TableCell>Memory</TableCell>
                  <TableCell>Uptime</TableCell>
                  <TableCell>Last Seen</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {instances.map((instance) => (
                  <TableRow key={instance.id}>
                    <TableCell>
                      <Box display="flex" alignItems="center">
                        {getStatusIcon(instance.status)}
                        <Chip
                          label={instance.status}
                          color={getStatusColor(instance.status)}
                          size="small"
                          sx={{ ml: 1 }}
                        />
                      </Box>
                    </TableCell>
                    
                    <TableCell>
                      <Box>
                        <Typography variant="body2" fontWeight="medium">
                          {instance.name || instance.id}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {instance.host}:{instance.port}
                        </Typography>
                      </Box>
                    </TableCell>
                    
                    <TableCell>
                      <Chip
                        label={instance.tags?.type || 'unknown'}
                        variant="outlined"
                        size="small"
                      />
                    </TableCell>
                    
                    <TableCell>
                      <Chip
                        label={instance.environment || 'unknown'}
                        variant="outlined"
                        size="small"
                      />
                    </TableCell>
                    
                    <TableCell>
                      {instance.metrics?.cpu_usage !== undefined ? (
                        <Box>
                          <Typography variant="body2">
                            {instance.metrics.cpu_usage.toFixed(1)}%
                          </Typography>
                          <LinearProgress
                            variant="determinate"
                            value={instance.metrics.cpu_usage}
                            sx={{ width: 60, height: 4 }}
                          />
                        </Box>
                      ) : (
                        'N/A'
                      )}
                    </TableCell>
                    
                    <TableCell>
                      {instance.metrics?.memory_usage ? (
                        formatBytes(instance.metrics.memory_usage)
                      ) : (
                        'N/A'
                      )}
                    </TableCell>
                    
                    <TableCell>
                      {instance.metrics?.uptime ? formatUptime(instance.metrics.uptime) : 'N/A'}
                    </TableCell>
                    
                    <TableCell>
                      <Tooltip title={new Date(instance.last_seen).toLocaleString()}>
                        <Typography variant="caption">
                          {new Date(instance.last_seen).toLocaleTimeString()}
                        </Typography>
                      </Tooltip>
                    </TableCell>
                    
                    <TableCell align="right">
                      {instance.tags?.type === 'remote' && user?.role === 'admin' && (
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleRemoveInstance(instance.id)}
                        >
                          <DeleteIcon />
                        </IconButton>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
                
                {instances.length === 0 && !loading && (
                  <TableRow>
                    <TableCell colSpan={9} align="center">
                      <Typography color="text.secondary">
                        No instances found
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>

      {/* Add Instance Dialog */}
      <Dialog open={addDialogOpen} onClose={() => setAddDialogOpen(false)}>
        <DialogTitle>Add Monitoring Instance</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 1 }}>
            <TextField
              autoFocus
              margin="dense"
              label="Host"
              fullWidth
              variant="outlined"
              value={newInstance.host}
              onChange={(e) => setNewInstance({ ...newInstance, host: e.target.value })}
              placeholder="localhost"
            />
            <TextField
              margin="dense"
              label="Port"
              type="number"
              fullWidth
              variant="outlined"
              value={newInstance.port}
              onChange={(e) => setNewInstance({ ...newInstance, port: parseInt(e.target.value) })}
              sx={{ mt: 2 }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAddDialogOpen(false)}>Cancel</Button>
          <Button 
            onClick={handleAddInstance}
            variant="contained"
            disabled={!newInstance.host || !newInstance.port}
          >
            Add Instance
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default InstanceManager;