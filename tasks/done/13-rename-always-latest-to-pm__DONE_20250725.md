# Task: Rename always-latest Command to pm (Package Manager)

## Objective
`always-latest` 명령어를 더 직관적이고 간결한 `pm` (package manager)로 변경하여 사용성을 개선한다.

## Requirements
- [x] 기존 always-latest 기능을 모두 pm으로 이전
- [x] docs/specs/package-manager.md 사양에 맞춰 구현
- [x] 기존 명령어에 대한 deprecation 경고 추가
- [x] 백워드 호환성 유지

## Steps

### 1. Analyze Current always-latest Command
- [x] cmd/always-latest/ 구조 분석
- [x] 지원하는 패키지 매니저 목록 확인
- [x] 현재 구현된 기능 파악
- [x] 테스트 코드 확인

### 2. Create New pm Command Structure
```go
// cmd/pm/pm.go
package pm

import (
    "github.com/spf13/cobra"
)

func NewPMCmd(ctx context.Context) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "pm",
        Short: "Manage development tools and package managers",
        Long:  `Unified package manager for development environments...`,
    }
    
    // Subcommands based on spec
    cmd.AddCommand(newStatusCmd(ctx))
    cmd.AddCommand(newInstallCmd(ctx))
    cmd.AddCommand(newUpdateCmd(ctx))
    cmd.AddCommand(newSyncCmd(ctx))
    cmd.AddCommand(newExportCmd(ctx))
    cmd.AddCommand(newValidateCmd(ctx))
    cmd.AddCommand(newCleanCmd(ctx))
    cmd.AddCommand(newBootstrapCmd(ctx))
    cmd.AddCommand(newUpgradeManagersCmd(ctx))
    cmd.AddCommand(newMigrateCmd(ctx))
    cmd.AddCommand(newSyncVersionsCmd(ctx))
    
    // Package manager specific commands
    cmd.AddCommand(newBrewCmd(ctx))
    cmd.AddCommand(newAsdfCmd(ctx))
    cmd.AddCommand(newSdkmanCmd(ctx))
    // ... other managers
    
    return cmd
}
```

### 3. Implement Core Commands Based on Spec

#### Status Command
```go
// gz pm status
// Shows status of all configured package managers
```

#### Install Command
```go
// gz pm install
// Install packages from configuration files
```

#### Update Command (existing functionality)
```go
// gz pm update --all
// gz pm update --manager brew
```

#### New Commands from Spec
- sync: Synchronize installed packages with configuration
- export: Export current installations to configuration
- validate: Validate configuration files
- clean: Clean unused packages
- bootstrap: Install missing package managers
- upgrade-managers: Upgrade package managers themselves
- migrate: Migrate packages between versions
- sync-versions: Sync version managers with package managers

### 4. Configuration Structure
```
~/.gzh/pm/
├── global.yml
├── brew.yml
├── asdf.yml
├── sdkman.yml
├── apt.yml
├── pip.yml
├── npm.yml
└── gem.yml
```

### 5. Add Deprecation to always-latest
```go
// cmd/always-latest/always_latest.go
func NewAlwaysLatestCmd(ctx context.Context) *cobra.Command {
    cmd := &cobra.Command{
        Use:        "always-latest",
        Deprecated: "use 'gz pm' instead",
        Short:      "Keep development tools and package managers up to date",
        // ...
    }
}
```

### 6. Update Root Command
```go
// cmd/root.go
// Add new pm command
cmd.AddCommand(pm.NewPMCmd(ctx))
// Keep always-latest for backward compatibility
cmd.AddCommand(alwayslatest.NewAlwaysLatestCmd(ctx))
```

### 7. Create Migration Helper
```go
// Intercept always-latest commands and suggest pm equivalents
// gz always-latest asdf → gz pm update asdf
// gz always-latest brew → gz pm update brew
```

### 8. Implement Package Manager Bootstrap
Based on spec section "Package Manager Bootstrap":
- Auto-detect missing package managers
- Install via official scripts
- Platform-specific installation methods
- Version verification

### 9. Implement Version Coordination
Based on spec section "Version Manager Coordination":
- Node.js/npm sync
- Ruby/gem migration
- Python/pip coordination
- Java multi-version support

### 10. Update Tests
- [ ] Migrate always-latest tests to pm
- [ ] Add tests for new functionality
- [ ] Ensure backward compatibility tests

### 11. Update Documentation
- [x] Update README.md
- [ ] Create docs/commands/pm.md
- [x] Update migration guide
- [x] Add pm to command structure docs

## Expected Output
- New `cmd/pm/` directory with all functionality
- Updated `cmd/root.go` with pm command
- Deprecated always-latest command
- Configuration structure in ~/.gzh/pm/
- Complete test coverage
- Updated documentation

## Verification Criteria
- [x] All always-latest functionality works via pm
- [x] New commands from spec are implemented
- [ ] Configuration files are properly structured
- [x] Deprecation warnings show correct alternatives
- [ ] Tests pass for both old and new commands
- [x] Documentation is complete and accurate

## Notes
- Maintain backward compatibility for at least 6 months
- Focus on user experience improvements
- Consider adding interactive mode for complex operations
- Ensure cross-platform compatibility