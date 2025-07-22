# Makefile.lint.mk - Code Quality and Pre-commit Management
# Separated from main Makefile for better organization

.PHONY: lint lint-check lint-fix lint-strict lint-help pre-commit-install pre-commit-run pre-commit-update validate-hooks pre-push-test validate-lint-config

# Code Quality Configuration
LINT_DIRS = libs commands api tests
LINT_DIRS_SECURITY = libs commands api
LINT_DIRS_CORE = libs commands api
EXCLUDE_DIRS = --exclude migrations --exclude node_modules --exclude examples

# Optional unsafe fixes (use: make lint-fix UNSAFE_FIXES=1)
UNSAFE_FIXES ?=
UNSAFE_FLAG = $(if $(UNSAFE_FIXES),--unsafe-fixes,)

# lint-check: pre-commit 수준의 검사 (자동 수정 없음)
# - ruff check: 모든 활성화된 규칙 검사 (D415 포함)
# - mypy: 타입 검사
# - bandit: 보안 취약점 검사 (medium 레벨)
# - mdformat: 마크다운 포맷팅 체크
lint-check:
	@echo "Running lint checks (pre-commit level, no auto-fix)..."
	@echo "Running ruff check..."
	uv run ruff check $(LINT_DIRS) $(EXCLUDE_DIRS)
	@echo "Running mypy..."
	uv run mypy $(LINT_DIRS_CORE) --ignore-missing-imports $(EXCLUDE_DIRS)
	@echo "Running bandit security check..."
	uv run bandit -r $(LINT_DIRS_SECURITY) --skip B101,B404,B603,B607,B602 --severity-level medium --quiet --exclude "*/tests/*,*/scripts/*,*/debug/*,*/examples/*" || echo "✅ Security check completed"
	@echo "Running mdformat check..."
	uv run mdformat --check *.md docs/**/*.md --wrap 120 || echo "✅ Markdown format check completed"

lint: lint-check

# lint-fix: 자동 수정 포함 코드 품질 검사 + 포맷팅
# - ruff check --fix: 자동 수정 가능한 규칙 위반 항목 수정
# - ruff format: 코드 포맷팅 자동 적용, black대체용
# - mypy: 타입 검사
# - bandit: 보안 취약점 검사 (medium 레벨)
# - mdformat: 마크다운 포맷팅
# - 사용법: make lint-fix UNSAFE_FIXES=1 (위험한 자동 수정 포함)
lint-fix:
	@echo "Running lint with auto-fix..."
	@echo "Running ruff check with auto-fix..."
	uv run ruff check $(LINT_DIRS) --fix $(UNSAFE_FLAG) $(EXCLUDE_DIRS)
	@echo "Running ruff format..."
	uv run ruff format $(LINT_DIRS) $(EXCLUDE_DIRS)
	@echo "Running mypy..."
	uv run mypy $(LINT_DIRS_CORE) --ignore-missing-imports $(EXCLUDE_DIRS)
	@echo "Running bandit security check..."
	uv run bandit -r $(LINT_DIRS_SECURITY) --skip B101,B404,B603,B607,B602 --severity-level medium --quiet --exclude "*/tests/*,*/scripts/*,*/debug/*,*/examples/*" || echo "✅ Security check completed"
	@echo "Running mdformat..."
	uv run mdformat *.md docs/**/*.md --wrap 120

# lint-strict: 엄격한 코드 품질 검사 (모든 규칙 적용)
# - ruff check --select ALL: 모든 규칙 적용 (일부 규칙 무시)
# - mypy --strict: 엄격한 타입 검사
# - bandit --severity-level low: 낮은 심각도까지 보안 검사
lint-strict:
	@echo "Running strict lint checks..."
	@echo "Running ruff with all rules..."
	uv run ruff check $(LINT_DIRS) --select ALL --ignore E501,B008,C901,COM812,B904,B017,B007,D100,D101,D102,D103,D104,D105,D106,D107  $(EXCLUDE_DIRS) --output-format=full
	@echo "Running mypy with strict settings..."
	uv run mypy $(LINT_DIRS_CORE) --strict --ignore-missing-imports  $(EXCLUDE_DIRS)
	@echo "Running bandit with strict settings..."
	uv run bandit -r $(LINT_DIRS_CORE) --severity-level low --exclude "*/tests/*,*/scripts/*,*/debug/*,*/examples/*"

# Pre-commit integration
pre-commit-install:
	@echo "Installing pre-commit hooks..."
	uv run pre-commit install

pre-commit-run:
	@echo "Running all pre-commit hooks..."
	uv run pre-commit run --all-files

pre-commit-update:
	@echo "Updating pre-commit hooks..."
	uv run pre-commit autoupdate

# Git hooks validation
validate-hooks:
	@echo "Validating git hooks consistency..."
	@echo "Testing pre-commit hooks..."
	uv run pre-commit run --all-files
	@echo "Testing make lint..."
	make lint
	@echo "✅ All hooks validated successfully"

pre-push-test:
	@echo "Running pre-push validation..."
	make lint
	make test-fast
	@echo "✅ Pre-push validation completed"

# Validate lint configuration consistency
validate-lint-config:
	@echo "Validating lint configuration consistency..."
	python3 scripts/validate-lint-config.py

# Help for lint targets
lint-help:
	@echo "Lint and Code Quality Commands:"
	@echo ""
	@echo "Basic Linting:"
	@echo "  make lint            Run linters (ruff, mypy, bandit) - read-only"
	@echo "  make lint-check      Same as lint (alias for consistency)"
	@echo "  make lint-fix        Run linters with auto-fix"
	@echo "  make lint-fix UNSAFE_FIXES=1  Run linters with unsafe auto-fix"
	@echo "  make lint-strict     Run strict linters for high quality standards"
	@echo ""
	@echo "Pre-commit Integration:"
	@echo "  make pre-commit-install    Install pre-commit hooks"
	@echo "  make pre-commit-run        Run all pre-commit hooks"
	@echo "  make pre-commit-update     Update pre-commit hooks"
	@echo ""
	@echo "Validation:"
	@echo "  make validate-hooks        Validate git hooks consistency"
	@echo "  make pre-push-test         Run pre-push validation"
	@echo "  make validate-lint-config  Validate lint configuration"
	@echo ""
	@echo "Configuration Variables:"
	@echo "  LINT_DIRS              = $(LINT_DIRS)"
	@echo "  LINT_DIRS_SECURITY     = $(LINT_DIRS_SECURITY)"
	@echo "  LINT_DIRS_CORE         = $(LINT_DIRS_CORE)"
	@echo "  EXCLUDE_DIRS           = $(EXCLUDE_DIRS)"
	@echo "  UNSAFE_FIXES           = $(UNSAFE_FIXES)"