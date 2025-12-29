# 🚀 Architecture Improvement Strategies (Non-Plugin)

플러그인 구조 전환 없이 현재 아키텍처를 개선하는 실용적 방법들을 제시합니다.

## 📋 Table of Contents

- [Current State Analysis](#current-state-analysis)
- [Improvement Strategy 1: Enhanced Registry System](#improvement-strategy-1-enhanced-registry-system)
- [Improvement Strategy 2: Build Variants](#improvement-strategy-2-build-variants)
- [Improvement Strategy 3: Command Lifecycle Management](#improvement-strategy-3-command-lifecycle-management)
- [Improvement Strategy 4: Configuration-Based Extensions](#improvement-strategy-4-configuration-based-extensions)
- [Improvement Strategy 5: Wrapper Pattern Enhancement](#improvement-strategy-5-wrapper-pattern-enhancement)
- [Implementation Priority](#implementation-priority)
- [Migration Guide](#migration-guide)

## 🔍 Current State Analysis

### Metrics

| Metric                 | Current Value                    | Status             |
| ---------------------- | -------------------------------- | ------------------ |
| **Total Commands**     | 11 core commands                 | ✅ Manageable      |
| **Go Files**           | 173 files in `cmd/`              | ✅ Well-organized  |
| **Binary Size**        | ~33MB (estimated)                | ✅ Acceptable      |
| **External Libraries** | 4 (git, quality, pm, shellforge) | ✅ Good separation |
| **Registry Pattern**   | Simple interface                 | ⚠️ Can be enhanced |

### Strengths

- ✅ Clean separation with external libraries (pm, quality, shellforge, git operations)
- ✅ Consistent registration pattern across all commands
- ✅ Single binary distribution
- ✅ Good modular structure

### Improvement Opportunities

- ⚠️ Registry lacks metadata (categories, priority, lifecycle)
- ⚠️ No conditional compilation for optional commands
- ⚠️ Limited command discovery and introspection
- ⚠️ No user-extensibility (aliases, custom shortcuts)
- ⚠️ Build always includes all commands

## 🎯 Improvement Strategy 1: Enhanced Registry System

### Goal

Add metadata and lifecycle management to command registry without changing plugin architecture.

### Implementation

#### 1.1 Rich Metadata Interface

```go
// cmd/registry/registry.go
type CommandProvider interface {
    Command() *cobra.Command
    Metadata() CommandMetadata
}

type CommandMetadata struct {
    Name         string          // Command name (e.g., "git")
    Category     CommandCategory // Grouping
    Version      string          // Command version
    Priority     int             // Display/execution order (lower = higher priority)
    Experimental bool            // Mark experimental features
    Dependencies []string        // Required external tools
    Tags         []string        // Searchable tags
    Lifecycle    LifecycleStage  // Development stage
}

type CommandCategory string

const (
    CategoryGit         CommandCategory = "git"         // Git operations
    CategoryDevelopment CommandCategory = "development" // dev-env, ide, pm
    CategoryQuality     CommandCategory = "quality"     // quality, doctor
    CategoryNetwork     CommandCategory = "network"     // net-env
    CategoryUtility     CommandCategory = "utility"     // profile, shell
    CategoryConfig      CommandCategory = "config"      // repo-config, synclone
)

type LifecycleStage string

const (
    LifecycleStable       LifecycleStage = "stable"       // Production ready
    LifecycleBeta         LifecycleStage = "beta"         // Feature complete, testing
    LifecycleExperimental LifecycleStage = "experimental" // Early development
    LifecycleDeprecated   LifecycleStage = "deprecated"   // Will be removed
)
```

#### 1.2 Enhanced Registry with Queries

```go
// cmd/registry/registry.go
type Registry struct {
    providers map[string]CommandProvider
    mu        sync.RWMutex
}

func NewRegistry() *Registry {
    return &Registry{
        providers: make(map[string]CommandProvider),
    }
}

// Register adds a command provider
func (r *Registry) Register(p CommandProvider) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    meta := p.Metadata()
    if _, exists := r.providers[meta.Name]; exists {
        return fmt.Errorf("command already registered: %s", meta.Name)
    }

    r.providers[meta.Name] = p
    return nil
}

// Query methods
func (r *Registry) ByCategory(cat CommandCategory) []CommandProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var result []CommandProvider
    for _, p := range r.providers {
        if p.Metadata().Category == cat {
            result = append(result, p)
        }
    }
    return result
}

func (r *Registry) StableCommands() []CommandProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var result []CommandProvider
    for _, p := range r.providers {
        if p.Metadata().Lifecycle == LifecycleStable {
            result = append(result, p)
        }
    }
    return result
}

func (r *Registry) ExperimentalCommands() []CommandProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var result []CommandProvider
    for _, p := range r.providers {
        if p.Metadata().Experimental {
            result = append(result, p)
        }
    }
    return result
}

// Sorted by priority
func (r *Registry) AllByPriority() []CommandProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]CommandProvider, 0, len(r.providers))
    for _, p := range r.providers {
        result = append(result, p)
    }

    sort.Slice(result, func(i, j int) bool {
        return result[i].Metadata().Priority < result[j].Metadata().Priority
    })

    return result
}
```

#### 1.3 Command Implementation Example

```go
// cmd/git/register.go
type gitCmdProvider struct {
    appCtx *app.AppContext
}

func (p gitCmdProvider) Command() *cobra.Command {
    return git.NewGitCmd(context.Background(), p.appCtx)
}

func (p gitCmdProvider) Metadata() registry.CommandMetadata {
    return registry.CommandMetadata{
        Name:         "git",
        Category:     registry.CategoryGit,
        Version:      "1.0.0",
        Priority:     10,
        Experimental: false,
        Dependencies: []string{"git"},
        Tags:         []string{"git", "repository", "vcs"},
        Lifecycle:    registry.LifecycleStable,
    }
}

func RegisterGitCmd(appCtx *app.AppContext) {
    registry.Register(gitCmdProvider{appCtx: appCtx})
}
```

#### 1.4 Root Command Integration

```go
// cmd/root.go
func NewRootCmd(ctx context.Context, version string, appCtx *app.AppContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "gz",
        Short: "개발 환경 및 Git 플랫폼 통합 관리 도구",
    }

    // Register all commands
    RegisterPMCmd(appCtx)
    RegisterQualityCmd(appCtx)
    git.RegisterGitCmd(appCtx)
    // ... other commands

    // Add commands by category for organized help
    addCommandsByCategory(cmd, registry.CategoryGit)
    addCommandsByCategory(cmd, registry.CategoryDevelopment)
    addCommandsByCategory(cmd, registry.CategoryQuality)

    // Add experimental commands only if flag enabled
    if os.Getenv("GZ_EXPERIMENTAL") == "1" {
        addExperimentalCommands(cmd)
    }

    return cmd
}

func addCommandsByCategory(parent *cobra.Command, cat registry.CommandCategory) {
    providers := registry.GlobalRegistry.ByCategory(cat)
    for _, p := range providers {
        parent.AddCommand(p.Command())
    }
}

func addExperimentalCommands(parent *cobra.Command) {
    providers := registry.GlobalRegistry.ExperimentalCommands()
    for _, p := range providers {
        cmd := p.Command()
        cmd.Short = "[EXPERIMENTAL] " + cmd.Short
        parent.AddCommand(cmd)
    }
}
```

### Benefits

- ✅ **Better organization**: Commands grouped by category
- ✅ **Lifecycle management**: Clear stable/beta/experimental markers
- ✅ **Dependency tracking**: Know what external tools are needed
- ✅ **Enhanced help**: Category-based help system
- ✅ **Introspection**: Auto-generate docs from metadata
- ✅ **Backward compatible**: Existing code works with simple wrapper

### Impact

- **Code changes**: Minimal (add Metadata() method to providers)
- **Binary size**: +5KB (metadata storage)
- **Performance**: Negligible
- **Compatibility**: 100% backward compatible

## 🏗️ Improvement Strategy 2: Build Variants

### Goal

Create multiple binary variants (full, core, custom) without plugin system.

### Implementation

#### 2.1 Build Tag System

```go
// cmd/quality_wrapper.go
//go:build quality || full
// +build quality full

package cmd

// Quality command only included with -tags quality or -tags full
```

```go
// cmd/synclone/register.go
//go:build synclone || full
// +build synclone full

package synclone
```

```go
// cmd/git/register.go
//go:build core || full
// +build core full

package git

// Git commands always included in core and full builds
```

#### 2.2 Conditional Registration

```go
// cmd/root.go
func NewRootCmd(ctx context.Context, version string, appCtx *app.AppContext) *cobra.Command {
    cmd := &cobra.Command{
        Use: "gz",
    }

    // Core commands (always included)
    registerCoreCommands(appCtx)

    // Optional commands (build tag dependent)
    registerOptionalCommands(appCtx)

    // Add all registered commands
    for _, provider := range registry.List() {
        cmd.AddCommand(provider.Command())
    }

    return cmd
}

func registerCoreCommands(appCtx *app.AppContext) {
    git.RegisterGitCmd(appCtx)
    devenv.RegisterDevEnvCmd(appCtx)
    versioncmd.RegisterVersionCmd()
    // Always available
}

//go:build quality || full
func registerOptionalCommands(appCtx *app.AppContext) {
    RegisterQualityCmd(appCtx)
    RegisterPMCmd(appCtx)
    synclone.RegisterSyncCloneCmd(appCtx)
    // Only with build tags
}

//go:build !quality && !full
func registerOptionalCommands(appCtx *app.AppContext) {
    // No-op for core build
}
```

#### 2.3 Build Scripts

```bash
# scripts/build-variants.sh
#!/bin/bash
set -euo pipefail

VERSION=${VERSION:-dev}
OUTPUT_DIR=${OUTPUT_DIR:-dist}

mkdir -p "$OUTPUT_DIR"

echo "Building gz variants..."

# Full build (all features)
echo "  - gz-full (all commands)"
go build -tags full \
    -ldflags "-X main.version=$VERSION" \
    -o "$OUTPUT_DIR/gz-full" \
    ./cmd/gz

# Core build (essential commands only)
echo "  - gz-core (git + dev-env only)"
go build -tags core \
    -ldflags "-X main.version=$VERSION" \
    -o "$OUTPUT_DIR/gz-core" \
    ./cmd/gz

# Git-focused build
echo "  - gz-git (git commands + quality)"
go build -tags "core,quality" \
    -ldflags "-X main.version=$VERSION" \
    -o "$OUTPUT_DIR/gz-git" \
    ./cmd/gz

# Developer build
echo "  - gz-dev (dev-env focused)"
go build -tags "core,quality,pm" \
    -ldflags "-X main.version=$VERSION" \
    -o "$OUTPUT_DIR/gz-dev" \
    ./cmd/gz

echo ""
echo "Build summary:"
ls -lh "$OUTPUT_DIR"/gz-* | awk '{print "  " $9 ": " $5}'
```

#### 2.4 Makefile Integration

```makefile
# .make/build.mk

# Build all variants
.PHONY: build-variants
build-variants: ## Build all binary variants (full, core, git, dev)
	@bash scripts/build-variants.sh

# Individual variant targets
.PHONY: build-full
build-full: ## Build full binary with all commands
	@go build -tags full -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT_DIR)/gz-full ./cmd/gz

.PHONY: build-core
build-core: ## Build core binary (git + dev-env only)
	@go build -tags core -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT_DIR)/gz-core ./cmd/gz

.PHONY: build-git
build-git: ## Build git-focused binary
	@go build -tags "core,quality" -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT_DIR)/gz-git ./cmd/gz

.PHONY: build-dev
build-dev: ## Build developer-focused binary
	@go build -tags "core,quality,pm" -ldflags "-X main.version=$(VERSION)" -o $(OUTPUT_DIR)/gz-dev ./cmd/gz

# Default build is full
.PHONY: build
build: build-full ## Build full binary (default)
	@cp $(OUTPUT_DIR)/gz-full gz
```

### Build Variant Comparison

| Variant     | Commands Included        | Size  | Use Case                |
| ----------- | ------------------------ | ----- | ----------------------- |
| **gz-full** | All (11 commands)        | ~33MB | Default, all features   |
| **gz-core** | git, dev-env, version    | ~8MB  | Minimal, essential only |
| **gz-git**  | core + quality, synclone | ~15MB | Git-focused workflows   |
| **gz-dev**  | core + quality, pm, ide  | ~20MB | Development environment |

### Benefits

- ✅ **Smaller binaries**: Core variant ~75% smaller
- ✅ **Faster distribution**: Download only needed features
- ✅ **Specialized builds**: Git-only, dev-only variants
- ✅ **Enterprise customization**: Custom feature combinations
- ✅ **No runtime complexity**: All compile-time decisions

### Impact

- **Code changes**: Moderate (add build tags to each command)
- **Build complexity**: +4 build targets
- **Maintenance**: Need to test multiple variants
- **Distribution**: Multiple binaries to distribute

## 🔄 Improvement Strategy 3: Command Lifecycle Management

### Goal

Manage command deprecation, experimental features, and version compatibility.

### Implementation

#### 3.1 Lifecycle Annotations

```go
// cmd/registry/lifecycle.go
type LifecycleManager struct {
    registry *Registry
}

func (lm *LifecycleManager) ValidateLifecycle() error {
    for _, p := range lm.registry.All() {
        meta := p.Metadata()

        // Warn about deprecated commands
        if meta.Lifecycle == LifecycleDeprecated {
            fmt.Fprintf(os.Stderr, "⚠️  Command '%s' is deprecated and will be removed\n", meta.Name)
        }

        // Check dependencies
        if err := lm.checkDependencies(meta); err != nil {
            return fmt.Errorf("command %s: %w", meta.Name, err)
        }
    }
    return nil
}

func (lm *LifecycleManager) checkDependencies(meta CommandMetadata) error {
    for _, dep := range meta.Dependencies {
        if _, err := exec.LookPath(dep); err != nil {
            return fmt.Errorf("missing dependency: %s", dep)
        }
    }
    return nil
}
```

#### 3.2 Experimental Flag System

```go
// cmd/root.go
func NewRootCmd(ctx context.Context, version string, appCtx *app.AppContext) *cobra.Command {
    cmd := &cobra.Command{
        Use: "gz",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            // Check if running experimental command
            if isExperimental(cmd) && !experimentalEnabled() {
                fmt.Fprintf(os.Stderr, "⚠️  This is an experimental command.\n")
                fmt.Fprintf(os.Stderr, "Enable with: export GZ_EXPERIMENTAL=1\n")
                os.Exit(1)
            }
        },
    }

    // Register commands
    registerAllCommands(appCtx)

    // Add lifecycle warnings to help
    enhanceHelpWithLifecycle(cmd)

    return cmd
}

func isExperimental(cmd *cobra.Command) bool {
    // Check command metadata
    for _, p := range registry.GlobalRegistry.All() {
        if p.Command().Name() == cmd.Name() {
            return p.Metadata().Experimental
        }
    }
    return false
}

func experimentalEnabled() bool {
    return os.Getenv("GZ_EXPERIMENTAL") == "1"
}

func enhanceHelpWithLifecycle(cmd *cobra.Command) {
    originalHelp := cmd.HelpFunc()
    cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
        // Show original help
        originalHelp(c, args)

        // Add lifecycle information
        fmt.Fprintf(os.Stderr, "\nCommand Status:\n")
        for _, p := range registry.GlobalRegistry.All() {
            meta := p.Metadata()
            status := lifecycleSymbol(meta.Lifecycle)
            fmt.Fprintf(os.Stderr, "  %s %s: %s\n", status, meta.Name, meta.Lifecycle)
        }
    })
}

func lifecycleSymbol(stage LifecycleStage) string {
    switch stage {
    case LifecycleStable:
        return "✅"
    case LifecycleBeta:
        return "🔶"
    case LifecycleExperimental:
        return "🧪"
    case LifecycleDeprecated:
        return "⚠️"
    default:
        return "❓"
    }
}
```

#### 3.3 Deprecation Warnings

```go
// cmd/deprecated/wrapper.go
func WrapDeprecatedCommand(cmd *cobra.Command, replacement string, removeVersion string) *cobra.Command {
    originalRun := cmd.Run
    originalRunE := cmd.RunE

    warningShown := false
    showWarning := func() {
        if !warningShown {
            fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: This command is deprecated\n")
            fmt.Fprintf(os.Stderr, "    Use '%s' instead\n", replacement)
            fmt.Fprintf(os.Stderr, "    Will be removed in version %s\n\n", removeVersion)
            warningShown = true
        }
    }

    if originalRun != nil {
        cmd.Run = func(c *cobra.Command, args []string) {
            showWarning()
            originalRun(c, args)
        }
    }

    if originalRunE != nil {
        cmd.RunE = func(c *cobra.Command, args []string) error {
            showWarning()
            return originalRunE(c, args)
        }
    }

    return cmd
}
```

### Benefits

- ✅ **Clear communication**: Users know what's stable/experimental
- ✅ **Safe deprecation**: Warn before breaking changes
- ✅ **Dependency checking**: Validate required tools exist
- ✅ **Better help**: Lifecycle status in help output

## ⚙️ Improvement Strategy 4: Configuration-Based Extensions

### Goal

Allow users to extend `gz` with custom aliases and shortcuts without modifying code.

### Implementation

#### 4.1 Configuration Schema

```yaml
# ~/.config/gzh-manager/extensions.yaml
aliases:
  # Simple command aliases
  pull-all: "git repo pull-all"
  update: "pm update --all"
  full-sync: "synclone run && pm update --all"

  # Multi-step workflows
  setup-project:
    description: "Complete project setup"
    steps:
      - "git clone {repo}"
      - "cd {repo}"
      - "pm bootstrap"
      - "dev-env setup"

  # Parameterized aliases
  quick-clone:
    description: "Clone and setup repository"
    params:
      - name: repo
        description: "Repository URL"
        required: true
    command: "git repo clone-or-update {repo} && cd $(basename {repo} .git)"

external:
  # Integrate external commands
  - name: terraform
    command: /usr/local/bin/terraform
    description: "Terraform infrastructure management"
    passthrough: true  # Pass all args directly

  - name: custom-lint
    command: /opt/tools/custom-lint
    description: "Custom linting tool"
    args:
      - "--config"
      - "$HOME/.config/lint.yaml"

hooks:
  # Pre/post command hooks
  pre-commit:
    - "gz quality check"
  post-clone:
    - "gz pm bootstrap"
```

#### 4.2 Extension Loader

```go
// cmd/extensions/loader.go
type ExtensionConfig struct {
    Aliases  map[string]AliasConfig    `yaml:"aliases"`
    External []ExternalCommandConfig   `yaml:"external"`
    Hooks    map[string][]string       `yaml:"hooks"`
}

type AliasConfig struct {
    Command     string              `yaml:"command,omitempty"`
    Description string              `yaml:"description"`
    Steps       []string            `yaml:"steps,omitempty"`
    Params      []ParamConfig       `yaml:"params,omitempty"`
}

type ExternalCommandConfig struct {
    Name        string   `yaml:"name"`
    Command     string   `yaml:"command"`
    Description string   `yaml:"description"`
    Passthrough bool     `yaml:"passthrough"`
    Args        []string `yaml:"args"`
}

type ParamConfig struct {
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
    Required    bool   `yaml:"required"`
}

func LoadExtensions() (*ExtensionConfig, error) {
    configPath := filepath.Join(os.Getenv("HOME"), ".config", "gzh-manager", "extensions.yaml")
    data, err := os.ReadFile(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            return &ExtensionConfig{}, nil // No extensions configured
        }
        return nil, err
    }

    var cfg ExtensionConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse extensions config: %w", err)
    }

    return &cfg, nil
}

func (cfg *ExtensionConfig) RegisterAll(rootCmd *cobra.Command) error {
    // Register aliases
    for name, alias := range cfg.Aliases {
        if err := registerAlias(rootCmd, name, alias); err != nil {
            return fmt.Errorf("register alias %s: %w", name, err)
        }
    }

    // Register external commands
    for _, ext := range cfg.External {
        if err := registerExternal(rootCmd, ext); err != nil {
            return fmt.Errorf("register external %s: %w", ext.Name, err)
        }
    }

    return nil
}

func registerAlias(parent *cobra.Command, name string, alias AliasConfig) error {
    cmd := &cobra.Command{
        Use:   name,
        Short: alias.Description,
        RunE: func(cmd *cobra.Command, args []string) error {
            if alias.Command != "" {
                // Simple alias
                return executeAlias(alias.Command, args)
            }
            if len(alias.Steps) > 0 {
                // Multi-step workflow
                return executeSteps(alias.Steps, args)
            }
            return fmt.Errorf("invalid alias configuration")
        },
    }

    // Add parameter flags
    for _, param := range alias.Params {
        cmd.Flags().String(param.Name, "", param.Description)
        if param.Required {
            cmd.MarkFlagRequired(param.Name)
        }
    }

    parent.AddCommand(cmd)
    return nil
}

func registerExternal(parent *cobra.Command, ext ExternalCommandConfig) error {
    // Check if external command exists
    if _, err := exec.LookPath(ext.Command); err != nil {
        fmt.Fprintf(os.Stderr, "⚠️  External command not found: %s\n", ext.Command)
        return nil // Don't fail, just skip
    }

    cmd := &cobra.Command{
        Use:   ext.Name,
        Short: fmt.Sprintf("[EXTERNAL] %s", ext.Description),
        RunE: func(cmd *cobra.Command, args []string) error {
            cmdArgs := append(ext.Args, args...)
            execCmd := exec.Command(ext.Command, cmdArgs...)
            execCmd.Stdin = os.Stdin
            execCmd.Stdout = os.Stdout
            execCmd.Stderr = os.Stderr
            return execCmd.Run()
        },
        DisableFlagParsing: ext.Passthrough,
    }

    parent.AddCommand(cmd)
    return nil
}

func executeAlias(aliasCmd string, args []string) error {
    // Parse and execute alias command
    parts := strings.Fields(aliasCmd)
    if len(parts) == 0 {
        return fmt.Errorf("empty alias command")
    }

    // Execute as gz subcommand
    cmd := exec.Command("gz", append(parts, args...)...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func executeSteps(steps []string, args []string) error {
    for i, step := range steps {
        fmt.Fprintf(os.Stderr, "Step %d/%d: %s\n", i+1, len(steps), step)
        if err := executeAlias(step, args); err != nil {
            return fmt.Errorf("step %d failed: %w", i+1, err)
        }
    }
    return nil
}
```

#### 4.3 Root Integration

```go
// cmd/root.go
func NewRootCmd(ctx context.Context, version string, appCtx *app.AppContext) *cobra.Command {
    cmd := &cobra.Command{
        Use: "gz",
    }

    // Register core commands
    registerAllCommands(appCtx)

    // Add registered commands
    for _, provider := range registry.List() {
        cmd.AddCommand(provider.Command())
    }

    // Load user extensions
    if extensions, err := extensions.LoadExtensions(); err == nil {
        if err := extensions.RegisterAll(cmd); err != nil {
            fmt.Fprintf(os.Stderr, "⚠️  Failed to load extensions: %v\n", err)
        }
    }

    return cmd
}
```

### Benefits

- ✅ **User extensibility**: Custom aliases without recompilation
- ✅ **Workflow automation**: Multi-step command sequences
- ✅ **External tool integration**: Wrap any command-line tool
- ✅ **No code changes**: Pure configuration-based
- ✅ **Optional**: Works without extensions.yaml

### Impact

- **Code changes**: Low (add extension loader)
- **Binary size**: +10KB
- **Performance**: Minimal (config loaded once)
- **User experience**: Greatly enhanced

## 📦 Improvement Strategy 5: Wrapper Pattern Enhancement

### Goal

Improve external library integration pattern for better maintainability.

### Current State

4 external libraries already integrated:

- `gzh-cli-git` (repo operations wrapper)
- `gzh-cli-quality` (quality_wrapper.go)
- `gzh-cli-package-manager` (pm_wrapper.go)
- `gzh-cli-shellforge` (shellforge_wrapper.go)

### Implementation

#### 5.1 Standardized Wrapper Interface

```go
// internal/wrapper/interface.go
package wrapper

import (
    "github.com/spf13/cobra"
    "github.com/Gizzahub/gzh-cli/internal/app"
)

// ExternalLibrary defines the interface for wrapping external libraries
type ExternalLibrary interface {
    // Name returns the library name
    Name() string

    // Version returns the library version
    Version() string

    // CreateCommand creates a cobra command from the external library
    CreateCommand(appCtx *app.AppContext) (*cobra.Command, error)

    // Validate checks if the library is properly configured
    Validate() error

    // Dependencies lists required external tools
    Dependencies() []string
}

// BaseWrapper provides common wrapper functionality
type BaseWrapper struct {
    name         string
    version      string
    dependencies []string
}

func (w *BaseWrapper) Name() string         { return w.name }
func (w *BaseWrapper) Version() string      { return w.version }
func (w *BaseWrapper) Dependencies() []string { return w.dependencies }

func (w *BaseWrapper) Validate() error {
    // Check dependencies
    for _, dep := range w.dependencies {
        if _, err := exec.LookPath(dep); err != nil {
            return fmt.Errorf("missing dependency: %s", dep)
        }
    }
    return nil
}
```

#### 5.2 Enhanced Wrapper Implementation

```go
// cmd/pm_wrapper.go (enhanced version)
package cmd

import (
    "context"

    pmcmd "github.com/gizzahub/gzh-cli-package-manager/cmd/pm/command"
    "github.com/Gizzahub/gzh-cli/cmd/registry"
    "github.com/Gizzahub/gzh-cli/internal/app"
    "github.com/Gizzahub/gzh-cli/internal/wrapper"
    "github.com/spf13/cobra"
)

// PMWrapper wraps gzh-cli-package-manager
type PMWrapper struct {
    wrapper.BaseWrapper
}

func NewPMWrapper() *PMWrapper {
    return &PMWrapper{
        BaseWrapper: wrapper.BaseWrapper{
            name:         "pm",
            version:      "1.0.0", // Could get from external library
            dependencies: []string{}, // Package managers optional
        },
    }
}

func (w *PMWrapper) CreateCommand(appCtx *app.AppContext) (*cobra.Command, error) {
    // Use external library implementation
    cmd := pmcmd.NewRootCmd()

    // Customize for gzh-cli context
    cmd.Use = "pm"
    cmd.Short = "Package manager operations"
    cmd.Long = `Manage multiple package managers with unified commands.

This command provides centralized management for multiple package managers including:
- System package managers: brew, apt, port, yum, dnf, pacman
- Version managers: asdf, rbenv, pyenv, nvm, sdkman
- Language package managers: pip, gem, npm, cargo, go, composer

Powered by: gzh-cli-package-manager ` + w.Version() + `

Examples:
  gz pm status      # Show status of all package managers
  gz pm update --all # Update all packages
  gz pm bootstrap   # Bootstrap missing package managers`

    return cmd, nil
}

// Provider implementation
type pmCmdProvider struct {
    appCtx  *app.AppContext
    wrapper *PMWrapper
}

func (p pmCmdProvider) Command() *cobra.Command {
    cmd, err := p.wrapper.CreateCommand(p.appCtx)
    if err != nil {
        // Fallback to minimal command with error
        return &cobra.Command{
            Use:   "pm",
            Short: "Package manager (unavailable)",
            RunE: func(cmd *cobra.Command, args []string) error {
                return fmt.Errorf("pm command unavailable: %w", err)
            },
        }
    }
    return cmd
}

func (p pmCmdProvider) Metadata() registry.CommandMetadata {
    return registry.CommandMetadata{
        Name:         p.wrapper.Name(),
        Category:     registry.CategoryDevelopment,
        Version:      p.wrapper.Version(),
        Priority:     30,
        Experimental: false,
        Dependencies: p.wrapper.Dependencies(),
        Tags:         []string{"package", "manager", "update", "install"},
        Lifecycle:    registry.LifecycleStable,
    }
}

func RegisterPMCmd(appCtx *app.AppContext) {
    wrapper := NewPMWrapper()
    if err := wrapper.Validate(); err != nil {
        fmt.Fprintf(os.Stderr, "⚠️  PM command validation failed: %v\n", err)
        // Still register but it will show error when used
    }

    registry.Register(pmCmdProvider{
        appCtx:  appCtx,
        wrapper: wrapper,
    })
}
```

#### 5.3 Wrapper Documentation Generator

```go
// cmd/wrapper/doc_generator.go
package wrapper

import (
    "fmt"
    "os"
    "text/template"
)

const wrapperDocTemplate = `# {{.Name}} Command Wrapper

**External Library**: {{.LibraryURL}}
**Version**: {{.Version}}
**Status**: {{.Status}}

## Overview

This command is implemented in an external library and integrated via wrapper pattern.

## Dependencies

{{range .Dependencies}}
- {{.}}
{{end}}

## Local Development

To work on the {{.Name}} command:

1. Clone the external library:
   \`\`\`bash
   git clone {{.LibraryURL}} ../{{.LibraryName}}
   \`\`\`

2. Update go.mod with local replace:
   \`\`\`go
   replace {{.LibraryModule}} => ../{{.LibraryName}}
   \`\`\`

3. Make changes in external library

4. Test integration:
   \`\`\`bash
   make build
   ./gz {{.Name}} --help
   \`\`\`

## Wrapper Maintenance

- **Core logic**: Modify in {{.LibraryURL}}
- **CLI integration**: Modify cmd/{{.Name}}_wrapper.go
- **Updates**: Update go.mod dependency version

## Testing

\`\`\`bash
# Test wrapper
go test ./cmd/{{.Name}}_wrapper_test.go

# Test external library (in library repo)
cd ../{{.LibraryName}} && go test ./...
\`\`\`
`

type WrapperDoc struct {
    Name          string
    LibraryURL    string
    LibraryName   string
    LibraryModule string
    Version       string
    Status        string
    Dependencies  []string
}

func GenerateWrapperDoc(lib ExternalLibrary, libURL string) error {
    doc := WrapperDoc{
        Name:          lib.Name(),
        LibraryURL:    libURL,
        LibraryName:   extractLibraryName(libURL),
        LibraryModule: extractModulePath(libURL),
        Version:       lib.Version(),
        Status:        "Active",
        Dependencies:  lib.Dependencies(),
    }

    tmpl, err := template.New("wrapper").Parse(wrapperDocTemplate)
    if err != nil {
        return err
    }

    filename := fmt.Sprintf("docs/integration/%s-wrapper.md", lib.Name())
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    return tmpl.Execute(f, doc)
}
```

### Benefits

- ✅ **Consistent pattern**: All wrappers follow same interface
- ✅ **Better validation**: Check dependencies before use
- ✅ **Auto-documentation**: Generate wrapper docs
- ✅ **Easier maintenance**: Clear separation of concerns
- ✅ **Version tracking**: Know which library version is integrated

## 📊 Implementation Priority

### Phase 1: Quick Wins (1-2 days)

**Impact**: High | **Effort**: Low

1. ✅ **Enhanced Registry Metadata** (Strategy 1)

   - Add `Metadata()` interface
   - Implement in existing commands
   - Add category grouping

1. ✅ **Configuration-Based Aliases** (Strategy 4)

   - Simple alias support
   - External command integration
   - No workflow features yet

**Deliverables**:

- Enhanced help system with categories
- User can add custom aliases
- Foundation for future improvements

### Phase 2: Build Optimization (3-5 days)

**Impact**: Medium | **Effort**: Medium

1. ✅ **Build Variants** (Strategy 2)

   - Add build tags to commands
   - Create build scripts
   - Test all variants

1. ✅ **Wrapper Pattern Enhancement** (Strategy 5)

   - Standardize wrapper interface
   - Update existing wrappers
   - Generate wrapper docs

**Deliverables**:

- Multiple binary sizes (core, full, custom)
- Consistent wrapper pattern
- Better external library integration

### Phase 3: Advanced Features (5-7 days)

**Impact**: Medium | **Effort**: High

1. ✅ **Lifecycle Management** (Strategy 3)

   - Experimental flag system
   - Deprecation warnings
   - Dependency checking

1. ✅ **Advanced Extensions** (Strategy 4)

   - Multi-step workflows
   - Parameterized aliases
   - Hook system

**Deliverables**:

- Complete lifecycle management
- Advanced user extensibility
- Production-ready extension system

## 🚀 Migration Guide

### For Developers

#### Adding Enhanced Metadata (5 minutes per command)

```go
// Before
type gitCmdProvider struct {
    appCtx *app.AppContext
}

func (p gitCmdProvider) Command() *cobra.Command {
    return git.NewGitCmd(context.Background(), p.appCtx)
}

// After - Add Metadata() method
func (p gitCmdProvider) Metadata() registry.CommandMetadata {
    return registry.CommandMetadata{
        Name:         "git",
        Category:     registry.CategoryGit,
        Version:      "1.0.0",
        Priority:     10,
        Experimental: false,
        Dependencies: []string{"git"},
        Tags:         []string{"git", "repository", "vcs"},
        Lifecycle:    registry.LifecycleStable,
    }
}
```

#### Adding Build Tags (2 minutes per command)

```go
// At top of file
//go:build quality || full
// +build quality full

package cmd
```

#### Creating Wrapper (30 minutes per library)

```go
// 1. Create wrapper struct
type MyWrapper struct {
    wrapper.BaseWrapper
}

// 2. Implement CreateCommand
func (w *MyWrapper) CreateCommand(appCtx *app.AppContext) (*cobra.Command, error) {
    cmd := externallib.NewCommand()
    // Customize
    return cmd, nil
}

// 3. Update provider to use wrapper
```

### For Users

#### Using Build Variants

```bash
# Download appropriate variant
wget https://releases/gz-core  # Minimal (~8MB)
wget https://releases/gz-full  # All features (~33MB)

# Or build custom
git clone https://github.com/gizzahub/gzh-cli
cd gzh-cli
make build-core  # or build-git, build-dev
```

#### Creating Aliases

```bash
# Create extensions config
mkdir -p ~/.config/gzh-manager
cat > ~/.config/gzh-manager/extensions.yaml << 'EOF'
aliases:
  update-all: "pm update --all"
  setup: "dev-env bootstrap && pm bootstrap"

external:
  - name: tf
    command: /usr/local/bin/terraform
    description: "Terraform shortcut"
    passthrough: true
EOF

# Use aliases
gz update-all
gz setup
gz tf plan
```

## 📈 Expected Improvements

| Metric                   | Before     | After Phase 1            | After Phase 3            |
| ------------------------ | ---------- | ------------------------ | ------------------------ |
| **Binary Sizes**         | 33MB       | 33MB                     | 8MB (core) - 33MB (full) |
| **User Extensibility**   | None       | Basic aliases            | Full workflow automation |
| **Command Organization** | Flat list  | Categories               | Categories + lifecycle   |
| **External Libraries**   | 4 wrappers | 4 wrappers               | Standardized pattern     |
| **Help Quality**         | Basic      | Enhanced with categories | Full metadata display    |
| **Build Variants**       | 1 (full)   | 1 (full)                 | 4 (core, git, dev, full) |

## ✅ Success Criteria

### Phase 1 Complete When:

- [ ] All commands have metadata
- [ ] Help system shows categories
- [ ] Users can define basic aliases
- [ ] External commands can be integrated

### Phase 2 Complete When:

- [ ] 4 build variants available
- [ ] All wrappers follow standard pattern
- [ ] Binary size reduced for core variant
- [ ] Build scripts automated

### Phase 3 Complete When:

- [ ] Experimental commands protected
- [ ] Deprecated commands show warnings
- [ ] Multi-step workflows work
- [ ] Hook system functional

______________________________________________________________________

**Document Version**: 1.0
**Last Updated**: 2025-12-01
**Implementation Status**: Planning Phase
**Next Review**: After Phase 1 completion
