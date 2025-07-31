#!/bin/bash
# GZH Manager v2.0 Unified Backward Compatibility Aliases
# These aliases will be removed in v3.0 (estimated: 2025-01-01)

# Detect shell type and use appropriate implementation
if [ -n "$BASH_VERSION" ]; then
    # Bash implementation
    source "$(dirname "${BASH_SOURCE[0]}")/aliases.bash"
elif [ -n "$ZSH_VERSION" ]; then
    # Zsh implementation (similar to bash)
    source "$(dirname "${BASH_SOURCE[0]}")/aliases.bash"
elif [ -n "$FISH_VERSION" ]; then
    # Fish implementation
    source "$(dirname "${BASH_SOURCE[0]}")/aliases.fish"
else
    echo "Warning: Shell not detected. Using bash-compatible aliases."
    source "$(dirname "${BASH_SOURCE[0]}")/aliases.bash"
fi
