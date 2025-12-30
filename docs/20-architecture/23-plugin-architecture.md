# 🔌 Plugin Architecture Analysis

This document analyzes the current command registration system and evaluates potential plugin architecture strategies.

## 📋 Table of Contents

- [Current Architecture](#current-architecture)
- [Plugin Architecture Options](#plugin-architecture-options)
- [Trade-offs Analysis](#trade-offs-analysis)
- [Recommendation](#recommendation)

## 🏗️ Current Architecture

### Hybrid Registry Pattern

gzh-cli uses a **registry-based abstraction layer** with **compile-time integration**.

#### Components

```
cmd/registry/registry.go       # CommandProvider interface + registry
cmd/root.go                    # Explicit command registration
cmd/*/register.go              # Individual command registration functions
```

#### Registration Flow

```go
// 1. Interface Definition (cmd/registry/registry.go)
type CommandProvider interface {
    Command() *cobra.Command
}

// 2. Command Registration (cmd/synclone/register.go)
type syncCloneCmdProvider struct {
    appCtx *app.AppContext
}

func (p syncCloneCmdProvider) Command() *cobra.Command {
    return NewSyncCloneCmd(context.Background(), p.appCtx)
}

func RegisterSyncCloneCmd(appCtx *app.AppContext) {
    registry.Register(syncCloneCmdProvider{appCtx: appCtx})
}

// 3. Explicit Registration (cmd/root.go)
synclone.RegisterSyncCloneCmd(appCtx)
devenv.RegisterDevEnvCmd(appCtx)
ide.RegisterIDECmd(appCtx)
// ... 11 total commands

// 4. Automatic Command Addition
for _, provider := range registry.List() {
    cmd.AddCommand(provider.Command())
}
```

### Characteristics

| Aspect            | Current State                        |
| ----------------- | ------------------------------------ |
| **Loading**       | Compile-time only                    |
| **Registration**  | Registry pattern with explicit calls |
| **Dependencies**  | Hardcoded imports in root.go         |
| **Binary Size**   | All commands included                |
| **Extensibility** | Requires recompilation               |
| **Distribution**  | Single monolithic binary             |

### External Library Integration Example

**pm_wrapper.go** demonstrates external package integration:

```go
// Import external library
pmcmd "github.com/gizzahub/gzh-cli-package-manager/cmd/pm/command"

// Wrap and customize
func NewPMCmd(ctx context.Context, appCtx *app.AppContext) *cobra.Command {
    cmd := pmcmd.NewRootCmd()           // Use external implementation
    cmd.Use = "pm"                       // Customize metadata
    cmd.Short = "Package manager ops"    // Customize description
    return cmd
}
```

This pattern allows:

- ✅ Library reuse across projects
- ✅ Separate versioning and maintenance
- ✅ Consistent functionality
- ❌ Still compiled into main binary

## 🔌 Plugin Architecture Options

### Option 1: Go Plugin System (plugin.so)

Build commands as shared libraries loaded at runtime.

#### Implementation

```go
// Plugin implementation (plugins/quality/quality_plugin.go)
package main

import "github.com/spf13/cobra"

type QualityPlugin struct{}

func (p *QualityPlugin) Command() *cobra.Command {
    return &cobra.Command{
        Use:   "quality",
        Short: "Code quality management",
        RunE:  func(cmd *cobra.Command, args []string) error {
            // Implementation
        },
    }
}

var Plugin QualityPlugin  // Exported symbol for plugin system

// Root command loader (cmd/root.go)
func loadPlugins(dir string) error {
    files, _ := filepath.Glob(filepath.Join(dir, "*.so"))
    for _, file := range files {
        p, err := plugin.Open(file)
        if err != nil {
            return err
        }

        sym, err := p.Lookup("Plugin")
        if err != nil {
            return err
        }

        provider, ok := sym.(registry.CommandProvider)
        if !ok {
            return fmt.Errorf("invalid plugin")
        }

        registry.Register(provider)
    }
    return nil
}
```

#### Build Process

```bash
# Build plugin
go build -buildmode=plugin -o plugins/quality.so plugins/quality/quality_plugin.go

# Build main binary (no quality command compiled in)
go build -o gz cmd/main.go

# Runtime loading
./gz --plugin-dir ./plugins quality check
```

### Option 2: Subprocess-Based Plugins (Git-Style)

Execute external binaries as subcommands.

#### Implementation

```go
// Command discovery (cmd/root.go)
func discoverPlugins() error {
    // Search PATH for gz-* executables
    paths := strings.Split(os.Getenv("PATH"), ":")
    for _, path := range paths {
        files, _ := filepath.Glob(filepath.Join(path, "gz-*"))
        for _, file := range files {
            name := strings.TrimPrefix(filepath.Base(file), "gz-")
            cmd := &cobra.Command{
                Use:   name,
                Short: fmt.Sprintf("Plugin command: %s", name),
                RunE: func(cmd *cobra.Command, args []string) error {
                    return executePlugin(file, args)
                },
            }
            rootCmd.AddCommand(cmd)
        }
    }
    return nil
}

// Plugin execution
func executePlugin(path string, args []string) error {
    cmd := exec.Command(path, args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

#### Plugin Structure

```bash
# Main binary
/usr/local/bin/gz

# Plugins (separate binaries)
/usr/local/bin/gz-quality
/usr/local/bin/gz-synclone
/usr/local/bin/gz-ide

# Usage
gz quality check       # Executes: gz-quality check
gz synclone run        # Executes: gz-synclone run
```

### Option 3: Configuration-Based Plugins

Define external commands in configuration files.

#### Implementation

```yaml
# ~/.config/gzh-manager/plugins.yaml
plugins:
  - name: quality
    command: /usr/local/bin/gz-quality
    description: Code quality management

  - name: custom-lint
    command: /opt/tools/custom-lint
    description: Custom linting tool
```

```go
// Plugin loading (cmd/root.go)
type PluginConfig struct {
    Name        string `yaml:"name"`
    Command     string `yaml:"command"`
    Description string `yaml:"description"`
}

func loadConfigPlugins() error {
    var cfg struct {
        Plugins []PluginConfig `yaml:"plugins"`
    }

    data, _ := os.ReadFile(pluginConfigPath)
    yaml.Unmarshal(data, &cfg)

    for _, p := range cfg.Plugins {
        cmd := &cobra.Command{
            Use:   p.Name,
            Short: p.Description,
            RunE: func(cmd *cobra.Command, args []string) error {
                return exec.Command(p.Command, args...).Run()
            },
        }
        rootCmd.AddCommand(cmd)
    }
    return nil
}
```

### Option 4: Library-Based Composition (Current Enhanced)

Keep current registry pattern, enhance with optional external libraries.

#### Implementation

```go
// Optional plugin loading (cmd/root.go)
func loadOptionalCommands(appCtx *app.AppContext) {
    // Try loading optional libraries at compile-time
    if hasQualitySupport() {
        RegisterQualityCmd(appCtx)
    }
    if hasSyncCloneSupport() {
        synclone.RegisterSyncCloneCmd(appCtx)
    }
    // Core commands always loaded
    git.RegisterGitCmd(appCtx)
    devenv.RegisterDevEnvCmd(appCtx)
}

// Build tags for conditional compilation
// go build -tags "quality,synclone" -o gz
```

**Build variations**:

```bash
# Full binary (all commands)
go build -o gz-full cmd/main.go

# Core only (git + devenv)
go build -tags core -o gz-core cmd/main.go

# Custom build
go build -tags "core,quality" -o gz-custom cmd/main.go
```

## ⚖️ Trade-offs Analysis

### Comparison Matrix

| Feature             | Current | Go Plugin | Subprocess | Config-Based | Library Comp |
| ------------------- | ------- | --------- | ---------- | ------------ | ------------ |
| **Dynamic Loading** | ❌      | ✅        | ✅         | ✅           | ❌           |
| **Type Safety**     | ✅✅    | ⚠️        | ❌         | ❌           | ✅✅         |
| **Performance**     | ✅✅    | ✅        | ⚠️         | ⚠️           | ✅✅         |
| **Binary Size**     | ⚠️      | ✅        | ✅✅       | ✅           | ⚠️           |
| **Cross-Platform**  | ✅✅    | ❌        | ✅✅       | ✅✅         | ✅✅         |
| **Maintenance**     | ✅      | ⚠️        | ✅         | ✅           | ✅           |
| **Distribution**    | ✅      | ⚠️        | ⚠️         | ⚠️           | ✅           |
| **User Install**    | ✅✅    | ⚠️        | ⚠️         | ⚠️           | ✅✅         |

### Detailed Analysis

#### Current Hybrid Registry

**Pros:**

- ✅ Simple single-binary distribution
- ✅ Full type safety and compile-time checks
- ✅ Excellent performance (no IPC overhead)
- ✅ Works on all platforms (Linux, macOS, Windows)
- ✅ Easy for users (just download and run)
- ✅ Registry abstraction allows easy command addition

**Cons:**

- ❌ All commands compiled into binary (larger size)
- ❌ Cannot extend without recompilation
- ❌ No third-party plugins
- ❌ Unused commands consume memory

**Best For:**

- Projects with stable command set
- Simple distribution requirements
- Teams prioritizing simplicity over extensibility

#### Go Plugin System

**Pros:**

- ✅ True dynamic loading
- ✅ Shared Go types and interfaces
- ✅ No subprocess overhead
- ✅ Smaller main binary

**Cons:**

- ❌ **Linux-only** (not supported on macOS/Windows)
- ❌ Version compatibility fragile (exact Go version + flags)
- ❌ Complex distribution (multiple .so files)
- ❌ Debugging difficulties
- ❌ Plugin updates require exact version matching

**Best For:**

- Linux-only deployments
- Advanced users comfortable with technical complexity
- Enterprise environments with controlled upgrade cycles

#### Subprocess-Based (Git-Style)

**Pros:**

- ✅ Language-agnostic plugins
- ✅ Complete isolation (safe)
- ✅ Independent versioning
- ✅ Cross-platform support
- ✅ Simple plugin development

**Cons:**

- ❌ Subprocess overhead (slower)
- ❌ No type safety
- ❌ Complex shared state management
- ❌ Distribution of multiple binaries
- ❌ PATH management required

**Best For:**

- Polyglot plugin ecosystems
- Projects with many optional commands
- Community-driven plugin development

#### Configuration-Based

**Pros:**

- ✅ User-defined custom commands
- ✅ Simple configuration
- ✅ Cross-platform
- ✅ No code changes needed

**Cons:**

- ❌ External binary dependencies
- ❌ No type safety or validation
- ❌ Limited integration capabilities
- ❌ Debugging challenges
- ❌ User responsibility for plugin availability

**Best For:**

- User customization and workflow automation
- Integration with existing tools
- Simple command aliasing

#### Library-Based Composition

**Pros:**

- ✅ Best of current + build flexibility
- ✅ Full type safety
- ✅ Conditional compilation reduces binary size
- ✅ Single-binary distribution options
- ✅ External library reuse (like pm_wrapper.go)

**Cons:**

- ❌ Still requires recompilation for changes
- ❌ Build tag complexity
- ⚠️ Multiple binary variants to maintain

**Best For:**

- Providing "lite" vs "full" versions
- Enterprise custom builds
- Library reuse across projects

## 🎯 Recommendation

### For gzh-cli: **Keep Current Architecture**

**Rationale:**

1. **Distribution Simplicity**

   - Single binary is easiest for users
   - No plugin installation/management complexity
   - Cross-platform consistency

1. **Performance**

   - Zero IPC overhead
   - Fast startup time
   - Efficient memory usage

1. **Type Safety**

   - Compile-time checking prevents runtime errors
   - Refactoring support with IDE tools
   - Clear dependency management

1. **Maintenance**

   - Simpler testing (no plugin compatibility matrix)
   - Clearer debugging
   - Unified versioning

### When to Consider Plugins

Consider switching to plugin architecture if:

1. **Binary size becomes prohibitive** (> 100MB)

   - Current size: ~20-30MB is acceptable
   - Mitigation: Build tags for "lite" versions

1. **Third-party extension ecosystem needed**

   - If community wants to add custom commands
   - Solution: Subprocess-based plugins

1. **Frequent command additions/removals**

   - If command set changes weekly
   - Solution: Configuration-based plugin loader

1. **Platform-specific commands**

   - If many Linux-only or macOS-only commands
   - Solution: Build tags + conditional compilation

### Enhancement Options (Without Full Plugin System)

#### 1. Improve Registry Discoverability

```go
// cmd/registry/registry.go - Add metadata
type CommandProvider interface {
    Command() *cobra.Command
    Metadata() CommandMetadata
}

type CommandMetadata struct {
    Name        string
    Category    string   // "git", "dev-env", "quality"
    Priority    int      // Display order
    Experimental bool    // Mark experimental features
}

// Auto-generate command documentation from registry
func GenerateCommandDocs() map[string]CommandMetadata {
    docs := make(map[string]CommandMetadata)
    for _, provider := range registry.List() {
        docs[provider.Metadata().Name] = provider.Metadata()
    }
    return docs
}
```

#### 2. Build Tag System for Optional Commands

```go
// cmd/root.go
func RegisterOptionalCommands(appCtx *app.AppContext) {
    // Always register core commands
    git.RegisterGitCmd(appCtx)
    devenv.RegisterDevEnvCmd(appCtx)

    // Optional commands with build tags
    registerOptionalQuality(appCtx)    // +build quality
    registerOptionalSynclone(appCtx)   // +build synclone
}
```

**Build variations**:

```bash
# Full build (default)
make build

# Core only (small binary ~5MB)
make build-core

# Custom build
go build -tags "core,git,devenv" -o gz-custom
```

#### 3. Configuration-Based Command Aliases

```yaml
# ~/.config/gzh-manager/aliases.yaml
aliases:
  update-all: "synclone run && pm update --all"
  full-setup: "dev-env bootstrap && pm bootstrap"

external-commands:
  custom-lint: /usr/local/bin/custom-lint
```

```go
// Load user-defined aliases as commands
func loadAliases() {
    // Simple configuration-based extension point
    // No type safety needed for user shortcuts
}
```

#### 4. External Library Integration Pattern

Enhance the existing `pm_wrapper.go` pattern:

```go
// Create cmd/*_wrapper.go for external integrations
// Example: cmd/terraform_wrapper.go

import tfcmd "github.com/hashicorp/terraform/command"

func NewTerraformCmd(appCtx *app.AppContext) *cobra.Command {
    // Wrap external tool with gzh-cli conventions
    cmd := adaptTerraformCommand(tfcmd.New())
    cmd.Use = "terraform"
    return cmd
}
```

**Benefits:**

- ✅ Reuse well-maintained external libraries
- ✅ Keep single binary distribution
- ✅ Maintain type safety
- ✅ Leverage external project updates

## 📊 Decision Matrix

| Use Case                     | Solution                                 |
| ---------------------------- | ---------------------------------------- |
| **Add new internal command** | Current registry pattern                 |
| **Integrate external tool**  | Library wrapper (pm_wrapper.go style)    |
| **Reduce binary size**       | Build tags + conditional compilation     |
| **User customization**       | Configuration-based aliases              |
| **Community plugins**        | **NOT RECOMMENDED** - complexity > value |

## 📝 Implementation Guide

### Current: Adding New Command

```bash
# 1. Create command package
mkdir cmd/newfeature

# 2. Implement command with registration
cat > cmd/newfeature/register.go << 'EOF'
package newfeature

import (
    "github.com/spf13/cobra"
    "github.com/gizzahub/gzh-cli/cmd/registry"
    "github.com/gizzahub/gzh-cli/internal/app"
)

type newFeatureCmdProvider struct {
    appCtx *app.AppContext
}

func (p newFeatureCmdProvider) Command() *cobra.Command {
    return NewNewFeatureCmd(p.appCtx)
}

func RegisterNewFeatureCmd(appCtx *app.AppContext) {
    registry.Register(newFeatureCmdProvider{appCtx: appCtx})
}
EOF

# 3. Add to root.go
# In cmd/root.go imports:
#   "github.com/gizzahub/gzh-cli/cmd/newfeature"
# In NewRootCmd():
#   newfeature.RegisterNewFeatureCmd(appCtx)

# 4. Build and test
make build
./gz newfeature --help
```

**Steps Required:** 3 (create, implement, register)
**Complexity:** Low
**Time:** 15 minutes

### Future: Build Tag Optimization

If binary size becomes an issue:

```bash
# 1. Add build tag to optional commands
// +build quality

package quality

# 2. Create build script
cat > scripts/build-variants.sh << 'EOF'
#!/bin/bash
# Full build
go build -o dist/gz-full

# Core build (no optional commands)
go build -tags core -o dist/gz-core

# Custom builds
go build -tags "core,git,synclone" -o dist/gz-git
EOF

# 3. Update Makefile
build-variants:
    @echo "Building variants..."
    bash scripts/build-variants.sh
```

## 🔮 Future Considerations

### Monitoring Metrics

Track these metrics to inform future plugin decisions:

1. **Binary Size Growth**

   - Current: ~20-30MB
   - Alert threshold: > 50MB
   - Critical threshold: > 100MB

1. **Command Usage Distribution**

   - Identify rarely-used commands
   - Candidates for optional builds
   - Usage telemetry: `gz doctor report`

1. **Community Requests**

   - Third-party integration requests
   - Custom command proposals
   - Extension API requests

1. **Distribution Complexity**

   - User feedback on binary size
   - Enterprise custom build requests
   - Platform-specific feature needs

### Decision Triggers

Reconsider plugin architecture if:

- Binary size exceeds 100MB
- 50%+ of commands are platform-specific
- Active community plugin development emerges
- Enterprise requires custom command isolation

______________________________________________________________________

**Document Version:** 1.0
**Last Updated:** 2025-12-01
**Status:** Current architecture recommended
**Review Trigger:** Binary size > 50MB or community plugin requests
