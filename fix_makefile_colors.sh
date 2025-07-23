#!/bin/bash

# Script to fix ANSI color output in Makefiles by adding -e flag to echo commands
# that contain color variables

echo "Fixing Makefile color output..."

# List of Makefile files to process
MAKEFILES=("Makefile" "Makefile.quality.mk" "Makefile.tools.mk" "Makefile.deps.mk" "Makefile.docker.mk" "Makefile.dev.mk" "Makefile.build.mk" "Makefile.test.mk")

for makefile in "${MAKEFILES[@]}"; do
    if [[ -f "$makefile" ]]; then
        echo "Processing $makefile..."

        # Create backup
        cp "$makefile" "$makefile.backup"

        # Replace @echo with @echo -e when the line contains color variables
        sed -i 's/@echo "\([^"]*\$([A-Z_]*)[^"]*\)"/@echo -e "\1"/g' "$makefile"

        # Also handle single quotes
        sed -i "s/@echo '\([^']*\$([A-Z_]*)[^']*\)'/@echo -e '\1'/g" "$makefile"

        echo "  Fixed $makefile"
    else
        echo "  Warning: $makefile not found"
    fi
done

echo "Done! Backups created with .backup extension"
echo ""
echo "Testing with make lint-count..."
