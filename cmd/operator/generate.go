package operator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Kubernetes operator components",
	Long: `Generate Kubernetes operator components including CRDs, controllers, and scaffolding.

Supports generation of:
- Custom Resource Definitions (CRDs)
- Controller implementation scaffolding
- Operator project structure
- RBAC configurations
- Deployment manifests
- Sample custom resources

Examples:
  gz operator generate --name myapp --group mycompany.io --version v1
  gz operator generate --name database --group db.example.com --version v1beta1
  gz operator generate --output ./operators/myapp --with-samples`,
	Run: runGenerate,
}

var (
	operatorName    string
	apiGroup        string
	apiVersion      string
	outputDir       string
	withSamples     bool
	withRBAC        bool
	withWebhooks    bool
	namespace       string
	operatorImage   string
	controllerImage string
)

func init() {
	GenerateCmd.Flags().StringVarP(&operatorName, "name", "n", "", "Operator name (required)")
	GenerateCmd.Flags().StringVarP(&apiGroup, "group", "g", "", "API group (e.g., mycompany.io)")
	GenerateCmd.Flags().StringVarP(&apiVersion, "version", "v", "v1", "API version")
	GenerateCmd.Flags().StringVarP(&outputDir, "output", "o", "./operator", "Output directory")
	GenerateCmd.Flags().BoolVar(&withSamples, "with-samples", true, "Generate sample custom resources")
	GenerateCmd.Flags().BoolVar(&withRBAC, "with-rbac", true, "Generate RBAC configurations")
	GenerateCmd.Flags().BoolVar(&withWebhooks, "with-webhooks", false, "Generate admission webhooks")
	GenerateCmd.Flags().StringVar(&namespace, "namespace", "operator-system", "Operator namespace")
	GenerateCmd.Flags().StringVar(&operatorImage, "operator-image", "", "Operator container image")
	GenerateCmd.Flags().StringVar(&controllerImage, "controller-image", "", "Controller container image")

	GenerateCmd.MarkFlagRequired("name")
	GenerateCmd.MarkFlagRequired("group")
}

// OperatorSpec holds operator generation specifications
type OperatorSpec struct {
	Name            string
	NameCamelCase   string
	NameLowerCase   string
	APIGroup        string
	APIVersion      string
	Namespace       string
	OutputDir       string
	WithSamples     bool
	WithRBAC        bool
	WithWebhooks    bool
	OperatorImage   string
	ControllerImage string
	Kind            string
	KindPlural      string
}

// CRDSpec represents a Custom Resource Definition
type CRDSpec struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   CRDMetadata       `yaml:"metadata"`
	Spec       CRDSpecDefinition `yaml:"spec"`
}

type CRDMetadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type CRDSpecDefinition struct {
	Group    string       `yaml:"group"`
	Versions []CRDVersion `yaml:"versions"`
	Scope    string       `yaml:"scope"`
	Names    CRDNames     `yaml:"names"`
}

type CRDVersion struct {
	Name    string                 `yaml:"name"`
	Served  bool                   `yaml:"served"`
	Storage bool                   `yaml:"storage"`
	Schema  map[string]interface{} `yaml:"schema"`
}

type CRDNames struct {
	Plural     string   `yaml:"plural"`
	Singular   string   `yaml:"singular"`
	Kind       string   `yaml:"kind"`
	ShortNames []string `yaml:"shortNames,omitempty"`
}

func runGenerate(cmd *cobra.Command, args []string) {
	if operatorName == "" {
		fmt.Println("Error: operator name is required")
		os.Exit(1)
	}

	if apiGroup == "" {
		fmt.Println("Error: API group is required")
		os.Exit(1)
	}

	// Create operator specification
	spec := OperatorSpec{
		Name:            operatorName,
		NameCamelCase:   toCamelCase(operatorName),
		NameLowerCase:   strings.ToLower(operatorName),
		APIGroup:        apiGroup,
		APIVersion:      apiVersion,
		Namespace:       namespace,
		OutputDir:       outputDir,
		WithSamples:     withSamples,
		WithRBAC:        withRBAC,
		WithWebhooks:    withWebhooks,
		OperatorImage:   operatorImage,
		ControllerImage: controllerImage,
		Kind:            toCamelCase(operatorName),
		KindPlural:      operatorName + "s",
	}

	// Set default images if not provided
	if spec.OperatorImage == "" {
		spec.OperatorImage = fmt.Sprintf("%s-operator:latest", spec.NameLowerCase)
	}
	if spec.ControllerImage == "" {
		spec.ControllerImage = fmt.Sprintf("%s-controller:latest", spec.NameLowerCase)
	}

	// Generate operator components
	if err := generateOperator(spec); err != nil {
		fmt.Printf("Error generating operator: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated Kubernetes operator: %s\n", outputDir)
	fmt.Printf("📋 API Group: %s\n", spec.APIGroup)
	fmt.Printf("📋 API Version: %s\n", spec.APIVersion)
	fmt.Printf("📋 Kind: %s\n", spec.Kind)
	fmt.Printf("📋 Namespace: %s\n", spec.Namespace)

	fmt.Println("\n📝 Next steps:")
	fmt.Printf("1. Review generated files in %s\n", outputDir)
	fmt.Printf("2. Implement controller logic in controller/controller.go\n")
	fmt.Printf("3. Build and deploy: kubectl apply -f %s/deploy/\n", outputDir)
}

func generateOperator(spec OperatorSpec) error {
	// Create directory structure
	dirs := []string{
		spec.OutputDir,
		filepath.Join(spec.OutputDir, "api", spec.APIVersion),
		filepath.Join(spec.OutputDir, "controller"),
		filepath.Join(spec.OutputDir, "deploy"),
		filepath.Join(spec.OutputDir, "samples"),
		filepath.Join(spec.OutputDir, "config"),
		filepath.Join(spec.OutputDir, "hack"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate CRD
	if err := generateCRD(spec); err != nil {
		return fmt.Errorf("failed to generate CRD: %w", err)
	}

	// Generate controller
	if err := generateController(spec); err != nil {
		return fmt.Errorf("failed to generate controller: %w", err)
	}

	// Generate deployment manifests
	if err := generateDeployment(spec); err != nil {
		return fmt.Errorf("failed to generate deployment: %w", err)
	}

	// Generate RBAC if requested
	if spec.WithRBAC {
		if err := generateRBAC(spec); err != nil {
			return fmt.Errorf("failed to generate RBAC: %w", err)
		}
	}

	// Generate samples if requested
	if spec.WithSamples {
		if err := generateSamples(spec); err != nil {
			return fmt.Errorf("failed to generate samples: %w", err)
		}
	}

	// Generate webhooks if requested
	if spec.WithWebhooks {
		if err := generateWebhooks(spec); err != nil {
			return fmt.Errorf("failed to generate webhooks: %w", err)
		}
	}

	// Generate supporting files
	if err := generateSupportingFiles(spec); err != nil {
		return fmt.Errorf("failed to generate supporting files: %w", err)
	}

	return nil
}

func generateCRD(spec OperatorSpec) error {
	crd := CRDSpec{
		APIVersion: "apiextensions.k8s.io/v1",
		Kind:       "CustomResourceDefinition",
		Metadata: CRDMetadata{
			Name: fmt.Sprintf("%s.%s", spec.KindPlural, spec.APIGroup),
			Labels: map[string]string{
				"app.kubernetes.io/name":       spec.NameLowerCase,
				"app.kubernetes.io/component":  "crd",
				"app.kubernetes.io/created-by": "gzh-manager",
			},
		},
		Spec: CRDSpecDefinition{
			Group: spec.APIGroup,
			Scope: "Namespaced",
			Names: CRDNames{
				Plural:     strings.ToLower(spec.KindPlural),
				Singular:   strings.ToLower(spec.Kind),
				Kind:       spec.Kind,
				ShortNames: []string{strings.ToLower(spec.Name)},
			},
			Versions: []CRDVersion{
				{
					Name:    spec.APIVersion,
					Served:  true,
					Storage: true,
					Schema: map[string]interface{}{
						"openAPIV3Schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"spec": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"replicas": map[string]interface{}{
											"type":    "integer",
											"minimum": 1,
											"default": 1,
										},
										"image": map[string]interface{}{
											"type": "string",
										},
										"resources": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"requests": map[string]interface{}{
													"type": "object",
													"properties": map[string]interface{}{
														"cpu":    map[string]interface{}{"type": "string"},
														"memory": map[string]interface{}{"type": "string"},
													},
												},
												"limits": map[string]interface{}{
													"type": "object",
													"properties": map[string]interface{}{
														"cpu":    map[string]interface{}{"type": "string"},
														"memory": map[string]interface{}{"type": "string"},
													},
												},
											},
										},
									},
								},
								"status": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"phase": map[string]interface{}{
											"type": "string",
											"enum": []string{"Pending", "Running", "Succeeded", "Failed"},
										},
										"replicas": map[string]interface{}{
											"type": "integer",
										},
										"readyReplicas": map[string]interface{}{
											"type": "integer",
										},
										"conditions": map[string]interface{}{
											"type": "array",
											"items": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"type": map[string]interface{}{
														"type": "string",
													},
													"status": map[string]interface{}{
														"type": "string",
													},
													"lastTransitionTime": map[string]interface{}{
														"type":   "string",
														"format": "date-time",
													},
													"reason": map[string]interface{}{
														"type": "string",
													},
													"message": map[string]interface{}{
														"type": "string",
													},
												},
												"required": []string{"type", "status"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(crd)
	if err != nil {
		return err
	}

	crdPath := filepath.Join(spec.OutputDir, "deploy", "crd.yaml")
	return os.WriteFile(crdPath, yamlData, 0o644)
}

func generateController(spec OperatorSpec) error {
	controllerTemplate := `package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	{{.NameLowerCase}}v1 "{{.APIGroup}}/{{.APIVersion}}"
)

// {{.Kind}}Reconciler reconciles a {{.Kind}} object
type {{.Kind}}Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups={{.APIGroup}},resources={{.KindPlural}},verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups={{.APIGroup}},resources={{.KindPlural}}/status,verbs=get;update;patch
//+kubebuilder:rbac:groups={{.APIGroup}},resources={{.KindPlural}}/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *{{.Kind}}Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the {{.Kind}} instance
	{{.NameLowerCase}} := &{{.NameLowerCase}}v1.{{.Kind}}{}
	err := r.Get(ctx, req.NamespacedName, {{.NameLowerCase}})
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			logger.Info("{{.Kind}} resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get {{.Kind}}")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: {{.NameLowerCase}}.Name, Namespace: {{.NameLowerCase}}.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentFor{{.Kind}}({{.NameLowerCase}})
		logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment replicas match the desired state
	size := {{.NameLowerCase}}.Spec.Replicas
	if *deployment.Spec.Replicas != size {
		deployment.Spec.Replicas = &size
		err = r.Update(ctx, deployment)
		if err != nil {
			logger.Error(err, "Failed to update Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}
		// Ask to requeue after 1 minute in order to give enough time for the
		// pods be created on the cluster side and the operand be able
		// to do the next update step accurately.
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Update the {{.Kind}} status with the pod names
	// List the pods for this {{.NameLowerCase}}'s deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace({{.NameLowerCase}}.Namespace),
		client.MatchingLabels(labelsFor{{.Kind}}({{.NameLowerCase}}.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		logger.Error(err, "Failed to list pods", "{{.Kind}}.Namespace", {{.NameLowerCase}}.Namespace, "{{.Kind}}.Name", {{.NameLowerCase}}.Name)
		return ctrl.Result{}, err
	}

	// Update status.Replicas if needed
	if {{.NameLowerCase}}.Status.Replicas != int32(len(podList.Items)) {
		{{.NameLowerCase}}.Status.Replicas = int32(len(podList.Items))
		err := r.Status().Update(ctx, {{.NameLowerCase}})
		if err != nil {
			logger.Error(err, "Failed to update {{.Kind}} status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentFor{{.Kind}} returns a {{.NameLowerCase}} Deployment object
func (r *{{.Kind}}Reconciler) deploymentFor{{.Kind}}(m *{{.NameLowerCase}}v1.{{.Kind}}) *appsv1.Deployment {
	ls := labelsFor{{.Kind}}(m.Name)
	replicas := m.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: m.Spec.Image,
						Name:  "{{.NameLowerCase}}",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "http",
						}},
						Resources: m.Spec.Resources,
					}},
				},
			},
		},
	}
	// Set {{.Kind}} instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsFor{{.Kind}} returns the labels for selecting the resources
// belonging to the given {{.NameLowerCase}} CR name.
func labelsFor{{.Kind}}(name string) map[string]string {
	return map[string]string{"app": "{{.NameLowerCase}}", "{{.NameLowerCase}}_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *{{.Kind}}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&{{.NameLowerCase}}v1.{{.Kind}}{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
`

	tmpl, err := template.New("controller").Parse(controllerTemplate)
	if err != nil {
		return err
	}

	controllerPath := filepath.Join(spec.OutputDir, "controller", "controller.go")
	file, err := os.Create(controllerPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, spec)
}

func generateDeployment(spec OperatorSpec) error {
	deploymentTemplate := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.NameLowerCase}}-operator
  namespace: {{.Namespace}}
  labels:
    app.kubernetes.io/name: {{.NameLowerCase}}-operator
    app.kubernetes.io/component: operator
    app.kubernetes.io/created-by: gzh-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: {{.NameLowerCase}}-operator
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{.NameLowerCase}}-operator
    spec:
      serviceAccountName: {{.NameLowerCase}}-operator
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
      containers:
      - name: operator
        image: {{.OperatorImage}}
        imagePullPolicy: IfNotPresent
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
        env:
        - name: OPERATOR_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        ports:
        - containerPort: 9443
          name: webhook-server
          protocol: TCP
        - containerPort: 8080
          name: metrics
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
`

	tmpl, err := template.New("deployment").Parse(deploymentTemplate)
	if err != nil {
		return err
	}

	deployPath := filepath.Join(spec.OutputDir, "deploy", "operator.yaml")
	file, err := os.Create(deployPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, spec)
}

func generateRBAC(spec OperatorSpec) error {
	rbacTemplate := `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.NameLowerCase}}-operator
  namespace: {{.Namespace}}
  labels:
    app.kubernetes.io/name: {{.NameLowerCase}}-operator
    app.kubernetes.io/component: rbac
    app.kubernetes.io/created-by: gzh-manager
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{.NameLowerCase}}-operator
  labels:
    app.kubernetes.io/name: {{.NameLowerCase}}-operator
    app.kubernetes.io/component: rbac
    app.kubernetes.io/created-by: gzh-manager
rules:
- apiGroups:
  - {{.APIGroup}}
  resources:
  - {{.KindPlural}}
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - {{.APIGroup}}
  resources:
  - {{.KindPlural}}/finalizers
  verbs:
  - update
- apiGroups:
  - {{.APIGroup}}
  resources:
  - {{.KindPlural}}/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services
  - pods
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.NameLowerCase}}-operator
  labels:
    app.kubernetes.io/name: {{.NameLowerCase}}-operator
    app.kubernetes.io/component: rbac
    app.kubernetes.io/created-by: gzh-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{.NameLowerCase}}-operator
subjects:
- kind: ServiceAccount
  name: {{.NameLowerCase}}-operator
  namespace: {{.Namespace}}
`

	tmpl, err := template.New("rbac").Parse(rbacTemplate)
	if err != nil {
		return err
	}

	rbacPath := filepath.Join(spec.OutputDir, "deploy", "rbac.yaml")
	file, err := os.Create(rbacPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, spec)
}

func generateSamples(spec OperatorSpec) error {
	sampleTemplate := `apiVersion: {{.APIGroup}}/{{.APIVersion}}
kind: {{.Kind}}
metadata:
  name: {{.NameLowerCase}}-sample
  namespace: default
  labels:
    app.kubernetes.io/name: {{.NameLowerCase}}
    app.kubernetes.io/instance: sample
    app.kubernetes.io/created-by: gzh-manager
spec:
  replicas: 3
  image: nginx:1.21
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
`

	tmpl, err := template.New("sample").Parse(sampleTemplate)
	if err != nil {
		return err
	}

	samplePath := filepath.Join(spec.OutputDir, "samples", fmt.Sprintf("%s-sample.yaml", spec.NameLowerCase))
	file, err := os.Create(samplePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, spec)
}

func generateWebhooks(spec OperatorSpec) error {
	webhookTemplate := `package webhooks

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	{{.NameLowerCase}}v1 "{{.APIGroup}}/{{.APIVersion}}"
)

// log is for logging in this package.
var {{.NameLowerCase}}log = logf.Log.WithName("{{.NameLowerCase}}-resource")

func (r *{{.Kind}}) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-{{.APIGroup}}-{{.APIVersion}}-{{.NameLowerCase}},mutating=true,failurePolicy=fail,sideEffects=None,groups={{.APIGroup}},resources={{.KindPlural}},verbs=create;update,versions={{.APIVersion}},name=m{{.NameLowerCase}}.{{.APIGroup}},admissionReviewVersions=v1

var _ webhook.Defaulter = &{{.Kind}}{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *{{.Kind}}) Default() {
	{{.NameLowerCase}}log.Info("default", "name", r.Name)

	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = 1
	}

	if r.Spec.Image == "" {
		r.Spec.Image = "nginx:latest"
	}
}

//+kubebuilder:webhook:path=/validate-{{.APIGroup}}-{{.APIVersion}}-{{.NameLowerCase}},mutating=false,failurePolicy=fail,sideEffects=None,groups={{.APIGroup}},resources={{.KindPlural}},verbs=create;update,versions={{.APIVersion}},name=v{{.NameLowerCase}}.{{.APIGroup}},admissionReviewVersions=v1

var _ webhook.Validator = &{{.Kind}}{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *{{.Kind}}) ValidateCreate() error {
	{{.NameLowerCase}}log.Info("validate create", "name", r.Name)

	return r.validateSpec()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *{{.Kind}}) ValidateUpdate(old runtime.Object) error {
	{{.NameLowerCase}}log.Info("validate update", "name", r.Name)

	return r.validateSpec()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *{{.Kind}}) ValidateDelete() error {
	{{.NameLowerCase}}log.Info("validate delete", "name", r.Name)

	return nil
}

func (r *{{.Kind}}) validateSpec() error {
	if r.Spec.Replicas < 0 {
		return fmt.Errorf("replicas cannot be negative")
	}

	if r.Spec.Image == "" {
		return fmt.Errorf("image cannot be empty")
	}

	return nil
}
`

	tmpl, err := template.New("webhook").Parse(webhookTemplate)
	if err != nil {
		return err
	}

	webhookPath := filepath.Join(spec.OutputDir, "api", spec.APIVersion, "webhook.go")
	file, err := os.Create(webhookPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, spec)
}

func generateSupportingFiles(spec OperatorSpec) error {
	// Generate API types
	typesTemplate := `package {{.APIVersion}}

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {{.Kind}}Spec defines the desired state of {{.Kind}}
type {{.Kind}}Spec struct {
	// Replicas is the number of desired replicas
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default=1
	Replicas int32 ` + "`" + `json:"replicas,omitempty"` + "`" + `

	// Image is the container image to run
	//+kubebuilder:validation:Required
	Image string ` + "`" + `json:"image"` + "`" + `

	// Resources defines compute resources
	//+optional
	Resources corev1.ResourceRequirements ` + "`" + `json:"resources,omitempty"` + "`" + `
}

// {{.Kind}}Status defines the observed state of {{.Kind}}
type {{.Kind}}Status struct {
	// Phase represents the current phase of the {{.Kind}}
	//+optional
	Phase string ` + "`" + `json:"phase,omitempty"` + "`" + `

	// Replicas is the number of actual replicas
	//+optional
	Replicas int32 ` + "`" + `json:"replicas,omitempty"` + "`" + `

	// ReadyReplicas is the number of ready replicas
	//+optional
	ReadyReplicas int32 ` + "`" + `json:"readyReplicas,omitempty"` + "`" + `

	// Conditions represent the latest available observations
	//+optional
	Conditions []metav1.Condition ` + "`" + `json:"conditions,omitempty"` + "`" + `
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced
//+kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
//+kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// {{.Kind}} is the Schema for the {{.KindPlural}} API
type {{.Kind}} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `

	Spec   {{.Kind}}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
	Status {{.Kind}}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

//+kubebuilder:object:root=true

// {{.Kind}}List contains a list of {{.Kind}}
type {{.Kind}}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{.Kind}} ` + "`" + `json:"items"` + "`" + `
}

func init() {
	SchemeBuilder.Register(&{{.Kind}}{}, &{{.Kind}}List{})
}
`

	tmpl, err := template.New("types").Parse(typesTemplate)
	if err != nil {
		return err
	}

	typesPath := filepath.Join(spec.OutputDir, "api", spec.APIVersion, "types.go")
	file, err := os.Create(typesPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, spec); err != nil {
		return err
	}

	// Generate scheme registration
	schemeTemplate := `package {{.APIVersion}}

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "{{.APIGroup}}", Version: "{{.APIVersion}}"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
`

	tmpl2, err := template.New("scheme").Parse(schemeTemplate)
	if err != nil {
		return err
	}

	schemePath := filepath.Join(spec.OutputDir, "api", spec.APIVersion, "groupversion_info.go")
	file2, err := os.Create(schemePath)
	if err != nil {
		return err
	}
	defer file2.Close()

	if err := tmpl2.Execute(file2, spec); err != nil {
		return err
	}

	// Generate Makefile
	makefileContent := `# Build and deployment helpers for {{.Name}} operator

# Image URL to use all building/pushing image targets
IMG ?= {{.OperatorImage}}

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Build manager binary
build:
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: build
	./bin/manager

# Install CRDs into a cluster
install:
	kubectl apply -f deploy/crd.yaml

# Uninstall CRDs from a cluster
uninstall:
	kubectl delete -f deploy/crd.yaml

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install
	kubectl apply -f deploy/

# Undeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	kubectl delete -f deploy/

# Build docker image
docker-build:
	docker build -t ${IMG} .

# Push docker image
docker-push:
	docker push ${IMG}

.PHONY: build run install uninstall deploy undeploy docker-build docker-push
`

	makefilePath := filepath.Join(spec.OutputDir, "Makefile")
	return os.WriteFile(makefilePath, []byte(makefileContent), 0o644)
}

func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}
