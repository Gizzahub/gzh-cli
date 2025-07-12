package operator

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy operator to Kubernetes cluster",
	Long: `Deploy operator components to a Kubernetes cluster.

Deploys the following components:
- Custom Resource Definitions (CRDs)
- RBAC configurations (ServiceAccount, ClusterRole, ClusterRoleBinding)
- Operator deployment
- Supporting resources

Examples:
  gz operator deploy --path ./operator
  gz operator deploy --path ./operator --namespace custom-system
  gz operator deploy --kubeconfig ~/.kube/config --dry-run`,
	Run: runDeploy,
}

var (
	deployPath      string
	deployNamespace string
	kubeconfig      string
	dryRun          bool
	waitForReady    bool
	deployTimeout   time.Duration
)

func init() {
	DeployCmd.Flags().StringVarP(&deployPath, "path", "p", "./operator", "Path to operator directory")
	DeployCmd.Flags().StringVarP(&deployNamespace, "namespace", "n", "operator-system", "Target namespace")
	DeployCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	DeployCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deployed without actually deploying")
	DeployCmd.Flags().BoolVar(&waitForReady, "wait", true, "Wait for operator to be ready")
	DeployCmd.Flags().DurationVar(&deployTimeout, "timeout", 5*time.Minute, "Deployment timeout")
}

func runDeploy(cmd *cobra.Command, args []string) {
	if deployPath == "" {
		fmt.Println("Error: operator path is required")
		os.Exit(1)
	}

	// Check if deploy directory exists
	deployDir := filepath.Join(deployPath, "deploy")
	if _, err := os.Stat(deployDir); os.IsNotExist(err) {
		fmt.Printf("Error: deploy directory not found: %s\n", deployDir)
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying operator from: %s\n", deployPath)
	fmt.Printf("📋 Target namespace: %s\n", deployNamespace)
	if dryRun {
		fmt.Println("📋 Mode: Dry run (no changes will be made)")
	}

	// Setup Kubernetes client
	clientset, dynamicClient, err := setupKubernetesClients()
	if err != nil {
		fmt.Printf("Error setting up Kubernetes clients: %v\n", err)
		os.Exit(1)
	}

	// Create namespace if it doesn't exist
	if err := ensureNamespace(clientset, deployNamespace); err != nil {
		fmt.Printf("Error ensuring namespace: %v\n", err)
		os.Exit(1)
	}

	// Deploy components in order
	deployOrder := []string{
		"crd.yaml",
		"rbac.yaml",
		"operator.yaml",
	}

	for _, filename := range deployOrder {
		manifestPath := filepath.Join(deployDir, filename)
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			fmt.Printf("⚠️ Skipping %s (not found)\n", filename)
			continue
		}

		fmt.Printf("📦 Deploying %s...\n", filename)
		if err := deployManifest(dynamicClient, manifestPath); err != nil {
			fmt.Printf("❌ Failed to deploy %s: %v\n", filename, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Successfully deployed %s\n", filename)
	}

	if waitForReady && !dryRun {
		fmt.Printf("⏳ Waiting for operator to be ready (timeout: %v)...\n", deployTimeout)
		if err := waitForOperatorReady(clientset, deployNamespace, deployTimeout); err != nil {
			fmt.Printf("❌ Operator not ready: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Operator is ready")
	}

	fmt.Println("🎉 Operator deployment completed successfully!")
	fmt.Println("\n📝 Next steps:")
	fmt.Printf("1. Verify operator status: kubectl get pods -n %s\n", deployNamespace)
	fmt.Printf("2. Check operator logs: kubectl logs -n %s -l app.kubernetes.io/component=operator\n", deployNamespace)
	fmt.Printf("3. Create custom resources in the samples/ directory\n")
}

func setupKubernetesClients() (*kubernetes.Clientset, dynamic.Interface, error) {
	// Use provided kubeconfig or default
	configPath := kubeconfig
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Build config
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build config: %w", err)
	}

	// Create clients
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return clientset, dynamicClient, nil
}

func ensureNamespace(clientset *kubernetes.Clientset, namespace string) error {
	if dryRun {
		fmt.Printf("📋 Would create namespace: %s\n", namespace)
		return nil
	}

	// Check if namespace exists
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("✅ Namespace %s already exists\n", namespace)
		return nil
	}

	// Create namespace
	nsCore := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/created-by": "gzh-manager",
			},
		},
	}

	_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), nsCore, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	fmt.Printf("✅ Created namespace: %s\n", namespace)
	return nil
}

func deployManifest(dynamicClient dynamic.Interface, manifestPath string) error {
	// Read manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse YAML documents
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 4096)

	for {
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode manifest: %w", err)
		}

		if len(obj.Object) == 0 {
			continue
		}

		if dryRun {
			fmt.Printf("📋 Would deploy: %s/%s (%s)\n",
				obj.GetKind(), obj.GetName(), obj.GetAPIVersion())
			continue
		}

		// Apply the resource
		if err := applyResource(dynamicClient, &obj); err != nil {
			return fmt.Errorf("failed to apply resource %s/%s: %w",
				obj.GetKind(), obj.GetName(), err)
		}

		fmt.Printf("  ✅ Applied: %s/%s\n", obj.GetKind(), obj.GetName())
	}

	return nil
}

func applyResource(dynamicClient dynamic.Interface, obj *unstructured.Unstructured) error {
	// Get GVR for the resource
	gvr, err := getGVRForObject(obj)
	if err != nil {
		return err
	}

	// Set namespace if not set and resource is namespaced
	namespace := obj.GetNamespace()
	if namespace == "" && !isClusterScoped(obj.GetKind()) {
		obj.SetNamespace(deployNamespace)
		namespace = deployNamespace
	}

	// Try to get existing resource
	var resourceClient dynamic.ResourceInterface
	if namespace != "" {
		resourceClient = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resourceClient = dynamicClient.Resource(gvr)
	}

	existing, err := resourceClient.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		// Resource doesn't exist, create it
		_, err = resourceClient.Create(context.TODO(), obj, metav1.CreateOptions{})
		return err
	}

	// Resource exists, update it
	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = resourceClient.Update(context.TODO(), obj, metav1.UpdateOptions{})
	return err
}

func getGVRForObject(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	// Map kind to resource name (simplified)
	kind := obj.GetKind()
	resource := strings.ToLower(kind)

	// Handle special cases
	switch kind {
	case "CustomResourceDefinition":
		resource = "customresourcedefinitions"
	case "ClusterRole":
		resource = "clusterroles"
	case "ClusterRoleBinding":
		resource = "clusterrolebindings"
	case "ServiceAccount":
		resource = "serviceaccounts"
	case "Deployment":
		resource = "deployments"
	default:
		// Add 's' for most resources
		if !strings.HasSuffix(resource, "s") {
			resource += "s"
		}
	}

	return schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: resource,
	}, nil
}

func isClusterScoped(kind string) bool {
	clusterScopedKinds := map[string]bool{
		"CustomResourceDefinition": true,
		"ClusterRole":              true,
		"ClusterRoleBinding":       true,
		"Namespace":                true,
		"Node":                     true,
		"PersistentVolume":         true,
		"StorageClass":             true,
	}
	return clusterScopedKinds[kind]
}

func waitForOperatorReady(clientset *kubernetes.Clientset, namespace string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for operator to be ready")
		default:
			// Check operator deployment
			deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/component=operator",
			})
			if err != nil {
				return fmt.Errorf("failed to list deployments: %w", err)
			}

			if len(deployments.Items) == 0 {
				time.Sleep(2 * time.Second)
				continue
			}

			for _, deployment := range deployments.Items {
				if deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
					deployment.Status.Replicas > 0 {
					return nil
				}
			}

			time.Sleep(2 * time.Second)
		}
	}
}
