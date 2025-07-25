# Task: Complete always-latest to pm Migration

## Objective
always-latest를 완전히 pm으로 통합하고 중복되거나 package-manager.md 스펙과 맞지 않는 항목을 제거한다.

## Completed Steps

### 1. Removed always-latest Command
- [x] Deleted entire `cmd/always-latest/` directory
- [x] Removed always-latest import from `cmd/root.go`
- [x] Removed always-latest command registration from root

### 2. Migrated Implementation to pm
- [x] Moved brew update logic to pm/update.go
- [x] Moved asdf update logic to pm/update.go
- [x] Moved sdkman update logic to pm/update.go
- [x] Moved apt update logic to pm/update.go
- [x] Added necessary imports (os, os/exec, strings)

### 3. Updated Documentation
- [x] Updated README.md to remove always-latest from command list
- [x] Updated CLAUDE.md command structure documentation
- [x] Updated CLAUDE.md development environment section

### 4. Maintained Backward Compatibility
- [x] Shell aliases (bash and fish) still support `gz always-latest` with deprecation warnings
- [x] Aliases redirect to `gz pm` commands

## Implementation Details

The pm command now includes all functionality from always-latest:
- `gz pm update --manager brew` - Update Homebrew packages
- `gz pm update --manager asdf` - Update asdf plugins and tools
- `gz pm update --manager sdkman` - Update SDKMAN candidates
- `gz pm update --manager apt` - Update APT packages
- `gz pm update --all` - Update all package managers

The implementation follows the package-manager.md specification with:
- Unified update strategies (latest, stable, minor, fixed)
- Dry-run support
- Configuration-based management
- Bootstrap and migration features

## Verification
- [x] Build successful (`go build ./...`)
- [x] Linter passes (with existing unrelated warnings)
- [x] No references to always-latest in active code
- [x] Backward compatibility maintained through aliases

## Notes
- The legacy.go file in pm already had placeholder commands for compatibility
- The update.go implementation now contains the actual logic from always-latest
- All package manager specific commands (asdf, brew, etc.) are available as subcommands of pm
- Migration guide already documents the deprecation