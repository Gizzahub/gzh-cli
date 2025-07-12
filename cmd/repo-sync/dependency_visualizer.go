package reposync

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DependencyVisualizer creates visual representations of dependency graphs
type DependencyVisualizer struct {
	logger *zap.Logger
	config *VisualizerConfig
}

// VisualizerConfig represents configuration for dependency visualization
type VisualizerConfig struct {
	OutputFormat     string            `json:"output_format"`     // dot, svg, html, json
	OutputPath       string            `json:"output_path"`
	IncludeExternal  bool              `json:"include_external"`
	IncludeInternal  bool              `json:"include_internal"`
	MaxNodes         int               `json:"max_nodes"`
	ClusterByType    bool              `json:"cluster_by_type"`
	ShowLabels       bool              `json:"show_labels"`
	NodeStyles       map[string]string `json:"node_styles"`
	EdgeStyles       map[string]string `json:"edge_styles"`
	LayoutAlgorithm  string            `json:"layout_algorithm"` // dot, neato, fdp, sfdp, twopi, circo
	Theme            string            `json:"theme"`            // light, dark, colorful
}

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	ID          string                 `json:"id"`
	Label       string                 `json:"label"`
	Type        string                 `json:"type"`        // module, file, package
	Language    string                 `json:"language"`
	Size        int                    `json:"size"`        // lines of code or file count
	Complexity  float64                `json:"complexity"`
	External    bool                   `json:"external"`
	Properties  map[string]interface{} `json:"properties"`
	Position    *NodePosition          `json:"position,omitempty"`
}

// GraphEdge represents an edge in the dependency graph
type GraphEdge struct {
	From       string             `json:"from"`
	To         string             `json:"to"`
	Type       DependencyType     `json:"type"`
	Strength   DependencyStrength `json:"strength"`
	Weight     float64            `json:"weight"`
	Properties map[string]interface{} `json:"properties"`
}

// NodePosition represents the position of a node in 2D space
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// DependencyGraph represents the complete dependency graph
type DependencyGraph struct {
	Nodes         []*GraphNode         `json:"nodes"`
	Edges         []*GraphEdge         `json:"edges"`
	Metadata      *GraphMetadata       `json:"metadata"`
	Clusters      map[string][]string  `json:"clusters,omitempty"`
	Statistics    *GraphStatistics     `json:"statistics"`
}

// GraphMetadata contains metadata about the graph
type GraphMetadata struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	GeneratedAt time.Time `json:"generated_at"`
	Repository  string    `json:"repository"`
	TotalNodes  int       `json:"total_nodes"`
	TotalEdges  int       `json:"total_edges"`
	Languages   []string  `json:"languages"`
}

// GraphStatistics contains statistical information about the graph
type GraphStatistics struct {
	NodesByType      map[string]int             `json:"nodes_by_type"`
	EdgesByType      map[DependencyType]int     `json:"edges_by_type"`
	EdgesByStrength  map[DependencyStrength]int `json:"edges_by_strength"`
	CentralityScores map[string]float64         `json:"centrality_scores"`
	ClusterSizes     map[string]int             `json:"cluster_sizes"`
	MaxDepth         int                        `json:"max_depth"`
	Density          float64                    `json:"density"`
}

// NewDependencyVisualizer creates a new dependency visualizer
func NewDependencyVisualizer(logger *zap.Logger, config *VisualizerConfig) *DependencyVisualizer {
	return &DependencyVisualizer{
		logger: logger,
		config: config,
	}
}

// CreateGraph creates a dependency graph from analysis results
func (dv *DependencyVisualizer) CreateGraph(result *DependencyResult) (*DependencyGraph, error) {
	dv.logger.Info("Creating dependency graph", 
		zap.String("repository", result.Repository))

	graph := &DependencyGraph{
		Nodes:      make([]*GraphNode, 0),
		Edges:      make([]*GraphEdge, 0),
		Clusters:   make(map[string][]string),
		Metadata: &GraphMetadata{
			Title:       fmt.Sprintf("Dependency Graph - %s", filepath.Base(result.Repository)),
			Description: "Generated dependency visualization",
			GeneratedAt: time.Now(),
			Repository:  result.Repository,
			Languages:   dv.extractLanguages(result),
		},
	}

	// Create nodes from modules
	nodeMap := make(map[string]*GraphNode)
	if dv.config.IncludeInternal {
		for moduleID, module := range result.Modules {
			node := dv.createModuleNode(moduleID, module)
			if node != nil {
				graph.Nodes = append(graph.Nodes, node)
				nodeMap[moduleID] = node
			}
		}
	}

	// Add external dependency nodes if enabled
	if dv.config.IncludeExternal {
		for _, extDep := range result.ExternalDeps {
			nodeID := fmt.Sprintf("ext_%s_%s", extDep.Language, extDep.Name)
			if _, exists := nodeMap[nodeID]; !exists {
				node := dv.createExternalNode(extDep)
				graph.Nodes = append(graph.Nodes, node)
				nodeMap[nodeID] = node
			}
		}
	}

	// Create edges from dependencies
	for _, dep := range result.Dependencies {
		if dv.shouldIncludeDependency(dep) {
			edge := dv.createEdge(dep, nodeMap)
			if edge != nil {
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}

	// Apply node limit if specified
	if dv.config.MaxNodes > 0 && len(graph.Nodes) > dv.config.MaxNodes {
		graph.Nodes = dv.selectTopNodes(graph.Nodes, dv.config.MaxNodes)
		graph.Edges = dv.filterEdgesByNodes(graph.Edges, graph.Nodes)
	}

	// Create clusters if enabled
	if dv.config.ClusterByType {
		graph.Clusters = dv.createClusters(graph.Nodes)
	}

	// Calculate statistics
	graph.Statistics = dv.calculateGraphStatistics(graph)

	// Update metadata
	graph.Metadata.TotalNodes = len(graph.Nodes)
	graph.Metadata.TotalEdges = len(graph.Edges)

	return graph, nil
}

// RenderGraph renders the dependency graph in the specified format
func (dv *DependencyVisualizer) RenderGraph(graph *DependencyGraph) error {
	switch strings.ToLower(dv.config.OutputFormat) {
	case "dot":
		return dv.renderDOT(graph)
	case "html":
		return dv.renderHTML(graph)
	case "json":
		return dv.renderJSON(graph)
	case "svg":
		return dv.renderSVG(graph)
	default:
		return fmt.Errorf("unsupported output format: %s", dv.config.OutputFormat)
	}
}

// createModuleNode creates a graph node from a module
func (dv *DependencyVisualizer) createModuleNode(moduleID string, module *ModuleDependencies) *GraphNode {
	size := len(module.Files)
	if size == 0 {
		size = len(module.Dependencies)
	}

	complexity := float64(len(module.Dependencies)) / 10.0 // Normalize complexity

	return &GraphNode{
		ID:         moduleID,
		Label:      dv.formatNodeLabel(moduleID, module.Language),
		Type:       "module",
		Language:   module.Language,
		Size:       size,
		Complexity: complexity,
		External:   false,
		Properties: map[string]interface{}{
			"version":     module.Version,
			"description": module.Description,
			"file_count":  len(module.Files),
			"exports":     len(module.Exports),
		},
	}
}

// createExternalNode creates a graph node from an external dependency
func (dv *DependencyVisualizer) createExternalNode(extDep ExternalDependency) *GraphNode {
	nodeID := fmt.Sprintf("ext_%s_%s", extDep.Language, extDep.Name)
	
	return &GraphNode{
		ID:         nodeID,
		Label:      dv.formatNodeLabel(extDep.Name, extDep.Language),
		Type:       "external",
		Language:   extDep.Language,
		Size:       extDep.UsageCount,
		Complexity: 0.0,
		External:   true,
		Properties: map[string]interface{}{
			"version":     extDep.Version,
			"description": extDep.Description,
			"license":     extDep.License,
			"repository":  extDep.Repository,
			"usage_count": extDep.UsageCount,
		},
	}
}

// createEdge creates a graph edge from a dependency
func (dv *DependencyVisualizer) createEdge(dep *Dependency, nodeMap map[string]*GraphNode) *GraphEdge {
	fromID := dep.From
	toID := dep.To

	// For external dependencies, use the external node ID format
	if dep.External {
		toID = fmt.Sprintf("ext_%s_%s", dep.Language, dep.To)
	}

	// Check if both nodes exist
	if _, fromExists := nodeMap[fromID]; !fromExists {
		return nil
	}
	if _, toExists := nodeMap[toID]; !toExists {
		return nil
	}

	weight := dv.calculateEdgeWeight(dep)

	return &GraphEdge{
		From:     fromID,
		To:       toID,
		Type:     dep.Type,
		Strength: dep.Strength,
		Weight:   weight,
		Properties: map[string]interface{}{
			"language": dep.Language,
			"version":  dep.Version,
			"location": dep.Location,
		},
	}
}

// shouldIncludeDependency determines if a dependency should be included in the graph
func (dv *DependencyVisualizer) shouldIncludeDependency(dep *Dependency) bool {
	if dep.External && !dv.config.IncludeExternal {
		return false
	}
	if !dep.External && !dv.config.IncludeInternal {
		return false
	}
	return true
}

// calculateEdgeWeight calculates the weight of an edge based on dependency properties
func (dv *DependencyVisualizer) calculateEdgeWeight(dep *Dependency) float64 {
	weight := 1.0

	// Adjust weight based on dependency strength
	switch dep.Strength {
	case DependencyStrengthStrong:
		weight = 3.0
	case DependencyStrengthWeak:
		weight = 1.5
	case DependencyStrengthOptional:
		weight = 0.5
	}

	// Adjust weight based on dependency type
	switch dep.Type {
	case DependencyTypeImport:
		weight *= 1.0
	case DependencyTypeRequire:
		weight *= 1.2
	case DependencyTypeInclude:
		weight *= 0.8
	default:
		weight *= 1.0
	}

	return weight
}

// formatNodeLabel formats the label for a node
func (dv *DependencyVisualizer) formatNodeLabel(name, language string) string {
	if !dv.config.ShowLabels {
		return ""
	}

	// Shorten long module names
	parts := strings.Split(name, "/")
	if len(parts) > 3 {
		return fmt.Sprintf(".../%s", strings.Join(parts[len(parts)-2:], "/"))
	}

	maxLen := 25
	if len(name) > maxLen {
		return name[:maxLen-3] + "..."
	}

	return name
}

// selectTopNodes selects the top N nodes based on importance (complexity, size, centrality)
func (dv *DependencyVisualizer) selectTopNodes(nodes []*GraphNode, maxNodes int) []*GraphNode {
	// Calculate importance score for each node
	for _, node := range nodes {
		importance := float64(node.Size) + node.Complexity*5
		if node.External {
			importance *= 0.5 // External nodes are less important for structure
		}
		node.Properties["importance"] = importance
	}

	// Sort by importance
	sort.Slice(nodes, func(i, j int) bool {
		importanceI := nodes[i].Properties["importance"].(float64)
		importanceJ := nodes[j].Properties["importance"].(float64)
		return importanceI > importanceJ
	})

	return nodes[:maxNodes]
}

// filterEdgesByNodes filters edges to only include those between existing nodes
func (dv *DependencyVisualizer) filterEdgesByNodes(edges []*GraphEdge, nodes []*GraphNode) []*GraphEdge {
	nodeSet := make(map[string]bool)
	for _, node := range nodes {
		nodeSet[node.ID] = true
	}

	filtered := make([]*GraphEdge, 0)
	for _, edge := range edges {
		if nodeSet[edge.From] && nodeSet[edge.To] {
			filtered = append(filtered, edge)
		}
	}

	return filtered
}

// createClusters creates clusters of nodes based on type or language
func (dv *DependencyVisualizer) createClusters(nodes []*GraphNode) map[string][]string {
	clusters := make(map[string][]string)

	for _, node := range nodes {
		var clusterKey string
		if node.External {
			clusterKey = fmt.Sprintf("external_%s", node.Language)
		} else {
			clusterKey = fmt.Sprintf("internal_%s", node.Language)
		}

		clusters[clusterKey] = append(clusters[clusterKey], node.ID)
	}

	return clusters
}

// calculateGraphStatistics calculates various statistics about the graph
func (dv *DependencyVisualizer) calculateGraphStatistics(graph *DependencyGraph) *GraphStatistics {
	stats := &GraphStatistics{
		NodesByType:      make(map[string]int),
		EdgesByType:      make(map[DependencyType]int),
		EdgesByStrength:  make(map[DependencyStrength]int),
		CentralityScores: make(map[string]float64),
		ClusterSizes:     make(map[string]int),
	}

	// Count nodes by type
	for _, node := range graph.Nodes {
		stats.NodesByType[node.Type]++
	}

	// Count edges by type and strength
	for _, edge := range graph.Edges {
		stats.EdgesByType[edge.Type]++
		stats.EdgesByStrength[edge.Strength]++
	}

	// Calculate cluster sizes
	for clusterName, nodeIDs := range graph.Clusters {
		stats.ClusterSizes[clusterName] = len(nodeIDs)
	}

	// Calculate centrality scores (degree centrality)
	nodeDegrees := make(map[string]int)
	for _, edge := range graph.Edges {
		nodeDegrees[edge.From]++
		nodeDegrees[edge.To]++
	}

	maxDegree := 0
	for _, degree := range nodeDegrees {
		if degree > maxDegree {
			maxDegree = degree
		}
	}

	if maxDegree > 0 {
		for nodeID, degree := range nodeDegrees {
			stats.CentralityScores[nodeID] = float64(degree) / float64(maxDegree)
		}
	}

	// Calculate graph density
	nodeCount := len(graph.Nodes)
	edgeCount := len(graph.Edges)
	if nodeCount > 1 {
		maxPossibleEdges := nodeCount * (nodeCount - 1)
		stats.Density = float64(edgeCount) / float64(maxPossibleEdges)
	}

	return stats
}

// extractLanguages extracts unique languages from the dependency result
func (dv *DependencyVisualizer) extractLanguages(result *DependencyResult) []string {
	languageSet := make(map[string]bool)
	for _, dep := range result.Dependencies {
		languageSet[dep.Language] = true
	}

	languages := make([]string, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	sort.Strings(languages)
	return languages
}

// renderJSON renders the graph as JSON
func (dv *DependencyVisualizer) renderJSON(graph *DependencyGraph) error {
	outputPath := dv.getOutputPath("json")
	
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph to JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	dv.logger.Info("Graph exported as JSON", zap.String("path", outputPath))
	return nil
}

// getOutputPath generates the output path for a given format
func (dv *DependencyVisualizer) getOutputPath(format string) string {
	if dv.config.OutputPath != "" {
		return dv.config.OutputPath
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("dependency_graph_%s.%s", timestamp, format)
	return filename
}