# 🚀 Deployment & Release Guide

Comprehensive guide for building, releasing, and deploying gzh-cli.

## 📋 Table of Contents

- [Release Process](#release-process)
- [Security Guidelines](#security-guidelines)
- [Distribution Methods](#distribution-methods)
- [CI/CD Integration](#cicd-integration)

## 📚 Deployment Documentation

### Release Management

- **[Release Process](70-releases.md)** - Complete release workflow and procedures
- **[Release Preparation](71-release-preparation.md)** - Pre-release checklist and validation
- **[Release Notes v1.0.0](72-release-notes-v1.0.0.md)** - Version 1.0.0 release documentation

### Security & Scanning

- **[Security Scanning](73-security-scanning.md)** - Security analysis and vulnerability scanning
- **[Security Guidelines](75-security-guidelines.md)** - Security policies and best practices

## 🎯 Quick Release Guide

### Prerequisites

- **Go 1.23.0+**
- **Git** with proper commit access
- **GPG Key** for signing releases
- **GitHub CLI** for release automation

### Standard Release Process

```bash
# 1. Prepare release
make pre-release-check
make update-version VERSION=v1.2.0

# 2. Run quality checks
make lint-all
make test-all
make security-scan

# 3. Build release artifacts
make release-build

# 4. Create and publish release
make release-publish VERSION=v1.2.0

# 5. Post-release verification
make post-release-verify
```

## 🏗️ Build System

### Local Build

```bash
# Development build
make build

# Install locally
make install

# Clean build
make clean && make build

# Debug build
make build-debug
```

### Release Build

```bash
# Cross-platform release build
make release

# Specific platform
make build-linux
make build-darwin
make build-windows

# With version and metadata
make release VERSION=v1.2.0 BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
```

### Build Artifacts

```
dist/
├── gz-linux-amd64           # Linux x86_64
├── gz-linux-arm64           # Linux ARM64
├── gz-darwin-amd64          # macOS Intel
├── gz-darwin-arm64          # macOS Apple Silicon
├── gz-windows-amd64.exe     # Windows x86_64
├── checksums.txt            # SHA256 checksums
└── gz-source.tar.gz         # Source code archive
```

## 📦 Distribution Methods

### GitHub Releases

```bash
# Create GitHub release
gh release create v1.2.0 \
  --title "Release v1.2.0" \
  --notes-file RELEASE_NOTES.md \
  --draft

# Upload artifacts
gh release upload v1.2.0 dist/*

# Publish release
gh release edit v1.2.0 --draft=false
```

### Package Managers

#### Homebrew

```ruby
# Formula update
class GzhCli < Formula
  desc "Comprehensive CLI tool for development environment management"
  homepage "https://github.com/gizzahub/gzh-cli"
  url "https://github.com/gizzahub/gzh-cli/archive/v1.2.0.tar.gz"
  sha256 "sha256-checksum-here"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "gz"
  end
end
```

#### Go Install

```bash
# Direct installation
go install github.com/gizzahub/gzh-cli@latest
go install github.com/gizzahub/gzh-cli@v1.2.0
```

#### Docker

```dockerfile
# Multi-stage build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /root/
COPY --from=builder /app/gz .
ENTRYPOINT ["./gz"]
```

## 🔒 Security & Compliance

### Security Scanning

```bash
# Vulnerability scanning
make security-scan

# Dependency audit
go mod audit

# SAST analysis
make static-analysis

# Container scanning (if using Docker)
make container-scan
```

### Code Signing

```bash
# Sign release binaries
make sign-release GPG_KEY_ID=your-key-id

# Verify signatures
make verify-signatures

# Generate checksums
make generate-checksums
```

### Supply Chain Security

```yaml
# GitHub Actions workflow for secure builds
name: Secure Release Build
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Run security checks
        run: |
          make security-scan
          make dependency-audit

      - name: Build release
        run: make release

      - name: Sign artifacts
        run: make sign-release
        env:
          GPG_PRIVATE_KEY: ${{ secrets.GPG_PRIVATE_KEY }}

      - name: Upload release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
          generate_release_notes: true
```

## 🚀 CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - run: make test-all

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make security-scan

  build:
    needs: [test, security]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - run: make release
      - uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - test
  - security
  - build
  - release

test:
  stage: test
  script:
    - make test-all

security:
  stage: security
  script:
    - make security-scan
  artifacts:
    reports:
      sast: security-report.json

build:
  stage: build
  script:
    - make release
  artifacts:
    paths:
      - dist/

release:
  stage: release
  script:
    - make release-publish
  only:
    - tags
```

## 📊 Release Metrics

### Tracking Release Success

```bash
# Download statistics
gh release view v1.2.0 --json assets

# Performance metrics
gz profile benchmark --release v1.2.0

# Usage analytics (if enabled)
gz stats export --format prometheus
```

### Release Health Monitoring

```bash
# Post-release verification
make post-release-verify

# Integration testing
make test-integration-release

# Performance regression testing
make performance-regression-test
```

## 🛠️ Environment-Specific Deployments

### Production Environment

```bash
# Production build with optimizations
make build-production

# Environment-specific configuration
export GZH_CONFIG_PATH=/etc/gzh-cli/production.yaml

# Health check endpoint
gz doctor --format json
```

### Container Deployment

```bash
# Build container image
make docker-build

# Push to registry
make docker-push REGISTRY=gcr.io/project/gzh-cli

# Deploy to Kubernetes
kubectl apply -f k8s/deployment.yaml
```

### Enterprise Deployment

```bash
# Enterprise build with additional features
make build-enterprise

# License verification
make verify-license

# Enterprise configuration
cp config/enterprise.yaml /etc/gzh-cli/gzh.yaml
```

## 🆘 Troubleshooting Deployment

### Common Build Issues

```bash
# Clean build environment
make clean-all
make bootstrap
make build

# Dependency issues
go mod tidy
go mod download
go mod verify

# Cross-compilation issues
GOOS=linux GOARCH=amd64 go build -o gz-linux-amd64
```

### Release Problems

```bash
# Verify release artifacts
make verify-release-artifacts

# Check signing
gpg --verify dist/gz-linux-amd64.sig dist/gz-linux-amd64

# Test installation
curl -L https://github.com/gizzahub/gzh-cli/releases/download/v1.2.0/gz-linux-amd64 -o gz
chmod +x gz
./gz version
```

### Security Issues

```bash
# Re-run security scans
make security-scan-detailed

# Update dependencies
go get -u all
go mod tidy

# Vulnerability assessment
make vulnerability-assessment
```

______________________________________________________________________

**Release Process**: Automated with GitHub Actions
**Security**: Code signing, vulnerability scanning, SAST
**Distribution**: GitHub Releases, Homebrew, Docker, go install
**Platforms**: Linux, macOS, Windows (x86_64, ARM64)
