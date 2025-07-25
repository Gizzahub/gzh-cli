# Task: Create Migration Scripts for Users

## Objective
사용자가 새로운 명령어 구조로 쉽게 전환할 수 있도록 자동 마이그레이션 스크립트를 제공한다.

## Requirements
- [ ] 사용자 설정 파일 자동 변환
- [ ] 스크립트 및 alias 업데이트
- [ ] 마이그레이션 전 백업 생성
- [ ] 롤백 기능 제공

## Steps

### 1. Migration Script Design
```bash
#!/bin/bash
# gz-migrate.sh - Migrate to new gz command structure

# Features:
# - Detect current gz version
# - Backup existing configurations
# - Convert config files to new format
# - Update shell aliases
# - Provide rollback option
```

### 2. Configuration Migration
```bash
# 설정 파일 마이그레이션 매핑
migrate_configs() {
    # gen-config → synclone
    if [ -f ~/.config/gzh-manager/gen-config.yaml ]; then
        echo "Migrating gen-config settings to synclone..."
        # YAML 변환 로직
    fi
    
    # 통합 config → 개별 명령어 config
    if [ -f ~/.config/gzh-manager/config.yaml ]; then
        echo "Splitting central config to command-specific configs..."
        # 설정 분리 로직
    fi
}
```

### 3. Script Updates Detection
```bash
# 사용자 스크립트에서 구 명령어 사용 검색
find_old_commands() {
    echo "Searching for old gz commands in scripts..."
    
    # Common script locations
    SEARCH_PATHS=(
        ~/.bashrc ~/.zshrc ~/.bash_profile
        ~/.config/fish/config.fish
        ~/bin ~/scripts
        ./.github/workflows
    )
    
    OLD_COMMANDS=(
        "gz gen-config"
        "gz repo-config"
        "gz event"
        "gz webhook"
        "gz ssh-config"
        "gz config"
        "gz doctor"
        "gz shell"
    )
    
    for path in "${SEARCH_PATHS[@]}"; do
        for cmd in "${OLD_COMMANDS[@]}"; do
            grep -r "$cmd" "$path" 2>/dev/null
        done
    done
}
```

### 4. Alias Generation
```bash
# 백워드 호환성을 위한 alias 생성
generate_aliases() {
    cat > ~/.config/gzh-manager/aliases.sh << 'EOF'
# GZ backward compatibility aliases
alias "gz gen-config"="gz synclone config generate"
alias "gz repo-config"="gz repo-sync config"
alias "gz event"="gz repo-sync event"
alias "gz webhook"="gz repo-sync webhook"
alias "gz ssh-config"="gz dev-env ssh"
alias "gz doctor"="gz validate --all"

# Function for config command
gz() {
    if [[ "$1" == "config" && "$2" != "" ]]; then
        echo "Warning: 'gz config' is deprecated. Use command-specific config." >&2
        # Attempt to route to appropriate command
        case "$3" in
            synclone|bulk-clone)
                command gz synclone config "${@:4}"
                ;;
            dev-env|development)
                command gz dev-env config "${@:4}"
                ;;
            net-env|network)
                command gz net-env config "${@:4}"
                ;;
            *)
                echo "Please specify the command: gz [command] config" >&2
                return 1
                ;;
        esac
    else
        command gz "$@"
    fi
}
EOF
}
```

### 5. Interactive Migration Tool
```go
// cmd/migrate/migrate.go
package main

import (
    "fmt"
    "github.com/AlecAivazis/survey/v2"
)

func main() {
    // Interactive migration wizard
    var migrate bool
    prompt := &survey.Confirm{
        Message: "Migrate gz configuration to new structure?",
    }
    survey.AskOne(prompt, &migrate)
    
    if migrate {
        steps := []MigrationStep{
            BackupConfigs,
            MigrateConfigs,
            UpdateScripts,
            GenerateAliases,
            ValidateSetup,
        }
        
        for _, step := range steps {
            if err := step(); err != nil {
                fmt.Printf("Migration failed: %v\n", err)
                promptRollback()
                return
            }
        }
    }
}
```

### 6. Backup and Rollback
```bash
# 백업 생성
create_backup() {
    BACKUP_DIR=~/.config/gzh-manager/backups/$(date +%Y%m%d_%H%M%S)
    mkdir -p "$BACKUP_DIR"
    
    # 설정 파일 백업
    cp -r ~/.config/gzh-manager/*.yaml "$BACKUP_DIR/"
    
    # 메타데이터 저장
    cat > "$BACKUP_DIR/metadata.json" << EOF
{
    "version": "$(gz version)",
    "date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "files": $(find ~/.config/gzh-manager -name "*.yaml" | jq -R -s -c 'split("\n")[:-1]')
}
EOF
}

# 롤백 기능
rollback() {
    LATEST_BACKUP=$(ls -t ~/.config/gzh-manager/backups | head -1)
    if [ -z "$LATEST_BACKUP" ]; then
        echo "No backup found"
        return 1
    fi
    
    echo "Rolling back to $LATEST_BACKUP..."
    cp -r ~/.config/gzh-manager/backups/"$LATEST_BACKUP"/*.yaml ~/.config/gzh-manager/
}
```

### 7. Validation Script
```bash
# 마이그레이션 검증
validate_migration() {
    echo "Validating migration..."
    
    # 새 명령어 테스트
    COMMANDS=(
        "gz synclone config validate"
        "gz dev-env validate"
        "gz net-env validate"
        "gz repo-sync validate"
    )
    
    for cmd in "${COMMANDS[@]}"; do
        if ! $cmd &>/dev/null; then
            echo "❌ Failed: $cmd"
            return 1
        else
            echo "✅ Passed: $cmd"
        fi
    done
}
```

### 8. User Communication
```markdown
# Migration Guide

## What's Changing
- `gen-config` → `synclone config`
- `repo-config`, `event`, `webhook` → `repo-sync`
- `ssh-config` → `dev-env ssh`
- `config` → distributed to each command
- `doctor` → `validate` in each command

## Auto-Migration
Run: `curl -sSL https://gz.dev/migrate | bash`

Or manually:
```bash
wget https://github.com/gzh/releases/latest/gz-migrate.sh
chmod +x gz-migrate.sh
./gz-migrate.sh
```

## Manual Migration Steps
1. Backup your configs
2. Install new gz version
3. Run migration script
4. Update your scripts
5. Test your workflows
```

## Expected Output
- `scripts/migrate.sh` - 메인 마이그레이션 스크립트
- `scripts/rollback.sh` - 롤백 스크립트
- `cmd/migrate/` - Go 기반 대화형 마이그레이션 도구
- `docs/migration/guide.md` - 사용자 가이드
- `~/.config/gzh-manager/aliases.sh` - 호환성 aliases

## Verification Criteria
- [ ] 모든 설정이 올바르게 마이그레이션됨
- [ ] 백업이 생성되고 복원 가능
- [ ] 사용자 스크립트의 구 명령어 감지
- [ ] Aliases가 정상 작동
- [ ] 롤백이 완벽하게 작동

## Notes
- 마이그레이션은 idempotent해야 함 (여러 번 실행 가능)
- 부분 실패 시 자동 롤백
- 사용자에게 각 단계 설명
- 마이그레이션 로그 저장