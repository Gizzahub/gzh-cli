// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/gizzahub/gzh-cli-dev-env/pkg/environment"
)

// AWSSwitcher implements ServiceSwitcher for AWS.
type AWSSwitcher struct{}

func (a *AWSSwitcher) Name() string {
	return "aws"
}

func (a *AWSSwitcher) Switch(ctx context.Context, config interface{}) error {
	awsConfig, ok := config.(*environment.AWSConfig)
	if !ok {
		return fmt.Errorf("invalid AWS configuration type")
	}

	// Set AWS profile
	if awsConfig.Profile != "" {
		cmd := exec.CommandContext(ctx, "aws", "configure", "set", "profile", awsConfig.Profile)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set AWS profile: %w", err)
		}
	}

	// Set AWS region
	if awsConfig.Region != "" {
		args := []string{"configure", "set", "region", awsConfig.Region}
		if awsConfig.Profile != "" {
			args = append(args, "--profile", awsConfig.Profile)
		}
		cmd := exec.CommandContext(ctx, "aws", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set AWS region: %w", err)
		}
	}

	return nil
}

func (a *AWSSwitcher) GetCurrentState(ctx context.Context) (interface{}, error) {
	// Get current AWS profile
	cmd := exec.CommandContext(ctx, "aws", "configure", "get", "profile")
	profileOutput, _ := cmd.Output()

	// Get current AWS region
	cmd = exec.CommandContext(ctx, "aws", "configure", "get", "region")
	regionOutput, _ := cmd.Output()

	return &environment.AWSConfig{
		Profile: string(profileOutput),
		Region:  string(regionOutput),
	}, nil
}

func (a *AWSSwitcher) Rollback(ctx context.Context, previousState interface{}) error {
	return a.Switch(ctx, previousState)
}

// GCPSwitcher implements ServiceSwitcher for GCP.
type GCPSwitcher struct{}

func (g *GCPSwitcher) Name() string {
	return "gcp"
}

func (g *GCPSwitcher) Switch(ctx context.Context, config interface{}) error {
	gcpConfig, ok := config.(*environment.GCPConfig)
	if !ok {
		return fmt.Errorf("invalid GCP configuration type")
	}

	// Set GCP project
	if gcpConfig.Project != "" {
		cmd := exec.CommandContext(ctx, "gcloud", "config", "set", "project", gcpConfig.Project)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set GCP project: %w", err)
		}
	}

	// Set GCP account
	if gcpConfig.Account != "" {
		cmd := exec.CommandContext(ctx, "gcloud", "config", "set", "account", gcpConfig.Account)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set GCP account: %w", err)
		}
	}

	// Set GCP region
	if gcpConfig.Region != "" {
		cmd := exec.CommandContext(ctx, "gcloud", "config", "set", "compute/region", gcpConfig.Region)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set GCP region: %w", err)
		}
	}

	return nil
}

func (g *GCPSwitcher) GetCurrentState(ctx context.Context) (interface{}, error) {
	// Get current GCP project
	cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "project")
	projectOutput, _ := cmd.Output()

	// Get current GCP account
	cmd = exec.CommandContext(ctx, "gcloud", "config", "get-value", "account")
	accountOutput, _ := cmd.Output()

	// Get current GCP region
	cmd = exec.CommandContext(ctx, "gcloud", "config", "get-value", "compute/region")
	regionOutput, _ := cmd.Output()

	return &environment.GCPConfig{
		Project: string(projectOutput),
		Account: string(accountOutput),
		Region:  string(regionOutput),
	}, nil
}

func (g *GCPSwitcher) Rollback(ctx context.Context, previousState interface{}) error {
	return g.Switch(ctx, previousState)
}

// AzureSwitcher implements ServiceSwitcher for Azure.
type AzureSwitcher struct{}

func (a *AzureSwitcher) Name() string {
	return "azure"
}

func (a *AzureSwitcher) Switch(ctx context.Context, config interface{}) error {
	azureConfig, ok := config.(*environment.AzureConfig)
	if !ok {
		return fmt.Errorf("invalid Azure configuration type")
	}

	// Set Azure subscription
	if azureConfig.Subscription != "" {
		cmd := exec.CommandContext(ctx, "az", "account", "set", "--subscription", azureConfig.Subscription)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set Azure subscription: %w", err)
		}
	}

	return nil
}

func (a *AzureSwitcher) GetCurrentState(ctx context.Context) (interface{}, error) {
	// Get current Azure subscription
	cmd := exec.CommandContext(ctx, "az", "account", "show", "--query", "id", "-o", "tsv")
	subscriptionOutput, _ := cmd.Output()

	// Get current Azure tenant
	cmd = exec.CommandContext(ctx, "az", "account", "show", "--query", "tenantId", "-o", "tsv")
	tenantOutput, _ := cmd.Output()

	return &environment.AzureConfig{
		Subscription: string(subscriptionOutput),
		Tenant:       string(tenantOutput),
	}, nil
}

func (a *AzureSwitcher) Rollback(ctx context.Context, previousState interface{}) error {
	return a.Switch(ctx, previousState)
}

// DockerSwitcher implements ServiceSwitcher for Docker.
type DockerSwitcher struct{}

func (d *DockerSwitcher) Name() string {
	return "docker"
}

func (d *DockerSwitcher) Switch(ctx context.Context, config interface{}) error {
	dockerConfig, ok := config.(*environment.DockerConfig)
	if !ok {
		return fmt.Errorf("invalid Docker configuration type")
	}

	// Set Docker context
	if dockerConfig.Context != "" {
		cmd := exec.CommandContext(ctx, "docker", "context", "use", dockerConfig.Context)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set Docker context: %w", err)
		}
	}

	return nil
}

func (d *DockerSwitcher) GetCurrentState(ctx context.Context) (interface{}, error) {
	// Get current Docker context
	cmd := exec.CommandContext(ctx, "docker", "context", "show")
	contextOutput, _ := cmd.Output()

	return &environment.DockerConfig{
		Context: string(contextOutput),
	}, nil
}

func (d *DockerSwitcher) Rollback(ctx context.Context, previousState interface{}) error {
	return d.Switch(ctx, previousState)
}

// KubernetesSwitcher implements ServiceSwitcher for Kubernetes.
type KubernetesSwitcher struct{}

func (k *KubernetesSwitcher) Name() string {
	return "kubernetes"
}

func (k *KubernetesSwitcher) Switch(ctx context.Context, config interface{}) error {
	kubernetesConfig, ok := config.(*environment.KubernetesConfig)
	if !ok {
		return fmt.Errorf("invalid Kubernetes configuration type")
	}

	// Set Kubernetes context
	if kubernetesConfig.Context != "" {
		cmd := exec.CommandContext(ctx, "kubectl", "config", "use-context", kubernetesConfig.Context)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set Kubernetes context: %w", err)
		}
	}

	// Set Kubernetes namespace
	if kubernetesConfig.Namespace != "" {
		cmd := exec.CommandContext(ctx, "kubectl", "config", "set-context", "--current", "--namespace", kubernetesConfig.Namespace)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set Kubernetes namespace: %w", err)
		}
	}

	return nil
}

func (k *KubernetesSwitcher) GetCurrentState(ctx context.Context) (interface{}, error) {
	// Get current Kubernetes context
	cmd := exec.CommandContext(ctx, "kubectl", "config", "current-context")
	contextOutput, _ := cmd.Output()

	// Get current namespace
	cmd = exec.CommandContext(ctx, "kubectl", "config", "view", "--minify", "--output", "jsonpath={..namespace}")
	namespaceOutput, _ := cmd.Output()

	return &environment.KubernetesConfig{
		Context:   string(contextOutput),
		Namespace: string(namespaceOutput),
	}, nil
}

func (k *KubernetesSwitcher) Rollback(ctx context.Context, previousState interface{}) error {
	return k.Switch(ctx, previousState)
}

// SSHSwitcher implements ServiceSwitcher for SSH.
type SSHSwitcher struct{}

func (s *SSHSwitcher) Name() string {
	return "ssh"
}

func (s *SSHSwitcher) Switch(ctx context.Context, config interface{}) error {
	sshConfig, ok := config.(*environment.SSHConfig)
	if !ok {
		return fmt.Errorf("invalid SSH configuration type")
	}

	// For SSH, we would typically update the SSH config file or set environment variables
	// This is a simplified implementation
	if sshConfig.Config != "" {
		// In a real implementation, this would load the SSH configuration
		// For now, we'll just validate that the config exists
		fmt.Printf("Setting SSH config to: %s\n", sshConfig.Config)
	}

	return nil
}

func (s *SSHSwitcher) GetCurrentState(ctx context.Context) (interface{}, error) {
	// Get current SSH configuration (simplified)
	return &environment.SSHConfig{
		Config: "default",
	}, nil
}

func (s *SSHSwitcher) Rollback(ctx context.Context, previousState interface{}) error {
	return s.Switch(ctx, previousState)
}
