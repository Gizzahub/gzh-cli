# Task: Setup and Understand New CI/CD Workflows

## Priority: HIGH
## Estimated Time: 45 minutes

## Context
Remote branch added separated CI/CD workflows for better organization:
- `.github/workflows/coverage.yml` - Code coverage reporting
- `.github/workflows/lint.yml` - Dedicated linting workflow
- `.github/workflows/test.yml` - Test execution workflow
- `.github/workflows/goreleaser.yml` - Automated release workflow
- `.github/workflows/dependabot-auto-merge.yml` - Dependency management

## Pre-requisites
- [ ] Task 01 completed (remote changes merged)
- [ ] GitHub Actions enabled in repository
- [ ] Golang 1.23+ installed locally

## Steps

### 1. Analyze New Workflow Files
```bash
# List all workflow files
ls -la .github/workflows/

# Review each new workflow
cat .github/workflows/coverage.yml
cat .github/workflows/lint.yml
cat .github/workflows/test.yml
cat .github/workflows/goreleaser.yml
```

### 2. Setup Local Testing for Workflows

#### Install act (GitHub Actions local runner)
```bash
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | bash
```

#### Test Coverage Workflow Locally
```bash
# Run coverage workflow
act -j coverage

# Expected output: Coverage report generation
```

#### Test Lint Workflow Locally
```bash
# Run lint workflow
act -j lint

# Verify it uses new .golangci.yml configuration
```

### 3. Configure GoReleaser Locally

#### Install GoReleaser
```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser@latest
```

#### Create .goreleaser.yml (if not exists)
```yaml
# .goreleaser.yml
before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: gz
    binary: gz
    main: ./main.go
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

#### Test GoReleaser
```bash
# Dry run to test configuration
goreleaser release --snapshot --skip-publish --rm-dist
```

### 4. Update Makefile Integration

Ensure Makefile has targets for new workflows:
```makefile
# Add to Makefile if missing
.PHONY: ci-lint
ci-lint:
	@echo "Running CI lint checks..."
	golangci-lint run --config .golangci.yml

.PHONY: ci-test
ci-test:
	@echo "Running CI tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: ci-coverage
ci-coverage: ci-test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html

.PHONY: release-dry
release-dry:
	@echo "Testing release process..."
	goreleaser release --snapshot --skip-publish --rm-dist
```

### 5. Setup Branch Protection Rules

Create script to configure branch protection:
```bash
# Create .github/scripts/setup-protection.sh
#!/bin/bash
gh api repos/:owner/:repo/branches/develop/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["lint","test","coverage"]}' \
  --field enforce_admins=false \
  --field required_pull_request_reviews='{"required_approving_review_count":1}' \
  --field restrictions=null
```

### 6. Verify Dependabot Configuration

Check `.github/dependabot.yml`:
```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    reviewers:
      - "archmagece"
    labels:
      - "dependencies"
      - "go"
```

## Expected Outcomes
- [ ] All workflows can be tested locally with `act`
- [ ] GoReleaser configuration is valid
- [ ] Makefile has CI-related targets
- [ ] Branch protection rules understand (ready to apply)
- [ ] Dependabot configuration is correct

## Verification Commands
```bash
# Test all CI steps locally
make ci-lint
make ci-test
make ci-coverage

# Verify goreleaser
goreleaser check

# List GitHub workflows
gh workflow list
```

## Next Steps
- Task 03: Implement bulk-clone performance improvements
- Task 04: Add programmatic usage examples