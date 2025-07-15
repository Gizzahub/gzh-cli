package template

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// PublishCmd represents the publish command
var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "템플릿 퍼블리시",
	Long: `템플릿을 마켓플레이스에 퍼블리시합니다.

퍼블리시 기능:
- 템플릿 메타데이터 검증
- 파일 패키징
- 버전 관리 및 태깅
- 의존성 검사
- 마켓플레이스 업로드
- 승인 워크플로우 처리

Examples:
  gz template publish
  gz template publish --path ./my-template
  gz template publish --registry private
  gz template publish --draft`,
	Run: runPublish,
}

var (
	publishPath     string
	publishRegistry string
	publishDraft    bool
	publishTag      string
	publishMessage  string
	skipValidation  bool
	autoApprove     bool
	publishAuthor   string
	publishServer   string
	publishAPIKey   string
	configPath      string
)

func init() {
	PublishCmd.Flags().StringVarP(&publishPath, "path", "p", ".", "퍼블리시할 템플릿 경로")
	PublishCmd.Flags().StringVarP(&publishRegistry, "registry", "r", "default", "대상 레지스트리")
	PublishCmd.Flags().BoolVar(&publishDraft, "draft", false, "드래프트로 퍼블리시")
	PublishCmd.Flags().StringVarP(&publishTag, "tag", "t", "", "버전 태그")
	PublishCmd.Flags().StringVarP(&publishMessage, "message", "m", "", "퍼블리시 메시지")
	PublishCmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "검증 건너뛰기")
	PublishCmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "자동 승인 요청")
	PublishCmd.Flags().StringVarP(&publishAuthor, "author", "a", "", "템플릿 작성자")
	PublishCmd.Flags().StringVar(&publishServer, "server", "http://localhost:8080", "템플릿 서버 URL")
	PublishCmd.Flags().StringVar(&publishAPIKey, "api-key", "", "API 키")
	PublishCmd.Flags().StringVar(&configPath, "config", "", "클라이언트 설정 파일")
}

func runPublish(cmd *cobra.Command, args []string) {
	fmt.Printf("📤 템플릿 퍼블리시\n")
	fmt.Printf("📁 경로: %s\n", publishPath)
	fmt.Printf("🏪 레지스트리: %s\n", publishRegistry)

	if publishDraft {
		fmt.Printf("📝 드래프트 모드\n")
	}

	// Publish template
	if err := publishTemplate(); err != nil {
		fmt.Printf("❌ 퍼블리시 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 템플릿 퍼블리시 완료\n")
}

func publishTemplate() error {
	// 1. Load client configuration
	client, err := setupClient()
	if err != nil {
		return fmt.Errorf("클라이언트 설정 실패: %w", err)
	}

	// 2. Validate template if not skipped
	if !skipValidation {
		fmt.Printf("🔍 템플릿 검증 중...\n")
		if err := validateTemplateForPublish(); err != nil {
			return fmt.Errorf("템플릿 검증 실패: %w", err)
		}
		fmt.Printf("✅ 템플릿 검증 완료\n")
	}

	// 3. Package template files
	fmt.Printf("📦 패키징 중...\n")
	packagePath, err := packageTemplate()
	if err != nil {
		return fmt.Errorf("패키징 실패: %w", err)
	}
	defer os.Remove(packagePath) // Clean up
	fmt.Printf("✅ 패키징 완료: %s\n", packagePath)

	// 4. Upload to server
	fmt.Printf("📤 업로드 중...\n")
	author := getAuthor()
	response, err := client.UploadTemplate(packagePath, author)
	if err != nil {
		return fmt.Errorf("업로드 실패: %w", err)
	}

	// 5. Display results
	fmt.Printf("✅ 업로드 완료\n")
	fmt.Printf("🆔 템플릿 ID: %s\n", response.TemplateID)
	fmt.Printf("💬 메시지: %s\n", response.Message)

	if response.ApprovalID != "" {
		fmt.Printf("🔄 승인 대기 중 (ID: %s)\n", response.ApprovalID)
	}

	return nil
}

func setupClient() (*TemplateClient, error) {
	var config *ClientConfig

	// Try to load from config file
	if configPath == "" {
		configPath = GetDefaultConfigPath()
	}

	if _, err := os.Stat(configPath); err == nil {
		loadedConfig, err := LoadClientConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("설정 파일 로드 실패: %w", err)
		}
		config = loadedConfig
	} else {
		// Create default config
		config = &ClientConfig{
			BaseURL: "http://localhost:8080",
			Timeout: 30,
		}
	}

	// Override with command line flags
	if publishServer != "" {
		config.BaseURL = publishServer
	}
	if publishAPIKey != "" {
		config.APIKey = publishAPIKey
	}

	return NewTemplateClient(config), nil
}

func validateTemplateForPublish() error {
	// Check if template.yaml exists
	metadataFile := filepath.Join(publishPath, "template.yaml")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		return fmt.Errorf("template.yaml 파일이 없습니다: %s", metadataFile)
	}

	// Validate required directories
	requiredDirs := []string{"templates"}
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(publishPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			return fmt.Errorf("필수 디렉터리가 없습니다: %s", dir)
		}
	}

	return nil
}

func packageTemplate() (string, error) {
	// Create temporary zip file
	tempFile, err := os.CreateTemp("", "template-*.zip")
	if err != nil {
		return "", fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	defer tempFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Walk through template directory
	err = filepath.Walk(publishPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(publishPath, path)
		if err != nil {
			return err
		}

		// Skip directories themselves
		if info.IsDir() {
			return nil
		}

		// Add file to zip
		fileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(fileWriter, file)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("파일 압축 실패: %w", err)
	}

	return tempFile.Name(), nil
}

func getPublishAuthor() string {
	if publishAuthor != "" {
		return publishAuthor
	}

	// Try to get from git config
	// In a real implementation, you might use git commands
	return "unknown"
}
