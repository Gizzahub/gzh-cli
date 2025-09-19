# Command: gz dev-env switch

## Scenario: Switch between development environments

### Input

**Command**:

```bash
gz dev-env switch --profile aws-prod
```

**Prerequisites**:

- [ ] Development environment profiles configured
- [ ] Required credentials available
- [ ] Network connectivity for cloud environments

### Expected Output

**Success Case**:

```text
🔄 Switching to development environment: aws-prod

📋 Environment Changes:
✅ AWS Profile: default → aws-prod-account
   • Region: us-east-1 → ap-northeast-2
   • Account: 123456789012 → 987654321098

✅ Docker Context: default → aws-prod-ecs
   • Endpoint: unix:///var/run/docker.sock → tcp://prod-ecs.amazonaws.com:2376

✅ Kubernetes Context: local → prod-k8s-cluster
   • Cluster: minikube → prod-eks-cluster
   • Namespace: default → production

✅ SSH Config: dev-bastion → prod-bastion
   • Host: dev.example.com → prod.example.com
   • User: dev-user → prod-user

🎉 Successfully switched to aws-prod environment!

💡 Active services:
   - AWS CLI configured for production account
   - Docker pointing to ECS cluster
   - kubectl configured for production cluster
   - SSH tunnels established

stderr: (empty)
Exit Code: 0
```

**Profile Not Found**:

```text
🔍 Searching for development environment: aws-prod

❌ Environment profile 'aws-prod' not found!

📋 Available profiles:
   • local (currently active)
   • aws-dev
   • aws-staging
   • docker-local
   • k8s-dev

💡 Create new profile:
   gz dev-env create --profile aws-prod

🚫 Environment switch failed.

stderr: profile not found
Exit Code: 1
```

**Credentials Missing**:

```text
🔄 Switching to development environment: aws-prod

⚠️  Credential validation failed:
   ❌ AWS credentials not found for profile 'aws-prod-account'
   ❌ Docker TLS certificates missing for remote context
   ✅ Kubernetes config valid

💡 Fix credentials:
   - AWS: aws configure --profile aws-prod-account
   - Docker: docker context create aws-prod-ecs --docker host=tcp://...

🔧 Partial switch completed. Fix credentials and retry.

stderr: missing credentials
Exit Code: 1
```

### Side Effects

**Files Created**:

- `~/.gzh/dev-env/current.yaml` - Active environment state
- `~/.gzh/dev-env/switch.log` - Environment switch log

**Files Modified**:

- `~/.aws/config` - AWS profile selection
- `~/.docker/config.json` - Docker context selection
- `~/.kube/config` - Kubernetes context selection
- `~/.ssh/config` - SSH configuration updates

**State Changes**:

- Environment variables exported to shell
- Active cloud provider context switched
- Container orchestration context switched
- VPN connections established/terminated

### Validation

**Automated Tests**:

```bash
# Test environment switch
result=$(gz dev-env switch --profile test-env 2>&1)
exit_code=$?

assert_contains "$result" "Switching to development environment"
assert_contains "$result" "Successfully switched"

# Verify state file creation
assert_file_exists "$HOME/.gzh/dev-env/current.yaml"
current_env=$(yq r "$HOME/.gzh/dev-env/current.yaml" 'active_profile')
assert_equals "$current_env" "test-env"
```

**Manual Verification**:

1. Switch between different environment profiles
1. Verify AWS CLI uses correct account/region
1. Check Docker context points to correct endpoint
1. Confirm kubectl uses correct cluster/namespace
1. Test SSH connections use correct bastion hosts

### Edge Cases

**Concurrent Switches**:

- Multiple terminal sessions attempting switches
- Lock file prevents concurrent operations
- Clear error message for locked environments

**Partial Failures**:

- Some services switch successfully, others fail
- Rollback mechanism for failed switches
- Clear indication of what succeeded/failed

**Network Issues**:

- Cloud provider API unavailable
- VPN connection failures
- Timeout handling for remote services

**Configuration Conflicts**:

- Conflicting environment variables
- Overlapping port bindings
- Resource allocation conflicts

### Performance Expectations

**Response Time**:

- Local environments: < 3 seconds
- Cloud environments: < 15 seconds
- Complex multi-service: < 30 seconds

**Resource Usage**:

- Memory: < 50MB
- Network: Minimal for validation calls
- Disk: Configuration file updates only

## Notes

- Supports AWS, Azure, GCP cloud environments
- Docker and Kubernetes context management
- SSH tunnel and bastion host configuration
- Environment variable isolation per profile
- Rollback capability for failed switches
- Integration with shell environment (bash, zsh, fish)
