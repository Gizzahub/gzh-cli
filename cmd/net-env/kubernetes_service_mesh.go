//nolint:tagliatelle // Service mesh configurations may need to match external API formats
package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	serviceMeshIstio = "istio"
)

// IstioConfig represents Istio-specific configuration.
type IstioConfig struct {
	VirtualServices  map[string]*IstioVirtualService  `yaml:"virtualServices,omitempty" json:"virtualServices,omitempty"`
	DestinationRules map[string]*IstioDestinationRule `yaml:"destinationRules,omitempty" json:"destinationRules,omitempty"`
	ServiceEntries   map[string]*IstioServiceEntry    `yaml:"serviceEntries,omitempty" json:"serviceEntries,omitempty"`
	Gateways         map[string]*IstioGateway         `yaml:"gateways,omitempty" json:"gateways,omitempty"`
	SidecarInjection bool                             `yaml:"sidecarInjection" json:"sidecarInjection"`
	MTLSMode         string                           `yaml:"mtlsMode,omitempty" json:"mtlsMode,omitempty"` // DISABLE, SIMPLE, MUTUAL
	CircuitBreaker   *CircuitBreakerConfig            `yaml:"circuitBreaker,omitempty" json:"circuitBreaker,omitempty"`
	RetryPolicy      *RetryPolicyConfig               `yaml:"retryPolicy,omitempty" json:"retryPolicy,omitempty"`
}

// LinkerdConfig represents Linkerd-specific configuration.
type LinkerdConfig struct {
	ServiceProfiles map[string]*LinkerdServiceProfile `yaml:"serviceProfiles,omitempty" json:"serviceProfiles,omitempty"`
	TrafficSplits   map[string]*LinkerdTrafficSplit   `yaml:"trafficSplits,omitempty" json:"trafficSplits,omitempty"`
	ProxyInjection  bool                              `yaml:"proxyInjection" json:"proxyInjection"`
	ProxyResources  *ProxyResourceConfig              `yaml:"proxyResources,omitempty" json:"proxyResources,omitempty"`
	TimeoutPolicy   *TimeoutPolicyConfig              `yaml:"timeoutPolicy,omitempty" json:"timeoutPolicy,omitempty"`
	RetryBudget     *RetryBudgetConfig                `yaml:"retryBudget,omitempty" json:"retryBudget,omitempty"`
}

// IstioVirtualService represents an Istio VirtualService.
type IstioVirtualService struct {
	Name     string           `yaml:"name" json:"name"`
	Hosts    []string         `yaml:"hosts" json:"hosts"`
	Gateways []string         `yaml:"gateways,omitempty" json:"gateways,omitempty"`
	HTTP     []IstioHTTPRoute `yaml:"http,omitempty" json:"http,omitempty"`
	TCP      []IstioTCPRoute  `yaml:"tcp,omitempty" json:"tcp,omitempty"`
	ExportTo []string         `yaml:"exportTo,omitempty" json:"exportTo,omitempty"`
}

// IstioHTTPRoute represents HTTP routing rules in VirtualService.
type IstioHTTPRoute struct {
	Name          string                      `yaml:"name,omitempty" json:"name,omitempty"`
	Match         []IstioHTTPMatchRequest     `yaml:"match,omitempty" json:"match,omitempty"`
	Route         []IstioHTTPRouteDestination `yaml:"route" json:"route"`
	Redirect      *IstioHTTPRedirect          `yaml:"redirect,omitempty" json:"redirect,omitempty"`
	Rewrite       *IstioHTTPRewrite           `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`
	Timeout       string                      `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries       *IstioHTTPRetry             `yaml:"retries,omitempty" json:"retries,omitempty"`
	Fault         *IstioHTTPFaultInjection    `yaml:"fault,omitempty" json:"fault,omitempty"`
	Mirror        *IstioDestination           `yaml:"mirror,omitempty" json:"mirror,omitempty"`
	MirrorPercent *int32                      `yaml:"mirrorPercent,omitempty" json:"mirrorPercent,omitempty"`
}

// IstioHTTPMatchRequest represents HTTP match conditions.
type IstioHTTPMatchRequest struct {
	URI         *StringMatch            `yaml:"uri,omitempty" json:"uri,omitempty"`
	Headers     map[string]*StringMatch `yaml:"headers,omitempty" json:"headers,omitempty"`
	Method      *StringMatch            `yaml:"method,omitempty" json:"method,omitempty"`
	QueryParams map[string]*StringMatch `yaml:"queryParams,omitempty" json:"queryParams,omitempty"`
}

// StringMatch represents different string matching types.
type StringMatch struct {
	Exact  string `yaml:"exact,omitempty" json:"exact,omitempty"`
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Regex  string `yaml:"regex,omitempty" json:"regex,omitempty"`
}

// IstioHTTPRouteDestination represents HTTP route destination.
type IstioHTTPRouteDestination struct {
	Destination *IstioDestination `yaml:"destination" json:"destination"`
	Weight      int32             `yaml:"weight,omitempty" json:"weight,omitempty"`
	Headers     *HeaderOperations `yaml:"headers,omitempty" json:"headers,omitempty"`
}

// IstioDestination represents a service destination.
type IstioDestination struct {
	Host   string        `yaml:"host" json:"host"`
	Subset string        `yaml:"subset,omitempty" json:"subset,omitempty"`
	Port   *PortSelector `yaml:"port,omitempty" json:"port,omitempty"`
}

// PortSelector represents port selection.
type PortSelector struct {
	Number int32 `yaml:"number,omitempty" json:"number,omitempty"`
}

// HeaderOperations represents header manipulation operations.
type HeaderOperations struct {
	Set    map[string]string `yaml:"set,omitempty" json:"set,omitempty"`
	Add    map[string]string `yaml:"add,omitempty" json:"add,omitempty"`
	Remove []string          `yaml:"remove,omitempty" json:"remove,omitempty"`
}

// IstioHTTPRedirect represents HTTP redirect.
type IstioHTTPRedirect struct {
	URI          string `yaml:"uri,omitempty" json:"uri,omitempty"`
	Authority    string `yaml:"authority,omitempty" json:"authority,omitempty"`
	RedirectCode int32  `yaml:"redirectCode,omitempty" json:"redirectCode,omitempty"`
}

// IstioHTTPRewrite represents HTTP rewrite.
type IstioHTTPRewrite struct {
	URI       string `yaml:"uri,omitempty" json:"uri,omitempty"`
	Authority string `yaml:"authority,omitempty" json:"authority,omitempty"`
}

// IstioHTTPRetry represents retry configuration.
type IstioHTTPRetry struct {
	Attempts      int32  `yaml:"attempts" json:"attempts"`
	PerTryTimeout string `yaml:"perTryTimeout,omitempty" json:"perTryTimeout,omitempty"`
	RetryOn       string `yaml:"retryOn,omitempty" json:"retryOn,omitempty"`
}

// IstioHTTPFaultInjection represents fault injection.
type IstioHTTPFaultInjection struct {
	Delay *FaultDelay `yaml:"delay,omitempty" json:"delay,omitempty"`
	Abort *FaultAbort `yaml:"abort,omitempty" json:"abort,omitempty"`
}

// FaultDelay represents delay injection.
type FaultDelay struct {
	Percentage int32  `yaml:"percentage" json:"percentage"`
	FixedDelay string `yaml:"fixedDelay" json:"fixedDelay"`
}

// FaultAbort represents abort injection.
type FaultAbort struct {
	Percentage int32 `yaml:"percentage" json:"percentage"`
	HTTPStatus int32 `yaml:"httpStatus" json:"httpStatus"`
}

// IstioTCPRoute represents TCP routing rules.
type IstioTCPRoute struct {
	Match []IstioL4MatchAttributes `yaml:"match,omitempty" json:"match,omitempty"`
	Route []IstioRouteDestination  `yaml:"route" json:"route"`
}

// IstioL4MatchAttributes represents L4 match attributes.
type IstioL4MatchAttributes struct {
	DestinationSubnets []string          `yaml:"destinationSubnets,omitempty" json:"destinationSubnets,omitempty"`
	Port               int32             `yaml:"port,omitempty" json:"port,omitempty"`
	SourceLabels       map[string]string `yaml:"sourceLabels,omitempty" json:"sourceLabels,omitempty"`
}

// IstioRouteDestination represents route destination.
type IstioRouteDestination struct {
	Destination *IstioDestination `yaml:"destination" json:"destination"`
	Weight      int32             `yaml:"weight,omitempty" json:"weight,omitempty"`
}

// IstioDestinationRule represents an Istio DestinationRule.
type IstioDestinationRule struct {
	Name          string              `yaml:"name" json:"name"`
	Host          string              `yaml:"host" json:"host"`
	TrafficPolicy *IstioTrafficPolicy `yaml:"traffic_policy,omitempty" json:"traffic_policy,omitempty"`
	Subsets       []IstioSubset       `yaml:"subsets,omitempty" json:"subsets,omitempty"`
	ExportTo      []string            `yaml:"export_to,omitempty" json:"export_to,omitempty"`
}

// IstioTrafficPolicy represents traffic policy configuration.
type IstioTrafficPolicy struct {
	LoadBalancer     *LoadBalancerSettings   `yaml:"load_balancer,omitempty" json:"load_balancer,omitempty"`
	ConnectionPool   *ConnectionPoolSettings `yaml:"connection_pool,omitempty" json:"connection_pool,omitempty"`
	OutlierDetection *OutlierDetection       `yaml:"outlier_detection,omitempty" json:"outlier_detection,omitempty"`
	TLS              *ClientTLSSettings      `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// LoadBalancerSettings represents load balancer configuration.
type LoadBalancerSettings struct {
	Simple string `yaml:"simple,omitempty" json:"simple,omitempty"` // ROUND_ROBIN, LEAST_CONN, RANDOM, PASSTHROUGH
}

// ConnectionPoolSettings represents connection pool configuration.
type ConnectionPoolSettings struct {
	TCP  *TCPSettings  `yaml:"tcp,omitempty" json:"tcp,omitempty"`
	HTTP *HTTPSettings `yaml:"http,omitempty" json:"http,omitempty"`
}

// TCPSettings represents TCP connection pool settings.
type TCPSettings struct {
	MaxConnections int32  `yaml:"max_connections,omitempty" json:"max_connections,omitempty"`
	ConnectTimeout string `yaml:"connect_timeout,omitempty" json:"connect_timeout,omitempty"`
}

// HTTPSettings represents HTTP connection pool settings.
type HTTPSettings struct {
	HTTP1MaxPendingRequests  int32  `yaml:"http1_max_pending_requests,omitempty" json:"http1_max_pending_requests,omitempty"`
	HTTP2MaxRequests         int32  `yaml:"http2_max_requests,omitempty" json:"http2_max_requests,omitempty"`
	MaxRequestsPerConnection int32  `yaml:"max_requests_per_connection,omitempty" json:"max_requests_per_connection,omitempty"`
	IdleTimeout              string `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
	UseClientProtocol        bool   `yaml:"use_client_protocol,omitempty" json:"use_client_protocol,omitempty"`
}

// OutlierDetection represents outlier detection configuration.
type OutlierDetection struct {
	ConsecutiveErrors  int32  `yaml:"consecutive_errors,omitempty" json:"consecutive_errors,omitempty"`
	Interval           string `yaml:"interval,omitempty" json:"interval,omitempty"`
	BaseEjectionTime   string `yaml:"base_ejection_time,omitempty" json:"base_ejection_time,omitempty"`
	MaxEjectionPercent int32  `yaml:"max_ejection_percent,omitempty" json:"max_ejection_percent,omitempty"`
	MinHealthPercent   int32  `yaml:"min_health_percent,omitempty" json:"min_health_percent,omitempty"`
}

// ClientTLSSettings represents client TLS settings.
type ClientTLSSettings struct {
	Mode              string   `yaml:"mode" json:"mode"` // DISABLE, SIMPLE, MUTUAL, ISTIO_MUTUAL
	ClientCertificate string   `yaml:"client_certificate,omitempty" json:"client_certificate,omitempty"`
	PrivateKey        string   `yaml:"private_key,omitempty" json:"private_key,omitempty"`
	CaCertificates    string   `yaml:"ca_certificates,omitempty" json:"ca_certificates,omitempty"`
	SubjectAltNames   []string `yaml:"subject_alt_names,omitempty" json:"subject_alt_names,omitempty"`
}

// IstioSubset represents a subset configuration.
type IstioSubset struct {
	Name          string              `yaml:"name" json:"name"`
	Labels        map[string]string   `yaml:"labels" json:"labels"`
	TrafficPolicy *IstioTrafficPolicy `yaml:"traffic_policy,omitempty" json:"traffic_policy,omitempty"`
}

// IstioServiceEntry represents an Istio ServiceEntry.
type IstioServiceEntry struct {
	Name       string          `yaml:"name" json:"name"`
	Hosts      []string        `yaml:"hosts" json:"hosts"`
	Ports      []ServicePort   `yaml:"ports" json:"ports"`
	Location   string          `yaml:"location" json:"location"`                         // MESH_EXTERNAL or MESH_INTERNAL
	Resolution string          `yaml:"resolution,omitempty" json:"resolution,omitempty"` // NONE, STATIC, DNS
	Endpoints  []WorkloadEntry `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
}

// WorkloadEntry represents a workload endpoint.
type WorkloadEntry struct {
	Address string            `yaml:"address" json:"address"`
	Ports   map[string]int32  `yaml:"ports,omitempty" json:"ports,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Weight  int32             `yaml:"weight,omitempty" json:"weight,omitempty"`
}

// IstioGateway represents an Istio Gateway.
type IstioGateway struct {
	Name     string            `yaml:"name" json:"name"`
	Selector map[string]string `yaml:"selector" json:"selector"`
	Servers  []IstioServer     `yaml:"servers" json:"servers"`
}

// IstioServer represents a server configuration in Gateway.
type IstioServer struct {
	Port  *GatewayPort       `yaml:"port" json:"port"`
	Hosts []string           `yaml:"hosts" json:"hosts"`
	TLS   *ServerTLSSettings `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// GatewayPort represents a gateway port.
type GatewayPort struct {
	Number   int32  `yaml:"number" json:"number"`
	Name     string `yaml:"name" json:"name"`
	Protocol string `yaml:"protocol" json:"protocol"`
}

// ServerTLSSettings represents server TLS settings.
type ServerTLSSettings struct {
	Mode              string   `yaml:"mode" json:"mode"` // PASSTHROUGH, SIMPLE, MUTUAL
	ServerCertificate string   `yaml:"server_certificate,omitempty" json:"server_certificate,omitempty"`
	PrivateKey        string   `yaml:"private_key,omitempty" json:"private_key,omitempty"`
	CaCertificates    string   `yaml:"ca_certificates,omitempty" json:"ca_certificates,omitempty"`
	SubjectAltNames   []string `yaml:"subject_alt_names,omitempty" json:"subject_alt_names,omitempty"`
}

// LinkerdServiceProfile represents a Linkerd ServiceProfile.
type LinkerdServiceProfile struct {
	Name        string             `yaml:"name" json:"name"`
	Routes      []LinkerdRoute     `yaml:"routes" json:"routes"`
	RetryBudget *RetryBudgetConfig `yaml:"retry_budget,omitempty" json:"retry_budget,omitempty"`
	OpaquePorts []int32            `yaml:"opaque_ports,omitempty" json:"opaque_ports,omitempty"`
}

// LinkerdRoute represents a route in ServiceProfile.
type LinkerdRoute struct {
	Name        string            `yaml:"name" json:"name"`
	Condition   *LinkerdCondition `yaml:"condition" json:"condition"`
	Timeout     string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	IsRetryable bool              `yaml:"is_retryable,omitempty" json:"is_retryable,omitempty"`
}

// LinkerdCondition represents a route condition.
type LinkerdCondition struct {
	Method    string `yaml:"method,omitempty" json:"method,omitempty"`
	PathRegex string `yaml:"path_regex,omitempty" json:"path_regex,omitempty"`
}

// LinkerdTrafficSplit represents a Linkerd TrafficSplit.
type LinkerdTrafficSplit struct {
	Name     string           `yaml:"name" json:"name"`
	Service  string           `yaml:"service" json:"service"`
	Backends []LinkerdBackend `yaml:"backends" json:"backends"`
}

// LinkerdBackend represents a traffic split backend.
type LinkerdBackend struct {
	Service string `yaml:"service" json:"service"`
	Weight  int32  `yaml:"weight" json:"weight"`
}

// CircuitBreakerConfig represents circuit breaker configuration.
type CircuitBreakerConfig struct {
	ConsecutiveErrors  int32  `yaml:"consecutiveErrors" json:"consecutiveErrors"`
	Interval           string `yaml:"interval" json:"interval"`
	BaseEjectionTime   string `yaml:"baseEjectionTime" json:"baseEjectionTime"`
	MaxEjectionPercent int32  `yaml:"maxEjectionPercent" json:"maxEjectionPercent"`
}

// RetryPolicyConfig represents retry policy configuration.
type RetryPolicyConfig struct {
	Attempts      int32    `yaml:"attempts" json:"attempts"`
	PerTryTimeout string   `yaml:"perTryTimeout,omitempty" json:"perTryTimeout,omitempty"`
	BackoffBase   string   `yaml:"backoffBase,omitempty" json:"backoffBase,omitempty"`
	BackoffMax    string   `yaml:"backoffMax,omitempty" json:"backoffMax,omitempty"`
	RetryOn       []string `yaml:"retryOn,omitempty" json:"retryOn,omitempty"`
}

// ProxyResourceConfig represents proxy resource configuration.
type ProxyResourceConfig struct {
	CPURequest    string `yaml:"cpuRequest" json:"cpuRequest"`
	CPULimit      string `yaml:"cpuLimit" json:"cpuLimit"`
	MemoryRequest string `yaml:"memoryRequest" json:"memoryRequest"`
	MemoryLimit   string `yaml:"memoryLimit" json:"memoryLimit"`
}

// TimeoutPolicyConfig represents timeout policy configuration.
type TimeoutPolicyConfig struct {
	RequestTimeout    string `yaml:"requestTimeout" json:"requestTimeout"`
	IdleTimeout       string `yaml:"idleTimeout" json:"idleTimeout"`
	StreamIdleTimeout string `yaml:"streamIdleTimeout" json:"streamIdleTimeout"`
}

// RetryBudgetConfig represents retry budget configuration.
type RetryBudgetConfig struct {
	RetryRatio          float32 `yaml:"retryRatio" json:"retryRatio"`
	MinRetriesPerSecond int32   `yaml:"minRetriesPerSecond" json:"minRetriesPerSecond"`
	TTL                 string  `yaml:"ttl" json:"ttl"`
}

// DetectServiceMesh detects which service mesh is installed in the cluster.
func (km *KubernetesNetworkManager) DetectServiceMesh(ctx context.Context) (string, error) {
	// Check for Istio
	istioCmd := "kubectl get namespace istio-system --no-headers 2>/dev/null"

	result, err := km.executor.ExecuteWithTimeout(ctx, istioCmd, 5*time.Second)
	if err == nil && result.ExitCode == 0 {
		// Check if Istio control plane is running
		pilotCmd := "kubectl get deployment -n istio-system istiod --no-headers 2>/dev/null"

		pilotResult, err := km.executor.ExecuteWithTimeout(ctx, pilotCmd, 5*time.Second)
		if err == nil && pilotResult.ExitCode == 0 {
			km.logger.Info("Detected Istio service mesh")
			return serviceMeshIstio, nil
		}
	}

	// Check for Linkerd
	linkerdCmd := "kubectl get namespace linkerd --no-headers 2>/dev/null"

	result, err = km.executor.ExecuteWithTimeout(ctx, linkerdCmd, 5*time.Second)
	if err == nil && result.ExitCode == 0 {
		// Check if Linkerd control plane is running
		controllerCmd := "kubectl get deployment -n linkerd linkerd-controller --no-headers 2>/dev/null"

		controllerResult, err := km.executor.ExecuteWithTimeout(ctx, controllerCmd, 5*time.Second)
		if err == nil && controllerResult.ExitCode == 0 {
			km.logger.Info("Detected Linkerd service mesh")
			return "linkerd", nil
		}
	}

	km.logger.Info("No service mesh detected")

	return "", nil
}

// ApplyServiceMeshConfig applies service mesh configuration.
func (km *KubernetesNetworkManager) ApplyServiceMeshConfig(profile *KubernetesNetworkProfile) error {
	if profile.ServiceMesh == nil || !profile.ServiceMesh.Enabled {
		km.logger.Info("Service mesh not enabled for profile", zap.String("profile", profile.Name))
		return nil
	}

	switch profile.ServiceMesh.Type {
	case "istio":
		return km.applyIstioConfig(profile)
	case "linkerd":
		return km.applyLinkerdConfig(profile)
	default:
		return fmt.Errorf("unsupported service mesh type: %s", profile.ServiceMesh.Type)
	}
}

// applyIstioConfig applies Istio-specific configuration.
func (km *KubernetesNetworkManager) applyIstioConfig(profile *KubernetesNetworkProfile) error {
	km.logger.Info("Applying Istio configuration",
		zap.String("profile", profile.Name),
		zap.String("namespace", profile.Namespace))

	// Enable automatic sidecar injection for namespace
	if err := km.enableIstioSidecarInjection(profile.Namespace); err != nil {
		return fmt.Errorf("failed to enable sidecar injection: %w", err)
	}

	// Apply Istio-specific resources from traffic policy
	if profile.ServiceMesh.TrafficPolicy != nil {
		if istioConfig, ok := profile.ServiceMesh.TrafficPolicy["istio"].(*IstioConfig); ok {
			// Apply VirtualServices
			for name, vs := range istioConfig.VirtualServices {
				if err := km.applyIstioVirtualService(profile.Namespace, name, vs); err != nil {
					return fmt.Errorf("failed to apply VirtualService %s: %w", name, err)
				}
			}

			// Apply DestinationRules
			for name, dr := range istioConfig.DestinationRules {
				if err := km.applyIstioDestinationRule(profile.Namespace, name, dr); err != nil {
					return fmt.Errorf("failed to apply DestinationRule %s: %w", name, err)
				}
			}

			// Apply ServiceEntries
			for name, se := range istioConfig.ServiceEntries {
				if err := km.applyIstioServiceEntry(profile.Namespace, name, se); err != nil {
					return fmt.Errorf("failed to apply ServiceEntry %s: %w", name, err)
				}
			}

			// Apply Gateways
			for name, gw := range istioConfig.Gateways {
				if err := km.applyIstioGateway(profile.Namespace, name, gw); err != nil {
					return fmt.Errorf("failed to apply Gateway %s: %w", name, err)
				}
			}
		}
	}

	km.logger.Info("Successfully applied Istio configuration", zap.String("profile", profile.Name))

	return nil
}

// applyLinkerdConfig applies Linkerd-specific configuration.
func (km *KubernetesNetworkManager) applyLinkerdConfig(profile *KubernetesNetworkProfile) error {
	km.logger.Info("Applying Linkerd configuration",
		zap.String("profile", profile.Name),
		zap.String("namespace", profile.Namespace))

	// Enable automatic proxy injection for namespace
	if err := km.enableLinkerdProxyInjection(profile.Namespace); err != nil {
		return fmt.Errorf("failed to enable proxy injection: %w", err)
	}

	// Apply Linkerd-specific resources from traffic policy
	if profile.ServiceMesh.TrafficPolicy != nil {
		if linkerdConfig, ok := profile.ServiceMesh.TrafficPolicy["linkerd"].(*LinkerdConfig); ok {
			// Apply ServiceProfiles
			for name, sp := range linkerdConfig.ServiceProfiles {
				if err := km.applyLinkerdServiceProfile(profile.Namespace, name, sp); err != nil {
					return fmt.Errorf("failed to apply ServiceProfile %s: %w", name, err)
				}
			}

			// Apply TrafficSplits
			for name, ts := range linkerdConfig.TrafficSplits {
				if err := km.applyLinkerdTrafficSplit(profile.Namespace, name, ts); err != nil {
					return fmt.Errorf("failed to apply TrafficSplit %s: %w", name, err)
				}
			}
		}
	}

	km.logger.Info("Successfully applied Linkerd configuration", zap.String("profile", profile.Name))

	return nil
}

// enableIstioSidecarInjection enables Istio sidecar injection for a namespace.
func (km *KubernetesNetworkManager) enableIstioSidecarInjection(namespace string) error {
	// Label namespace for automatic sidecar injection
	labelCmd := fmt.Sprintf("kubectl label namespace %s istio-injection=enabled --overwrite", namespace)

	result, err := km.executor.ExecuteWithTimeout(context.Background(), labelCmd, 10*time.Second)
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("failed to enable sidecar injection: %s", result.Error)
	}

	km.logger.Info("Enabled Istio sidecar injection", zap.String("namespace", namespace))

	return nil
}

// enableLinkerdProxyInjection enables Linkerd proxy injection for a namespace.
func (km *KubernetesNetworkManager) enableLinkerdProxyInjection(namespace string) error {
	// Annotate namespace for automatic proxy injection
	annotateCmd := fmt.Sprintf("kubectl annotate namespace %s linkerd.io/inject=enabled --overwrite", namespace)

	result, err := km.executor.ExecuteWithTimeout(context.Background(), annotateCmd, 10*time.Second)
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("failed to enable proxy injection: %s", result.Error)
	}

	km.logger.Info("Enabled Linkerd proxy injection", zap.String("namespace", namespace))

	return nil
}

// GenerateIstioVirtualService generates an Istio VirtualService manifest.
func (km *KubernetesNetworkManager) GenerateIstioVirtualService(namespace, name string, vs *IstioVirtualService) (map[string]interface{}, error) {
	manifest := map[string]interface{}{
		"apiVersion": "networking.istio.io/v1beta1",
		"kind":       "VirtualService",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"hosts": vs.Hosts,
		},
	}

	spec, ok := manifest["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type in manifest")
	}

	if len(vs.Gateways) > 0 {
		spec["gateways"] = vs.Gateways
	}

	if len(vs.HTTP) > 0 {
		httpRoutes := make([]map[string]interface{}, 0, len(vs.HTTP))
		for _, route := range vs.HTTP {
			httpRoute := km.convertHTTPRoute(route)
			httpRoutes = append(httpRoutes, httpRoute)
		}

		spec["http"] = httpRoutes
	}

	return manifest, nil
}

// convertHTTPRoute converts HTTPRoute to map for YAML generation.
func (km *KubernetesNetworkManager) convertHTTPRoute(route IstioHTTPRoute) map[string]interface{} {
	httpRoute := make(map[string]interface{})

	if route.Name != "" {
		httpRoute["name"] = route.Name
	}

	if len(route.Match) > 0 {
		matches := make([]map[string]interface{}, 0, len(route.Match))
		for _, match := range route.Match {
			m := make(map[string]interface{})
			if match.URI != nil {
				m["uri"] = km.convertStringMatch(match.URI)
			}

			if match.Method != nil {
				m["method"] = km.convertStringMatch(match.Method)
			}

			if len(match.Headers) > 0 {
				headers := make(map[string]interface{})
				for k, v := range match.Headers {
					headers[k] = km.convertStringMatch(v)
				}

				m["headers"] = headers
			}

			matches = append(matches, m)
		}

		httpRoute["match"] = matches
	}

	if len(route.Route) > 0 {
		routes := make([]map[string]interface{}, 0, len(route.Route))
		for _, r := range route.Route {
			rm := map[string]interface{}{
				"destination": map[string]interface{}{
					"host": r.Destination.Host,
				},
			}
			if r.Destination.Subset != "" {
				if dest, ok := rm["destination"].(map[string]interface{}); ok {
					dest["subset"] = r.Destination.Subset
				}
			}

			if r.Weight > 0 {
				rm["weight"] = r.Weight
			}

			routes = append(routes, rm)
		}

		httpRoute["route"] = routes
	}

	if route.Timeout != "" {
		httpRoute["timeout"] = route.Timeout
	}

	if route.Retries != nil {
		httpRoute["retries"] = map[string]interface{}{
			"attempts": route.Retries.Attempts,
		}
		if route.Retries.PerTryTimeout != "" {
			if retries, ok := httpRoute["retries"].(map[string]interface{}); ok {
				retries["perTryTimeout"] = route.Retries.PerTryTimeout
			}
		}
	}

	return httpRoute
}

// convertStringMatch converts StringMatch to map.
func (km *KubernetesNetworkManager) convertStringMatch(sm *StringMatch) map[string]interface{} {
	match := make(map[string]interface{})
	switch {
	case sm.Exact != "":
		match["exact"] = sm.Exact
	case sm.Prefix != "":
		match["prefix"] = sm.Prefix
	case sm.Regex != "":
		match["regex"] = sm.Regex
	}

	return match
}

// applyIstioVirtualService applies an Istio VirtualService.
func (km *KubernetesNetworkManager) applyIstioVirtualService(namespace, name string, vs *IstioVirtualService) error {
	manifest, err := km.GenerateIstioVirtualService(namespace, name, vs)
	if err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal VirtualService: %w", err)
	}

	return km.applyResource(yamlData)
}

// GenerateIstioDestinationRule generates an Istio DestinationRule manifest.
func (km *KubernetesNetworkManager) GenerateIstioDestinationRule(namespace, name string, dr *IstioDestinationRule) (map[string]interface{}, error) {
	manifest := map[string]interface{}{
		"apiVersion": "networking.istio.io/v1beta1",
		"kind":       "DestinationRule",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"host": dr.Host,
		},
	}

	spec, ok := manifest["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type in manifest")
	}

	if dr.TrafficPolicy != nil {
		tp := km.buildTrafficPolicy(dr.TrafficPolicy)
		spec["trafficPolicy"] = tp
	}

	if len(dr.Subsets) > 0 {
		spec["subsets"] = km.buildSubsets(dr.Subsets)
	}

	return manifest, nil
}

// buildTrafficPolicy builds the traffic policy configuration.
func (km *KubernetesNetworkManager) buildTrafficPolicy(tp *IstioTrafficPolicy) map[string]interface{} {
	result := make(map[string]interface{})

	if tp.LoadBalancer != nil && tp.LoadBalancer.Simple != "" {
		result["loadBalancer"] = map[string]interface{}{
			"simple": tp.LoadBalancer.Simple,
		}
	}

	if tp.ConnectionPool != nil {
		result["connectionPool"] = km.buildConnectionPool(tp.ConnectionPool)
	}

	if tp.OutlierDetection != nil {
		result["outlierDetection"] = km.buildOutlierDetection(tp.OutlierDetection)
	}

	if tp.TLS != nil {
		result["tls"] = map[string]interface{}{
			"mode": tp.TLS.Mode,
		}
	}

	return result
}

// buildConnectionPool builds the connection pool configuration.
func (km *KubernetesNetworkManager) buildConnectionPool(cp *ConnectionPoolSettings) map[string]interface{} {
	result := make(map[string]interface{})

	if cp.TCP != nil {
		tcp := make(map[string]interface{})
		if cp.TCP.MaxConnections > 0 {
			tcp["maxConnections"] = cp.TCP.MaxConnections
		}
		if cp.TCP.ConnectTimeout != "" {
			tcp["connectTimeout"] = cp.TCP.ConnectTimeout
		}
		result["tcp"] = tcp
	}

	if cp.HTTP != nil {
		http := make(map[string]interface{})
		if cp.HTTP.HTTP1MaxPendingRequests > 0 {
			http["http1MaxPendingRequests"] = cp.HTTP.HTTP1MaxPendingRequests
		}
		if cp.HTTP.HTTP2MaxRequests > 0 {
			http["http2MaxRequests"] = cp.HTTP.HTTP2MaxRequests
		}
		result["http"] = http
	}

	return result
}

// buildOutlierDetection builds the outlier detection configuration.
func (km *KubernetesNetworkManager) buildOutlierDetection(od *OutlierDetection) map[string]interface{} {
	result := make(map[string]interface{})

	if od.ConsecutiveErrors > 0 {
		result["consecutiveErrors"] = od.ConsecutiveErrors
	}
	if od.Interval != "" {
		result["interval"] = od.Interval
	}
	if od.BaseEjectionTime != "" {
		result["baseEjectionTime"] = od.BaseEjectionTime
	}

	return result
}

// buildSubsets builds the subsets configuration.
func (km *KubernetesNetworkManager) buildSubsets(subsets []IstioSubset) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(subsets))
	for _, subset := range subsets {
		s := map[string]interface{}{
			"name":   subset.Name,
			"labels": subset.Labels,
		}
		result = append(result, s)
	}
	return result
}

// applyIstioDestinationRule applies an Istio DestinationRule.
func (km *KubernetesNetworkManager) applyIstioDestinationRule(namespace, name string, dr *IstioDestinationRule) error {
	manifest, err := km.GenerateIstioDestinationRule(namespace, name, dr)
	if err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal DestinationRule: %w", err)
	}

	return km.applyResource(yamlData)
}

// GenerateIstioServiceEntry generates an Istio ServiceEntry manifest.
func (km *KubernetesNetworkManager) GenerateIstioServiceEntry(namespace, name string, se *IstioServiceEntry) (map[string]interface{}, error) {
	manifest := map[string]interface{}{
		"apiVersion": "networking.istio.io/v1beta1",
		"kind":       "ServiceEntry",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"hosts":    se.Hosts,
			"ports":    km.convertServicePorts(se.Ports),
			"location": se.Location,
		},
	}

	spec, ok := manifest["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type in manifest")
	}

	if se.Resolution != "" {
		spec["resolution"] = se.Resolution
	}

	if len(se.Endpoints) > 0 {
		endpoints := make([]map[string]interface{}, 0, len(se.Endpoints))
		for _, ep := range se.Endpoints {
			endpoint := map[string]interface{}{
				"address": ep.Address,
			}
			if len(ep.Ports) > 0 {
				endpoint["ports"] = ep.Ports
			}

			if len(ep.Labels) > 0 {
				endpoint["labels"] = ep.Labels
			}

			endpoints = append(endpoints, endpoint)
		}

		spec["endpoints"] = endpoints
	}

	return manifest, nil
}

// convertServicePorts converts ServicePort slice to interface slice.
func (km *KubernetesNetworkManager) convertServicePorts(ports []ServicePort) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(ports))
	for _, port := range ports {
		p := map[string]interface{}{
			"number":   port.Port,
			"protocol": port.Protocol,
			"name":     port.Name,
		}
		result = append(result, p)
	}

	return result
}

// applyIstioServiceEntry applies an Istio ServiceEntry.
func (km *KubernetesNetworkManager) applyIstioServiceEntry(namespace, name string, se *IstioServiceEntry) error {
	manifest, err := km.GenerateIstioServiceEntry(namespace, name, se)
	if err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal ServiceEntry: %w", err)
	}

	return km.applyResource(yamlData)
}

// GenerateIstioGateway generates an Istio Gateway manifest.
func (km *KubernetesNetworkManager) GenerateIstioGateway(namespace, name string, gw *IstioGateway) (map[string]interface{}, error) {
	manifest := map[string]interface{}{
		"apiVersion": "networking.istio.io/v1beta1",
		"kind":       "Gateway",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"selector": gw.Selector,
			"servers":  km.convertGatewayServers(gw.Servers),
		},
	}

	return manifest, nil
}

// convertGatewayServers converts IstioServer slice to interface slice.
func (km *KubernetesNetworkManager) convertGatewayServers(servers []IstioServer) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(servers))
	for _, server := range servers {
		s := map[string]interface{}{
			"port": map[string]interface{}{
				"number":   server.Port.Number,
				"name":     server.Port.Name,
				"protocol": server.Port.Protocol,
			},
			"hosts": server.Hosts,
		}
		if server.TLS != nil {
			s["tls"] = map[string]interface{}{
				"mode": server.TLS.Mode,
			}
		}

		result = append(result, s)
	}

	return result
}

// applyIstioGateway applies an Istio Gateway.
func (km *KubernetesNetworkManager) applyIstioGateway(namespace, name string, gw *IstioGateway) error {
	manifest, err := km.GenerateIstioGateway(namespace, name, gw)
	if err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal Gateway: %w", err)
	}

	return km.applyResource(yamlData)
}

// GenerateLinkerdServiceProfile generates a Linkerd ServiceProfile manifest.
func (km *KubernetesNetworkManager) GenerateLinkerdServiceProfile(namespace, name string, sp *LinkerdServiceProfile) (map[string]interface{}, error) {
	manifest := map[string]interface{}{
		"apiVersion": "linkerd.io/v1alpha2",
		"kind":       "ServiceProfile",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{},
	}

	spec, ok := manifest["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type in manifest")
	}

	if len(sp.Routes) > 0 {
		routes := make([]map[string]interface{}, 0, len(sp.Routes))
		for _, route := range sp.Routes {
			r := map[string]interface{}{
				"name": route.Name,
			}
			if route.Condition != nil {
				condition := make(map[string]interface{})
				if route.Condition.Method != "" {
					condition["method"] = route.Condition.Method
				}

				if route.Condition.PathRegex != "" {
					condition["pathRegex"] = route.Condition.PathRegex
				}

				r["condition"] = condition
			}

			if route.Timeout != "" {
				r["timeout"] = route.Timeout
			}

			if route.IsRetryable {
				r["isRetryable"] = true
			}

			routes = append(routes, r)
		}

		spec["routes"] = routes
	}

	if sp.RetryBudget != nil {
		spec["retryBudget"] = map[string]interface{}{
			"retryRatio":          sp.RetryBudget.RetryRatio,
			"minRetriesPerSecond": sp.RetryBudget.MinRetriesPerSecond,
			"ttl":                 sp.RetryBudget.TTL,
		}
	}

	return manifest, nil
}

// applyLinkerdServiceProfile applies a Linkerd ServiceProfile.
func (km *KubernetesNetworkManager) applyLinkerdServiceProfile(namespace, name string, sp *LinkerdServiceProfile) error {
	manifest, err := km.GenerateLinkerdServiceProfile(namespace, name, sp)
	if err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal ServiceProfile: %w", err)
	}

	return km.applyResource(yamlData)
}

// GenerateLinkerdTrafficSplit generates a Linkerd TrafficSplit manifest.
func (km *KubernetesNetworkManager) GenerateLinkerdTrafficSplit(namespace, name string, ts *LinkerdTrafficSplit) (map[string]interface{}, error) {
	manifest := map[string]interface{}{
		"apiVersion": "split.smi-spec.io/v1alpha1",
		"kind":       "TrafficSplit",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"service": ts.Service,
		},
	}

	spec, ok := manifest["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type in manifest")
	}

	if len(ts.Backends) > 0 {
		backends := make([]map[string]interface{}, 0, len(ts.Backends))
		for _, backend := range ts.Backends {
			b := map[string]interface{}{
				"service": backend.Service,
				"weight":  backend.Weight,
			}
			backends = append(backends, b)
		}

		spec["backends"] = backends
	}

	return manifest, nil
}

// applyLinkerdTrafficSplit applies a Linkerd TrafficSplit.
func (km *KubernetesNetworkManager) applyLinkerdTrafficSplit(namespace, name string, ts *LinkerdTrafficSplit) error {
	manifest, err := km.GenerateLinkerdTrafficSplit(namespace, name, ts)
	if err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal TrafficSplit: %w", err)
	}

	return km.applyResource(yamlData)
}

// ValidateServiceMeshConfig validates service mesh configuration.
func (km *KubernetesNetworkManager) ValidateServiceMeshConfig(config *ServiceMeshConfig) error {
	if config == nil {
		return nil
	}

	if config.Type != "" && config.Type != "istio" && config.Type != "linkerd" {
		return fmt.Errorf("unsupported service mesh type: %s (must be 'istio' or 'linkerd')", config.Type)
	}

	// Validate namespace
	if config.Namespace == "" {
		config.Namespace = "default"
	}

	// Type-specific validation
	if config.TrafficPolicy != nil {
		switch config.Type {
		case "istio":
			if istioConfig, ok := config.TrafficPolicy["istio"]; ok {
				// Validate Istio configuration
				if err := km.validateIstioConfig(istioConfig); err != nil {
					return fmt.Errorf("invalid Istio configuration: %w", err)
				}
			}
		case "linkerd":
			if linkerdConfig, ok := config.TrafficPolicy["linkerd"]; ok {
				// Validate Linkerd configuration
				if err := km.validateLinkerdConfig(linkerdConfig); err != nil {
					return fmt.Errorf("invalid Linkerd configuration: %w", err)
				}
			}
		}
	}

	return nil
}

// validateIstioConfig validates Istio-specific configuration.
func (km *KubernetesNetworkManager) validateIstioConfig(config interface{}) error {
	// Convert to IstioConfig if possible
	istioConfig, ok := config.(*IstioConfig)
	if !ok {
		// Try to convert from map
		configMap, ok := config.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid Istio configuration format")
		}
		// Convert map to IstioConfig
		jsonData, err := json.Marshal(configMap)
		if err != nil {
			return err
		}

		istioConfig = &IstioConfig{}
		if err := json.Unmarshal(jsonData, istioConfig); err != nil {
			return err
		}
	}

	// Validate mTLS mode
	if istioConfig.MTLSMode != "" {
		validModes := map[string]bool{
			"DISABLE":      true,
			"SIMPLE":       true,
			"MUTUAL":       true,
			"ISTIO_MUTUAL": true,
		}
		if !validModes[istioConfig.MTLSMode] {
			return fmt.Errorf("invalid mTLS mode: %s", istioConfig.MTLSMode)
		}
	}

	// Validate VirtualServices
	for name, vs := range istioConfig.VirtualServices {
		if len(vs.Hosts) == 0 {
			return fmt.Errorf("VirtualService %s must have at least one host", name)
		}
	}

	// Validate DestinationRules
	for name, dr := range istioConfig.DestinationRules {
		if dr.Host == "" {
			return fmt.Errorf("DestinationRule %s must have a host", name)
		}
	}

	return nil
}

// validateLinkerdConfig validates Linkerd-specific configuration.
func (km *KubernetesNetworkManager) validateLinkerdConfig(config interface{}) error {
	// Convert to LinkerdConfig if possible
	linkerdConfig, ok := config.(*LinkerdConfig)
	if !ok {
		// Try to convert from map
		configMap, ok := config.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid Linkerd configuration format")
		}
		// Convert map to LinkerdConfig
		jsonData, err := json.Marshal(configMap)
		if err != nil {
			return err
		}

		linkerdConfig = &LinkerdConfig{}
		if err := json.Unmarshal(jsonData, linkerdConfig); err != nil {
			return err
		}
	}

	// Validate ServiceProfiles
	for name, sp := range linkerdConfig.ServiceProfiles {
		if len(sp.Routes) == 0 {
			km.logger.Warn("ServiceProfile has no routes defined", zap.String("profile", name))
		}
	}

	// Validate TrafficSplits
	for name, ts := range linkerdConfig.TrafficSplits {
		if ts.Service == "" {
			return fmt.Errorf("TrafficSplit %s must have a service", name)
		}

		if len(ts.Backends) == 0 {
			return fmt.Errorf("TrafficSplit %s must have at least one backend", name)
		}

		// Validate weights sum to 100
		totalWeight := int32(0)
		for _, backend := range ts.Backends {
			totalWeight += backend.Weight
		}

		if totalWeight != 100 {
			return fmt.Errorf("TrafficSplit %s backend weights must sum to 100, got %d", name, totalWeight)
		}
	}

	return nil
}

// GetServiceMeshStatus gets the status of service mesh in a namespace.
func (km *KubernetesNetworkManager) GetServiceMeshStatus(namespace string) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Detect service mesh type
	meshType, err := km.DetectServiceMesh(context.Background())
	if err != nil {
		return nil, err
	}

	status["mesh_type"] = meshType
	status["namespace"] = namespace

	switch meshType {
	case "istio":
		// Check Istio injection status
		labelCmd := fmt.Sprintf("kubectl get namespace %s -o jsonpath='{.metadata.labels.istio-injection}'", namespace)
		result, _ := km.executor.ExecuteWithTimeout(context.Background(), labelCmd, 5*time.Second)
		status["sidecar_injection"] = strings.Trim(result.Output, "'") == "enabled"

		// Count Istio resources
		resources := []string{"virtualservices", "destinationrules", "serviceentries", "gateways"}
		resourceCounts := make(map[string]int)

		for _, resource := range resources {
			countCmd := fmt.Sprintf("kubectl get %s -n %s --no-headers 2>/dev/null | wc -l", resource, namespace)
			result, _ := km.executor.ExecuteWithTimeout(context.Background(), countCmd, 5*time.Second)
			count := 0
			_, _ = fmt.Sscanf(strings.TrimSpace(result.Output), "%d", &count)
			resourceCounts[resource] = count
		}

		status["resources"] = resourceCounts

	case "linkerd":
		// Check Linkerd injection status
		annotationCmd := fmt.Sprintf("kubectl get namespace %s -o jsonpath='{.metadata.annotations.linkerd\\.io/inject}'", namespace)
		result, _ := km.executor.ExecuteWithTimeout(context.Background(), annotationCmd, 5*time.Second)
		status["proxy_injection"] = strings.Trim(result.Output, "'") == "enabled"

		// Count Linkerd resources
		resources := []string{"serviceprofiles.linkerd.io", "trafficsplits.split.smi-spec.io"}
		resourceCounts := make(map[string]int)

		for _, resource := range resources {
			countCmd := fmt.Sprintf("kubectl get %s -n %s --no-headers 2>/dev/null | wc -l", resource, namespace)
			result, _ := km.executor.ExecuteWithTimeout(context.Background(), countCmd, 5*time.Second)
			count := 0
			_, _ = fmt.Sscanf(strings.TrimSpace(result.Output), "%d", &count)
			resourceCounts[resource] = count
		}

		status["resources"] = resourceCounts

	default:
		status["enabled"] = false
	}

	return status, nil
}
