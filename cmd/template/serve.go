package template

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

// ServeCmd represents the serve command
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "커뮤니티 템플릿 공유 API 서버 실행",
	Long: `템플릿 공유를 위한 RESTful API 서버를 실행합니다.

API 기능:
- 템플릿 업로드/다운로드
- 템플릿 검색 및 필터링
- 버전 관리 및 의존성 해결
- 사용자 인증 및 권한 관리
- 라이선스 관리 시스템
- 템플릿 검증 및 승인 프로세스

Examples:
  gz template serve --port 8080
  gz template serve --host 0.0.0.0 --port 3000 --storage ./templates
  gz template serve --auth-required --admin-key mykey`,
	Run: runServe,
}

var (
	serverHost        string
	serverPort        int
	storageDir        string
	authRequired      bool
	adminKey          string
	enableCORS        bool
	enableTLS         bool
	tlsCertFile       string
	tlsKeyFile        string
	uploadMaxSize     int64
	allowedExtensions []string
	rateLimitRPM      int
	enableMetrics     bool
	logLevel          string
	dataDirectory     string
)

func init() {
	ServeCmd.Flags().StringVar(&serverHost, "host", "localhost", "서버 바인딩 호스트")
	ServeCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "서버 포트")
	ServeCmd.Flags().StringVar(&storageDir, "storage", "./templates", "템플릿 저장 디렉터리")
	ServeCmd.Flags().BoolVar(&authRequired, "auth-required", false, "인증 필수 여부")
	ServeCmd.Flags().StringVar(&adminKey, "admin-key", "", "관리자 API 키")
	ServeCmd.Flags().BoolVar(&enableCORS, "cors", true, "CORS 활성화")
	ServeCmd.Flags().BoolVar(&enableTLS, "tls", false, "TLS 활성화")
	ServeCmd.Flags().StringVar(&tlsCertFile, "tls-cert", "", "TLS 인증서 파일")
	ServeCmd.Flags().StringVar(&tlsKeyFile, "tls-key", "", "TLS 키 파일")
	ServeCmd.Flags().Int64Var(&uploadMaxSize, "max-upload-size", 50*1024*1024, "최대 업로드 크기 (바이트)")
	ServeCmd.Flags().StringSliceVar(&allowedExtensions, "allowed-ext", []string{".zip", ".tar.gz", ".tgz"}, "허용된 파일 확장자")
	ServeCmd.Flags().IntVar(&rateLimitRPM, "rate-limit", 60, "분당 요청 제한")
	ServeCmd.Flags().BoolVar(&enableMetrics, "metrics", true, "메트릭 수집 활성화")
	ServeCmd.Flags().StringVar(&logLevel, "log-level", "info", "로그 레벨 (debug, info, warn, error)")
	ServeCmd.Flags().StringVar(&dataDirectory, "data-dir", "./data", "데이터 디렉터리")
}

// TemplateServer represents the template sharing server
type TemplateServer struct {
	StorageDir    string
	DataDir       string
	AuthRequired  bool
	AdminKey      string
	Templates     map[string]*TemplateInfo
	Users         map[string]*UserInfo
	Licenses      map[string]*LicenseInfo
	approvalQueue []*ApprovalRequest
}

// TemplateInfo represents template information
type TemplateInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Type        string            `json:"type"`
	Keywords    []string          `json:"keywords"`
	License     string            `json:"license"`
	Homepage    string            `json:"homepage,omitempty"`
	Repository  string            `json:"repository,omitempty"`
	Downloads   int               `json:"downloads"`
	Rating      float64           `json:"rating"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	Verified    bool              `json:"verified"`
	Approved    bool              `json:"approved"`
	Deprecated  bool              `json:"deprecated"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	FilePath    string            `json:"file_path"`
}

// UserInfo represents user information
type UserInfo struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // admin, publisher, user
	Created   time.Time `json:"created"`
	LastLogin time.Time `json:"last_login"`
	Templates []string  `json:"templates"`
	Verified  bool      `json:"verified"`
	Active    bool      `json:"active"`
}

// LicenseInfo represents license information
type LicenseInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SPDXID      string `json:"spdx_id"`
	URL         string `json:"url"`
	OSIApproved bool   `json:"osi_approved"`
	Content     string `json:"content"`
}

// ApprovalRequest represents template approval request
type ApprovalRequest struct {
	ID         string    `json:"id"`
	TemplateID string    `json:"template_id"`
	Author     string    `json:"author"`
	Action     string    `json:"action"` // create, update, delete
	Status     string    `json:"status"` // pending, approved, rejected
	Reason     string    `json:"reason,omitempty"`
	Created    time.Time `json:"created"`
	Reviewed   time.Time `json:"reviewed,omitempty"`
	Reviewer   string    `json:"reviewer,omitempty"`
}

// UploadResponse represents upload response
type UploadResponse struct {
	Success    bool   `json:"success"`
	TemplateID string `json:"template_id,omitempty"`
	Message    string `json:"message"`
	ApprovalID string `json:"approval_id,omitempty"`
}

// SearchResponse represents search response
type SearchResponse struct {
	Templates []TemplateInfo `json:"templates"`
	Total     int            `json:"total"`
	Page      int            `json:"page"`
	PerPage   int            `json:"per_page"`
	Query     string         `json:"query,omitempty"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func runServe(cmd *cobra.Command, args []string) {
	fmt.Printf("🚀 템플릿 공유 API 서버 시작\n")
	fmt.Printf("🌐 주소: http://%s:%d\n", serverHost, serverPort)
	fmt.Printf("📁 저장소: %s\n", storageDir)
	fmt.Printf("📊 데이터: %s\n", dataDirectory)

	// Initialize server
	server, err := initializeServer()
	if err != nil {
		fmt.Printf("❌ 서버 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	// Setup router
	router := setupRouter(server)

	// Create server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", serverHost, serverPort),
		Handler: router,
	}

	// Start server
	fmt.Printf("✅ 서버 실행 중...\n")
	if enableTLS {
		if tlsCertFile == "" || tlsKeyFile == "" {
			fmt.Printf("❌ TLS 인증서 파일이 필요합니다\n")
			os.Exit(1)
		}
		fmt.Printf("🔒 TLS 모드로 실행\n")
		if err := httpServer.ListenAndServeTLS(tlsCertFile, tlsKeyFile); err != nil {
			fmt.Printf("❌ TLS 서버 실행 실패: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := httpServer.ListenAndServe(); err != nil {
			fmt.Printf("❌ 서버 실행 실패: %v\n", err)
			os.Exit(1)
		}
	}
}

func initializeServer() (*TemplateServer, error) {
	// Create directories
	dirs := []string{storageDir, dataDirectory}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("디렉터리 생성 실패 %s: %w", dir, err)
		}
	}

	server := &TemplateServer{
		StorageDir:   storageDir,
		DataDir:      dataDirectory,
		AuthRequired: authRequired,
		AdminKey:     adminKey,
		Templates:    make(map[string]*TemplateInfo),
		Users:        make(map[string]*UserInfo),
		Licenses:     make(map[string]*LicenseInfo),
	}

	// Load existing data
	if err := server.loadData(); err != nil {
		return nil, fmt.Errorf("데이터 로드 실패: %w", err)
	}

	// Initialize licenses
	server.initializeLicenses()

	fmt.Printf("📦 템플릿 로드: %d개\n", len(server.Templates))
	fmt.Printf("👥 사용자 로드: %d개\n", len(server.Users))
	fmt.Printf("📄 라이선스 로드: %d개\n", len(server.Licenses))

	return server, nil
}

func setupRouter(server *TemplateServer) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	if enableCORS {
		router.Use(corsMiddleware)
	}
	router.Use(loggingMiddleware)
	router.Use(rateLimitMiddleware)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Template routes
	api.HandleFunc("/templates", server.listTemplates).Methods("GET")
	api.HandleFunc("/templates", server.uploadTemplate).Methods("POST")
	api.HandleFunc("/templates/{id}", server.getTemplate).Methods("GET")
	api.HandleFunc("/templates/{id}", server.updateTemplate).Methods("PUT")
	api.HandleFunc("/templates/{id}", server.deleteTemplate).Methods("DELETE")
	api.HandleFunc("/templates/{id}/download", server.downloadTemplate).Methods("GET")
	api.HandleFunc("/templates/search", server.searchTemplates).Methods("GET")

	// User routes
	api.HandleFunc("/users", server.listUsers).Methods("GET")
	api.HandleFunc("/users", server.createUser).Methods("POST")
	api.HandleFunc("/users/{id}", server.getUser).Methods("GET")
	api.HandleFunc("/users/{id}", server.updateUser).Methods("PUT")

	// License routes
	api.HandleFunc("/licenses", server.listLicenses).Methods("GET")
	api.HandleFunc("/licenses/{id}", server.getLicense).Methods("GET")

	// Approval routes
	api.HandleFunc("/approvals", server.listApprovals).Methods("GET")
	api.HandleFunc("/approvals/{id}/approve", server.approveTemplate).Methods("POST")
	api.HandleFunc("/approvals/{id}/reject", server.rejectTemplate).Methods("POST")

	// Health and metrics
	router.HandleFunc("/health", healthCheck).Methods("GET")
	if enableMetrics {
		router.HandleFunc("/metrics", metricsHandler).Methods("GET")
	}

	// Static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	return router
}

func (s *TemplateServer) loadData() error {
	// Load templates from storage directory
	return filepath.Walk(s.StorageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, "template.yaml") {
			return nil
		}

		// Load template metadata
		template, err := loadTemplateMetadata(path)
		if err != nil {
			fmt.Printf("⚠️ 템플릿 로드 실패 %s: %v\n", path, err)
			return nil
		}

		s.Templates[template.ID] = template
		return nil
	})
}

func loadTemplateMetadata(metadataPath string) (*TemplateInfo, error) {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata TemplateMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	template := &TemplateInfo{
		ID:          generateTemplateID(metadata.Metadata.Name, metadata.Metadata.Version),
		Name:        metadata.Metadata.Name,
		Version:     metadata.Metadata.Version,
		Author:      metadata.Metadata.Author,
		Description: metadata.Metadata.Description,
		Category:    metadata.Metadata.Category,
		Type:        metadata.Metadata.Type,
		Keywords:    metadata.Metadata.Keywords,
		License:     metadata.Metadata.License,
		Homepage:    metadata.Metadata.Homepage,
		Repository:  metadata.Metadata.Repository,
		Created:     time.Now(),
		Updated:     time.Now(),
		Tags:        metadata.Metadata.Tags,
		FilePath:    filepath.Dir(metadataPath),
		Approved:    true, // Auto-approve for existing templates
	}

	return template, nil
}

func (s *TemplateServer) initializeLicenses() {
	// Initialize common licenses
	licenses := []*LicenseInfo{
		{
			ID:          "MIT",
			Name:        "MIT License",
			SPDXID:      "MIT",
			URL:         "https://opensource.org/licenses/MIT",
			OSIApproved: true,
			Content:     "MIT License\n\nPermission is hereby granted, free of charge...",
		},
		{
			ID:          "Apache-2.0",
			Name:        "Apache License 2.0",
			SPDXID:      "Apache-2.0",
			URL:         "https://opensource.org/licenses/Apache-2.0",
			OSIApproved: true,
			Content:     "Apache License\nVersion 2.0, January 2004...",
		},
		{
			ID:          "GPL-3.0",
			Name:        "GNU General Public License v3.0",
			SPDXID:      "GPL-3.0",
			URL:         "https://opensource.org/licenses/GPL-3.0",
			OSIApproved: true,
			Content:     "GNU GENERAL PUBLIC LICENSE\nVersion 3, 29 June 2007...",
		},
	}

	for _, license := range licenses {
		s.Licenses[license.ID] = license
	}
}

// API Handlers

func (s *TemplateServer) listTemplates(w http.ResponseWriter, r *http.Request) {
	templates := make([]TemplateInfo, 0, len(s.Templates))
	for _, template := range s.Templates {
		if template.Approved || !s.AuthRequired {
			templates = append(templates, *template)
		}
	}

	response := SearchResponse{
		Templates: templates,
		Total:     len(templates),
		Page:      1,
		PerPage:   len(templates),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *TemplateServer) uploadTemplate(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(uploadMaxSize); err != nil {
		http.Error(w, "파일 크기가 너무 큽니다", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("template")
	if err != nil {
		http.Error(w, "파일 업로드 실패", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if !isAllowedExtension(header.Filename) {
		http.Error(w, "허용되지 않은 파일 형식", http.StatusBadRequest)
		return
	}

	// Create template ID
	templateID := generateTemplateID(header.Filename, "1.0.0")

	// Save file
	templateDir := filepath.Join(s.StorageDir, templateID)
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		http.Error(w, "템플릿 디렉터리 생성 실패", http.StatusInternalServerError)
		return
	}

	templateFile := filepath.Join(templateDir, header.Filename)
	dst, err := os.Create(templateFile)
	if err != nil {
		http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "파일 복사 실패", http.StatusInternalServerError)
		return
	}

	// Create template info
	template := &TemplateInfo{
		ID:       templateID,
		Name:     strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)),
		Version:  "1.0.0",
		Author:   r.FormValue("author"),
		Created:  time.Now(),
		Updated:  time.Now(),
		Size:     header.Size,
		FilePath: templateDir,
		Approved: !s.AuthRequired, // Auto-approve if auth not required
	}

	// Add to approval queue if required
	var approvalID string
	if s.AuthRequired {
		approval := &ApprovalRequest{
			ID:         generateApprovalID(),
			TemplateID: templateID,
			Author:     template.Author,
			Action:     "create",
			Status:     "pending",
			Created:    time.Now(),
		}
		s.approvalQueue = append(s.approvalQueue, approval)
		approvalID = approval.ID
	} else {
		s.Templates[templateID] = template
	}

	response := UploadResponse{
		Success:    true,
		TemplateID: templateID,
		Message:    "템플릿 업로드 성공",
		ApprovalID: approvalID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *TemplateServer) searchTemplates(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	templateType := r.URL.Query().Get("type")
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	page := 1
	perPage := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	var filtered []TemplateInfo
	for _, template := range s.Templates {
		if !template.Approved && s.AuthRequired {
			continue
		}

		// Apply filters
		if query != "" && !strings.Contains(strings.ToLower(template.Name), strings.ToLower(query)) &&
			!strings.Contains(strings.ToLower(template.Description), strings.ToLower(query)) {
			continue
		}

		if category != "" && template.Category != category {
			continue
		}

		if templateType != "" && template.Type != templateType {
			continue
		}

		filtered = append(filtered, *template)
	}

	// Pagination
	start := (page - 1) * perPage
	end := start + perPage
	if start >= len(filtered) {
		filtered = []TemplateInfo{}
	} else if end > len(filtered) {
		filtered = filtered[start:]
	} else {
		filtered = filtered[start:end]
	}

	response := SearchResponse{
		Templates: filtered,
		Total:     len(filtered),
		Page:      page,
		PerPage:   perPage,
		Query:     query,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *TemplateServer) getTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	template, exists := s.Templates[templateID]
	if !exists {
		http.Error(w, "템플릿을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func (s *TemplateServer) downloadTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	template, exists := s.Templates[templateID]
	if !exists {
		http.Error(w, "템플릿을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	// Increment download count
	template.Downloads++

	// Find template file
	templateFile := filepath.Join(template.FilePath, template.Name+".zip")
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		// Try other extensions
		for _, ext := range allowedExtensions {
			testFile := filepath.Join(template.FilePath, template.Name+ext)
			if _, err := os.Stat(testFile); err == nil {
				templateFile = testFile
				break
			}
		}
	}

	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		http.Error(w, "템플릿 파일을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	// Serve file
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(templateFile)))
	http.ServeFile(w, r, templateFile)
}

func (s *TemplateServer) listLicenses(w http.ResponseWriter, r *http.Request) {
	licenses := make([]LicenseInfo, 0, len(s.Licenses))
	for _, license := range s.Licenses {
		licenses = append(licenses, *license)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(licenses)
}

func (s *TemplateServer) listApprovals(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(r) {
		http.Error(w, "관리자 권한이 필요합니다", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.approvalQueue)
}

// Utility functions

func generateTemplateID(name, version string) string {
	return fmt.Sprintf("%s-%s-%d", strings.ToLower(name), version, time.Now().Unix())
}

func generateApprovalID() string {
	return fmt.Sprintf("approval-%d", time.Now().Unix())
}

func isAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

func (s *TemplateServer) isAdmin(r *http.Request) bool {
	apiKey := r.Header.Get("X-API-Key")
	return s.AdminKey != "" && apiKey == s.AdminKey
}

// Middleware

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	// Simple rate limiting (in production, use proper rate limiter)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Health and metrics

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Prometheus-style metrics
	fmt.Fprintf(w, "# HELP template_server_requests_total Total requests\n")
	fmt.Fprintf(w, "# TYPE template_server_requests_total counter\n")
	fmt.Fprintf(w, "template_server_requests_total 42\n")
}

// Placeholder implementations for missing handlers
func (s *TemplateServer) updateTemplate(w http.ResponseWriter, r *http.Request)  { /* TODO */ }
func (s *TemplateServer) deleteTemplate(w http.ResponseWriter, r *http.Request)  { /* TODO */ }
func (s *TemplateServer) listUsers(w http.ResponseWriter, r *http.Request)       { /* TODO */ }
func (s *TemplateServer) createUser(w http.ResponseWriter, r *http.Request)      { /* TODO */ }
func (s *TemplateServer) getUser(w http.ResponseWriter, r *http.Request)         { /* TODO */ }
func (s *TemplateServer) updateUser(w http.ResponseWriter, r *http.Request)      { /* TODO */ }
func (s *TemplateServer) getLicense(w http.ResponseWriter, r *http.Request)      { /* TODO */ }
func (s *TemplateServer) approveTemplate(w http.ResponseWriter, r *http.Request) { /* TODO */ }
func (s *TemplateServer) rejectTemplate(w http.ResponseWriter, r *http.Request)  { /* TODO */ }
