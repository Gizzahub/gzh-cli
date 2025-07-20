package docker

import (
	"encoding/json"
	"os"
	"strings"
)

// LanguageOptimizer provides language-specific optimizations.
type LanguageOptimizer struct {
	Language string
	Project  ProjectInfo
}

// OptimizationConfig holds optimization settings.
type OptimizationConfig struct {
	BaseImage        string            `json:"baseImage"`
	BuildArgs        map[string]string `json:"buildArgs"`
	RuntimePackages  []string          `json:"runtimePackages"`
	BuildPackages    []string          `json:"buildPackages"`
	EnvironmentVars  map[string]string `json:"environmentVars"`
	ExposedPorts     []int             `json:"exposedPorts"`
	HealthCheckPath  string            `json:"healthCheckPath"`
	StartCommand     []string          `json:"startCommand"`
	SecuritySettings SecuritySettings  `json:"securitySettings"`
}

// SecuritySettings holds security-related configurations.
type SecuritySettings struct {
	RunAsNonRoot     bool     `json:"runAsNonRoot"`
	ReadOnlyRootFS   bool     `json:"readOnlyRootFs"`
	DropCapabilities []string `json:"dropCapabilities"`
	AddCapabilities  []string `json:"addCapabilities"`
	SeccompProfile   string   `json:"seccompProfile"`
	ApparmorProfile  string   `json:"apparmorProfile"`
}

// NewLanguageOptimizer creates a new language optimizer.
func NewLanguageOptimizer(language string, project ProjectInfo) *LanguageOptimizer {
	return &LanguageOptimizer{
		Language: language,
		Project:  project,
	}
}

// GetOptimizedConfig returns optimized configuration for the language.
func (lo *LanguageOptimizer) GetOptimizedConfig() OptimizationConfig {
	switch lo.Language {
	case langGo:
		return lo.getGoOptimizations()
	case langNode:
		return lo.getNodeOptimizations()
	case langPython:
		return lo.getPythonOptimizations()
	case langRuby:
		return lo.getRubyOptimizations()
	case "rust":
		return lo.getRustOptimizations()
	case langJava:
		return lo.getJavaOptimizations()
	default:
		return lo.getGenericOptimizations()
	}
}

func (lo *LanguageOptimizer) getGoOptimizations() OptimizationConfig {
	// Detect Go framework if any
	framework := lo.detectGoFramework()

	config := OptimizationConfig{
		BaseImage: "golang:1.21-alpine",
		BuildArgs: map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        "${TARGETOS:-linux}",
			"GOARCH":      "${TARGETARCH:-amd64}",
		},
		RuntimePackages: []string{"ca-certificates", "tzdata"},
		BuildPackages:   []string{"git", "curl"},
		EnvironmentVars: map[string]string{
			"PATH": "/app:$PATH",
		},
		ExposedPorts:    []int{8080},
		HealthCheckPath: "/health",
		StartCommand:    []string{"./app"},
		SecuritySettings: SecuritySettings{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   true,
			DropCapabilities: []string{"ALL"},
		},
	}

	// Framework-specific optimizations
	switch framework {
	case "gin":
		config.ExposedPorts = []int{8080}
		config.EnvironmentVars["GIN_MODE"] = "release"
	case "echo":
		config.ExposedPorts = []int{1323}
	case "fiber":
		config.ExposedPorts = []int{3000}
	case "chi":
		config.ExposedPorts = []int{8080}
	}

	return config
}

func (lo *LanguageOptimizer) getNodeOptimizations() OptimizationConfig {
	// Detect Node.js framework
	framework := lo.detectNodeFramework()

	config := OptimizationConfig{
		BaseImage: "node:20-alpine",
		BuildArgs: map[string]string{
			"NODE_ENV": "production",
		},
		RuntimePackages: []string{"tini", "curl"},
		BuildPackages:   []string{"python3", "make", "g++"},
		EnvironmentVars: map[string]string{
			"NODE_ENV":              "production",
			"NPM_CONFIG_LOGLEVEL":   "warn",
			"NPM_CONFIG_PRODUCTION": "true",
			"GENERATE_SOURCEMAP":    "false",
		},
		ExposedPorts:    []int{3000},
		HealthCheckPath: "/health",
		StartCommand:    []string{"node", "index.js"},
		SecuritySettings: SecuritySettings{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   false, // Node apps often need write access
			DropCapabilities: []string{"ALL"},
		},
	}

	// Framework-specific optimizations
	switch framework {
	case "express":
		config.ExposedPorts = []int{3000}
	case "fastify":
		config.ExposedPorts = []int{3000}
		config.EnvironmentVars["FASTIFY_CLOSE_GRACE_DELAY"] = "500"
	case "koa":
		config.ExposedPorts = []int{3000}
	case "nextjs":
		config.ExposedPorts = []int{3000}
		config.EnvironmentVars["NEXT_TELEMETRY_DISABLED"] = "1"
		config.StartCommand = []string{"node", "server.js"}
	case "nuxtjs":
		config.ExposedPorts = []int{3000}
		config.EnvironmentVars["NUXT_TELEMETRY_DISABLED"] = "1"
	}

	return config
}

func (lo *LanguageOptimizer) getPythonOptimizations() OptimizationConfig {
	// Detect Python framework
	framework := lo.detectPythonFramework()

	config := OptimizationConfig{
		BaseImage: "python:3.11-slim",
		BuildArgs: map[string]string{
			"PYTHONUNBUFFERED":              "1",
			"PYTHONDONTWRITEBYTECODE":       "1",
			"PIP_NO_CACHE_DIR":              "1",
			"PIP_DISABLE_PIP_VERSION_CHECK": "1",
		},
		RuntimePackages: []string{"curl"},
		BuildPackages:   []string{"build-essential", "git"},
		EnvironmentVars: map[string]string{
			"PYTHONPATH":              "/app",
			"PYTHONUNBUFFERED":        "1",
			"PYTHONDONTWRITEBYTECODE": "1",
		},
		ExposedPorts:    []int{8000},
		HealthCheckPath: "/health",
		StartCommand:    []string{"python", "app.py"},
		SecuritySettings: SecuritySettings{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   false, // Python apps often need write access
			DropCapabilities: []string{"ALL"},
		},
	}

	// Framework-specific optimizations
	switch framework {
	case "flask":
		config.ExposedPorts = []int{5000}
		config.EnvironmentVars["FLASK_ENV"] = "production"
		config.StartCommand = []string{"gunicorn", "--bind", "0.0.0.0:5000", "app:app"}
	case "django":
		config.ExposedPorts = []int{8000}
		config.EnvironmentVars["DJANGO_SETTINGS_MODULE"] = "settings.production"
		config.StartCommand = []string{"gunicorn", "--bind", "0.0.0.0:8000", "wsgi:application"}
	case "fastapi":
		config.ExposedPorts = []int{8000}
		config.StartCommand = []string{"uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"}
	case "tornado":
		config.ExposedPorts = []int{8888}
	}

	return config
}

func (lo *LanguageOptimizer) getRubyOptimizations() OptimizationConfig {
	// Detect Ruby framework
	framework := lo.detectRubyFramework()

	config := OptimizationConfig{
		BaseImage: "ruby:3.2-alpine",
		BuildArgs: map[string]string{
			"BUNDLE_DEPLOYMENT": "1",
			"BUNDLE_WITHOUT":    "development:test",
		},
		RuntimePackages: []string{"curl", "tzdata"},
		BuildPackages:   []string{"build-base", "git"},
		EnvironmentVars: map[string]string{
			"RAILS_ENV":           "production",
			"RACK_ENV":            "production",
			"RAILS_LOG_TO_STDOUT": "true",
		},
		ExposedPorts:    []int{3000},
		HealthCheckPath: "/health",
		StartCommand:    []string{"ruby", "app.rb"},
		SecuritySettings: SecuritySettings{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   false,
			DropCapabilities: []string{"ALL"},
		},
	}

	// Framework-specific optimizations
	switch framework {
	case "rails":
		config.ExposedPorts = []int{3000}
		config.StartCommand = []string{"rails", "server", "-b", "0.0.0.0"}
		config.EnvironmentVars["RAILS_SERVE_STATIC_FILES"] = "true"
	case "sinatra":
		config.ExposedPorts = []int{4567}
		config.StartCommand = []string{"ruby", "app.rb"}
	}

	return config
}

func (lo *LanguageOptimizer) getRustOptimizations() OptimizationConfig {
	// Detect Rust framework
	framework := lo.detectRustFramework()

	config := OptimizationConfig{
		BaseImage: "rust:1.70-alpine",
		BuildArgs: map[string]string{
			"CARGO_NET_GIT_FETCH_WITH_CLI": "true",
		},
		RuntimePackages: []string{"ca-certificates"},
		BuildPackages:   []string{"musl-dev", "git"},
		EnvironmentVars: map[string]string{
			"RUST_LOG": "info",
		},
		ExposedPorts:    []int{8080},
		HealthCheckPath: "/health",
		StartCommand:    []string{"./app"},
		SecuritySettings: SecuritySettings{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   true,
			DropCapabilities: []string{"ALL"},
		},
	}

	// Framework-specific optimizations
	switch framework {
	case "actix":
		config.ExposedPorts = []int{8080}
	case "warp":
		config.ExposedPorts = []int{3030}
	case "rocket":
		config.ExposedPorts = []int{8000}
	case "axum":
		config.ExposedPorts = []int{3000}
	}

	return config
}

func (lo *LanguageOptimizer) getJavaOptimizations() OptimizationConfig {
	// Detect Java framework
	framework := lo.detectJavaFramework()

	config := OptimizationConfig{
		BaseImage: "openjdk:21-jre-slim",
		BuildArgs: map[string]string{
			"MAVEN_OPTS": "-XX:+TieredCompilation -XX:TieredStopAtLevel=1",
		},
		RuntimePackages: []string{"curl"},
		BuildPackages:   []string{},
		EnvironmentVars: map[string]string{
			"JAVA_OPTS":              "-Xmx512m -Xms256m -XX:+UseContainerSupport",
			"SPRING_PROFILES_ACTIVE": "production",
		},
		ExposedPorts:    []int{8080},
		HealthCheckPath: "/actuator/health",
		StartCommand:    []string{"java", "-jar", "app.jar"},
		SecuritySettings: SecuritySettings{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   false,
			DropCapabilities: []string{"ALL"},
		},
	}

	// Framework-specific optimizations
	switch framework {
	case "spring":
		config.ExposedPorts = []int{8080}
		config.HealthCheckPath = "/actuator/health"
		config.EnvironmentVars["MANAGEMENT_ENDPOINTS_WEB_EXPOSURE_INCLUDE"] = "health,info"
	case "quarkus":
		config.ExposedPorts = []int{8080}
		config.EnvironmentVars["QUARKUS_HTTP_HOST"] = "0.0.0.0"
	case "micronaut":
		config.ExposedPorts = []int{8080}
	}

	return config
}

func (lo *LanguageOptimizer) getGenericOptimizations() OptimizationConfig {
	return OptimizationConfig{
		BaseImage:       "alpine:latest",
		BuildArgs:       map[string]string{},
		RuntimePackages: []string{"curl"},
		BuildPackages:   []string{},
		EnvironmentVars: map[string]string{},
		ExposedPorts:    []int{8080},
		HealthCheckPath: "/health",
		StartCommand:    []string{"./app"},
		SecuritySettings: SecuritySettings{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   true,
			DropCapabilities: []string{"ALL"},
		},
	}
}

// Framework detection methods.
func (lo *LanguageOptimizer) detectGoFramework() string {
	if lo.fileContains("go.mod", "gin-gonic/gin") {
		return "gin"
	}

	if lo.fileContains("go.mod", "labstack/echo") {
		return "echo"
	}

	if lo.fileContains("go.mod", "gofiber/fiber") {
		return "fiber"
	}

	if lo.fileContains("go.mod", "go-chi/chi") {
		return "chi"
	}

	return ""
}

func (lo *LanguageOptimizer) detectNodeFramework() string {
	if lo.fileContains("package.json", "express") {
		return "express"
	}

	if lo.fileContains("package.json", "fastify") {
		return "fastify"
	}

	if lo.fileContains("package.json", "koa") {
		return "koa"
	}

	if lo.fileContains("package.json", "next") {
		return "nextjs"
	}

	if lo.fileContains("package.json", "nuxt") {
		return "nuxtjs"
	}

	return ""
}

func (lo *LanguageOptimizer) detectPythonFramework() string {
	if lo.fileContains("requirements.txt", "Flask") || lo.fileContains("pyproject.toml", "flask") {
		return "flask"
	}

	if lo.fileContains("requirements.txt", "Django") || lo.fileContains("pyproject.toml", "django") {
		return "django"
	}

	if lo.fileContains("requirements.txt", "fastapi") || lo.fileContains("pyproject.toml", "fastapi") {
		return "fastapi"
	}

	if lo.fileContains("requirements.txt", "tornado") || lo.fileContains("pyproject.toml", "tornado") {
		return "tornado"
	}

	return ""
}

func (lo *LanguageOptimizer) detectRubyFramework() string {
	if lo.fileContains("Gemfile", "rails") {
		return "rails"
	}

	if lo.fileContains("Gemfile", "sinatra") {
		return "sinatra"
	}

	return ""
}

func (lo *LanguageOptimizer) detectRustFramework() string {
	if lo.fileContains("Cargo.toml", "actix-web") {
		return "actix"
	}

	if lo.fileContains("Cargo.toml", "warp") {
		return "warp"
	}

	if lo.fileContains("Cargo.toml", "rocket") {
		return "rocket"
	}

	if lo.fileContains("Cargo.toml", "axum") {
		return "axum"
	}

	return ""
}

func (lo *LanguageOptimizer) detectJavaFramework() string {
	if lo.fileContains("pom.xml", "spring-boot") || lo.fileContains("build.gradle", "spring-boot") {
		return "spring"
	}

	if lo.fileContains("pom.xml", "quarkus") || lo.fileContains("build.gradle", "quarkus") {
		return "quarkus"
	}

	if lo.fileContains("pom.xml", "micronaut") || lo.fileContains("build.gradle", "micronaut") {
		return "micronaut"
	}

	return ""
}

func (lo *LanguageOptimizer) fileContains(filename, text string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), text)
}

// GetOptimizationSummary returns a summary of applied optimizations.
func (lo *LanguageOptimizer) GetOptimizationSummary() map[string]interface{} {
	config := lo.GetOptimizedConfig()
	framework := ""

	switch lo.Language {
	case "go":
		framework = lo.detectGoFramework()
	case langNode:
		framework = lo.detectNodeFramework()
	case langPython:
		framework = lo.detectPythonFramework()
	case langRuby:
		framework = lo.detectRubyFramework()
	case langRust:
		framework = lo.detectRustFramework()
	case langJava:
		framework = lo.detectJavaFramework()
	}

	return map[string]interface{}{
		"language":           lo.Language,
		"detected_framework": framework,
		"base_image":         config.BaseImage,
		"exposed_ports":      config.ExposedPorts,
		"security_enabled":   config.SecuritySettings.RunAsNonRoot,
		"optimizations_applied": []string{
			"Multi-stage build",
			"Minimal runtime image",
			"Non-root user",
			"Security scanning",
			"Language-specific optimizations",
		},
	}
}

// SaveOptimizationConfig saves the optimization config to a file.
func (lo *LanguageOptimizer) SaveOptimizationConfig(filename string) error {
	config := lo.GetOptimizedConfig()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o644)
}
