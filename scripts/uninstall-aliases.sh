#!/bin/bash
# uninstall-aliases.sh - Remove GZ backward compatibility aliases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

CONFIG_DIR="$HOME/.config/gzh-manager"

echo -e "${YELLOW}Removing GZ backward compatibility aliases...${NC}"

# Function to remove line from file
remove_line() {
    local file="$1"
    local pattern="$2"
    
    if [ -f "$file" ]; then
        # Create a temporary file
        local temp_file=$(mktemp)
        
        # Remove lines containing the pattern
        grep -v "$pattern" "$file" > "$temp_file" 2>/dev/null || true
        
        # Replace original file
        mv "$temp_file" "$file"
        
        echo -e "${GREEN}✓ Removed aliases from $file${NC}"
    fi
}

# Remove from shell configurations
remove_line "$HOME/.bashrc" "gzh-manager/aliases"
remove_line "$HOME/.zshrc" "gzh-manager/aliases"
remove_line "$HOME/.config/fish/config.fish" "gzh-manager/aliases"

# Remove alias files
if [ -d "$CONFIG_DIR" ]; then
    rm -f "$CONFIG_DIR/aliases.bash"
    rm -f "$CONFIG_DIR/aliases.fish"
    rm -f "$CONFIG_DIR/deprecation-schedule.yaml"
    echo -e "${GREEN}✓ Removed alias files${NC}"
fi

echo ""
echo -e "${GREEN}Uninstallation complete!${NC}"
echo ""
echo "Please reload your shell or open a new terminal for changes to take effect."