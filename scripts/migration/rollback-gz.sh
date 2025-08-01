#!/bin/bash
# rollback-gz.sh - Rollback helper for gz command migration

set -e

echo "GZ Command Rollback Helper"
echo "========================="
echo ""

# Remove aliases file
ALIAS_FILE="$HOME/.config/gzh-manager/aliases.sh"

if [ -f "$ALIAS_FILE" ]; then
    echo "Removing compatibility aliases file..."
    rm -f "$ALIAS_FILE"
    echo "✓ Removed: $ALIAS_FILE"
else
    echo "No aliases file found."
fi

echo ""
echo "Rollback complete!"
echo ""
echo "Note: The deprecated commands will continue to work with warnings."
echo "If you modified any scripts, you'll need to revert those changes manually."
