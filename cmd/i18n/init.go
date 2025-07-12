package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/i18n"
	"github.com/spf13/cobra"
)

// InitCmd represents the init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize i18n configuration and locale files",
	Long: `Initialize internationalization configuration and create default locale files.

This command sets up the i18n system with default configurations and creates
skeleton locale files for supported languages.

Examples:
  gz i18n init --locales ./locales --languages en,ko,ja
  gz i18n init --default en --fallback en`,
	Run: runInit,
}

var (
	localesDir       string
	defaultLanguage  string
	fallbackLanguage string
	languages        []string
	force            bool
)

func init() {
	InitCmd.Flags().StringVar(&localesDir, "locales", "locales", "Directory for locale files")
	InitCmd.Flags().StringVar(&defaultLanguage, "default", "en", "Default language")
	InitCmd.Flags().StringVar(&fallbackLanguage, "fallback", "en", "Fallback language")
	InitCmd.Flags().StringSliceVar(&languages, "languages", []string{"en", "ko", "ja", "zh"}, "Supported languages")
	InitCmd.Flags().BoolVar(&force, "force", false, "Force overwrite existing files")
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("🌍 Initializing i18n configuration...")

	// Validate languages
	if defaultLanguage == "" {
		fmt.Println("❌ Default language cannot be empty")
		os.Exit(1)
	}

	if fallbackLanguage == "" {
		fmt.Println("❌ Fallback language cannot be empty")
		os.Exit(1)
	}

	// Check if default language is in supported languages
	var hasDefault bool
	for _, lang := range languages {
		if lang == defaultLanguage {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		languages = append([]string{defaultLanguage}, languages...)
	}

	fmt.Printf("📁 Locales directory: %s\n", localesDir)
	fmt.Printf("🏠 Default language: %s\n", defaultLanguage)
	fmt.Printf("🔄 Fallback language: %s\n", fallbackLanguage)
	fmt.Printf("🌐 Supported languages: %s\n", strings.Join(languages, ", "))

	// Create locales directory
	if err := os.MkdirAll(localesDir, 0o755); err != nil {
		fmt.Printf("❌ Failed to create locales directory: %v\n", err)
		os.Exit(1)
	}

	// Create i18n configuration
	config := &i18n.Config{
		LocalesDir:         localesDir,
		DefaultLanguage:    defaultLanguage,
		FallbackLanguage:   fallbackLanguage,
		SupportedLanguages: languages,
	}

	// Initialize manager
	manager, err := i18n.NewManager(config)
	if err != nil {
		fmt.Printf("❌ Failed to initialize i18n manager: %v\n", err)
		os.Exit(1)
	}

	// Generate locale files
	fmt.Println("📝 Generating locale files...")
	for _, lang := range languages {
		if err := generateLocaleFile(localesDir, lang, force); err != nil {
			fmt.Printf("❌ Failed to generate locale file for %s: %v\n", lang, err)
			continue
		}
		fmt.Printf("✅ Generated locale file for %s\n", lang)
	}

	// Create configuration file
	configFile := filepath.Join(localesDir, "config.json")
	if err := createConfigFile(configFile, config, force); err != nil {
		fmt.Printf("❌ Failed to create config file: %v\n", err)
	} else {
		fmt.Printf("✅ Created configuration file: %s\n", configFile)
	}

	// Test the configuration
	fmt.Println("🧪 Testing configuration...")
	testMessage := manager.T(i18n.MsgWelcome)
	if testMessage != "" {
		fmt.Printf("✅ Test successful: %s\n", testMessage)
	} else {
		fmt.Println("⚠️  Test warning: No welcome message found")
	}

	fmt.Println("\n🎉 i18n initialization completed!")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  1. Extract messages from your code: gz i18n extract")
	fmt.Println("  2. Translate messages in the generated locale files")
	fmt.Println("  3. Validate translations: gz i18n validate")
	fmt.Println("  4. Use in your code:")
	fmt.Println("     import \"github.com/gizzahub/gzh-manager-go/pkg/i18n\"")
	fmt.Println("     i18n.Init(config)")
	fmt.Println("     message := i18n.T(\"your.message.key\")")
}

// generateLocaleFile creates a locale file for the specified language
func generateLocaleFile(localesDir, lang string, force bool) error {
	filename := filepath.Join(localesDir, fmt.Sprintf("%s.json", lang))

	// Check if file exists
	if !force {
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("⚠️  Locale file already exists: %s (use --force to overwrite)\n", filename)
			return nil
		}
	}

	// Create locale bundle with common messages
	bundle := createDefaultBundle(lang)

	// Save to file
	return saveLocaleBundle(filename, bundle)
}

// createDefaultBundle creates a default locale bundle for a language
func createDefaultBundle(lang string) *i18n.LocalizationBundle {
	bundle := &i18n.LocalizationBundle{
		Language: lang,
		Version:  "1.0.0",
		Messages: make(map[string]i18n.MessageConfig),
	}

	// Add common messages based on language
	switch lang {
	case "en":
		bundle.Messages = map[string]i18n.MessageConfig{
			i18n.MsgWelcome: {
				ID:      i18n.MsgWelcome,
				Message: "Welcome to GZH Manager",
			},
			i18n.MsgError: {
				ID:      i18n.MsgError,
				Message: "Error",
			},
			i18n.MsgSuccess: {
				ID:      i18n.MsgSuccess,
				Message: "Success",
			},
			i18n.MsgCloneStarting: {
				ID:      i18n.MsgCloneStarting,
				Message: "Starting bulk clone operation",
			},
			i18n.MsgCloneCompleted: {
				ID:      i18n.MsgCloneCompleted,
				Message: "Clone operation completed successfully",
			},
			i18n.MsgCloneStats: {
				ID:      i18n.MsgCloneStats,
				Message: "Cloned {{.Cloned}} repositories, {{.Failed}} failed, {{.Skipped}} skipped",
			},
			i18n.MsgDockerGenerating: {
				ID:      i18n.MsgDockerGenerating,
				Message: "Generating Dockerfile for {{.Language}} project",
			},
			i18n.MsgPluginLoading: {
				ID:      i18n.MsgPluginLoading,
				Message: "Loading plugin: {{.Name}}",
			},
		}

	case "ko":
		bundle.Messages = map[string]i18n.MessageConfig{
			i18n.MsgWelcome: {
				ID:      i18n.MsgWelcome,
				Message: "GZH Manager에 오신 것을 환영합니다",
			},
			i18n.MsgError: {
				ID:      i18n.MsgError,
				Message: "오류",
			},
			i18n.MsgSuccess: {
				ID:      i18n.MsgSuccess,
				Message: "성공",
			},
			i18n.MsgCloneStarting: {
				ID:      i18n.MsgCloneStarting,
				Message: "대량 복제 작업을 시작합니다",
			},
			i18n.MsgCloneCompleted: {
				ID:      i18n.MsgCloneCompleted,
				Message: "복제 작업이 성공적으로 완료되었습니다",
			},
			i18n.MsgCloneStats: {
				ID:      i18n.MsgCloneStats,
				Message: "{{.Cloned}}개 저장소 복제됨, {{.Failed}}개 실패, {{.Skipped}}개 건너뜀",
			},
			i18n.MsgDockerGenerating: {
				ID:      i18n.MsgDockerGenerating,
				Message: "{{.Language}} 프로젝트용 Dockerfile을 생성합니다",
			},
			i18n.MsgPluginLoading: {
				ID:      i18n.MsgPluginLoading,
				Message: "플러그인 로딩 중: {{.Name}}",
			},
		}

	case "ja":
		bundle.Messages = map[string]i18n.MessageConfig{
			i18n.MsgWelcome: {
				ID:      i18n.MsgWelcome,
				Message: "GZH Managerへようこそ",
			},
			i18n.MsgError: {
				ID:      i18n.MsgError,
				Message: "エラー",
			},
			i18n.MsgSuccess: {
				ID:      i18n.MsgSuccess,
				Message: "成功",
			},
			i18n.MsgCloneStarting: {
				ID:      i18n.MsgCloneStarting,
				Message: "一括クローン操作を開始します",
			},
			i18n.MsgCloneCompleted: {
				ID:      i18n.MsgCloneCompleted,
				Message: "クローン操作が正常に完了しました",
			},
			i18n.MsgCloneStats: {
				ID:      i18n.MsgCloneStats,
				Message: "{{.Cloned}}個のリポジトリをクローン、{{.Failed}}個失敗、{{.Skipped}}個スキップ",
			},
			i18n.MsgDockerGenerating: {
				ID:      i18n.MsgDockerGenerating,
				Message: "{{.Language}}プロジェクト用のDockerfileを生成します",
			},
			i18n.MsgPluginLoading: {
				ID:      i18n.MsgPluginLoading,
				Message: "プラグインをロード中: {{.Name}}",
			},
		}

	case "zh", "zh-CN":
		bundle.Messages = map[string]i18n.MessageConfig{
			i18n.MsgWelcome: {
				ID:      i18n.MsgWelcome,
				Message: "欢迎使用 GZH Manager",
			},
			i18n.MsgError: {
				ID:      i18n.MsgError,
				Message: "错误",
			},
			i18n.MsgSuccess: {
				ID:      i18n.MsgSuccess,
				Message: "成功",
			},
			i18n.MsgCloneStarting: {
				ID:      i18n.MsgCloneStarting,
				Message: "开始批量克隆操作",
			},
			i18n.MsgCloneCompleted: {
				ID:      i18n.MsgCloneCompleted,
				Message: "克隆操作成功完成",
			},
			i18n.MsgCloneStats: {
				ID:      i18n.MsgCloneStats,
				Message: "已克隆 {{.Cloned}} 个仓库，{{.Failed}} 个失败，{{.Skipped}} 个跳过",
			},
			i18n.MsgDockerGenerating: {
				ID:      i18n.MsgDockerGenerating,
				Message: "正在为 {{.Language}} 项目生成 Dockerfile",
			},
			i18n.MsgPluginLoading: {
				ID:      i18n.MsgPluginLoading,
				Message: "正在加载插件: {{.Name}}",
			},
		}

	default:
		// For other languages, use English as base
		bundle.Messages = map[string]i18n.MessageConfig{
			i18n.MsgWelcome: {
				ID:      i18n.MsgWelcome,
				Message: "Welcome to GZH Manager",
			},
			i18n.MsgError: {
				ID:      i18n.MsgError,
				Message: "Error",
			},
			i18n.MsgSuccess: {
				ID:      i18n.MsgSuccess,
				Message: "Success",
			},
		}
	}

	return bundle
}

// saveLocaleBundle saves a locale bundle to a JSON file
func saveLocaleBundle(filename string, bundle *i18n.LocalizationBundle) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bundle)
}

// createConfigFile creates an i18n configuration file
func createConfigFile(filename string, config *i18n.Config, force bool) error {
	if !force {
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("⚠️  Config file already exists: %s (use --force to overwrite)\n", filename)
			return nil
		}
	}

	// Create a simplified config for the JSON file
	configData := map[string]interface{}{
		"locales_dir":         config.LocalesDir,
		"default_language":    config.DefaultLanguage,
		"fallback_language":   config.FallbackLanguage,
		"supported_languages": config.SupportedLanguages,
		"version":             "1.0.0",
		"created_by":          "gz i18n init",
	}

	return saveJSONFile(filename, configData)
}

// saveJSONFile saves data to a JSON file
func saveJSONFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
