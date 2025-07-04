package bulk_clone

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// >>>>>>>>>> default >>>>>>>>>>
type BulkCloneDefault struct {
	Protocol string                 `yaml:"protocol" validate:"required,oneof=http https ssh"`
	Github   BulkCloneDefaultGithub `yaml:"github"`
	Gitlab   BulkCloneDefaultGitlab `yaml:"gitlab"`
}

type BulkCloneDefaultGithub struct {
	RootPath string `yaml:"root_path"`
	Provider string `yaml:"provider"`
	Protocol string `yaml:"protocol"`
	OrgName  string `yaml:"org_name"`
}

type BulkCloneDefaultGitlab struct {
	RootPath  string `yaml:"root_path"`
	Provider  string `yaml:"provider"`
	Url       string `yaml:"url"`
	Recursive bool   `yaml:"recursive"`
	Protocol  string `yaml:"protocol"`
	GroupName string `yaml:"group_name"`
}

// <<<<<<<<<< default <<<<<<<<<<

type BulkCloneGithub struct {
	RootPath string `yaml:"root_path" validate:"required"`
	Provider string `yaml:"provider" validate:"required"`
	Protocol string `yaml:"protocol" validate:"required,oneof=http https ssh"`
	OrgName  string `yaml:"org_name" validate:"required"`
}

type BulkCloneGitlab struct {
	RootPath  string `yaml:"root_path" validate:"required"`
	Provider  string `yaml:"provider" validate:"required"`
	Url       string `yaml:"url"`
	Recursive bool   `yaml:"recursive"`
	Protocol  string `yaml:"protocol" validate:"required,oneof=http https ssh"`
	GroupName string `yaml:"group_name" binding:"required"`
}

type bulkCloneConfig struct {
	Version           string            `yaml:"version"`
	Default           BulkCloneDefault  `yaml:"default"`
	IgnoreNameRegexes []string          `yaml:"ignore_names"`
	RepoRoots         []BulkCloneGithub `yaml:"repo_roots"`
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func (cfg *bulkCloneConfig) ConfigExists(targetPath string) bool {
	return fileExists(path.Join(targetPath, "bulk-clone.yaml"))
}

func (cfg *bulkCloneConfig) ReadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	err = cfg.validateConfig()
	if err != nil {
		printValidationErrors(err)
		return fmt.Errorf("failed to validate config file: %v", err)
	}

	return nil
}

// ReadConfigWithoutValidation reads config file without validation (used for overlays)
func (cfg *bulkCloneConfig) ReadConfigWithoutValidation(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	return nil
}

// ReadConfigFromDir reads config from a directory (legacy support)
func (cfg *bulkCloneConfig) ReadConfigFromDir(targetPath string) {
	configPath := path.Join(targetPath, "bulk-clone.yaml")
	if err := cfg.ReadConfig(configPath); err != nil {
		log.Fatal(err)
	}
}

// 커스텀 오류 메시지 맵
var errorMessages = map[string]string{
	"required": "필수 필드입니다.",
	"url":      "유효한 URL을 입력해야 합니다.",
	"oneof":    "허용되는 값이 아닙니다 (http, https, ssh).",
}

// 유효성 검사 오류 상세 메시지 출력 함수
func printValidationErrors(err error) {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		for _, e := range errs {
			// 기본 메시지
			msg, exists := errorMessages[e.Tag()]
			if !exists {
				msg = fmt.Sprintf("필드 '%s'은(는) '%s' 규칙을 만족해야 합니다.", e.Field(), e.Tag())
			}

			// 추가 정보가 필요한 경우 (예: oneof)
			if e.Tag() == "oneof" {
				msg = fmt.Sprintf("필드 '%s'은(는) 허용되는 값 중 하나여야 합니다: %s.", e.Field(), e.Param())
			}

			fmt.Printf("오류: %s\n", msg)
		}
	}
}

// 유효성 검사 함수
func (cfg *bulkCloneConfig) validateConfig() error {
	// validate := validator.New(validator.WithRequiredStructEnabled())
	validate := validator.New()
	return validate.Struct(cfg)
}
