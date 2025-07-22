# Makefile.docker - Docker targets for gzh-manager-go
# Container build, optimization, and deployment

# ==============================================================================
# Docker Configuration
# ==============================================================================

# Docker configuration
DOCKER_REGISTRY ?= gizzahub
DOCKER_IMAGE_NAME ?= $(projectname)
DOCKER_TAG ?= $(VERSION)
DOCKER_FULL_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)
DOCKER_LATEST := $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest

# ==============================================================================
# Docker Build Targets
# ==============================================================================

.PHONY: docker-build docker-build-dev docker-build-prod docker-tag docker-push
.PHONY: docker-run docker-run-dev docker-stop docker-logs docker-clean

docker-build: ## build Docker image
	@echo "$(CYAN)Building Docker image $(DOCKER_FULL_IMAGE)...$(RESET)"
	@docker build -t $(DOCKER_FULL_IMAGE) .
	@docker tag $(DOCKER_FULL_IMAGE) $(DOCKER_LATEST)
	@echo "$(GREEN)✅ Docker image built successfully$(RESET)"

docker-build-dev: ## build Docker image for development
	@echo "$(CYAN)Building development Docker image...$(RESET)"
	@docker build --target development -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):dev .
	@echo "$(GREEN)✅ Development Docker image built$(RESET)"

docker-build-prod: ## build Docker image for production
	@echo "$(CYAN)Building production Docker image...$(RESET)"
	@docker build --target production -t $(DOCKER_FULL_IMAGE) .
	@docker tag $(DOCKER_FULL_IMAGE) $(DOCKER_LATEST)
	@echo "$(GREEN)✅ Production Docker image built$(RESET)"

docker-tag: ## tag Docker image with version
	@echo "$(CYAN)Tagging Docker image...$(RESET)"
	@docker tag $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest $(DOCKER_FULL_IMAGE)
	@echo "$(GREEN)✅ Docker image tagged: $(DOCKER_FULL_IMAGE)$(RESET)"

docker-push: docker-build ## push Docker image to registry
	@echo "$(CYAN)Pushing Docker image to registry...$(RESET)"
	@docker push $(DOCKER_FULL_IMAGE)
	@docker push $(DOCKER_LATEST)
	@echo "$(GREEN)✅ Docker image pushed to registry$(RESET)"

# ==============================================================================
# Docker Runtime Targets
# ==============================================================================

docker-run: ## run Docker container
	@echo "$(CYAN)Running Docker container...$(RESET)"
	@docker run -d --name $(executablename) -p 8080:8080 $(DOCKER_FULL_IMAGE)
	@echo "$(GREEN)✅ Docker container started$(RESET)"

docker-run-dev: ## run Docker container in development mode
	@echo "$(CYAN)Running development Docker container...$(RESET)"
	@docker run -d --name $(executablename)-dev \
		-v $(PWD):/app \
		-p 8080:8080 \
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):dev
	@echo "$(GREEN)✅ Development Docker container started$(RESET)"

docker-stop: ## stop and remove Docker container
	@echo "$(CYAN)Stopping Docker container...$(RESET)"
	@docker stop $(executablename) 2>/dev/null || true
	@docker rm $(executablename) 2>/dev/null || true
	@docker stop $(executablename)-dev 2>/dev/null || true
	@docker rm $(executablename)-dev 2>/dev/null || true
	@echo "$(GREEN)✅ Docker container stopped$(RESET)"

docker-logs: ## show Docker container logs
	@echo "$(CYAN)Docker container logs:$(RESET)"
	@docker logs $(executablename) 2>/dev/null || docker logs $(executablename)-dev 2>/dev/null || echo "$(YELLOW)No containers running$(RESET)"

docker-exec: ## execute shell in running container
	@echo "$(CYAN)Executing shell in container...$(RESET)"
	@docker exec -it $(executablename) /bin/sh 2>/dev/null || docker exec -it $(executablename)-dev /bin/sh 2>/dev/null || echo "$(RED)No containers running$(RESET)"

# ==============================================================================
# Docker Optimization and Analysis
# ==============================================================================

.PHONY: docker-optimize docker-scan docker-size docker-history docker-lint

docker-optimize: ## optimize Docker image size
	@echo "$(CYAN)Analyzing Docker image for optimization...$(RESET)"
	@echo "$(YELLOW)Current image size:$(RESET)"
	@docker images $(DOCKER_FULL_IMAGE) --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}"
	@echo ""
	@echo "$(YELLOW)Layer breakdown:$(RESET)"
	@docker history $(DOCKER_FULL_IMAGE) --format "table {{.CreatedBy}}\t{{.Size}}" | head -10
	@echo ""
	@echo "$(BLUE)💡 Optimization suggestions:$(RESET)"
	@echo "  • Use multi-stage builds to reduce final image size"
	@echo "  • Remove unnecessary packages and files"
	@echo "  • Use .dockerignore to exclude build artifacts"
	@echo "  • Consider using Alpine Linux base images"
	@echo "  • Combine RUN commands to reduce layers"

docker-scan: ## scan Docker image for vulnerabilities
	@echo "$(CYAN)Scanning Docker image for vulnerabilities...$(RESET)"
	@if command -v trivy >/dev/null 2>&1; then \
		trivy image $(DOCKER_FULL_IMAGE); \
	elif command -v docker >/dev/null 2>&1 && docker version | grep -q "Engine"; then \
		echo "$(YELLOW)Using docker scan (requires Docker Desktop)...$(RESET)"; \
		docker scan $(DOCKER_FULL_IMAGE) || echo "$(YELLOW)Docker scan not available$(RESET)"; \
	else \
		echo "$(YELLOW)Install trivy for vulnerability scanning: https://github.com/aquasecurity/trivy$(RESET)"; \
	fi

docker-size: ## analyze Docker image size
	@echo "$(CYAN)Docker image size analysis:$(RESET)"
	@echo ""
	@echo "$(YELLOW)All $(DOCKER_IMAGE_NAME) images:$(RESET)"
	@docker images $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME) --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
	@echo ""
	@echo "$(YELLOW)Image layers:$(RESET)"
	@docker history $(DOCKER_FULL_IMAGE) --format "table {{.CreatedBy}}\t{{.Size}}" | head -5

docker-history: ## show Docker image history
	@echo "$(CYAN)Docker image build history:$(RESET)"
	@docker history $(DOCKER_FULL_IMAGE)

docker-lint: ## lint Dockerfile
	@echo "$(CYAN)Linting Dockerfile...$(RESET)"
	@if command -v hadolint >/dev/null 2>&1; then \
		hadolint Dockerfile; \
		echo "$(GREEN)✅ Dockerfile linting completed$(RESET)"; \
	else \
		echo "$(YELLOW)Install hadolint for Dockerfile linting: https://github.com/hadolint/hadolint$(RESET)"; \
	fi

# ==============================================================================
# Docker Cleanup Targets
# ==============================================================================

.PHONY: docker-clean docker-clean-images docker-clean-volumes docker-clean-all

docker-clean: ## clean up Docker containers and unused images
	@echo "$(CYAN)Cleaning Docker containers and images...$(RESET)"
	@docker container prune -f
	@docker image prune -f
	@echo "$(GREEN)✅ Docker cleanup completed$(RESET)"

docker-clean-images: ## remove all project Docker images
	@echo "$(CYAN)Removing all $(DOCKER_IMAGE_NAME) images...$(RESET)"
	@docker images $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME) -q | xargs -r docker rmi -f
	@echo "$(GREEN)✅ All project Docker images removed$(RESET)"

docker-clean-volumes: ## clean unused Docker volumes
	@echo "$(CYAN)Cleaning unused Docker volumes...$(RESET)"
	@docker volume prune -f
	@echo "$(GREEN)✅ Docker volumes cleaned$(RESET)"

docker-clean-all: docker-stop docker-clean docker-clean-images docker-clean-volumes ## comprehensive Docker cleanup
	@echo "$(GREEN)✅ Comprehensive Docker cleanup completed$(RESET)"

# ==============================================================================
# Docker Compose Integration
# ==============================================================================

.PHONY: docker-compose-up docker-compose-down docker-compose-logs docker-compose-build

docker-compose-up: ## start services with docker-compose
	@echo "$(CYAN)Starting services with docker-compose...$(RESET)"
	@if [ -f "docker-compose.yml" ]; then \
		docker-compose up -d; \
		echo "$(GREEN)✅ Services started$(RESET)"; \
	else \
		echo "$(YELLOW)No docker-compose.yml found$(RESET)"; \
	fi

docker-compose-down: ## stop services with docker-compose
	@echo "$(CYAN)Stopping services with docker-compose...$(RESET)"
	@if [ -f "docker-compose.yml" ]; then \
		docker-compose down; \
		echo "$(GREEN)✅ Services stopped$(RESET)"; \
	else \
		echo "$(YELLOW)No docker-compose.yml found$(RESET)"; \
	fi

docker-compose-logs: ## show docker-compose logs
	@echo "$(CYAN)Docker-compose logs:$(RESET)"
	@if [ -f "docker-compose.yml" ]; then \
		docker-compose logs -f; \
	else \
		echo "$(YELLOW)No docker-compose.yml found$(RESET)"; \
	fi

docker-compose-build: ## build services with docker-compose
	@echo "$(CYAN)Building services with docker-compose...$(RESET)"
	@if [ -f "docker-compose.yml" ]; then \
		docker-compose build; \
		echo "$(GREEN)✅ Services built$(RESET)"; \
	else \
		echo "$(YELLOW)No docker-compose.yml found$(RESET)"; \
	fi

# ==============================================================================
# Docker Information
# ==============================================================================

.PHONY: docker-info docker-status

docker-info: ## show Docker information and available targets
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         $(YELLOW)Docker Information$(CYAN)                              ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo "$(RESET)"
	@echo "$(GREEN)🐳 Docker Configuration:$(RESET)"
	@echo "  Registry:       $(YELLOW)$(DOCKER_REGISTRY)$(RESET)"
	@echo "  Image Name:     $(YELLOW)$(DOCKER_IMAGE_NAME)$(RESET)"
	@echo "  Version Tag:    $(YELLOW)$(DOCKER_TAG)$(RESET)"
	@echo "  Full Image:     $(YELLOW)$(DOCKER_FULL_IMAGE)$(RESET)"
	@echo ""
	@echo "$(GREEN)🏗️  Build Targets:$(RESET)"
	@echo "  • $(CYAN)docker-build$(RESET)        Build Docker image"
	@echo "  • $(CYAN)docker-build-dev$(RESET)    Build development image"
	@echo "  • $(CYAN)docker-build-prod$(RESET)   Build production image"
	@echo "  • $(CYAN)docker-push$(RESET)         Push image to registry"
	@echo ""
	@echo "$(GREEN)🚀 Runtime Targets:$(RESET)"
	@echo "  • $(CYAN)docker-run$(RESET)          Run container"
	@echo "  • $(CYAN)docker-run-dev$(RESET)      Run development container"
	@echo "  • $(CYAN)docker-stop$(RESET)         Stop and remove containers"
	@echo "  • $(CYAN)docker-logs$(RESET)         Show container logs"
	@echo "  • $(CYAN)docker-exec$(RESET)         Execute shell in container"
	@echo ""
	@echo "$(GREEN)🔍 Analysis Targets:$(RESET)"
	@echo "  • $(CYAN)docker-optimize$(RESET)     Analyze for optimization"
	@echo "  • $(CYAN)docker-scan$(RESET)         Scan for vulnerabilities"
	@echo "  • $(CYAN)docker-size$(RESET)         Analyze image size"
	@echo "  • $(CYAN)docker-lint$(RESET)         Lint Dockerfile"

docker-status: ## show current Docker status
	@echo "$(CYAN)Docker Status$(RESET)"
	@echo "$(BLUE)==============$(RESET)"
	@echo ""
	@echo "$(GREEN)🐳 Docker Environment:$(RESET)"
	@printf "  %-20s " "Docker Version:"; docker --version 2>/dev/null | cut -d' ' -f3 | cut -d',' -f1 || echo "$(RED)Not installed$(RESET)"
	@printf "  %-20s " "Docker Running:"; if docker info >/dev/null 2>&1; then echo "$(GREEN)Yes$(RESET)"; else echo "$(RED)No$(RESET)"; fi
	@echo ""
	@echo "$(GREEN)📦 Project Images:$(RESET)"
	@docker images $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME) --format "  {{.Repository}}:{{.Tag}}\t{{.Size}}" 2>/dev/null || echo "  $(YELLOW)No project images found$(RESET)"
	@echo ""
	@echo "$(GREEN)🔄 Running Containers:$(RESET)"
	@docker ps --filter "name=$(executablename)" --format "  {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "  $(YELLOW)No containers running$(RESET)"
