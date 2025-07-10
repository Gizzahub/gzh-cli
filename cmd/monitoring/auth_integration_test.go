package monitoring

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthenticationIntegration tests the full authentication flow
func TestAuthenticationIntegration(t *testing.T) {
	server := NewMonitoringServer(&ServerConfig{
		Host:  "localhost",
		Port:  0,
		Debug: true,
	})

	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	t.Run("Login and access protected resource", func(t *testing.T) {
		// Step 1: Login
		credentials := Credentials{
			Username: "admin",
			Password: "admin123",
		}
		jsonData, _ := json.Marshal(credentials)

		loginReq, _ := http.NewRequest("POST", testServer.URL+"/auth/login", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := http.DefaultClient.Do(loginReq)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		assert.Equal(t, http.StatusOK, loginResp.StatusCode)

		var loginResponse map[string]interface{}
		err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		require.NoError(t, err)

		token := loginResponse["token"].(string)
		assert.NotEmpty(t, token)

		// Step 2: Access protected resource
		statusReq, _ := http.NewRequest("GET", testServer.URL+"/api/v1/status", nil)
		statusReq.Header.Set("Authorization", "Bearer "+token)

		statusResp, err := http.DefaultClient.Do(statusReq)
		require.NoError(t, err)
		defer statusResp.Body.Close()

		assert.Equal(t, http.StatusOK, statusResp.StatusCode)

		var status SystemStatus
		err = json.NewDecoder(statusResp.Body).Decode(&status)
		require.NoError(t, err)

		assert.Equal(t, "healthy", status.Status)
	})

	t.Run("Access protected resource without token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/status", nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Permission denied for insufficient role", func(t *testing.T) {
		// Login as viewer
		credentials := Credentials{
			Username: "viewer",
			Password: "viewer123",
		}
		jsonData, _ := json.Marshal(credentials)

		loginReq, _ := http.NewRequest("POST", testServer.URL+"/auth/login", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := http.DefaultClient.Do(loginReq)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		var loginResponse map[string]interface{}
		json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		token := loginResponse["token"].(string)

		// Try to access admin-only endpoint
		req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
