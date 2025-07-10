package monitoring

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	// JWT related constants
	JWTSecretEnvVar    = "GZH_JWT_SECRET"
	JWTExpirationHours = 24
	AdminRole          = "admin"
	ViewerRole         = "viewer"
	OperatorRole       = "operator"
)

// User represents a system user
type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Active   bool      `json:"active"`
	LastSeen time.Time `json:"last_seen"`
	Created  time.Time `json:"created"`
}

// Credentials represents login credentials
type Credentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthManager manages authentication and authorization
type AuthManager struct {
	jwtSecret []byte
	users     map[string]*User  // In-memory store (replace with database in production)
	passwords map[string]string // username -> hashed password
	logger    *zap.Logger
}

// Permission levels for different operations
var rolePermissions = map[string][]string{
	AdminRole: {
		"read:system", "write:system", "read:tasks", "write:tasks",
		"read:alerts", "write:alerts", "read:config", "write:config",
		"read:users", "write:users", "delete:users",
	},
	OperatorRole: {
		"read:system", "read:tasks", "write:tasks",
		"read:alerts", "write:alerts", "read:config",
	},
	ViewerRole: {
		"read:system", "read:tasks", "read:alerts", "read:config",
	},
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(logger *zap.Logger) *AuthManager {
	// Generate or load JWT secret
	jwtSecret := generateJWTSecret()

	auth := &AuthManager{
		jwtSecret: jwtSecret,
		users:     make(map[string]*User),
		passwords: make(map[string]string),
		logger:    logger,
	}

	// Initialize with default admin user
	auth.initializeDefaultUsers()

	return auth
}

// generateJWTSecret creates a secure JWT secret
func generateJWTSecret() []byte {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		// Fallback to a default secret (not recommended for production)
		return []byte("gzh-manager-default-secret-change-me")
	}
	return secret
}

// initializeDefaultUsers creates default system users
func (a *AuthManager) initializeDefaultUsers() {
	// Create default admin user
	adminPassword := "admin123" // Should be changed on first login
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	admin := &User{
		ID:       "admin-001",
		Username: "admin",
		Email:    "admin@localhost",
		Role:     AdminRole,
		Active:   true,
		Created:  time.Now(),
		LastSeen: time.Now(),
	}

	a.users[admin.Username] = admin
	a.passwords[admin.Username] = string(hashedPassword)

	// Create default viewer user
	viewerPassword := "viewer123"
	hashedViewerPassword, _ := bcrypt.GenerateFromPassword([]byte(viewerPassword), bcrypt.DefaultCost)

	viewer := &User{
		ID:       "viewer-001",
		Username: "viewer",
		Email:    "viewer@localhost",
		Role:     ViewerRole,
		Active:   true,
		Created:  time.Now(),
		LastSeen: time.Now(),
	}

	a.users[viewer.Username] = viewer
	a.passwords[viewer.Username] = string(hashedViewerPassword)

	a.logger.Info("Default users initialized",
		zap.String("admin_username", "admin"),
		zap.String("viewer_username", "viewer"),
		zap.String("note", "Change default passwords immediately"))
}

// Authenticate validates user credentials
func (a *AuthManager) Authenticate(username, password string) (*User, string, error) {
	user, exists := a.users[username]
	if !exists || !user.Active {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	hashedPassword, exists := a.passwords[username]
	if !exists {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := a.generateJWTToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last seen
	user.LastSeen = time.Now()

	return user, token, nil
}

// generateJWTToken creates a JWT token for the user
func (a *AuthManager) generateJWTToken(user *User) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWTExpirationHours * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gzh-manager",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ValidateJWTToken validates and parses a JWT token
func (a *AuthManager) ValidateJWTToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HasPermission checks if a user has the required permission
func (a *AuthManager) HasPermission(userRole, permission string) bool {
	permissions, exists := rolePermissions[userRole]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// CreateUser creates a new user
func (a *AuthManager) CreateUser(username, email, password, role string) (*User, error) {
	if _, exists := a.users[username]; exists {
		return nil, fmt.Errorf("user already exists")
	}

	if role != AdminRole && role != OperatorRole && role != ViewerRole {
		return nil, fmt.Errorf("invalid role")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		ID:       generateUserID(),
		Username: username,
		Email:    email,
		Role:     role,
		Active:   true,
		Created:  time.Now(),
		LastSeen: time.Time{},
	}

	a.users[username] = user
	a.passwords[username] = string(hashedPassword)

	a.logger.Info("User created",
		zap.String("username", username),
		zap.String("role", role))

	return user, nil
}

// UpdateUserPassword updates a user's password
func (a *AuthManager) UpdateUserPassword(username, newPassword string) error {
	if _, exists := a.users[username]; !exists {
		return fmt.Errorf("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	a.passwords[username] = string(hashedPassword)

	a.logger.Info("User password updated",
		zap.String("username", username))

	return nil
}

// DeactivateUser deactivates a user account
func (a *AuthManager) DeactivateUser(username string) error {
	user, exists := a.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.Active = false

	a.logger.Info("User deactivated",
		zap.String("username", username))

	return nil
}

// GetUsers returns all users
func (a *AuthManager) GetUsers() []*User {
	users := make([]*User, 0, len(a.users))
	for _, user := range a.users {
		users = append(users, user)
	}
	return users
}

// GetUser returns a specific user
func (a *AuthManager) GetUser(username string) (*User, error) {
	user, exists := a.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// generateUserID generates a unique user ID
func generateUserID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "user-" + hex.EncodeToString(bytes)
}

// Middleware functions

// JWTAuthMiddleware validates JWT tokens for API endpoints
func (a *AuthManager) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := a.ValidateJWTToken(tokenParts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// RequirePermission middleware checks if user has required permission
func (a *AuthManager) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "user role not found"})
			c.Abort()
			return
		}

		if !a.HasPermission(userRole.(string), permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens but doesn't require them
func (a *AuthManager) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := a.ValidateJWTToken(tokenParts[1])
		if err != nil {
			c.Next()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// WebSocket authentication middleware
func (a *AuthManager) AuthenticateWebSocket(r *http.Request) (*User, error) {
	// Check for token in query parameter or header
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token = tokenParts[1]
			}
		}
	}

	if token == "" {
		return nil, fmt.Errorf("no authentication token provided")
	}

	claims, err := a.ValidateJWTToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, exists := a.users[claims.Username]
	if !exists || !user.Active {
		return nil, fmt.Errorf("user not found or inactive")
	}

	return user, nil
}
