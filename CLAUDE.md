# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gzh-manager-go is a CLI tool for managing multiple Git repositories across GitHub, GitLab, Gitea, and Gogs. It specializes in bulk operations like cloning entire organizations and synchronizing repositories.

## Essential Commands

### Development Setup
```bash
make bootstrap  # Install all build dependencies - run this first
```

### Building and Running
```bash
make build      # Creates 'gz' executable
make install    # Installs to GOPATH/bin
make run        # Run with version tag
```

### Code Quality - ALWAYS RUN BEFORE COMMITTING
```bash
make fmt        # Format code with gofumpt and gci
make lint       # Run golangci-lint checks
make test       # Run all tests with coverage
```

### Testing
```bash
make test       # Run unit tests
make cover      # Show coverage with race detection
go test ./cmd/bulk-clone -v  # Run specific package tests
```

## Architecture

### Command Structure
- **cmd/** - CLI commands using cobra framework
  - `root.go` - Main entry point and command setup
  - `bulk-clone/` - Service-specific bulk clone implementations (GitHub, GitLab, Gitea, Gogs)
  - `gen-config/` - Configuration generation

### Core Packages
- **internal/** - Private packages
  - `config/` - YAML configuration parsing
  - `git/` - Git operations wrapper
  - `sync/` - Repository synchronization logic
  - `auth/` - Authentication handling
  - `logger/` - Structured logging

- **pkg/** - Public packages
  - Service-specific implementations (github/, gitlab/, gitea/, gogs/)
  - Configuration structures

### Key Patterns
1. Each Git service has its own implementation following a common interface
2. Configuration-driven design using YAML files (see samples/bulk-clone.yml)
3. Uses cobra for CLI, viper for config, testify for testing
4. Extensive error handling with wrapped errors

## Testing Guidelines
- Test files use `*_test.go` convention
- Uses testify for assertions
- Mock external services when possible
- Check for GITHUB_TOKEN in environment for integration tests

## Important Notes
- The binary is named 'gz' not 'gzh-manager-go'
- Always run `make fmt` before committing code
- Configuration files support both CLI flags and YAML config
- Supports multiple authentication methods per service