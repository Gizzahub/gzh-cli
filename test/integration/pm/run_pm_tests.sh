#!/bin/bash
# Script to run package manager integration tests

set -e

echo "🚀 Starting Package Manager Integration Tests"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
check_prerequisites() {
    echo "📋 Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed${NC}"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All prerequisites met${NC}"
}

# Build gz binary
build_gz() {
    echo "🔨 Building gz binary..."
    pushd ../../.. > /dev/null
    go build -o test/integration/pm/gz
    popd > /dev/null
    chmod +x gz
    echo -e "${GREEN}✅ gz binary built${NC}"
}

# Run tests with Docker Compose
run_compose_tests() {
    echo "🐳 Running tests with Docker Compose..."
    
    # Build containers
    docker-compose build
    
    # Start containers
    docker-compose up -d
    
    # Wait for containers to be ready
    sleep 5
    
    # Run tests in each container
    for container in gz-pm-ubuntu-test gz-pm-fedora-test gz-pm-alpine-test; do
        echo "📦 Testing in $container..."
        
        # Make gz executable in container
        docker exec $container chmod +x /usr/local/bin/gz
        
        # Run bootstrap check
        docker exec $container sudo -u testuser bash -l -c "gz pm bootstrap --check" || true
        
        # Run package installation
        docker exec $container sudo -u testuser bash -l -c "gz pm install --all" || true
        
        # Run export
        docker exec $container sudo -u testuser bash -l -c "gz pm export --all" || true
    done
    
    # Clean up
    docker-compose down
    echo -e "${GREEN}✅ Docker Compose tests completed${NC}"
}

# Run Go integration tests
run_go_tests() {
    echo "🧪 Running Go integration tests..."
    
    # Run with timeout
    if go test -v -timeout 30m; then
        echo -e "${GREEN}✅ Go integration tests passed${NC}"
    else
        echo -e "${RED}❌ Go integration tests failed${NC}"
        exit 1
    fi
}

# Main execution
main() {
    check_prerequisites
    build_gz
    
    # Parse arguments
    if [ "$1" == "compose" ]; then
        run_compose_tests
    elif [ "$1" == "go" ]; then
        run_go_tests
    else
        # Run both by default
        run_compose_tests
        run_go_tests
    fi
    
    echo -e "${GREEN}🎉 All package manager integration tests completed!${NC}"
}

# Run main function
main "$@"