#!/bin/bash
# Git Extensions Uninstall Script for gzh-manager-go
# Removes git-synclone and other Git extensions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="git-synclone"
BACKUP_DIR="${HOME}/.local/backup/gzh"

# Function to print colored messages
print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to find all installations
find_installations() {
    local installations=()
    
    # Common installation locations
    local search_paths=(
        "${HOME}/.local/bin"
        "${HOME}/bin"
        "/usr/local/bin"
        "/usr/bin"
        "$(go env GOPATH 2>/dev/null)/bin"
    )
    
    for path in "${search_paths[@]}"; do
        if [ -n "$path" ] && [ -f "$path/$BINARY_NAME" ]; then
            installations+=("$path/$BINARY_NAME")
        fi
    done
    
    printf '%s\n' "${installations[@]}"
}

# Function to backup before removal
backup_binary() {
    local binary_path="$1"
    
    if [ -f "$binary_path" ]; then
        print_step "Backing up $binary_path..."
        mkdir -p "$BACKUP_DIR"
        local backup_file="$BACKUP_DIR/${BINARY_NAME}_$(date +%Y%m%d_%H%M%S)"
        if cp "$binary_path" "$backup_file"; then
            print_success "Backed up to: $backup_file"
            return 0
        else
            print_error "Failed to backup $binary_path"
            return 1
        fi
    fi
}

# Function to remove binary
remove_binary() {
    local binary_path="$1"
    
    if [ -f "$binary_path" ]; then
        print_step "Removing $binary_path..."
        if rm -f "$binary_path"; then
            print_success "Removed $binary_path"
            return 0
        else
            print_error "Failed to remove $binary_path"
            return 1
        fi
    else
        print_warning "$binary_path not found"
        return 1
    fi
}

# Function to clean up configuration (optional)
cleanup_config() {
    local config_dir="${HOME}/.config/gzh-manager"
    
    if [ -d "$config_dir" ]; then
        echo
        read -p "Remove configuration directory $config_dir? [y/N]: " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_step "Removing configuration directory..."
            if rm -rf "$config_dir"; then
                print_success "Removed configuration directory"
            else
                print_error "Failed to remove configuration directory"
            fi
        else
            print_info "Configuration directory preserved"
        fi
    fi
}

# Function to verify uninstallation
verify_uninstall() {
    print_step "Verifying uninstallation..."
    
    if command_exists "$BINARY_NAME"; then
        print_warning "$BINARY_NAME is still available in PATH"
        local remaining_path
        remaining_path=$(command -v "$BINARY_NAME")
        print_info "Found at: $remaining_path"
        return 1
    else
        print_success "$BINARY_NAME is no longer available"
    fi
    
    # Test git integration
    if git synclone --help >/dev/null 2>&1; then
        print_warning "'git synclone' command still works"
        return 1
    else
        print_success "'git synclone' command removed"
    fi
    
    return 0
}

# Function to show removal summary
show_summary() {
    local removed_count="$1"
    
    echo
    echo -e "${CYAN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                  Uninstall Summary                          │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"
    
    if [ "$removed_count" -gt 0 ]; then
        print_success "Removed $removed_count installation(s) of $BINARY_NAME"
    else
        print_info "No installations found to remove"
    fi
    
    if [ -d "$BACKUP_DIR" ]; then
        print_info "Backups available in: $BACKUP_DIR"
    fi
    
    echo
    print_info "To restore from backup:"
    echo "  cp $BACKUP_DIR/${BINARY_NAME}_* ~/.local/bin/${BINARY_NAME}"
    echo "  chmod +x ~/.local/bin/${BINARY_NAME}"
    echo
    print_info "To reinstall:"
    echo "  ./scripts/install-git-extensions.sh"
    echo "  # or"
    echo "  go install github.com/gizzahub/gzh-manager-go/cmd/git-synclone@latest"
}

# Main uninstall function
main() {
    echo -e "${CYAN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                Git Extensions Uninstaller                   │"
    echo "│                   gzh-manager-go                            │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"
    
    # Find all installations
    print_step "Searching for $BINARY_NAME installations..."
    readarray -t installations < <(find_installations)
    
    if [ ${#installations[@]} -eq 0 ]; then
        print_info "No installations of $BINARY_NAME found"
        exit 0
    fi
    
    echo "Found ${#installations[@]} installation(s):"
    for installation in "${installations[@]}"; do
        echo "  - $installation"
    done
    echo
    
    # Confirm removal
    if [ "$1" != "--force" ]; then
        read -p "Remove all installations? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Uninstall cancelled"
            exit 0
        fi
    fi
    
    # Remove each installation
    removed_count=0
    for installation in "${installations[@]}"; do
        if backup_binary "$installation" && remove_binary "$installation"; then
            ((removed_count++))
        fi
    done
    
    # Optional: cleanup configuration
    cleanup_config
    
    # Verify uninstallation
    verify_uninstall
    
    # Show summary
    show_summary "$removed_count"
    
    if [ "$removed_count" -gt 0 ]; then
        print_success "Uninstallation completed successfully"
        exit 0
    else
        print_error "Uninstallation failed"
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -h, --help     Show this help message"
            echo "  --force        Remove without confirmation"
            exit 0
            ;;
        --force)
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            print_info "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main uninstall
main "$@"