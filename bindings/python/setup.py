"""
Setup script for GZH Manager Python client.
"""

import os
import platform
import subprocess
import sys
from pathlib import Path
from setuptools import setup, find_packages
from setuptools.command.build_ext import build_ext
from setuptools.command.install import install


class BuildGZHLibrary(build_ext):
    """Custom build command to compile the Go library."""
    
    def run(self):
        """Build the Go shared library."""
        # Check if Go is available
        try:
            subprocess.run(["go", "version"], check=True, capture_output=True)
        except (subprocess.CalledProcessError, FileNotFoundError):
            raise RuntimeError("Go compiler not found. Please install Go to build the library.")
        
        # Determine library extension based on platform
        system = platform.system().lower()
        if system == "windows":
            lib_ext = ".dll"
        elif system == "darwin":
            lib_ext = ".dylib"
        else:
            lib_ext = ".so"
        
        lib_name = f"libgzh{lib_ext}"
        
        # Build command
        build_cmd = [
            "go", "build",
            "-buildmode=c-shared",
            "-o", lib_name,
            "libgzh.go"
        ]
        
        print(f"Building Go library: {' '.join(build_cmd)}")
        
        try:
            subprocess.run(build_cmd, check=True, cwd=Path(__file__).parent)
            print(f"✅ Successfully built {lib_name}")
        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Failed to build Go library: {e}")
        
        # Run the standard build_ext
        super().run()


class InstallWithLibrary(install):
    """Custom install command that includes the shared library."""
    
    def run(self):
        # Build the library first
        self.run_command('build_ext')
        super().run()


def read_file(filename):
    """Read file contents."""
    with open(Path(__file__).parent / filename, 'r', encoding='utf-8') as f:
        return f.read()


def get_version():
    """Get version from __init__.py."""
    init_file = Path(__file__).parent / "gzhclient" / "__init__.py"
    with open(init_file, 'r', encoding='utf-8') as f:
        for line in f:
            if line.startswith('__version__'):
                return line.split('=')[1].strip().strip('"\'')
    return "0.1.0"


# Read long description from README
long_description = """
# GZH Manager Python Client

A Python client library for GZH Manager, providing programmatic access to bulk repository operations, 
plugin management, system monitoring, and more.

## Features

- **Bulk Repository Cloning**: Clone repositories from GitHub, GitLab, Gitea, and Gogs
- **Plugin Management**: Load, execute, and manage plugins
- **System Monitoring**: Collect CPU, memory, and disk metrics
- **Event System**: Subscribe to and handle system events
- **Platform-specific Clients**: Dedicated clients for each Git platform

## Quick Start

```python
import gzhclient
from gzhclient import PlatformConfig

# Create client
with gzhclient.Client() as client:
    # Check health
    health = client.health()
    print(f"Status: {health.overall.value}")
    
    # Bulk clone repositories
    result = client.bulk_clone(
        platforms=[
            PlatformConfig(
                type='github',
                token='your-token',
                organizations=['your-org']
            )
        ],
        output_dir='./repos'
    )
    print(f"Cloned {result.success_count} repositories")
```

## Requirements

- Python 3.8+
- Go 1.19+ (for building the underlying library)
- GZH Manager library dependencies

## Installation

```bash
pip install gzhclient
```

For development installation:

```bash
git clone https://github.com/gizzahub/gzh-manager-go.git
cd gzh-manager-go/bindings/python
pip install -e .
```

## Documentation

For complete documentation and examples, visit:
https://github.com/gizzahub/gzh-manager-go/tree/main/bindings/python

## License

Licensed under the same terms as GZH Manager.
"""

setup(
    name="gzhclient",
    version=get_version(),
    author="GZH Manager Team",
    author_email="support@gzh-manager.com",
    description="Python client library for GZH Manager",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/gizzahub/gzh-manager-go",
    project_urls={
        "Bug Reports": "https://github.com/gizzahub/gzh-manager-go/issues",
        "Source": "https://github.com/gizzahub/gzh-manager-go/tree/main/bindings/python",
        "Documentation": "https://github.com/gizzahub/gzh-manager-go/tree/main/bindings/python/README.md",
    },
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Software Development :: Version Control :: Git",
        "Topic :: System :: Systems Administration",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Programming Language :: Go",
        "Operating System :: OS Independent",
        "Operating System :: POSIX :: Linux",
        "Operating System :: MacOS",
        "Operating System :: Microsoft :: Windows",
    ],
    keywords="git, repository, bulk, clone, github, gitlab, gitea, gogs, devops, automation",
    python_requires=">=3.8",
    install_requires=[
        # No external dependencies - uses only Python standard library
    ],
    extras_require={
        "dev": [
            "pytest>=6.0",
            "pytest-cov>=2.0",
            "black>=21.0",
            "flake8>=3.8",
            "mypy>=0.812",
            "build>=0.7",
            "twine>=3.4",
        ],
        "test": [
            "pytest>=6.0",
            "pytest-cov>=2.0",
            "pytest-mock>=3.0",
        ],
    },
    package_data={
        "gzhclient": ["*.so", "*.dylib", "*.dll", "*.h"],
    },
    include_package_data=True,
    zip_safe=False,
    cmdclass={
        'build_ext': BuildGZHLibrary,
        'install': InstallWithLibrary,
    },
    entry_points={
        "console_scripts": [
            "gzh-python-example=gzhclient.examples:main",
        ],
    },
)