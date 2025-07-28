#!/bin/bash
# Online Installation Script for gzh Git Extensions
# This script downloads and installs git-synclone from GitHub releases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REPO="gizzahub/gzh-manager-go"
BINARY_NAME="git-synclone"
INSTALL_DIR="${HOME}/.local/bin"
GITHUB_API="https://api.github.com/repos"
GITHUB_RELEASES="https://github.com"

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

# Function to detect OS and architecture
detect_platform() {
    local os arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          os="unknown" ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        armv7l)       arch="arm" ;;
        i386|i686)    arch="386" ;;
        *)            arch="unknown" ;;
    esac
    
    if [ "$os" = "unknown" ] || [ "$arch" = "unknown" ]; then
        print_error "Unsupported platform: $(uname -s) $(uname -m)"
        print_info "Supported platforms:"
        print_info "  - Linux (amd64, arm64, arm, 386)"
        print_info "  - macOS (amd64, arm64)"
        print_info "  - Windows (amd64, 386)"
        exit 1
    fi
    
    echo "${os}_${arch}"
}

# Function to get latest release version
get_latest_version() {
    if command_exists curl; then
        curl -s "${GITHUB_API}/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command_exists wget; then
        wget -qO- "${GITHUB_API}/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        print_error "Neither curl nor wget is available"
        print_info "Please install curl or wget to continue"
        exit 1
    fi
}

# Function to download file
download_file() {
    local url="$1"
    local output="$2"
    
    print_step "Downloading from: $url"
    
    if command_exists curl; then
        curl -L -o "$output" "$url"
    elif command_exists wget; then
        wget -O "$output" "$url"
    else
        print_error "Neither curl nor wget is available"
        exit 1
    fi
}

# Function to install from source (fallback)
install_from_source() {
    print_step "Installing from source..."
    
    # Check if Go is installed
    if ! command_exists go; then
        print_error "Go is not installed and no pre-built binary available"
        print_info "Please install Go from: https://golang.org/doc/install"
        print_info "Then run: go install github.com/${REPO}/cmd/git-synclone@latest"
        exit 1
    fi
    
    # Install using go install
    print_step "Installing using 'go install'..."
    if go install "github.com/${REPO}/cmd/git-synclone@latest"; then
        print_success "Installed using 'go install'"
        
        # Check if GOPATH/bin is in PATH
        local gopath_bin
        gopath_bin="$(go env GOPATH)/bin"
        if [[ ":$PATH:" != *":$gopath_bin:"* ]]; then
            print_warning "$gopath_bin is not in PATH"
            print_info "Add the following to your shell profile:"
            echo "  export PATH=\"\$PATH:$gopath_bin\""
        fi
        
        return 0
    else
        print_error "Failed to install from source"
        exit 1
    fi
}

# Function to install binary
install_binary() {
    local version="$1"
    local platform="$2"
    
    print_step "Installing git-synclone version $version for $platform..."
    
    # Construct download URL
    local filename="${BINARY_NAME}_${version}_${platform}"
    if [[ "$platform" == *"windows"* ]]; then
        filename="${filename}.exe"
    fi
    
    local download_url="${GITHUB_RELEASES}/${REPO}/releases/download/${version}/${filename}"
    local temp_file="/tmp/${filename}"
    
    # Download binary
    if ! download_file "$download_url" "$temp_file"; then
        print_warning "Failed to download pre-built binary"
        print_info "Falling back to source installation..."
        install_from_source
        return
    fi
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Install binary
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    cp "$temp_file" "$binary_path"
    chmod +x "$binary_path"
    
    # Cleanup
    rm -f "$temp_file"
    
    print_success "Installed $BINARY_NAME to $INSTALL_DIR"
}

# Function to verify installation
verify_installation() {
    print_step "Verifying installation..."
    
    # Check if binary is accessible
    if command_exists "$BINARY_NAME"; then
        print_success "$BINARY_NAME is available in PATH"
        
        # Check version
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        print_info "Installed version: $version"
        
        # Test git integration
        if git synclone --help >/dev/null 2>&1; then
            print_success "Git integration working: 'git synclone' available"
        else
            print_warning "Git integration issue"
        fi
    else
        print_warning "$BINARY_NAME not found in PATH"
        print_info "Binary installed at: $INSTALL_DIR/$BINARY_NAME"
        
        # Check PATH
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            print_warning "$INSTALL_DIR is not in PATH"
            print_info "Add the following to your shell profile:"
            echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
            print_info "Then reload your shell or run: source ~/.bashrc"
        fi
    fi
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
    echo "  git synclone doctor                    # Check installation"
    echo
    print_info "Configuration:"
    echo "  ~/.config/gzh-manager/synclone.yaml   # Configuration file"
    echo
    print_info "For more information:"
    echo "  https://github.com/${REPO}"
}

# Main installation function
main() {
    echo -e "${CYAN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│           Git Extensions Online Installer                   │"
    echo "│                 gzh-manager-go                              │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"
    
    # Check prerequisites
    print_step "Checking prerequisites..."
    
    if ! command_exists git; then
        print_error "Git is not installed"
        print_info "Please install Git first: https://git-scm.com/downloads"
        exit 1
    fi
    
    print_success "Git is installed"
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    print_info "Detected platform: $platform"
    
    # Get latest version
    print_step "Fetching latest release..."
    local version
    version=$(get_latest_version)
    
    if [ -z "$version" ]; then
        print_warning "Could not fetch latest version from GitHub"
        print_info "Falling back to source installation..."
        install_from_source
    else
        print_info "Latest version: $version"
        install_binary "$version" "$platform"
    fi
    
    # Verify and show usage
    verify_installation
    show_usage
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -h, --help         Show this help message"
            echo "  --install-dir DIR  Custom installation directory"
            echo "  --version VERSION  Install specific version"
            exit 0
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
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