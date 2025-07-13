package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TemplateClient represents a client for template sharing API
type TemplateClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

// ClientConfig represents client configuration
type ClientConfig struct {
	BaseURL   string `yaml:"base_url" json:"base_url"`
	APIKey    string `yaml:"api_key" json:"api_key"`
	Timeout   int    `yaml:"timeout" json:"timeout"`
	UserAgent string `yaml:"user_agent" json:"user_agent"`
}

// NewTemplateClient creates a new template client
func NewTemplateClient(config *ClientConfig) *TemplateClient {
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	userAgent := "gzh-manager/1.0.0"
	if config.UserAgent != "" {
		userAgent = config.UserAgent
	}

	return &TemplateClient{
		BaseURL: strings.TrimSuffix(config.BaseURL, "/"),
		APIKey:  config.APIKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		UserAgent: userAgent,
	}
}

// UploadTemplate uploads a template to the server
func (c *TemplateClient) UploadTemplate(templatePath, author string) (*UploadResponse, error) {
	// Open template file
	file, err := os.Open(templatePath)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file field
	fileWriter, err := writer.CreateFormFile("template", filepath.Base(templatePath))
	if err != nil {
		return nil, fmt.Errorf("폼 필드 생성 실패: %w", err)
	}

	if _, err := io.Copy(fileWriter, file); err != nil {
		return nil, fmt.Errorf("파일 복사 실패: %w", err)
	}

	// Add author field
	if err := writer.WriteField("author", author); err != nil {
		return nil, fmt.Errorf("작성자 필드 추가 실패: %w", err)
	}

	writer.Close()

	// Create request
	url := fmt.Sprintf("%s/api/v1/templates", c.BaseURL)
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &response, fmt.Errorf("업로드 실패: %s", response.Message)
	}

	return &response, nil
}

// DownloadTemplate downloads a template from the server
func (c *TemplateClient) DownloadTemplate(templateID, outputPath string) error {
	url := fmt.Sprintf("%s/api/v1/templates/%s/download", c.BaseURL, templateID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("요청 생성 실패: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("다운로드 실패: 상태 코드 %d", resp.StatusCode)
	}

	// Create output file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("출력 디렉터리 생성 실패: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("출력 파일 생성 실패: %w", err)
	}
	defer outFile.Close()

	// Copy content
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("파일 저장 실패: %w", err)
	}

	return nil
}

// SearchTemplates searches for templates
func (c *TemplateClient) SearchTemplates(query, category, templateType string, page, perPage int) (*SearchResponse, error) {
	url := fmt.Sprintf("%s/api/v1/templates/search", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if query != "" {
		q.Add("q", query)
	}
	if category != "" {
		q.Add("category", category)
	}
	if templateType != "" {
		q.Add("type", templateType)
	}
	if page > 0 {
		q.Add("page", fmt.Sprintf("%d", page))
	}
	if perPage > 0 {
		q.Add("per_page", fmt.Sprintf("%d", perPage))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	var response SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	return &response, nil
}

// GetTemplate gets template details
func (c *TemplateClient) GetTemplate(templateID string) (*TemplateInfo, error) {
	url := fmt.Sprintf("%s/api/v1/templates/%s", c.BaseURL, templateID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("템플릿을 찾을 수 없습니다: %s", templateID)
	}

	var template TemplateInfo
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	return &template, nil
}

// ListTemplates lists all templates
func (c *TemplateClient) ListTemplates() (*SearchResponse, error) {
	url := fmt.Sprintf("%s/api/v1/templates", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	var response SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	return &response, nil
}

// ListLicenses lists available licenses
func (c *TemplateClient) ListLicenses() ([]LicenseInfo, error) {
	url := fmt.Sprintf("%s/api/v1/licenses", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	var licenses []LicenseInfo
	if err := json.NewDecoder(resp.Body).Decode(&licenses); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	return licenses, nil
}

// HealthCheck checks server health
func (c *TemplateClient) HealthCheck() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/health", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	return health, nil
}

// LoadClientConfig loads client configuration from file
func LoadClientConfig(configPath string) (*ClientConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}

	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("설정 파싱 실패: %w", err)
	}

	return &config, nil
}

// SaveClientConfig saves client configuration to file
func SaveClientConfig(config *ClientConfig, configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("설정 디렉터리 생성 실패: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 마샬링 실패: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("설정 파일 쓰기 실패: %w", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./template-client.json"
	}
	return filepath.Join(homeDir, ".config", "gzh-manager", "template-client.json")
}
