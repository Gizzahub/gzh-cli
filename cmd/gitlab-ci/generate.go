package gitlabci

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate GitLab CI/CD pipelines",
	Long: `Generate GitLab CI/CD pipelines for various project types and deployment scenarios.

Supports generation of:
- Multi-stage CI/CD pipelines
- Language-specific build templates
- Docker and container workflows
- Kubernetes deployment pipelines
- Security scanning and compliance
- Multi-environment deployments
- Parallel and matrix builds
- Custom runner configurations

Examples:
  gz gitlab-ci generate --language go --with-docker
  gz gitlab-ci generate --type deploy --environment staging,production
  gz gitlab-ci generate --template microservice --with-security`,
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
	withCaching    bool
	environments   []string
	stages         []string
	pipelineVars   []string
	runners        []string
	templateType   string
	includeRules   bool
	parallelJobs   bool
	matrixBuild    bool
	customImage    string
	beforeScript   []string
	afterScript    []string
)

func init() {
	GenerateCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language (go, node, python, java, etc.)")
	GenerateCmd.Flags().StringVarP(&pipelineType, "type", "t", "ci", "Pipeline type (ci, deploy, release, test)")
	GenerateCmd.Flags().StringVarP(&outputFile, "output", "o", ".gitlab-ci.yml", "Output file path")
	GenerateCmd.Flags().BoolVar(&withDocker, "with-docker", false, "Include Docker build stages")
	GenerateCmd.Flags().BoolVar(&withKubernetes, "with-k8s", false, "Include Kubernetes deployment")
	GenerateCmd.Flags().BoolVar(&withSecurity, "with-security", true, "Include security scanning")
	GenerateCmd.Flags().BoolVar(&withTesting, "with-testing", true, "Include testing stages")
	GenerateCmd.Flags().BoolVar(&withCaching, "with-caching", true, "Enable pipeline caching")
	GenerateCmd.Flags().StringSliceVar(&environments, "environments", []string{"staging", "production"}, "Deployment environments")
	GenerateCmd.Flags().StringSliceVar(&stages, "stages", []string{}, "Custom pipeline stages")
	GenerateCmd.Flags().StringSliceVar(&pipelineVars, "variables", []string{}, "Pipeline variables (key=value)")
	GenerateCmd.Flags().StringSliceVar(&runners, "runners", []string{}, "Runner tags")
	GenerateCmd.Flags().StringVar(&templateType, "template", "standard", "Template type (standard, microservice, monorepo)")
	GenerateCmd.Flags().BoolVar(&includeRules, "include-rules", true, "Include pipeline rules and conditions")
	GenerateCmd.Flags().BoolVar(&parallelJobs, "parallel-jobs", false, "Enable parallel job execution")
	GenerateCmd.Flags().BoolVar(&matrixBuild, "matrix-build", false, "Use matrix build strategy")
	GenerateCmd.Flags().StringVar(&customImage, "image", "", "Custom Docker image for jobs")
	GenerateCmd.Flags().StringSliceVar(&beforeScript, "before-script", []string{}, "Global before_script commands")
	GenerateCmd.Flags().StringSliceVar(&afterScript, "after-script", []string{}, "Global after_script commands")
}

// PipelineSpec holds pipeline generation specifications
type PipelineSpec struct {
	Language       string
	Type           string
	WithDocker     bool
	WithKubernetes bool
	WithSecurity   bool
	WithTesting    bool
	WithCaching    bool
	Environments   []string
	Stages         []string
	Variables      map[string]string
	Runners        []string
	TemplateType   string
	IncludeRules   bool
	ParallelJobs   bool
	MatrixBuild    bool
	CustomImage    string
	BeforeScript   []string
	AfterScript    []string
	PackageManager string
	BuildCommands  []string
	TestCommands   []string
	LintCommands   []string
}

// GitLabCIConfig represents a complete GitLab CI configuration
type GitLabCIConfig struct {
	Image        interface{}            `yaml:"image,omitempty"`
	Services     []ServiceConfig        `yaml:"services,omitempty"`
	BeforeScript []string               `yaml:"before_script,omitempty"`
	AfterScript  []string               `yaml:"after_script,omitempty"`
	Stages       []string               `yaml:"stages,omitempty"`
	Variables    map[string]interface{} `yaml:"variables,omitempty"`
	Cache        *CacheConfig           `yaml:"cache,omitempty"`
	Include      []IncludeConfig        `yaml:"include,omitempty"`
	Jobs         map[string]JobConfig   `yaml:",inline"`
}

type ServiceConfig struct {
	Name  string            `yaml:"name,omitempty"`
	Alias string            `yaml:"alias,omitempty"`
	Image string            `yaml:"image,omitempty"`
	Env   map[string]string `yaml:"environment,omitempty"`
}

type CacheConfig struct {
	Key    interface{} `yaml:"key,omitempty"`
	Paths  []string    `yaml:"paths,omitempty"`
	Policy string      `yaml:"policy,omitempty"`
	When   string      `yaml:"when,omitempty"`
}

type IncludeConfig struct {
	Template string `yaml:"template,omitempty"`
	Local    string `yaml:"local,omitempty"`
	Remote   string `yaml:"remote,omitempty"`
	Project  string `yaml:"project,omitempty"`
	File     string `yaml:"file,omitempty"`
}

type JobConfig struct {
	Stage              string                 `yaml:"stage,omitempty"`
	Image              interface{}            `yaml:"image,omitempty"`
	Services           []ServiceConfig        `yaml:"services,omitempty"`
	BeforeScript       []string               `yaml:"before_script,omitempty"`
	Script             []string               `yaml:"script,omitempty"`
	AfterScript        []string               `yaml:"after_script,omitempty"`
	Variables          map[string]interface{} `yaml:"variables,omitempty"`
	Cache              *CacheConfig           `yaml:"cache,omitempty"`
	Artifacts          *ArtifactsConfig       `yaml:"artifacts,omitempty"`
	Dependencies       []string               `yaml:"dependencies,omitempty"`
	Needs              []interface{}          `yaml:"needs,omitempty"`
	Rules              []RuleConfig           `yaml:"rules,omitempty"`
	Only               *OnlyConfig            `yaml:"only,omitempty"`
	Except             *ExceptConfig          `yaml:"except,omitempty"`
	Tags               []string               `yaml:"tags,omitempty"`
	AllowFailure       bool                   `yaml:"allow_failure,omitempty"`
	When               string                 `yaml:"when,omitempty"`
	Environment        *EnvironmentConfig     `yaml:"environment,omitempty"`
	Coverage           string                 `yaml:"coverage,omitempty"`
	Retry              *RetryConfig           `yaml:"retry,omitempty"`
	Timeout            string                 `yaml:"timeout,omitempty"`
	Parallel           interface{}            `yaml:"parallel,omitempty"`
	InterruptibleValue bool                   `yaml:"interruptible,omitempty"`
}

type ArtifactsConfig struct {
	Name     string            `yaml:"name,omitempty"`
	Paths    []string          `yaml:"paths,omitempty"`
	Exclude  []string          `yaml:"exclude,omitempty"`
	ExpireIn string            `yaml:"expire_in,omitempty"`
	When     string            `yaml:"when,omitempty"`
	Reports  *ArtifactsReports `yaml:"reports,omitempty"`
	Expose   string            `yaml:"expose_as,omitempty"`
}

type ArtifactsReports struct {
	JUnit       []string `yaml:"junit,omitempty"`
	Coverage    []string `yaml:"cobertura,omitempty"`
	Codequality []string `yaml:"codequality,omitempty"`
	SAST        []string `yaml:"sast,omitempty"`
	Dependency  []string `yaml:"dependency_scanning,omitempty"`
	Container   []string `yaml:"container_scanning,omitempty"`
	Performance []string `yaml:"performance,omitempty"`
	LoadTesting []string `yaml:"load_performance,omitempty"`
}

type RuleConfig struct {
	If           string            `yaml:"if,omitempty"`
	Changes      []string          `yaml:"changes,omitempty"`
	Exists       []string          `yaml:"exists,omitempty"`
	Variables    map[string]string `yaml:"variables,omitempty"`
	When         string            `yaml:"when,omitempty"`
	AllowFailure bool              `yaml:"allow_failure,omitempty"`
}

type OnlyConfig struct {
	Refs       []string `yaml:"refs,omitempty"`
	Variables  []string `yaml:"variables,omitempty"`
	Changes    []string `yaml:"changes,omitempty"`
	Kubernetes string   `yaml:"kubernetes,omitempty"`
}

type ExceptConfig struct {
	Refs      []string `yaml:"refs,omitempty"`
	Variables []string `yaml:"variables,omitempty"`
	Changes   []string `yaml:"changes,omitempty"`
}

type EnvironmentConfig struct {
	Name           string `yaml:"name,omitempty"`
	URL            string `yaml:"url,omitempty"`
	Action         string `yaml:"action,omitempty"`
	OnStop         string `yaml:"on_stop,omitempty"`
	AutoStop       string `yaml:"auto_stop_in,omitempty"`
	KubernetesNS   string `yaml:"kubernetes,omitempty"`
	DeploymentTier string `yaml:"deployment_tier,omitempty"`
}

type RetryConfig struct {
	Max  int      `yaml:"max,omitempty"`
	When []string `yaml:"when,omitempty"`
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

	fmt.Printf("🚀 Generating GitLab CI/CD pipeline\n")
	fmt.Printf("📋 Language: %s, Type: %s, Template: %s\n", language, pipelineType, templateType)

	// Parse variables
	variableMap := parsePipelineVariables(pipelineVars)

	// Create pipeline specification
	spec := PipelineSpec{
		Language:       language,
		Type:           pipelineType,
		WithDocker:     withDocker,
		WithKubernetes: withKubernetes,
		WithSecurity:   withSecurity,
		WithTesting:    withTesting,
		WithCaching:    withCaching,
		Environments:   environments,
		Stages:         stages,
		Variables:      variableMap,
		Runners:        runners,
		TemplateType:   templateType,
		IncludeRules:   includeRules,
		ParallelJobs:   parallelJobs,
		MatrixBuild:    matrixBuild,
		CustomImage:    customImage,
		BeforeScript:   beforeScript,
		AfterScript:    afterScript,
	}

	// Enhance spec with language-specific information
	if err := enhanceSpecForLanguage(&spec); err != nil {
		fmt.Printf("Error enhancing spec: %v\n", err)
		os.Exit(1)
	}

	// Generate pipeline
	pipeline, err := generatePipeline(spec)
	if err != nil {
		fmt.Printf("Error generating pipeline: %v\n", err)
		os.Exit(1)
	}

	// Write pipeline file
	if err := writePipeline(pipeline, outputFile); err != nil {
		fmt.Printf("Error writing pipeline: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated GitLab CI pipeline: %s\n", outputFile)

	// Generate additional files if requested
	if withDocker {
		if err := generateDockerfiles(spec); err != nil {
			fmt.Printf("Warning: Failed to generate Dockerfiles: %v\n", err)
		} else {
			fmt.Println("✅ Generated Dockerfile templates")
		}
	}

	if withKubernetes {
		if err := generateK8sManifests(spec); err != nil {
			fmt.Printf("Warning: Failed to generate K8s manifests: %v\n", err)
		} else {
			fmt.Println("✅ Generated Kubernetes manifests")
		}
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Review and customize the generated pipeline")
	fmt.Println("2. Set up GitLab CI/CD variables and secrets")
	fmt.Println("3. Configure GitLab Runners with appropriate tags")
	fmt.Println("4. Test pipeline with a sample commit")
	fmt.Printf("5. Monitor pipeline execution in GitLab: %s/-/pipelines\n", "${CI_PROJECT_URL}")
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

func parsePipelineVariables(vars []string) map[string]string {
	result := make(map[string]string)
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func enhanceSpecForLanguage(spec *PipelineSpec) error {
	switch spec.Language {
	case "go":
		spec.PackageManager = "go modules"
		spec.BuildCommands = []string{"go build ./..."}
		spec.TestCommands = []string{"go test -v ./..."}
		spec.LintCommands = []string{"golangci-lint run"}
		if spec.CustomImage == "" {
			spec.CustomImage = "golang:1.21"
		}
	case "node", "javascript", "typescript":
		spec.PackageManager = "npm"
		spec.BuildCommands = []string{"npm run build"}
		spec.TestCommands = []string{"npm test"}
		spec.LintCommands = []string{"npm run lint"}
		if spec.CustomImage == "" {
			spec.CustomImage = "node:18"
		}
	case "python":
		spec.PackageManager = "pip"
		spec.BuildCommands = []string{"python -m build"}
		spec.TestCommands = []string{"pytest"}
		spec.LintCommands = []string{"flake8", "black --check ."}
		if spec.CustomImage == "" {
			spec.CustomImage = "python:3.11"
		}
	case "java":
		spec.PackageManager = "maven"
		spec.BuildCommands = []string{"mvn compile"}
		spec.TestCommands = []string{"mvn test"}
		spec.LintCommands = []string{"mvn checkstyle:check"}
		if spec.CustomImage == "" {
			spec.CustomImage = "maven:3.8-openjdk-17"
		}
	case "rust":
		spec.PackageManager = "cargo"
		spec.BuildCommands = []string{"cargo build"}
		spec.TestCommands = []string{"cargo test"}
		spec.LintCommands = []string{"cargo clippy"}
		if spec.CustomImage == "" {
			spec.CustomImage = "rust:1.70"
		}
	case "php":
		spec.PackageManager = "composer"
		spec.BuildCommands = []string{"composer install --no-dev --optimize-autoloader"}
		spec.TestCommands = []string{"vendor/bin/phpunit"}
		spec.LintCommands = []string{"vendor/bin/phpcs"}
		if spec.CustomImage == "" {
			spec.CustomImage = "php:8.2"
		}
	default:
		spec.PackageManager = "generic"
		spec.BuildCommands = []string{"make build"}
		spec.TestCommands = []string{"make test"}
		spec.LintCommands = []string{"make lint"}
		if spec.CustomImage == "" {
			spec.CustomImage = "alpine:latest"
		}
	}

	// Set default stages if not specified
	if len(spec.Stages) == 0 {
		spec.Stages = generateDefaultStages(spec)
	}

	return nil
}

func generateDefaultStages(spec *PipelineSpec) []string {
	stages := []string{"prepare", "build"}

	if spec.WithTesting {
		stages = append(stages, "test")
	}

	if spec.WithSecurity {
		stages = append(stages, "security")
	}

	if spec.WithDocker {
		stages = append(stages, "package")
	}

	// Add environment-specific deploy stages
	for _, env := range spec.Environments {
		stages = append(stages, "deploy:"+env)
	}

	stages = append(stages, "cleanup")

	return stages
}

func generatePipeline(spec PipelineSpec) (*GitLabCIConfig, error) {
	pipeline := &GitLabCIConfig{
		Stages:    spec.Stages,
		Variables: make(map[string]interface{}),
		Jobs:      make(map[string]JobConfig),
	}

	// Set global image
	if spec.CustomImage != "" {
		pipeline.Image = spec.CustomImage
	}

	// Set global before/after scripts
	if len(spec.BeforeScript) > 0 {
		pipeline.BeforeScript = spec.BeforeScript
	}
	if len(spec.AfterScript) > 0 {
		pipeline.AfterScript = spec.AfterScript
	}

	// Add variables
	for k, v := range spec.Variables {
		pipeline.Variables[k] = v
	}

	// Add language-specific variables
	addLanguageVariables(pipeline, spec)

	// Generate cache configuration
	if spec.WithCaching {
		pipeline.Cache = generateCacheConfig(spec)
	}

	// Generate jobs based on pipeline type
	switch spec.Type {
	case "ci":
		addCIJobs(pipeline, spec)
	case "deploy":
		addDeploymentJobs(pipeline, spec)
	case "release":
		addReleaseJobs(pipeline, spec)
	case "test":
		addTestJobs(pipeline, spec)
	default:
		addCIJobs(pipeline, spec)
	}

	return pipeline, nil
}

func addLanguageVariables(pipeline *GitLabCIConfig, spec PipelineSpec) {
	switch spec.Language {
	case "go":
		pipeline.Variables["CGO_ENABLED"] = "0"
		pipeline.Variables["GOOS"] = "linux"
		pipeline.Variables["GOARCH"] = "amd64"
	case "node":
		pipeline.Variables["NODE_ENV"] = "production"
		pipeline.Variables["NPM_CONFIG_CACHE"] = ".npm"
	case "python":
		pipeline.Variables["PIP_CACHE_DIR"] = ".pip-cache"
		pipeline.Variables["PYTHONDONTWRITEBYTECODE"] = "1"
	case "java":
		pipeline.Variables["MAVEN_OPTS"] = "-Dmaven.repo.local=.m2/repository"
		pipeline.Variables["MAVEN_CLI_OPTS"] = "--batch-mode --errors --fail-at-end --show-version"
	}
}

func generateCacheConfig(spec PipelineSpec) *CacheConfig {
	cache := &CacheConfig{
		Key:    "$CI_COMMIT_REF_SLUG",
		Policy: "pull-push",
	}

	switch spec.Language {
	case "go":
		cache.Paths = []string{".go/pkg/mod/"}
	case "node":
		cache.Paths = []string{".npm/", "node_modules/"}
	case "python":
		cache.Paths = []string{".pip-cache/", "venv/"}
	case "java":
		cache.Paths = []string{".m2/repository/"}
	case "rust":
		cache.Paths = []string{"target/", ".cargo/"}
	case "php":
		cache.Paths = []string{"vendor/"}
	default:
		cache.Paths = []string{".cache/"}
	}

	return cache
}

func addCIJobs(pipeline *GitLabCIConfig, spec PipelineSpec) {
	// Prepare job
	prepareJob := JobConfig{
		Stage:  "prepare",
		Script: generatePrepareScript(spec),
		Cache: &CacheConfig{
			Key:    "$CI_COMMIT_REF_SLUG",
			Paths:  pipeline.Cache.Paths,
			Policy: "push",
		},
	}

	if len(spec.Runners) > 0 {
		prepareJob.Tags = spec.Runners
	}

	pipeline.Jobs["prepare:dependencies"] = prepareJob

	// Build job
	buildJob := JobConfig{
		Stage:        "build",
		Script:       generateBuildScript(spec),
		Dependencies: []string{"prepare:dependencies"},
		Cache: &CacheConfig{
			Key:    "$CI_COMMIT_REF_SLUG",
			Paths:  pipeline.Cache.Paths,
			Policy: "pull",
		},
	}

	if spec.WithDocker {
		buildJob.Artifacts = &ArtifactsConfig{
			Paths:    []string{"dist/", "build/"},
			ExpireIn: "1 week",
		}
	}

	pipeline.Jobs["build"] = buildJob

	// Test jobs
	if spec.WithTesting {
		addTestJobs(pipeline, spec)
	}

	// Security jobs
	if spec.WithSecurity {
		addSecurityJobs(pipeline, spec)
	}

	// Docker build job
	if spec.WithDocker {
		addDockerJobs(pipeline, spec)
	}
}

func addTestJobs(pipeline *GitLabCIConfig, spec PipelineSpec) {
	// Unit tests
	unitTestJob := JobConfig{
		Stage:        "test",
		Script:       generateTestScript(spec),
		Dependencies: []string{"build"},
		Coverage:     "/coverage: (\\d+\\.\\d+)%/",
		Artifacts: &ArtifactsConfig{
			Reports: &ArtifactsReports{
				JUnit:    []string{"reports/junit.xml"},
				Coverage: []string{"reports/coverage.xml"},
			},
			ExpireIn: "30 days",
		},
	}

	if spec.IncludeRules {
		unitTestJob.Rules = []RuleConfig{
			{
				If:   "$CI_PIPELINE_SOURCE == 'merge_request_event'",
				When: "always",
			},
			{
				If:   "$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH",
				When: "always",
			},
		}
	}

	pipeline.Jobs["test:unit"] = unitTestJob

	// Lint job
	if len(spec.LintCommands) > 0 {
		lintJob := JobConfig{
			Stage:        "test",
			Script:       generateLintScript(spec),
			Dependencies: []string{"prepare:dependencies"},
			AllowFailure: true,
		}
		pipeline.Jobs["test:lint"] = lintJob
	}

	// Integration tests (if specified)
	if spec.TemplateType == "microservice" {
		integrationJob := JobConfig{
			Stage:  "test",
			Script: generateIntegrationTestScript(spec),
			Services: []ServiceConfig{
				{
					Name:  "postgres:13",
					Alias: "postgres",
				},
				{
					Name:  "redis:6",
					Alias: "redis",
				},
			},
			Variables: map[string]interface{}{
				"POSTGRES_DB":       "test",
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
			},
		}
		pipeline.Jobs["test:integration"] = integrationJob
	}
}

func addSecurityJobs(pipeline *GitLabCIConfig, spec PipelineSpec) {
	// SAST job
	sastJob := JobConfig{
		Stage: "security",
		Script: []string{
			"echo 'Running SAST scan...'",
			"# Include SAST template in main pipeline config",
		},
	}
	pipeline.Jobs["sast"] = sastJob

	// Dependency scanning
	depScanJob := JobConfig{
		Stage: "security",
		Script: []string{
			"echo 'Running dependency scan...'",
			"# Include dependency scanning template",
		},
	}
	pipeline.Jobs["dependency_scanning"] = depScanJob

	// Secret detection
	secretJob := JobConfig{
		Stage: "security",
		Script: []string{
			"echo 'Running secret detection...'",
			"# Include secret detection template",
		},
	}
	pipeline.Jobs["secret_detection"] = secretJob

	// Container scanning (if Docker is enabled)
	if spec.WithDocker {
		containerScanJob := JobConfig{
			Stage: "security",
			Script: []string{
				"echo 'Running container scan...'",
				"# Include container scanning template",
			},
		}
		pipeline.Jobs["container_scanning"] = containerScanJob
	}
}

func addDockerJobs(pipeline *GitLabCIConfig, spec PipelineSpec) {
	dockerJob := JobConfig{
		Stage: "package",
		Image: "docker:24.0.5",
		Services: []ServiceConfig{
			{
				Name: "docker:24.0.5-dind",
			},
		},
		Variables: map[string]interface{}{
			"DOCKER_TLS_CERTDIR": "/certs",
			"DOCKER_DRIVER":      "overlay2",
		},
		BeforeScript: []string{
			"docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY",
		},
		Script:       generateDockerScript(spec),
		Dependencies: []string{"build"},
	}

	if spec.IncludeRules {
		dockerJob.Rules = []RuleConfig{
			{
				If:   "$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH",
				When: "always",
			},
			{
				If:   "$CI_COMMIT_TAG",
				When: "always",
			},
		}
	}

	pipeline.Jobs["docker:build"] = dockerJob
}

func addDeploymentJobs(pipeline *GitLabCIConfig, spec PipelineSpec) {
	for _, env := range spec.Environments {
		deployJob := JobConfig{
			Stage:  "deploy:" + env,
			Script: generateDeployScript(spec, env),
			Environment: &EnvironmentConfig{
				Name: env,
				URL:  fmt.Sprintf("https://%s.example.com", env),
			},
		}

		// Add deployment rules
		if spec.IncludeRules {
			switch env {
			case "staging":
				deployJob.Rules = []RuleConfig{
					{
						If:   "$CI_COMMIT_BRANCH == 'develop'",
						When: "always",
					},
				}
			case "production":
				deployJob.Rules = []RuleConfig{
					{
						If:   "$CI_COMMIT_TAG",
						When: "manual",
					},
					{
						If:   "$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH",
						When: "manual",
					},
				}
			}
		}

		// Add Kubernetes deployment if enabled
		if spec.WithKubernetes {
			deployJob.Script = append(deployJob.Script, generateKubernetesDeployScript(spec, env)...)
		}

		pipeline.Jobs[fmt.Sprintf("deploy:%s", env)] = deployJob
	}
}

func addReleaseJobs(pipeline *GitLabCIConfig, spec PipelineSpec) {
	releaseJob := JobConfig{
		Stage: "deploy:production",
		Script: []string{
			"echo 'Creating release...'",
			"apk add --no-cache curl",
			"curl --request POST --header \"PRIVATE-TOKEN: $CI_JOB_TOKEN\" \"$CI_API_V4_URL/projects/$CI_PROJECT_ID/releases\" --data \"tag_name=$CI_COMMIT_TAG&description=Release%20$CI_COMMIT_TAG\"",
		},
		Rules: []RuleConfig{
			{
				If:   "$CI_COMMIT_TAG",
				When: "always",
			},
		},
	}
	pipeline.Jobs["release"] = releaseJob
}

func generatePrepareScript(spec PipelineSpec) []string {
	var script []string

	switch spec.Language {
	case "go":
		script = []string{
			"go version",
			"go mod download",
		}
	case "node":
		script = []string{
			"node --version",
			"npm --version",
			"npm ci --cache .npm --prefer-offline",
		}
	case "python":
		script = []string{
			"python --version",
			"pip install --upgrade pip",
			"pip install -r requirements.txt --cache-dir .pip-cache",
		}
	case "java":
		script = []string{
			"java -version",
			"mvn $MAVEN_CLI_OPTS dependency:resolve",
		}
	case "rust":
		script = []string{
			"rustc --version",
			"cargo --version",
			"cargo fetch",
		}
	case "php":
		script = []string{
			"php --version",
			"composer --version",
			"composer install --no-dev --optimize-autoloader",
		}
	default:
		script = []string{
			"echo 'Preparing dependencies...'",
		}
	}

	return script
}

func generateBuildScript(spec PipelineSpec) []string {
	script := []string{"echo 'Building project...'"}
	script = append(script, spec.BuildCommands...)
	return script
}

func generateTestScript(spec PipelineSpec) []string {
	script := []string{"echo 'Running tests...'"}
	script = append(script, spec.TestCommands...)

	// Add coverage generation based on language
	switch spec.Language {
	case "go":
		script = append(script, "go test -v -race -coverprofile=coverage.out ./...")
		script = append(script, "go tool cover -html=coverage.out -o coverage.html")
	case "node":
		script = append(script, "npm run test:coverage")
	case "python":
		script = append(script, "pytest --cov=. --cov-report=xml --cov-report=html")
	}

	return script
}

func generateLintScript(spec PipelineSpec) []string {
	script := []string{"echo 'Running linters...'"}
	script = append(script, spec.LintCommands...)
	return script
}

func generateIntegrationTestScript(spec PipelineSpec) []string {
	return []string{
		"echo 'Running integration tests...'",
		"sleep 10", // Wait for services to start
		"npm run test:integration || pytest tests/integration/ || go test -tags=integration ./...",
	}
}

func generateDockerScript(spec PipelineSpec) []string {
	return []string{
		"docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .",
		"docker tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:latest",
		"docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA",
		"docker push $CI_REGISTRY_IMAGE:latest",
	}
}

func generateDeployScript(spec PipelineSpec, environment string) []string {
	script := []string{
		fmt.Sprintf("echo 'Deploying to %s environment...'", environment),
	}

	if spec.WithDocker {
		script = append(script,
			"docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY",
			fmt.Sprintf("docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA"),
			fmt.Sprintf("docker run -d --name app-%s -p 8080:8080 $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA", environment),
		)
	}

	return script
}

func generateKubernetesDeployScript(spec PipelineSpec, environment string) []string {
	return []string{
		"kubectl config use-context " + environment,
		"kubectl set image deployment/app app=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA",
		"kubectl rollout status deployment/app",
	}
}

func writePipeline(pipeline *GitLabCIConfig, filePath string) error {
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

func generateDockerfiles(spec PipelineSpec) error {
	// Generate basic Dockerfile based on language
	dockerfileContent := generateDockerfileContent(spec)

	if err := os.WriteFile("Dockerfile", []byte(dockerfileContent), 0o644); err != nil {
		return err
	}

	// Generate .dockerignore
	dockerignoreContent := generateDockerignoreContent(spec)
	return os.WriteFile(".dockerignore", []byte(dockerignoreContent), 0o644)
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

func generateDockerignoreContent(spec PipelineSpec) string {
	common := `.git
.gitignore
README.md
Dockerfile
.dockerignore
node_modules
npm-debug.log
coverage/
.nyc_output
`

	switch spec.Language {
	case "go":
		return common + `
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
vendor/
`
	case "python":
		return common + `
__pycache__
*.pyc
*.pyo
*.pyd
.Python
env/
venv/
.venv/
pip-log.txt
pip-delete-this-directory.txt
`
	case "java":
		return common + `
target/
*.jar
*.war
*.ear
*.class
`
	default:
		return common
	}
}

func generateK8sManifests(spec PipelineSpec) error {
	// Create k8s directory
	if err := os.MkdirAll("k8s", 0o755); err != nil {
		return err
	}

	// Generate deployment manifest
	deploymentContent := generateK8sDeployment(spec)
	if err := os.WriteFile("k8s/deployment.yaml", []byte(deploymentContent), 0o644); err != nil {
		return err
	}

	// Generate service manifest
	serviceContent := generateK8sService(spec)
	if err := os.WriteFile("k8s/service.yaml", []byte(serviceContent), 0o644); err != nil {
		return err
	}

	// Generate ingress manifest
	ingressContent := generateK8sIngress(spec)
	return os.WriteFile("k8s/ingress.yaml", []byte(ingressContent), 0o644)
}

func generateK8sDeployment(spec PipelineSpec) string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: ${CI_REGISTRY_IMAGE}:${CI_COMMIT_SHA}
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
`
}

func generateK8sService(spec PipelineSpec) string {
	return `apiVersion: v1
kind: Service
metadata:
  name: app-service
spec:
  selector:
    app: app
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP
`
}

func generateK8sIngress(spec PipelineSpec) string {
	return `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - app.example.com
    secretName: app-tls
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: app-service
            port:
              number: 80
`
}
