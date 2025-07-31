#!/bin/bash
# GZH Manager v2.0 Backward Compatibility Aliases
# These aliases will be removed in v3.0 (estimated: 2025-01-01)
# 
# Usage:
#   source scripts/deprecated-aliases.sh  # For bash/zsh
#   source scripts/deprecated-aliases.fish  # For fish shell

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
    echo -e "This alias will be removed in v3.0.\n" >&2
}

# Function wrapper for gz command
gz() {
    case "$1" in
        # Removed commands with error messages
        "shell")
            echo -e "${RED}Error:${NC} 'gz shell' has been removed." >&2
            echo -e "For debugging, use: ${YELLOW}gz --debug-shell${NC}" >&2
            return 1
            ;;
        "config")
            echo -e "${RED}Error:${NC} 'gz config' has been removed." >&2
            echo -e "Use: ${YELLOW}gz [command] config${NC} instead." >&2
            return 1
            ;;
        "docker")
            echo -e "${RED}Error:${NC} 'gz docker' has been removed." >&2
            echo -e "Please use Docker CLI directly." >&2
            return 1
            ;;
        # Deprecated but still working
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

# Export the function (bash only)
if [[ -n "$BASH_VERSION" ]]; then
    export -f gz _gz_deprecation_warning
fi