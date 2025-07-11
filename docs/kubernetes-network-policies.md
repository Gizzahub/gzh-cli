# Kubernetes Network Policy Management

The Kubernetes network policy management feature in `gz net-env kubernetes-network` provides comprehensive namespace-specific network configuration capabilities. This feature allows you to manage Kubernetes NetworkPolicies as profiles, enabling consistent security policies across different environments.

## Overview

Kubernetes network profiles allow you to:
- Define and manage NetworkPolicies for namespace-specific network segmentation
- Automatically generate NetworkPolicy YAML manifests
- Export existing namespace policies to reusable profiles
- Apply consistent network security across environments
- Integrate with service mesh solutions (Istio/Linkerd)

## Basic Usage

### Creating a Network Profile

```bash
# Create a simple network profile
gz net-env kubernetes-network create myapp-policies \
  --namespace production \
  --description "Production network policies"

# Create a profile interactively
gz net-env kubernetes-network create myapp-policies --interactive
```

### Managing Network Policies

```bash
# Add a network policy to a profile
gz net-env kubernetes-network policy add myapp-policies frontend-policy \
  --pod-selector app=frontend,tier=web \
  --policy-types Ingress,Egress \
  --allow-from pod:app=gateway --allow-from ns:name=monitoring \
  --allow-to pod:app=api \
  --ports TCP:80,TCP:443

# Show policy details
gz net-env kubernetes-network policy show myapp-policies frontend-policy

# Remove a policy from profile
gz net-env kubernetes-network policy remove myapp-policies old-policy
```

### Applying Profiles

```bash
# Apply a network profile (creates NetworkPolicies in the namespace)
gz net-env kubernetes-network apply myapp-policies

# Dry run to see what would be applied
gz net-env kubernetes-network apply myapp-policies --dry-run
```

### Exporting and Importing

```bash
# Export existing namespace policies to a profile
gz net-env kubernetes-network export namespace production prod-policies

# Export profile to file
gz net-env kubernetes-network export profile myapp-policies myapp-policies.yaml
```

## Advanced Features

### Zero-Trust Network Security

Implement a deny-all default policy with specific exceptions:

```yaml
name: zero-trust
namespace: secure-app
description: Zero trust network security baseline
policies:
  default-deny-all:
    name: default-deny-all
    pod_selector: {}  # Empty selector applies to all pods
    policy_types:
      - Ingress
      - Egress
    # No rules = deny all traffic
  
  allow-dns:
    name: allow-dns
    pod_selector: {}  # All pods need DNS
    policy_types:
      - Egress
    egress:
      - to:
        - namespace_selector:
            name: kube-system
          pod_selector:
            k8s-app: kube-dns
        ports:
        - protocol: UDP
          port: 53
```

### Microservices Network Policies

Create comprehensive policies for microservices architecture:

```yaml
name: microservices
namespace: production
description: Microservices network segmentation
policies:
  frontend-policy:
    name: frontend-policy
    pod_selector:
      tier: frontend
    policy_types:
      - Ingress
      - Egress
    ingress:
      - from:
        - ip_block:
            cidr: 0.0.0.0/0  # Allow from internet
        ports:
        - protocol: TCP
          port: 80
        - protocol: TCP
          port: 443
    egress:
      - to:
        - pod_selector:
            tier: api
        ports:
        - protocol: TCP
          port: 8080
      - to:  # DNS resolution
        - namespace_selector:
            name: kube-system
        ports:
        - protocol: UDP
          port: 53
  
  api-policy:
    name: api-policy
    pod_selector:
      tier: api
    policy_types:
      - Ingress
      - Egress
    ingress:
      - from:
        - pod_selector:
            tier: frontend
        - namespace_selector:
            name: monitoring
        ports:
        - protocol: TCP
          port: 8080
    egress:
      - to:
        - pod_selector:
            tier: database
        ports:
        - protocol: TCP
          port: 5432  # PostgreSQL
      - to:
        - pod_selector:
            tier: cache
        ports:
        - protocol: TCP
          port: 6379  # Redis
  
  database-policy:
    name: database-policy
    pod_selector:
      tier: database
    policy_types:
      - Ingress
    ingress:
      - from:
        - pod_selector:
            tier: api
        - pod_selector:
            app: backup
        ports:
        - protocol: TCP
          port: 5432
```

### Multi-Tenant Isolation

Implement namespace isolation for multi-tenant environments:

```yaml
name: tenant-isolation
namespace: tenant-a
description: Tenant A network isolation
policies:
  isolate-tenant:
    name: isolate-tenant
    pod_selector: {}
    policy_types:
      - Ingress
      - Egress
    ingress:
      - from:
        - pod_selector: {}  # Only from same namespace
    egress:
      - to:
        - pod_selector: {}  # Only to same namespace
      - to:  # Allow external services
        - namespace_selector:
            name: kube-system
      - to:  # Allow internet access
        - ip_block:
            cidr: 0.0.0.0/0
            except:
              - 10.0.0.0/8
              - 172.16.0.0/12
              - 192.168.0.0/16
```

## Network Policy Examples

### Allow Specific Pods

```bash
# Allow traffic from specific pods
gz net-env kubernetes-network policy add myapp web-policy \
  --pod-selector app=web \
  --allow-from pod:app=frontend,tier=ui \
  --ports TCP:8080
```

### Allow from Namespace

```bash
# Allow traffic from specific namespace
gz net-env kubernetes-network policy add myapp monitoring-policy \
  --pod-selector app=api \
  --allow-from ns:name=monitoring \
  --ports TCP:9090
```

### Allow IP Blocks

```bash
# Allow traffic from specific IP ranges
gz net-env kubernetes-network policy add myapp external-policy \
  --pod-selector app=public-api \
  --allow-from 203.0.113.0/24 \
  --ports TCP:443
```

### Port Ranges

```bash
# Allow port ranges
gz net-env kubernetes-network policy add myapp range-policy \
  --pod-selector app=service \
  --allow-from pod:app=client \
  --ports TCP:8080-8090
```

## Service Mesh Integration

### Istio Integration

When using Istio, you can include service mesh metadata:

```yaml
name: istio-app
namespace: istio-enabled
metadata:
  service-mesh: istio
  mtls-mode: STRICT
  traffic-policy: round-robin
  circuit-breaker: enabled
  retry-attempts: "3"
  timeout: 30s
policies:
  # NetworkPolicies work alongside Istio policies
  allow-istio-system:
    name: allow-istio-system
    pod_selector: {}
    policy_types:
      - Ingress
      - Egress
    ingress:
      - from:
        - namespace_selector:
            name: istio-system
    egress:
      - to:
        - namespace_selector:
            name: istio-system
```

### Linkerd Integration

For Linkerd service mesh:

```yaml
name: linkerd-app
namespace: linkerd-enabled
metadata:
  service-mesh: linkerd
  proxy-cpu-request: 100m
  proxy-cpu-limit: 250m
  proxy-memory: 64Mi
  tap-enabled: "true"
policies:
  # NetworkPolicies complement Linkerd's mTLS
```

## Best Practices

1. **Default Deny**: Start with a deny-all policy and explicitly allow required traffic
2. **Least Privilege**: Only allow the minimum required network access
3. **Namespace Isolation**: Use namespace selectors for cross-namespace communication
4. **Label Consistency**: Maintain consistent pod labels across your deployments
5. **DNS Access**: Remember to allow DNS (port 53/UDP to kube-system)
6. **Testing**: Test policies in non-production environments first
7. **Documentation**: Document the purpose of each policy in the description

## Generating Manifests

Generate NetworkPolicy YAML files for version control:

```bash
# Generate all policies from a profile
gz net-env kubernetes-network generate policies myapp-policies ./manifests/

# This creates files like:
# manifests/networkpolicy-frontend-policy.yaml
# manifests/networkpolicy-api-policy.yaml
# manifests/networkpolicy-database-policy.yaml
```

## Monitoring and Status

Check the status of NetworkPolicies:

```bash
# Show policies in default namespace
gz net-env kubernetes-network status

# Show policies in specific namespace
gz net-env kubernetes-network status --namespace production

# Show policies in all namespaces
gz net-env kubernetes-network status --all-namespaces

# JSON output for automation
gz net-env kubernetes-network status --output json
```

## Troubleshooting

### Common Issues

1. **No Network Connectivity**
   - Check if a deny-all policy is blocking traffic
   - Verify pod labels match the selectors
   - Ensure DNS resolution is allowed

2. **Partial Connectivity**
   - Review policy precedence (NetworkPolicies are additive)
   - Check both ingress and egress rules
   - Verify port numbers and protocols

3. **Policy Not Applied**
   - Confirm the namespace exists
   - Check if NetworkPolicy API is enabled in your cluster
   - Verify RBAC permissions

### Debugging Commands

```bash
# Check if policies are created
kubectl get networkpolicies -n <namespace>

# Describe a specific policy
kubectl describe networkpolicy <policy-name> -n <namespace>

# Test connectivity between pods
kubectl exec -it <source-pod> -- nc -zv <target-pod> <port>

# Check pod labels
kubectl get pods -n <namespace> --show-labels
```

## Migration Guide

### From Existing Policies

If you have existing NetworkPolicies, export them to a profile:

```bash
# Export all policies from a namespace
gz net-env kubernetes-network export namespace production prod-profile

# Review the exported profile
gz net-env kubernetes-network list -o yaml

# Apply to another environment
gz net-env kubernetes-network apply prod-profile --namespace staging
```

### From Other Tools

Convert policies from other formats:
- Calico policies can be manually converted to NetworkPolicy format
- Cilium policies may need adjustment for standard NetworkPolicy API
- Service mesh policies (Istio/Linkerd) can coexist with NetworkPolicies

## Security Considerations

1. **Defense in Depth**: Use NetworkPolicies alongside other security measures
2. **Regular Audits**: Periodically review and update policies
3. **Egress Control**: Don't forget to restrict outbound traffic
4. **Compliance**: Ensure policies meet regulatory requirements
5. **Backup**: Keep profile backups in version control

## Future Enhancements

- Support for Kubernetes 1.21+ EndPort field in all scenarios
- Enhanced service mesh policy generation
- Network policy visualization
- Automated policy recommendations based on traffic analysis
- Integration with admission controllers
- Support for GlobalNetworkPolicies (Calico CRDs)