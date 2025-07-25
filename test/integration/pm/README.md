# Package Manager Integration Tests

This directory contains Docker-based integration tests for the `gz pm` (package manager) functionality. The tests verify package manager installation, configuration, version coordination, and migration features across different operating systems.

## Overview

The integration tests use Docker containers to provide isolated environments for testing package manager operations without affecting the host system. This approach allows us to:

- Test on multiple operating systems (Ubuntu, Alpine, Fedora)
- Verify package manager bootstrap functionality
- Test version coordination (e.g., Node.js/npm, Ruby/gem)
- Validate package migration between versions
- Ensure cross-platform compatibility

## Test Structure

```
test/integration/pm/
├── Dockerfile.ubuntu    # Ubuntu 22.04 test environment
├── Dockerfile.alpine    # Alpine Linux test environment
├── Dockerfile.fedora    # Fedora 39 test environment
├── docker-compose.yml   # Container orchestration
├── pm_integration_test.go # Main test file
├── README.md           # This file
└── helpers/            # Test utilities (if needed)

test/fixtures/pm/
├── global.yml          # Global PM configuration
├── brew.yml           # Homebrew test packages
├── asdf.yml           # ASDF plugins and versions
├── npm.yml            # NPM packages
└── pip.yml            # Python packages
```

## Prerequisites

- Docker Engine 20.10+
- Go 1.21+
- At least 4GB RAM for containers
- Internet connection for package downloads

## Running Tests

### Run All PM Integration Tests

```bash
# From project root
go test ./test/integration/pm/... -v
```

### Run Specific OS Tests

```bash
# Test Ubuntu only
go test ./test/integration/pm -v -run "TestPMIntegration/Ubuntu"

# Test Alpine only
go test ./test/integration/pm -v -run "TestPMIntegration/Alpine"

# Test Fedora only
go test ./test/integration/pm -v -run "TestPMIntegration/Fedora"
```

### Run with Docker Compose

```bash
# Build gz binary first
cd test/integration/pm
go build -o gz ../../../

# Start all test containers
docker-compose up -d

# Run tests in a specific container
docker exec gz-pm-ubuntu-test bash -l -c "gz pm bootstrap --check"
docker exec gz-pm-ubuntu-test bash -l -c "gz pm install --all"

# Clean up
docker-compose down
```

## Test Scenarios

### 1. Bootstrap Tests
- Checks which package managers need installation
- Verifies package manager detection
- Tests installation of missing managers

### 2. Package Installation Tests
- **Homebrew**: Install formulae on Linux
- **ASDF**: Install plugins and language versions
- **NPM**: Install global Node.js packages
- **pip**: Install Python packages

### 3. Version Coordination Tests
- Node.js/npm version synchronization
- Ruby/gem migration when switching versions
- Python/pip virtual environment handling

### 4. Migration Tests
- Migrate packages between language versions
- Test intelligent migration strategies
- Verify package compatibility

### 5. Export/Import Tests
- Export current package configurations
- Import configurations on new systems
- Validate configuration integrity

## Container Environments

### Ubuntu 22.04
- Full support for all package managers
- Includes: Homebrew, nvm, rbenv, pyenv, asdf, SDKMAN
- Used for comprehensive testing

### Alpine Linux
- Lightweight environment
- Limited package manager support
- Tests: asdf, native package management

### Fedora 39
- RPM-based system testing
- Full package manager support
- Tests DNF integration

## Package Managers Tested

1. **System Package Managers**
   - Homebrew (Linux)
   - APT (Ubuntu)
   - DNF (Fedora)
   - APK (Alpine)

2. **Version Managers**
   - asdf (all languages)
   - nvm (Node.js)
   - rbenv (Ruby)
   - pyenv (Python)
   - SDKMAN (JVM)

3. **Language Package Managers**
   - npm/yarn/pnpm (Node.js)
   - pip (Python)
   - gem (Ruby)

## Test Configuration

The test configurations in `test/fixtures/pm/` are minimal to ensure fast test execution:

- `global.yml`: Enables essential managers with test settings
- `brew.yml`: Small set of formulae (jq, tree, htop, ripgrep)
- `asdf.yml`: Node.js and Python versions
- `npm.yml`: Essential packages (yarn, typescript, nodemon)
- `pip.yml`: Basic packages (black, pytest, httpie)

## Troubleshooting

### Docker Build Failures

```bash
# Clean Docker cache
docker system prune -f

# Rebuild without cache
docker-compose build --no-cache
```

### Test Timeouts

```bash
# Increase timeout
go test ./test/integration/pm -v -timeout 30m
```

### Package Manager Installation Issues

```bash
# Check container logs
docker logs gz-pm-ubuntu-test

# Interactive debugging
docker exec -it gz-pm-ubuntu-test bash
```

## CI/CD Integration

```yaml
# Example GitHub Actions workflow
name: PM Integration Tests
on: [push, pull_request]

jobs:
  pm-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21

      - name: Run PM Integration Tests
        run: |
          go test ./test/integration/pm/... -v -timeout 20m
```

## Future Enhancements

- [ ] Windows container support (when available)
- [ ] macOS testing (using macOS runners)
- [ ] Performance benchmarking
- [ ] Package conflict resolution testing
- [ ] Multi-version testing (e.g., Python 2 vs 3)
- [ ] Cross-platform configuration validation
- [ ] Package security scanning integration
