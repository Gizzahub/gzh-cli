# Command: gz dev-env status

## Scenario: Display current development environment status

### Input

**Command**:

```bash
gz dev-env status
```

**Prerequisites**:

- [ ] Development environment system initialized
- [ ] Access to configuration files

### Expected Output

**Active Environment**:

```text
🌍 Development Environment Status

📋 Active Profile: aws-prod
   ⏰ Switched: 2025-09-02 14:30:15 KST
   👤 User: devops-team
   🔄 Duration: 2h 15m 30s

☁️  AWS Configuration
✅ Profile: aws-prod-account (active)
   • Region: ap-northeast-2
   • Account: 987654321098
   • Credentials: valid (expires in 11h 45m)

🐳 Docker Configuration
✅ Context: aws-prod-ecs (active)
   • Endpoint: tcp://prod-ecs.amazonaws.com:2376
   • TLS: verified
   • Connection: healthy

☸️  Kubernetes Configuration
✅ Context: prod-k8s-cluster (active)
   • Cluster: prod-eks-cluster
   • Namespace: production
   • Server: https://k8s-prod.example.com
   • Connection: healthy

🔗 SSH Configuration
✅ Active tunnels: 3
   • prod-bastion: dev.example.com:22 → prod.example.com:22
   • db-tunnel: localhost:5432 → prod-db.internal:5432
   • redis-tunnel: localhost:6379 → prod-redis.internal:6379

🌐 Network Status
✅ VPN: connected to prod-vpc
✅ DNS: prod-dns.example.com (8.8.8.8 fallback)
✅ Proxy: corp-proxy.example.com:8080

📊 Resource Usage
   • Memory: 45MB
   • Active connections: 12
   • Background processes: 5

stderr: (empty)
Exit Code: 0
```

**No Active Environment**:

```text
🌍 Development Environment Status

❌ No active development environment

📋 Available profiles:
   • local - Local development setup
   • aws-dev - AWS development environment
   • aws-staging - AWS staging environment
   • aws-prod - AWS production environment
   • docker-local - Local Docker setup
   • k8s-dev - Kubernetes development cluster

💡 Switch to environment:
   gz dev-env switch --profile <profile-name>

💡 Create new profile:
   gz dev-env create --profile <new-profile>

🚫 Initialize an environment to get started.

stderr: (empty)
Exit Code: 1
```

**Degraded Environment**:

```text
🌍 Development Environment Status

⚠️  Active Profile: aws-prod (degraded)
   ⏰ Switched: 2025-09-02 14:30:15 KST
   🔄 Duration: 2h 15m 30s

☁️  AWS Configuration
❌ Profile: aws-prod-account (expired)
   • Region: ap-northeast-2
   • Account: 987654321098
   • Credentials: expired 15m ago

🐳 Docker Configuration
⚠️  Context: aws-prod-ecs (connection issues)
   • Endpoint: tcp://prod-ecs.amazonaws.com:2376
   • TLS: verified
   • Connection: timeout (retry in 30s)

☸️  Kubernetes Configuration
✅ Context: prod-k8s-cluster (active)
   • Cluster: prod-eks-cluster
   • Namespace: production
   • Connection: healthy

🔗 SSH Configuration
❌ Tunnel failures: 2/3
   ✅ prod-bastion: connected
   ❌ db-tunnel: connection refused
   ❌ redis-tunnel: authentication failed

⚠️  Environment requires attention!

💡 Fix issues:
   - Refresh AWS credentials: aws sso login --profile aws-prod-account
   - Check Docker endpoint: docker context inspect aws-prod-ecs
   - Restart SSH tunnels: gz dev-env switch --profile aws-prod --force

stderr: degraded environment detected
Exit Code: 2
```

### Side Effects

**Files Created**:

- `~/.gzh/dev-env/status-cache.json` - Status information cache
- `~/.gzh/dev-env/health-check.log` - Health check results

**Files Modified**: None
**State Changes**: Status cache updated with latest health checks

### Validation

**Automated Tests**:

```bash
# Test status display
result=$(gz dev-env status 2>&1)
exit_code=$?

# Should show either active environment or no environment message
assert_contains "$result" "Development Environment Status"

# Check cache file creation
assert_file_exists "$HOME/.gzh/dev-env/status-cache.json"
cache_content=$(cat "$HOME/.gzh/dev-env/status-cache.json")
assert_contains "$cache_content" '"timestamp":'
```

**Manual Verification**:

1. Check status with active environment
1. Verify all service connections are accurate
1. Test status with no active environment
1. Confirm degraded state detection works
1. Validate credential expiration warnings

### Edge Cases

**Stale State Detection**:

- Environment switched in different session
- Configuration files modified externally
- Services restarted outside of tool

**Network Connectivity Issues**:

- Cloud provider API unreachable
- SSH tunnels dropped unexpectedly
- VPN disconnection detection

**Credential Rotation**:

- AWS SSO token refresh needed
- Kubernetes token expiration
- SSH key rotation required

**Resource Exhaustion**:

- Too many open connections
- Memory usage monitoring
- Process cleanup for orphaned connections

### Performance Expectations

**Response Time**:

- Cached status: < 1 second
- Full health check: < 5 seconds
- Network checks: < 10 seconds

**Resource Usage**:

- Memory: < 30MB
- Network: Minimal health check calls
- CPU: Low impact status collection

## Notes

- Real-time health monitoring of all services
- Credential expiration tracking and warnings
- Connection pooling for efficient status checks
- Background health monitoring (optional daemon)
- Integration with system monitoring tools
- Export capabilities (JSON, YAML, Prometheus metrics)
