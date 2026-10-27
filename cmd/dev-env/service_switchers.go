// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/gizzahub/gzh-cli-dev-env/pkg/azure"
	"github.com/gizzahub/gzh-cli-dev-env/pkg/docker"
	"github.com/gizzahub/gzh-cli-dev-env/pkg/environment"
	"github.com/gizzahub/gzh-cli-dev-env/pkg/gcp"
	"github.com/gizzahub/gzh-cli-dev-env/pkg/kubernetes"
	"github.com/gizzahub/gzh-cli-dev-env/pkg/ssh"
)

// AWSSwitcher implements ServiceSwitcher for AWS.
type AWSSwitcher struct{}

func (a *AWSSwitcher) Name() string {
	return "aws"
}

func (a *AWSSwitcher) Switch(ctx context.Context, config any) error {
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

func (a *AWSSwitcher) GetCurrentState(ctx context.Context) (any, error) {
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

func (a *AWSSwitcher) Rollback(ctx context.Context, previousState any) error {
	return a.Switch(ctx, previousState)
}

// GCPSwitcher implements ServiceSwitcher for GCP.
type GCPSwitcher struct{}

func (g *GCPSwitcher) Name() string {
	return "gcp"
}

func (g *GCPSwitcher) Switch(ctx context.Context, config any) error {
	return gcp.NewSwitcher().Switch(ctx, config)
}

func (g *GCPSwitcher) GetCurrentState(ctx context.Context) (any, error) {
	return gcp.NewSwitcher().GetCurrentState(ctx)
}

func (g *GCPSwitcher) Rollback(ctx context.Context, previousState any) error {
	return gcp.NewSwitcher().Rollback(ctx, previousState)
}

// AzureSwitcher implements ServiceSwitcher for Azure.
type AzureSwitcher struct{}

func (a *AzureSwitcher) Name() string {
	return "azure"
}

func (a *AzureSwitcher) Switch(ctx context.Context, config any) error {
	return azure.NewSwitcher().Switch(ctx, config)
}

func (a *AzureSwitcher) GetCurrentState(ctx context.Context) (any, error) {
	return azure.NewSwitcher().GetCurrentState(ctx)
}

func (a *AzureSwitcher) Rollback(ctx context.Context, previousState any) error {
	return azure.NewSwitcher().Rollback(ctx, previousState)
}

// DockerSwitcher implements ServiceSwitcher for Docker.
type DockerSwitcher struct{}

func (d *DockerSwitcher) Name() string {
	return "docker"
}

func (d *DockerSwitcher) Switch(ctx context.Context, config any) error {
	return docker.NewSwitcher().Switch(ctx, config)
}

func (d *DockerSwitcher) GetCurrentState(ctx context.Context) (any, error) {
	return docker.NewSwitcher().GetCurrentState(ctx)
}

func (d *DockerSwitcher) Rollback(ctx context.Context, previousState any) error {
	return docker.NewSwitcher().Rollback(ctx, previousState)
}

// KubernetesSwitcher implements ServiceSwitcher for Kubernetes.
type KubernetesSwitcher struct{}

func (k *KubernetesSwitcher) Name() string {
	return "kubernetes"
}

func (k *KubernetesSwitcher) Switch(ctx context.Context, config any) error {
	return kubernetes.NewSwitcher().Switch(ctx, config)
}

func (k *KubernetesSwitcher) GetCurrentState(ctx context.Context) (any, error) {
	return kubernetes.NewSwitcher().GetCurrentState(ctx)
}

func (k *KubernetesSwitcher) Rollback(ctx context.Context, previousState any) error {
	return kubernetes.NewSwitcher().Rollback(ctx, previousState)
}

// SSHSwitcher implements ServiceSwitcher for SSH.
type SSHSwitcher struct{}

func (s *SSHSwitcher) Name() string {
	return "ssh"
}

func (s *SSHSwitcher) Switch(ctx context.Context, config any) error {
	return ssh.NewSwitcher().Switch(ctx, config)
}

func (s *SSHSwitcher) GetCurrentState(ctx context.Context) (any, error) {
	return ssh.NewSwitcher().GetCurrentState(ctx)
}

func (s *SSHSwitcher) Rollback(ctx context.Context, previousState any) error {
	return ssh.NewSwitcher().Rollback(ctx, previousState)
}
