# Development Container Configuration

This directory contains the development container configuration for the GZH Manager Go project. The development container provides a consistent, isolated development environment with all necessary tools and dependencies pre-installed.

## Features

### 🛠️ Pre-installed Tools

- **Go 1.24.0** - Primary development language
- **Node.js 20** - For React dashboard and Node.js bindings
- **Python 3.12** - For Python bindings
- **Docker** - Container support and integration testing
- **Git & GitHub CLI** - Version control and GitHub integration
- **Development Tools**:
  - golangci-lint, gosec, gofumpt, gci
  - ESLint, Prettier, TypeScript
  - Black, isort, pylint, mypy
  - pre-commit, make, cmake

### 🎨 VS Code Integration

- **Extensions**: Comprehensive language support for Go, TypeScript, Python, Docker, YAML
- **Settings**: Optimized for the project structure with proper linting, formatting, and testing
- **Debugging**: Ready-to-use debugging configurations
- **IntelliSense**: Full language server support for all technologies

### 🔧 Development Environment

- **Multi-language support**: Go, Node.js, Python, React
- **Hot reload**: File watching and automatic rebuilds
- **Port forwarding**: Automatic forwarding of development ports
- **Git integration**: SSH key mounting and configuration
- **Shell enhancements**: Zsh with Oh My Zsh and helpful aliases

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running
- [Visual Studio Code](https://code.visualstudio.com/) with [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

### Opening the Project

1. **Clone the repository**:

   ```bash
   git clone https://github.com/gizzahub/gzh-manager-go.git
   cd gzh-manager-go
   ```

2. **Open in VS Code**:

   ```bash
   code .
   ```

3. **Open in Container**:
   - When prompted, click "Reopen in Container"
   - Or use Command Palette: `Remote-Containers: Reopen in Container`
   - Or use Command Palette: `Remote-Containers: Open Folder in Container`

4. **Wait for setup**: The container will build and configure automatically (5-10 minutes on first run)

### First Commands

Once the container is ready:

```bash
# Install dependencies
make bootstrap

# Build the project
make build

# Run tests
make test

# Start development
./gz --help
```

## Container Structure

### File Organization

```
.devcontainer/
├── devcontainer.json     # Main configuration
├── Dockerfile           # Container image definition
├── README.md           # This file
└── scripts/
    ├── initialize.sh   # Pre-container setup
    ├── post-create.sh  # Initial container setup
    ├── post-start.sh   # Container startup tasks
    └── post-attach.sh  # VS Code attachment tasks
```

### Port Forwarding

| Port | Service                  | Auto-forward        |
| ---- | ------------------------ | ------------------- |
| 8080 | GZH API Server           | Yes                 |
| 3000 | React Development Server | Yes (opens preview) |
| 9090 | Prometheus               | No                  |
| 9093 | Alertmanager             | No                  |
| 6060 | Go pprof                 | No                  |

### Volume Mounts

- **Source code**: `/workspace` (full project)
- **Git config**: `~/.gitconfig` (bind mount from host)
- **SSH keys**: `~/.ssh` (bind mount from host)
- **Docker socket**: `/var/run/docker.sock` (Docker-in-Docker)

## Development Workflows

### Go Development

```bash
# Build and test
make build
make test
make lint

# Run specific tests
go test ./cmd/bulk-clone -v
go test -tags integration ./...

# Code quality
golangci-lint run
gosec ./...
```

### React Dashboard

```bash
# Start development server
cd web
npm start
# Open http://localhost:3000

# Build for production
npm run build
```

### Node.js Bindings

```bash
# Development
cd bindings/nodejs
npm run build
npm test

# Native compilation
npm run build:native
```

### Python Bindings

```bash
# Activate environment
cd bindings/python
source venv/bin/activate

# Development
pip install -e .
python -m pytest
```

### Docker Development

```bash
# Build image
docker build -t gzh-manager .

# Run integration tests
make test-docker

# Test container
docker run -it gzh-manager --help
```

## Environment Variables

### Required for Development

| Variable       | Purpose                  | Default |
| -------------- | ------------------------ | ------- |
| `GZH_DEV_MODE` | Enable development mode  | `true`  |
| `GO111MODULE`  | Go modules support       | `on`    |
| `CGO_ENABLED`  | CGO support for bindings | `1`     |

### Optional for Features

| Variable          | Purpose                | Default |
| ----------------- | ---------------------- | ------- |
| `GITHUB_TOKEN`    | GitHub API access      | (none)  |
| `GITLAB_TOKEN`    | GitLab API access      | (none)  |
| `DOCKER_BUILDKIT` | Enhanced Docker builds | `1`     |

## Customization

### Adding Extensions

Edit `.devcontainer/devcontainer.json`:

```json
{
  "customizations": {
    "vscode": {
      "extensions": ["your.extension.id"]
    }
  }
}
```

### Modifying Environment

Edit `.devcontainer/Dockerfile` to add tools:

```dockerfile
# Install additional tools
RUN apt-get update && apt-get install -y \
    your-package \
    && apt-get clean
```

### Custom Scripts

Add scripts to `.devcontainer/scripts/` and reference them in `devcontainer.json`:

```json
{
  "postCreateCommand": ".devcontainer/scripts/my-setup.sh"
}
```

## Troubleshooting

### Common Issues

1. **Container won't start**:

   ```bash
   # Check Docker is running
   docker info

   # Rebuild container
   # Command Palette: "Remote-Containers: Rebuild Container"
   ```

2. **Go modules issues**:

   ```bash
   # Inside container
   go clean -modcache
   go mod download
   ```

3. **Node.js build failures**:

   ```bash
   # Clear npm cache
   npm cache clean --force

   # Rebuild node-gyp
   cd bindings/nodejs
   npm run clean
   npm install
   ```

4. **Python environment issues**:

   ```bash
   # Recreate virtual environment
   cd bindings/python
   rm -rf venv
   python3 -m venv venv
   source venv/bin/activate
   pip install -e .
   ```

5. **Permission issues**:

   ```bash
   # Fix ownership
   sudo chown -R vscode:vscode /workspace

   # Fix permissions
   chmod +x .devcontainer/scripts/*.sh
   ```

### Performance Tips

1. **Use Docker volume for node_modules**:

   ```bash
   # Avoid mounting node_modules from host
   echo "node_modules" >> .dockerignore
   ```

2. **Exclude build artifacts**:

   ```bash
   # Add to .dockerignore
   dist/
   build/
   *.log
   ```

3. **Optimize container rebuilds**:
   ```bash
   # Use multi-stage builds
   # Copy dependency files first
   # Install dependencies
   # Copy source code last
   ```

## Logs and Debugging

### Container Logs

```bash
# View setup logs
tail -f /workspace/.devcontainer/logs/setup.log

# View startup logs
tail -f /workspace/.devcontainer/logs/startup.log

# View VS Code logs
tail -f /workspace/.devcontainer/logs/vscode.log
```

### VS Code Debugging

1. **Remote-Containers logs**:
   - Command Palette: "Remote-Containers: Show Container Log"

2. **Extension logs**:
   - View → Output → Select extension

3. **Docker logs**:
   ```bash
   # On host machine
   docker logs <container_id>
   ```

## Security Considerations

### Secrets Management

- Never commit secrets to the container image
- Use environment variables for tokens
- Mount secrets from host when needed
- Use Docker secrets in production

### Network Security

- Container runs in isolated network
- Only necessary ports are forwarded
- No direct host network access

### Access Control

- Container runs as non-root user (`vscode`)
- Limited sudo access for development tasks
- SSH keys mounted read-only when possible

## Additional Resources

- [VS Code Dev Containers Documentation](https://code.visualstudio.com/docs/remote/containers)
- [Docker Best Practices](https://docs.docker.com/develop/best-practices/)
- [Go Development with VS Code](https://code.visualstudio.com/docs/languages/go)
- [Node.js Development with VS Code](https://code.visualstudio.com/docs/nodejs/nodejs-tutorial)
- [Python Development with VS Code](https://code.visualstudio.com/docs/python/python-tutorial)

---

**Note**: This development container is optimized for the GZH Manager project structure and may need adjustments for other projects.
