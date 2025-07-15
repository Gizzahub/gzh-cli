package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// MarketplaceCmd represents the marketplace command
var MarketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "템플릿 마켓플레이스 관리",
	Long: `템플릿 마켓플레이스 시스템을 관리합니다.

마켓플레이스 기능:
- 저장소 초기화 및 설정
- 템플릿 메타데이터 인덱싱
- 의존성 그래프 생성
- 버전 관리 및 호환성 검사
- 접근 제어 및 권한 관리

Examples:
  gz template marketplace init --type community
  gz template marketplace init --type private --auth ldap
  gz template marketplace index
  gz template marketplace stats`,
	Run: runMarketplace,
}

var (
	marketplaceType   string
	marketplacePath   string
	authProvider      string
	enableVersioning  bool
	enableApproval    bool
	enableMetrics     bool
	adminUsers        []string
	allowedCategories []string
)

func init() {
	MarketplaceCmd.Flags().StringVar(&marketplaceType, "type", "community", "마켓플레이스 타입 (community, private, hybrid)")
	MarketplaceCmd.Flags().StringVar(&marketplacePath, "path", "./marketplace", "마켓플레이스 저장소 경로")
	MarketplaceCmd.Flags().StringVar(&authProvider, "auth", "local", "인증 제공자 (local, ldap, oauth)")
	MarketplaceCmd.Flags().BoolVar(&enableVersioning, "versioning", true, "버전 관리 활성화")
	MarketplaceCmd.Flags().BoolVar(&enableApproval, "approval", false, "승인 워크플로우 활성화")
	MarketplaceCmd.Flags().BoolVar(&enableMetrics, "metrics", true, "메트릭 수집 활성화")
	MarketplaceCmd.Flags().StringSliceVar(&adminUsers, "admins", []string{}, "관리자 사용자 목록")
	MarketplaceCmd.Flags().StringSliceVar(&allowedCategories, "categories", []string{}, "허용된 카테고리 목록")
}

// MarketplaceConfig represents marketplace configuration
type MarketplaceConfig struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	Metadata   struct {
		Name        string    `yaml:"name" json:"name"`
		Description string    `yaml:"description" json:"description"`
		Created     time.Time `yaml:"created" json:"created"`
		Updated     time.Time `yaml:"updated" json:"updated"`
	} `yaml:"metadata" json:"metadata"`
	Spec MarketplaceSpec `yaml:"spec" json:"spec"`
}

// MarketplaceSpec represents marketplace specification
type MarketplaceSpec struct {
	Type            string           `yaml:"type" json:"type"`
	Storage         StorageConfig    `yaml:"storage" json:"storage"`
	Authentication  AuthConfig       `yaml:"authentication" json:"authentication"`
	Authorization   AuthzConfig      `yaml:"authorization" json:"authorization"`
	Versioning      VersioningConfig `yaml:"versioning" json:"versioning"`
	Approval        ApprovalConfig   `yaml:"approval" json:"approval"`
	Indexing        IndexingConfig   `yaml:"indexing" json:"indexing"`
	Metrics         MetricsConfig    `yaml:"metrics" json:"metrics"`
	Categories      []string         `yaml:"categories" json:"categories"`
	DefaultLicense  string           `yaml:"defaultLicense" json:"defaultLicense"`
	MaxTemplateSize int64            `yaml:"maxTemplateSize" json:"maxTemplateSize"`
	RetentionPolicy RetentionPolicy  `yaml:"retentionPolicy" json:"retentionPolicy"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type       string            `yaml:"type" json:"type"`
	Path       string            `yaml:"path" json:"path"`
	Repository string            `yaml:"repository,omitempty" json:"repository,omitempty"`
	Bucket     string            `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	Region     string            `yaml:"region,omitempty" json:"region,omitempty"`
	Options    map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Provider string                 `yaml:"provider" json:"provider"`
	Config   map[string]string      `yaml:"config,omitempty" json:"config,omitempty"`
	LDAP     *MarketplaceLDAPConfig `yaml:"ldap,omitempty" json:"ldap,omitempty"`
	OAuth    *OAuthConfig           `yaml:"oauth,omitempty" json:"oauth,omitempty"`
}

// MarketplaceLDAPConfig represents LDAP configuration for marketplace
type MarketplaceLDAPConfig struct {
	Server  string `yaml:"server" json:"server"`
	Port    int    `yaml:"port" json:"port"`
	BaseDN  string `yaml:"baseDN" json:"baseDN"`
	UserDN  string `yaml:"userDN" json:"userDN"`
	GroupDN string `yaml:"groupDN" json:"groupDN"`
	TLS     bool   `yaml:"tls" json:"tls"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	Provider     string   `yaml:"provider" json:"provider"`
	ClientID     string   `yaml:"clientId" json:"clientId"`
	ClientSecret string   `yaml:"clientSecret" json:"clientSecret"`
	RedirectURL  string   `yaml:"redirectUrl" json:"redirectUrl"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
}

// AuthzConfig represents authorization configuration
type AuthzConfig struct {
	Enabled     bool                    `yaml:"enabled" json:"enabled"`
	AdminUsers  []string                `yaml:"adminUsers" json:"adminUsers"`
	AdminGroups []string                `yaml:"adminGroups" json:"adminGroups"`
	Roles       []MarketplaceRole       `yaml:"roles" json:"roles"`
	Permissions []MarketplacePermission `yaml:"permissions" json:"permissions"`
	DefaultRole string                  `yaml:"defaultRole" json:"defaultRole"`
}

// MarketplaceRole represents a user role in marketplace
type MarketplaceRole struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Permissions []string `yaml:"permissions" json:"permissions"`
}

// MarketplacePermission represents a permission in marketplace
type MarketplacePermission struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Resource    string   `yaml:"resource" json:"resource"`
	Actions     []string `yaml:"actions" json:"actions"`
}

// VersioningConfig represents versioning configuration
type VersioningConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	Strategy      string `yaml:"strategy" json:"strategy"`
	MaxVersions   int    `yaml:"maxVersions" json:"maxVersions"`
	AutoCleanup   bool   `yaml:"autoCleanup" json:"autoCleanup"`
	ImmutableTags bool   `yaml:"immutableTags" json:"immutableTags"`
}

// ApprovalConfig represents approval workflow configuration
type ApprovalConfig struct {
	Enabled      bool     `yaml:"enabled" json:"enabled"`
	RequiredFor  []string `yaml:"requiredFor" json:"requiredFor"`
	Approvers    []string `yaml:"approvers" json:"approvers"`
	MinApprovals int      `yaml:"minApprovals" json:"minApprovals"`
	AutoApprove  bool     `yaml:"autoApprove" json:"autoApprove"`
	TimeoutHours int      `yaml:"timeoutHours" json:"timeoutHours"`
}

// IndexingConfig represents indexing configuration
type IndexingConfig struct {
	Enabled       bool     `yaml:"enabled" json:"enabled"`
	RefreshRate   string   `yaml:"refreshRate" json:"refreshRate"`
	SearchFields  []string `yaml:"searchFields" json:"searchFields"`
	FacetFields   []string `yaml:"facetFields" json:"facetFields"`
	FullTextIndex bool     `yaml:"fullTextIndex" json:"fullTextIndex"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Provider  string `yaml:"provider" json:"provider"`
	Endpoint  string `yaml:"endpoint" json:"endpoint"`
	Interval  string `yaml:"interval" json:"interval"`
	Retention string `yaml:"retention" json:"retention"`
}

// RetentionPolicy represents data retention policy
type RetentionPolicy struct {
	KeepVersions int    `yaml:"keepVersions" json:"keepVersions"`
	KeepDays     int    `yaml:"keepDays" json:"keepDays"`
	ArchiveOld   bool   `yaml:"archiveOld" json:"archiveOld"`
	ArchivePath  string `yaml:"archivePath,omitempty" json:"archivePath,omitempty"`
}

// MarketplaceIndex represents the marketplace index
type MarketplaceIndex struct {
	Version    string              `json:"version"`
	Generated  time.Time           `json:"generated"`
	Total      int                 `json:"total"`
	Templates  map[string]Template `json:"templates"`
	Categories map[string]int      `json:"categories"`
	Stats      IndexStats          `json:"stats"`
}

// Template represents a template in the index
type Template struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Category     string            `json:"category"`
	Type         string            `json:"type"`
	Keywords     []string          `json:"keywords"`
	Downloads    int               `json:"downloads"`
	Rating       float64           `json:"rating"`
	Created      time.Time         `json:"created"`
	Updated      time.Time         `json:"updated"`
	Dependencies []string          `json:"dependencies"`
	Size         int64             `json:"size"`
	License      string            `json:"license"`
	Homepage     string            `json:"homepage,omitempty"`
	Repository   string            `json:"repository,omitempty"`
	Verified     bool              `json:"verified"`
	Deprecated   bool              `json:"deprecated"`
	Tags         []string          `json:"tags"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// IndexStats represents index statistics
type IndexStats struct {
	TotalTemplates   int            `json:"totalTemplates"`
	TotalDownloads   int            `json:"totalDownloads"`
	TotalSize        int64          `json:"totalSize"`
	CategorieCount   map[string]int `json:"categoriesCount"`
	TypesCount       map[string]int `json:"typesCount"`
	PopularTemplates []Template     `json:"popularTemplates"`
	RecentTemplates  []Template     `json:"recentTemplates"`
	TopAuthors       []AuthorStats  `json:"topAuthors"`
}

// AuthorStats represents author statistics
type AuthorStats struct {
	Author         string  `json:"author"`
	TemplateCount  int     `json:"templateCount"`
	TotalDownloads int     `json:"totalDownloads"`
	AverageRating  float64 `json:"averageRating"`
}

func runMarketplace(cmd *cobra.Command, args []string) {
	fmt.Printf("🏪 템플릿 마켓플레이스 관리\n")
	fmt.Printf("📁 경로: %s\n", marketplacePath)
	fmt.Printf("🏷️  타입: %s\n", marketplaceType)

	// Initialize marketplace
	if err := initializeMarketplace(); err != nil {
		fmt.Printf("❌ 마켓플레이스 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	// Generate index
	if err := generateMarketplaceIndex(); err != nil {
		fmt.Printf("❌ 인덱스 생성 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 마켓플레이스 초기화 완료\n")
}

func initializeMarketplace() error {
	fmt.Printf("🏗️ 마켓플레이스 초기화 중...\n")

	// Create marketplace directory structure
	if err := createMarketplaceStructure(); err != nil {
		return err
	}

	// Generate configuration
	if err := generateMarketplaceConfig(); err != nil {
		return err
	}

	// Initialize storage
	if err := initializeStorage(); err != nil {
		return err
	}

	return nil
}

func createMarketplaceStructure() error {
	dirs := []string{
		marketplacePath,
		filepath.Join(marketplacePath, "templates"),
		filepath.Join(marketplacePath, "index"),
		filepath.Join(marketplacePath, "metadata"),
		filepath.Join(marketplacePath, "config"),
		filepath.Join(marketplacePath, "logs"),
		filepath.Join(marketplacePath, "cache"),
		filepath.Join(marketplacePath, "backup"),
	}

	if marketplaceType == "private" {
		dirs = append(dirs, []string{
			filepath.Join(marketplacePath, "auth"),
			filepath.Join(marketplacePath, "approval"),
			filepath.Join(marketplacePath, "audit"),
		}...)
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("디렉터리 생성 실패 %s: %w", dir, err)
		}
	}

	fmt.Printf("📁 마켓플레이스 구조 생성 완료\n")
	return nil
}

func generateMarketplaceConfig() error {
	config := MarketplaceConfig{
		APIVersion: "v1",
		Kind:       "Marketplace",
	}

	config.Metadata.Name = "gz-template-marketplace"
	config.Metadata.Description = fmt.Sprintf("%s 템플릿 마켓플레이스", marketplaceType)
	config.Metadata.Created = time.Now()
	config.Metadata.Updated = time.Now()

	// Configure spec based on type
	config.Spec = generateMarketplaceSpec()

	// Write config file
	configFile := filepath.Join(marketplacePath, "config", "marketplace.yaml")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 마샬링 실패: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0o644); err != nil {
		return fmt.Errorf("설정 파일 쓰기 실패: %w", err)
	}

	fmt.Printf("⚙️  설정 파일 생성: %s\n", configFile)
	return nil
}

func generateMarketplaceSpec() MarketplaceSpec {
	spec := MarketplaceSpec{
		Type: marketplaceType,
		Storage: StorageConfig{
			Type: "filesystem",
			Path: filepath.Join(marketplacePath, "templates"),
		},
		Authentication: AuthConfig{
			Provider: authProvider,
		},
		Versioning: VersioningConfig{
			Enabled:       enableVersioning,
			Strategy:      "semantic",
			MaxVersions:   10,
			AutoCleanup:   true,
			ImmutableTags: true,
		},
		Indexing: IndexingConfig{
			Enabled:       true,
			RefreshRate:   "1h",
			SearchFields:  []string{"name", "description", "keywords", "author"},
			FacetFields:   []string{"category", "type", "author", "license"},
			FullTextIndex: true,
		},
		Metrics: MetricsConfig{
			Enabled:   enableMetrics,
			Provider:  "prometheus",
			Interval:  "30s",
			Retention: "30d",
		},
		Categories: []string{
			"web", "database", "infrastructure", "cicd",
			"monitoring", "security", "analytics", "general",
		},
		DefaultLicense:  "MIT",
		MaxTemplateSize: 100 * 1024 * 1024, // 100MB
		RetentionPolicy: RetentionPolicy{
			KeepVersions: 5,
			KeepDays:     365,
			ArchiveOld:   true,
		},
	}

	// Configure based on marketplace type
	if marketplaceType == "private" {
		spec.Authorization = AuthzConfig{
			Enabled:     true,
			AdminUsers:  adminUsers,
			DefaultRole: "user",
			Roles: []MarketplaceRole{
				{
					Name:        "admin",
					Description: "관리자 역할",
					Permissions: []string{"template.create", "template.update", "template.delete", "template.approve", "marketplace.admin"},
				},
				{
					Name:        "publisher",
					Description: "퍼블리셔 역할",
					Permissions: []string{"template.create", "template.update", "template.publish"},
				},
				{
					Name:        "user",
					Description: "일반 사용자 역할",
					Permissions: []string{"template.read", "template.download"},
				},
			},
		}

		spec.Approval = ApprovalConfig{
			Enabled:      enableApproval,
			RequiredFor:  []string{"create", "update", "publish"},
			MinApprovals: 1,
			TimeoutHours: 72,
		}
	}

	// Override categories if specified
	if len(allowedCategories) > 0 {
		spec.Categories = allowedCategories
	}

	return spec
}

func initializeStorage() error {
	// Create initial index structure
	indexDir := filepath.Join(marketplacePath, "index")

	initialIndex := MarketplaceIndex{
		Version:    "1.0.0",
		Generated:  time.Now(),
		Total:      0,
		Templates:  make(map[string]Template),
		Categories: make(map[string]int),
		Stats: IndexStats{
			CategorieCount:   make(map[string]int),
			TypesCount:       make(map[string]int),
			PopularTemplates: []Template{},
			RecentTemplates:  []Template{},
			TopAuthors:       []AuthorStats{},
		},
	}

	indexFile := filepath.Join(indexDir, "index.json")
	data, err := json.MarshalIndent(initialIndex, "", "  ")
	if err != nil {
		return fmt.Errorf("인덱스 마샬링 실패: %w", err)
	}

	if err := os.WriteFile(indexFile, data, 0o644); err != nil {
		return fmt.Errorf("인덱스 파일 쓰기 실패: %w", err)
	}

	fmt.Printf("📇 초기 인덱스 생성: %s\n", indexFile)
	return nil
}

func generateMarketplaceIndex() error {
	fmt.Printf("📇 마켓플레이스 인덱스 생성 중...\n")

	// Scan templates directory
	templatesDir := filepath.Join(marketplacePath, "templates")
	templates := make(map[string]Template)
	categories := make(map[string]int)

	// Walk through templates
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, "template.yaml") {
			return nil
		}

		// Load template metadata
		template, err := loadTemplateFromPath(path)
		if err != nil {
			fmt.Printf("⚠️  템플릿 로드 실패 %s: %v\n", path, err)
			return nil
		}

		templates[template.Name] = template
		categories[template.Category]++
		return nil
	})
	if err != nil {
		return fmt.Errorf("템플릿 스캔 실패: %w", err)
	}

	// Generate statistics
	stats := generateIndexStats(templates)

	// Create index
	index := MarketplaceIndex{
		Version:    "1.0.0",
		Generated:  time.Now(),
		Total:      len(templates),
		Templates:  templates,
		Categories: categories,
		Stats:      stats,
	}

	// Write index file
	indexFile := filepath.Join(marketplacePath, "index", "index.json")
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("인덱스 마샬링 실패: %w", err)
	}

	if err := os.WriteFile(indexFile, data, 0o644); err != nil {
		return fmt.Errorf("인덱스 파일 쓰기 실패: %w", err)
	}

	fmt.Printf("📇 인덱스 생성 완료: %d개 템플릿\n", len(templates))
	return nil
}

func loadTemplateFromPath(metadataPath string) (Template, error) {
	var template Template

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return template, err
	}

	// Parse metadata (simplified)
	var metadata TemplateMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return template, err
	}

	// Convert to index template
	template = Template{
		Name:        metadata.Metadata.Name,
		Version:     metadata.Metadata.Version,
		Description: metadata.Metadata.Description,
		Author:      metadata.Metadata.Author,
		Category:    metadata.Metadata.Category,
		Type:        metadata.Metadata.Type,
		Keywords:    metadata.Metadata.Keywords,
		License:     metadata.Metadata.License,
		Homepage:    metadata.Metadata.Homepage,
		Repository:  metadata.Metadata.Repository,
		Created:     time.Now(), // In real implementation, parse from metadata
		Updated:     time.Now(),
		Tags:        metadata.Metadata.Tags,
	}

	return template, nil
}

func generateIndexStats(templates map[string]Template) IndexStats {
	stats := IndexStats{
		TotalTemplates:   len(templates),
		CategorieCount:   make(map[string]int),
		TypesCount:       make(map[string]int),
		PopularTemplates: []Template{},
		RecentTemplates:  []Template{},
		TopAuthors:       []AuthorStats{},
	}

	// Count by categories and types
	for _, template := range templates {
		stats.CategorieCount[template.Category]++
		stats.TypesCount[template.Type]++
		stats.TotalDownloads += template.Downloads
		stats.TotalSize += template.Size
	}

	return stats
}

// Additional commands for search, install, publish will be added in separate files
