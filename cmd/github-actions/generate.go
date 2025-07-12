package githubactions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate GitHub Actions workflows",
	Long: `Generate GitHub Actions workflows for various project types and use cases.

Supports generation of:
- CI/CD pipelines for multiple languages
- Release automation workflows
- Security scanning workflows
- Multi-platform build workflows
- Deployment workflows
- Reusable actions and workflows

Examples:
  gz github-actions generate --type ci --language go
  gz github-actions generate --type release --with-docker
  gz github-actions generate --type security --enable-codecov`,
	Run: runGenerate,
}

var (
	workflowType   string
	language       string
	outputDir      string
	withDocker     bool
	withKubernetes bool
	withSecurity   bool
	withCodeCov    bool
	withRelease    bool
	platforms      []string
	secrets        []string
	customActions  []string
	workflowName   string
	enableCaching  bool
	parallelJobs   bool
	matrixStrategy bool
)

func init() {
	GenerateCmd.Flags().StringVarP(&workflowType, "type", "t", "ci", "Workflow type (ci, release, security, deploy)")
	GenerateCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language (go, node, python, java, etc.)")
	GenerateCmd.Flags().StringVarP(&outputDir, "output", "o", ".github/workflows", "Output directory")
	GenerateCmd.Flags().BoolVar(&withDocker, "with-docker", false, "Include Docker build steps")
	GenerateCmd.Flags().BoolVar(&withKubernetes, "with-k8s", false, "Include Kubernetes deployment")
	GenerateCmd.Flags().BoolVar(&withSecurity, "with-security", true, "Include security scanning")
	GenerateCmd.Flags().BoolVar(&withCodeCov, "with-codecov", false, "Include code coverage reporting")
	GenerateCmd.Flags().BoolVar(&withRelease, "with-release", false, "Include release automation")
	GenerateCmd.Flags().StringSliceVar(&platforms, "platforms", []string{"ubuntu-latest"}, "Build platforms")
	GenerateCmd.Flags().StringSliceVar(&secrets, "secrets", []string{}, "Required secrets")
	GenerateCmd.Flags().StringSliceVar(&customActions, "custom-actions", []string{}, "Custom actions to include")
	GenerateCmd.Flags().StringVarP(&workflowName, "name", "n", "", "Workflow name")
	GenerateCmd.Flags().BoolVar(&enableCaching, "enable-caching", true, "Enable dependency caching")
	GenerateCmd.Flags().BoolVar(&parallelJobs, "parallel-jobs", true, "Enable parallel job execution")
	GenerateCmd.Flags().BoolVar(&matrixStrategy, "matrix-strategy", false, "Use matrix build strategy")
}

// WorkflowSpec holds workflow generation specifications
type WorkflowSpec struct {
	Name           string
	Type           string
	Language       string
	OutputDir      string
	WithDocker     bool
	WithKubernetes bool
	WithSecurity   bool
	WithCodeCov    bool
	WithRelease    bool
	Platforms      []string
	Secrets        []string
	CustomActions  []string
	EnableCaching  bool
	ParallelJobs   bool
	MatrixStrategy bool
	PackageManager string
	BuildCommands  []string
	TestCommands   []string
	LintCommands   []string
}

// WorkflowConfig represents a complete GitHub Actions workflow
type WorkflowConfig struct {
	Name string                 `yaml:"name"`
	On   WorkflowTriggers       `yaml:"on"`
	Env  map[string]string      `yaml:"env,omitempty"`
	Jobs map[string]WorkflowJob `yaml:"jobs"`
}

type WorkflowTriggers struct {
	Push        *PushTrigger        `yaml:"push,omitempty"`
	PullRequest *PullRequestTrigger `yaml:"pull_request,omitempty"`
	Schedule    []ScheduleTrigger   `yaml:"schedule,omitempty"`
	Workflow    *WorkflowTrigger    `yaml:"workflow_dispatch,omitempty"`
	Release     *ReleaseTrigger     `yaml:"release,omitempty"`
}

type PushTrigger struct {
	Branches []string `yaml:"branches,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
	Paths    []string `yaml:"paths,omitempty"`
}

type PullRequestTrigger struct {
	Branches []string `yaml:"branches,omitempty"`
	Paths    []string `yaml:"paths,omitempty"`
}

type ScheduleTrigger struct {
	Cron string `yaml:"cron"`
}

type WorkflowTrigger struct {
	Inputs map[string]WorkflowInput `yaml:"inputs,omitempty"`
}

type ReleaseTrigger struct {
	Types []string `yaml:"types,omitempty"`
}

type WorkflowInput struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default,omitempty"`
	Type        string `yaml:"type,omitempty"`
}

type WorkflowJob struct {
	Name      string                 `yaml:"name,omitempty"`
	RunsOn    interface{}            `yaml:"runs-on"` // string or matrix
	Needs     []string               `yaml:"needs,omitempty"`
	If        string                 `yaml:"if,omitempty"`
	Strategy  *MatrixStrategy        `yaml:"strategy,omitempty"`
	Env       map[string]string      `yaml:"env,omitempty"`
	Steps     []WorkflowStep         `yaml:"steps"`
	Outputs   map[string]string      `yaml:"outputs,omitempty"`
	Container *ContainerSpec         `yaml:"container,omitempty"`
	Services  map[string]ServiceSpec `yaml:"services,omitempty"`
}

type MatrixStrategy struct {
	Matrix map[string]interface{} `yaml:"matrix"`
}

type WorkflowStep struct {
	Name            string            `yaml:"name,omitempty"`
	ID              string            `yaml:"id,omitempty"`
	Uses            string            `yaml:"uses,omitempty"`
	Run             string            `yaml:"run,omitempty"`
	With            map[string]string `yaml:"with,omitempty"`
	Env             map[string]string `yaml:"env,omitempty"`
	If              string            `yaml:"if,omitempty"`
	ContinueOnError bool              `yaml:"continue-on-error,omitempty"`
	WorkingDir      string            `yaml:"working-directory,omitempty"`
	Shell           string            `yaml:"shell,omitempty"`
}

type ContainerSpec struct {
	Image   string            `yaml:"image"`
	Options string            `yaml:"options,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
}

type ServiceSpec struct {
	Image   string            `yaml:"image"`
	Env     map[string]string `yaml:"env,omitempty"`
	Options string            `yaml:"options,omitempty"`
	Ports   []string          `yaml:"ports,omitempty"`
}

func runGenerate(cmd *cobra.Command, args []string) {
	// Auto-detect language if not specified
	if language == "" {
		detectedLang, err := detectProjectLanguage()
		if err != nil {
			fmt.Printf("Warning: Could not auto-detect language: %v\n", err)
			language = "generic"
		} else {
			language = detectedLang
			fmt.Printf("🔍 Auto-detected language: %s\n", language)
		}
	}

	// Set default workflow name if not specified
	if workflowName == "" {
		workflowName = fmt.Sprintf("%s %s", strings.Title(workflowType), strings.Title(language))
	}

	fmt.Printf("🚀 Generating GitHub Actions workflow: %s\n", workflowName)
	fmt.Printf("📋 Type: %s, Language: %s\n", workflowType, language)

	// Create workflow specification
	spec := WorkflowSpec{
		Name:           workflowName,
		Type:           workflowType,
		Language:       language,
		OutputDir:      outputDir,
		WithDocker:     withDocker,
		WithKubernetes: withKubernetes,
		WithSecurity:   withSecurity,
		WithCodeCov:    withCodeCov,
		WithRelease:    withRelease,
		Platforms:      platforms,
		Secrets:        secrets,
		CustomActions:  customActions,
		EnableCaching:  enableCaching,
		ParallelJobs:   parallelJobs,
		MatrixStrategy: matrixStrategy,
	}

	// Enhance spec with language-specific information
	if err := enhanceSpecForLanguage(&spec); err != nil {
		fmt.Printf("Error enhancing spec: %v\n", err)
		os.Exit(1)
	}

	// Generate workflow
	workflow, err := generateWorkflow(spec)
	if err != nil {
		fmt.Printf("Error generating workflow: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write workflow file
	filename := fmt.Sprintf("%s.yml", strings.ReplaceAll(strings.ToLower(workflowName), " ", "-"))
	filePath := filepath.Join(outputDir, filename)

	if err := writeWorkflow(workflow, filePath); err != nil {
		fmt.Printf("Error writing workflow: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated workflow: %s\n", filePath)

	// Generate additional files if needed
	if withSecurity {
		if err := generateSecurityWorkflow(spec); err != nil {
			fmt.Printf("Warning: Failed to generate security workflow: %v\n", err)
		} else {
			fmt.Println("✅ Generated security scanning workflow")
		}
	}

	if withRelease {
		if err := generateReleaseWorkflow(spec); err != nil {
			fmt.Printf("Warning: Failed to generate release workflow: %v\n", err)
		} else {
			fmt.Println("✅ Generated release automation workflow")
		}
	}

	// Generate custom actions if specified
	if len(customActions) > 0 {
		if err := generateCustomActions(spec); err != nil {
			fmt.Printf("Warning: Failed to generate custom actions: %v\n", err)
		} else {
			fmt.Println("✅ Generated custom actions")
		}
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Review and customize the generated workflows")
	fmt.Println("2. Add required secrets to your GitHub repository")
	fmt.Println("3. Configure environment-specific variables")
	fmt.Println("4. Test workflows with a sample commit or PR")
}

func detectProjectLanguage() (string, error) {
	// Check for various language indicators
	if _, err := os.Stat("go.mod"); err == nil {
		return "go", nil
	}
	if _, err := os.Stat("package.json"); err == nil {
		return "node", nil
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "python", nil
	}
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return "python", nil
	}
	if _, err := os.Stat("setup.py"); err == nil {
		return "python", nil
	}
	if _, err := os.Stat("pom.xml"); err == nil {
		return "java", nil
	}
	if _, err := os.Stat("build.gradle"); err == nil {
		return "java", nil
	}
	if _, err := os.Stat("Cargo.toml"); err == nil {
		return "rust", nil
	}
	if _, err := os.Stat(".csproj"); err == nil {
		return "dotnet", nil
	}
	if _, err := os.Stat("*.sln"); err == nil {
		return "dotnet", nil
	}

	return "", fmt.Errorf("could not detect project language")
}

func enhanceSpecForLanguage(spec *WorkflowSpec) error {
	switch spec.Language {
	case "go":
		spec.PackageManager = "go modules"
		spec.BuildCommands = []string{"go build ./..."}
		spec.TestCommands = []string{"go test -v ./..."}
		spec.LintCommands = []string{"golangci-lint run"}
	case "node", "javascript", "typescript":
		spec.PackageManager = "npm"
		spec.BuildCommands = []string{"npm run build"}
		spec.TestCommands = []string{"npm test"}
		spec.LintCommands = []string{"npm run lint"}
	case "python":
		spec.PackageManager = "pip"
		spec.BuildCommands = []string{"python -m build"}
		spec.TestCommands = []string{"pytest"}
		spec.LintCommands = []string{"flake8", "black --check ."}
	case "java":
		spec.PackageManager = "maven"
		spec.BuildCommands = []string{"mvn compile"}
		spec.TestCommands = []string{"mvn test"}
		spec.LintCommands = []string{"mvn checkstyle:check"}
	case "rust":
		spec.PackageManager = "cargo"
		spec.BuildCommands = []string{"cargo build"}
		spec.TestCommands = []string{"cargo test"}
		spec.LintCommands = []string{"cargo clippy"}
	case "dotnet":
		spec.PackageManager = "dotnet"
		spec.BuildCommands = []string{"dotnet build"}
		spec.TestCommands = []string{"dotnet test"}
		spec.LintCommands = []string{"dotnet format --verify-no-changes"}
	default:
		spec.PackageManager = "generic"
		spec.BuildCommands = []string{"make build"}
		spec.TestCommands = []string{"make test"}
		spec.LintCommands = []string{"make lint"}
	}

	return nil
}

func generateWorkflow(spec WorkflowSpec) (*WorkflowConfig, error) {
	workflow := &WorkflowConfig{
		Name: spec.Name,
		On:   generateTriggers(spec),
		Env:  generateGlobalEnv(spec),
		Jobs: make(map[string]WorkflowJob),
	}

	// Generate jobs based on type
	switch spec.Type {
	case "ci":
		workflow.Jobs = generateCIJobs(spec)
	case "release":
		workflow.Jobs = generateReleaseJobs(spec)
	case "security":
		workflow.Jobs = generateSecurityJobs(spec)
	case "deploy":
		workflow.Jobs = generateDeployJobs(spec)
	default:
		return nil, fmt.Errorf("unsupported workflow type: %s", spec.Type)
	}

	return workflow, nil
}

func generateTriggers(spec WorkflowSpec) WorkflowTriggers {
	triggers := WorkflowTriggers{
		Workflow: &WorkflowTrigger{},
	}

	switch spec.Type {
	case "ci":
		triggers.Push = &PushTrigger{
			Branches: []string{"main", "master", "develop"},
		}
		triggers.PullRequest = &PullRequestTrigger{
			Branches: []string{"main", "master", "develop"},
		}
	case "release":
		triggers.Push = &PushTrigger{
			Tags: []string{"v*"},
		}
		triggers.Release = &ReleaseTrigger{
			Types: []string{"published"},
		}
	case "security":
		triggers.Schedule = []ScheduleTrigger{
			{Cron: "0 0 * * 1"}, // Weekly on Monday
		}
		triggers.Push = &PushTrigger{
			Branches: []string{"main", "master"},
		}
	case "deploy":
		triggers.Workflow = &WorkflowTrigger{
			Inputs: map[string]WorkflowInput{
				"environment": {
					Description: "Deployment environment",
					Required:    true,
					Default:     "staging",
					Type:        "choice",
				},
			},
		}
	}

	return triggers
}

func generateGlobalEnv(spec WorkflowSpec) map[string]string {
	env := map[string]string{}

	// Language-specific environment variables
	switch spec.Language {
	case "go":
		env["GO_VERSION"] = "1.21"
		env["CGO_ENABLED"] = "0"
	case "node":
		env["NODE_VERSION"] = "18"
	case "python":
		env["PYTHON_VERSION"] = "3.11"
	case "java":
		env["JAVA_VERSION"] = "17"
	case "rust":
		env["RUST_VERSION"] = "stable"
	}

	// Docker-specific variables
	if spec.WithDocker {
		env["REGISTRY"] = "ghcr.io"
		env["IMAGE_NAME"] = "${{ github.repository }}"
	}

	return env
}

func generateCIJobs(spec WorkflowSpec) map[string]WorkflowJob {
	jobs := make(map[string]WorkflowJob)

	// Test job
	testJob := WorkflowJob{
		Name:   "Test",
		RunsOn: generateRunsOn(spec),
		Steps:  generateTestSteps(spec),
	}

	if spec.MatrixStrategy {
		testJob.Strategy = generateMatrixStrategy(spec)
	}

	jobs["test"] = testJob

	// Lint job (parallel with test)
	if len(spec.LintCommands) > 0 {
		jobs["lint"] = WorkflowJob{
			Name:   "Lint",
			RunsOn: spec.Platforms[0],
			Steps:  generateLintSteps(spec),
		}
	}

	// Security job (parallel with test)
	if spec.WithSecurity {
		jobs["security"] = WorkflowJob{
			Name:   "Security",
			RunsOn: spec.Platforms[0],
			Steps:  generateSecuritySteps(spec),
		}
	}

	// Build job (after test passes)
	if len(spec.BuildCommands) > 0 {
		buildJob := WorkflowJob{
			Name:   "Build",
			RunsOn: spec.Platforms[0],
			Needs:  []string{"test"},
			Steps:  generateBuildSteps(spec),
		}

		if spec.WithDocker {
			buildJob.Steps = append(buildJob.Steps, generateDockerSteps(spec)...)
		}

		jobs["build"] = buildJob
	}

	return jobs
}

func generateTestSteps(spec WorkflowSpec) []WorkflowStep {
	steps := []WorkflowStep{
		{
			Name: "Checkout code",
			Uses: "actions/checkout@v4",
		},
	}

	// Setup language environment
	steps = append(steps, generateLanguageSetupSteps(spec)...)

	// Cache dependencies
	if spec.EnableCaching {
		steps = append(steps, generateCacheSteps(spec)...)
	}

	// Install dependencies
	steps = append(steps, generateDependencySteps(spec)...)

	// Run tests
	for _, testCmd := range spec.TestCommands {
		steps = append(steps, WorkflowStep{
			Name: "Run tests",
			Run:  testCmd,
		})
	}

	// Code coverage
	if spec.WithCodeCov {
		steps = append(steps, WorkflowStep{
			Name: "Upload coverage to Codecov",
			Uses: "codecov/codecov-action@v3",
			With: map[string]string{
				"token": "${{ secrets.CODECOV_TOKEN }}",
			},
		})
	}

	return steps
}

func generateLanguageSetupSteps(spec WorkflowSpec) []WorkflowStep {
	var steps []WorkflowStep

	switch spec.Language {
	case "go":
		steps = append(steps, WorkflowStep{
			Name: "Setup Go",
			Uses: "actions/setup-go@v4",
			With: map[string]string{
				"go-version": "${{ env.GO_VERSION }}",
			},
		})
	case "node":
		steps = append(steps, WorkflowStep{
			Name: "Setup Node.js",
			Uses: "actions/setup-node@v3",
			With: map[string]string{
				"node-version": "${{ env.NODE_VERSION }}",
			},
		})
	case "python":
		steps = append(steps, WorkflowStep{
			Name: "Setup Python",
			Uses: "actions/setup-python@v4",
			With: map[string]string{
				"python-version": "${{ env.PYTHON_VERSION }}",
			},
		})
	case "java":
		steps = append(steps, WorkflowStep{
			Name: "Setup Java",
			Uses: "actions/setup-java@v3",
			With: map[string]string{
				"java-version": "${{ env.JAVA_VERSION }}",
				"distribution": "temurin",
			},
		})
	case "rust":
		steps = append(steps, WorkflowStep{
			Name: "Setup Rust",
			Uses: "actions-rs/toolchain@v1",
			With: map[string]string{
				"toolchain": "${{ env.RUST_VERSION }}",
				"override":  "true",
			},
		})
	}

	return steps
}

func generateCacheSteps(spec WorkflowSpec) []WorkflowStep {
	var steps []WorkflowStep

	switch spec.Language {
	case "go":
		steps = append(steps, WorkflowStep{
			Name: "Cache Go modules",
			Uses: "actions/cache@v3",
			With: map[string]string{
				"path": "~/go/pkg/mod",
				"key":  "${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}",
			},
		})
	case "node":
		steps = append(steps, WorkflowStep{
			Name: "Cache node modules",
			Uses: "actions/cache@v3",
			With: map[string]string{
				"path": "~/.npm",
				"key":  "${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}",
			},
		})
	case "python":
		steps = append(steps, WorkflowStep{
			Name: "Cache pip packages",
			Uses: "actions/cache@v3",
			With: map[string]string{
				"path": "~/.cache/pip",
				"key":  "${{ runner.os }}-pip-${{ hashFiles('**/requirements.txt') }}",
			},
		})
	}

	return steps
}

func generateDependencySteps(spec WorkflowSpec) []WorkflowStep {
	var steps []WorkflowStep

	switch spec.Language {
	case "go":
		steps = append(steps, WorkflowStep{
			Name: "Download dependencies",
			Run:  "go mod download",
		})
	case "node":
		steps = append(steps, WorkflowStep{
			Name: "Install dependencies",
			Run:  "npm ci",
		})
	case "python":
		steps = append(steps, WorkflowStep{
			Name: "Install dependencies",
			Run:  "pip install -r requirements.txt",
		})
	case "java":
		steps = append(steps, WorkflowStep{
			Name: "Cache Maven packages",
			Uses: "actions/cache@v3",
			With: map[string]string{
				"path": "~/.m2",
				"key":  "${{ runner.os }}-m2-${{ hashFiles('**/pom.xml') }}",
			},
		})
	}

	return steps
}

func generateLintSteps(spec WorkflowSpec) []WorkflowStep {
	steps := []WorkflowStep{
		{
			Name: "Checkout code",
			Uses: "actions/checkout@v4",
		},
	}

	steps = append(steps, generateLanguageSetupSteps(spec)...)

	if spec.EnableCaching {
		steps = append(steps, generateCacheSteps(spec)...)
	}

	steps = append(steps, generateDependencySteps(spec)...)

	// Run lint commands
	for _, lintCmd := range spec.LintCommands {
		steps = append(steps, WorkflowStep{
			Name: "Run linter",
			Run:  lintCmd,
		})
	}

	return steps
}

func generateSecuritySteps(spec WorkflowSpec) []WorkflowStep {
	steps := []WorkflowStep{
		{
			Name: "Checkout code",
			Uses: "actions/checkout@v4",
		},
		{
			Name: "Run Trivy vulnerability scanner",
			Uses: "aquasecurity/trivy-action@master",
			With: map[string]string{
				"scan-type": "fs",
				"scan-ref":  ".",
				"format":    "sarif",
				"output":    "trivy-results.sarif",
			},
		},
		{
			Name: "Upload Trivy scan results to GitHub Security",
			Uses: "github/codeql-action/upload-sarif@v2",
			With: map[string]string{
				"sarif_file": "trivy-results.sarif",
			},
		},
	}

	return steps
}

func generateBuildSteps(spec WorkflowSpec) []WorkflowStep {
	steps := []WorkflowStep{
		{
			Name: "Checkout code",
			Uses: "actions/checkout@v4",
		},
	}

	steps = append(steps, generateLanguageSetupSteps(spec)...)

	if spec.EnableCaching {
		steps = append(steps, generateCacheSteps(spec)...)
	}

	steps = append(steps, generateDependencySteps(spec)...)

	// Run build commands
	for _, buildCmd := range spec.BuildCommands {
		steps = append(steps, WorkflowStep{
			Name: "Build project",
			Run:  buildCmd,
		})
	}

	return steps
}

func generateDockerSteps(spec WorkflowSpec) []WorkflowStep {
	return []WorkflowStep{
		{
			Name: "Set up Docker Buildx",
			Uses: "docker/setup-buildx-action@v3",
		},
		{
			Name: "Log in to Container Registry",
			Uses: "docker/login-action@v3",
			With: map[string]string{
				"registry": "${{ env.REGISTRY }}",
				"username": "${{ github.actor }}",
				"password": "${{ secrets.GITHUB_TOKEN }}",
			},
		},
		{
			Name: "Extract metadata",
			ID:   "meta",
			Uses: "docker/metadata-action@v5",
			With: map[string]string{
				"images": "${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}",
			},
		},
		{
			Name: "Build and push Docker image",
			Uses: "docker/build-push-action@v5",
			With: map[string]string{
				"context":   ".",
				"push":      "true",
				"tags":      "${{ steps.meta.outputs.tags }}",
				"labels":    "${{ steps.meta.outputs.labels }}",
				"platforms": "linux/amd64,linux/arm64",
			},
		},
	}
}

func generateRunsOn(spec WorkflowSpec) interface{} {
	if spec.MatrixStrategy && len(spec.Platforms) > 1 {
		return "${{ matrix.os }}"
	}
	return spec.Platforms[0]
}

func generateMatrixStrategy(spec WorkflowSpec) *MatrixStrategy {
	matrix := map[string]interface{}{}

	if len(spec.Platforms) > 1 {
		matrix["os"] = spec.Platforms
	}

	// Add language version matrix for some languages
	switch spec.Language {
	case "go":
		matrix["go-version"] = []string{"1.20", "1.21"}
	case "node":
		matrix["node-version"] = []string{"16", "18", "20"}
	case "python":
		matrix["python-version"] = []string{"3.9", "3.10", "3.11"}
	}

	return &MatrixStrategy{Matrix: matrix}
}

func generateReleaseJobs(spec WorkflowSpec) map[string]WorkflowJob {
	// Implementation for release jobs
	return map[string]WorkflowJob{
		"release": {
			Name:   "Release",
			RunsOn: "ubuntu-latest",
			Steps: []WorkflowStep{
				{
					Name: "Checkout code",
					Uses: "actions/checkout@v4",
				},
				{
					Name: "Create Release",
					Uses: "actions/create-release@v1",
					Env: map[string]string{
						"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
					},
				},
			},
		},
	}
}

func generateSecurityJobs(spec WorkflowSpec) map[string]WorkflowJob {
	return map[string]WorkflowJob{
		"security": {
			Name:   "Security Scan",
			RunsOn: "ubuntu-latest",
			Steps:  generateSecuritySteps(spec),
		},
	}
}

func generateDeployJobs(spec WorkflowSpec) map[string]WorkflowJob {
	jobs := map[string]WorkflowJob{
		"deploy": {
			Name:   "Deploy",
			RunsOn: "ubuntu-latest",
			Steps: []WorkflowStep{
				{
					Name: "Checkout code",
					Uses: "actions/checkout@v4",
				},
			},
		},
	}

	if spec.WithKubernetes {
		deployJob := jobs["deploy"]
		deployJob.Steps = append(deployJob.Steps, WorkflowStep{
			Name: "Deploy to Kubernetes",
			Run:  "kubectl apply -f k8s/",
		})
		jobs["deploy"] = deployJob
	}

	return jobs
}

func writeWorkflow(workflow *WorkflowConfig, filePath string) error {
	data, err := yaml.Marshal(workflow)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

func generateSecurityWorkflow(spec WorkflowSpec) error {
	// Generate dedicated security workflow
	securityWorkflow := &WorkflowConfig{
		Name: "Security Scan",
		On: WorkflowTriggers{
			Schedule: []ScheduleTrigger{
				{Cron: "0 0 * * 1"},
			},
			Push: &PushTrigger{
				Branches: []string{"main", "master"},
			},
		},
		Jobs: generateSecurityJobs(spec),
	}

	filename := "security.yml"
	filePath := filepath.Join(spec.OutputDir, filename)
	return writeWorkflow(securityWorkflow, filePath)
}

func generateReleaseWorkflow(spec WorkflowSpec) error {
	// Generate dedicated release workflow
	releaseWorkflow := &WorkflowConfig{
		Name: "Release",
		On: WorkflowTriggers{
			Push: &PushTrigger{
				Tags: []string{"v*"},
			},
		},
		Jobs: generateReleaseJobs(spec),
	}

	filename := "release.yml"
	filePath := filepath.Join(spec.OutputDir, filename)
	return writeWorkflow(releaseWorkflow, filePath)
}

func generateCustomActions(spec WorkflowSpec) error {
	// Generate custom actions directory structure
	actionsDir := filepath.Join(spec.OutputDir, "..", "actions")

	for _, actionName := range spec.CustomActions {
		actionDir := filepath.Join(actionsDir, actionName)
		if err := os.MkdirAll(actionDir, 0o755); err != nil {
			return err
		}

		// Generate action.yml
		actionConfig := map[string]interface{}{
			"name":        actionName,
			"description": fmt.Sprintf("Custom action for %s", actionName),
			"runs": map[string]string{
				"using": "composite",
			},
		}

		data, err := yaml.Marshal(actionConfig)
		if err != nil {
			return err
		}

		actionFile := filepath.Join(actionDir, "action.yml")
		if err := os.WriteFile(actionFile, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}
