# Task: Update Root Command to Reflect New Structure

## Objective
cmd/root.go를 업데이트하여 새로운 간소화된 명령어 구조를 반영하고, deprecated 명령어들을 제거한다.

## Requirements
- [ ] 새로운 명령어 구조로 root command 업데이트
- [ ] Deprecated 명령어 제거 또는 숨김 처리
- [ ] 명령어 그룹화 및 정렬
- [ ] Help 텍스트 개선

## Steps

### 1. Analyze Current Root Command
- [ ] cmd/root.go의 현재 구조 분석
- [ ] 모든 AddCommand 호출 목록화
- [ ] 명령어 초기화 순서 파악
- [ ] 전역 플래그 확인

### 2. New Command Structure
```go
// 새로운 명령어 구조
Core Commands:
  synclone      Synchronize and clone repositories (includes gen-config)
  dev-env       Development environment management (includes ssh-config)
  net-env       Network environment management
  repo-sync     Advanced repository synchronization (includes webhook, event, repo-config)
  
Tool Commands:
  ide           IDE configuration management
  always-latest Package manager updates
  docker        Container management
  
Utility Commands:
  validate      Run validation across all components (replaces doctor)
  completion    Generate shell completions
  version       Show version information
  help          Help about any command
```

### 3. Update Root Command
```go
// cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
    
    // Core commands
    "github.com/yourusername/gzh-manager-go/cmd/synclone"
    "github.com/yourusername/gzh-manager-go/cmd/devenv"
    "github.com/yourusername/gzh-manager-go/cmd/netenv"
    "github.com/yourusername/gzh-manager-go/cmd/reposync"
    
    // Tool commands
    "github.com/yourusername/gzh-manager-go/cmd/ide"
    "github.com/yourusername/gzh-manager-go/cmd/alwayslatest"
    "github.com/yourusername/gzh-manager-go/cmd/docker"
    
    // Utility commands
    "github.com/yourusername/gzh-manager-go/cmd/validate"
    "github.com/yourusername/gzh-manager-go/cmd/completion"
    "github.com/yourusername/gzh-manager-go/cmd/version"
)

var rootCmd = &cobra.Command{
    Use:   "gz",
    Short: "GZH Manager - Unified development environment and repository management",
    Long: `GZH Manager (gz) is a comprehensive CLI tool for managing development 
environments and Git repositories across multiple platforms.

Core Features:
  • Repository synchronization from GitHub, GitLab, Gitea, and Gogs
  • Development environment configuration management
  • Network environment transitions and profiles
  • Advanced repository synchronization with webhooks

Use "gz [command] --help" for more information about a command.`,
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: $HOME/.config/gzh-manager/config.yaml)")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
    rootCmd.PersistentFlags().Bool("json", false, "output in JSON format")
    
    // Hidden debug flag
    rootCmd.PersistentFlags().Bool("debug-shell", false, "")
    rootCmd.PersistentFlags().MarkHidden("debug-shell")
    
    // Add commands in organized groups
    addCoreCommands()
    addToolCommands()
    addUtilityCommands()
    
    // Handle deprecated commands
    handleDeprecatedCommands()
}

func addCoreCommands() {
    rootCmd.AddCommand(synclone.NewCommand())
    rootCmd.AddCommand(devenv.NewCommand())
    rootCmd.AddCommand(netenv.NewCommand())
    rootCmd.AddCommand(reposync.NewCommand())
}

func addToolCommands() {
    rootCmd.AddCommand(ide.NewCommand())
    rootCmd.AddCommand(alwayslatest.NewCommand())
    rootCmd.AddCommand(docker.NewCommand())
}

func addUtilityCommands() {
    rootCmd.AddCommand(validate.NewCommand())
    rootCmd.AddCommand(completion.NewCommand())
    rootCmd.AddCommand(version.NewCommand())
}
```

### 4. Handle Deprecated Commands
```go
// cmd/deprecated.go
package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

func handleDeprecatedCommands() {
    // Create hidden deprecated commands that show migration messages
    deprecatedMappings := map[string]string{
        "gen-config":  "synclone config generate",
        "repo-config": "repo-sync config",
        "event":       "repo-sync event",
        "webhook":     "repo-sync webhook",
        "ssh-config":  "dev-env ssh",
        "doctor":      "validate --all",
        "shell":       "--debug-shell flag",
    }
    
    for old, new := range deprecatedMappings {
        oldCmd := old
        newCmd := new
        cmd := &cobra.Command{
            Use:    oldCmd,
            Hidden: true,
            Short:  fmt.Sprintf("(DEPRECATED) Use 'gz %s' instead", newCmd),
            Run: func(cmd *cobra.Command, args []string) {
                fmt.Fprintf(os.Stderr, 
                    "Error: '%s' has been deprecated.\n"+
                    "Please use 'gz %s' instead.\n\n"+
                    "For more information, run: gz help migrate\n", 
                    oldCmd, newCmd)
                os.Exit(1)
            },
        }
        rootCmd.AddCommand(cmd)
    }
    
    // Special handling for 'config' command
    configCmd := &cobra.Command{
        Use:    "config",
        Hidden: true,
        Short:  "(DEPRECATED) Use command-specific config",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Fprintf(os.Stderr, 
                "Error: 'config' has been distributed to individual commands.\n"+
                "Use 'gz [command] config' instead:\n"+
                "  - gz synclone config\n"+
                "  - gz dev-env config\n"+
                "  - gz net-env config\n"+
                "  - gz repo-sync config\n\n"+
                "For more information, run: gz help migrate\n")
            os.Exit(1)
        },
    }
    rootCmd.AddCommand(configCmd)
}
```

### 5. Improve Help Output
```go
// cmd/root.go
func init() {
    // Custom help template
    rootCmd.SetHelpTemplate(`{{.Long}}

{{if .HasAvailableSubCommands}}Core Commands:{{range .Commands}}{{if and (not .Hidden) (has .Annotations "group:core")}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Tool Commands:{{range .Commands}}{{if and (not .Hidden) (has .Annotations "group:tool")}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Utility Commands:{{range .Commands}}{{if and (not .Hidden) (has .Annotations "group:utility")}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

{{end}}{{if .HasAvailableFlags}}Flags:
{{.Flags.FlagUsages | trimTrailingWhitespaces}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`)
}

// In each command's init
func NewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "synclone",
        Short: "Synchronize and clone repositories from multiple Git platforms",
        Annotations: map[string]string{
            "group": "core",
        },
    }
    return cmd
}
```

### 6. Add Migration Help
```go
// cmd/migrate_help.go
var migrateHelpCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Show migration guide for deprecated commands",
    Long: `Migration Guide for GZ Command Structure Changes

The following commands have been reorganized:

  Old Command          →  New Command
  ─────────────────────────────────────────
  gen-config           →  synclone config generate
  repo-config          →  repo-sync config
  event                →  repo-sync event  
  webhook              →  repo-sync webhook
  ssh-config           →  dev-env ssh
  config               →  [command] config
  doctor               →  validate --all
  shell                →  --debug-shell flag

To automatically migrate your configuration:
  curl -sSL https://gz.dev/migrate | bash

Or manually:
  1. Update your scripts to use new commands
  2. Run 'gz validate --all' to check your setup
  3. Update any aliases or automation

For backward compatibility, add these aliases:
  source ~/.config/gzh-manager/aliases.sh
`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(cmd.Long)
    },
}
```

## Expected Output
- Updated `cmd/root.go` with new structure
- `cmd/deprecated.go` for handling old commands
- `cmd/migrate_help.go` for migration guidance
- Improved help output with command grouping
- Hidden debug features

## Verification Criteria
- [ ] Only 10 main commands visible in help
- [ ] Commands are logically grouped
- [ ] Deprecated commands show helpful error messages
- [ ] Help text is clear and organized
- [ ] Debug features are hidden but functional
- [ ] All new commands are properly registered

## Notes
- Maintain alphabetical order within groups
- Ensure consistent naming conventions
- Keep help text concise but informative
- Test help output on different terminal widths