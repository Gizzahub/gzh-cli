#!/bin/bash
#
# Example workflow for using gz bulk-clone command
# This script demonstrates common usage patterns

set -e

# Configuration
REPOS_DIR="${HOME}/repos"
CONFIG_FILE="${HOME}/.config/gzh-manager/bulk-clone.yaml"
LOG_FILE="/tmp/gz-clone-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $*" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" | tee -a "$LOG_FILE" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" | tee -a "$LOG_FILE"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    # Check if gz is installed
    if ! command -v gz &> /dev/null; then
        error "gz command not found. Please install gzh-manager-go first."
        exit 1
    fi

    # Check if config exists
    if [[ ! -f "$CONFIG_FILE" ]]; then
        warn "Config file not found at $CONFIG_FILE"
        log "Creating example config..."
        mkdir -p "$(dirname "$CONFIG_FILE")"
        cat > "$CONFIG_FILE" <<EOF
# Bulk clone configuration
github:
  - organization: "my-org"
    token: "\${GITHUB_TOKEN}"
    targetPath: "${REPOS_DIR}/github"
    strategy: "reset"

settings:
  concurrency: 5
  retryAttempts: 3
  retryDelay: "5s"
EOF
        log "Please edit $CONFIG_FILE with your settings"
        exit 1
    fi

    # Check tokens
    if [[ -z "$GITHUB_TOKEN" ]]; then
        warn "GITHUB_TOKEN not set. Only public repos will be accessible."
    fi
}

# Validate configuration
validate_config() {
    log "Validating configuration..."

    if ! gz bulk-clone validate --config "$CONFIG_FILE"; then
        error "Configuration validation failed"
        exit 1
    fi

    log "Configuration is valid"
}

# Discovery phase - see what would be cloned
discover_repos() {
    log "Discovering repositories..."

    # Run in dry-run mode to see what would be cloned
    gz bulk-clone \
        --config "$CONFIG_FILE" \
        --dry-run \
        --log-level info \
        2>&1 | tee -a "$LOG_FILE"

    echo
    read -p "Proceed with cloning? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log "Clone cancelled by user"
        exit 0
    fi
}

# Perform the actual clone
clone_repos() {
    log "Starting bulk clone operation..."
    log "Target directory: $REPOS_DIR"
    log "Log file: $LOG_FILE"

    # Create target directory
    mkdir -p "$REPOS_DIR"

    # Run bulk clone
    if gz bulk-clone \
        --config "$CONFIG_FILE" \
        --log-level info \
        --progress \
        2>&1 | tee -a "$LOG_FILE"; then
        log "Bulk clone completed successfully"
    else
        error "Bulk clone failed. Check log file: $LOG_FILE"
        exit 1
    fi
}

# Post-clone operations
post_clone() {
    log "Running post-clone operations..."

    # Generate summary
    log "Generating repository summary..."

    total_repos=$(find "$REPOS_DIR" -name ".git" -type d | wc -l)
    total_size=$(du -sh "$REPOS_DIR" 2>/dev/null | cut -f1)

    cat >> "$LOG_FILE" <<EOF

=== Clone Summary ===
Total repositories: $total_repos
Total disk usage: $total_size
Target directory: $REPOS_DIR
Timestamp: $(date)
====================
EOF

    log "Summary:"
    log "  - Total repositories: $total_repos"
    log "  - Total disk usage: $total_size"

    # Optional: Update all repos
    read -p "Update all repositories now? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "Updating all repositories..."
        gz bulk-clone refresh \
            --target-path "$REPOS_DIR" \
            --strategy pull \
            --concurrency 5
    fi
}

# Main workflow
main() {
    log "Starting gz bulk clone workflow"

    check_prerequisites
    validate_config
    discover_repos
    clone_repos
    post_clone

    log "Workflow completed. Log saved to: $LOG_FILE"
}

# Run main function
main "$@"
