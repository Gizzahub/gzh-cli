import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Switch,
  TextField,
  Button,
  Alert,
  Snackbar,
  FormControlLabel,
  Grid,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
} from '@mui/material';
import {
  Send as SendIcon,
  Settings as SettingsIcon,
  Tag as SlackIcon,
  Email as EmailIcon,
  Chat as DiscordIcon,
  Business as TeamsIcon,
  PlayArrow as TestIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';

const NotificationSettings = () => {
  const { user } = useAuth();
  const [settings, setSettings] = useState({
    slack: {
      enabled: false,
      webhookUrl: '',
      channel: '#monitoring',
      username: 'GZH Monitoring',
      iconEmoji: ':robot_face:',
    },
    email: {
      enabled: false,
      smtpHost: '',
      smtpPort: 587,
      username: '',
      password: '',
      from: '',
      recipients: [],
    },
    discord: {
      enabled: false,
      webhookUrl: '',
      username: 'GZH Monitoring',
    },
    teams: {
      enabled: false,
      webhookUrl: '',
    },
  });

  const [testDialogOpen, setTestDialogOpen] = useState(false);
  const [testType, setTestType] = useState('');
  const [testMessage, setTestMessage] = useState('Test notification from GZH Monitoring');
  const [loading, setLoading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'info' });
  const [newRecipient, setNewRecipient] = useState('');

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      // In a real implementation, this would load from the server
      const savedSettings = localStorage.getItem('notificationSettings');
      if (savedSettings) {
        setSettings(JSON.parse(savedSettings));
      }
    } catch (error) {
      console.error('Failed to load notification settings:', error);
    }
  };

  const saveSettings = async () => {
    try {
      setLoading(true);
      
      // In a real implementation, this would save to the server
      localStorage.setItem('notificationSettings', JSON.stringify(settings));
      
      showSnackbar('Settings saved successfully', 'success');
    } catch (error) {
      console.error('Failed to save settings:', error);
      showSnackbar('Failed to save settings', 'error');
    } finally {
      setLoading(false);
    }
  };

  const testNotification = async (type, message) => {
    try {
      setLoading(true);
      
      await api.post('/notifications/test', {
        type: type,
        target: '',
        message: message,
      });
      
      showSnackbar(`${type} test notification sent successfully`, 'success');
      setTestDialogOpen(false);
    } catch (error) {
      console.error('Failed to send test notification:', error);
      showSnackbar(
        error.response?.data?.error || 'Failed to send test notification',
        'error'
      );
    } finally {
      setLoading(false);
    }
  };

  const showSnackbar = (message, severity) => {
    setSnackbar({ open: true, message, severity });
  };

  const handleSettingChange = (provider, field, value) => {
    setSettings(prev => ({
      ...prev,
      [provider]: {
        ...prev[provider],
        [field]: value,
      },
    }));
  };

  const addEmailRecipient = () => {
    if (newRecipient && !settings.email.recipients.includes(newRecipient)) {
      handleSettingChange('email', 'recipients', [
        ...settings.email.recipients,
        newRecipient,
      ]);
      setNewRecipient('');
    }
  };

  const removeEmailRecipient = (email) => {
    handleSettingChange('email', 'recipients', 
      settings.email.recipients.filter(r => r !== email)
    );
  };

  const openTestDialog = (type) => {
    setTestType(type);
    setTestDialogOpen(true);
  };

  const renderSlackSettings = () => (
    <Card sx={{ mb: 3 }}>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
          <Box display="flex" alignItems="center">
            <SlackIcon sx={{ mr: 1, color: '#4A154B' }} />
            <Typography variant="h6">Slack Integration</Typography>
          </Box>
          <Box>
            <IconButton onClick={() => openTestDialog('slack')} disabled={!settings.slack.enabled}>
              <TestIcon />
            </IconButton>
            <FormControlLabel
              control={
                <Switch
                  checked={settings.slack.enabled}
                  onChange={(e) => handleSettingChange('slack', 'enabled', e.target.checked)}
                />
              }
              label="Enabled"
            />
          </Box>
        </Box>

        <Grid container spacing={2}>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label="Webhook URL"
              value={settings.slack.webhookUrl}
              onChange={(e) => handleSettingChange('slack', 'webhookUrl', e.target.value)}
              type="password"
              placeholder="https://hooks.slack.com/services/..."
              disabled={!settings.slack.enabled}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Channel"
              value={settings.slack.channel}
              onChange={(e) => handleSettingChange('slack', 'channel', e.target.value)}
              placeholder="#monitoring"
              disabled={!settings.slack.enabled}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Username"
              value={settings.slack.username}
              onChange={(e) => handleSettingChange('slack', 'username', e.target.value)}
              placeholder="GZH Monitoring"
              disabled={!settings.slack.enabled}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Icon Emoji"
              value={settings.slack.iconEmoji}
              onChange={(e) => handleSettingChange('slack', 'iconEmoji', e.target.value)}
              placeholder=":robot_face:"
              disabled={!settings.slack.enabled}
            />
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );

  const renderEmailSettings = () => (
    <Card sx={{ mb: 3 }}>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
          <Box display="flex" alignItems="center">
            <EmailIcon sx={{ mr: 1, color: '#1976d2' }} />
            <Typography variant="h6">Email Notifications</Typography>
          </Box>
          <Box>
            <IconButton onClick={() => openTestDialog('email')} disabled={!settings.email.enabled}>
              <TestIcon />
            </IconButton>
            <FormControlLabel
              control={
                <Switch
                  checked={settings.email.enabled}
                  onChange={(e) => handleSettingChange('email', 'enabled', e.target.checked)}
                />
              }
              label="Enabled"
            />
          </Box>
        </Box>

        <Grid container spacing={2}>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="SMTP Host"
              value={settings.email.smtpHost}
              onChange={(e) => handleSettingChange('email', 'smtpHost', e.target.value)}
              placeholder="smtp.gmail.com"
              disabled={!settings.email.enabled}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="SMTP Port"
              type="number"
              value={settings.email.smtpPort}
              onChange={(e) => handleSettingChange('email', 'smtpPort', parseInt(e.target.value))}
              disabled={!settings.email.enabled}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Username"
              value={settings.email.username}
              onChange={(e) => handleSettingChange('email', 'username', e.target.value)}
              disabled={!settings.email.enabled}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Password"
              type="password"
              value={settings.email.password}
              onChange={(e) => handleSettingChange('email', 'password', e.target.value)}
              disabled={!settings.email.enabled}
            />
          </Grid>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label="From Address"
              value={settings.email.from}
              onChange={(e) => handleSettingChange('email', 'from', e.target.value)}
              placeholder="monitoring@company.com"
              disabled={!settings.email.enabled}
            />
          </Grid>
          <Grid item xs={12}>
            <Typography variant="subtitle2" gutterBottom>
              Recipients
            </Typography>
            <Box display="flex" gap={1} mb={1}>
              <TextField
                size="small"
                placeholder="Add email recipient"
                value={newRecipient}
                onChange={(e) => setNewRecipient(e.target.value)}
                disabled={!settings.email.enabled}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    addEmailRecipient();
                  }
                }}
              />
              <Button
                variant="outlined"
                onClick={addEmailRecipient}
                disabled={!settings.email.enabled || !newRecipient}
              >
                <AddIcon />
              </Button>
            </Box>
            <Box display="flex" flexWrap="wrap" gap={1}>
              {settings.email.recipients.map((email, index) => (
                <Chip
                  key={index}
                  label={email}
                  onDelete={() => removeEmailRecipient(email)}
                  disabled={!settings.email.enabled}
                />
              ))}
            </Box>
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );

  const renderDiscordSettings = () => (
    <Card sx={{ mb: 3 }}>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
          <Box display="flex" alignItems="center">
            <DiscordIcon sx={{ mr: 1, color: '#5865F2' }} />
            <Typography variant="h6">Discord Integration</Typography>
          </Box>
          <Box>
            <IconButton onClick={() => openTestDialog('discord')} disabled={!settings.discord.enabled}>
              <TestIcon />
            </IconButton>
            <FormControlLabel
              control={
                <Switch
                  checked={settings.discord.enabled}
                  onChange={(e) => handleSettingChange('discord', 'enabled', e.target.checked)}
                />
              }
              label="Enabled"
            />
          </Box>
        </Box>

        <Grid container spacing={2}>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label="Webhook URL"
              value={settings.discord.webhookUrl}
              onChange={(e) => handleSettingChange('discord', 'webhookUrl', e.target.value)}
              type="password"
              placeholder="https://discord.com/api/webhooks/..."
              disabled={!settings.discord.enabled}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Username"
              value={settings.discord.username}
              onChange={(e) => handleSettingChange('discord', 'username', e.target.value)}
              placeholder="GZH Monitoring"
              disabled={!settings.discord.enabled}
            />
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );

  if (user?.role !== 'admin') {
    return (
      <Alert severity="warning">
        You need administrator privileges to access notification settings.
      </Alert>
    );
  }

  return (
    <Box>
      <Box mb={3} display="flex" justifyContent="space-between" alignItems="center">
        <Box>
          <Typography variant="h4" gutterBottom>
            Notification Settings
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Configure alert notifications for Slack, email, Discord, and Microsoft Teams
          </Typography>
        </Box>
        <Button
          variant="contained"
          onClick={saveSettings}
          disabled={loading}
          startIcon={<SettingsIcon />}
        >
          Save Settings
        </Button>
      </Box>

      {renderSlackSettings()}
      {renderEmailSettings()}
      {renderDiscordSettings()}

      {/* Test Notification Dialog */}
      <Dialog open={testDialogOpen} onClose={() => setTestDialogOpen(false)}>
        <DialogTitle>Test {testType} Notification</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Test Message"
            fullWidth
            multiline
            rows={3}
            variant="outlined"
            value={testMessage}
            onChange={(e) => setTestMessage(e.target.value)}
            sx={{ mt: 1 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setTestDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={() => testNotification(testType, testMessage)}
            variant="contained"
            disabled={loading}
            startIcon={<SendIcon />}
          >
            Send Test
          </Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
      >
        <Alert
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity}
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default NotificationSettings;