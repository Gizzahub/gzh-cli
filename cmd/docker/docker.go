package docker

import (
	"github.com/spf13/cobra"
)

// DockerCmd represents the docker command
var DockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker-related commands for containerization",
	Long: `Docker-related commands for generating optimized Dockerfiles, 
managing container builds, and integrating security scanning.

Available commands:
  dockerfile  Generate optimized multi-stage Dockerfile for projects`,
}

func init() {
	DockerCmd.AddCommand(DockerfileCmd)
}
