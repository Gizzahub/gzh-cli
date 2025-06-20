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
Clone repositories by GitHub account (user, org) or GitLab group.

- bulk-clone
  - git
  - gitea
  - github
  - gitlab
  - gogs
- gen-config

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

### Bulk Clone config

```yaml
# bulk-clone.yaml 
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
