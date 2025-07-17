#!/bin/bash

# Add all changes
git add -A

# Create comprehensive commit
git commit -m "feat(claude): comprehensive infrastructure improvements

- Add .golangci.yml with comprehensive linting rules
- Add error recovery system with circuit breaker patterns
- Add structured logging with JSON output support
- Add modern integration test runners with timeout handling
- Update configuration files and documentation
- Improve test coverage and reliability
- Remove obsolete .golang-ci.yml file

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

echo "Comprehensive commit completed successfully!"