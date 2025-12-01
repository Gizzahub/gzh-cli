# Command Lifecycle Management

gz implements a comprehensive lifecycle management system for commands, allowing safe rollout of new features and graceful deprecation of old ones.

## Overview

Commands progress through well-defined lifecycle stages:

1. **Experimental** - Early development, requires opt-in
2. **Beta** - Feature-complete, testing phase
3. **Stable** - Production-ready
4. **Deprecated** - Will be removed in future versions

## Lifecycle Stages

### Experimental

**Characteristics**:
- Early development stage
- API may change significantly
- Disabled by default
- Requires explicit enablement

**Behavior**:
- Hidden from help unless enabled
- Shows warning when executed
- May have incomplete features
- Breaking changes possible

**Enablement**:
```bash
# Via environment variable
export GZ_EXPERIMENTAL=1
gz experimental-command

# Via flag
gz --experimental experimental-command
```

**Warning Message**:
```
⚠️  Warning: Command 'experimental-command' is experimental and may change or be removed
   Version: 0.1.0 | Status: experimental
```

### Beta

**Characteristics**:
- Feature-complete
- API relatively stable
- Testing phase
- Available by default

**Behavior**:
- Shown in help
- Shows info message when executed
- Minor changes possible
- User feedback welcome

**Info Message**:
```
ℹ️  Info: Command 'beta-command' is in beta testing
   Version: 0.9.0 | Please report any issues
```

### Stable

**Characteristics**:
- Production-ready
- API locked (breaking changes require deprecation)
- Fully tested
- Default stage

**Behavior**:
- No special warnings
- Full support
- Backward compatibility guaranteed
- Documentation complete

### Deprecated

**Characteristics**:
- Scheduled for removal
- Replacement available
- Still functional
- Discouraged use

**Behavior**:
- Shown in help with deprecation notice
- Shows warning when executed
- Still works normally
- Migration guide provided

**Warning Message**:
```
⚠️  DEPRECATED: Command 'old-command' is deprecated and will be removed in a future version
   Current Version: 1.0.0 | Please migrate to alternatives
```

## Command Metadata

Commands expose metadata including lifecycle stage:

```go
func (p *cmdProvider) Metadata() registry.CommandMetadata {
    return registry.CommandMetadata{
        Name:         "my-command",
        Category:     registry.CategoryUtility,
        Version:      "1.0.0",
        Lifecycle:    registry.LifecycleStable,
        Dependencies: []string{"git"},
        Tags:         []string{"utility", "tool"},
    }
}
```

## Lifecycle Transitions

### Experimental → Beta

**Requirements**:
- Core functionality complete
- API surface defined
- Basic tests passing
- Documentation written

**Process**:
1. Update Lifecycle field to `LifecycleBeta`
2. Update version (e.g., 0.1.0 → 0.9.0)
3. Document any API changes
4. Announce in changelog

### Beta → Stable

**Requirements**:
- All features implemented
- Comprehensive test coverage
- API frozen
- User testing complete
- Documentation complete

**Process**:
1. Update Lifecycle field to `LifecycleStable`
2. Update version to 1.0.0
3. Announce stability in release notes
4. Add migration guide if needed

### Stable → Deprecated

**Requirements**:
- Replacement available
- Migration path documented
- Deprecation timeline announced
- Major version bump

**Process**:
1. Update Lifecycle field to `LifecycleDeprecated`
2. Add deprecation notice to documentation
3. Update help text with migration info
4. Set removal timeline (usually 2-3 versions)

### Deprecated → Removed

**Requirements**:
- Deprecation period complete (2-3 versions)
- Users notified multiple times
- Alternative well-established
- Breaking change announced

**Process**:
1. Remove command registration
2. Update documentation
3. Add removal note to changelog
4. Major version bump

## Dependency Validation

Commands can declare dependencies on external tools:

```go
Metadata: CommandMetadata{
    Dependencies: []string{"git", "docker", "terraform"},
}
```

### Validation Behavior

At command execution:
1. Check all dependencies available
2. Show warning for missing dependencies
3. Allow execution (command may handle gracefully)

### Warning Message

```
⚠️  Warning: Command 'docker-cmd' requires missing dependencies:
   - docker
```

## Best Practices

### For Developers

**Introducing New Features**:
1. Start as Experimental
2. Gather user feedback
3. Stabilize API in Beta
4. Release as Stable

**Deprecating Features**:
1. Announce deprecation in advance
2. Provide migration guide
3. Keep working during deprecation period
4. Remove only after sufficient notice

**Version Numbering**:
- Experimental: 0.1.x - 0.8.x
- Beta: 0.9.x
- Stable: 1.x.x+
- Deprecated: Mark in help text, not version

### For Users

**Using Experimental Features**:
- Expect breaking changes
- Report bugs and feedback
- Don't use in production
- Stay updated with changes

**Using Beta Features**:
- Safe for testing environments
- Report issues promptly
- Expect minor changes
- Migration may be required

**Using Stable Features**:
- Production-ready
- Backward compatibility guaranteed
- Full support provided
- Safe to depend on

**Handling Deprecated Features**:
- Plan migration soon
- Follow migration guide
- Test replacement thoroughly
- Update before removal

## Command Filtering

### By Lifecycle Stage

```bash
# All commands (default)
gz --help

# Only stable commands
# (experimental filtered out by default)
gz --help

# Include experimental commands
gz --experimental --help
export GZ_EXPERIMENTAL=1
gz --help
```

### Registry API

Programmatic filtering available:

```go
// Get only stable commands
stableProviders := registry.StableCommands()

// Get experimental commands
expProviders := registry.ExperimentalCommands()

// Filter by lifecycle
for _, p := range registry.List() {
    meta := registry.GetMetadata(p)
    if meta.Lifecycle == registry.LifecycleStable {
        // Use stable command
    }
}
```

## Configuration

### Global Settings

No global configuration needed. Lifecycle is determined by command metadata.

### Environment Variables

- `GZ_EXPERIMENTAL=1` - Enable experimental features

## Migration Examples

### Example 1: Experimental to Stable

```go
// Before (Experimental)
Metadata: CommandMetadata{
    Name:      "new-feature",
    Version:   "0.1.0",
    Lifecycle: LifecycleExperimental,
}

// After Beta testing
Metadata: CommandMetadata{
    Name:      "new-feature",
    Version:   "0.9.0",
    Lifecycle: LifecycleBeta,
}

// After stabilization
Metadata: CommandMetadata{
    Name:      "new-feature",
    Version:   "1.0.0",
    Lifecycle: LifecycleStable,
}
```

### Example 2: Stable to Deprecated

```go
// Before
Metadata: CommandMetadata{
    Name:      "old-cmd",
    Version:   "1.5.0",
    Lifecycle: LifecycleStable,
}

// Deprecation (version 2.0.0)
Metadata: CommandMetadata{
    Name:      "old-cmd",
    Version:   "2.0.0",
    Lifecycle: LifecycleDeprecated,
}

// Removal (version 3.0.0)
// Command registration removed entirely
```

## See Also

- [Extensions System](38-extensions-system.md)
- [Registry System](../20-architecture/25-registry-pattern.md)
- [Command Development Guide](../50-development/development-guide.md)
