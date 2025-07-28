#!/bin/bash
# Git Extensions Installation Script for gzh-manager-go
# Installs git-synclone and other Git extensions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="gzh-manager-go"
BINARY_NAME="git-synclone"
INSTALL_DIR="${HOME}/.local/bin"
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

# Function to check Go installation
check_go() {
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go first."
        print_info "Visit: https://golang.org/doc/install"
        exit 1
    fi
    
    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Found Go version: $go_version"
}

# Function to check Git installation
check_git() {
    if ! command_exists git; then
        print_error "Git is not installed. Please install Git first."
        exit 1
    fi
    
    local git_version
    git_version=$(git --version | awk '{print $3}')
    print_info "Found Git version: $git_version"
}

# Function to check Make installation
check_make() {
    if ! command_exists make; then
        print_error "Make is not installed. Please install Make first."
        exit 1
    fi
}

# Function to backup existing installation
backup_existing() {
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_step "Backing up existing installation..."
        mkdir -p "$BACKUP_DIR"
        local backup_file="$BACKUP_DIR/${BINARY_NAME}_$(date +%Y%m%d_%H%M%S)"
        cp "$INSTALL_DIR/$BINARY_NAME" "$backup_file"
        print_success "Backed up to: $backup_file"
    fi
}

# Function to build the binary
build_binary() {
    print_step "Building git extensions..."
    
    # Check if we're in the project directory
    if [ ! -f "Makefile" ] || [ ! -d "cmd/git-synclone" ]; then
        print_error "Please run this script from the project root directory"
        print_info "Expected files: Makefile, cmd/git-synclone/"
        exit 1
    fi
    
    # Build using Makefile
    if ! make build-git-extensions; then
        print_error "Build failed. Please check the error messages above."
        exit 1
    fi
    
    # Verify binary was created
    if [ ! -f "$BINARY_NAME" ]; then
        print_error "Binary '$BINARY_NAME' was not created"
        exit 1
    fi
    
    print_success "Binary built successfully"
}

# Function to install binary
install_binary() {
    print_step "Installing git extensions..."
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Copy binary
    cp "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    print_success "Installed $BINARY_NAME to $INSTALL_DIR"
}

# Function to check PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        print_warning "$INSTALL_DIR is not in PATH"
        print_info "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo -e "${YELLOW}  export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
        print_info "Then reload your shell or run: source ~/.bashrc (or ~/.zshrc)"
        return 1
    fi
    return 0
}

# Function to test installation
test_installation() {
    print_step "Testing installation..."
    
    # Test if binary can be found
    if command_exists "$BINARY_NAME"; then
        print_success "$BINARY_NAME is available in PATH"
    else
        print_warning "$BINARY_NAME is not in PATH"
        return 1
    fi
    
    # Test git integration
    if git synclone --help >/dev/null 2>&1; then
        print_success "Git integration working: 'git synclone' available"
    else
        print_warning "Git integration issue: 'git synclone' not working"
        print_info "Make sure $INSTALL_DIR is in your PATH"
        return 1
    fi
    
    return 0
}

# Function to show usage information
show_usage() {
    print_success "Installation completed!"
    echo
    print_info "Usage examples:"
    echo "  git synclone --help                    # Show help"
    echo "  git synclone github -o myorg           # Clone GitHub organization"
    echo "  git synclone gitlab -g mygroup         # Clone GitLab group"
    echo "  git synclone gitea -o myorg            # Clone Gitea organization"
    echo
    print_info "Configuration:"
    echo "  ~/.config/gzh-manager/synclone.yaml   # Configuration file"
    echo
    print_info "For more information, visit:"
    echo "  https://github.com/gizzahub/gzh-manager-go"
}

# Function to clean up
cleanup() {
    if [ -f "$BINARY_NAME" ]; then
        rm -f "$BINARY_NAME"
    fi
}

# Main installation function
main() {
    echo -e "${CYAN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                Git Extensions Installer                     │"
    echo "│                   gzh-manager-go                            │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"
    
    # Trap to cleanup on exit
    trap cleanup EXIT
    
    # Check prerequisites
    print_step "Checking prerequisites..."
    check_go
    check_git
    check_make
    print_success "All prerequisites satisfied"
    
    # Backup existing installation
    backup_existing
    
    # Build and install
    build_binary
    install_binary
    
    # Check PATH and test installation
    path_ok=true
    if ! check_path; then
        path_ok=false
    fi
    
    if $path_ok && test_installation; then
        show_usage
        exit 0
    else
        print_warning "Installation completed but with issues"
        print_info "Please check the PATH configuration above"
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
            echo "  --install-dir  Custom installation directory (default: ~/.local/bin)"
            exit 0
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            print_info "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main installation
main "$@"