import axios from 'axios';

class ApiService {
  constructor() {
    this.client = axios.create({
      baseURL: '/api/v1',
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor to handle errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Token expired or invalid
          localStorage.removeItem('token');
          window.location.href = '/';
        }
        return Promise.reject(error);
      }
    );
  }

  setAuthToken(token) {
    if (token) {
      this.client.defaults.headers.Authorization = `Bearer ${token}`;
    } else {
      delete this.client.defaults.headers.Authorization;
    }
  }

  // Auth endpoints
  async login(credentials) {
    const response = await axios.post('/auth/login', credentials);
    return response;
  }

  async logout() {
    return this.client.post('/auth/logout');
  }

  async getCurrentUser() {
    return this.client.get('/auth/me');
  }

  // System endpoints
  async getSystemStatus() {
    return this.client.get('/status');
  }

  async getHealth() {
    return this.client.get('/health');
  }

  async getMetrics(format = 'json') {
    return this.client.get('/metrics', { params: { format } });
  }

  // Task endpoints
  async getTasks(limit = 50, offset = 0, status = '') {
    return this.client.get('/tasks', {
      params: { limit, offset, status }
    });
  }

  async getTask(id) {
    return this.client.get(`/tasks/${id}`);
  }

  async stopTask(id) {
    return this.client.post(`/tasks/${id}/stop`);
  }

  // Alert endpoints
  async getAlerts() {
    return this.client.get('/alerts');
  }

  async createAlert(alert) {
    return this.client.post('/alerts', alert);
  }

  async updateAlert(id, alert) {
    return this.client.put(`/alerts/${id}`, alert);
  }

  async deleteAlert(id) {
    return this.client.delete(`/alerts/${id}`);
  }

  // Notification endpoints
  async getNotifications() {
    return this.client.get('/notifications');
  }

  async testNotification(notification) {
    return this.client.post('/notifications/test', notification);
  }

  // Configuration endpoints
  async getConfig() {
    return this.client.get('/config');
  }

  async updateConfig(config) {
    return this.client.put('/config', config);
  }

  // User management endpoints
  async getUsers() {
    return this.client.get('/users');
  }

  async getUser(username) {
    return this.client.get(`/users/${username}`);
  }

  async createUser(user) {
    return this.client.post('/users', user);
  }

  async updateUserPassword(username, password) {
    return this.client.put(`/users/${username}/password`, { password });
  }

  async deleteUser(username) {
    return this.client.delete(`/users/${username}`);
  }

  // Generic methods
  async get(url, config = {}) {
    return this.client.get(url, config);
  }

  async post(url, data = {}, config = {}) {
    return this.client.post(url, data, config);
  }

  async put(url, data = {}, config = {}) {
    return this.client.put(url, data, config);
  }

  async delete(url, config = {}) {
    return this.client.delete(url, config);
  }
}

const api = new ApiService();
export default api;