# Release Process

This document describes the automated release process for the gz CLI tool using GoReleaser and GitHub Actions.

## Overview

The project uses a fully automated release pipeline that:

1. **Builds** cross-platform binaries for Linux, macOS, and Windows
2. **Packages** releases as archives, Linux packages (deb/rpm/apk), and container images
3. **Publishes** to multiple distribution channels (GitHub Releases, Docker Hub, Homebrew, etc.)
4. **Signs** artifacts with Cosign for supply chain security
5. **Announces** releases via Slack/Discord webhooks

## Release Channels

### Package Managers

| Platform | Package Manager | Installation Command |
|----------|----------------|---------------------|
| **macOS** | Homebrew | `brew install gizzahub/tap/gz` |
| **Windows** | Chocolatey | `choco install gz` |
| **Windows** | Scoop | `scoop bucket add gizzahub https://github.com/Gizzahub/scoop-bucket && scoop install gz` |
| **Arch Linux** | AUR | `yay -S gz-bin` |
| **Linux** | APT (deb) | `dpkg -i gz_*.deb` |
| **Linux** | YUM/DNF (rpm) | `rpm -i gz_*.rpm` |
| **Alpine** | APK | `apk add gz_*.apk` |

### Container Images

| Registry | Image | Pull Command |
|----------|-------|--------------|
| **Docker Hub** | `gizzahub/gzh-manager-go` | `docker pull gizzahub/gzh-manager-go:latest` |
| **GitHub Container Registry** | `ghcr.io/gizzahub/gzh-manager-go` | `docker pull ghcr.io/gizzahub/gzh-manager-go:latest` |

### Direct Downloads

- **GitHub Releases**: Pre-built binaries for all platforms
- **Source Code**: Available as tarball and zip archives

## Release Workflow

### Automated Release (Recommended)

1. **Create and push a git tag**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions automatically**:
   - Runs CI tests and security scans
   - Builds cross-platform binaries
   - Creates packages for all supported platforms
   - Builds and pushes container images
   - Signs artifacts with Cosign
   - Creates GitHub Release with changelog
   - Publishes to package managers
   - Sends notifications

### Manual Release (Development)

For testing releases locally:

```bash
# Install goreleaser
make install-goreleaser

# Check configuration
make release-check

# Dry run (no publishing)
make release-dry-run

# Create snapshot (local testing)
make release-snapshot
```

## Versioning

The project follows [Semantic Versioning (SemVer)](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., `v1.2.3`)
- **Pre-releases**: `v1.2.3-rc.1`, `v1.2.3-beta.1`, `v1.2.3-alpha.1`

### Version Examples

| Version | Type | Description |
|---------|------|-------------|
| `v1.0.0` | Major | Breaking changes, new major features |
| `v1.1.0` | Minor | New features, backward compatible |
| `v1.1.1` | Patch | Bug fixes, security updates |
| `v1.2.0-rc.1` | Release Candidate | Pre-release testing |
| `v1.2.0-beta.1` | Beta | Feature complete, testing |
| `v1.2.0-alpha.1` | Alpha | Early development, unstable |

## Release Configuration

### GoReleaser Configuration

The release process is configured in `.goreleaser.yml`:

- **Builds**: Cross-platform binaries with optimized ldflags
- **Archives**: Compressed releases with documentation
- **Packages**: Native packages for Linux distributions
- **Docker**: Multi-architecture container images
- **Signing**: Cosign signatures for supply chain security
- **Distribution**: Multiple package managers and registries

### CI/CD Pipeline

GitHub Actions workflows in `.github/workflows/`:

- **`ci.yml`**: Continuous integration (tests, linting, security)
- **`release.yml`**: Automated release process
- **`dependabot-auto-merge.yml`**: Automatic dependency updates

## Build Information

Each release includes build metadata:

```bash
gz version
# Output: gz version v1.0.0
```

Build-time information embedded in binaries:
- **Version**: Git tag (e.g., `v1.0.0`)
- **Commit**: Git commit SHA
- **Date**: Build timestamp
- **Built By**: Build system (e.g., `goreleaser`)

## Security

### Artifact Signing

All release artifacts are signed with [Cosign](https://github.com/sigstore/cosign):

- **Checksums**: Signed with keyless signing
- **Container Images**: Signed with OIDC identity
- **Verification**: Public transparency log

```bash
# Verify container image signature
cosign verify ghcr.io/gizzahub/gzh-manager-go:v1.0.0

# Verify checksum signature
cosign verify-blob --certificate checksums.txt.pem --signature checksums.txt.sig checksums.txt
```

### Supply Chain Security

- **SBOM**: Software Bill of Materials included
- **Provenance**: Build provenance attestation
- **Vulnerability Scanning**: Automated security scanning
- **Dependency Updates**: Automated with Dependabot

## Environment Variables

Required secrets for automated releases:

| Secret | Purpose | Required |
|--------|---------|----------|
| `GITHUB_TOKEN` | GitHub API access | ✅ |
| `DOCKERHUB_USERNAME` | Docker Hub publishing | ✅ |
| `DOCKERHUB_TOKEN` | Docker Hub authentication | ✅ |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Homebrew formula updates | Optional |
| `SCOOP_BUCKET_GITHUB_TOKEN` | Scoop manifest updates | Optional |
| `AUR_KEY` | Arch Linux AUR publishing | Optional |
| `SLACK_WEBHOOK_URL` | Slack notifications | Optional |
| `DISCORD_WEBHOOK_URL` | Discord notifications | Optional |

## Changelog Generation

Changelogs are automatically generated from commit messages:

### Commit Message Format

Follow [Conventional Commits](https://conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Examples

```bash
feat(cli): add new bulk-clone command
fix(config): resolve validation error
docs(readme): update installation instructions
chore(deps): bump golang.org/x/text from 0.3.7 to 0.3.8
```

### Changelog Sections

- **New Features**: `feat:` commits
- **Bug Fixes**: `fix:` commits  
- **Security Updates**: `sec:` commits
- **Performance Improvements**: `perf:` commits
- **Documentation Updates**: `docs:` commits
- **Dependency Updates**: `feat(deps):` or `fix(deps):` commits

## Testing Releases

### Pre-Release Testing

Before tagging a release:

1. **Run full test suite**:
   ```bash
   make test-all
   ```

2. **Test release configuration**:
   ```bash
   make release-check
   make release-dry-run
   ```

3. **Build and test binary**:
   ```bash
   make build
   ./gz version
   ./gz --help
   ```

### Post-Release Verification

After release:

1. **Verify GitHub Release** was created
2. **Test installation** from package managers
3. **Pull container images** and test
4. **Check artifact signatures**

## Troubleshooting

### Common Issues

1. **GoReleaser fails**:
   ```bash
   # Check configuration
   make release-check
   
   # Verify scripts are executable
   chmod +x scripts/*.sh
   ```

2. **Package manager publishing fails**:
   - Check repository tokens and permissions
   - Verify tap/bucket repositories exist
   - Ensure proper branch protection rules

3. **Container image push fails**:
   - Verify Docker Hub credentials
   - Check repository permissions
   - Ensure registry authentication

### Debug Commands

```bash
# Verbose goreleaser output
goreleaser release --debug

# Test specific publisher
goreleaser release --publisher docker

# Skip specific steps
goreleaser release --skip=docker,homebrew
```

## Best Practices

1. **Test thoroughly** before tagging releases
2. **Use semantic versioning** consistently
3. **Write clear commit messages** for better changelogs
4. **Review generated artifacts** before publishing
5. **Monitor release metrics** and user feedback
6. **Keep documentation updated** with each release
7. **Coordinate major releases** with team announcements

## Release Metrics

Track these metrics for each release:

- **Download statistics** from GitHub Releases
- **Package installation counts** from registries
- **Container image pulls** from Docker Hub/GHCR
- **User feedback** and issue reports
- **Security scan results** and vulnerability assessments

## Future Enhancements

Planned improvements:

- **Multi-stage releases** with beta/rc channels
- **Automated rollback** on critical issues
- **Release notes automation** with AI assistance
- **Performance benchmarks** in release pipeline
- **User notification system** for major updates