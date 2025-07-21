#!/bin/bash

# Migration script: bulk-clone.yaml → gzh.yaml
# This script helps automate the basic migration process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BULK_CLONE_FILE="bulk-clone.yaml"
GZH_YAML_FILE="gzh.yaml"
BACKUP_SUFFIX=".backup.$(date +%Y%m%d_%H%M%S)"

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  gzh.yaml Migration Tool${NC}"
    echo -e "${BLUE}================================${NC}"
    echo
}

print_step() {
    echo -e "${GREEN}▶${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

check_file_exists() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        print_error "File not found: $file"
        return 1
    fi
    return 0
}

backup_file() {
    local file="$1"
    if [[ -f "$file" ]]; then
        local backup="${file}${BACKUP_SUFFIX}"
        cp "$file" "$backup"
        print_success "Backup created: $backup"
    fi
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -i, --input FILE     Input bulk-clone.yaml file (default: bulk-clone.yaml)"
    echo "  -o, --output FILE    Output gzh.yaml file (default: gzh.yaml)"
    echo "  -h, --help          Show this help message"
    echo "  --dry-run           Preview migration without creating files"
    echo "  --backup           Create backup of existing files"
    echo
    echo "Examples:"
    echo "  $0                                    # Migrate bulk-clone.yaml to gzh.yaml"
    echo "  $0 -i old.yaml -o new.yaml          # Custom input/output files"
    echo "  $0 --dry-run                         # Preview migration"
    echo "  $0 --backup                         # Create backups before migration"
}

create_basic_gzh_yaml() {
    local output="$1"
    local dry_run="$2"

    local content=$(cat << 'EOF'
# Migrated from bulk-clone.yaml
# Please review and update as needed

version: "1.0.0"
default_provider: github

providers:
  github:
    # IMPORTANT: Set this environment variable
    # export GITHUB_TOKEN="your_github_personal_access_token"
    token: "${GITHUB_TOKEN}"
    orgs:
      # Add your GitHub organizations here
      # Example:
      # - name: "your-org-name"
      #   visibility: "all"
      #   clone_dir: "${HOME}/repos/github"
      #   exclude: ["test-.*", ".*-archive"]
      #   strategy: "reset"

# Migration Notes:
# 1. Replace protocol-based authentication with tokens
# 2. Update repo_roots entries to providers.github.orgs
# 3. Move ignore_names to exclude patterns per organization
# 4. Set up environment variables for authentication
#
# See migration guide: docs/migration-guide-bulk-clone-to-gzh.md
EOF
)

    if [[ "$dry_run" == "true" ]]; then
        echo -e "${BLUE}Generated gzh.yaml content:${NC}"
        echo "$content"
    else
        echo "$content" > "$output"
        print_success "Created basic gzh.yaml template: $output"
    fi
}

extract_organizations() {
    local input="$1"

    print_step "Analyzing existing configuration..."

    if command -v yq >/dev/null 2>&1; then
        print_step "Using yq to extract organizations..."

        # Extract organizations using yq
        local orgs=$(yq '.repo_roots[].org_name' "$input" 2>/dev/null | grep -v null | sort -u)

        if [[ -n "$orgs" ]]; then
            echo -e "${BLUE}Found organizations:${NC}"
            echo "$orgs" | while read -r org; do
                echo "  - $org"
            done
            echo
        fi

        # Extract ignore patterns
        local ignore_patterns=$(yq '.ignore_names[]' "$input" 2>/dev/null | grep -v null)

        if [[ -n "$ignore_patterns" ]]; then
            echo -e "${BLUE}Found ignore patterns:${NC}"
            echo "$ignore_patterns" | while read -r pattern; do
                echo "  - $pattern"
            done
            echo
        fi
    else
        print_warning "yq not found. Install yq for better migration support."
        print_warning "Falling back to basic template generation."
    fi
}

check_environment() {
    print_step "Checking environment..."

    # Check for required tools
    local missing_tools=()

    if ! command -v yq >/dev/null 2>&1; then
        missing_tools+=("yq")
    fi

    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        print_warning "Optional tools not found: ${missing_tools[*]}"
        print_warning "For better migration experience, install:"
        for tool in "${missing_tools[@]}"; do
            echo "  - $tool"
        done
        echo
    fi

    # Check for existing tokens
    if [[ -z "$GITHUB_TOKEN" ]]; then
        print_warning "GITHUB_TOKEN environment variable not set"
        echo "  You'll need to set this for authentication:"
        echo "  export GITHUB_TOKEN=\"your_github_personal_access_token\""
        echo
    fi
}

main() {
    local input_file="$BULK_CLONE_FILE"
    local output_file="$GZH_YAML_FILE"
    local dry_run="false"
    local create_backup="false"

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -i|--input)
                input_file="$2"
                shift 2
                ;;
            -o|--output)
                output_file="$2"
                shift 2
                ;;
            --dry-run)
                dry_run="true"
                shift
                ;;
            --backup)
                create_backup="true"
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    print_header

    # Check if input file exists
    if ! check_file_exists "$input_file"; then
        print_error "Cannot proceed without input file"
        echo
        echo "Please ensure your bulk-clone.yaml file exists, or specify a different input file:"
        echo "  $0 -i /path/to/your/bulk-clone.yaml"
        exit 1
    fi

    # Create backup if requested
    if [[ "$create_backup" == "true" ]]; then
        print_step "Creating backups..."
        backup_file "$input_file"
        [[ -f "$output_file" ]] && backup_file "$output_file"
    fi

    # Check environment
    check_environment

    # Extract information from existing config
    extract_organizations "$input_file"

    # Generate new configuration
    print_step "Generating gzh.yaml configuration..."
    create_basic_gzh_yaml "$output_file" "$dry_run"

    # Provide next steps
    echo
    print_success "Migration preparation complete!"
    echo
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Review the generated gzh.yaml file"
    echo "2. Update organization configurations based on your bulk-clone.yaml"
    echo "3. Set up environment variables for authentication:"
    echo "   export GITHUB_TOKEN=\"your_github_token\""
    echo "4. Validate the configuration:"
    echo "   gzh config validate"
    echo "5. Test with dry run:"
    echo "   gzh bulk-clone --dry-run --use-gzh-config"
    echo
    echo -e "${BLUE}For detailed migration instructions, see:${NC}"
    echo "  docs/migration-guide-bulk-clone-to-gzh.md"
    echo

    if [[ "$dry_run" != "true" ]]; then
        print_success "Configuration template created: $output_file"
    fi
}

# Run main function with all arguments
main "$@"
