# SelfUpdate Command - Agent Guidelines

## Overview

The selfupdate command enables automatic updates of the gz binary by downloading the latest release from GitHub. It provides safe binary replacement with backup and rollback capabilities.

## Architecture

### Core Components

1. **Updater**: Main update logic with GitHub API integration
1. **GitHubRelease**: Release metadata parsing
1. **Platform Detection**: Cross-platform binary name resolution
1. **Safe Replacement**: Binary replacement with rollback support

### Key Features

- GitHub Releases API integration
- Cross-platform binary name detection
- Safe binary replacement (backup on Windows)
- Version comparison with semantic versioning support
- Force update option
- Context-aware cancellation
- Progress logging

## Development Guidelines

### Code Organization

```
cmd/selfupdate/
├── register.go      # Command registration
├── selfupdate.go    # Main implementation
├── selfupdate_test.go # Unit tests
└── AGENTS.md        # This documentation
```

### Testing Strategy

#### Unit Tests

- Version comparison logic
- Asset name generation
- Platform detection
- Error handling

#### Integration Tests (Future)

- GitHub API mock tests
- Binary replacement simulation
- Network error handling

#### Manual Testing

```bash
# Test version check (dry run would be ideal)
gz selfupdate --help

# Test with force flag
gz selfupdate --force

# Test current version detection
gz version && gz selfupdate
```

### Platform Support

#### Asset Naming Convention

```
gz_${GOOS}_${GOARCH}${EXT}
```

#### Supported Platforms

- Linux (x86_64, i386, arm64)
- Windows (x86_64, i386) with .exe extension
- Darwin/macOS (x86_64, arm64)

#### Architecture Mapping

- `amd64` → `x86_64`
- `386` → `i386`
- Other architectures use Go names directly

### Security Considerations

1. **HTTPS Only**: All GitHub API calls use HTTPS
1. **Temporary Files**: Downloads to system temp directory
1. **Atomic Replacement**: Binary replacement is atomic where possible
1. **Backup Strategy**: Windows backup before replacement
1. **Permission Preservation**: Unix executable permissions maintained

### Error Handling

#### Network Errors

- Timeout handling (30s for API, 5min for download)
- HTTP status code validation
- JSON parsing error handling

#### File System Errors

- Temporary file creation
- Permission issues
- Disk space validation (implicit)
- Backup/restore on Windows

#### GitHub API Errors

- Rate limiting awareness
- Asset not found handling
- Release format validation

### Logging Strategy

#### Info Level

- Update check started
- Version comparison results
- Download progress
- Successful completion

#### Error Level

- Network failures
- File system errors
- API response errors
- Rollback situations

### Configuration

#### Environment Variables

- Uses system temp directory
- Respects HTTP proxy settings via Go's net/http

#### Command Flags

- `--force`: Skip version check, force update

### Future Enhancements

#### Potential Features

1. **Checksum Verification**: SHA256 checksums for downloads
1. **Incremental Updates**: Delta updates for large binaries
1. **Rollback Command**: `gz selfupdate --rollback`
1. **Update Channels**: Stable, beta, nightly releases
1. **Auto-update Scheduling**: Background update checks
1. **Configuration**: Update preferences in config file

#### API Enhancements

1. **Rate Limit Handling**: GitHub API rate limit respect
1. **Authentication**: Support for GitHub tokens
1. **Enterprise Support**: GitHub Enterprise Server support
1. **Mirror Support**: Alternative download sources

### Testing Commands

```bash
# Unit tests (when Go version supports)
go test ./cmd/selfupdate -v

# Lint checks
make lint

# Format code
make fmt

# Integration with full build
make build && ./gz selfupdate --help
```

### Version Handling

#### Version Detection

1. Root command version (preferred)
1. Fallback to "dev" for development builds
1. Support for "v" prefixed versions

#### Comparison Logic

- Simple string comparison after prefix removal
- "dev" and empty versions always trigger updates
- Future: Semantic version comparison

### Cross-Platform Considerations

#### Windows Specific

- .exe file extension
- Backup before replacement (file locking)
- Cleanup of .old backup files

#### Unix/Linux Specific

- Preserve executable permissions (0755)
- Direct file replacement
- Symlink resolution

#### macOS Specific

- Same as Unix/Linux
- Future: Code signing considerations

## Command Usage

```bash
# Check and update to latest version
gz selfupdate

# Force update even if already latest
gz selfupdate --force

# Get help
gz selfupdate --help
```

## Implementation Notes

### GitHub API Integration

- Uses public GitHub API (no authentication required)
- Respects API rate limits
- Handles JSON response parsing
- Timeout configuration

### Binary Replacement Strategy

- Atomic replacement where possible
- Windows: backup → replace → cleanup
- Unix: direct replacement
- Error recovery through backup restoration

### Asset Selection Logic

1. Generate expected asset name for current platform
1. Search release assets for exact match
1. Error if no matching asset found
1. Download and replace binary

This documentation should be updated as the selfupdate command evolves and new features are added.
