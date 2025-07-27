# Generic Package Manager Pattern Design

## Overview

This document outlines the design for implementing a generic `gz pm [manager]` pattern that would allow dynamic access to any package manager without explicitly defining each one.

## Current State

Currently, we have explicit commands for each package manager:
- `gz pm brew`
- `gz pm asdf`
- `gz pm pip`
- `gz pm npm`
- etc.

## Proposed Generic Pattern

### Command Structure

```bash
gz pm [manager] [subcommand] [args...]
```

### Implementation Approach

1. **Dynamic Command Registration**
   - Use cobra's `DisableFlagParsing` to capture all arguments
   - Parse the first argument as the package manager name
   - Pass remaining arguments to the package manager

2. **Package Manager Registry**
   ```go
   type PackageManager interface {
       Name() string
       IsInstalled() bool
       Execute(args []string) error
   }
   
   var registry = map[string]PackageManager{
       "brew": &BrewManager{},
       "apt":  &AptManager{},
       // etc.
   }
   ```

3. **Fallback to System Command**
   - If manager not in registry, attempt to execute as system command
   - Provides flexibility for new/unknown package managers

### Example Implementation

```go
func newGenericPMCmd(ctx context.Context) *cobra.Command {
    return &cobra.Command{
        Use:                "pm [manager] [args...]",
        Short:              "Execute package manager commands",
        DisableFlagParsing: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) < 1 {
                return fmt.Errorf("specify a package manager")
            }
            
            manager := args[0]
            pmArgs := args[1:]
            
            // Check registry first
            if pm, ok := registry[manager]; ok {
                return pm.Execute(pmArgs)
            }
            
            // Fallback to system command
            return executeSystemCommand(ctx, manager, pmArgs)
        },
    }
}
```

### Challenges

1. **Command Discovery**: How to provide help/completion for dynamic commands
2. **Error Handling**: Distinguishing between "manager not found" vs "command failed"
3. **Security**: Preventing arbitrary command execution
4. **Configuration Integration**: How to integrate with unified config system

### Benefits

1. **Extensibility**: Support new package managers without code changes
2. **Flexibility**: Users can access any package manager command
3. **Consistency**: Unified interface for all package managers

### Migration Path

1. Keep existing explicit commands for backward compatibility
2. Implement generic pattern alongside explicit commands
3. Eventually deprecate explicit commands in favor of generic pattern

## Decision

For now, we'll continue with explicit commands as they provide:
- Better documentation and help text
- Type safety and validation
- Clear command structure
- Better integration with configuration system

The generic pattern can be revisited when there's a clear need for supporting many more package managers dynamically.