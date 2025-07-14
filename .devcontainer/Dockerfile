# Multi-stage development container for GZH Manager Go
# Supports Go, Node.js, Python, Docker, and all development tooling

FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.24.0
ENV NODE_VERSION=20
ENV PYTHON_VERSION=3.12
ENV CGO_ENABLED=1
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV DOCKER_BUILDKIT=1
ENV COMPOSE_DOCKER_CLI_BUILD=1

# Update system packages
RUN apt-get update && apt-get install -y \
    # Build essentials
    build-essential \
    pkg-config \
    # Git and version control
    git \
    git-lfs \
    # Network tools
    curl \
    wget \
    # Text processing
    jq \
    yq \
    # Archive tools
    zip \
    unzip \
    tar \
    # Development libraries
    libssl-dev \
    libffi-dev \
    # Python development
    python3-dev \
    python3-pip \
    python3-venv \
    # Node.js development (node-gyp dependencies)
    libnode-dev \
    # Documentation
    man-db \
    manpages-dev \
    # Shell enhancements
    zsh \
    bash-completion \
    # Process management
    htop \
    # Security tools
    gnupg2 \
    ca-certificates \
    # Clean up
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget -O go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

# Install Node.js and npm
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs

# Install Python and pip
RUN apt-get update && apt-get install -y \
    python${PYTHON_VERSION} \
    python${PYTHON_VERSION}-dev \
    python${PYTHON_VERSION}-venv \
    && ln -sf /usr/bin/python${PYTHON_VERSION} /usr/bin/python3 \
    && curl -sS https://bootstrap.pypa.io/get-pip.py | python3

# Set up PATH for Go
ENV PATH=/usr/local/go/bin:$PATH
ENV GOPATH=/go
ENV GOROOT=/usr/local/go
ENV PATH=$GOPATH/bin:$PATH

# Create Go workspace
RUN mkdir -p $GOPATH/src $GOPATH/bin $GOPATH/pkg

# Install Go development tools
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4 \
    && go install mvdan.cc/gofumpt@latest \
    && go install github.com/daixiang0/gci@v0.13.5 \
    && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest \
    && go install go.uber.org/mock/mockgen@latest \
    && go install github.com/goreleaser/goreleaser@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install Node.js development tools globally
RUN npm install -g \
    typescript \
    ts-node \
    eslint \
    prettier \
    node-gyp \
    yarn \
    pnpm

# Install Python development tools
RUN pip3 install --no-cache-dir \
    black \
    isort \
    pylint \
    flake8 \
    mypy \
    pytest \
    setuptools \
    wheel \
    build \
    twine \
    pre-commit

# Install pre-commit
RUN pip3 install pre-commit

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh

# Install Docker CLI and Docker Compose
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y docker-ce-cli docker-compose-plugin

# Install additional development tools
RUN apt-get update && apt-get install -y \
    # Make and build tools
    make \
    cmake \
    # Database clients
    postgresql-client \
    mysql-client \
    redis-tools \
    # Network debugging
    netcat \
    telnet \
    nmap \
    # File operations
    rsync \
    tree \
    # Text processing
    vim \
    nano \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create directories for development
RUN mkdir -p /workspace \
    && mkdir -p /home/vscode/.config \
    && mkdir -p /home/vscode/.local/bin

# Set up vscode user
RUN useradd -m -s /bin/zsh vscode \
    && usermod -aG docker vscode \
    && echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install Oh My Zsh for vscode user
USER vscode
RUN sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

# Configure zsh with useful plugins
RUN sed -i 's/plugins=(git)/plugins=(git golang docker docker-compose npm yarn python pip)/g' ~/.zshrc \
    && echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.zshrc \
    && echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc \
    && echo 'alias ll="ls -la"' >> ~/.zshrc \
    && echo 'alias la="ls -A"' >> ~/.zshrc \
    && echo 'alias l="ls -CF"' >> ~/.zshrc \
    && echo 'alias gz="./gz"' >> ~/.zshrc \
    && echo 'alias make-build="make build"' >> ~/.zshrc \
    && echo 'alias make-test="make test"' >> ~/.zshrc \
    && echo 'alias make-lint="make lint"' >> ~/.zshrc \
    && echo 'alias make-fmt="make fmt"' >> ~/.zshrc

# Switch back to root for final setup
USER root

# Set up Go environment for vscode user
RUN mkdir -p /home/vscode/go/src /home/vscode/go/bin /home/vscode/go/pkg \
    && chown -R vscode:vscode /home/vscode/go

# Configure Git (will be overridden by bind mounts)
USER vscode
RUN git config --global init.defaultBranch main \
    && git config --global pull.rebase false \
    && git config --global user.name "Development Container" \
    && git config --global user.email "dev@container.local"

# Set final working directory
WORKDIR /workspace

# Expose common development ports
EXPOSE 8080 3000 9090 9093 6060

# Set default shell
SHELL ["/bin/zsh", "-c"]

# Health check for development container
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD go version && node --version && python3 --version || exit 1

# Default command
CMD ["/bin/zsh"]