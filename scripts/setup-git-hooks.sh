#!/bin/bash
#
# Setup Git hooks for the project

set -e

echo "Setting up Git hooks for gzh-cli..."

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "Error: pre-commit is not installed."
    echo "Please install it using one of the following methods:"
    echo "  - pip install pre-commit"
    echo "  - brew install pre-commit (macOS)"
    echo "  - Download from https://pre-commit.com/#install"
    exit 1
fi

# Install pre-commit hooks
echo "Installing pre-commit hooks..."
pre-commit install --install-hooks

# Install commit-msg hook
echo "Installing commit-msg hook..."
pre-commit install --hook-type commit-msg

# Install pre-push hook
echo "Installing pre-push hook..."
pre-commit install --hook-type pre-push

# Run pre-commit on all files to check current status
echo ""
echo "Running pre-commit checks on all files..."
echo "This may take a while on first run..."
if pre-commit run --all-files; then
    echo "✅ All pre-commit checks passed!"
else
    echo "⚠️  Some pre-commit checks failed. Please fix the issues and commit again."
    echo "You can run 'pre-commit run --all-files' to see the issues."
fi

# Create a custom prepare-commit-msg hook for better commit messages
cat > .git/hooks/prepare-commit-msg << 'EOF'
#!/bin/bash
#
# Prepare commit message with branch name

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# Only add branch name if commit message is empty
if [ -z "$COMMIT_SOURCE" ]; then
    # Get current branch name
    BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)

    # Extract issue number if present (e.g., feature/ISSUE-123-description)
    ISSUE_NUMBER=$(echo "$BRANCH_NAME" | grep -oE '[A-Z]+-[0-9]+' || true)

    # Don't modify if on main/master/develop
    if [[ "$BRANCH_NAME" =~ ^(main|master|develop)$ ]]; then
        exit 0
    fi

    # Read current commit message
    CURRENT_MSG=$(cat "$COMMIT_MSG_FILE")

    # If message is empty and we have an issue number, prepend it
    if [ -z "$CURRENT_MSG" ] && [ -n "$ISSUE_NUMBER" ]; then
        echo "[$ISSUE_NUMBER] " > "$COMMIT_MSG_FILE"
    fi
fi
EOF

chmod +x .git/hooks/prepare-commit-msg

echo ""
echo "✅ Git hooks setup complete!"
echo ""
echo "Hooks installed:"
echo "  - pre-commit: Runs formatting and linting checks"
echo "  - commit-msg: Validates commit message format"
echo "  - pre-push: Runs tests and coverage checks"
echo "  - prepare-commit-msg: Auto-adds issue numbers from branch names"
echo ""
echo "To manually run hooks:"
echo "  - All files: pre-commit run --all-files"
echo "  - Specific hook: pre-commit run <hook-id>"
echo "  - Push hooks: pre-commit run --hook-stage pre-push"
echo ""
echo "To bypass hooks (use sparingly):"
echo "  - git commit --no-verify"
echo "  - git push --no-verify"
