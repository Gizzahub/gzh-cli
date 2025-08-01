#!/bin/bash
# install-aliases.sh - Install GZ backward compatibility aliases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect shell type
SHELL_TYPE=$(basename "$SHELL")
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$HOME/.config/gzh-manager"

# Create config directory if it doesn't exist
mkdir -p "$CONFIG_DIR"

# Copy alias files to config directory
echo -e "${GREEN}Installing GZ backward compatibility aliases...${NC}"

# Copy appropriate alias file
cp "$SCRIPT_DIR/aliases.bash" "$CONFIG_DIR/"
cp "$SCRIPT_DIR/aliases.fish" "$CONFIG_DIR/"

# Function to check if line exists in file
line_exists() {
    local file="$1"
    local pattern="$2"
    [ -f "$file" ] && grep -q "$pattern" "$file" 2>/dev/null
}

# Install aliases based on shell type
install_aliases() {
    case "$SHELL_TYPE" in
        bash)
            local rc_file="$HOME/.bashrc"
            local source_line="source $CONFIG_DIR/aliases.bash"

            if ! line_exists "$rc_file" "gzh-manager/aliases"; then
                echo "" >> "$rc_file"
                echo "# GZ backward compatibility aliases" >> "$rc_file"
                echo "$source_line" >> "$rc_file"
                echo -e "${GREEN}✓ Aliases added to $rc_file${NC}"
            else
                echo -e "${YELLOW}⚠ Aliases already installed in $rc_file${NC}"
            fi
            ;;

        zsh)
            local rc_file="$HOME/.zshrc"
            local source_line="source $CONFIG_DIR/aliases.bash"

            if ! line_exists "$rc_file" "gzh-manager/aliases"; then
                echo "" >> "$rc_file"
                echo "# GZ backward compatibility aliases" >> "$rc_file"
                echo "$source_line" >> "$rc_file"
                echo -e "${GREEN}✓ Aliases added to $rc_file${NC}"
            else
                echo -e "${YELLOW}⚠ Aliases already installed in $rc_file${NC}"
            fi
            ;;

        fish)
            local config_file="$HOME/.config/fish/config.fish"
            local source_line="source $CONFIG_DIR/aliases.fish"

            # Create fish config directory if it doesn't exist
            mkdir -p "$HOME/.config/fish"

            if ! line_exists "$config_file" "gzh-manager/aliases"; then
                echo "" >> "$config_file"
                echo "# GZ backward compatibility aliases" >> "$config_file"
                echo "$source_line" >> "$config_file"
                echo -e "${GREEN}✓ Aliases added to $config_file${NC}"
            else
                echo -e "${YELLOW}⚠ Aliases already installed in $config_file${NC}"
            fi
            ;;

        *)
            echo -e "${RED}Unsupported shell: $SHELL_TYPE${NC}"
            echo "Please manually add the following to your shell configuration:"
            echo "  source $CONFIG_DIR/aliases.bash  (for bash/zsh)"
            echo "  source $CONFIG_DIR/aliases.fish  (for fish)"
            return 1
            ;;
    esac
}

# Main installation
install_aliases

# Create deprecation schedule file
cat > "$CONFIG_DIR/deprecation-schedule.yaml" << 'EOF'
# GZ Deprecation Schedule
deprecation_schedule:
  v2.0.0:
    deprecated_commands:
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

  migration_guide: "docs/migration/command-migration-guide.md"
EOF

echo -e "${GREEN}✓ Deprecation schedule created${NC}"

# Print instructions
echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Reload your shell configuration:"
echo "   - Bash/Zsh: source ~/.${SHELL_TYPE}rc"
echo "   - Fish: source ~/.config/fish/config.fish"
echo "   - Or simply open a new terminal"
echo ""
echo "2. Test the aliases:"
echo "   gz gen-config    # Should show deprecation warning"
echo "   gz migrate help  # Should show migration guide"
echo ""
echo -e "${YELLOW}Note: These aliases will be removed in v3.0 (2025-01-01)${NC}"
