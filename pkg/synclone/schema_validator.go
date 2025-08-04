// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// LoadSchemaFromFile loads the JSON schema from the docs directory.
func LoadSchemaFromFile() (string, error) {
	return LoadSchemaFromFileWithEnv(env.NewOSEnvironment())
}

// LoadSchemaFromFileWithEnv loads the JSON schema using the provided environment.
func LoadSchemaFromFileWithEnv(environment env.Environment) (string, error) {
	// Try to find the schema file relative to the current working directory
	goPath := environment.Get("GOPATH")
	paths := []string{
		"docs/bulk-clone-schema.json",
		"../../docs/bulk-clone-schema.json",
	}

	if goPath != "" {
		paths = append(paths, filepath.Join(goPath, "src", "github.com", "gizzahub", "gzh-manager-go", "docs", "bulk-clone-schema.json"))
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			return string(data), nil
		}
	}

	// If not found, return the embedded schema
	return bulkCloneSchemaJSON, nil
}

// bulkCloneSchemaJSON contains the embedded JSON schema for validating bulk-clone configuration files.
var bulkCloneSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Bulk Clone Configuration Schema",
  "description": "Schema for gzh-manager bulk-clone configuration files",
  "type": "object",
  "required": ["version"],
  "properties": {
    "version": {
      "type": "string",
      "description": "Configuration schema version",
      "enum": ["1.0"]
    },
    "default": {
      "type": "object",
      "description": "Default settings for all providers",
      "properties": {
        "protocol": {
          "type": "string",
          "description": "Default protocol for Git operations",
          "enum": ["http", "https", "ssh"]
        },
        "github": {
          "$ref": "#/definitions/githubDefault"
        },
        "gitlab": {
          "$ref": "#/definitions/gitlabDefault"
        }
      },
      "additionalProperties": false
    },
    "repo_roots": {
      "type": "array",
      "description": "Repository configurations for different organizations",
      "items": {
        "$ref": "#/definitions/githubRepoRoot"
      }
    },
    "ignore_names": {
      "type": "array",
      "description": "Repository name patterns to ignore (regex)",
      "items": {
        "type": "string"
      }
    }
  },
  "additionalProperties": false,
  "definitions": {
    "githubDefault": {
      "type": "object",
      "description": "Default settings for GitHub",
      "properties": {
        "root_path": {
          "type": "string",
          "description": "Base directory for GitHub repositories"
        },
        "provider": {
          "type": "string",
          "description": "Provider identifier"
        },
        "protocol": {
          "type": "string",
          "description": "Protocol override for GitHub",
          "enum": ["", "http", "https", "ssh"]
        },
        "org_name": {
          "type": "string",
          "description": "Default organization name"
        }
      },
      "additionalProperties": false
    },
    "gitlabDefault": {
      "type": "object",
      "description": "Default settings for GitLab",
      "properties": {
        "root_path": {
          "type": "string",
          "description": "Base directory for GitLab repositories"
        },
        "provider": {
          "type": "string",
          "description": "Provider identifier"
        },
        "url": {
          "type": "string",
          "description": "GitLab instance URL"
        },
        "recursive": {
          "type": "boolean",
          "description": "Whether to clone subgroups recursively"
        },
        "protocol": {
          "type": "string",
          "description": "Protocol override for GitLab",
          "enum": ["", "http", "https", "ssh"]
        },
        "group_name": {
          "type": "string",
          "description": "Default group name"
        }
      },
      "additionalProperties": false
    },
    "githubRepoRoot": {
      "type": "object",
      "description": "GitHub organization configuration",
      "required": ["root_path", "provider", "protocol", "org_name"],
      "properties": {
        "root_path": {
          "type": "string",
          "description": "Directory where repositories will be cloned"
        },
        "provider": {
          "type": "string",
          "description": "Git provider",
          "const": "github"
        },
        "protocol": {
          "type": "string",
          "description": "Protocol for Git operations",
          "enum": ["http", "https", "ssh"]
        },
        "org_name": {
          "type": "string",
          "description": "GitHub organization name",
          "minLength": 1
        }
      },
      "additionalProperties": false
    }
  }
}`

// ValidateConfigWithSchema validates a configuration against the JSON schema.
func ValidateConfigWithSchema(configPath string) error {
	// Load the schema
	schemaLoader := gojsonschema.NewStringLoader(bulkCloneSchemaJSON)

	// Read and convert YAML config to JSON for validation
	cfg := &bulkCloneConfig{}
	if err := cfg.ReadConfig(configPath); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Convert to JSON
	jsonData, err := configToJSON(cfg)
	if err != nil {
		return fmt.Errorf("failed to convert config to JSON: %w", err)
	}

	// Load the document
	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, err := range result.Errors() {
			errors = append(errors, fmt.Sprintf("- %s", err))
		}

		return fmt.Errorf("config validation failed:\n%s", joinStrings(errors, "\n"))
	}

	return nil
}

// configToJSON converts the config struct to JSON bytes.
func configToJSON(cfg *bulkCloneConfig) ([]byte, error) {
	// First convert to a generic map to handle the YAML tags
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	var genericData map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &genericData); err != nil {
		return nil, err
	}

	// Then convert to JSON
	return json.Marshal(genericData)
}

// joinStrings joins strings with a separator (simple helper).
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}

	return result
}
