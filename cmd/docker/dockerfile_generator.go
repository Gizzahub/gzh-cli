package docker

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// DockerfileCmd represents the dockerfile command.
var DockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate optimized Dockerfile for projects",
	Long: `Generate optimized multi-stage Dockerfile templates based on project language and requirements.

Supports automatic detection of project language and generates appropriate Dockerfile with:
- Multi-stage builds for optimal image size
- Language-specific optimizations
- Security best practices
- Built-in security scanning integration`,
	Run: runDockerfileGenerate,
}

var (
	outputPath       string
	projectLanguage  string
	projectName      string
	baseImage        string
	includeScanning  bool
	includeMultiArch bool
	production       bool
)

func init() {
	DockerfileCmd.Flags().StringVarP(&outputPath, "output", "o", "./Dockerfile", "Output path for generated Dockerfile")
	DockerfileCmd.Flags().StringVarP(&projectLanguage, "language", "l", "", "Project language (auto-detect if not specified)")
	DockerfileCmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name (auto-detect from directory if not specified)")
	DockerfileCmd.Flags().StringVarP(&baseImage, "base", "b", "", "Custom base image (uses language defaults if not specified)")
	DockerfileCmd.Flags().BoolVar(&includeScanning, "security-scan", true, "Include security scanning in Dockerfile")
	DockerfileCmd.Flags().BoolVar(&includeMultiArch, "multi-arch", false, "Generate multi-architecture build support")
	DockerfileCmd.Flags().BoolVar(&production, "production", true, "Generate production-ready Dockerfile")
}

// ProjectInfo holds detected project information.
type ProjectInfo struct {
	Language        string
	Name            string
	BaseImage       string
	HasGoMod        bool
	HasPackageJSON  bool
	HasRequirements bool
	HasGemfile      bool
	HasCargoToml    bool
	HasPomXML       bool
	HasBuildGradle  bool
	Framework       string
	Port            int
}

// DockerfileTemplate holds template data.
type DockerfileTemplate struct {
	Project          ProjectInfo
	IncludeScanning  bool
	IncludeMultiArch bool
	Production       bool
	SecurityTools    []string
	BuildArgs        []string
	HealthCheck      string
	User             string
	Workdir          string
}

func runDockerfileGenerate(cmd *cobra.Command, args []string) {
	// Auto-detect project information
	projectInfo, err := detectProjectInfo()
	if err != nil {
		fmt.Printf("Error detecting project info: %v\n", err)
		os.Exit(1)
	}

	// Override with user-specified values
	if projectLanguage != "" {
		projectInfo.Language = projectLanguage
	}

	if projectName != "" {
		projectInfo.Name = projectName
	}

	if baseImage != "" {
		projectInfo.BaseImage = baseImage
	}

	// Validate detected/specified language
	if projectInfo.Language == "" {
		fmt.Println("Could not detect project language. Please specify with --language flag.")
		fmt.Println("Supported languages: go, node, python, ruby, rust, java")
		os.Exit(1)
	}

	// Set default base image if not specified
	if projectInfo.BaseImage == "" {
		projectInfo.BaseImage = getDefaultBaseImage(projectInfo.Language)
	}

	// Generate Dockerfile
	templateData := DockerfileTemplate{
		Project:          projectInfo,
		IncludeScanning:  includeScanning,
		IncludeMultiArch: includeMultiArch,
		Production:       production,
		SecurityTools:    getSecurityTools(),
		BuildArgs:        getBuildArgs(projectInfo),
		HealthCheck:      getHealthCheck(projectInfo),
		User:             getNonRootUser(projectInfo.Language),
		Workdir:          "/app",
	}

	dockerfile, err := generateDockerfile(templateData)
	if err != nil {
		fmt.Printf("Error generating Dockerfile: %v\n", err)
		os.Exit(1)
	}

	// Write Dockerfile
	err = os.WriteFile(outputPath, []byte(dockerfile), 0o644)
	if err != nil {
		fmt.Printf("Error writing Dockerfile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated optimized Dockerfile: %s\n", outputPath)
	fmt.Printf("📋 Language: %s\n", projectInfo.Language)
	fmt.Printf("📋 Base image: %s\n", projectInfo.BaseImage)
	fmt.Printf("📋 Security scanning: %v\n", includeScanning)
	fmt.Printf("📋 Multi-architecture: %v\n", includeMultiArch)

	// Generate .dockerignore if it doesn't exist
	if err := generateDockerignore(projectInfo); err != nil {
		fmt.Printf("Warning: Could not generate .dockerignore: %v\n", err)
	} else {
		fmt.Println("✅ Generated .dockerignore file")
	}
}

func detectProjectInfo() (ProjectInfo, error) {
	info := ProjectInfo{
		Name: filepath.Base(getCurrentDir()),
		Port: 8080, // default port
	}

	// Check for language-specific files
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // ignore errors
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		switch d.Name() {
		case "go.mod":
			info.Language = "go"
			info.HasGoMod = true
		case "package.json":
			info.Language = "node"
			info.HasPackageJSON = true
			info.Port = 3000 // common Node.js port
		case "requirements.txt", "pyproject.toml", "Pipfile":
			info.Language = "python"
			info.HasRequirements = true
		case "Gemfile":
			info.Language = "ruby"
			info.HasGemfile = true
		case "Cargo.toml":
			info.Language = "rust"
			info.HasCargoToml = true
		case "pom.xml":
			info.Language = "java"
			info.HasPomXML = true
		case "build.gradle", "build.gradle.kts":
			info.Language = "java"
			info.HasBuildGradle = true
		}

		return nil
	})

	return info, err
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "app"
	}

	return dir
}

func getDefaultBaseImage(language string) string {
	switch language {
	case "go":
		return "golang:1.21-alpine"
	case "node":
		return "node:20-alpine"
	case "python":
		return "python:3.11-slim"
	case "ruby":
		return "ruby:3.2-alpine"
	case "rust":
		return "rust:1.70-alpine"
	case "java":
		return "openjdk:21-jdk-slim"
	default:
		return "alpine:latest"
	}
}

func getSecurityTools() []string {
	tools := []string{}
	if includeScanning {
		tools = append(tools, "trivy", "grype")
	}

	return tools
}

func getBuildArgs(info ProjectInfo) []string {
	args := []string{
		"BUILDPLATFORM",
		"TARGETPLATFORM",
		"TARGETOS",
		"TARGETARCH",
	}

	switch info.Language {
	case "go":
		args = append(args, "CGO_ENABLED=0", "GOOS=${TARGETOS}", "GOARCH=${TARGETARCH}")
	case "node":
		args = append(args, "NODE_ENV=production")
	case "python":
		args = append(args, "PYTHONUNBUFFERED=1", "PYTHONDONTWRITEBYTECODE=1")
	}

	return args
}

func getHealthCheck(info ProjectInfo) string {
	switch info.Language {
	case "go", "node", "python", "ruby", "java":
		return fmt.Sprintf("HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\\n  CMD curl -f http://localhost:%d/health || exit 1", info.Port)
	default:
		return "# Add appropriate health check for your application"
	}
}

func getNonRootUser(language string) string {
	switch language {
	case "go":
		return "nobody"
	case "node":
		return "node"
	case "python":
		return "python"
	case "ruby":
		return "ruby"
	case "java":
		return "java"
	default:
		return "app"
	}
}

func generateDockerfile(data DockerfileTemplate) (string, error) {
	templateStr := getDockerfileTemplate(data.Project.Language)

	tmpl, err := template.New("dockerfile").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func getDockerfileTemplate(language string) string {
	switch language {
	case "go":
		return goDockerfileTemplate
	case "node":
		return nodeDockerfileTemplate
	case "python":
		return pythonDockerfileTemplate
	case "ruby":
		return rubyDockerfileTemplate
	case "rust":
		return rustDockerfileTemplate
	case "java":
		return javaDockerfileTemplate
	default:
		return genericDockerfileTemplate
	}
}

func generateDockerignore(info ProjectInfo) error {
	dockerignorePath := ".dockerignore"

	// Check if .dockerignore already exists
	if _, err := os.Stat(dockerignorePath); err == nil {
		return nil // file exists, don't overwrite
	}

	content := getDockerignoreTemplate(info.Language)

	return os.WriteFile(dockerignorePath, []byte(content), 0o644)
}
