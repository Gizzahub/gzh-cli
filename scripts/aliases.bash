#!/bin/bash
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
            if [[ "$2" == "help" ]] || [[ -z "$2" ]]; then
                echo "GZ Migration Guide - Command Changes in v2.0"
                echo "==========================================="
                echo ""
                echo "The following commands have been renamed or restructured:"
                echo ""
                echo "  Old Command          →  New Command"
                echo "  -----------             -----------"
                echo "  gz gen-config        →  gz synclone config generate"
                echo "  gz repo-config       →  gz repo-sync config"
                echo "  gz event             →  gz repo-sync event"
                echo "  gz webhook           →  gz repo-sync webhook"
                echo "  gz ssh-config        →  gz dev-env ssh"
                echo "  gz doctor            →  gz validate --all"
                echo "  gz shell             →  gz --debug-shell (hidden)"
                echo "  gz config            →  gz [command] config"
                echo ""
                echo "To install backward compatibility aliases:"
                echo "  source ~/.config/gzh-manager/aliases.bash"
                echo ""
                echo "For more details, see: docs/migration/command-migration-guide.md"
                return 0
            fi
            shift
            command gz "$@"
            ;;
        
        # Default: pass through to actual gz command
        *)
            command gz "$@"
            ;;
    esac
}

# Export the function (only works in bash, not in zsh)
if [[ -n "$BASH_VERSION" ]]; then
    export -f gz
    export -f _gz_deprecation_warning
fi