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
        # repo-config -> kept as is
        "repo-config")
            # Still available, no deprecation needed
            shift
            command gz repo-config "$@"
            ;;
        
        # event -> kept as is
        "event")
            # Still available, no deprecation needed
            shift
            command gz event "$@"
            ;;
        
        # webhook -> kept as is
        "webhook")
            # Still available, no deprecation needed
            shift
            command gz webhook "$@"
            ;;
        
        # doctor -> still available
        "doctor")
            # Still available, no deprecation needed
            shift
            command gz doctor "$@"
            ;;
        
        # shell -> error with instruction
        "shell")
            echo -e "${RED}Error:${NC} 'gz shell' has been removed." >&2
            echo -e "For debugging, use: ${YELLOW}gz --debug-shell${NC}" >&2
            echo -e "or set: ${YELLOW}export GZH_DEBUG_SHELL=1${NC}" >&2
            return 1
            ;;
        
        # config -> removed
        "config")
            echo -e "${RED}Error:${NC} 'gz config' has been removed." >&2
            echo -e "Use: ${YELLOW}gz [command] config${NC} instead. For example:" >&2
            echo -e "  - gz synclone config" >&2
            echo -e "  - gz dev-env config" >&2
            echo -e "  - gz net-env config" >&2
            return 1
            ;;
        
        # docker -> removed
        "docker")
            echo -e "${RED}Error:${NC} 'gz docker' has been removed." >&2
            echo -e "Please use Docker CLI directly for container management." >&2
            return 1
            ;;
        
        # always-latest -> pm
        "always-latest")
            _gz_deprecation_warning "always-latest" "pm"
            shift
            command gz pm "$@"
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