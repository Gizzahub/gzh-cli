# Development Container Guide

This document provides comprehensive guidance for using the development container setup for the GZH Manager Go project.

## Overview

The development container provides a consistent, reproducible development environment that includes:

- **Go 1.24.0** with all development tools
- **Node.js 20** for React dashboard
- **Python 3.12** for scripting
- **Docker-in-Docker** for container development
- **Comprehensive tooling** for linting, testing, and debugging
- **VS Code integration** with optimized settings and extensions

## Quick Start

### Prerequisites

1. **Docker Desktop** or **Docker Engine** installed and running
2. **Visual Studio Code** with the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

### Opening the Project

1. Clone and open the repository:

   ```bash
   git clone https://github.com/gizzahub/gzh-cli.git
   cd gzh-cli
   code .
   ```

2. When VS Code opens, you'll see a notification to "Reopen in Container":
   - Click **"Reopen in Container"**
   - Or use Command Palette (`Ctrl+Shift+P`): **"Remote-Containers: Reopen in Container"**

3. Wait for the container to build and configure (5-10 minutes on first run)

4. Once ready, run the initial setup:
   ```bash
   make bootstrap
   make build
   ./gz --help
   ```

## Container Architecture

### Base Image

The container is built on `mcr.microsoft.com/devcontainers/base:ubuntu-22.04` with comprehensive multi-language support.

### Installed Tools

#### Go Development

- **Go 1.24.0** - Primary language
- **golangci-lint 1.63.4** - Comprehensive linting
- **gosec** - Security analysis
- **gofumpt** - Enhanced formatting
- **gci** - Import organization
- **mockgen** - Mock generation
- **goreleaser** - Release automation
- **staticcheck** - Static analysis

#### Node.js/JavaScript Development

- **Node.js 20** - JavaScript runtime
- **npm/yarn/pnpm** - Package managers
- **TypeScript** - Type-safe JavaScript
- **ESLint** - JavaScript linting
- **Prettier** - Code formatting
- **node-gyp** - Native addon builds

#### Python Development

- **Python 3.12** - Latest Python
- **pip** - Package installer
- **black** - Code formatting
- **isort** - Import sorting
- **pylint** - Code analysis
- **mypy** - Type checking
- **pytest** - Testing framework

#### Development Tools

- **Docker CLI** - Container management
- **Docker Compose** - Multi-container apps
- **GitHub CLI** - GitHub integration
- **pre-commit** - Git hook framework
- **make/cmake** - Build systems
- **git/git-lfs** - Version control

#### System Tools

- **zsh + Oh My Zsh** - Enhanced shell
- **vim/nano** - Text editors
- **htop** - Process monitor
- **jq/yq** - JSON/YAML processing
- **curl/wget** - HTTP clients

### Directory Structure

```
/workspace/                 # Project root (mapped from host)
├── .devcontainer/         # Container configuration
│   ├── devcontainer.json  # Main configuration
│   ├── Dockerfile         # Container definition
│   ├── scripts/           # Setup scripts
│   └── logs/             # Container logs
├── .vscode/              # VS Code settings
│   └── launch.json       # Debug configurations
└── [project files]       # Your source code

/home/vscode/             # Container user home
├── .zshrc                # Shell configuration
├── .gitconfig            # Git configuration (mounted)
├── .ssh/                 # SSH keys (mounted)
└── go/                   # Go workspace
```

## Development Workflows

### Go Development

#### Building and Testing

```bash
# Full build process
make bootstrap    # Install dependencies
make build        # Build gz binary
make test         # Run tests
make lint         # Run linting
make fmt          # Format code

# Quick development cycle
make build && ./gz version
go test ./cmd/bulk-clone -v
golangci-lint run ./cmd/...
```

#### Debugging

1. Open VS Code Debug panel (`Ctrl+Shift+D`)
2. Select a debug configuration:
   - **Debug GZH CLI** - Debug main application
   - **Debug GZH Bulk Clone** - Debug bulk clone command
   - **Debug Go Test** - Debug specific tests
3. Set breakpoints and press `F5`

#### Package Development

```bash
# Test specific packages
go test ./pkg/github -v
go test ./cmd/bulk-clone -v
go test -tags integration ./...

# Generate mocks
make generate-mocks

# Security analysis
make security
gosec ./...
```

### React Dashboard Development

#### Setup and Development

```bash
# Navigate to web directory
cd web

# Install dependencies (if not done by container)
npm ci

# Start development server
npm start
# Opens http://localhost:3000

# Build for production
npm run build
```

#### Testing

```bash
# Run React tests
cd web
npm test

# Run linting
npm run lint
npm run format
```

### Docker Development

#### Building Images

```bash
# Build project Docker image
docker build -t gzh-manager .

# Test the image
docker run -it gzh-manager --help

# Run with volume mount
docker run -v $(pwd):/workspace gzh-manager bulk-clone --help
```

#### Integration Testing

```bash
# Run Docker-based integration tests
make test-docker

# Run specific service tests
make test-gitlab
make test-gitea
make test-redis
```

## Port Forwarding

The container automatically forwards these ports:

| Port | Service      | Auto-Open | Description              |
| ---- | ------------ | --------- | ------------------------ |
| 8080 | GZH API      | Notify    | Main API server          |
| 3000 | React Dev    | Preview   | React development server |
| 9090 | Prometheus   | No        | Metrics collection       |
| 9093 | Alertmanager | No        | Alert management         |
| 6060 | Go pprof     | No        | Performance profiling    |

### Accessing Services

```bash
# Start API server
./gz serve --port 8080
# Access at http://localhost:8080

# Start React development
cd web && npm start
# Access at http://localhost:3000

# Enable Go profiling
import _ "net/http/pprof"
# Access at http://localhost:6060/debug/pprof/
```

## Environment Variables

### Container Environment

Set automatically in the container:

```bash
GZH_DEV_MODE=true          # Enable development features
GO111MODULE=on             # Go modules support
GOPROXY=https://proxy.golang.org,direct
GOSUMDB=sum.golang.org
CGO_ENABLED=1              # CGO support for native dependencies
DOCKER_BUILDKIT=1          # Enhanced Docker builds
COMPOSE_DOCKER_CLI_BUILD=1 # Docker Compose v2
```

### Project-Specific Variables

Set these for development:

```bash
# GitHub integration
export GITHUB_TOKEN="ghp_..."

# GitLab integration
export GITLAB_TOKEN="glpat-..."

# Debug mode
export GZH_LOG_LEVEL="debug"
export GZH_TRACE_ENABLED="true"
```

### Setting Variables

1. **In VS Code Terminal**:

   ```bash
   export GITHUB_TOKEN="your-token"
   ```

2. **In devcontainer.json**:

   ```json
   {
     "containerEnv": {
       "GITHUB_TOKEN": "${localEnv:GITHUB_TOKEN}"
     }
   }
   ```

3. **In .env file** (add to .gitignore):
   ```bash
   GITHUB_TOKEN=your-token
   GITLAB_TOKEN=your-token
   ```

## VS Code Integration

### Installed Extensions

The container includes these extensions:

#### Go Development

- `golang.go` - Official Go extension
- `golang.go-nightly` - Latest Go features

#### JavaScript/TypeScript

- `ms-vscode.vscode-typescript-next` - TypeScript support
- `esbenp.prettier-vscode` - Code formatting
- `dbaeumer.vscode-eslint` - Linting

#### Python

- `ms-python.python` - Python support
- `ms-python.black-formatter` - Code formatting
- `ms-python.pylint` - Linting
- `ms-python.isort` - Import sorting

#### Development Tools

- `ms-azuretools.vscode-docker` - Docker support
- `eamodio.gitlens` - Enhanced Git features
- `github.vscode-pull-request-github` - GitHub integration
- `redhat.vscode-yaml` - YAML support
- `ms-vscode.makefile-tools` - Makefile support

### Optimized Settings

Key VS Code settings configured:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "gofumpt",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit",
    "source.fixAll.eslint": "explicit"
  },
  "yaml.schemas": {
    "./docs/bulk-clone-schema.json": ["examples/bulk-clone*.yaml"]
  }
}
```

### Debug Configurations

Pre-configured debug scenarios:

1. **Debug GZH CLI** - Debug main application with `--help`
2. **Debug GZH Bulk Clone** - Debug bulk clone with sample config
3. **Debug GZH Config Validate** - Debug configuration validation
4. **Debug GZH Web Dashboard** - Debug web server
5. **Debug Current Go File** - Debug the currently open file
6. **Debug Go Test** - Debug tests in current package
7. **Attach to Process** - Attach to running process

## File Mounting

### Host Mounts

These directories are mounted from your host:

```bash
# Source code (read-write)
${workspaceFolder} → /workspace

# Git configuration (read-only)
~/.gitconfig → /home/vscode/.gitconfig

# SSH keys (read-only)
~/.ssh → /home/vscode/.ssh

# Docker socket (for Docker-in-Docker)
/var/run/docker.sock → /var/run/docker.sock
```

### Volume Persistence

These are managed by Docker:

- **Go module cache** - Persisted between container rebuilds
- **Node.js cache** - npm/yarn cache persistence
- **VS Code extensions** - Extension cache and settings

## Customization

### Adding Extensions

Edit `.devcontainer/devcontainer.json`:

```json
{
  "customizations": {
    "vscode": {
      "extensions": ["existing.extension", "your.new.extension"]
    }
  }
}
```

### Installing Additional Tools

Edit `.devcontainer/Dockerfile`:

```dockerfile
# Add after existing RUN commands
RUN apt-get update && apt-get install -y \
    your-package \
    another-tool \
    && apt-get clean

# Or install with specific package managers
RUN go install github.com/your/tool@latest
RUN npm install -g your-npm-tool
RUN pip3 install your-python-tool
```

### Custom Setup Scripts

Add scripts to `.devcontainer/scripts/` and reference them:

```json
{
  "postCreateCommand": ".devcontainer/scripts/my-setup.sh",
  "postStartCommand": ".devcontainer/scripts/my-startup.sh"
}
```

### Environment Customization

Add to `devcontainer.json`:

```json
{
  "containerEnv": {
    "MY_CUSTOM_VAR": "value",
    "PATH": "${containerEnv:PATH}:/my/custom/path"
  }
}
```

## Troubleshooting

### Container Issues

#### Container Won't Start

```bash
# Check Docker is running
docker info

# Check Docker Desktop status
# Restart Docker Desktop if needed

# Rebuild container
# VS Code Command Palette: "Remote-Containers: Rebuild Container"
```

#### Slow Container Performance

```bash
# Check Docker resource allocation
# Docker Desktop → Settings → Resources
# Increase CPU and memory if needed

# Optimize file syncing (macOS/Windows)
# Add to .dockerignore:
node_modules
dist
build
*.log
```

#### Permission Issues

```bash
# Fix ownership inside container
sudo chown -R vscode:vscode /workspace

# Fix script permissions
chmod +x .devcontainer/scripts/*.sh

# Check mount permissions
ls -la ~/.ssh
ls -la ~/.gitconfig
```

### Development Issues

#### Go Module Issues

```bash
# Clean module cache
go clean -modcache
go mod download
go mod tidy

# Verify Go environment
go env
echo $GOPATH
echo $GOROOT
```

#### Git Configuration Issues

```bash
# Check Git configuration
git config --list

# Reconfigure if needed
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Trust workspace directory
git config --global --add safe.directory /workspace
```

### VS Code Issues

#### Extensions Not Working

```bash
# Reload VS Code window
# Command Palette: "Developer: Reload Window"

# Reinstall extensions
# Command Palette: "Remote-Containers: Rebuild Container"

# Check extension logs
# View → Output → Select extension
```

#### Debug Not Working

```bash
# Check Go extension status
# Command Palette: "Go: Show Current GOPATH"

# Verify debug configuration
# Check .vscode/launch.json

# Rebuild binary
make clean
make build
```

#### IntelliSense Issues

```bash
# Restart Go language server
# Command Palette: "Go: Restart Language Server"

# Verify module cache
go mod download
go mod tidy

# Check workspace settings
# File → Preferences → Settings (Workspace)
```

## Performance Optimization

### Docker Performance

1. **Allocate sufficient resources**:
   - CPU: 4+ cores
   - Memory: 8GB+
   - Disk: SSD recommended

2. **Optimize file syncing**:

   ```dockerfile
   # Add to .dockerignore
   node_modules
   dist
   build
   .git
   *.log
   coverage
   ```

3. **Use multi-stage builds**:
   ```dockerfile
   # Cache dependencies separately
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN go build
   ```

### Build Performance

1. **Parallel builds**:

   ```bash
   # Use make parallel jobs
   make -j$(nproc) build

   # Go parallel compilation
   export GOMAXPROCS=$(nproc)
   ```

2. **Incremental builds**:

   ```bash
   # Only rebuild changed packages
   go install ./cmd/...

   # Use go build cache
   export GOCACHE=/workspace/.gocache
   ```

3. **Dependency caching**:

   ```dockerfile
   # Cache Go modules
   COPY go.mod go.sum ./
   RUN go mod download

   # Cache npm dependencies
   COPY package*.json ./
   RUN npm ci
   ```

## Best Practices

### Development Workflow

1. **Regular container updates**:

   ```bash
   # Weekly container rebuild
   # Command Palette: "Remote-Containers: Rebuild Container"
   ```

2. **Code quality checks**:

   ```bash
   # Before committing
   make fmt
   make lint
   make test
   pre-commit run --all-files
   ```

3. **Dependency management**:

   ```bash
   # Weekly dependency updates
   go get -u ./...
   go mod tidy

   cd web && npm update
   ```

### Security

1. **Secret management**:

   ```bash
   # Never commit secrets
   echo "*.env" >> .gitignore
   echo "secrets/" >> .gitignore

   # Use environment variables
   export GITHUB_TOKEN="$(cat ~/.github_token)"
   ```

2. **Container security**:

   ```bash
   # Run as non-root user
   USER vscode

   # Minimal attack surface
   RUN apt-get autoremove -y
   RUN rm -rf /var/lib/apt/lists/*
   ```

3. **Network security**:
   ```bash
   # Use HTTPS for all external calls
   # Validate certificates
   # Use secure protocols
   ```

### Resource Management

1. **Memory management**:

   ```bash
   # Monitor memory usage
   htop
   docker stats

   # Tune Go GC
   export GOGC=100
   export GOMEMLIMIT=6GiB
   ```

2. **Disk management**:

   ```bash
   # Clean build artifacts
   make clean
   docker system prune

   # Monitor disk usage
   du -sh /workspace/*
   df -h
   ```

## Additional Resources

- [VS Code Dev Containers Documentation](https://code.visualstudio.com/docs/remote/containers)
- [Docker Best Practices](https://docs.docker.com/develop/best-practices/)
- [Go Development Guide](https://golang.org/doc/code.html)
- [Node.js Best Practices](https://nodejs.org/en/docs/guides/)
- [Python Development Guide](https://docs.python.org/3/tutorial/)
- [GZH Manager Documentation](../README.md)

---

**Note**: This development container configuration is specifically optimized for the GZH Manager project structure and workflows. Adjustments may be needed for different project requirements.
