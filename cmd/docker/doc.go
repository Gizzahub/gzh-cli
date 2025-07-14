// Package docker provides Docker container management and integration
// functionality for the GZH Manager CLI tool.
//
// This package implements comprehensive Docker operations including
// container lifecycle management, image handling, network configuration,
// and development workflow integration.
//
// Key Components:
//
// Container Management:
//   - Container creation, startup, and shutdown
//   - Container health monitoring and status
//   - Container networking and port management
//   - Volume and storage management
//
// Image Operations:
//   - Docker image building and tagging
//   - Image registry integration
//   - Multi-stage build optimization
//   - Image security scanning
//
// Development Integration:
//   - Development container setup
//   - Hot reload and live development
//   - Debug container configuration
//   - Test environment containerization
//
// Compose Integration:
//   - Docker Compose file management
//   - Multi-service orchestration
//   - Service dependency management
//   - Environment-specific configurations
//
// Features:
//   - Cross-platform Docker support
//   - Docker Desktop integration
//   - Container resource optimization
//   - Security best practices enforcement
//   - Performance monitoring and metrics
//
// Example usage:
//
//	manager := docker.NewManager()
//
//	container := docker.NewContainer(config)
//	err := manager.StartContainer(container)
//
//	image := docker.NewImage(dockerfile)
//	err = manager.BuildImage(image)
//
// The package provides seamless Docker integration for development
// workflows and container-based deployments.
package docker
