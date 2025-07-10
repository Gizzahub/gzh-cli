import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Box } from '@mui/material';
import { useAuth } from './contexts/AuthContext';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import Layout from './components/Layout';
import LoadingSpinner from './components/LoadingSpinner';

function App() {
  const { user, loading } = useAuth();

  if (loading) {
    return <LoadingSpinner />;
  }

  if (!user) {
    return <Login />;
  }

  return (
    <Box sx={{ display: 'flex', height: '100vh' }}>
      <Layout>
        <Routes>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/tasks" element={<div>Tasks Page (Coming Soon)</div>} />
          <Route path="/alerts" element={<div>Alerts Page (Coming Soon)</div>} />
          <Route path="/metrics" element={<div>Metrics Page (Coming Soon)</div>} />
          <Route path="/users" element={<div>Users Page (Coming Soon)</div>} />
          <Route path="/settings" element={<div>Settings Page (Coming Soon)</div>} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </Layout>
    </Box>
  );
}

export default App;