// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

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

type bulkCloneDefault struct {
	Protocol string                 `yaml:"protocol" validate:"required,oneof=http https ssh"`
	Github   bulkCloneDefaultGithub `yaml:"github"`
	Gitlab   bulkCloneDefaultGitlab `yaml:"gitlab"`
}

type bulkCloneDefaultGithub struct {
	RootPath string `yaml:"rootPath"`
	Provider string `yaml:"provider"`
	Protocol string `yaml:"protocol"`
	OrgName  string `yaml:"orgName"`
}

type bulkCloneDefaultGitlab struct {
	RootPath  string `yaml:"rootPath"`
	Provider  string `yaml:"provider"`
	URL       string `yaml:"url"`
	Recursive bool   `yaml:"recursive"`
	Protocol  string `yaml:"protocol"`
	GroupName string `yaml:"groupName"`
}

// BulkCloneGithub represents GitHub bulk clone configuration.
type BulkCloneGithub struct { //nolint:revive // Type name maintained for clarity in configuration structs
	RootPath string `yaml:"rootPath" validate:"required"`
	Provider string `yaml:"provider" validate:"required"`
	Protocol string `yaml:"protocol" validate:"required,oneof=http https ssh"`
	OrgName  string `yaml:"orgName" validate:"required"`
}

// BulkCloneGitlab represents GitLab bulk clone configuration.
type BulkCloneGitlab struct { //nolint:revive // Type name maintained for clarity in configuration structs
	RootPath  string `yaml:"rootPath" validate:"required"`
	Provider  string `yaml:"provider" validate:"required"`
	URL       string `yaml:"url"`
	Recursive bool   `yaml:"recursive"`
	Protocol  string `yaml:"protocol" validate:"required,oneof=http https ssh"`
	GroupName string `yaml:"groupName" validate:"required"`
}

type bulkCloneConfig struct {
	Version           string            `yaml:"version"`
	Default           bulkCloneDefault  `yaml:"default"`
	IgnoreNameRegexes []string          `yaml:"ignoreNames"`
	RepoRoots         []BulkCloneGithub `yaml:"repoRoots"`
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
