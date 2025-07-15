package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// NetworkPolicyManifest represents a Kubernetes NetworkPolicy in YAML format
type NetworkPolicyManifest struct {
	APIVersion string                    `yaml:"apiVersion"`
	Kind       string                    `yaml:"kind"`
	Metadata   NetworkPolicyMetadata     `yaml:"metadata"`
	Spec       NetworkPolicySpecManifest `yaml:"spec"`
}

// NetworkPolicyMetadata represents metadata for NetworkPolicy
type NetworkPolicyMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// NetworkPolicySpecManifest represents the spec section of NetworkPolicy
type NetworkPolicySpecManifest struct {
	PodSelector LabelSelector              `yaml:"podSelector"`
	PolicyTypes []string                   `yaml:"policyTypes,omitempty"`
	Ingress     []NetworkPolicyIngressSpec `yaml:"ingress,omitempty"`
	Egress      []NetworkPolicyEgressSpec  `yaml:"egress,omitempty"`
}

// LabelSelector represents a label selector
type LabelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

// NetworkPolicyIngressSpec represents ingress rules in manifest format
type NetworkPolicyIngressSpec struct {
	From  []NetworkPolicyPeerSpec `yaml:"from,omitempty"`
	Ports []NetworkPolicyPortSpec `yaml:"ports,omitempty"`
}

// NetworkPolicyEgressSpec represents egress rules in manifest format
type NetworkPolicyEgressSpec struct {
	To    []NetworkPolicyPeerSpec `yaml:"to,omitempty"`
	Ports []NetworkPolicyPortSpec `yaml:"ports,omitempty"`
}

// NetworkPolicyPeerSpec represents a peer in manifest format
type NetworkPolicyPeerSpec struct {
	PodSelector       *LabelSelector `yaml:"podSelector,omitempty"`
	NamespaceSelector *LabelSelector `yaml:"namespaceSelector,omitempty"`
	IPBlock           *IPBlockSpec   `yaml:"ipBlock,omitempty"`
}

// IPBlockSpec represents IP block in manifest format
type IPBlockSpec struct {
	CIDR   string   `yaml:"cidr"`
	Except []string `yaml:"except,omitempty"`
}

// NetworkPolicyPortSpec represents port configuration in manifest format
type NetworkPolicyPortSpec struct {
	Protocol string      `yaml:"protocol,omitempty"`
	Port     interface{} `yaml:"port,omitempty"`
	EndPort  *int32      `yaml:"endPort,omitempty"`
}

// GenerateNetworkPolicy generates a Kubernetes NetworkPolicy from configuration
func (km *KubernetesNetworkManager) GenerateNetworkPolicy(namespace string, config *NetworkPolicyConfig) (*NetworkPolicyManifest, error) {
	policy := &NetworkPolicyManifest{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
		Metadata: NetworkPolicyMetadata{
			Name:      config.Name,
			Namespace: namespace,
		},
		Spec: NetworkPolicySpecManifest{
			PodSelector: LabelSelector{
				MatchLabels: config.PodSelector,
			},
			PolicyTypes: config.PolicyTypes,
		},
	}

	// Convert ingress rules
	for _, ingressRule := range config.Ingress {
		rule := NetworkPolicyIngressSpec{}

		// Convert peers
		for _, peer := range ingressRule.From {
			policyPeer := NetworkPolicyPeerSpec{}

			if len(peer.PodSelector) > 0 {
				policyPeer.PodSelector = &LabelSelector{
					MatchLabels: peer.PodSelector,
				}
			}

			if len(peer.NamespaceSelector) > 0 {
				policyPeer.NamespaceSelector = &LabelSelector{
					MatchLabels: peer.NamespaceSelector,
				}
			}

			if peer.IPBlock != nil {
				policyPeer.IPBlock = &IPBlockSpec{
					CIDR:   peer.IPBlock.CIDR,
					Except: peer.IPBlock.Except,
				}
			}

			rule.From = append(rule.From, policyPeer)
		}

		// Convert ports
		for _, port := range ingressRule.Ports {
			policyPort := NetworkPolicyPortSpec{
				Protocol: port.Protocol,
			}

			if port.Port != nil {
				policyPort.Port = *port.Port
			}

			if port.EndPort != nil {
				policyPort.EndPort = port.EndPort
			}

			rule.Ports = append(rule.Ports, policyPort)
		}

		policy.Spec.Ingress = append(policy.Spec.Ingress, rule)
	}

	// Convert egress rules
	for _, egressRule := range config.Egress {
		rule := NetworkPolicyEgressSpec{}

		// Convert peers
		for _, peer := range egressRule.To {
			policyPeer := NetworkPolicyPeerSpec{}

			if len(peer.PodSelector) > 0 {
				policyPeer.PodSelector = &LabelSelector{
					MatchLabels: peer.PodSelector,
				}
			}

			if len(peer.NamespaceSelector) > 0 {
				policyPeer.NamespaceSelector = &LabelSelector{
					MatchLabels: peer.NamespaceSelector,
				}
			}

			if peer.IPBlock != nil {
				policyPeer.IPBlock = &IPBlockSpec{
					CIDR:   peer.IPBlock.CIDR,
					Except: peer.IPBlock.Except,
				}
			}

			rule.To = append(rule.To, policyPeer)
		}

		// Convert ports
		for _, port := range egressRule.Ports {
			policyPort := NetworkPolicyPortSpec{
				Protocol: port.Protocol,
			}

			if port.Port != nil {
				policyPort.Port = *port.Port
			}

			if port.EndPort != nil {
				policyPort.EndPort = port.EndPort
			}

			rule.Ports = append(rule.Ports, policyPort)
		}

		policy.Spec.Egress = append(policy.Spec.Egress, rule)
	}

	return policy, nil
}

// // ExecuteWithTimeout executes a kubectl command with timeout
// func (executor *KubernetesCommandExecutor) ExecuteWithTimeout(ctx context.Context, command string, timeout time.Duration) (*KubernetesCommandResult, error) {
// 	// Check cache first (for read-only commands)
// 	if strings.Contains(command, "get") && !strings.Contains(command, "watch") {
// 		if cached := executor.getCachedResult(command); cached != nil {
// 			return cached, nil
// 		}
// 	}
// 
// 	// Create context with timeout
// 	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
// 	defer cancel()
// 
// 	// Parse command
// 	parts := strings.Fields(command)
// 	if len(parts) == 0 {
// 		return nil, fmt.Errorf("empty command")
// 	}
// 
// 	start := time.Now()
// 	cmd := exec.CommandContext(timeoutCtx, parts[0], parts[1:]...)
// 	output, err := cmd.CombinedOutput()
// 	duration := time.Since(start)
// 
// 	result := &KubernetesCommandResult{
// 		Output:   string(output),
// 		Duration: duration,
// 		CachedAt: time.Now(),
// 	}
// 
// 	if err != nil {
// 		result.Error = err.Error()
// 		if exitError, ok := err.(*exec.ExitError); ok {
// 			result.ExitCode = exitError.ExitCode()
// 		} else {
// 			result.ExitCode = 1
// 		}
// 	}
// 
// 	// Cache read-only command results for 30 seconds
// 	if strings.Contains(command, "get") && !strings.Contains(command, "watch") {
// 		executor.setCachedResult(command, result)
// 	}
// 
// 	return result, nil
// }
// 
// GetNamespaceNetworkPolicies gets all network policies in a namespace
func (km *KubernetesNetworkManager) GetNamespaceNetworkPolicies(namespace string) ([]map[string]interface{}, error) {
	cmd := fmt.Sprintf("kubectl get networkpolicies -n %s -o json", namespace)
	result, err := km.executor.ExecuteWithTimeout(context.Background(), cmd, 10*time.Second)

	if err != nil || result.ExitCode != 0 {
		return nil, fmt.Errorf("failed to get network policies: %w", err)
	}

	var policyList map[string]interface{}
	if err := json.Unmarshal([]byte(result.Output), &policyList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network policies: %w", err)
	}

	if items, ok := policyList["items"].([]interface{}); ok {
		policies := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			if policy, ok := item.(map[string]interface{}); ok {
				policies = append(policies, policy)
			}
		}
		return policies, nil
	}

	return []map[string]interface{}{}, nil
}

// ExportNamespacePolicies exports existing namespace policies to a profile
func (km *KubernetesNetworkManager) ExportNamespacePolicies(namespace, profileName string) error {
	policies, err := km.GetNamespaceNetworkPolicies(namespace)
	if err != nil {
		return fmt.Errorf("failed to get namespace policies: %w", err)
	}

	profile := &KubernetesNetworkProfile{
		Name:        profileName,
		Description: fmt.Sprintf("Exported from namespace %s", namespace),
		Namespace:   namespace,
		Policies:    make(map[string]*NetworkPolicyConfig),
		Metadata:    make(map[string]string),
	}

	// Convert Kubernetes NetworkPolicies to our config format
	for _, policyData := range policies {
		metadata, _ := policyData["metadata"].(map[string]interface{})
		name, _ := metadata["name"].(string)

		spec, _ := policyData["spec"].(map[string]interface{})

		config := &NetworkPolicyConfig{
			Name:        name,
			PolicyTypes: make([]string, 0),
		}

		// Extract pod selector
		if podSelector, ok := spec["podSelector"].(map[string]interface{}); ok {
			if matchLabels, ok := podSelector["matchLabels"].(map[string]interface{}); ok {
				config.PodSelector = make(map[string]string)
				for k, v := range matchLabels {
					config.PodSelector[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		// Extract policy types
		if policyTypes, ok := spec["policyTypes"].([]interface{}); ok {
			for _, pt := range policyTypes {
				if policyType, ok := pt.(string); ok {
					config.PolicyTypes = append(config.PolicyTypes, policyType)
				}
			}
		}

		// Extract ingress rules
		if ingressRules, ok := spec["ingress"].([]interface{}); ok {
			config.Ingress = make([]NetworkPolicyIngressRule, 0)
			for _, rule := range ingressRules {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					ingressRule := NetworkPolicyIngressRule{}

					// Extract from peers
					if fromPeers, ok := ruleMap["from"].([]interface{}); ok {
						for _, peer := range fromPeers {
							if peerMap, ok := peer.(map[string]interface{}); ok {
								networkPeer := NetworkPolicyPeer{}

								// Pod selector
								if ps, ok := peerMap["podSelector"].(map[string]interface{}); ok {
									if matchLabels, ok := ps["matchLabels"].(map[string]interface{}); ok {
										networkPeer.PodSelector = make(map[string]string)
										for k, v := range matchLabels {
											networkPeer.PodSelector[k] = fmt.Sprintf("%v", v)
										}
									}
								}

								// Namespace selector
								if ns, ok := peerMap["namespaceSelector"].(map[string]interface{}); ok {
									if matchLabels, ok := ns["matchLabels"].(map[string]interface{}); ok {
										networkPeer.NamespaceSelector = make(map[string]string)
										for k, v := range matchLabels {
											networkPeer.NamespaceSelector[k] = fmt.Sprintf("%v", v)
										}
									}
								}

								// IP block
								if ipBlock, ok := peerMap["ipBlock"].(map[string]interface{}); ok {
									networkPeer.IPBlock = &IPBlock{}
									if cidr, ok := ipBlock["cidr"].(string); ok {
										networkPeer.IPBlock.CIDR = cidr
									}
									if except, ok := ipBlock["except"].([]interface{}); ok {
										for _, e := range except {
											if exceptCIDR, ok := e.(string); ok {
												networkPeer.IPBlock.Except = append(networkPeer.IPBlock.Except, exceptCIDR)
											}
										}
									}
								}

								ingressRule.From = append(ingressRule.From, networkPeer)
							}
						}
					}

					// Extract ports
					if ports, ok := ruleMap["ports"].([]interface{}); ok {
						for _, port := range ports {
							if portMap, ok := port.(map[string]interface{}); ok {
								networkPort := NetworkPolicyPort{}

								if protocol, ok := portMap["protocol"].(string); ok {
									networkPort.Protocol = protocol
								}

								if portNum, ok := portMap["port"].(float64); ok {
									p := int32(portNum)
									networkPort.Port = &p
								}

								if endPort, ok := portMap["endPort"].(float64); ok {
									ep := int32(endPort)
									networkPort.EndPort = &ep
								}

								ingressRule.Ports = append(ingressRule.Ports, networkPort)
							}
						}
					}

					config.Ingress = append(config.Ingress, ingressRule)
				}
			}
		}

		// Extract egress rules (similar to ingress)
		if egressRules, ok := spec["egress"].([]interface{}); ok {
			config.Egress = make([]NetworkPolicyEgressRule, 0)
			// Similar parsing logic as ingress
		}

		profile.Policies[name] = config
	}

	profile.Metadata["exported_at"] = time.Now().Format(time.RFC3339)
	profile.Metadata["total_policies"] = fmt.Sprintf("%d", len(policies))

	return km.CreateProfile(profile)
}
