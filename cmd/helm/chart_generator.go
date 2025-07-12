package helm

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ChartCmd represents the chart command
var ChartCmd = &cobra.Command{
	Use:   "chart",
	Short: "Generate Helm chart for projects",
	Long: `Generate optimized Helm chart templates based on project requirements.

Supports automatic detection of project type and generates appropriate Helm chart with:
- Chart template library with best practices
- Values file management system
- Dependency chart handling
- Kubernetes resource templates

Examples:
  gz helm chart                           # Generate chart in ./chart directory
  gz helm chart --name myapp --output ./charts/myapp
  gz helm chart --values custom-values.yaml`,
	Run: runChartGenerate,
}

var (
	chartName        string
	chartOutput      string
	chartVersion     string
	appVersion       string
	valuesFile       string
	includeDeps      bool
	includeIngress   bool
	includeService   bool
	includeHPA       bool
	includePDB       bool
	chartDescription string
	namespace        string
)

func init() {
	ChartCmd.Flags().StringVarP(&chartName, "name", "n", "", "Chart name (auto-detect from directory if not specified)")
	ChartCmd.Flags().StringVarP(&chartOutput, "output", "o", "./chart", "Output directory for generated chart")
	ChartCmd.Flags().StringVarP(&chartVersion, "chart-version", "c", "0.1.0", "Chart version")
	ChartCmd.Flags().StringVarP(&appVersion, "app-version", "a", "1.0.0", "Application version")
	ChartCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "Custom values file to include")
	ChartCmd.Flags().BoolVar(&includeDeps, "include-deps", false, "Include common dependencies (redis, postgresql)")
	ChartCmd.Flags().BoolVar(&includeIngress, "include-ingress", true, "Include Ingress resource")
	ChartCmd.Flags().BoolVar(&includeService, "include-service", true, "Include Service resource")
	ChartCmd.Flags().BoolVar(&includeHPA, "include-hpa", false, "Include HorizontalPodAutoscaler")
	ChartCmd.Flags().BoolVar(&includePDB, "include-pdb", false, "Include PodDisruptionBudget")
	ChartCmd.Flags().StringVar(&chartDescription, "description", "", "Chart description")
	ChartCmd.Flags().StringVar(&namespace, "namespace", "default", "Target namespace")
}

// ProjectInfo holds detected project information for Helm chart generation
type ProjectInfo struct {
	Name        string
	Language    string
	Framework   string
	Port        int
	HasDatabase bool
	HasRedis    bool
	HasIngress  bool
	HasService  bool
}

// ChartTemplate holds chart template data
type ChartTemplate struct {
	Project        ProjectInfo
	ChartName      string
	ChartVersion   string
	AppVersion     string
	Description    string
	Namespace      string
	IncludeDeps    bool
	IncludeIngress bool
	IncludeService bool
	IncludeHPA     bool
	IncludePDB     bool
	Dependencies   []Dependency
	Values         map[string]interface{}
}

// Dependency represents a Helm chart dependency
type Dependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
	Condition  string `yaml:"condition,omitempty"`
}

// ChartMetadata represents Chart.yaml content
type ChartMetadata struct {
	APIVersion   string       `yaml:"apiVersion"`
	Name         string       `yaml:"name"`
	Description  string       `yaml:"description"`
	Type         string       `yaml:"type"`
	Version      string       `yaml:"version"`
	AppVersion   string       `yaml:"appVersion"`
	Keywords     []string     `yaml:"keywords,omitempty"`
	Maintainers  []Maintainer `yaml:"maintainers,omitempty"`
	Dependencies []Dependency `yaml:"dependencies,omitempty"`
}

// Maintainer represents chart maintainer information
type Maintainer struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email,omitempty"`
	URL   string `yaml:"url,omitempty"`
}

func runChartGenerate(cmd *cobra.Command, args []string) {
	// Auto-detect project information
	projectInfo, err := detectProjectInfo()
	if err != nil {
		fmt.Printf("Error detecting project info: %v\n", err)
		os.Exit(1)
	}

	// Override with user-specified values
	if chartName != "" {
		projectInfo.Name = chartName
	}

	// Set defaults
	if projectInfo.Name == "" {
		projectInfo.Name = filepath.Base(getCurrentDir())
	}

	if chartDescription == "" {
		chartDescription = fmt.Sprintf("A Helm chart for %s", projectInfo.Name)
	}

	// Generate chart
	templateData := ChartTemplate{
		Project:        projectInfo,
		ChartName:      projectInfo.Name,
		ChartVersion:   chartVersion,
		AppVersion:     appVersion,
		Description:    chartDescription,
		Namespace:      namespace,
		IncludeDeps:    includeDeps,
		IncludeIngress: includeIngress,
		IncludeService: includeService,
		IncludeHPA:     includeHPA,
		IncludePDB:     includePDB,
		Dependencies:   generateDependencies(projectInfo),
		Values:         generateDefaultValues(projectInfo),
	}

	if err := generateHelmChart(templateData); err != nil {
		fmt.Printf("Error generating Helm chart: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated Helm chart: %s\n", chartOutput)
	fmt.Printf("📋 Chart name: %s\n", templateData.ChartName)
	fmt.Printf("📋 Chart version: %s\n", templateData.ChartVersion)
	fmt.Printf("📋 App version: %s\n", templateData.AppVersion)
	fmt.Printf("📋 Language: %s\n", projectInfo.Language)
	if projectInfo.Framework != "" {
		fmt.Printf("📋 Framework: %s\n", projectInfo.Framework)
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Printf("1. Review and customize values in %s/values.yaml\n", chartOutput)
	fmt.Printf("2. Install: helm install %s %s\n", templateData.ChartName, chartOutput)
	fmt.Printf("3. Upgrade: helm upgrade %s %s\n", templateData.ChartName, chartOutput)
}

func detectProjectInfo() (ProjectInfo, error) {
	info := ProjectInfo{
		Name: filepath.Base(getCurrentDir()),
		Port: 8080, // default port
	}

	// Check for language-specific files
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // ignore errors
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
			info.Framework = detectGoFramework()
		case "package.json":
			info.Language = "node"
			info.Framework = detectNodeFramework()
			info.Port = 3000
		case "requirements.txt", "pyproject.toml", "Pipfile":
			info.Language = "python"
			info.Framework = detectPythonFramework()
		case "Gemfile":
			info.Language = "ruby"
			info.Framework = detectRubyFramework()
		case "Cargo.toml":
			info.Language = "rust"
		case "pom.xml", "build.gradle", "build.gradle.kts":
			info.Language = "java"
			info.Framework = detectJavaFramework()
		case "docker-compose.yml", "docker-compose.yaml":
			info.HasDatabase = true
			info.HasRedis = true
		}

		return nil
	})

	return info, err
}

func detectGoFramework() string {
	// Check go.mod for common Go frameworks
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "github.com/gin-gonic/gin") {
		return "gin"
	}
	if strings.Contains(contentStr, "github.com/gorilla/mux") {
		return "gorilla"
	}
	if strings.Contains(contentStr, "github.com/labstack/echo") {
		return "echo"
	}
	if strings.Contains(contentStr, "github.com/gofiber/fiber") {
		return "fiber"
	}
	return ""
}

func detectNodeFramework() string {
	// Check package.json for common Node.js frameworks
	content, err := os.ReadFile("package.json")
	if err != nil {
		return ""
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "\"express\"") {
		return "express"
	}
	if strings.Contains(contentStr, "\"fastify\"") {
		return "fastify"
	}
	if strings.Contains(contentStr, "\"next\"") {
		return "nextjs"
	}
	if strings.Contains(contentStr, "\"nuxt\"") {
		return "nuxtjs"
	}
	return ""
}

func detectPythonFramework() string {
	// Check requirements.txt for common Python frameworks
	files := []string{"requirements.txt", "pyproject.toml", "Pipfile"}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "django") {
			return "django"
		}
		if strings.Contains(contentStr, "flask") {
			return "flask"
		}
		if strings.Contains(contentStr, "fastapi") {
			return "fastapi"
		}
	}
	return ""
}

func detectRubyFramework() string {
	// Check Gemfile for common Ruby frameworks
	content, err := os.ReadFile("Gemfile")
	if err != nil {
		return ""
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "rails") {
		return "rails"
	}
	if strings.Contains(contentStr, "sinatra") {
		return "sinatra"
	}
	return ""
}

func detectJavaFramework() string {
	// Check for Spring Boot in pom.xml or build.gradle
	files := []string{"pom.xml", "build.gradle", "build.gradle.kts"}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "spring-boot") {
			return "spring-boot"
		}
		if strings.Contains(contentStr, "quarkus") {
			return "quarkus"
		}
		if strings.Contains(contentStr, "micronaut") {
			return "micronaut"
		}
	}
	return ""
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return dir
}

func generateDependencies(info ProjectInfo) []Dependency {
	var deps []Dependency

	if includeDeps {
		if info.HasDatabase {
			deps = append(deps, Dependency{
				Name:       "postgresql",
				Version:    "12.1.5",
				Repository: "https://charts.bitnami.com/bitnami",
				Condition:  "postgresql.enabled",
			})
		}

		if info.HasRedis {
			deps = append(deps, Dependency{
				Name:       "redis",
				Version:    "17.3.7",
				Repository: "https://charts.bitnami.com/bitnami",
				Condition:  "redis.enabled",
			})
		}
	}

	return deps
}

func generateDefaultValues(info ProjectInfo) map[string]interface{} {
	values := map[string]interface{}{
		"replicaCount": 1,
		"image": map[string]interface{}{
			"repository": info.Name,
			"pullPolicy": "IfNotPresent",
			"tag":        "",
		},
		"imagePullSecrets": []string{},
		"nameOverride":     "",
		"fullnameOverride": "",
		"serviceAccount": map[string]interface{}{
			"create":      true,
			"annotations": map[string]interface{}{},
			"name":        "",
		},
		"podAnnotations": map[string]interface{}{},
		"podSecurityContext": map[string]interface{}{
			"fsGroup": 2000,
		},
		"securityContext": map[string]interface{}{
			"capabilities": map[string]interface{}{
				"drop": []string{"ALL"},
			},
			"readOnlyRootFilesystem":   true,
			"runAsNonRoot":             true,
			"runAsUser":                1000,
			"allowPrivilegeEscalation": false,
		},
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "500m",
				"memory": "512Mi",
			},
			"requests": map[string]interface{}{
				"cpu":    "250m",
				"memory": "256Mi",
			},
		},
		"nodeSelector":   map[string]interface{}{},
		"tolerations":    []interface{}{},
		"affinity":       map[string]interface{}{},
		"env":            []interface{}{},
		"envFrom":        []interface{}{},
		"volumes":        []interface{}{},
		"volumeMounts":   []interface{}{},
		"livenessProbe":  generateProbe(info, "liveness"),
		"readinessProbe": generateProbe(info, "readiness"),
		"startupProbe":   generateProbe(info, "startup"),
	}

	if includeService {
		values["service"] = map[string]interface{}{
			"type": "ClusterIP",
			"port": info.Port,
		}
	}

	if includeIngress {
		values["ingress"] = map[string]interface{}{
			"enabled":   false,
			"className": "",
			"annotations": map[string]interface{}{
				"kubernetes.io/ingress.class": "nginx",
			},
			"hosts": []map[string]interface{}{
				{
					"host": fmt.Sprintf("%s.local", info.Name),
					"paths": []map[string]interface{}{
						{
							"path":     "/",
							"pathType": "Prefix",
						},
					},
				},
			},
			"tls": []interface{}{},
		}
	}

	if includeHPA {
		values["autoscaling"] = map[string]interface{}{
			"enabled":                        false,
			"minReplicas":                    1,
			"maxReplicas":                    100,
			"targetCPUUtilizationPercentage": 80,
		}
	}

	if includePDB {
		values["podDisruptionBudget"] = map[string]interface{}{
			"enabled":      false,
			"minAvailable": 1,
		}
	}

	// Add dependency-specific values
	if includeDeps {
		if info.HasDatabase {
			values["postgresql"] = map[string]interface{}{
				"enabled": false,
				"auth": map[string]interface{}{
					"postgresPassword": "changeme",
					"database":         info.Name,
				},
			}
		}

		if info.HasRedis {
			values["redis"] = map[string]interface{}{
				"enabled": false,
				"auth": map[string]interface{}{
					"enabled":  false,
					"password": "",
				},
			}
		}
	}

	return values
}

func generateProbe(info ProjectInfo, probeType string) map[string]interface{} {
	probe := map[string]interface{}{
		"httpGet": map[string]interface{}{
			"path": "/health",
			"port": "http",
		},
	}

	switch probeType {
	case "liveness":
		probe["initialDelaySeconds"] = 30
		probe["periodSeconds"] = 10
		probe["timeoutSeconds"] = 5
		probe["failureThreshold"] = 3
	case "readiness":
		probe["initialDelaySeconds"] = 5
		probe["periodSeconds"] = 10
		probe["timeoutSeconds"] = 5
		probe["failureThreshold"] = 3
	case "startup":
		probe["initialDelaySeconds"] = 10
		probe["periodSeconds"] = 10
		probe["timeoutSeconds"] = 5
		probe["failureThreshold"] = 30
	}

	return probe
}

func generateHelmChart(data ChartTemplate) error {
	// Create chart directory structure
	dirs := []string{
		chartOutput,
		filepath.Join(chartOutput, "templates"),
		filepath.Join(chartOutput, "templates", "tests"),
		filepath.Join(chartOutput, "charts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate Chart.yaml
	if err := generateChartMetadata(data); err != nil {
		return fmt.Errorf("failed to generate Chart.yaml: %w", err)
	}

	// Generate values.yaml
	if err := generateValuesFile(data); err != nil {
		return fmt.Errorf("failed to generate values.yaml: %w", err)
	}

	// Generate templates
	if err := generateTemplates(data); err != nil {
		return fmt.Errorf("failed to generate templates: %w", err)
	}

	// Generate .helmignore
	if err := generateHelmignore(); err != nil {
		return fmt.Errorf("failed to generate .helmignore: %w", err)
	}

	// Generate NOTES.txt
	if err := generateNotes(data); err != nil {
		return fmt.Errorf("failed to generate NOTES.txt: %w", err)
	}

	return nil
}

func generateChartMetadata(data ChartTemplate) error {
	metadata := ChartMetadata{
		APIVersion:  "v2",
		Name:        data.ChartName,
		Description: data.Description,
		Type:        "application",
		Version:     data.ChartVersion,
		AppVersion:  data.AppVersion,
		Keywords:    []string{data.Project.Language, "kubernetes", "helm"},
		Maintainers: []Maintainer{
			{
				Name: "Generated by GZH Manager",
			},
		},
		Dependencies: data.Dependencies,
	}

	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(chartOutput, "Chart.yaml"), yamlData, 0o644)
}

func generateValuesFile(data ChartTemplate) error {
	// Load custom values if specified
	if valuesFile != "" {
		customValues, err := loadCustomValues()
		if err != nil {
			return fmt.Errorf("failed to load custom values: %w", err)
		}
		// Merge custom values with defaults
		for k, v := range customValues {
			data.Values[k] = v
		}
	}

	yamlData, err := yaml.Marshal(data.Values)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(chartOutput, "values.yaml"), yamlData, 0o644)
}

func loadCustomValues() (map[string]interface{}, error) {
	data, err := os.ReadFile(valuesFile)
	if err != nil {
		return nil, err
	}

	var values map[string]interface{}
	err = yaml.Unmarshal(data, &values)
	return values, err
}

func generateTemplates(data ChartTemplate) error {
	templates := map[string]string{
		"deployment.yaml":     getDeploymentTemplate(),
		"service.yaml":        getServiceTemplate(),
		"serviceaccount.yaml": getServiceAccountTemplate(),
		"configmap.yaml":      getConfigMapTemplate(),
		"secret.yaml":         getSecretTemplate(),
		"_helpers.tpl":        getHelpersTemplate(),
	}

	if data.IncludeIngress {
		templates["ingress.yaml"] = getIngressTemplate()
	}

	if data.IncludeHPA {
		templates["hpa.yaml"] = getHPATemplate()
	}

	if data.IncludePDB {
		templates["pdb.yaml"] = getPDBTemplate()
	}

	// Generate test template
	templates["tests/test-connection.yaml"] = getTestTemplate()

	for filename, templateStr := range templates {
		if err := generateTemplateFile(filepath.Join(chartOutput, "templates", filename), templateStr, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	return nil
}

func generateTemplateFile(filepath, templateStr string, data ChartTemplate) error {
	// Create directory if it doesn't exist
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func generateHelmignore() error {
	content := `# Patterns to ignore when building packages.
# This supports shell glob matching, relative path matching, and
# negation (prefixed with !). Only one pattern per line.
.DS_Store
# Common VCS dirs
.git/
.gitignore
.bzr/
.bzrignore
.hg/
.hgignore
.svn/
# Common backup files
*.swp
*.bak
*.tmp
*.orig
*~
# Various IDEs
.project
.idea/
*.tmproj
.vscode/
`

	return os.WriteFile(filepath.Join(chartOutput, ".helmignore"), []byte(content), 0o644)
}

func generateNotes(data ChartTemplate) error {
	content := fmt.Sprintf(`1. Get the application URL by running these commands:
{{- if .Values.ingress.enabled }}
{{- range $host := .Values.ingress.hosts }}
  {{- range .paths }}
  http{{ if $.Values.ingress.tls }}s{{ end }}://{{ $host.host }}{{ .path }}
  {{- end }}
{{- end }}
{{- else if contains "NodePort" .Values.service.type }}
  export NODE_PORT=$(kubectl get --namespace {{ .Release.Namespace }} -o jsonpath="{.spec.ports[0].nodePort}" services {{ include "%s.fullname" . }})
  export NODE_IP=$(kubectl get nodes --namespace {{ .Release.Namespace }} -o jsonpath="{.items[0].status.addresses[0].address}")
  echo http://$NODE_IP:$NODE_PORT
{{- else if contains "LoadBalancer" .Values.service.type }}
     NOTE: It may take a few minutes for the LoadBalancer IP to be available.
           You can watch the status of by running 'kubectl get --namespace {{ .Release.Namespace }} svc -w {{ include "%s.fullname" . }}'
  export SERVICE_IP=$(kubectl get svc --namespace {{ .Release.Namespace }} {{ include "%s.fullname" . }} --template "{{"{{ range (index .status.loadBalancer.ingress 0) }}{{.}}{{ end }}"}}")
  echo http://$SERVICE_IP:{{ .Values.service.port }}
{{- else if contains "ClusterIP" .Values.service.type }}
  export POD_NAME=$(kubectl get pods --namespace {{ .Release.Namespace }} -l "app.kubernetes.io/name={{ include "%s.name" . }},app.kubernetes.io/instance={{ .Release.Name }}" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace {{ .Release.Namespace }} $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl --namespace {{ .Release.Namespace }} port-forward $POD_NAME 8080:$CONTAINER_PORT
{{- end }}
`, data.ChartName, data.ChartName, data.ChartName, data.ChartName)

	return os.WriteFile(filepath.Join(chartOutput, "templates", "NOTES.txt"), []byte(content), 0o644)
}
