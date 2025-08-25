# Docker-based Integration Tests

This directory contains Docker-based integration tests using testcontainers-go. These tests spin up real service containers (GitLab, Gitea, Redis) to test integration scenarios.

## Prerequisites

### Docker Environment

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 4GB RAM available for containers
- Internet connection for pulling container images

### Go Dependencies

The following dependencies are automatically managed:

- `github.com/testcontainers/testcontainers-go`
- Standard testing packages

## Container Images Used

| Service   | Image              | Version      | Purpose                    |
| --------- | ------------------ | ------------ | -------------------------- |
| GitLab CE | `gitlab/gitlab-ce` | 16.11.0-ce.0 | GitLab integration testing |
| Gitea     | `gitea/gitea`      | 1.21.10      | Gitea integration testing  |
| Redis     | `redis`            | 7.2-alpine   | Cache integration testing  |

## Running Tests

### Run All Docker Integration Tests

```bash
go test ./test/integration/docker/... -v
```

### Run Specific Test Suite

```bash
# GitLab integration tests
go test ./test/integration/docker -v -run TestBulkClone_GitLab

# Gitea integration tests
go test ./test/integration/docker -v -run TestBulkClone_Gitea

# Redis cache integration tests
go test ./test/integration/docker -v -run TestBulkClone_Redis

# Multi-provider integration tests
go test ./test/integration/docker -v -run TestMultiProvider
```

### Skip Docker Tests (Short Mode)

```bash
go test ./test/integration/docker/... -v -short
```

### Run with Extended Timeout

```bash
go test ./test/integration/docker/... -v -timeout 30m
```

## Test Scenarios

### 1. GitLab Integration (`TestBulkClone_GitLab_Integration`)

- **Duration**: ~15 minutes
- **Resources**: GitLab CE container (2GB RAM)
- **Tests**:
  - GitLab container startup and readiness
  - Configuration loading with GitLab provider
  - GitLab API connectivity validation

### 2. Gitea Integration (`TestBulkClone_Gitea_Integration`)

- **Duration**: ~10 minutes
- **Resources**: Gitea container (512MB RAM)
- **Tests**:
  - Gitea container startup and readiness
  - Configuration loading with Gitea provider
  - Gitea API connectivity validation

### 3. Redis Cache Integration (`TestBulkClone_Redis_Cache_Integration`)

- **Duration**: ~5 minutes
- **Resources**: Redis container (100MB RAM)
- **Tests**:
  - Redis container startup
  - Cache configuration validation
  - Redis connectivity testing

### 4. Multi-Provider Integration (`TestMultiProvider_Integration`)

- **Duration**: ~15 minutes
- **Resources**: GitLab + Gitea + Redis containers (2.6GB RAM)
- **Tests**:
  - Multiple container orchestration
  - Cross-provider configuration
  - Cache integration with multiple providers

## Container Configuration

### GitLab Container

```yaml
Image: gitlab/gitlab-ce:16.11.0-ce.0
Ports: 80/tcp, 22/tcp
Environment:
  - GITLAB_OMNIBUS_CONFIG: Optimized for testing
  - initial_root_password: testpassword123
Memory: ~2GB
Startup Time: ~10 minutes
```

### Gitea Container

```yaml
Image: gitea/gitea:1.21.10
Ports: 3000/tcp, 22/tcp
Environment:
  - DB_TYPE: sqlite3
  - INSTALL_LOCK: true
  - Preconfigured admin credentials
Memory: ~512MB
Startup Time: ~2 minutes
```

### Redis Container

```yaml
Image: redis:7.2-alpine
Ports: 6379/tcp
Configuration:
  - appendonly: yes
  - maxmemory: 100mb
  - maxmemory-policy: allkeys-lru
Memory: ~100MB
Startup Time: ~30 seconds
```

## Performance Considerations

### Resource Usage

- **Total Memory**: Up to 2.6GB for full test suite
- **CPU**: Moderate during container startup
- **Disk**: ~5GB for container images
- **Network**: Container-to-container communication

### Optimization Strategies

1. **Parallel Execution**: Tests run containers independently
1. **Resource Limits**: Containers configured with memory limits
1. **Fast Cleanup**: Containers terminated immediately after tests
1. **Image Caching**: Docker images cached between runs

## Troubleshooting

### Common Issues

#### Docker Not Available

```
Error: Cannot connect to the Docker daemon
```

**Solution**: Ensure Docker is running and accessible

```bash
docker version
systemctl start docker  # Linux
```

#### Insufficient Memory

```
Error: Container startup timeout
```

**Solution**: Increase available memory or reduce parallel tests

```bash
docker system prune -f  # Clean up unused containers
```

#### Network Connectivity Issues

```
Error: Cannot pull container image
```

**Solution**: Check internet connectivity and proxy settings

```bash
docker pull gitlab/gitlab-ce:16.11.0-ce.0  # Test manually
```

#### Port Conflicts

```
Error: Port already in use
```

**Solution**: Testcontainers automatically assigns random ports, but ensure no conflicts

```bash
docker ps  # Check running containers
```

### Debug Commands

```bash
# Check container logs
docker logs <container-id>

# Check container status
docker ps -a

# Check resource usage
docker stats

# Clean up test containers
docker container prune -f
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Docker Integration Tests
on: [push, pull_request]

jobs:
  docker-tests:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:20.10-dind
        options: --privileged

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.24

      - name: Run Docker Integration Tests
        run: |
          go test ./test/integration/docker/... -v -timeout 30m
```

### Resource Requirements for CI

- **Memory**: 8GB+ recommended
- **Storage**: 20GB+ for container images
- **Network**: Stable internet for image pulls
- **Time**: 30-45 minutes for full test suite

## Security Considerations

### Test Isolation

- Each test uses temporary directories
- Containers use test-specific credentials
- No production data or secrets

### Container Security

- Containers run with minimal privileges
- Test-only configuration (not production-ready)
- Automatic cleanup prevents resource leaks

### Network Security

- Containers use isolated bridge networks
- No external network exposure
- Test traffic only between containers

## Future Enhancements

- [ ] **Gogs Container**: Add Gogs integration testing
- [ ] **GitHub Enterprise**: Test with GitHub Enterprise container
- [ ] **Database Integration**: Add PostgreSQL/MySQL containers
- [ ] **Monitoring**: Add Prometheus/Grafana test containers
- [ ] **Load Testing**: Performance testing with multiple containers
- [ ] **Chaos Testing**: Network failure simulation
- [ ] **Cross-Platform**: Windows container testing

## Maintenance

### Regular Tasks

1. **Update Container Images**: Monthly security updates
1. **Performance Monitoring**: Track test execution times
1. **Resource Optimization**: Monitor memory/CPU usage
1. **Documentation Updates**: Keep README current

### Version Management

- Container images pinned to specific versions
- Upgrade strategy: test new versions in separate branch
- Rollback plan: revert to previous working versions
