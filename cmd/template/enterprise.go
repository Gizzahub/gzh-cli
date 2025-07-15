package template

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

// EnterpriseCmd represents the enterprise command
var EnterpriseCmd = &cobra.Command{
	Use:   "enterprise",
	Short: "기업용 프라이빗 마켓플레이스 관리",
	Long: `기업용 프라이빗 마켓플레이스 기능을 제공합니다.

엔터프라이즈 기능:
- 역할 기반 접근 제어 (RBAC)
- 승인 워크플로우 시스템
- 감사 로그 및 추적성
- 조직별 템플릿 관리
- SSO(Single Sign-On) 통합
- 컴플라이언스 및 정책 관리
- 사용량 분석 및 보고서
- 백업 및 복원 시스템

Examples:
  gz template enterprise init --org mycompany
  gz template enterprise user add --email user@company.com --role developer
  gz template enterprise approval create --template-id abc123 --reviewer admin
  gz template enterprise audit --from 2024-01-01 --to 2024-12-31`,
	Run: runEnterprise,
}

var (
	orgName          string
	enterpriseConfig string
	adminEmail       string
	ssoProvider      string
	ldapConfig       string
	auditLogDir      string
	backupDir        string
	licenseFile      string
)

func init() {
	EnterpriseCmd.Flags().StringVar(&orgName, "org", "", "조직 이름")
	EnterpriseCmd.Flags().StringVar(&enterpriseConfig, "config", "", "엔터프라이즈 설정 파일")
	EnterpriseCmd.Flags().StringVar(&adminEmail, "admin-email", "", "관리자 이메일")
	EnterpriseCmd.Flags().StringVar(&ssoProvider, "sso-provider", "", "SSO 제공자 (saml, oauth2, ldap)")
	EnterpriseCmd.Flags().StringVar(&ldapConfig, "ldap-config", "", "LDAP 설정 파일")
	EnterpriseCmd.Flags().StringVar(&auditLogDir, "audit-dir", "./audit", "감사 로그 디렉터리")
	EnterpriseCmd.Flags().StringVar(&backupDir, "backup-dir", "./backups", "백업 디렉터리")
	EnterpriseCmd.Flags().StringVar(&licenseFile, "license", "", "엔터프라이즈 라이선스 파일")

	// Add subcommands
	EnterpriseCmd.AddCommand(enterpriseInitCmd)
	EnterpriseCmd.AddCommand(enterpriseUserCmd)
	EnterpriseCmd.AddCommand(enterpriseApprovalCmd)
	EnterpriseCmd.AddCommand(enterpriseAuditCmd)
	EnterpriseCmd.AddCommand(enterprisePolicyCmd)
	EnterpriseCmd.AddCommand(enterpriseBackupCmd)
	EnterpriseCmd.AddCommand(SecurityCmd)
	EnterpriseCmd.AddCommand(ComplianceCmd)
}

// EnterpriseServer extends TemplateServer with enterprise features
type EnterpriseServer struct {
	*TemplateServer
	Organization *Organization
	Users        map[string]*EnterpriseUser
	Roles        map[string]*Role
	Permissions  map[string]*Permission
	Policies     map[string]*Policy
	AuditLogger  *AuditLogger
	SSO          *SSOConfig
	LDAP         *LDAPConfig
	Workflows    map[string]*ApprovalWorkflow
}

// Organization represents enterprise organization
type Organization struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Domain       string            `json:"domain"`
	Settings     map[string]string `json:"settings"`
	Subscription *Subscription     `json:"subscription"`
	Created      time.Time         `json:"created"`
	Updated      time.Time         `json:"updated"`
}

// EnterpriseUser extends UserInfo with enterprise features
type EnterpriseUser struct {
	*UserInfo
	Department   string            `json:"department"`
	Manager      string            `json:"manager"`
	Roles        []string          `json:"roles"`
	Permissions  []string          `json:"permissions"`
	Groups       []string          `json:"groups"`
	LastActivity time.Time         `json:"last_activity"`
	Status       string            `json:"status"` // active, suspended, disabled
	Profile      map[string]string `json:"profile"`
	Settings     map[string]string `json:"settings"`
	SSO          *SSOUserInfo      `json:"sso,omitempty"`
}

// Role represents user role
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// Permission represents specific permission
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
}

// Policy represents compliance policy
type Policy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // security, compliance, governance
	Rules       []PolicyRule      `json:"rules"`
	Enabled     bool              `json:"enabled"`
	Enforcement string            `json:"enforcement"` // warn, block, log
	Metadata    map[string]string `json:"metadata"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
}

// PolicyRule represents policy rule
type PolicyRule struct {
	ID         string            `json:"id"`
	Condition  string            `json:"condition"`
	Action     string            `json:"action"`
	Parameters map[string]string `json:"parameters"`
	Message    string            `json:"message"`
}

// ApprovalWorkflow represents approval workflow
type ApprovalWorkflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Trigger     string         `json:"trigger"` // template_upload, user_create, etc.
	Steps       []ApprovalStep `json:"steps"`
	Enabled     bool           `json:"enabled"`
	Created     time.Time      `json:"created"`
	Updated     time.Time      `json:"updated"`
}

// ApprovalStep represents approval step
type ApprovalStep struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"` // manual, automatic, condition
	Approvers     []string          `json:"approvers"`
	RequiredVotes int               `json:"required_votes"`
	Timeout       time.Duration     `json:"timeout"`
	Conditions    map[string]string `json:"conditions"`
	Actions       []string          `json:"actions"`
}

// AuditLog represents audit log entry
type AuditLog struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	UserID     string            `json:"user_id"`
	UserEmail  string            `json:"user_email"`
	Action     string            `json:"action"`
	Resource   string            `json:"resource"`
	ResourceID string            `json:"resource_id"`
	IPAddress  string            `json:"ip_address"`
	UserAgent  string            `json:"user_agent"`
	Result     string            `json:"result"` // success, failure, error
	Details    map[string]string `json:"details"`
	SessionID  string            `json:"session_id"`
}

// AuditLogger manages audit logging
type AuditLogger struct {
	LogDir      string
	MaxLogSize  int64
	MaxLogFiles int
	Format      string // json, csv, syslog
}

// SSOConfig represents SSO configuration
type SSOConfig struct {
	Provider    string            `json:"provider"`
	EntityID    string            `json:"entity_id"`
	MetadataURL string            `json:"metadata_url"`
	CertFile    string            `json:"cert_file"`
	KeyFile     string            `json:"key_file"`
	Attributes  map[string]string `json:"attributes"`
	Enabled     bool              `json:"enabled"`
}

// SSOUserInfo represents SSO user information
type SSOUserInfo struct {
	SubjectID  string            `json:"subject_id"`
	Attributes map[string]string `json:"attributes"`
	LastLogin  time.Time         `json:"last_login"`
	SessionID  string            `json:"session_id"`
	Provider   string            `json:"provider"`
}

// LDAPConfig represents LDAP configuration
type LDAPConfig struct {
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	BaseDN       string            `json:"base_dn"`
	BindDN       string            `json:"bind_dn"`
	BindPassword string            `json:"bind_password"`
	UserFilter   string            `json:"user_filter"`
	GroupFilter  string            `json:"group_filter"`
	Attributes   map[string]string `json:"attributes"`
	TLS          bool              `json:"tls"`
	Enabled      bool              `json:"enabled"`
}

// Subscription represents enterprise subscription
type Subscription struct {
	ID           string    `json:"id"`
	Plan         string    `json:"plan"`
	Status       string    `json:"status"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	UserLimit    int       `json:"user_limit"`
	StorageLimit int64     `json:"storage_limit"`
	Features     []string  `json:"features"`
}

func runEnterprise(cmd *cobra.Command, args []string) {
	fmt.Printf("🏢 기업용 프라이빗 마켓플레이스\n")
	fmt.Printf("📋 사용 가능한 하위 명령어:\n")
	fmt.Printf("  • init      - 조직 초기화\n")
	fmt.Printf("  • user      - 사용자 관리\n")
	fmt.Printf("  • approval  - 승인 워크플로우 관리\n")
	fmt.Printf("  • audit     - 감사 로그 관리\n")
	fmt.Printf("  • policy    - 정책 관리\n")
	fmt.Printf("  • backup    - 백업/복원 관리\n")
	fmt.Printf("\n💡 자세한 도움말: gz template enterprise <command> --help\n")
}

// Enterprise subcommands
var enterpriseInitCmd = &cobra.Command{
	Use:   "init",
	Short: "엔터프라이즈 조직 초기화",
	Run:   runEnterpriseInit,
}

var enterpriseUserCmd = &cobra.Command{
	Use:   "user",
	Short: "사용자 관리",
	Run:   runEnterpriseUser,
}

var enterpriseApprovalCmd = &cobra.Command{
	Use:   "approval",
	Short: "승인 워크플로우 관리",
	Run:   runEnterpriseApproval,
}

var enterpriseAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "감사 로그 관리",
	Run:   runEnterpriseAudit,
}

var enterprisePolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "정책 관리",
	Run:   runEnterprisePolicy,
}

var enterpriseBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "백업/복원 관리",
	Run:   runEnterpriseBackup,
}

func runEnterpriseInit(cmd *cobra.Command, args []string) {
	if orgName == "" {
		fmt.Printf("❌ 조직 이름이 필요합니다 (--org)\n")
		os.Exit(1)
	}

	fmt.Printf("🏢 조직 초기화: %s\n", orgName)

	// Initialize enterprise server
	server, err := initializeEnterpriseServer()
	if err != nil {
		fmt.Printf("❌ 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	// Create organization
	org := &Organization{
		ID:       generateOrgID(),
		Name:     orgName,
		Settings: make(map[string]string),
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	server.Organization = org

	// Create default roles and permissions
	if err := createDefaultRolesAndPermissions(server); err != nil {
		fmt.Printf("❌ 기본 역할 생성 실패: %v\n", err)
		os.Exit(1)
	}

	// Save configuration
	if err := saveEnterpriseConfig(server); err != nil {
		fmt.Printf("❌ 설정 저장 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 조직 초기화 완료\n")
	fmt.Printf("🆔 조직 ID: %s\n", org.ID)
}

func runEnterpriseUser(cmd *cobra.Command, args []string) {
	fmt.Printf("👥 사용자 관리\n")
	// Implementation for user management
}

func runEnterpriseApproval(cmd *cobra.Command, args []string) {
	fmt.Printf("🔄 승인 워크플로우 관리\n")
	// Implementation for approval workflow management
}

func runEnterpriseAudit(cmd *cobra.Command, args []string) {
	fmt.Printf("📊 감사 로그 관리\n")
	// Implementation for audit log management
}

func runEnterprisePolicy(cmd *cobra.Command, args []string) {
	fmt.Printf("📋 정책 관리\n")
	// Implementation for policy management
}

func runEnterpriseBackup(cmd *cobra.Command, args []string) {
	fmt.Printf("💾 백업/복원 관리\n")
	// Implementation for backup/restore management
}

// Enterprise server methods

func initializeEnterpriseServer() (*EnterpriseServer, error) {
	// Initialize base template server
	baseServer, err := initializeServer()
	if err != nil {
		return nil, err
	}

	// Create audit logger
	auditLogger := &AuditLogger{
		LogDir:      auditLogDir,
		MaxLogSize:  100 * 1024 * 1024, // 100MB
		MaxLogFiles: 30,
		Format:      "json",
	}

	// Create directories
	dirs := []string{auditLogDir, backupDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("디렉터리 생성 실패 %s: %w", dir, err)
		}
	}

	enterpriseServer := &EnterpriseServer{
		TemplateServer: baseServer,
		Users:          make(map[string]*EnterpriseUser),
		Roles:          make(map[string]*Role),
		Permissions:    make(map[string]*Permission),
		Policies:       make(map[string]*Policy),
		Workflows:      make(map[string]*ApprovalWorkflow),
		AuditLogger:    auditLogger,
	}

	return enterpriseServer, nil
}

func createDefaultRolesAndPermissions(server *EnterpriseServer) error {
	// Create default permissions
	permissions := []*Permission{
		{ID: "template.read", Name: "템플릿 읽기", Resource: "template", Action: "read"},
		{ID: "template.write", Name: "템플릿 쓰기", Resource: "template", Action: "write"},
		{ID: "template.delete", Name: "템플릿 삭제", Resource: "template", Action: "delete"},
		{ID: "template.approve", Name: "템플릿 승인", Resource: "template", Action: "approve"},
		{ID: "user.read", Name: "사용자 읽기", Resource: "user", Action: "read"},
		{ID: "user.write", Name: "사용자 쓰기", Resource: "user", Action: "write"},
		{ID: "user.admin", Name: "사용자 관리", Resource: "user", Action: "admin"},
		{ID: "audit.read", Name: "감사 로그 읽기", Resource: "audit", Action: "read"},
		{ID: "system.admin", Name: "시스템 관리", Resource: "system", Action: "admin"},
	}

	for _, perm := range permissions {
		perm.Created = time.Now()
		server.Permissions[perm.ID] = perm
	}

	// Create default roles
	roles := []*Role{
		{
			ID:          "viewer",
			Name:        "뷰어",
			Description: "템플릿 읽기 전용",
			Permissions: []string{"template.read"},
		},
		{
			ID:          "developer",
			Name:        "개발자",
			Description: "템플릿 읽기/쓰기",
			Permissions: []string{"template.read", "template.write"},
		},
		{
			ID:          "approver",
			Name:        "승인자",
			Description: "템플릿 승인 권한",
			Permissions: []string{"template.read", "template.write", "template.approve"},
		},
		{
			ID:          "admin",
			Name:        "관리자",
			Description: "모든 권한",
			Permissions: []string{
				"template.read", "template.write", "template.delete", "template.approve",
				"user.read", "user.write", "user.admin",
				"audit.read", "system.admin",
			},
		},
	}

	for _, role := range roles {
		role.Created = time.Now()
		role.Updated = time.Now()
		server.Roles[role.ID] = role
	}

	return nil
}

func saveEnterpriseConfig(server *EnterpriseServer) error {
	configDir := "./enterprise"
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	// Save organization config
	orgFile := filepath.Join(configDir, "organization.json")
	orgData, err := json.MarshalIndent(server.Organization, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(orgFile, orgData, 0o644); err != nil {
		return err
	}

	// Save roles config
	rolesFile := filepath.Join(configDir, "roles.json")
	rolesData, err := json.MarshalIndent(server.Roles, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(rolesFile, rolesData, 0o644); err != nil {
		return err
	}

	// Save permissions config
	permsFile := filepath.Join(configDir, "permissions.json")
	permsData, err := json.MarshalIndent(server.Permissions, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(permsFile, permsData, 0o644); err != nil {
		return err
	}

	return nil
}

// Utility functions

func generateOrgID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "org-" + hex.EncodeToString(bytes)
}

// Enterprise middleware for template server

func (es *EnterpriseServer) rbacMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check user authentication and authorization
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "인증이 필요합니다", http.StatusUnauthorized)
			return
		}

		user, exists := es.Users[userID]
		if !exists || user.Status != "active" {
			http.Error(w, "유효하지 않은 사용자", http.StatusUnauthorized)
			return
		}

		// Check permissions
		resource, action := extractResourceAction(r)
		if !es.hasPermission(user, resource, action) {
			http.Error(w, "권한이 없습니다", http.StatusForbidden)
			es.logAudit(user, r, "access_denied", resource, "failure")
			return
		}

		// Log audit
		es.logAudit(user, r, action, resource, "success")

		next.ServeHTTP(w, r)
	})
}

func (es *EnterpriseServer) hasPermission(user *EnterpriseUser, resource, action string) bool {
	permissionID := resource + "." + action

	// Check direct permissions
	for _, perm := range user.Permissions {
		if perm == permissionID {
			return true
		}
	}

	// Check role permissions
	for _, roleID := range user.Roles {
		if role, exists := es.Roles[roleID]; exists {
			for _, perm := range role.Permissions {
				if perm == permissionID {
					return true
				}
			}
		}
	}

	return false
}

func (es *EnterpriseServer) logAudit(user *EnterpriseUser, r *http.Request, action, resource, result string) {
	auditLog := &AuditLog{
		ID:        generateAuditID(),
		Timestamp: time.Now(),
		UserID:    user.ID,
		UserEmail: user.Email,
		Action:    action,
		Resource:  resource,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Result:    result,
		Details:   make(map[string]string),
	}

	// Save audit log
	es.saveAuditLog(auditLog)
}

func (es *EnterpriseServer) saveAuditLog(log *AuditLog) {
	// In a real implementation, this would save to a database or file
	logFile := filepath.Join(es.AuditLogger.LogDir, fmt.Sprintf("audit-%s.json", time.Now().Format("2006-01-02")))

	data, _ := json.Marshal(log)
	data = append(data, '\n')

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		file.Write(data)
		file.Close()
	}
}

func extractResourceAction(r *http.Request) (string, string) {
	// Extract resource and action from HTTP request
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	parts := strings.Split(path, "/")

	resource := "unknown"
	action := "unknown"

	if len(parts) > 0 {
		resource = parts[0]
	}

	switch r.Method {
	case "GET":
		action = "read"
	case "POST":
		action = "write"
	case "PUT":
		action = "write"
	case "DELETE":
		action = "delete"
	}

	return resource, action
}

func generateAuditID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "audit-" + hex.EncodeToString(bytes)
}

// Extended template server with enterprise features
func setupEnterpriseRouter(server *EnterpriseServer) *mux.Router {
	router := setupRouter(server.TemplateServer)

	// Add enterprise middleware
	router.Use(server.rbacMiddleware)

	// Add enterprise API routes
	api := router.PathPrefix("/api/v1/enterprise").Subrouter()

	// Organization routes
	api.HandleFunc("/organization", server.getOrganization).Methods("GET")
	api.HandleFunc("/organization", server.updateOrganization).Methods("PUT")

	// User management routes
	api.HandleFunc("/users", server.listEnterpriseUsers).Methods("GET")
	api.HandleFunc("/users", server.createEnterpriseUser).Methods("POST")
	api.HandleFunc("/users/{id}", server.getEnterpriseUser).Methods("GET")
	api.HandleFunc("/users/{id}", server.updateEnterpriseUser).Methods("PUT")
	api.HandleFunc("/users/{id}", server.deleteEnterpriseUser).Methods("DELETE")

	// Role and permission routes
	api.HandleFunc("/roles", server.listRoles).Methods("GET")
	api.HandleFunc("/roles", server.createRole).Methods("POST")
	api.HandleFunc("/permissions", server.listPermissions).Methods("GET")

	// Policy routes
	api.HandleFunc("/policies", server.listPolicies).Methods("GET")
	api.HandleFunc("/policies", server.createPolicy).Methods("POST")
	api.HandleFunc("/policies/{id}", server.getPolicy).Methods("GET")
	api.HandleFunc("/policies/{id}", server.updatePolicy).Methods("PUT")

	// Audit routes
	api.HandleFunc("/audit", server.getAuditLogs).Methods("GET")
	api.HandleFunc("/audit/export", server.exportAuditLogs).Methods("GET")

	// Workflow routes
	api.HandleFunc("/workflows", server.listWorkflows).Methods("GET")
	api.HandleFunc("/workflows", server.createWorkflow).Methods("POST")

	return router
}

// Placeholder implementations for enterprise handlers
func (es *EnterpriseServer) getOrganization(w http.ResponseWriter, r *http.Request)      { /* TODO */ }
func (es *EnterpriseServer) updateOrganization(w http.ResponseWriter, r *http.Request)   { /* TODO */ }
func (es *EnterpriseServer) listEnterpriseUsers(w http.ResponseWriter, r *http.Request)  { /* TODO */ }
func (es *EnterpriseServer) createEnterpriseUser(w http.ResponseWriter, r *http.Request) { /* TODO */ }
func (es *EnterpriseServer) getEnterpriseUser(w http.ResponseWriter, r *http.Request)    { /* TODO */ }
func (es *EnterpriseServer) updateEnterpriseUser(w http.ResponseWriter, r *http.Request) { /* TODO */ }
func (es *EnterpriseServer) deleteEnterpriseUser(w http.ResponseWriter, r *http.Request) { /* TODO */ }
func (es *EnterpriseServer) listRoles(w http.ResponseWriter, r *http.Request)            { /* TODO */ }
func (es *EnterpriseServer) createRole(w http.ResponseWriter, r *http.Request)           { /* TODO */ }
func (es *EnterpriseServer) listPermissions(w http.ResponseWriter, r *http.Request)      { /* TODO */ }
func (es *EnterpriseServer) listPolicies(w http.ResponseWriter, r *http.Request)         { /* TODO */ }
func (es *EnterpriseServer) createPolicy(w http.ResponseWriter, r *http.Request)         { /* TODO */ }
func (es *EnterpriseServer) getPolicy(w http.ResponseWriter, r *http.Request)            { /* TODO */ }
func (es *EnterpriseServer) updatePolicy(w http.ResponseWriter, r *http.Request)         { /* TODO */ }
func (es *EnterpriseServer) getAuditLogs(w http.ResponseWriter, r *http.Request)         { /* TODO */ }
func (es *EnterpriseServer) exportAuditLogs(w http.ResponseWriter, r *http.Request)      { /* TODO */ }
func (es *EnterpriseServer) listWorkflows(w http.ResponseWriter, r *http.Request)        { /* TODO */ }
func (es *EnterpriseServer) createWorkflow(w http.ResponseWriter, r *http.Request)       { /* TODO */ }
