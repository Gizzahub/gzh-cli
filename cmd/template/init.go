package template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// InitCmd represents the init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "템플릿 저장소 초기화",
	Long: `새로운 템플릿 저장소를 초기화합니다.

템플릿 저장소 구조:
- 메타데이터 파일 (template.yaml)
- 템플릿 파일들 (templates/ 디렉터리)
- 문서 파일들 (docs/ 디렉터리)
- 예제 파일들 (examples/ 디렉터리)
- 테스트 파일들 (tests/ 디렉터리)

Examples:
  gz template init --name my-template --type docker
  gz template init --name web-app --type helm --category web
  gz template init --path ./custom-template`,
	Run: runInit,
}

var (
	templateName     string
	templateType     string
	templateCategory string
	templatePath     string
	author           string
	description      string
	license          string
	force            bool
)

func init() {
	InitCmd.Flags().StringVarP(&templateName, "name", "n", "", "템플릿 이름")
	InitCmd.Flags().StringVarP(&templateType, "type", "t", "generic", "템플릿 타입 (docker, helm, terraform, ansible, github-actions)")
	InitCmd.Flags().StringVarP(&templateCategory, "category", "c", "general", "템플릿 카테고리")
	InitCmd.Flags().StringVarP(&templatePath, "path", "p", "", "템플릿 저장소 경로")
	InitCmd.Flags().StringVarP(&author, "author", "a", "", "템플릿 작성자")
	InitCmd.Flags().StringVarP(&description, "description", "d", "", "템플릿 설명")
	InitCmd.Flags().StringVar(&license, "license", "MIT", "라이선스")
	InitCmd.Flags().BoolVar(&force, "force", false, "기존 디렉터리 덮어쓰기")
}

// TemplateMetadata represents the template metadata structure
type TemplateMetadata struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string            `yaml:"name"`
		Version     string            `yaml:"version"`
		Description string            `yaml:"description"`
		Author      string            `yaml:"author"`
		License     string            `yaml:"license"`
		Homepage    string            `yaml:"homepage,omitempty"`
		Repository  string            `yaml:"repository,omitempty"`
		Keywords    []string          `yaml:"keywords,omitempty"`
		Category    string            `yaml:"category"`
		Type        string            `yaml:"type"`
		Tags        []string          `yaml:"tags,omitempty"`
		Labels      map[string]string `yaml:"labels,omitempty"`
		Created     string            `yaml:"created"`
		Updated     string            `yaml:"updated"`
	} `yaml:"metadata"`
	Spec TemplateSpec `yaml:"spec"`
}

// TemplateSpec represents the template specification
type TemplateSpec struct {
	MinVersion    string                 `yaml:"minVersion,omitempty"`
	MaxVersion    string                 `yaml:"maxVersion,omitempty"`
	Dependencies  []Dependency           `yaml:"dependencies,omitempty"`
	Parameters    []Parameter            `yaml:"parameters,omitempty"`
	Files         []FileEntry            `yaml:"files"`
	Hooks         map[string][]Hook      `yaml:"hooks,omitempty"`
	Configuration map[string]interface{} `yaml:"configuration,omitempty"`
	Requirements  Requirements           `yaml:"requirements,omitempty"`
}

// Dependency represents a template dependency
type Dependency struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Repository  string `yaml:"repository,omitempty"`
	Optional    bool   `yaml:"optional,omitempty"`
	Condition   string `yaml:"condition,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Parameter represents a template parameter
type Parameter struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default,omitempty"`
	Required    bool        `yaml:"required,omitempty"`
	Pattern     string      `yaml:"pattern,omitempty"`
	Enum        []string    `yaml:"enum,omitempty"`
	Min         *int        `yaml:"min,omitempty"`
	Max         *int        `yaml:"max,omitempty"`
}

// FileEntry represents a template file
type FileEntry struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Template    bool   `yaml:"template,omitempty"`
	Executable  bool   `yaml:"executable,omitempty"`
	Condition   string `yaml:"condition,omitempty"`
}

// Hook represents a template hook
type Hook struct {
	Name        string            `yaml:"name"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	WorkingDir  string            `yaml:"workingDir,omitempty"`
	Condition   string            `yaml:"condition,omitempty"`
}

// Requirements represents template requirements
type Requirements struct {
	Platform     []string          `yaml:"platform,omitempty"`
	Tools        []string          `yaml:"tools,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty"`
	MinResources ResourceLimits    `yaml:"minResources,omitempty"`
}

// ResourceLimits represents resource requirements
type ResourceLimits struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
	Disk   string `yaml:"disk,omitempty"`
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Printf("🏗️ 템플릿 저장소 초기화\n")

	if templateName == "" {
		fmt.Printf("❌ 템플릿 이름이 필요합니다 (--name)\n")
		os.Exit(1)
	}

	// Determine template path
	if templatePath == "" {
		templatePath = templateName
	}

	// Check if directory exists
	if _, err := os.Stat(templatePath); err == nil && !force {
		fmt.Printf("❌ 디렉터리가 이미 존재합니다: %s (--force로 강제 생성 가능)\n", templatePath)
		os.Exit(1)
	}

	fmt.Printf("📁 경로: %s\n", templatePath)
	fmt.Printf("📦 타입: %s\n", templateType)
	fmt.Printf("📂 카테고리: %s\n", templateCategory)

	// Create template repository structure
	if err := createTemplateStructure(); err != nil {
		fmt.Printf("❌ 템플릿 구조 생성 실패: %v\n", err)
		os.Exit(1)
	}

	// Generate metadata file
	if err := generateMetadata(); err != nil {
		fmt.Printf("❌ 메타데이터 생성 실패: %v\n", err)
		os.Exit(1)
	}

	// Generate example files
	if err := generateExampleFiles(); err != nil {
		fmt.Printf("❌ 예제 파일 생성 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 템플릿 저장소 초기화 완료\n")
	fmt.Printf("\n📝 다음 단계:\n")
	fmt.Printf("1. 템플릿 파일 추가: %s/templates/\n", templatePath)
	fmt.Printf("2. 메타데이터 수정: %s/template.yaml\n", templatePath)
	fmt.Printf("3. 문서 작성: %s/README.md\n", templatePath)
	fmt.Printf("4. 템플릿 검증: gz template validate\n")
}

func createTemplateStructure() error {
	dirs := []string{
		templatePath,
		filepath.Join(templatePath, "templates"),
		filepath.Join(templatePath, "docs"),
		filepath.Join(templatePath, "examples"),
		filepath.Join(templatePath, "tests"),
		filepath.Join(templatePath, ".gz-template"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("디렉터리 생성 실패 %s: %w", dir, err)
		}
	}

	fmt.Printf("📁 디렉터리 구조 생성 완료\n")
	return nil
}

func generateMetadata() error {
	metadata := TemplateMetadata{
		APIVersion: "v1",
		Kind:       "Template",
	}

	metadata.Metadata.Name = templateName
	metadata.Metadata.Version = "0.1.0"
	metadata.Metadata.Description = getDescription()
	metadata.Metadata.Author = getAuthor()
	metadata.Metadata.License = license
	metadata.Metadata.Category = templateCategory
	metadata.Metadata.Type = templateType
	metadata.Metadata.Created = getCurrentTimestamp()
	metadata.Metadata.Updated = getCurrentTimestamp()

	// Set template-specific configuration
	metadata.Spec = generateTemplateSpec()

	// Write metadata file
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("메타데이터 마샬링 실패: %w", err)
	}

	metadataFile := filepath.Join(templatePath, "template.yaml")
	if err := os.WriteFile(metadataFile, data, 0o644); err != nil {
		return fmt.Errorf("메타데이터 파일 쓰기 실패: %w", err)
	}

	fmt.Printf("📄 메타데이터 파일 생성: %s\n", metadataFile)
	return nil
}

func generateTemplateSpec() TemplateSpec {
	spec := TemplateSpec{
		Files: []FileEntry{},
	}

	switch templateType {
	case "docker":
		spec.Files = []FileEntry{
			{Source: "Dockerfile.tpl", Destination: "Dockerfile", Template: true},
			{Source: "docker-compose.yml.tpl", Destination: "docker-compose.yml", Template: true},
			{Source: ".dockerignore.tpl", Destination: ".dockerignore", Template: true},
		}
		spec.Parameters = []Parameter{
			{Name: "base_image", Type: "string", Description: "베이스 Docker 이미지", Default: "alpine:latest", Required: true},
			{Name: "app_name", Type: "string", Description: "애플리케이션 이름", Required: true},
			{Name: "port", Type: "integer", Description: "노출할 포트", Default: 8080},
		}
		spec.Requirements.Tools = []string{"docker", "docker-compose"}

	case "helm":
		spec.Files = []FileEntry{
			{Source: "Chart.yaml.tpl", Destination: "Chart.yaml", Template: true},
			{Source: "values.yaml.tpl", Destination: "values.yaml", Template: true},
			{Source: "templates/deployment.yaml.tpl", Destination: "templates/deployment.yaml", Template: true},
			{Source: "templates/service.yaml.tpl", Destination: "templates/service.yaml", Template: true},
		}
		spec.Parameters = []Parameter{
			{Name: "chart_name", Type: "string", Description: "Helm 차트 이름", Required: true},
			{Name: "app_version", Type: "string", Description: "애플리케이션 버전", Default: "1.0.0"},
			{Name: "namespace", Type: "string", Description: "Kubernetes 네임스페이스", Default: "default"},
		}
		spec.Requirements.Tools = []string{"helm", "kubectl"}

	case "terraform":
		spec.Files = []FileEntry{
			{Source: "main.tf.tpl", Destination: "main.tf", Template: true},
			{Source: "variables.tf.tpl", Destination: "variables.tf", Template: true},
			{Source: "outputs.tf.tpl", Destination: "outputs.tf", Template: true},
			{Source: "terraform.tfvars.example.tpl", Destination: "terraform.tfvars.example", Template: true},
		}
		spec.Parameters = []Parameter{
			{Name: "provider", Type: "string", Description: "클라우드 제공자", Enum: []string{"aws", "gcp", "azure"}, Required: true},
			{Name: "region", Type: "string", Description: "리전", Required: true},
			{Name: "environment", Type: "string", Description: "환경", Default: "development"},
		}
		spec.Requirements.Tools = []string{"terraform"}

	case "ansible":
		spec.Files = []FileEntry{
			{Source: "playbook.yml.tpl", Destination: "playbook.yml", Template: true},
			{Source: "inventory.ini.tpl", Destination: "inventory.ini", Template: true},
			{Source: "roles/main/tasks/main.yml.tpl", Destination: "roles/main/tasks/main.yml", Template: true},
		}
		spec.Parameters = []Parameter{
			{Name: "playbook_name", Type: "string", Description: "플레이북 이름", Required: true},
			{Name: "target_hosts", Type: "string", Description: "대상 호스트 그룹", Default: "all"},
			{Name: "become_user", Type: "string", Description: "권한 상승 사용자", Default: "root"},
		}
		spec.Requirements.Tools = []string{"ansible", "ansible-playbook"}

	case "github-actions":
		spec.Files = []FileEntry{
			{Source: ".github/workflows/ci.yml.tpl", Destination: ".github/workflows/ci.yml", Template: true},
			{Source: ".github/workflows/cd.yml.tpl", Destination: ".github/workflows/cd.yml", Template: true},
		}
		spec.Parameters = []Parameter{
			{Name: "workflow_name", Type: "string", Description: "워크플로우 이름", Required: true},
			{Name: "trigger_branches", Type: "array", Description: "트리거 브랜치", Default: []string{"main", "develop"}},
			{Name: "node_version", Type: "string", Description: "Node.js 버전", Default: "18"},
		}

	default:
		spec.Files = []FileEntry{
			{Source: "README.md.tpl", Destination: "README.md", Template: true},
			{Source: "config.yaml.tpl", Destination: "config.yaml", Template: true},
		}
		spec.Parameters = []Parameter{
			{Name: "project_name", Type: "string", Description: "프로젝트 이름", Required: true},
			{Name: "version", Type: "string", Description: "버전", Default: "0.1.0"},
		}
	}

	return spec
}

func generateExampleFiles() error {
	// Generate README.md
	readmeContent := generateReadmeContent()
	readmeFile := filepath.Join(templatePath, "README.md")
	if err := os.WriteFile(readmeFile, []byte(readmeContent), 0o644); err != nil {
		return fmt.Errorf("README 파일 생성 실패: %w", err)
	}

	// Generate example template files
	if err := generateExampleTemplates(); err != nil {
		return err
	}

	// Generate test file
	testContent := generateTestContent()
	testFile := filepath.Join(templatePath, "tests", "template_test.yaml")
	if err := os.WriteFile(testFile, []byte(testContent), 0o644); err != nil {
		return fmt.Errorf("테스트 파일 생성 실패: %w", err)
	}

	fmt.Printf("📝 예제 파일 생성 완료\n")
	return nil
}

func generateReadmeContent() string {
	return fmt.Sprintf("# %s\n\n%s\n\n## 설명\n\n%s 타입의 템플릿입니다.\n\n## 사용법\n\n```bash\n# 템플릿 설치\ngz template install %s\n\n# 템플릿 적용\ngz template apply %s --param key=value\n```\n\n## 매개변수\n\n템플릿 매개변수는 template.yaml 파일에서 확인할 수 있습니다.\n\n## 요구사항\n\n- gz CLI 도구\n- %s 관련 도구들\n\n## 라이선스\n\n%s\n\n## 작성자\n\n%s\n", templateName, getDescription(), templateType, templateName, templateName, templateType, license, getAuthor())
}

func generateExampleTemplates() error {
	templatesDir := filepath.Join(templatePath, "templates")

	switch templateType {
	case "docker":
		dockerfileContent := "FROM {{ .base_image }}\n\nWORKDIR /app\n\nCOPY . .\n\n{{ if .port }}\nEXPOSE {{ .port }}\n{{ end }}\n\nCMD [\"./{{ .app_name }}\"]"

		if err := os.WriteFile(filepath.Join(templatesDir, "Dockerfile.tpl"), []byte(dockerfileContent), 0o644); err != nil {
			return err
		}

		composeContent := "version: '3.8'\n\nservices:\n  {{ .app_name }}:\n    build: .\n    ports:\n      - \"{{ .port }}:{{ .port }}\"\n    environment:\n      - NODE_ENV=production"

		if err := os.WriteFile(filepath.Join(templatesDir, "docker-compose.yml.tpl"), []byte(composeContent), 0o644); err != nil {
			return err
		}

	case "helm":
		chartContent := "apiVersion: v2\nname: {{ .chart_name }}\ndescription: A Helm chart for {{ .chart_name }}\ntype: application\nversion: 0.1.0\nappVersion: \"{{ .app_version }}\""

		if err := os.MkdirAll(filepath.Join(templatesDir, "templates"), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(templatesDir, "Chart.yaml.tpl"), []byte(chartContent), 0o644); err != nil {
			return err
		}

	default:
		// Generic template
		genericContent := "# {{ .project_name }}\n\nVersion: {{ .version }}\n\nThis is a generic template file."

		if err := os.WriteFile(filepath.Join(templatesDir, "README.md.tpl"), []byte(genericContent), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func generateTestContent() string {
	return fmt.Sprintf("# Template Test Configuration\napiVersion: v1\nkind: TemplateTest\nmetadata:\n  name: %s-test\nspec:\n  scenarios:\n    - name: default-parameters\n      description: Test with default parameters\n      parameters:\n        app_name: test-app\n      expected:\n        files:\n          - path: README.md\n            exists: true\n    - name: custom-parameters\n      description: Test with custom parameters\n      parameters:\n        app_name: custom-app\n        version: 1.0.0\n      expected:\n        files:\n          - path: README.md\n            exists: true\n            contains: [\"custom-app\", \"1.0.0\"]\n", templateName)
}

func getDescription() string {
	if description != "" {
		return description
	}
	return fmt.Sprintf("%s용 템플릿", templateType)
}

func getAuthor() string {
	if author != "" {
		return author
	}
	return "gz-template"
}

func getCurrentTimestamp() string {
	return "2025-01-13T00:00:00Z" // In real implementation, use time.Now()
}
