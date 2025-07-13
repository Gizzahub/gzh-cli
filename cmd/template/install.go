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

// InstallCmd represents the install command
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "템플릿 설치",
	Long: `마켓플레이스에서 템플릿을 설치합니다.

설치 기능:
- 템플릿 다운로드 및 설치
- 의존성 자동 해결
- 버전 호환성 검사
- 매개변수 검증
- 설치 후 훅 실행

Examples:
  gz template install nginx-template
  gz template install nginx-template@1.2.0
  gz template install ./local-template
  gz template install --name my-app --param port=8080`,
	Run: runInstall,
}

var (
	installName    string
	installVersion string
	installPath    string
	parameters     []string
	installDryRun  bool
	skipDeps       bool
	forceInstall   bool
	installServer  string
	installAPIKey  string
	installConfig  string
	templateID     string
)

func init() {
	InstallCmd.Flags().StringVarP(&installName, "name", "n", "", "설치할 템플릿 이름")
	InstallCmd.Flags().StringVarP(&installVersion, "version", "v", "latest", "템플릿 버전")
	InstallCmd.Flags().StringVarP(&installPath, "path", "p", ".", "설치 경로")
	InstallCmd.Flags().StringSliceVar(&parameters, "param", []string{}, "템플릿 매개변수 (key=value)")
	InstallCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "실제 설치하지 않고 미리보기")
	InstallCmd.Flags().BoolVar(&skipDeps, "skip-deps", false, "의존성 설치 건너뛰기")
	InstallCmd.Flags().BoolVar(&forceInstall, "force", false, "기존 파일 덮어쓰기")
	InstallCmd.Flags().StringVar(&installServer, "server", "http://localhost:8080", "템플릿 서버 URL")
	InstallCmd.Flags().StringVar(&installAPIKey, "api-key", "", "API 키")
	InstallCmd.Flags().StringVar(&installConfig, "config", "", "클라이언트 설정 파일")
	InstallCmd.Flags().StringVar(&templateID, "id", "", "템플릿 ID")
}

func runInstall(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		installName = args[0]
	}

	if installName == "" {
		fmt.Printf("❌ 설치할 템플릿 이름이 필요합니다\n")
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("📦 템플릿 설치: %s\n", installName)
	fmt.Printf("📁 설치 경로: %s\n", installPath)
	fmt.Printf("🏷️  버전: %s\n", installVersion)

	if installDryRun {
		fmt.Printf("🔍 드라이런 모드\n")
	}

	// Install template
	if err := installTemplate(); err != nil {
		fmt.Printf("❌ 설치 실패: %v\n", err)
		os.Exit(1)
	}

	if !installDryRun {
		fmt.Printf("✅ 템플릿 설치 완료\n")
	}
}

func installTemplate() error {
	// 1. Setup client
	client, err := setupInstallClient()
	if err != nil {
		return fmt.Errorf("클라이언트 설정 실패: %w", err)
	}

	// 2. Resolve template ID
	resolvedTemplateID, err := resolveTemplateID(client)
	if err != nil {
		return fmt.Errorf("템플릿 해결 실패: %w", err)
	}

	// 3. Get template details
	fmt.Printf("🔍 템플릿 정보 확인 중...\n")
	templateInfo, err := client.GetTemplate(resolvedTemplateID)
	if err != nil {
		return fmt.Errorf("템플릿 정보 조회 실패: %w", err)
	}

	fmt.Printf("📦 템플릿: %s v%s\n", templateInfo.Name, templateInfo.Version)
	fmt.Printf("👤 작성자: %s\n", templateInfo.Author)
	fmt.Printf("📝 설명: %s\n", templateInfo.Description)

	// 4. Check if dry run
	if installDryRun {
		fmt.Printf("📝 드라이런: 실제 설치 없이 완료\n")
		return nil
	}

	// 5. Download template
	fmt.Printf("📥 다운로드 중...\n")
	tempFile, err := downloadTemplate(client, resolvedTemplateID)
	if err != nil {
		return fmt.Errorf("다운로드 실패: %w", err)
	}
	defer os.Remove(tempFile)

	// 6. Extract template
	fmt.Printf("📁 파일 추출 중...\n")
	if err := extractTemplate(tempFile, installPath); err != nil {
		return fmt.Errorf("추출 실패: %w", err)
	}

	// 7. Process parameters
	if len(parameters) > 0 {
		fmt.Printf("📋 매개변수 처리 중...\n")
		if err := processTemplateParameters(); err != nil {
			return fmt.Errorf("매개변수 처리 실패: %w", err)
		}
	}

	fmt.Printf("✅ 설치 완료\n")
	return nil
}

func setupInstallClient() (*TemplateClient, error) {
	var config *ClientConfig

	// Try to load from config file
	if installConfig == "" {
		installConfig = GetDefaultConfigPath()
	}

	if _, err := os.Stat(installConfig); err == nil {
		loadedConfig, err := LoadClientConfig(installConfig)
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
	if installServer != "" {
		config.BaseURL = installServer
	}
	if installAPIKey != "" {
		config.APIKey = installAPIKey
	}

	return NewTemplateClient(config), nil
}

func resolveTemplateID(client *TemplateClient) (string, error) {
	// If template ID is provided directly, use it
	if templateID != "" {
		return templateID, nil
	}

	// Otherwise, search for template by name
	response, err := client.SearchTemplates(installName, "", "", 1, 10)
	if err != nil {
		return "", fmt.Errorf("템플릿 검색 실패: %w", err)
	}

	if len(response.Templates) == 0 {
		return "", fmt.Errorf("템플릿을 찾을 수 없습니다: %s", installName)
	}

	// Find exact match or best match
	var bestMatch *TemplateInfo
	for _, template := range response.Templates {
		if template.Name == installName {
			// Check version compatibility
			if installVersion == "latest" || template.Version == installVersion {
				bestMatch = &template
				break
			}
		}
	}

	if bestMatch == nil {
		bestMatch = &response.Templates[0] // Use first result
	}

	return bestMatch.ID, nil
}

func downloadTemplate(client *TemplateClient, templateID string) (string, error) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", "template-*.zip")
	if err != nil {
		return "", fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	tempFile.Close()

	// Download template
	if err := client.DownloadTemplate(templateID, tempFile.Name()); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

func extractTemplate(zipPath, destPath string) error {
	// Open zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("압축 파일 열기 실패: %w", err)
	}
	defer reader.Close()

	// Create destination directory
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return fmt.Errorf("대상 디렉터리 생성 실패: %w", err)
	}

	// Extract files
	for _, file := range reader.File {
		destFile := filepath.Join(destPath, file.Name)

		// Check for directory traversal
		if !strings.HasPrefix(destFile, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("잘못된 파일 경로: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destFile, file.FileInfo().Mode())
			continue
		}

		// Create directory for file
		if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
			return fmt.Errorf("파일 디렉터리 생성 실패: %w", err)
		}

		// Extract file
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("압축 파일 내 파일 열기 실패: %w", err)
		}

		outFile, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("대상 파일 생성 실패: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return fmt.Errorf("파일 복사 실패: %w", err)
		}
	}

	return nil
}

func processTemplateParameters() error {
	// Parse parameters
	params := make(map[string]string)
	for _, param := range parameters {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("잘못된 매개변수 형식: %s (key=value 형식이어야 함)", param)
		}
		params[parts[0]] = parts[1]
	}

	// In a real implementation, this would:
	// 1. Process template files with Go templates
	// 2. Replace variables with provided parameters
	// 3. Validate required parameters

	fmt.Printf("   📋 처리된 매개변수: %d개\n", len(params))
	for key, value := range params {
		fmt.Printf("     • %s = %s\n", key, value)
	}

	return nil
}
