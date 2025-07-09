# Configuration Priority Testing Guide

This document provides test scenarios to verify that the configuration priority system works as documented.

## Test Scenarios

### Test 1: CLI Flag Override

**Setup:**
```yaml
# test-config.yaml
version: "1.0.0"
global:
  default_strategy: reset
  concurrency:
    clone_workers: 10
providers:
  github:
    token: "ghp_config_token"
```

```bash
export GITHUB_TOKEN=ghp_env_token
```

**Test Command:**
```bash
gz bulk-clone --config=test-config.yaml --strategy=pull --parallel=20 --token=ghp_flag_token
```

**Expected Result:**
- Strategy: `pull` (CLI flag overrides config file)
- Parallel: `20` (CLI flag overrides config file)
- Token: `ghp_flag_token` (CLI flag overrides environment variable)

### Test 2: Environment Variable Override

**Setup:**
```yaml
# test-config.yaml
version: "1.0.0"
providers:
  github:
    token: "ghp_config_token"
```

```bash
export GITHUB_TOKEN=ghp_env_token
```

**Test Command:**
```bash
gz bulk-clone --config=test-config.yaml
```

**Expected Result:**
- Token: `ghp_env_token` (environment variable overrides config file)

### Test 3: Configuration File Priority

**Setup:**
```yaml
# test-config.yaml
version: "1.0.0"
global:
  default_strategy: reset
  concurrency:
    clone_workers: 15
```

**Test Command:**
```bash
gz bulk-clone --config=test-config.yaml
```

**Expected Result:**
- Strategy: `reset` (from config file)
- Parallel: `15` (from config file)

### Test 4: Default Values

**Setup:**
No configuration file, no environment variables

**Test Command:**
```bash
gz bulk-clone
```

**Expected Result:**
- Strategy: `reset` (default value)
- Parallel: `10` (default value)

### Test 5: Environment Variable Expansion

**Setup:**
```yaml
# test-config.yaml
version: "1.0.0"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    api_url: "${GITHUB_API_URL:-https://api.github.com}"
```

```bash
export GITHUB_TOKEN=ghp_expanded_token
# GITHUB_API_URL not set
```

**Test Command:**
```bash
gz bulk-clone --config=test-config.yaml
```

**Expected Result:**
- Token: `ghp_expanded_token` (expanded from environment variable)
- API URL: `https://api.github.com` (default value from expansion)

### Test 6: Configuration File Search Order

**Setup:**
```yaml
# ./gzh.yaml
version: "1.0.0"
global:
  default_strategy: pull
```

```yaml
# ~/.config/gzh-manager/gzh.yaml
version: "1.0.0"
global:
  default_strategy: reset
```

**Test Command:**
```bash
gz bulk-clone  # No --config flag
```

**Expected Result:**
- Strategy: `pull` (current directory takes precedence over user config)

## Manual Testing Commands

To manually test the priority system:

```bash
# Test CLI flag override
gz config show --strategy=pull --parallel=20

# Test environment variable
export GITHUB_TOKEN=test_token
gz config show

# Test configuration file
gz config show --config=test-config.yaml

# Test default values
gz config show --no-config

# Test search order
gz config paths
```

## Automated Testing

The priority system should be covered by unit tests in:
- `pkg/config/priority_test.go`
- Integration tests in `cmd/*/cmd_test.go`

**Test Structure:**
```go
func TestConfigurationPriority(t *testing.T) {
    // Test CLI flag override
    // Test environment variable override
    // Test configuration file priority
    // Test default values
}
```

## Validation Checklist

- [ ] CLI flags override all other sources
- [ ] Environment variables override config files
- [ ] Configuration files override default values
- [ ] Environment variable expansion works correctly
- [ ] Configuration file search order is correct
- [ ] Default values are used when no other source provides values
- [ ] Priority documentation matches actual behavior
- [ ] All commands follow the same priority rules
- [ ] Debug commands show correct priority resolution

## Common Issues

1. **Environment variable not expanding**: Check syntax `${VAR_NAME}`
2. **Config file not found**: Use `gz config paths` to check search order
3. **Priority not working**: Use `gz config show` to verify effective configuration
4. **Default values unexpected**: Check documentation for correct default values

## References

- [Configuration Priority Guide](configuration-priority.md)
- [Configuration System](configuration.md)
- [Configuration Migration Guide](configuration-migration.md)