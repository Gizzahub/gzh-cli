// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package bulkclone provides configuration structures and validation
// for bulk repository cloning operations across multiple Git platforms
// including GitHub, GitLab, Gitea, and Gogs.
package bulkclone

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// bulkCloneDefault defines default configuration settings for bulk clone operations.
type bulkCloneDefault struct {
	Protocol string                 `yaml:"protocol" validate:"required,oneof=http https ssh"`
	Github   bulkCloneDefaultGithub `yaml:"github"`
	Gitlab   bulkCloneDefaultGitlab `yaml:"gitlab"`
}

// bulkCloneDefaultGithub defines default GitHub-specific configuration.
//
// 키 표기는 snake_case다. 이 패키지가 함께 들고 있는 JSON 스키마
// (schema_validator.go), examples/synclone/*.yaml, 마이그레이션 문서가 모두
// snake_case를 쓴다. 예전에는 여기만 camelCase라서 설정 파일의 값이 통째로
// 버려졌다 -- yaml.Unmarshal은 모르는 키를 오류로 보지 않고 그냥 지나치므로
// 아무 데도 흔적이 남지 않았다. tagliatelle은 yaml 태그를 camel로 요구하지만
// 그건 린터 설정이 실제 파일 형식과 어긋난 것이다(tasks/issue/33).
type bulkCloneDefaultGithub struct { // 설정 파일 형식이 snake_case다
	RootPath string `yaml:"root_path"`
	Provider string `yaml:"provider"`
	Protocol string `yaml:"protocol"`
	OrgName  string `yaml:"org_name"`
}

// bulkCloneDefaultGitlab defines default GitLab-specific configuration.
type bulkCloneDefaultGitlab struct { // 설정 파일 형식이 snake_case다
	RootPath  string `yaml:"root_path"`
	Provider  string `yaml:"provider"`
	URL       string `yaml:"url"`
	Recursive bool   `yaml:"recursive"`
	Protocol  string `yaml:"protocol"`
	GroupName string `yaml:"group_name"`
}

// BulkCloneGithub represents GitHub bulk clone configuration.
//
//nolint:revive // Type name maintained for clarity; 설정 파일 형식이 snake_case다
type BulkCloneGithub struct {
	RootPath string `yaml:"root_path" validate:"required"`
	Provider string `yaml:"provider" validate:"required"`
	Protocol string `yaml:"protocol" validate:"required,oneof=http https ssh"`
	OrgName  string `yaml:"org_name" validate:"required"`
}

// BulkCloneGitlab represents GitLab bulk clone configuration.
//
//nolint:revive // Type name maintained for clarity; 설정 파일 형식이 snake_case다
type BulkCloneGitlab struct {
	RootPath  string `yaml:"root_path" validate:"required"`
	Provider  string `yaml:"provider" validate:"required"`
	URL       string `yaml:"url"`
	Recursive bool   `yaml:"recursive"`
	Protocol  string `yaml:"protocol" validate:"required,oneof=http https ssh"`
	GroupName string `yaml:"group_name" validate:"required"`
}

// 설정 파일 형식이 snake_case다.
type bulkCloneConfig struct {
	Version           string           `yaml:"version"`
	Default           bulkCloneDefault `yaml:"default"`
	IgnoreNameRegexes []string         `yaml:"ignore_names"`
	// dive가 있어야 원소까지 들어간다.
	//
	// validator는 구조체 필드(Default 같은)는 알아서 따라 들어가지만
	// 조각(slice)은 dive를 붙이지 않으면 원소를 건드리지 않는다. 그래서
	// BulkCloneGithub의 required와 oneof 네 개가 여기서는 전부 죽어
	// 있었다. root_path도 org_name도 없는 repo_roots 항목이 그대로
	// 통과했고, protocol에 "carrier-pigeon"을 적어도 아무 말이 없었다.
	// 설정이 틀린 채로 통과하니 오류는 한참 뒤 clone 단계에서 엉뚱한
	// 모습으로 나타난다.
	RepoRoots []BulkCloneGithub `yaml:"repo_roots" validate:"dive"`
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
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	err = cfg.validateConfig()
	if err != nil {
		printValidationErrors(err)
		return fmt.Errorf("failed to validate config file: %w", err)
	}

	return nil
}

// ReadConfigWithoutValidation reads config file without validation (used for overlays).
func (cfg *bulkCloneConfig) ReadConfigWithoutValidation(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}

// ReadConfigFromDir reads config from a directory (legacy support).
func (cfg *bulkCloneConfig) ReadConfigFromDir(targetPath string) {
	configPath := path.Join(targetPath, "bulk-clone.yaml")
	if err := cfg.ReadConfig(configPath); err != nil {
		log.Fatal(err)
	}
}

// errorMessages contains custom error messages for validation.
var errorMessages = map[string]string{
	"required": "This field is required.",
	"url":      "Please enter a valid URL.",
	"oneof":    "Invalid value (allowed: http, https, ssh).",
}

// printValidationErrors prints detailed validation error messages.
func printValidationErrors(err error) {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		for _, e := range errs {
			// Default message
			msg, exists := errorMessages[e.Tag()]
			if !exists {
				msg = fmt.Sprintf("Field '%s' must satisfy '%s' rule.", e.Field(), e.Tag())
			}

			// Additional information for specific cases (e.g., oneof)
			if e.Tag() == "oneof" {
				msg = fmt.Sprintf("Field '%s' must be one of the allowed values: %s.", e.Field(), e.Param())
			}

			fmt.Printf("Error: %s\n", msg)
		}
	}
}

// validateConfig validates the configuration structure.
func (cfg *bulkCloneConfig) validateConfig() error {
	validate := validator.New()
	return validate.Struct(cfg)
}
