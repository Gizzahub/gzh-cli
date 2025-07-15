package jenkins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Jenkins pipelines and configurations",
	Long: `Generate Jenkins pipelines and configurations for various project types.

Supports generation of:
- Declarative and Scripted Jenkinsfiles
- Multi-branch pipeline configurations
- Shared library implementations
- Plugin configuration files
- Docker-based build environments
- Kubernetes deployment pipelines
- Blue Ocean compatible pipelines

Examples:
  gz jenkins generate --language go --type declarative
  gz jenkins generate --type shared-library --name deploy-utils
  gz jenkins generate --multi-branch --with-docker`,
	Run: runGenerate,
}

var (
	language       string
	pipelineType   string
	outputFile     string
	withDocker     bool
	withKubernetes bool
	withSecurity   bool
	withTesting    bool
	withDeployment bool
	environments   []string
	agents         []string
	sharedLibrary  string
	libraryName    string
	multiBranch    bool
	blueOcean      bool
	parallelStages bool
	customTools    []string
	credentials    []string
	notifications  []string
	postActions    []string
	timeoutMinutes int
)

func init() {
	GenerateCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language (go, node, python, java, etc.)")
	GenerateCmd.Flags().StringVarP(&pipelineType, "type", "t", "declarative", "Pipeline type (declarative, scripted, shared-library)")
	GenerateCmd.Flags().StringVarP(&outputFile, "output", "o", "Jenkinsfile", "Output file path")
	GenerateCmd.Flags().BoolVar(&withDocker, "with-docker", false, "Include Docker build steps")
	GenerateCmd.Flags().BoolVar(&withKubernetes, "with-k8s", false, "Include Kubernetes deployment")
	GenerateCmd.Flags().BoolVar(&withSecurity, "with-security", true, "Include security scanning")
	GenerateCmd.Flags().BoolVar(&withTesting, "with-testing", true, "Include testing stages")
	GenerateCmd.Flags().BoolVar(&withDeployment, "with-deployment", false, "Include deployment stages")
	GenerateCmd.Flags().StringSliceVar(&environments, "environments", []string{"staging", "production"}, "Deployment environments")
	GenerateCmd.Flags().StringSliceVar(&agents, "agents", []string{}, "Jenkins agent labels")
	GenerateCmd.Flags().StringVar(&sharedLibrary, "shared-library", "", "Shared library to use")
	GenerateCmd.Flags().StringVar(&libraryName, "library-name", "", "Name for shared library generation")
	GenerateCmd.Flags().BoolVar(&multiBranch, "multi-branch", false, "Generate multi-branch pipeline")
	GenerateCmd.Flags().BoolVar(&blueOcean, "blue-ocean", false, "Blue Ocean compatible pipeline")
	GenerateCmd.Flags().BoolVar(&parallelStages, "parallel-stages", false, "Use parallel stages")
	GenerateCmd.Flags().StringSliceVar(&customTools, "tools", []string{}, "Custom tools configuration")
	GenerateCmd.Flags().StringSliceVar(&credentials, "credentials", []string{}, "Credentials to configure")
	GenerateCmd.Flags().StringSliceVar(&notifications, "notifications", []string{}, "Notification configurations")
	GenerateCmd.Flags().StringSliceVar(&postActions, "post-actions", []string{}, "Post-build actions")
	GenerateCmd.Flags().IntVar(&timeoutMinutes, "timeout", 60, "Pipeline timeout in minutes")
}

// PipelineSpec holds pipeline generation specifications
type PipelineSpec struct {
	Language       string
	Type           string
	WithDocker     bool
	WithKubernetes bool
	WithSecurity   bool
	WithTesting    bool
	WithDeployment bool
	Environments   []string
	Agents         []string
	SharedLibrary  string
	LibraryName    string
	MultiBranch    bool
	BlueOcean      bool
	ParallelStages bool
	CustomTools    []string
	Credentials    []string
	Notifications  []string
	PostActions    []string
	TimeoutMinutes int
	PackageManager string
	BuildCommands  []string
	TestCommands   []string
	LintCommands   []string
}

// JenkinsConfig represents Jenkins configuration
type JenkinsConfig struct {
	Pipeline        PipelineConfig         `yaml:"pipeline,omitempty"`
	SharedLibraries []SharedLibraryConfig  `yaml:"shared_libraries,omitempty"`
	Tools           map[string]interface{} `yaml:"tools,omitempty"`
	Credentials     []CredentialConfig     `yaml:"credentials,omitempty"`
	GlobalPipelines []string               `yaml:"global_pipelines,omitempty"`
}

type PipelineConfig struct {
	Agent      AgentConfig             `yaml:"agent,omitempty"`
	Tools      map[string]string       `yaml:"tools,omitempty"`
	Options    OptionsConfig           `yaml:"options,omitempty"`
	Triggers   []TriggerConfig         `yaml:"triggers,omitempty"`
	Parameters []ParameterConfig       `yaml:"parameters,omitempty"`
	Stages     []StageConfig           `yaml:"stages,omitempty"`
	Post       map[string][]StepConfig `yaml:"post,omitempty"`
}

type AgentConfig struct {
	Label      string            `yaml:"label,omitempty"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Image      string            `yaml:"image,omitempty"`
	Args       string            `yaml:"args,omitempty"`
	None       bool              `yaml:"none,omitempty"`
	Any        bool              `yaml:"any,omitempty"`
	Node       string            `yaml:"node,omitempty"`
	Custom     map[string]string `yaml:"custom,omitempty"`
}

type OptionsConfig struct {
	BuildDiscarder    map[string]interface{} `yaml:"buildDiscarder,omitempty"`
	DisableConcurrent bool                   `yaml:"disableConcurrentBuilds,omitempty"`
	OverrideIndex     bool                   `yaml:"overrideIndexTriggers,omitempty"`
	SkipStagesAfter   string                 `yaml:"skipStagesAfterUnstable,omitempty"`
	Timeout           string                 `yaml:"timeout,omitempty"`
	Retry             int                    `yaml:"retry,omitempty"`
	Timestamps        bool                   `yaml:"timestamps,omitempty"`
	CheckoutToSubdir  string                 `yaml:"checkoutToSubdirectory,omitempty"`
	NewContainerPerSt bool                   `yaml:"newContainerPerStage,omitempty"`
}

type TriggerConfig struct {
	Type       string            `yaml:"type"`
	Schedule   string            `yaml:"schedule,omitempty"`
	Upstream   []string          `yaml:"upstream,omitempty"`
	PollSCM    string            `yaml:"pollSCM,omitempty"`
	GenericURL map[string]string `yaml:"genericURL,omitempty"`
}

type ParameterConfig struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Default     interface{} `yaml:"default,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Choices     []string    `yaml:"choices,omitempty"`
	Trim        bool        `yaml:"trim,omitempty"`
}

type StageConfig struct {
	Name     string                  `yaml:"name"`
	Agent    *AgentConfig            `yaml:"agent,omitempty"`
	When     *WhenConfig             `yaml:"when,omitempty"`
	Steps    []StepConfig            `yaml:"steps,omitempty"`
	Parallel map[string]StageConfig  `yaml:"parallel,omitempty"`
	Matrix   *MatrixConfig           `yaml:"matrix,omitempty"`
	Input    *InputConfig            `yaml:"input,omitempty"`
	Options  *OptionsConfig          `yaml:"options,omitempty"`
	Tools    map[string]string       `yaml:"tools,omitempty"`
	Env      map[string]string       `yaml:"environment,omitempty"`
	Post     map[string][]StepConfig `yaml:"post,omitempty"`
}

type WhenConfig struct {
	Branch        string            `yaml:"branch,omitempty"`
	BuildingTag   bool              `yaml:"buildingTag,omitempty"`
	ChangeRequest bool              `yaml:"changeRequest,omitempty"`
	Environment   string            `yaml:"environment,omitempty"`
	Expression    string            `yaml:"expression,omitempty"`
	Tag           string            `yaml:"tag,omitempty"`
	Not           *WhenConfig       `yaml:"not,omitempty"`
	AllOf         []WhenConfig      `yaml:"allOf,omitempty"`
	AnyOf         []WhenConfig      `yaml:"anyOf,omitempty"`
	Custom        map[string]string `yaml:"custom,omitempty"`
}

type StepConfig struct {
	Script     string                 `yaml:"script,omitempty"`
	Shell      string                 `yaml:"sh,omitempty"`
	Bat        string                 `yaml:"bat,omitempty"`
	PowerShell string                 `yaml:"powershell,omitempty"`
	Echo       string                 `yaml:"echo,omitempty"`
	Dir        string                 `yaml:"dir,omitempty"`
	Git        *GitConfig             `yaml:"git,omitempty"`
	Checkout   *CheckoutConfig        `yaml:"checkout,omitempty"`
	Build      *BuildConfig           `yaml:"build,omitempty"`
	Archive    *ArchiveConfig         `yaml:"archiveArtifacts,omitempty"`
	Publish    *PublishConfig         `yaml:"publishTestResults,omitempty"`
	Docker     *DockerConfig          `yaml:"docker,omitempty"`
	Kubernetes *KubernetesConfig      `yaml:"kubernetes,omitempty"`
	Custom     map[string]interface{} `yaml:"custom,omitempty"`
}

type GitConfig struct {
	URL        string `yaml:"url"`
	Branch     string `yaml:"branch,omitempty"`
	Credential string `yaml:"credentialsId,omitempty"`
}

type CheckoutConfig struct {
	SCM    string `yaml:"scm"`
	Poll   bool   `yaml:"poll,omitempty"`
	Change bool   `yaml:"changelog,omitempty"`
}

type BuildConfig struct {
	Job        string            `yaml:"job"`
	Parameters map[string]string `yaml:"parameters,omitempty"`
	Wait       bool              `yaml:"wait,omitempty"`
}

type ArchiveConfig struct {
	Artifacts        string `yaml:"artifacts"`
	Excludes         string `yaml:"excludes,omitempty"`
	Fingerprint      bool   `yaml:"fingerprint,omitempty"`
	OnlyIfSuccessful bool   `yaml:"onlyIfSuccessful,omitempty"`
}

type PublishConfig struct {
	TestResults string `yaml:"testResultsPattern"`
	AllowEmpty  bool   `yaml:"allowEmptyResults,omitempty"`
	KeepLong    bool   `yaml:"keepLongStdio,omitempty"`
}

type DockerConfig struct {
	Image    string            `yaml:"image,omitempty"`
	Build    string            `yaml:"build,omitempty"`
	Push     []string          `yaml:"push,omitempty"`
	Run      *DockerRunConfig  `yaml:"run,omitempty"`
	Registry string            `yaml:"registry,omitempty"`
	Args     string            `yaml:"args,omitempty"`
	Custom   map[string]string `yaml:"custom,omitempty"`
}

type DockerRunConfig struct {
	Image string `yaml:"image"`
	Args  string `yaml:"args,omitempty"`
}

type KubernetesConfig struct {
	ConfigFile string            `yaml:"configFile,omitempty"`
	Namespace  string            `yaml:"namespace,omitempty"`
	Manifests  []string          `yaml:"manifests,omitempty"`
	DryRun     bool              `yaml:"dryRun,omitempty"`
	Custom     map[string]string `yaml:"custom,omitempty"`
}

type MatrixConfig struct {
	Axes    []AxisConfig    `yaml:"axes"`
	Stages  []StageConfig   `yaml:"stages,omitempty"`
	Exclude []ExcludeConfig `yaml:"excludes,omitempty"`
}

type AxisConfig struct {
	Name   string   `yaml:"name"`
	Values []string `yaml:"values"`
}

type ExcludeConfig struct {
	Axis   string   `yaml:"axis"`
	Values []string `yaml:"values"`
}

type InputConfig struct {
	Message    string            `yaml:"message"`
	Ok         string            `yaml:"ok,omitempty"`
	Submitter  string            `yaml:"submitter,omitempty"`
	Parameters []ParameterConfig `yaml:"parameters,omitempty"`
}

type SharedLibraryConfig struct {
	Name                 string `yaml:"name"`
	Version              string `yaml:"version,omitempty"`
	Retrieval            string `yaml:"retrieval"`
	DefaultVersion       string `yaml:"defaultVersion,omitempty"`
	Implicit             bool   `yaml:"implicit,omitempty"`
	AllowVersionOverride bool   `yaml:"allowVersionOverride,omitempty"`
}

type CredentialConfig struct {
	ID          string `yaml:"id"`
	Type        string `yaml:"type"`
	Description string `yaml:"description,omitempty"`
	Username    string `yaml:"username,omitempty"`
	Password    string `yaml:"password,omitempty"`
	SecretText  string `yaml:"secretText,omitempty"`
	PrivateKey  string `yaml:"privateKey,omitempty"`
	Passphrase  string `yaml:"passphrase,omitempty"`
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

	fmt.Printf("🚀 Generating Jenkins pipeline\n")
	fmt.Printf("📋 Language: %s, Type: %s\n", language, pipelineType)

	// Create pipeline specification
	spec := PipelineSpec{
		Language:       language,
		Type:           pipelineType,
		WithDocker:     withDocker,
		WithKubernetes: withKubernetes,
		WithSecurity:   withSecurity,
		WithTesting:    withTesting,
		WithDeployment: withDeployment,
		Environments:   environments,
		Agents:         agents,
		SharedLibrary:  sharedLibrary,
		LibraryName:    libraryName,
		MultiBranch:    multiBranch,
		BlueOcean:      blueOcean,
		ParallelStages: parallelStages,
		CustomTools:    customTools,
		Credentials:    credentials,
		Notifications:  notifications,
		PostActions:    postActions,
		TimeoutMinutes: timeoutMinutes,
	}

	// Enhance spec with language-specific information
	if err := enhanceSpecForLanguage(&spec); err != nil {
		fmt.Printf("Error enhancing spec: %v\n", err)
		os.Exit(1)
	}

	// Generate based on pipeline type
	var content string
	var err error

	switch spec.Type {
	case "declarative":
		content, err = generateDeclarativePipeline(spec)
	case "scripted":
		content, err = generateScriptedPipeline(spec)
	case "shared-library":
		err = generateSharedLibrary(spec)
		if err == nil {
			fmt.Println("✅ Generated shared library structure")
			return
		}
	default:
		content, err = generateDeclarativePipeline(spec)
	}

	if err != nil {
		fmt.Printf("Error generating pipeline: %v\n", err)
		os.Exit(1)
	}

	// Write pipeline file
	if err := os.WriteFile(outputFile, []byte(content), 0o644); err != nil {
		fmt.Printf("Error writing pipeline: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated Jenkins pipeline: %s\n", outputFile)

	// Generate additional files if requested
	if withDocker {
		if err := generateDockerfiles(spec); err != nil {
			fmt.Printf("Warning: Failed to generate Dockerfiles: %v\n", err)
		} else {
			fmt.Println("✅ Generated Dockerfile")
		}
	}

	if multiBranch {
		if err := generateMultiBranchConfig(spec); err != nil {
			fmt.Printf("Warning: Failed to generate multi-branch config: %v\n", err)
		} else {
			fmt.Println("✅ Generated multi-branch configuration")
		}
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Review and customize the generated pipeline")
	fmt.Println("2. Configure Jenkins credentials and tools")
	fmt.Println("3. Set up Jenkins agents if using specific labels")
	fmt.Println("4. Test pipeline with a sample build")
	fmt.Println("5. Configure webhooks for automatic triggering")
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
	if _, err := os.Stat("composer.json"); err == nil {
		return "php", nil
	}

	return "", fmt.Errorf("could not detect project language")
}

func enhanceSpecForLanguage(spec *PipelineSpec) error {
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
	case "php":
		spec.PackageManager = "composer"
		spec.BuildCommands = []string{"composer install --no-dev"}
		spec.TestCommands = []string{"vendor/bin/phpunit"}
		spec.LintCommands = []string{"vendor/bin/phpcs"}
	default:
		spec.PackageManager = "generic"
		spec.BuildCommands = []string{"make build"}
		spec.TestCommands = []string{"make test"}
		spec.LintCommands = []string{"make lint"}
	}

	return nil
}

func generateDeclarativePipeline(spec PipelineSpec) (string, error) {
	var pipeline strings.Builder

	// Pipeline header
	pipeline.WriteString("pipeline {\n")

	// Agent configuration
	if len(spec.Agents) > 0 {
		pipeline.WriteString(fmt.Sprintf("    agent {\n        label '%s'\n    }\n", spec.Agents[0]))
	} else if spec.WithDocker {
		pipeline.WriteString("    agent {\n        docker {\n            image 'ubuntu:20.04'\n        }\n    }\n")
	} else {
		pipeline.WriteString("    agent any\n")
	}

	// Options
	pipeline.WriteString("    options {\n")
	pipeline.WriteString(fmt.Sprintf("        timeout(time: %d, unit: 'MINUTES')\n", spec.TimeoutMinutes))
	pipeline.WriteString("        buildDiscarder(logRotator(numToKeepStr: '10'))\n")
	pipeline.WriteString("        timestamps()\n")
	if spec.BlueOcean {
		pipeline.WriteString("        skipDefaultCheckout(true)\n")
	}
	pipeline.WriteString("    }\n")

	// Tools
	pipeline.WriteString("    tools {\n")
	switch spec.Language {
	case "go":
		pipeline.WriteString("        go 'go-1.21'\n")
	case "node":
		pipeline.WriteString("        nodejs 'nodejs-18'\n")
	case "python":
		pipeline.WriteString("        python 'python-3.11'\n")
	case "java":
		pipeline.WriteString("        maven 'maven-3.9'\n")
		pipeline.WriteString("        jdk 'jdk-17'\n")
	}
	for _, tool := range spec.CustomTools {
		pipeline.WriteString(fmt.Sprintf("        %s\n", tool))
	}
	pipeline.WriteString("    }\n")

	// Environment variables
	pipeline.WriteString("    environment {\n")
	switch spec.Language {
	case "go":
		pipeline.WriteString("        CGO_ENABLED = '0'\n")
		pipeline.WriteString("        GOOS = 'linux'\n")
	case "node":
		pipeline.WriteString("        NODE_ENV = 'production'\n")
	case "python":
		pipeline.WriteString("        PYTHONDONTWRITEBYTECODE = '1'\n")
	}
	pipeline.WriteString("    }\n")

	// Stages
	pipeline.WriteString("    stages {\n")

	// Checkout stage
	pipeline.WriteString("        stage('Checkout') {\n")
	pipeline.WriteString("            steps {\n")
	pipeline.WriteString("                checkout scm\n")
	pipeline.WriteString("            }\n")
	pipeline.WriteString("        }\n")

	// Dependencies stage
	pipeline.WriteString("        stage('Dependencies') {\n")
	pipeline.WriteString("            steps {\n")
	pipeline.WriteString("                script {\n")
	for _, cmd := range generateDependencyCommands(spec) {
		pipeline.WriteString(fmt.Sprintf("                    sh '%s'\n", cmd))
	}
	pipeline.WriteString("                }\n")
	pipeline.WriteString("            }\n")
	pipeline.WriteString("        }\n")

	// Build stage
	pipeline.WriteString("        stage('Build') {\n")
	pipeline.WriteString("            steps {\n")
	pipeline.WriteString("                script {\n")
	for _, cmd := range spec.BuildCommands {
		pipeline.WriteString(fmt.Sprintf("                    sh '%s'\n", cmd))
	}
	pipeline.WriteString("                }\n")
	pipeline.WriteString("            }\n")
	if spec.WithDocker {
		pipeline.WriteString("            post {\n")
		pipeline.WriteString("                success {\n")
		pipeline.WriteString("                    archiveArtifacts artifacts: 'dist/**, build/**', fingerprint: true\n")
		pipeline.WriteString("                }\n")
		pipeline.WriteString("            }\n")
	}
	pipeline.WriteString("        }\n")

	// Test stage
	if spec.WithTesting {
		if spec.ParallelStages {
			pipeline.WriteString("        stage('Test') {\n")
			pipeline.WriteString("            parallel {\n")

			// Unit tests
			pipeline.WriteString("                stage('Unit Tests') {\n")
			pipeline.WriteString("                    steps {\n")
			pipeline.WriteString("                        script {\n")
			for _, cmd := range spec.TestCommands {
				pipeline.WriteString(fmt.Sprintf("                            sh '%s'\n", cmd))
			}
			pipeline.WriteString("                        }\n")
			pipeline.WriteString("                    }\n")
			pipeline.WriteString("                    post {\n")
			pipeline.WriteString("                        always {\n")
			pipeline.WriteString("                            publishTestResults testResultsPattern: 'test-results.xml'\n")
			pipeline.WriteString("                        }\n")
			pipeline.WriteString("                    }\n")
			pipeline.WriteString("                }\n")

			// Lint
			if len(spec.LintCommands) > 0 {
				pipeline.WriteString("                stage('Lint') {\n")
				pipeline.WriteString("                    steps {\n")
				pipeline.WriteString("                        script {\n")
				for _, cmd := range spec.LintCommands {
					pipeline.WriteString(fmt.Sprintf("                            sh '%s'\n", cmd))
				}
				pipeline.WriteString("                        }\n")
				pipeline.WriteString("                    }\n")
				pipeline.WriteString("                }\n")
			}

			pipeline.WriteString("            }\n")
			pipeline.WriteString("        }\n")
		} else {
			// Sequential test stages
			pipeline.WriteString("        stage('Test') {\n")
			pipeline.WriteString("            steps {\n")
			pipeline.WriteString("                script {\n")
			for _, cmd := range spec.TestCommands {
				pipeline.WriteString(fmt.Sprintf("                    sh '%s'\n", cmd))
			}
			pipeline.WriteString("                }\n")
			pipeline.WriteString("            }\n")
			pipeline.WriteString("            post {\n")
			pipeline.WriteString("                always {\n")
			pipeline.WriteString("                    publishTestResults testResultsPattern: 'test-results.xml'\n")
			pipeline.WriteString("                }\n")
			pipeline.WriteString("            }\n")
			pipeline.WriteString("        }\n")
		}
	}

	// Security stage
	if spec.WithSecurity {
		pipeline.WriteString("        stage('Security Scan') {\n")
		pipeline.WriteString("            steps {\n")
		pipeline.WriteString("                script {\n")
		pipeline.WriteString("                    sh 'echo \"Running security scans...\"'\n")
		switch spec.Language {
		case "go":
			pipeline.WriteString("                    sh 'go list -json -m all | nancy sleuth'\n")
		case "node":
			pipeline.WriteString("                    sh 'npm audit'\n")
		case "python":
			pipeline.WriteString("                    sh 'safety check'\n")
		}
		pipeline.WriteString("                }\n")
		pipeline.WriteString("            }\n")
		pipeline.WriteString("        }\n")
	}

	// Docker stage
	if spec.WithDocker {
		pipeline.WriteString("        stage('Docker Build') {\n")
		pipeline.WriteString("            steps {\n")
		pipeline.WriteString("                script {\n")
		pipeline.WriteString("                    def image = docker.build(\"${env.JOB_NAME}:${env.BUILD_NUMBER}\")\n")
		pipeline.WriteString("                    docker.withRegistry('https://registry.hub.docker.com', 'docker-hub-credentials') {\n")
		pipeline.WriteString("                        image.push()\n")
		pipeline.WriteString("                        image.push('latest')\n")
		pipeline.WriteString("                    }\n")
		pipeline.WriteString("                }\n")
		pipeline.WriteString("            }\n")
		pipeline.WriteString("        }\n")
	}

	// Deployment stages
	if spec.WithDeployment {
		for _, env := range spec.Environments {
			pipeline.WriteString(fmt.Sprintf("        stage('Deploy to %s') {\n", strings.Title(env)))

			// Add approval step for production
			if env == "production" {
				pipeline.WriteString("            input {\n")
				pipeline.WriteString("                message 'Deploy to production?'\n")
				pipeline.WriteString("                ok 'Deploy'\n")
				pipeline.WriteString("                submitterParameter 'DEPLOYER'\n")
				pipeline.WriteString("            }\n")
			}

			pipeline.WriteString("            steps {\n")
			pipeline.WriteString("                script {\n")
			pipeline.WriteString(fmt.Sprintf("                    echo 'Deploying to %s environment'\n", env))

			if spec.WithKubernetes {
				pipeline.WriteString(fmt.Sprintf("                    sh 'kubectl apply -f k8s/%s/ --validate=false'\n", env))
				pipeline.WriteString(fmt.Sprintf("                    sh 'kubectl rollout status deployment/app -n %s'\n", env))
			} else {
				pipeline.WriteString(fmt.Sprintf("                    sh 'echo \"Deploy to %s\"'\n", env))
			}

			pipeline.WriteString("                }\n")
			pipeline.WriteString("            }\n")
			pipeline.WriteString("        }\n")
		}
	}

	pipeline.WriteString("    }\n")

	// Post actions
	pipeline.WriteString("    post {\n")
	pipeline.WriteString("        always {\n")
	pipeline.WriteString("            cleanWs()\n")
	pipeline.WriteString("        }\n")
	pipeline.WriteString("        success {\n")
	pipeline.WriteString("            echo 'Pipeline succeeded!'\n")
	for _, notification := range spec.Notifications {
		pipeline.WriteString(fmt.Sprintf("            %s\n", notification))
	}
	pipeline.WriteString("        }\n")
	pipeline.WriteString("        failure {\n")
	pipeline.WriteString("            echo 'Pipeline failed!'\n")
	pipeline.WriteString("            emailext (\n")
	pipeline.WriteString("                subject: \"Build Failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER}\",\n")
	pipeline.WriteString("                body: \"Build failed. Check console output at ${env.BUILD_URL}\",\n")
	pipeline.WriteString("                to: \"${env.CHANGE_AUTHOR_EMAIL}\"\n")
	pipeline.WriteString("            )\n")
	pipeline.WriteString("        }\n")
	for _, action := range spec.PostActions {
		pipeline.WriteString(fmt.Sprintf("        %s\n", action))
	}
	pipeline.WriteString("    }\n")

	pipeline.WriteString("}")

	return pipeline.String(), nil
}

func generateScriptedPipeline(spec PipelineSpec) (string, error) {
	var pipeline strings.Builder

	// Node configuration
	if len(spec.Agents) > 0 {
		pipeline.WriteString(fmt.Sprintf("node('%s') {\n", spec.Agents[0]))
	} else {
		pipeline.WriteString("node {\n")
	}

	// Properties
	pipeline.WriteString("    properties([\n")
	pipeline.WriteString("        buildDiscarder(logRotator(numToKeepStr: '10')),\n")
	pipeline.WriteString(fmt.Sprintf("        timeout(time: %d, unit: 'MINUTES')\n", spec.TimeoutMinutes))
	pipeline.WriteString("    ])\n")

	// Try-catch block
	pipeline.WriteString("    try {\n")

	// Checkout stage
	pipeline.WriteString("        stage('Checkout') {\n")
	pipeline.WriteString("            checkout scm\n")
	pipeline.WriteString("        }\n")

	// Dependencies stage
	pipeline.WriteString("        stage('Dependencies') {\n")
	for _, cmd := range generateDependencyCommands(spec) {
		pipeline.WriteString(fmt.Sprintf("            sh '%s'\n", cmd))
	}
	pipeline.WriteString("        }\n")

	// Build stage
	pipeline.WriteString("        stage('Build') {\n")
	for _, cmd := range spec.BuildCommands {
		pipeline.WriteString(fmt.Sprintf("            sh '%s'\n", cmd))
	}
	pipeline.WriteString("        }\n")

	// Test stage
	if spec.WithTesting {
		pipeline.WriteString("        stage('Test') {\n")
		for _, cmd := range spec.TestCommands {
			pipeline.WriteString(fmt.Sprintf("            sh '%s'\n", cmd))
		}
		pipeline.WriteString("            publishTestResults testResultsPattern: 'test-results.xml'\n")
		pipeline.WriteString("        }\n")
	}

	// Security stage
	if spec.WithSecurity {
		pipeline.WriteString("        stage('Security') {\n")
		pipeline.WriteString("            sh 'echo \"Running security scans...\"'\n")
		pipeline.WriteString("        }\n")
	}

	// Docker stage
	if spec.WithDocker {
		pipeline.WriteString("        stage('Docker') {\n")
		pipeline.WriteString("            def image = docker.build(\"${env.JOB_NAME}:${env.BUILD_NUMBER}\")\n")
		pipeline.WriteString("            docker.withRegistry('https://registry.hub.docker.com', 'docker-hub-credentials') {\n")
		pipeline.WriteString("                image.push()\n")
		pipeline.WriteString("                image.push('latest')\n")
		pipeline.WriteString("            }\n")
		pipeline.WriteString("        }\n")
	}

	// Deployment stages
	if spec.WithDeployment {
		for _, env := range spec.Environments {
			pipeline.WriteString(fmt.Sprintf("        stage('Deploy %s') {\n", strings.Title(env)))
			if env == "production" {
				pipeline.WriteString("            input message: 'Deploy to production?', ok: 'Deploy'\n")
			}
			pipeline.WriteString(fmt.Sprintf("            echo 'Deploying to %s'\n", env))
			pipeline.WriteString("        }\n")
		}
	}

	// Catch block
	pipeline.WriteString("    } catch (Exception e) {\n")
	pipeline.WriteString("        currentBuild.result = 'FAILURE'\n")
	pipeline.WriteString("        throw e\n")
	pipeline.WriteString("    } finally {\n")
	pipeline.WriteString("        cleanWs()\n")
	pipeline.WriteString("    }\n")

	pipeline.WriteString("}")

	return pipeline.String(), nil
}

func generateSharedLibrary(spec PipelineSpec) error {
	if spec.LibraryName == "" {
		return fmt.Errorf("library name is required for shared library generation")
	}

	libDir := fmt.Sprintf("shared-library-%s", spec.LibraryName)

	// Create directory structure
	dirs := []string{
		filepath.Join(libDir, "vars"),
		filepath.Join(libDir, "src", "org", "example", spec.LibraryName),
		filepath.Join(libDir, "resources"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// Generate vars/deploy.groovy
	deployContent := generateDeployFunction(spec)
	if err := os.WriteFile(filepath.Join(libDir, "vars", "deploy.groovy"), []byte(deployContent), 0o644); err != nil {
		return err
	}

	// Generate vars/buildAndTest.groovy
	buildTestContent := generateBuildTestFunction(spec)
	if err := os.WriteFile(filepath.Join(libDir, "vars", "buildAndTest.groovy"), []byte(buildTestContent), 0o644); err != nil {
		return err
	}

	// Generate utility class
	utilsContent := generateUtilsClass(spec)
	utilsPath := filepath.Join(libDir, "src", "org", "example", spec.LibraryName, "Utils.groovy")
	if err := os.WriteFile(utilsPath, []byte(utilsContent), 0o644); err != nil {
		return err
	}

	// Generate README
	readmeContent := generateLibraryReadme(spec)
	if err := os.WriteFile(filepath.Join(libDir, "README.md"), []byte(readmeContent), 0o644); err != nil {
		return err
	}

	return nil
}

func generateDependencyCommands(spec PipelineSpec) []string {
	switch spec.Language {
	case "go":
		return []string{
			"go version",
			"go mod download",
		}
	case "node":
		return []string{
			"node --version",
			"npm --version",
			"npm ci",
		}
	case "python":
		return []string{
			"python --version",
			"pip install --upgrade pip",
			"pip install -r requirements.txt",
		}
	case "java":
		return []string{
			"java -version",
			"mvn --version",
			"mvn dependency:resolve",
		}
	case "rust":
		return []string{
			"rustc --version",
			"cargo --version",
			"cargo fetch",
		}
	case "php":
		return []string{
			"php --version",
			"composer --version",
			"composer install --no-dev",
		}
	default:
		return []string{
			"echo 'Setting up dependencies...'",
		}
	}
}

func generateDeployFunction(spec PipelineSpec) string {
	return fmt.Sprintf(`def call(Map config) {
    def environment = config.environment ?: 'staging'
    def image = config.image ?: "${env.JOB_NAME}:${env.BUILD_NUMBER}"
    
    stage("Deploy to ${environment}") {
        script {
            echo "Deploying ${image} to ${environment}"
            
            // Add deployment logic here
            if (environment == 'production') {
                input message: 'Deploy to production?', ok: 'Deploy'
            }
            
            // Example deployment commands
            sh "echo 'Deploying to ${environment}'"
            
            // If using Kubernetes
            if (config.kubernetes) {
                sh "kubectl apply -f k8s/${environment}/"
                sh "kubectl set image deployment/app app=${image}"
                sh "kubectl rollout status deployment/app"
            }
        }
    }
}`)
}

func generateBuildTestFunction(spec PipelineSpec) string {
	return fmt.Sprintf(`def call(Map config = [:]) {
    def language = config.language ?: '%s'
    def withTests = config.withTests ?: true
    
    stage('Build') {
        script {
            switch(language) {
                case 'go':
                    sh 'go build ./...'
                    break
                case 'node':
                    sh 'npm run build'
                    break
                case 'python':
                    sh 'python -m build'
                    break
                case 'java':
                    sh 'mvn compile'
                    break
                default:
                    sh 'make build'
            }
        }
    }
    
    if (withTests) {
        stage('Test') {
            script {
                switch(language) {
                    case 'go':
                        sh 'go test -v ./...'
                        break
                    case 'node':
                        sh 'npm test'
                        break
                    case 'python':
                        sh 'pytest'
                        break
                    case 'java':
                        sh 'mvn test'
                        break
                    default:
                        sh 'make test'
                }
            }
            post {
                always {
                    publishTestResults testResultsPattern: 'test-results.xml'
                }
            }
        }
    }
}`, spec.Language)
}

func generateUtilsClass(spec PipelineSpec) string {
	return fmt.Sprintf(`package org.example.%s

class Utils implements Serializable {
    def script
    
    Utils(script) {
        this.script = script
    }
    
    def getVersion() {
        return script.sh(
            script: 'git describe --tags --always',
            returnStdout: true
        ).trim()
    }
    
    def notifySlack(String message, String color = 'good') {
        script.slackSend(
            color: color,
            message: message,
            channel: '#deployments'
        )
    }
    
    def dockerBuild(String imageName, String tag = 'latest') {
        def image = script.docker.build("${imageName}:${tag}")
        return image
    }
    
    def dockerPush(def image, String registry = '') {
        if (registry) {
            script.docker.withRegistry("https://${registry}", 'docker-registry-credentials') {
                image.push()
            }
        } else {
            image.push()
        }
    }
    
    def kubernetesDeploy(String namespace, String manifests) {
        script.sh "kubectl apply -f ${manifests} -n ${namespace}"
        script.sh "kubectl rollout status deployment/app -n ${namespace}"
    }
}`, spec.LibraryName)
}

func generateLibraryReadme(spec PipelineSpec) string {
	return fmt.Sprintf(`# %s Shared Library

This Jenkins shared library provides reusable pipeline functions for %s projects.

## Functions

### deploy(Map config)
Deploy applications to various environments.

**Parameters:**
- `+"`environment`"+`: Target environment (staging, production)
- `+"`image`"+`: Docker image to deploy
- `+"`kubernetes`"+`: Enable Kubernetes deployment

**Example:**
`+"```groovy"+`
deploy([
    environment: 'staging',
    image: 'myapp:latest',
    kubernetes: true
])
`+"```"+`

### buildAndTest(Map config)
Build and test the application.

**Parameters:**
- `+"`language`"+`: Programming language
- `+"`withTests`"+`: Run tests (default: true)

**Example:**
`+"```groovy"+`
buildAndTest([
    language: '%s',
    withTests: true
])
`+"```"+`

## Usage

1. Configure this library in Jenkins Global Configuration
2. Use `+"`@Library('%s') _`"+` at the top of your Jenkinsfile
3. Call the functions in your pipeline

## Example Jenkinsfile

`+"```groovy"+`
@Library('%s') _

pipeline {
    agent any
    
    stages {
        stage('Build & Test') {
            steps {
                buildAndTest()
            }
        }
        
        stage('Deploy') {
            steps {
                deploy([
                    environment: 'staging',
                    kubernetes: true
                ])
            }
        }
    }
}
`+"```"+`
`, spec.LibraryName, spec.Language, spec.Language, spec.LibraryName, spec.LibraryName)
}

func generateDockerfiles(spec PipelineSpec) error {
	dockerfileContent := generateDockerfileContent(spec)
	return os.WriteFile("Dockerfile", []byte(dockerfileContent), 0o644)
}

func generateDockerfileContent(spec PipelineSpec) string {
	switch spec.Language {
	case "go":
		return `# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
`
	case "node":
		return `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
`
	case "python":
		return `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "app.py"]
`
	default:
		return `FROM alpine:latest
WORKDIR /app
COPY . .
CMD ["./start.sh"]
`
	}
}

func generateMultiBranchConfig(spec PipelineSpec) error {
	configContent := generateMultiBranchConfigContent(spec)
	return os.WriteFile("multibranch-config.xml", []byte(configContent), 0o644)
}

func generateMultiBranchConfigContent(spec PipelineSpec) string {
	return fmt.Sprintf(`<?xml version='1.1' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch">
  <actions/>
  <description>Multi-branch pipeline for %s project</description>
  <properties>
    <org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig plugin="pipeline-model-definition">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons"/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder">
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder">
    <pruneDeadBranches>true</pruneDeadBranches>
    <daysToKeep>-1</daysToKeep>
    <numToKeep>-1</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <disabled>false</disabled>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api">
    <data>
      <jenkins.branch.BranchSource>
        <source class="jenkins.plugins.git.GitSCMSource" plugin="git">
          <id>origin</id>
          <remote>https://github.com/example/repo.git</remote>
          <credentialsId></credentialsId>
          <traits>
            <jenkins.plugins.git.traits.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </jenkins.plugins.git.traits.BranchDiscoveryTrait>
            <jenkins.plugins.git.traits.OriginPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
            </jenkins.plugins.git.traits.OriginPullRequestDiscoveryTrait>
            <jenkins.plugins.git.traits.ForkPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
              <trust class="jenkins.plugins.git.traits.ForkPullRequestDiscoveryTrait$TrustPermission"/>
            </jenkins.plugins.git.traits.ForkPullRequestDiscoveryTrait>
          </traits>
        </source>
        <strategy class="jenkins.branch.DefaultBranchPropertyStrategy">
          <properties class="empty-list"/>
        </strategy>
      </jenkins.branch.BranchSource>
    </data>
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
    <scriptPath>Jenkinsfile</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`, spec.Language)
}
