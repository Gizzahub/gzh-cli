#!/bin/bash

# Commit 1: Infrastructure improvements
echo "Creating infrastructure improvements commit..."
git add .golangci.yml
git commit -m "feat(claude): add comprehensive linting config

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Commit 2: Error handling and logging
echo "Creating error handling commit..."
git add internal/errors/ internal/logger/
git commit -m "feat(claude): add error recovery and structured logging

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Commit 3: Test enhancements
echo "Creating test enhancements commit..."
git add test/integration/bulk_clone_modern_test.go test/integration/run_all_tests.sh test/integration/run_modern_tests.sh
git commit -m "test(claude): add modern integration test runners

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Commit 4: Configuration updates
echo "Creating configuration updates commit..."
git add .claude/settings.local.json Makefile README.md
git commit -m "chore(claude): update config and documentation

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Commit 5: Test file updates
echo "Creating test file updates commit..."
git add cmd/always-latest/*_test.go cmd/gen-config/gen_config_test.go cmd/net-env/*_test.go pkg/legacy/errors_test.go test/integration/net-env/net_env_integration_test.go
git commit -m "test(claude): update test files with improvements

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Commit 6: Bulk clone updates
echo "Creating bulk clone updates commit..."
git add cmd/bulk-clone/bulk_clone_github.go cmd/doctor/doctor.go examples/bulk-clone.home.yaml examples/bulk-clone.work.yaml
git commit -m "feat(claude): update bulk clone and doctor commands

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Remove obsolete file
echo "Removing obsolete file..."
git rm .golang-ci.yml
git commit -m "chore(claude): remove obsolete golangci config

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

echo "All commits completed successfully!"
