Gizzahub Manager
================

<div style="text-align: center;">
Comprehensive CLI Tool
<br>
<br>
<img src="https://github.com/gizzahub/gzh-manager-go/actions/workflows/test.yml/badge.svg" alt="Test Status"/>
<img src="https://github.com/gizzahub/gzh-manager-go/actions/workflows/lint.yml/badge.svg" alt="Lint Status"/>
<img src="https://pkg.go.dev/badge/github.com/gizzahub/gzh-manager-go.svg" alt="GoDoc"/>
<img src="https://codecov.io/gh/Gizzahub/gzh-manager-go/branch/main/graph/badge.svg" alt="Code Coverage"/>
<img src="https://img.shields.io/github/v/release/Gizzahub/gzh-manager-go" alt="Latest Release"/>
<img src="https://img.shields.io/docker/pulls/Gizzahub/gzh-manager-go" alt="Docker Pulls"/>
<img src="https://img.shields.io/github/downloads/Gizzahub/gzh-manager-go/total.svg" alt="Total Downloads"/>
</div>


# Table of Contents
<!--ts-->
  * [Usage](#usage)
  * [Features](#features)
  * [Project Layout](#project-layout)
  * [How to use this template](#how-to-use-this-template)
  * [Demo Application](#demo-application)
  * [Makefile Targets](#makefile-targets)
  * [Contribute](#contribute)

<!-- Added by: morelly_t1, at: Tue 10 Aug 2021 08:54:24 AM CEST -->

<!--te-->

# Usage

## 기능
Clone repositories by GitHub account (user, org) or GitLab group and manage repository configurations at scale.

- bulk-clone
  - git
  - gitea
  - github
  - gitlab
  - gogs
- gen-config
- repo-config (GitHub repository configuration management)
  - Apply configuration templates
  - Enforce security policies
  - Audit compliance
  - Manage branch protection
  - Bulk operations

### CLI


```sh
$> bulk-clone -h
golang-cli cli application by managing bulk-clone

Usage:
  gzh [flags]
  gzh [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  bulk-clone    Clone repositories in bulk
  help        Help about any command
  version     bulk-clone version

Flags:
  -h, --help   help for bulk-clone

Use "gzh-manager [command] --help" for more information about a command.
```

First, create a configuration file in the desired path. Refer to
[bulk-clone.yaml](pkg/bulk-clone/bulk-clone.yaml)

```sh
$> gzh bulk-clone -t $HOME/mywork

This won't work:
$> gzh bulk-clone -t ./mywork
$> gzh bulk-clone -t $HOME/mywork
$> gzh bulk-clone -t ~/mywork
```

### Bulk Clone Config File Support

The bulk-clone command now supports configuration files to manage multiple organizations and their settings. This allows you to define clone operations once and reuse them.

#### Configuration File Locations

The tool searches for configuration files in the following order:
1. Environment variable: `GZH_CONFIG_PATH`
2. Current directory: `./bulk-clone.yaml` or `./bulk-clone.yml`
3. User config directory: `~/.config/gzh-manager/bulk-clone.yaml`
4. System config: `/etc/gzh-manager/bulk-clone.yaml`

#### Using Configuration Files

```bash
# Use config file from standard locations
gzh bulk-clone github --use-config -o myorg

# Use specific config file
gzh bulk-clone github -c /path/to/config.yaml -o myorg

# Override config values with CLI flags
gzh bulk-clone github -c config.yaml -o myorg -t /different/path
```

#### Configuration File Examples

Several example configuration files are provided in the `samples/` directory:

1. **bulk-clone-simple.yaml** - A minimal working configuration
2. **bulk-clone-example.yaml** - A comprehensive example with detailed comments
3. **bulk-clone.yml** - Advanced features (planned/future implementation)

##### Simple Configuration Example

```yaml
# bulk-clone-simple.yaml
version: "0.1"

default:
  protocol: https
  github:
    root_path: "$HOME/github-repos"
  gitlab:
    root_path: "$HOME/gitlab-repos"

repo_roots:
  - root_path: "$HOME/work/mycompany"
    provider: "github"
    protocol: "ssh"
    org_name: "mycompany"
  
  - root_path: "$HOME/opensource"
    provider: "github"
    protocol: "https"
    org_name: "kubernetes"

ignore_names:
  - "test-.*"
  - ".*-archive"
```

See `samples/bulk-clone-example.yaml` for a comprehensive example with all available options and detailed comments.

#### Configuration Schema

The configuration file structure is formally defined in:
- **JSON Schema**: `docs/bulk-clone-schema.json` - Machine-readable schema definition
- **YAML Schema**: `docs/bulk-clone-schema.yaml` - Human-readable schema documentation

##### Validating Your Configuration

You can validate your configuration file using the built-in validator:

```bash
# Validate a specific config file
gzh bulk-clone validate -c /path/to/bulk-clone.yaml

# Validate config from standard locations
gzh bulk-clone validate --use-config
```

The validator checks:
- Required fields are present
- Values match allowed enums (protocol, provider, etc.)
- Structure follows the schema
- Regex patterns are valid

#### Advanced Configuration (Future)

```yaml
# bulk-clone.yaml (advanced example for future implementation)
github:
  ScriptonBasestar:
   auth: token
   proto: https
   targetPath: $HOME/mywork/ScriptonBasestar
   default:
    strategy: include
    branch: develop
   include:
    proxynd:
      branch: develop
    devops-minim-engine:
      branch: dev
   exclude:
    - sb-wp-*
   override:
    include:
  nginxinc:
   targetPath: $HOME/mywork/nginxinc
```

```bash
gzh bulk-clone -o nginxinc
gzh bulk-clone -o nginxinc -t $HOME/mywork/nginxinc
gzh bulk-clone -o nginxinc -t $HOME/mywork/nginxinc --auth token
gzh bulk-clone -o nginxinc -t $HOME/mywork/nginxinc -s pull
```

### Strategy Options

The `-s` or `--strategy` flag controls how existing repositories are synchronized:

- `reset` (default): Performs `git reset --hard HEAD` followed by `git pull`. This discards all local changes and ensures a clean sync with the remote repository.
- `pull`: Only performs `git pull` without resetting. This attempts to merge remote changes with local changes. May fail if there are conflicts.
- `fetch`: Only performs `git fetch` without modifying the working directory. This updates remote tracking branches but doesn't change your local files.

Example usage:
```bash
# Default behavior (reset strategy)
gzh bulk-clone github -o myorg -t ~/repos

# Preserve local changes and merge with remote
gzh bulk-clone github -o myorg -t ~/repos -s pull

# Only fetch updates without modifying local files
gzh bulk-clone github -o myorg -t ~/repos -s fetch
```

## Repository Configuration Management

The `gz repo-config` command allows you to manage GitHub repository configurations at scale, including settings, security policies, branch protection rules, and compliance auditing.

### Quick Start

1. **Create a configuration file** (`repo-config.yaml`):
   ```yaml
   version: "1.0.0"
   organization: "your-org"
   
   templates:
     standard:
       description: "Standard repository settings"
       settings:
         has_issues: true
         has_wiki: false
         delete_branch_on_merge: true
       security:
         vulnerability_alerts: true
         branch_protection:
           main:
             required_reviews: 2
             enforce_admins: true
   
   repositories:
     - name: "*"
       template: "standard"
   ```

2. **Apply configuration**:
   ```bash
   # Preview changes (dry run)
   gz repo-config apply --config repo-config.yaml --dry-run
   
   # Apply configuration
   gz repo-config apply --config repo-config.yaml
   ```

3. **Audit compliance**:
   ```bash
   gz repo-config audit --config repo-config.yaml
   ```

### Key Features

- **Templates**: Define reusable repository configurations
- **Policies**: Enforce security and compliance rules
- **Pattern Matching**: Apply configurations based on repository name patterns
- **Exception Handling**: Allow documented exceptions to policies
- **Compliance Auditing**: Generate reports on policy violations
- **Bulk Operations**: Update multiple repositories efficiently

### Documentation

- [Quick Start Guide](docs/repo-config-quick-start.md) - Get started in 5 minutes
- [User Guide](docs/repo-config-user-guide.md) - Complete documentation
- [Policy Examples](docs/repo-config-policy-examples.md) - Ready-to-use policy templates
- [Configuration Schema](docs/repo-config-schema.yaml) - Configuration file reference

### Example: Enterprise Configuration

```yaml
version: "1.0.0"
organization: "enterprise-org"

templates:
  backend:
    description: "Backend service configuration"
    settings:
      private: true
    security:
      secret_scanning: true
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks: ["ci/build", "ci/test"]

policies:
  security:
    description: "Security requirements"
    rules:
      must_be_private:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "Production services must be private"

patterns:
  - pattern: "*-service"
    template: "backend"
    policies: ["security"]
```

## Trigger

와이파이 변경.. 등

# Features
- [goreleaser](https://goreleaser.com/) with `deb.` and `.rpm` packer and container (`docker.hub` and `ghcr.io`) releasing including `manpages` and `shell completions` and grouped Changelog generation.
- [golangci-lint](https://golangci-lint.run/) for linting and formatting
- [Github Actions](.github/worflows) Stages (Lint, Test (`windows`, `linux`, `mac-os`), Build, Release) 
- [Gitlab CI](.gitlab-ci.yml) Configuration (Lint, Test, Build, Release)
- [cobra](https://cobra.dev/) example setup including tests
- [Makefile](Makefile) - with various useful targets and documentation (see Makefile Targets)
- [Github Pages](_config.yml) using [jekyll-theme-minimal](https://github.com/pages-themes/minimal) (checkout [https://Gizzahub.github.io/gzh-manager-go/](https://Gizzahub.github.io/gzh-manager-go/))
- Useful `README.md` badges
- [pre-commit-hooks](https://pre-commit.com/) for formatting and validating code before committing

## Project Layout
* [assets/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/assets) => docs, images, etc
* [cmd/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/cmd)  => command-line configurations (flags, subcommands)
* [pkg/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/pkg)  => packages that are okay to import for other projects
* [internal/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/pkg)  => packages that are only for project internal purposes
- [`tools/`](tools/) => for automatically shipping all required dependencies when running `go get` (or `make bootstrap`) such as `golang-ci-lint` (see: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
- [`scripts/`](scripts/) => build scripts 

# Makefile Targets
```sh
$> make
bootstrap                      install build deps
build                          build golang binary
clean                          clean up environment
cover                          display test coverage
docker-build                   dockerize golang application
fmt                            format go files
help                           list makefile targets
install                        install golang binary
lint                           lint go files
pre-commit                     run pre-commit hooks
run                            run the app
test                           display test coverage
```
