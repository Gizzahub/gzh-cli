# 📦 Installation Guide

This guide covers the installation and initial setup of gzh-cli (`gz`).

## 🚀 Quick Installation

### Prerequisites

- **Go 1.24.0+** (with toolchain go1.24.5)
- **Git** (any recent version)
- **Network access** to Git platforms (GitHub, GitLab, etc.)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/gizzahub/gzh-cli.git
cd gzh-cli

# Install build dependencies
make bootstrap

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Verify Installation

```bash
# Check if gz is available
gz --version

# Test basic functionality
gz --help
```

## 🔧 Initial Setup

### 1. Environment Configuration

Set up authentication tokens for the Git platforms you plan to use:

```bash
# GitHub
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# GitLab
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"

# Gitea/Gogs (if needed)
export GITEA_TOKEN="your_gitea_token"
export GOGS_TOKEN="your_gogs_token"
```

### 2. Configuration Directory

Create the configuration directory:

```bash
mkdir -p ~/.config/gzh-manager
```

### 3. Basic Configuration File

Create a basic configuration file:

```bash
cat > ~/.config/gzh-manager/gzh.yaml << 'EOF'
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "your-org"
        clone_dir: "$HOME/repos/github/your-org"
EOF
```

## 🧪 Test Your Installation

### Quick Test Commands

```bash
# System diagnostics
gz doctor

# Configuration validation
gz config validate

# List available commands
gz --help

# Test repository cloning (optional)
gz git repo clone-or-update https://github.com/octocat/Hello-World.git /tmp/test-repo
```

## 🔨 Development Setup (Optional)

If you plan to contribute or develop gzh-cli:

### Additional Dependencies

```bash
# Install development tools
make pre-commit-install

# Run code quality checks
make lint-all

# Run tests
make test
```

### IDE Integration

For JetBrains IDEs or VS Code:

```bash
# Start IDE monitoring (if using JetBrains IDEs)
gz ide monitor &

# This will help sync IDE settings and detect configuration issues
```

## 🌍 Platform-Specific Notes

### Linux

- Ensure `$GOPATH/bin` is in your `PATH`
- Some distributions may require additional packages for Git

### macOS

- Works on both Intel and Apple Silicon
- Consider using Homebrew for Go installation
- Ensure Xcode Command Line Tools are installed

### Windows

- **WSL recommended** for the best experience
- PowerShell support available but limited
- Ensure Git for Windows is properly configured

## 🔗 Network Configuration

### Proxy Support

If you're behind a corporate proxy:

```bash
# Set proxy environment variables
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="http://proxy.company.com:8080"
export NO_PROXY="localhost,127.0.0.1,.company.com"
```

### SSH Key Setup

For Git operations over SSH:

```bash
# Ensure your SSH key is added to ssh-agent
ssh-add ~/.ssh/id_rsa

# Test SSH connectivity
ssh -T git@github.com
```

## 🆘 Troubleshooting

### Common Issues

#### "gz: command not found"

```bash
# Check if $GOPATH/bin is in PATH
echo $PATH | grep -o $GOPATH/bin

# If not, add to your shell profile
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Token Authentication Issues

```bash
# Test token validity
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# For GitLab
curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" https://gitlab.com/api/v4/user
```

#### Build Issues

```bash
# Clean and rebuild
make clean
make bootstrap
make build
```

### Getting Help

- **Command Help**: `gz <command> --help`
- **System Diagnostics**: `gz doctor`
- **Configuration Check**: `gz config validate`
- **Verbose Logging**: Add `--verbose` or `--debug` to any command

## 📋 Next Steps

After installation, proceed to:

1. **[Quick Start Guide](11-quick-start.md)** - Try your first commands
1. **[Configuration Guide](../40-configuration/40-configuration-guide.md)** - Detailed configuration
1. **[Command Reference](../50-api-reference/50-command-reference.md)** - Complete command documentation

______________________________________________________________________

**Installation Requirements**: Go 1.24.0+, Git
**Supported Platforms**: Linux, macOS, Windows (WSL)
**Last Updated**: 2025-08-19
