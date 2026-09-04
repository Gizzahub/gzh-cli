# Release Process

This document describes the automated release process for the gz CLI tool using GoReleaser and GitHub Actions.

## Overview

The project uses a fully automated release pipeline that:

1. **Builds** cross-platform binaries for Linux, macOS, and Windows
1. **Packages** releases as archives, Linux packages (deb/rpm/apk), and container images
1. **Publishes** to multiple distribution channels (GitHub Releases, Docker Hub, Homebrew, etc.)
1. **Signs** artifacts with Cosign for supply chain security
1. **Leaves notifications disabled** until the protected workflow supplies provider credentials

## Release Channels

### Package Managers

| Platform       | Package Manager | Installation Command                                                                           |
| -------------- | --------------- | ---------------------------------------------------------------------------------------------- |
| **macOS**      | Homebrew Cask   | `brew install --cask gizzahub/tap/gz`                                                          |
| **Windows**    | Chocolatey      | _Not supported_ — see [Chocolatey: explicitly unsupported](#chocolatey-explicitly-unsupported) |
| **Windows**    | Scoop           | `scoop bucket add gizzahub https://github.com/gizzahub/scoop-bucket && scoop install gz`       |
| **Arch Linux** | AUR             | `yay -S gz-bin`                                                                                |
| **Linux**      | APT (deb)       | `dpkg -i gz_*.deb`                                                                             |
| **Linux**      | YUM/DNF (rpm)   | `rpm -i gz_*.rpm`                                                                              |
| **Alpine**     | APK             | `apk add gz_*.apk`                                                                             |

Upgrade a Cask installation with `brew upgrade --cask gz`.

The generated Cask removes `com.apple.quarantine` from the staged `gz` binary in a
macOS-only postflight hook because the binary is not Apple code-signed or notarized.
This is separate from Cosign artifact signing. Snapshot validation checks the generated
Cask structure and binary locally; an actual tap publish and `brew install --cask` remain
external release validation steps.

### Container Images

| Registry                      | Image                      | Pull Command                                  |
| ----------------------------- | -------------------------- | --------------------------------------------- |
| **Docker Hub**                | `gizzahub/gzh-cli`         | `docker pull gizzahub/gzh-cli:latest`         |
| **GitHub Container Registry** | `ghcr.io/gizzahub/gzh-cli` | `docker pull ghcr.io/gizzahub/gzh-cli:latest` |

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

1. **GitHub Actions automatically**:

   - Runs CI tests and security scans
   - Builds cross-platform binaries
   - Creates packages for all supported platforms
   - Builds and pushes container images
   - Signs artifacts with Cosign
   - Creates GitHub Release with changelog
   - Publishes to package managers
   - Leaves Slack and Discord notifications disabled pending protected workflow wiring

### Manual Release (Development)

For testing releases locally:

```bash
# Install the exact pinned toolchain into bin/tools (gitignored)
make install-release-tools

# Fail fast if any pinned tool is missing or has drifted
make verify-release-tools

# Same check plus `goreleaser healthcheck`, which asks GoReleaser itself
# which external binaries the configured pipes need
make release-healthcheck

# Validate .goreleaser.yml with the pinned GoReleaser
make release-check

# Dry run: builds and Docker images, no publish, no signing
make release-dry-run

# Snapshot: publish, Docker and signing excluded — the routine local gate
make release-snapshot
```

`sign` is excluded from both local targets deliberately. Cosign signing here is
keyless: it mints a Fulcio certificate from an ambient OIDC token and writes an
entry to the public Rekor transparency log, which cannot be undone. The release
job has that ambient token (`id-token: write`); a laptop does not, so locally
cosign would fall back to an interactive browser flow and, if completed, would
publish a throwaway snapshot to a public log. The signing configuration is
instead gated by `goreleaser check` and by `release-healthcheck`, which fail if
cosign is absent.

## Release Toolchain Pins

The release is only reproducible if every tool in it is an exact version. None
of these resolve to `latest`.

| Tool           | Pinned version | Upstream evidence                                                                                |
| -------------- | -------------- | ------------------------------------------------------------------------------------------------ |
| **GoReleaser** | `v2.18.0`      | [Release v2.18.0](https://github.com/goreleaser/goreleaser/releases/tag/v2.18.0) — latest stable |
| **syft**       | `v1.51.1`      | [Release v1.51.1](https://github.com/anchore/syft/releases/tag/v1.51.1) — latest stable          |
| **cosign**     | `v3.1.3`       | [Release v3.1.3](https://github.com/sigstore/cosign/releases/tag/v3.1.3) — latest stable         |

Each version is declared in exactly two places, which must be changed together:

| Where                           | How                                                      |
| ------------------------------- | -------------------------------------------------------- |
| `.make/tools.mk`                | `GORELEASER_VERSION` / `SYFT_VERSION` / `COSIGN_VERSION` |
| `.github/workflows/release.yml` | the job-level `env:` block of the same three names       |

Locally the tools are installed into the gitignored `bin/tools/` with
`go install <module>@<version>` and verified with `go version -m`, which reports
the true module version even for a binary built without the upstream release
ldflags. CI installs the official release binaries through pinned actions and
verifies them with each tool's own `version` output. Both paths are exact.

### Why GoReleaser v2.18.0

- The module path for GoReleaser v2 is `github.com/goreleaser/goreleaser/v2`.
  The unsuffixed `github.com/goreleaser/goreleaser` module stops at `v1.26.2`,
  so a `go install github.com/goreleaser/goreleaser@latest` silently installs
  **v1**, which cannot read a `version: 2` config.
- `homebrew_casks[].binary` was replaced by `binaries` in v2.13.0 and is now a
  hard `goreleaser check` failure ("configuration is valid, but uses deprecated
  properties"), so the config uses `binaries`.
- v2.18.0 carries fixes that matter to this config directly: `fix(sign): artifacts: none masks real signing failures`, `fix(cask): a cask without a repository stops the ones after it`, and `fix(cask): emit Casks that pass brew style`.

### Why cosign v3.1.3 changed the signing config

cosign 3.x removed the working `--output-certificate` / `--output-signature`
path for `sign-blob`; invoking it now fails with `must specify --bundle with --new-bundle-format`. The `signs` block therefore follows the current upstream
GoReleaser example and emits a single Sigstore bundle:

```yaml
signs:
  - cmd: cosign
    signature: "${artifact}.sigstore.json"
    args: ["sign-blob", "--bundle=${signature}", "${artifact}", "--yes"]
    artifacts: checksum
```

This changes the published signature artifact from `checksums.txt.pem` plus
`checksums.txt.sig` to a single `checksums.txt.sigstore.json`. No public release
has shipped yet, so no existing verification instructions break.

### Chocolatey: explicitly unsupported

The `chocolateys` pipe shells out to the `choco` CLI, which GoReleaser does not
install ("GoReleaser will not install `chocolatey`/`choco` nor any of its
dependencies for you"). `choco` is a Windows-only .NET tool and is not present
on the `ubuntu-latest` release runner, so with the section configured
`goreleaser healthcheck` exits 1 with `choco - not present in path`.

Splitting the pipe into a separate Windows job requires `goreleaser release --split` plus `goreleaser continue`, which are GoReleaser **Pro** features and
are unavailable to this OSS distribution.

**Decision**: Chocolatey is an explicitly unsupported publish channel. The
`chocolateys` section has been removed from `.goreleaser.yml` rather than
skipped at run time with `--skip=chocolatey`, so the configuration and the
healthcheck agree on what the release actually ships and nothing is dropped
silently.

To adopt it later: move the release job (or a dedicated packaging job) to a
`windows-latest` runner with `choco` on `PATH`, restore the `chocolateys`
section from git history (it is preserved in the commit that removed it, and
`.goreleaser.yml` carries a pointer comment at the removal site), and add
`CHOCOLATEY_API_KEY` to the job env and to the credential preflight list.

### Toolchain failure vs. missing credentials

The release job runs two separate preflights so the two failure classes can
never be confused:

1. **`Preflight — publishing credentials`** runs first, before any tool is
   installed. It asserts that every secret `.goreleaser.yml` consumes is
   non-empty, reports all missing names at once, and never echoes a value. Its
   messages are prefixed `CREDENTIAL FAILURE` and state explicitly that the
   toolchain is fine.
1. **`Preflight — release toolchain`** runs after installation. It asserts the
   exact version of each tool on `PATH` and then runs `goreleaser healthcheck`,
   which fails on any binary the configured pipes need. Its messages are
   prefixed `TOOLCHAIN FAILURE`.

A missing secret therefore never looks like a broken toolchain, and a drifted
tool never looks like a missing secret.

## Versioning

The project follows [Semantic Versioning (SemVer)](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., `v1.2.3`)
- **Pre-releases**: `v1.2.3-rc.1`, `v1.2.3-beta.1`, `v1.2.3-alpha.1`

### Version Examples

| Version          | Type              | Description                          |
| ---------------- | ----------------- | ------------------------------------ |
| `v1.0.0`         | Major             | Breaking changes, new major features |
| `v1.1.0`         | Minor             | New features, backward compatible    |
| `v1.1.1`         | Patch             | Bug fixes, security updates          |
| `v1.2.0-rc.1`    | Release Candidate | Pre-release testing                  |
| `v1.2.0-beta.1`  | Beta              | Feature complete, testing            |
| `v1.2.0-alpha.1` | Alpha             | Early development, unstable          |

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
# Output: gz version 1.0.0
```

GoReleaser removes the tag's `v` prefix and embeds the resulting release
version (for example, `1.0.0`) in `internal/version.Version`. The current
`gz version` contract does not expose separate commit, build-date, or builder
fields.

## Security

### Artifact Signing

All release artifacts are signed with [Cosign](https://github.com/sigstore/cosign):

- **Checksums**: Signed with keyless signing, emitted as a `.sigstore.json` bundle
- **Container Images**: Signed with OIDC identity
- **Verification**: Public transparency log

Keyless verification has no public key to pin, so cosign requires you to say
*which signer identity you trust*; it refuses to run without one rather than
accepting any valid Sigstore signature from anyone. Both commands therefore
carry `--certificate-identity-regexp` (the workflow that signed the release) and
`--certificate-oidc-issuer` (the OIDC provider that vouched for it). Omitting
either fails immediately with `--certificate-identity or --certificate-identity-regexp is required for verification in keyless mode`.

```bash
# Verify container image signature
cosign verify ghcr.io/gizzahub/gzh-cli:v1.0.0 \
  --certificate-identity-regexp 'https://github.com/Gizzahub/gzh-cli/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

# Verify checksum signature (Sigstore bundle; see "Why cosign v3.1.3" above)
cosign verify-blob --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/Gizzahub/gzh-cli/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
```

The regexp is deliberately scoped to this repository. Widening it to `.*` would
accept a signature from any GitHub Actions workflow anywhere, which verifies
that *something* signed the artifact but not that *we* did.

### Supply Chain Security

- **SBOM**: Software Bill of Materials included
- **Provenance**: Build provenance attestation
- **Vulnerability Scanning**: Automated security scanning
- **Dependency Updates**: Automated with Dependabot

## Environment Variables

Required secrets for automated releases:

| Secret                        | Consumed by                                | Required |
| ----------------------------- | ------------------------------------------ | -------- |
| `GITHUB_TOKEN`                | GitHub Release, provided automatically     | ✅       |
| `DOCKERHUB_USERNAME`          | `docker login` for the `dockers` pipes     | ✅       |
| `DOCKERHUB_TOKEN`             | `docker login` for the `dockers` pipes     | ✅       |
| `HOMEBREW_TAP_GITHUB_TOKEN`   | `homebrew_casks[].repository.token`        | ✅       |
| `SCOOP_BUCKET_GITHUB_TOKEN`   | `scoops[].repository.token`                | ✅       |
| `AUR_KEY`                     | `aurs[].private_key`                       | ✅       |
| `GORELEASER_MAINTAINER_EMAIL` | `nfpms[].maintainer`, `aurs[].maintainers` | ✅       |

The credential preflight enforces exactly **six** of these before any tool is
installed: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `HOMEBREW_TAP_GITHUB_TOKEN`,
`SCOOP_BUCKET_GITHUB_TOKEN`, `AUR_KEY` and `GORELEASER_MAINTAINER_EMAIL`. Each is
referenced unconditionally by a configured pipe in `.goreleaser.yml`, so a
missing value is a failed release, not a skipped channel. A value that is only
whitespace counts as missing. If a channel should genuinely not ship, remove
its pipe from `.goreleaser.yml` rather than leaving its secret unset.

`GITHUB_TOKEN` is not preflight-checked: it is injected by the Actions runner,
so it is always present and there is nothing for a preflight to catch. What can
go wrong is its *permissions*, not its existence — hence the job-level
`packages: write` and `id-token: write` grants. A rejected login is reported by
the registry login step as a `CREDENTIAL FAILURE`.

`nfpms[].maintainer` and `aurs[].maintainers` use `{{ .Env.GORELEASER_MAINTAINER_EMAIL }}`.
GoReleaser errors if that key is unset, so a local `make release-snapshot`
without the env fails closed instead of publishing a literal `${...}` string.
Export a value for local snapshots; CI gets it from the repository secret.

Slack and Discord announcements are deliberately disabled in
`.goreleaser.yml`. Enabling them requires a separately approved protected
workflow change that passes the exact credentials expected by each GoReleaser
provider. A repository secret alone does not enable notifications.

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

1. **Test release configuration**:

   ```bash
   make release-check
   make release-dry-run
   ```

1. **Build and test binary**:

   ```bash
   make build
   ./gz version
   ./gz --help
   ```

### Post-Release Verification

After release:

1. **Verify GitHub Release** was created
1. **Test installation** from package managers
1. **Pull container images** and test
1. **Check artifact signatures**

## Troubleshooting

### Common Issues

1. **GoReleaser fails**:

   ```bash
   # Check configuration
   make release-check

   # Verify scripts are executable
   chmod +x scripts/*.sh
   ```

1. **Package manager publishing fails**:

   - Check repository tokens and permissions
   - Verify tap/bucket repositories exist
   - Ensure proper branch protection rules

1. **Container image push fails**:

   - Verify Docker Hub credentials
   - Check repository permissions
   - Ensure registry authentication

### Debug Commands

```bash
# Which external binaries do the configured pipes need?
bin/tools/goreleaser healthcheck

# Verbose goreleaser output
bin/tools/goreleaser release --verbose

# Skip specific steps
bin/tools/goreleaser release --skip=docker,homebrew
```

Always invoke the pinned copy in `bin/tools/` rather than whatever `goreleaser`
resolves to on `PATH`; the make targets do this for you and additionally put
`bin/tools` first on `PATH` so GoReleaser picks up the pinned `syft` and
`cosign` when it shells out to them by name.

## Best Practices

1. **Test thoroughly** before tagging releases
1. **Use semantic versioning** consistently
1. **Write clear commit messages** for better changelogs
1. **Review generated artifacts** before publishing
1. **Monitor release metrics** and user feedback
1. **Keep documentation updated** with each release
1. **Coordinate major releases** with team announcements

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
