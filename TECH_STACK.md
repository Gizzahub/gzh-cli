<!-- 🚫 AI_MODIFY_PROHIBITED -->
<!-- This file should not be modified by AI agents -->

# Tech Stack

## Core Technologies

- **Language**: Go 1.21+
- **Framework**: Cobra CLI framework
- **Database**: File-based configuration (YAML/JSON)
- **Cloud Platform**: Multi-cloud support (AWS, GCP, Azure)

## Development Tools

- **Build System**: Make + Go modules
- **Testing**: testify framework with gomock
- **Linting**: golangci-lint v2 with comprehensive rules
- **Package Manager**: Go modules

## External Dependencies

### Core Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `gopkg.in/yaml.v3` - YAML processing
- `github.com/go-git/go-git/v5` - Git operations
- `github.com/google/go-github/v45` - GitHub API
- `github.com/xanzy/go-gitlab` - GitLab API
- `github.com/fatih/color` - Terminal colors
- `github.com/schollz/progressbar/v3` - Progress bars

### Development Dependencies

- `github.com/stretchr/testify` - Testing framework
- `github.com/golang/mock/gomock` - Mock generation
- `github.com/golangci/golangci-lint` - Code linting
- `golang.org/x/tools/go/packages` - Code analysis
- `github.com/xeipuuv/gojsonschema` - JSON schema validation

## Architecture Overview

gzh-manager-go follows a modular CLI architecture with the following key components:

### Command Structure

```
cmd/
├── root.go              # Main CLI entry point
├── bulk-clone/          # Repository cloning commands
├── always-latest/       # Package manager updates
├── dev-env/            # Development environment management
├── net-env/            # Network environment transitions
├── ide/                # IDE settings management
└── webhook/            # Webhook management
```

### Package Architecture

```
pkg/
├── bulk-clone/         # Configuration loading and validation
├── github/            # GitHub API integration
├── gitlab/            # GitLab API integration
├── gitea/             # Gitea API integration
├── gogs/              # Gogs API integration (planned)
└── cloud/             # Cloud provider abstractions
```

### Design Patterns

1. **Service-specific implementations**: Each Git platform has dedicated packages following common interfaces
2. **Configuration-driven design**: Extensive YAML configuration with schema validation
3. **Cross-platform support**: Native OS detection and platform-specific implementations
4. **Atomic operations**: Commands designed for safe execution with backup and rollback
5. **Comprehensive testing**: Mock services and environment-specific tests

## Deployment

### Binary Distribution

- **Binary Name**: `gz`
- **Platforms**: Linux, macOS, Windows
- **Installation**:
  - Go install: `go install github.com/gizzahub/gzh-manager-go@latest`
  - Manual build: `make build && make install`

### Configuration

- **Config Hierarchy**:
  1. Environment variable: `GZH_CONFIG_PATH`
  2. Current directory: `./bulk-clone.yaml`
  3. User config: `~/.config/gzh-manager/bulk-clone.yaml`
  4. System config: `/etc/gzh-manager/bulk-clone.yaml`

### Authentication

- Token-based authentication for private repositories
- Environment variable support (GITHUB_TOKEN, GITLAB_TOKEN, etc.)
- SSH key management and configuration

### Quality Assurance

- Pre-commit hooks with golangci-lint
- Comprehensive test suite with coverage reporting
- JSON Schema validation for configuration files
- Cross-platform testing
