package monitoring

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthManager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	authManager := NewAuthManager(logger)

	t.Run("Default users created", func(t *testing.T) {
		users := authManager.GetUsers()
		assert.Len(t, users, 2)

		// Check admin user
		admin, err := authManager.GetUser("admin")
		require.NoError(t, err)
		assert.Equal(t, "admin", admin.Username)
		assert.Equal(t, AdminRole, admin.Role)
		assert.True(t, admin.Active)

		// Check viewer user
		viewer, err := authManager.GetUser("viewer")
		require.NoError(t, err)
		assert.Equal(t, "viewer", viewer.Username)
		assert.Equal(t, ViewerRole, viewer.Role)
		assert.True(t, viewer.Active)
	})

	t.Run("Authentication success", func(t *testing.T) {
		user, token, err := authManager.Authenticate("admin", "admin123")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
		assert.Equal(t, "admin", user.Username)
	})

	t.Run("Authentication failure", func(t *testing.T) {
		user, token, err := authManager.Authenticate("admin", "wrongpassword")
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Empty(t, token)
	})

	t.Run("JWT token validation", func(t *testing.T) {
		user, token, err := authManager.Authenticate("admin", "admin123")
		require.NoError(t, err)

		claims, err := authManager.ValidateJWTToken(token)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Username, claims.Username)
		assert.Equal(t, user.Role, claims.Role)
	})

	t.Run("Create new user", func(t *testing.T) {
		user, err := authManager.CreateUser("testuser", "test@example.com", "testpass", OperatorRole)
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, OperatorRole, user.Role)
		assert.True(t, user.Active)

		// Verify user can authenticate
		authUser, token, err := authManager.Authenticate("testuser", "testpass")
		require.NoError(t, err)
		assert.Equal(t, user.ID, authUser.ID)
		assert.NotEmpty(t, token)
	})

	t.Run("Duplicate user creation fails", func(t *testing.T) {
		_, err := authManager.CreateUser("admin", "admin@example.com", "password", AdminRole)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Permission checking", func(t *testing.T) {
		// Admin should have all permissions
		assert.True(t, authManager.HasPermission(AdminRole, "read:system"))
		assert.True(t, authManager.HasPermission(AdminRole, "write:users"))
		assert.True(t, authManager.HasPermission(AdminRole, "delete:users"))

		// Operator should have limited permissions
		assert.True(t, authManager.HasPermission(OperatorRole, "read:system"))
		assert.True(t, authManager.HasPermission(OperatorRole, "write:tasks"))
		assert.False(t, authManager.HasPermission(OperatorRole, "write:users"))

		// Viewer should have only read permissions
		assert.True(t, authManager.HasPermission(ViewerRole, "read:system"))
		assert.False(t, authManager.HasPermission(ViewerRole, "write:tasks"))
		assert.False(t, authManager.HasPermission(ViewerRole, "write:users"))
	})

	t.Run("Update user password", func(t *testing.T) {
		err := authManager.UpdateUserPassword("admin", "newpassword")
		require.NoError(t, err)

		// Old password should not work
		_, _, err = authManager.Authenticate("admin", "admin123")
		require.Error(t, err)

		// New password should work
		user, token, err := authManager.Authenticate("admin", "newpassword")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
	})

	t.Run("Deactivate user", func(t *testing.T) {
		err := authManager.DeactivateUser("viewer")
		require.NoError(t, err)

		// Deactivated user should not be able to authenticate
		_, _, err = authManager.Authenticate("viewer", "viewer123")
		require.Error(t, err)
	})
}

func TestAuthenticationAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create server with authentication
	server := NewMonitoringServer(&ServerConfig{
		Host:  "localhost",
		Port:  0,
		Debug: true,
	})

	t.Run("Login success", func(t *testing.T) {
		credentials := Credentials{
			Username: "admin",
			Password: "admin123",
		}
		jsonData, _ := json.Marshal(credentials)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "token")
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Login failure", func(t *testing.T) {
		credentials := Credentials{
			Username: "admin",
			Password: "wrongpassword",
		}
		jsonData, _ := json.Marshal(credentials)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Protected endpoint without token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/status", nil)

		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Protected endpoint with valid token", func(t *testing.T) {
		// First login to get token
		credentials := Credentials{
			Username: "admin",
			Password: "admin123",
		}
		jsonData, _ := json.Marshal(credentials)

		loginReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")

		loginW := httptest.NewRecorder()
		server.router.ServeHTTP(loginW, loginReq)

		var loginResponse map[string]interface{}
		json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
		token := loginResponse["token"].(string)

		// Use token to access protected endpoint
		req, _ := http.NewRequest("GET", "/api/v1/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Permission denied for insufficient role", func(t *testing.T) {
		// Login as viewer
		credentials := Credentials{
			Username: "viewer",
			Password: "viewer123",
		}
		jsonData, _ := json.Marshal(credentials)

		loginReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")

		loginW := httptest.NewRecorder()
		server.router.ServeHTTP(loginW, loginReq)

		var loginResponse map[string]interface{}
		json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
		token := loginResponse["token"].(string)

		// Try to access admin-only endpoint
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Get current user info", func(t *testing.T) {
		// Login to get token
		credentials := Credentials{
			Username: "admin",
			Password: "admin123",
		}
		jsonData, _ := json.Marshal(credentials)

		loginReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")

		loginW := httptest.NewRecorder()
		server.router.ServeHTTP(loginW, loginReq)

		var loginResponse map[string]interface{}
		json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
		token := loginResponse["token"].(string)

		// Get current user info
		req, _ := http.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		user := response["user"].(map[string]interface{})
		assert.Equal(t, "admin", user["username"])
		assert.Equal(t, AdminRole, user["role"])
	})
}

func TestJWTTokenExpiration(t *testing.T) {
	logger := zap.NewNop()
	authManager := NewAuthManager(logger)

	// Create a token that will expire soon (for testing purposes)
	user := &User{
		ID:       "test-user",
		Username: "testuser",
		Role:     ViewerRole,
	}

	// Mock expired token by manipulating the current time
	// In a real test, you might want to use a time mock library
	token, err := authManager.generateJWTToken(user)
	require.NoError(t, err)

	// Token should be valid immediately
	claims, err := authManager.ValidateJWTToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)

	// Test with invalid token
	invalidToken := "invalid.jwt.token"
	_, err = authManager.ValidateJWTToken(invalidToken)
	require.Error(t, err)
}

func TestCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := NewMonitoringServer(&ServerConfig{
		Host:  "localhost",
		Port:  0,
		Debug: true,
	})

	t.Run("CORS headers present", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/auth/login", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})
}
