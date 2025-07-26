#!/bin/bash
# gz-migrate.sh - Helper script for migrating to new gz command structure

set -e

echo "GZ Command Migration Helper"
echo "=========================="
echo ""

# Check if gz is installed
if ! command -v gz &> /dev/null; then
    echo "Error: gz command not found. Please install gz first."
    exit 1
fi

# Show deprecated commands and their replacements
echo "Command changes in the new version:"
echo ""
echo "  Old Command          →  New Command"
echo "  ─────────────────────────────────────────"
echo "  gz gen-config        →  gz synclone config generate"
echo "  gz repo-config       →  (still available)"
echo "  gz event             →  (still available)"
echo "  gz webhook           →  (still available)"
echo "  gz ssh-config        →  gz dev-env ssh"
echo ""
echo "Note: The old commands will show deprecation warnings."
echo ""

# Search for old commands in common files
echo "Searching for old commands in your configuration files..."
echo ""

SEARCH_PATHS=(
    ~/.bashrc
    ~/.zshrc
    ~/.bash_profile
    ~/.config/fish/config.fish
    ~/.aliases
    ~/.bash_aliases
)

OLD_COMMANDS=(
    "gz gen-config"
    "gz repo-config"
    "gz event"
    "gz webhook"
    "gz ssh-config"
)

found_any=false

for path in "${SEARCH_PATHS[@]}"; do
    if [ -f "$path" ]; then
        for cmd in "${OLD_COMMANDS[@]}"; do
            if grep -q "$cmd" "$path" 2>/dev/null; then
                if [ "$found_any" = false ]; then
                    echo "Found old commands in:"
                    found_any=true
                fi
                echo "  - $path (contains: $cmd)"
            fi
        done
    fi
done

if [ "$found_any" = false ]; then
    echo "No old commands found in common configuration files."
else
    echo ""
    echo "Please update these files manually with the new commands."
fi

echo ""

# Create aliases file
ALIAS_FILE="$HOME/.config/gzh-manager/aliases.sh"
mkdir -p "$HOME/.config/gzh-manager"

echo "Creating compatibility aliases in $ALIAS_FILE..."

cat > "$ALIAS_FILE" << 'EOF'
# GZ backward compatibility aliases
# Source this file in your shell to use old command names

# Command aliases
alias "gz-gen-config"="echo 'Deprecated: use gz synclone config generate' && gz synclone config generate"
# repo-config, event, webhook are still available as standalone commands
# No aliases needed for these
alias "gz-ssh-config"="echo 'Deprecated: use gz dev-env ssh' && gz dev-env ssh"

# Function to intercept gz commands and show deprecation warnings
gz_compat() {
    case "$1" in
        doctor)
            echo "Warning: 'gz doctor' is deprecated. Use 'gz validate --all' instead." >&2
            shift
            command gz validate --all "$@"
            ;;
        shell)
            echo "Warning: 'gz shell' has been removed. Use 'gz --debug-shell' instead." >&2
            return 1
            ;;
        *)
            command gz "$@"
            ;;
    esac
}

# Uncomment to enable compatibility wrapper
# alias gz=gz_compat
EOF

echo ""
echo "Migration helper complete!"
echo ""
echo "To install full backward compatibility aliases (recommended):"
echo "  ./scripts/install-aliases.sh"
echo ""
echo "Or to use the basic compatibility aliases created here:"
echo "  source $ALIAS_FILE"
echo ""
echo "For more information, see: docs/migration/command-migration-guide.md"