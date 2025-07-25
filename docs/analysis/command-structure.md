# Current Command Structure Analysis

## Overview
GZH Manager (gz) currently has 18 top-level commands registered in `cmd/root.go`.

## Command List and Structure

### Core Commands

1. **synclone** (`cmd/synclone/`)
   - Purpose: Synchronize and clone repositories from multiple Git hosting services
   - Context: Uses context for cancellation

2. **always-latest** (`cmd/always-latest/`)
   - Purpose: Keep development tools and package managers up to date
   - Context: Uses context

3. **dev-env** (`cmd/dev-env/`)
   - Purpose: Manage development environment configurations
   - Context: No context propagation

4. **net-env** (`cmd/net-env/`)
   - Purpose: Manage network environment transitions
   - Context: Uses context

5. **ide** (`cmd/ide/`)
   - Purpose: Monitor and manage IDE configuration changes
   - Context: Uses context

6. **docker** (`cmd/docker/`)
   - Purpose: Container image management and automation

7. **repo-sync** (`cmd/repo-sync/`)
   - Purpose: Advanced repository synchronization and management
   - Context: Uses context

### Configuration Commands

8. **config** (`cmd/config/`)
   - Purpose: Configuration management commands (generic)
   - Context: No context propagation

9. **gen-config** (`cmd/gen-config/`)
   - Purpose: Generate bulk-clone configuration files
   - Context: Uses context

10. **repo-config** (`cmd/repo-config/`)
    - Purpose: GitHub repository configuration management
    - Context: No context propagation

11. **ssh-config** (`cmd/ssh-config/`)
    - Purpose: SSH configuration management for Git operations

### Utility Commands

12. **doctor** (`cmd/doctor/`)
    - Purpose: Diagnose system health and configuration issues

13. **migrate** (`cmd/migrate/`)
    - Purpose: Migrate configuration files to unified format
    - Context: No context propagation

14. **shell** (`cmd/shell/`)
    - Purpose: Start interactive debugging shell (REPL)

15. **version** (inline in root.go)
    - Purpose: Show version information

### GitHub Integration Commands

16. **webhook** (`cmd/webhook.go`)
    - Purpose: GitHub webhook management

17. **event** (`cmd/event.go`)
    - Purpose: GitHub event management and webhook server

## Command Registration Order

```go
1. version (inline)
2. always-latest
3. synclone  
4. config
5. doctor
6. dev-env
7. docker
8. gen-config
9. ide
10. migrate
11. net-env
12. repo-config
13. repo-sync
14. shell
15. ssh-config
16. webhook
17. event
```

## Global Flags

- `--verbose, -v`: Enable verbose logging
- `--debug`: Enable debug logging (shows all log levels)
- `--quiet, -q`: Suppress all logs except critical errors

## Context Usage Pattern

Commands fall into two categories:
1. **Context-aware**: Uses context for cancellation and timeouts
   - always-latest, synclone, gen-config, ide, net-env, repo-sync
2. **No context**: Commands that don't propagate context
   - config, dev-env, migrate, repo-config, event

## Key Observations

1. **Overlapping Functionality**:
   - `gen-config` generates configs, but `config` is the generic config manager
   - `repo-config` and `repo-sync` both handle repository management
   - `webhook` and `event` are closely related GitHub features

2. **Inconsistent Patterns**:
   - Some commands use NewXxxCmd() pattern, others use direct Cmd export
   - Context propagation is inconsistent

3. **Command Grouping Opportunities**:
   - GitHub-related: webhook, event, repo-config → could merge into repo-sync
   - Config-related: config, gen-config → could merge
   - Environment: dev-env could absorb ssh-config