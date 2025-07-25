# Task: Create Backward Compatibility Aliases for Old Commands

## Objective
기존 사용자의 스크립트와 워크플로우가 깨지지 않도록 이전 명령어에 대한 백워드 호환성 별칭을 제공한다.

## Requirements
- [x] 모든 deprecated 명령어에 대한 alias 생성
- [x] 다양한 shell 환경 지원 (bash, zsh, fish)
- [x] Deprecation 경고 메시지 표시
- [x] 일정 기간 후 제거 계획

## Steps

### 1. Alias Strategy Design
```bash
# 별칭 동작 방식
# 1. 구 명령어 실행 시 경고 메시지 표시
# 2. 새 명령어로 자동 리다이렉트
# 3. 사용 통계 로깅 (선택적)
# 4. 제거 예정일 안내
```

### 2. Shell-Specific Alias Files

#### Bash/Zsh Aliases
```bash
# ~/.config/gzh-manager/aliases.bash
# GZH Manager v2.0 Backward Compatibility Aliases
# These aliases will be removed in v3.0 (estimated: 2025-01-01)

# Color codes for warnings
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Deprecation warning function
_gz_deprecation_warning() {
    local old_cmd="$1"
    local new_cmd="$2"
    echo -e "${YELLOW}Warning:${NC} '${RED}gz ${old_cmd}${NC}' is deprecated." >&2
    echo -e "Please use '${YELLOW}gz ${new_cmd}${NC}' instead." >&2
    echo -e "This alias will be removed in v3.0. Run 'gz help migrate' for details.\n" >&2
}

# Function wrapper for gz command
gz() {
    case "$1" in
        # gen-config -> synclone config
        "gen-config")
            _gz_deprecation_warning "gen-config" "synclone config generate"
            shift
            command gz synclone config generate "$@"
            ;;
        
        # repo-config -> repo-sync config
        "repo-config")
            _gz_deprecation_warning "repo-config" "repo-sync config"
            shift
            command gz repo-sync config "$@"
            ;;
        
        # event -> repo-sync event
        "event")
            _gz_deprecation_warning "event" "repo-sync event"
            shift
            command gz repo-sync event "$@"
            ;;
        
        # webhook -> repo-sync webhook
        "webhook")
            _gz_deprecation_warning "webhook" "repo-sync webhook"
            shift
            command gz repo-sync webhook "$@"
            ;;
        
        # ssh-config -> dev-env ssh
        "ssh-config")
            _gz_deprecation_warning "ssh-config" "dev-env ssh"
            shift
            command gz dev-env ssh "$@"
            ;;
        
        # doctor -> validate --all
        "doctor")
            _gz_deprecation_warning "doctor" "validate --all"
            shift
            command gz validate --all "$@"
            ;;
        
        # shell -> error with instruction
        "shell")
            echo -e "${RED}Error:${NC} 'gz shell' has been removed." >&2
            echo -e "For debugging, use: ${YELLOW}gz --debug-shell${NC}" >&2
            echo -e "or set: ${YELLOW}export GZH_DEBUG_SHELL=1${NC}" >&2
            return 1
            ;;
        
        # config -> special handling
        "config")
            if [[ -n "$2" ]]; then
                echo -e "${RED}Error:${NC} 'gz config' has been distributed to individual commands." >&2
                echo -e "Use: ${YELLOW}gz [command] config${NC} instead. For example:" >&2
                echo -e "  - gz synclone config" >&2
                echo -e "  - gz dev-env config" >&2
                echo -e "  - gz net-env config" >&2
                echo -e "  - gz repo-sync config" >&2
                return 1
            fi
            ;;
        
        # migrate -> special command for migration help
        "migrate")
            shift
            command gz help migrate "$@"
            ;;
        
        # Default: pass through to actual gz command
        *)
            command gz "$@"
            ;;
    esac
}

# Export the function
export -f gz
export -f _gz_deprecation_warning
```

#### Fish Shell Aliases
```fish
# ~/.config/gzh-manager/aliases.fish
# GZH Manager v2.0 Backward Compatibility Aliases

function _gz_deprecation_warning
    set -l old_cmd $argv[1]
    set -l new_cmd $argv[2]
    echo (set_color yellow)"Warning:"(set_color normal) \
         (set_color red)"gz $old_cmd"(set_color normal) "is deprecated." >&2
    echo "Please use" (set_color yellow)"gz $new_cmd"(set_color normal) "instead." >&2
    echo "This alias will be removed in v3.0. Run 'gz help migrate' for details.\n" >&2
end

function gz
    switch $argv[1]
        case "gen-config"
            _gz_deprecation_warning "gen-config" "synclone config generate"
            command gz synclone config generate $argv[2..-1]
        
        case "repo-config"
            _gz_deprecation_warning "repo-config" "repo-sync config"
            command gz repo-sync config $argv[2..-1]
        
        case "event"
            _gz_deprecation_warning "event" "repo-sync event"
            command gz repo-sync event $argv[2..-1]
        
        case "webhook"
            _gz_deprecation_warning "webhook" "repo-sync webhook"
            command gz repo-sync webhook $argv[2..-1]
        
        case "ssh-config"
            _gz_deprecation_warning "ssh-config" "dev-env ssh"
            command gz dev-env ssh $argv[2..-1]
        
        case "doctor"
            _gz_deprecation_warning "doctor" "validate --all"
            command gz validate --all $argv[2..-1]
        
        case "shell"
            echo (set_color red)"Error:"(set_color normal) "'gz shell' has been removed." >&2
            echo "For debugging, use:" (set_color yellow)"gz --debug-shell"(set_color normal) >&2
            return 1
        
        case "config"
            if test (count $argv) -gt 1
                echo (set_color red)"Error:"(set_color normal) \
                     "'gz config' has been distributed to individual commands." >&2
                echo "Use:" (set_color yellow)"gz [command] config"(set_color normal) "instead." >&2
                return 1
            end
        
        case "*"
            command gz $argv
    end
end
```

### 3. PowerShell Aliases
```powershell
# ~/.config/gzh-manager/aliases.ps1
# GZH Manager v2.0 Backward Compatibility Aliases

function Show-GzDeprecationWarning {
    param(
        [string]$OldCommand,
        [string]$NewCommand
    )
    Write-Host "Warning: " -ForegroundColor Yellow -NoNewline
    Write-Host "'gz $OldCommand'" -ForegroundColor Red -NoNewline
    Write-Host " is deprecated." -ForegroundColor White
    Write-Host "Please use " -NoNewline
    Write-Host "'gz $NewCommand'" -ForegroundColor Yellow -NoNewline
    Write-Host " instead." -ForegroundColor White
    Write-Host "This alias will be removed in v3.0.`n" -ForegroundColor Gray
}

function gz {
    $cmd = $args[0]
    $remainingArgs = $args[1..($args.Length-1)]
    
    switch ($cmd) {
        "gen-config" {
            Show-GzDeprecationWarning "gen-config" "synclone config generate"
            & gz synclone config generate @remainingArgs
        }
        "repo-config" {
            Show-GzDeprecationWarning "repo-config" "repo-sync config"
            & gz repo-sync config @remainingArgs
        }
        # ... other aliases
        default {
            & gz @args
        }
    }
}
```

### 4. Installation Script
```bash
#!/bin/bash
# install-aliases.sh

SHELL_TYPE=$(basename "$SHELL")
CONFIG_DIR="$HOME/.config/gzh-manager"

install_aliases() {
    echo "Installing GZ backward compatibility aliases..."
    
    case "$SHELL_TYPE" in
        bash)
            echo "source $CONFIG_DIR/aliases.bash" >> ~/.bashrc
            echo "Aliases added to ~/.bashrc"
            ;;
        zsh)
            echo "source $CONFIG_DIR/aliases.bash" >> ~/.zshrc
            echo "Aliases added to ~/.zshrc"
            ;;
        fish)
            echo "source $CONFIG_DIR/aliases.fish" >> ~/.config/fish/config.fish
            echo "Aliases added to ~/.config/fish/config.fish"
            ;;
        *)
            echo "Unsupported shell: $SHELL_TYPE"
            echo "Please manually source the appropriate alias file"
            ;;
    esac
}

# Check if aliases already installed
if grep -q "gzh-manager/aliases" ~/.*rc 2>/dev/null; then
    echo "Aliases already installed"
else
    install_aliases
fi
```

### 5. Usage Tracking (Optional)
```bash
# Add to alias function for usage statistics
_gz_track_deprecated_usage() {
    local cmd="$1"
    local log_file="$HOME/.config/gzh-manager/deprecated-usage.log"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) $cmd" >> "$log_file"
}

# In each deprecated command:
_gz_track_deprecated_usage "gen-config"
```

### 6. Removal Schedule
```yaml
# ~/.config/gzh-manager/deprecation-schedule.yaml
deprecation_schedule:
  v2.0.0:
    deprecated:
      - gen-config
      - repo-config
      - event
      - webhook
      - ssh-config
      - config
      - doctor
      - shell
    removal_target: v3.0.0
    removal_date: "2025-01-01"
    
  warnings:
    first_warning: "2024-06-01"  # 6 months before removal
    final_warning: "2024-12-01"  # 1 month before removal
```

### 7. Auto-Update Check
```bash
# Add to aliases to check deprecation status
_gz_check_deprecation_date() {
    local removal_date="2025-01-01"
    local current_date=$(date +%Y-%m-%d)
    local days_until_removal=$(( ($(date -d "$removal_date" +%s) - $(date +%s)) / 86400 ))
    
    if [[ $days_until_removal -lt 30 ]]; then
        echo -e "${RED}FINAL WARNING: These aliases will be removed in $days_until_removal days!${NC}" >&2
    fi
}
```

## Expected Output
- `~/.config/gzh-manager/aliases.bash` - Bash/Zsh aliases
- `~/.config/gzh-manager/aliases.fish` - Fish shell aliases
- `~/.config/gzh-manager/aliases.ps1` - PowerShell aliases
- `scripts/install-aliases.sh` - Installation script
- `~/.config/gzh-manager/deprecation-schedule.yaml` - Removal timeline

## Verification Criteria
- [x] All deprecated commands show warning messages
- [x] Commands redirect to correct new versions
- [x] Aliases work in bash, zsh, and fish
- [x] Installation script correctly modifies shell configs
- [x] Usage tracking works (if implemented)
- [x] Removal date is clearly communicated

## Notes
- 최소 6개월의 deprecation 기간 제공
- 경고 메시지는 표준 에러(stderr)로 출력
- 사용 통계는 선택적 기능으로 제공
- 제거 예정일이 가까워질수록 경고 강도 증가