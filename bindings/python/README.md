# GZH Manager Python Client

[![PyPI version](https://badge.fury.io/py/gzhclient.svg)](https://badge.fury.io/py/gzhclient)
[![Python Support](https://img.shields.io/pypi/pyversions/gzhclient.svg)](https://pypi.org/project/gzhclient/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Python client library for [GZH Manager](https://github.com/gizzahub/gzh-manager-go), providing programmatic access to bulk repository operations, plugin management, system monitoring, and more.

## 🚀 Features

- **🔄 Bulk Repository Operations**: Clone repositories from GitHub, GitLab, Gitea, and Gogs
- **🔌 Plugin Management**: Load, execute, and manage plugins
- **📊 System Monitoring**: Collect CPU, memory, and disk metrics
- **⚡ Event System**: Subscribe to and handle system events
- **🌐 Multi-Platform Support**: Works on Linux, macOS, and Windows
- **🐍 Pythonic API**: Clean, intuitive Python interface
- **🔒 Type Safety**: Full type hints for better development experience

## 📦 Installation

### Prerequisites

- Python 3.8 or higher
- Go 1.19+ (for building the underlying library)

### Install from PyPI

```bash
pip install gzhclient
```

### Install from Source

```bash
git clone https://github.com/gizzahub/gzh-manager-go.git
cd gzh-manager-go/bindings/python
pip install -e .
```

## 🏃‍♂️ Quick Start

### Basic Usage

```python
import gzhclient
from gzhclient import PlatformConfig

# Create client with default configuration
with gzhclient.Client() as client:
    # Check client health
    health = client.health()
    print(f"Client status: {health.overall.value}")
    
    # Get system metrics
    metrics = client.get_system_metrics()
    print(f"CPU usage: {metrics.cpu.usage:.1f}%")
    print(f"Memory usage: {metrics.memory.usage:.1f}%")
```

### Bulk Repository Cloning

```python
import gzhclient
from gzhclient import PlatformConfig, CloneFilters
from datetime import datetime, timedelta

# Configure platforms
platforms = [
    PlatformConfig(
        type="github",
        token="your-github-token",
        organizations=["kubernetes", "golang"]
    ),
    PlatformConfig(
        type="gitlab",
        url="https://gitlab.com",
        token="your-gitlab-token",
        organizations=["gitlab-org"]
    )
]

# Configure filters
filters = CloneFilters(
    languages=["go", "python", "javascript"],
    updated_after=datetime.now() - timedelta(days=30),
    exclude_repos=["test-*", "archive-*"]
)

# Perform bulk clone
with gzhclient.Client() as client:
    result = client.bulk_clone(
        platforms=platforms,
        output_dir="./repositories",
        concurrency=5,
        strategy="reset",
        filters=filters
    )
    
    print(f"Successfully cloned {result.success_count} repositories")
    print(f"Failed: {result.failure_count}")
    print(f"Duration: {result.duration:.2f} seconds")
```

### Plugin Management

```python
with gzhclient.Client() as client:
    # List available plugins
    plugins = client.list_plugins()
    for plugin in plugins:
        print(f"Plugin: {plugin.name} v{plugin.version}")
        print(f"  Description: {plugin.description}")
        print(f"  Status: {plugin.status}")
    
    # Execute a plugin
    if plugins:
        result = client.execute_plugin(
            plugin_name=plugins[0].name,
            method="process",
            args={"input": "test data"},
            timeout=30
        )
        print(f"Plugin result: {result.result}")
```

### Custom Configuration

```python
from gzhclient import Client, ClientConfig

# Custom configuration
config = ClientConfig(
    timeout=120,  # 2 minutes
    retry_count=5,
    enable_plugins=True,
    plugin_dir="/path/to/plugins",
    log_level="debug",
    features={
        "bulk_clone": True,
        "monitoring": True,
        "plugins": True
    }
)

with Client(config) as client:
    # Use client with custom configuration
    health = client.health()
    print(f"Client configured with {config.timeout}s timeout")
```

## 📚 Examples

The `examples/` directory contains comprehensive examples:

- [`basic_usage.py`](examples/basic_usage.py) - Basic client usage and health checks
- [`bulk_clone_example.py`](examples/bulk_clone_example.py) - Advanced bulk cloning with filters

Run examples:

```bash
cd examples
python basic_usage.py
python bulk_clone_example.py
```

## 🔧 Development

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/gizzahub/gzh-manager-go.git
cd gzh-manager-go/bindings/python

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install in development mode
pip install -e ".[dev]"

# Build the Go library
go build -buildmode=c-shared -o libgzh.so libgzh.go
```

### Running Tests

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=gzhclient --cov-report=html

# Run specific test categories
pytest -m unit
pytest -m integration
```

### Code Quality

```bash
# Format code
black gzhclient/ examples/ tests/

# Type checking
mypy gzhclient/

# Linting
flake8 gzhclient/
```

### Building Distribution

```bash
# Build package
python -m build

# Upload to PyPI (maintainers only)
twine upload dist/*
```

## 🌍 Environment Variables

The client respects several environment variables:

- `GZHCLIENT_LIB_PATH`: Path to the GZH Manager shared library
- `GITHUB_TOKEN`: GitHub personal access token
- `GITLAB_TOKEN`: GitLab personal access token
- `GITEA_TOKEN`: Gitea personal access token
- `GITEA_URL`: Gitea instance URL (default: https://gitea.com)

## 🔗 Integration

### With GitHub Actions

```yaml
- name: Setup GZH Manager
  run: |
    pip install gzhclient
    
- name: Clone repositories
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    python -c "
    import gzhclient
    from gzhclient import PlatformConfig
    
    with gzhclient.Client() as client:
        result = client.bulk_clone(
            platforms=[PlatformConfig(
                type='github',
                token='${{ secrets.GITHUB_TOKEN }}',
                organizations=['my-org']
            )],
            output_dir='./repos'
        )
        print(f'Cloned {result.success_count} repositories')
    "
```

### With Docker

```dockerfile
FROM python:3.11-slim

# Install Go for building the library
RUN apt-get update && apt-get install -y golang-go

# Install gzhclient
RUN pip install gzhclient

# Your application code
COPY . /app
WORKDIR /app

CMD ["python", "your_script.py"]
```

## 🐛 Troubleshooting

### Library Not Found

If you get an error about the shared library not being found:

1. Make sure Go is installed and available in PATH
2. Build the library manually: `go build -buildmode=c-shared -o libgzh.so libgzh.go`
3. Set the `GZHCLIENT_LIB_PATH` environment variable to the directory containing the library

### Permission Errors

On Linux/macOS, you might need to mark the library as executable:

```bash
chmod +x libgzh.so  # or libgzh.dylib on macOS
```

### Plugin Issues

If plugins are not loading:

1. Check that the plugin directory exists and is readable
2. Ensure plugins are compiled for the correct architecture
3. Enable debug logging: `ClientConfig(log_level="debug")`

## 📖 API Reference

### Client

- `Client(config: Optional[ClientConfig] = None)` - Create client instance
- `health() -> HealthStatus` - Get client health status
- `bulk_clone(platforms, output_dir, **kwargs) -> BulkCloneResult` - Perform bulk clone
- `list_plugins() -> List[PluginInfo]` - List available plugins
- `execute_plugin(plugin_name, method, args, timeout) -> PluginExecuteResult` - Execute plugin
- `get_system_metrics() -> SystemMetrics` - Get system metrics
- `close()` - Close client and cleanup resources

### Configuration

- `ClientConfig` - Client configuration options
- `PlatformConfig` - Git platform configuration
- `CloneFilters` - Repository filtering options

### Results

- `BulkCloneResult` - Bulk clone operation result
- `PluginExecuteResult` - Plugin execution result
- `SystemMetrics` - System metrics data
- `HealthStatus` - Client health information

For complete API documentation, see the docstrings in the source code.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## 🙏 Acknowledgments

- Built on top of [GZH Manager](https://github.com/gizzahub/gzh-manager-go)
- Uses Go's C shared library capabilities for Python integration
- Inspired by the need for better developer tooling in multi-repository environments

## 📞 Support

- 📖 [Documentation](https://github.com/gizzahub/gzh-manager-go/tree/main/bindings/python)
- 🐛 [Issue Tracker](https://github.com/gizzahub/gzh-manager-go/issues)
- 💬 [Discussions](https://github.com/gizzahub/gzh-manager-go/discussions)

---

Made with ❤️ by the GZH Manager team