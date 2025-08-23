// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// NetworkPolicyManifest represents a Kubernetes NetworkPolicy in YAML format.
type NetworkPolicyManifest struct {
	APIVersion string                    `yaml:"apiVersion"`
	Kind       string                    `yaml:"kind"`
	Metadata   NetworkPolicyMetadata     `yaml:"metadata"`
	Spec       NetworkPolicySpecManifest `yaml:"spec"`
}

// NetworkPolicyMetadata represents metadata for NetworkPolicy.
type NetworkPolicyMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// NetworkPolicySpecManifest represents the spec section of NetworkPolicy.
type NetworkPolicySpecManifest struct {
	PodSelector LabelSelector              `yaml:"podSelector"`
	PolicyTypes []string                   `yaml:"policyTypes,omitempty"`
	Ingress     []NetworkPolicyIngressSpec `yaml:"ingress,omitempty"`
	Egress      []NetworkPolicyEgressSpec  `yaml:"egress,omitempty"`
}

// LabelSelector represents a label selector.
type LabelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

// NetworkPolicyIngressSpec represents ingress rules in manifest format.
type NetworkPolicyIngressSpec struct {
	From  []NetworkPolicyPeerSpec `yaml:"from,omitempty"`
	Ports []NetworkPolicyPortSpec `yaml:"ports,omitempty"`
}

// NetworkPolicyEgressSpec represents egress rules in manifest format.
type NetworkPolicyEgressSpec struct {
	To    []NetworkPolicyPeerSpec `yaml:"to,omitempty"`
	Ports []NetworkPolicyPortSpec `yaml:"ports,omitempty"`
}

// NetworkPolicyPeerSpec represents a peer in manifest format.
type NetworkPolicyPeerSpec struct {
	PodSelector       *LabelSelector `yaml:"podSelector,omitempty"`
	NamespaceSelector *LabelSelector `yaml:"namespaceSelector,omitempty"`
	IPBlock           *IPBlockSpec   `yaml:"ipBlock,omitempty"`
}

// IPBlockSpec represents IP block in manifest format.
type IPBlockSpec struct {
	CIDR   string   `yaml:"cidr"`
	Except []string `yaml:"except,omitempty"`
}

// NetworkPolicyPortSpec represents port configuration in manifest format.
type NetworkPolicyPortSpec struct {
	Protocol string      `yaml:"protocol,omitempty"`
	Port     interface{} `yaml:"port,omitempty"`
	EndPort  *int32      `yaml:"endPort,omitempty"`
}

// GenerateNetworkPolicy generates a Kubernetes NetworkPolicy from configuration.
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
		rule := NetworkPolicyIngressSpec{
			From:  km.convertNetworkPolicyPeers(ingressRule.From),
			Ports: km.convertNetworkPolicyPorts(ingressRule.Ports),
		}
		policy.Spec.Ingress = append(policy.Spec.Ingress, rule)
	}

	// Convert egress rules
	for _, egressRule := range config.Egress {
		rule := NetworkPolicyEgressSpec{
			To:    km.convertNetworkPolicyPeers(egressRule.To),
			Ports: km.convertNetworkPolicyPorts(egressRule.Ports),
		}
		policy.Spec.Egress = append(policy.Spec.Egress, rule)
	}

	return policy, nil
}

// convertNetworkPolicyPeers converts configuration peers to manifest format.
func (km *KubernetesNetworkManager) convertNetworkPolicyPeers(peers []NetworkPolicyPeer) []NetworkPolicyPeerSpec {
	result := make([]NetworkPolicyPeerSpec, 0, len(peers))
	for _, peer := range peers {
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

		result = append(result, policyPeer)
	}
	return result
}

// convertNetworkPolicyPorts converts configuration ports to manifest format.
func (km *KubernetesNetworkManager) convertNetworkPolicyPorts(ports []NetworkPolicyPort) []NetworkPolicyPortSpec {
	result := make([]NetworkPolicyPortSpec, 0, len(ports))
	for _, port := range ports {
		policyPort := NetworkPolicyPortSpec{
			Protocol: port.Protocol,
		}

		if port.Port != nil {
			policyPort.Port = *port.Port
		}

		if port.EndPort != nil {
			policyPort.EndPort = port.EndPort
		}

		result = append(result, policyPort)
	}
	return result
}

// GetNamespaceNetworkPolicies gets all network policies in a namespace.
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

// ExportNamespacePolicies exports existing namespace policies to a profile.
func (km *KubernetesNetworkManager) ExportNamespacePolicies(namespace, profileName string) error {
	policies, err := km.GetNamespaceNetworkPolicies(namespace)
	if err != nil {
		return fmt.Errorf("failed to get namespace policies: %w", err)
	}

	profile := km.createNetworkProfile(namespace, profileName, len(policies))

	// Convert Kubernetes NetworkPolicies to our config format
	for _, policyData := range policies {
		config, err := km.convertPolicyToConfig(policyData)
		if err != nil {
			continue // Skip invalid policies
		}
		profile.Policies[config.Name] = config
	}

	return km.CreateProfile(profile)
}

// createNetworkProfile creates a new network profile with basic metadata.
func (km *KubernetesNetworkManager) createNetworkProfile(namespace, profileName string, policyCount int) *KubernetesNetworkProfile {
	profile := &KubernetesNetworkProfile{
		Name:        profileName,
		Description: fmt.Sprintf("Exported from namespace %s", namespace),
		Namespace:   namespace,
		Policies:    make(map[string]*NetworkPolicyConfig),
		Metadata:    make(map[string]string),
	}

	profile.Metadata["exported_at"] = time.Now().Format(time.RFC3339)
	profile.Metadata["total_policies"] = fmt.Sprintf("%d", policyCount)

	return profile
}

// convertPolicyToConfig converts a Kubernetes NetworkPolicy to our internal config format.
func (km *KubernetesNetworkManager) convertPolicyToConfig(policyData map[string]interface{}) (*NetworkPolicyConfig, error) {
	metadata, ok := policyData["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid metadata")
	}

	name, ok := metadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid policy name")
	}

	spec, ok := policyData["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid policy spec")
	}

	config := &NetworkPolicyConfig{
		Name:        name,
		PolicyTypes: make([]string, 0),
	}

	// Extract pod selector
	km.extractPodSelector(spec, config)

	// Extract policy types
	km.extractPolicyTypes(spec, config)

	// Extract ingress rules
	km.extractIngressRules(spec, config)

	// Extract egress rules
	km.extractEgressRules(spec, config)

	return config, nil
}

// extractPodSelector extracts pod selector from policy spec.
func (km *KubernetesNetworkManager) extractPodSelector(spec map[string]interface{}, config *NetworkPolicyConfig) {
	if podSelector, ok := spec["podSelector"].(map[string]interface{}); ok {
		config.PodSelector = km.extractMatchLabels(podSelector)
	}
}

// extractPolicyTypes extracts policy types from policy spec.
func (km *KubernetesNetworkManager) extractPolicyTypes(spec map[string]interface{}, config *NetworkPolicyConfig) {
	if policyTypes, ok := spec["policyTypes"].([]interface{}); ok {
		for _, pt := range policyTypes {
			if policyType, ok := pt.(string); ok {
				config.PolicyTypes = append(config.PolicyTypes, policyType)
			}
		}
	}
}

// extractIngressRules extracts ingress rules from policy spec.
func (km *KubernetesNetworkManager) extractIngressRules(spec map[string]interface{}, config *NetworkPolicyConfig) {
	if ingressRules, ok := spec["ingress"].([]interface{}); ok {
		config.Ingress = make([]NetworkPolicyIngressRule, 0)
		for _, rule := range ingressRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				ingressRule := km.parseIngressRule(ruleMap)
				config.Ingress = append(config.Ingress, ingressRule)
			}
		}
	}
}

// extractEgressRules extracts egress rules from policy spec.
func (km *KubernetesNetworkManager) extractEgressRules(spec map[string]interface{}, config *NetworkPolicyConfig) {
	if egressRules, ok := spec["egress"].([]interface{}); ok {
		config.Egress = make([]NetworkPolicyEgressRule, 0)
		for _, rule := range egressRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				egressRule := km.parseEgressRule(ruleMap)
				config.Egress = append(config.Egress, egressRule)
			}
		}
	}
}

// parseIngressRule parses a single ingress rule from rule map.
func (km *KubernetesNetworkManager) parseIngressRule(ruleMap map[string]interface{}) NetworkPolicyIngressRule {
	ingressRule := NetworkPolicyIngressRule{}

	// Extract from peers
	if fromPeers, ok := ruleMap["from"].([]interface{}); ok {
		for _, peer := range fromPeers {
			if peerMap, ok := peer.(map[string]interface{}); ok {
				networkPeer := km.parseNetworkPeer(peerMap)
				ingressRule.From = append(ingressRule.From, networkPeer)
			}
		}
	}

	// Extract ports
	if ports, ok := ruleMap["ports"].([]interface{}); ok {
		ingressRule.Ports = km.parseNetworkPorts(ports)
	}

	return ingressRule
}

// parseEgressRule parses a single egress rule from rule map.
func (km *KubernetesNetworkManager) parseEgressRule(ruleMap map[string]interface{}) NetworkPolicyEgressRule {
	egressRule := NetworkPolicyEgressRule{}

	// Extract to peers
	if toPeers, ok := ruleMap["to"].([]interface{}); ok {
		for _, peer := range toPeers {
			if peerMap, ok := peer.(map[string]interface{}); ok {
				networkPeer := km.parseNetworkPeer(peerMap)
				egressRule.To = append(egressRule.To, networkPeer)
			}
		}
	}

	// Extract ports
	if ports, ok := ruleMap["ports"].([]interface{}); ok {
		egressRule.Ports = km.parseNetworkPorts(ports)
	}

	return egressRule
}

// parseNetworkPeer parses a network peer from peer map.
func (km *KubernetesNetworkManager) parseNetworkPeer(peerMap map[string]interface{}) NetworkPolicyPeer {
	networkPeer := NetworkPolicyPeer{}

	// Pod selector
	if ps, ok := peerMap["podSelector"].(map[string]interface{}); ok {
		networkPeer.PodSelector = km.extractMatchLabels(ps)
	}

	// Namespace selector
	if ns, ok := peerMap["namespaceSelector"].(map[string]interface{}); ok {
		networkPeer.NamespaceSelector = km.extractMatchLabels(ns)
	}

	// IP block
	if ipBlock, ok := peerMap["ipBlock"].(map[string]interface{}); ok {
		networkPeer.IPBlock = km.parseIPBlock(ipBlock)
	}

	return networkPeer
}

// parseIPBlock parses an IP block from ipBlock map.
func (km *KubernetesNetworkManager) parseIPBlock(ipBlock map[string]interface{}) *IPBlock {
	block := &IPBlock{}

	if cidr, ok := ipBlock["cidr"].(string); ok {
		block.CIDR = cidr
	}

	if except, ok := ipBlock["except"].([]interface{}); ok {
		for _, e := range except {
			if exceptCIDR, ok := e.(string); ok {
				block.Except = append(block.Except, exceptCIDR)
			}
		}
	}

	return block
}

// parseNetworkPorts parses network ports from ports interface slice.
func (km *KubernetesNetworkManager) parseNetworkPorts(ports []interface{}) []NetworkPolicyPort {
	networkPorts := make([]NetworkPolicyPort, 0, len(ports))

	for _, port := range ports {
		portMap, ok := port.(map[string]interface{})
		if !ok {
			continue
		}

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

		networkPorts = append(networkPorts, networkPort)
	}

	return networkPorts
}

// extractMatchLabels extracts match labels from a selector map.
func (km *KubernetesNetworkManager) extractMatchLabels(selector map[string]interface{}) map[string]string {
	labels := make(map[string]string)

	if matchLabels, ok := selector["matchLabels"].(map[string]interface{}); ok {
		for k, v := range matchLabels {
			labels[k] = fmt.Sprintf("%v", v)
		}
	}

	return labels
}
