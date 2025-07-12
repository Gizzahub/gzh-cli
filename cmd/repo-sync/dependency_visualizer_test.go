package reposync

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewDependencyVisualizer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VisualizerConfig{
		OutputFormat:    "dot",
		IncludeExternal: true,
		IncludeInternal: true,
		MaxNodes:        100,
	}

	visualizer := NewDependencyVisualizer(logger, config)

	assert.NotNil(t, visualizer)
	assert.Equal(t, config, visualizer.config)
	assert.NotNil(t, visualizer.logger)
}

func TestDependencyVisualizer_CreateGraph(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VisualizerConfig{
		OutputFormat:    "json",
		IncludeExternal: true,
		IncludeInternal: true,
		ClusterByType:   true,
		ShowLabels:      true,
	}

	visualizer := NewDependencyVisualizer(logger, config)

	// Create test dependency result
	result := &DependencyResult{
		Repository: "/test/repo",
		Dependencies: []*Dependency{
			{From: "module1", To: "module2", Type: DependencyTypeImport, Language: "go", Strength: DependencyStrengthStrong, External: false},
			{From: "module1", To: "external_lib", Type: DependencyTypeRequire, Language: "go", Strength: DependencyStrengthStrong, External: true},
			{From: "module2", To: "module3", Type: DependencyTypeImport, Language: "javascript", Strength: DependencyStrengthWeak, External: false},
		},
		Modules: map[string]*ModuleDependencies{
			"module1": {
				ModulePath: "module1",
				Language:   "go",
				Files:      []string{"file1.go", "file2.go"},
				Version:    "v1.0.0",
			},
			"module2": {
				ModulePath: "module2",
				Language:   "go",
				Files:      []string{"file3.go"},
				Version:    "v2.0.0",
			},
			"module3": {
				ModulePath: "module3",
				Language:   "javascript",
				Files:      []string{"file4.js"},
			},
		},
		ExternalDeps: []ExternalDependency{
			{
				Name:        "external_lib",
				Version:     "v1.2.3",
				Language:    "go",
				UsageCount:  5,
				Description: "External library",
			},
		},
	}

	graph, err := visualizer.CreateGraph(result)
	require.NoError(t, err)

	// Verify graph structure
	assert.NotNil(t, graph)
	assert.NotNil(t, graph.Metadata)
	assert.NotNil(t, graph.Statistics)

	// Should have 4 nodes (3 modules + 1 external)
	assert.Len(t, graph.Nodes, 4)

	// Should have 3 edges
	assert.Len(t, graph.Edges, 3)

	// Verify metadata
	assert.Equal(t, "/test/repo", graph.Metadata.Repository)
	assert.Equal(t, 4, graph.Metadata.TotalNodes)
	assert.Equal(t, 3, graph.Metadata.TotalEdges)
	assert.Contains(t, graph.Metadata.Languages, "go")
	assert.Contains(t, graph.Metadata.Languages, "javascript")

	// Verify clusters
	assert.NotEmpty(t, graph.Clusters)
	assert.Contains(t, graph.Clusters, "internal_go")
	assert.Contains(t, graph.Clusters, "external_go")
	assert.Contains(t, graph.Clusters, "internal_javascript")

	// Verify statistics
	assert.Equal(t, 3, graph.Statistics.NodesByType["module"])
	assert.Equal(t, 1, graph.Statistics.NodesByType["external"])
}

func TestDependencyVisualizer_CreateGraphWithLimits(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VisualizerConfig{
		OutputFormat:    "json",
		IncludeExternal: true,
		IncludeInternal: true,
		MaxNodes:        2, // Limit to 2 nodes
	}

	visualizer := NewDependencyVisualizer(logger, config)

	// Create test data with more nodes than the limit
	result := &DependencyResult{
		Repository: "/test/repo",
		Dependencies: []*Dependency{
			{From: "module1", To: "module2", External: false},
			{From: "module2", To: "module3", External: false},
			{From: "module3", To: "module1", External: false},
		},
		Modules: map[string]*ModuleDependencies{
			"module1": {ModulePath: "module1", Language: "go", Files: []string{"f1", "f2", "f3"}},
			"module2": {ModulePath: "module2", Language: "go", Files: []string{"f4"}},
			"module3": {ModulePath: "module3", Language: "go", Files: []string{"f5"}},
		},
		ExternalDeps: []ExternalDependency{},
	}

	graph, err := visualizer.CreateGraph(result)
	require.NoError(t, err)

	// Should be limited to 2 nodes
	assert.Len(t, graph.Nodes, 2)

	// Edges should be filtered to only include those between remaining nodes
	for _, edge := range graph.Edges {
		assert.True(t, nodeExists(graph.Nodes, edge.From))
		assert.True(t, nodeExists(graph.Nodes, edge.To))
	}
}

func TestDependencyVisualizer_CreateGraphExternalOnly(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VisualizerConfig{
		OutputFormat:    "json",
		IncludeExternal: true,
		IncludeInternal: false, // Only external dependencies
	}

	visualizer := NewDependencyVisualizer(logger, config)

	result := &DependencyResult{
		Repository: "/test/repo",
		Dependencies: []*Dependency{
			{From: "module1", To: "module2", External: false},
			{From: "module1", To: "external_lib", External: true},
		},
		Modules: map[string]*ModuleDependencies{
			"module1": {ModulePath: "module1", Language: "go"},
		},
		ExternalDeps: []ExternalDependency{
			{Name: "external_lib", Language: "go", UsageCount: 1},
		},
	}

	graph, err := visualizer.CreateGraph(result)
	require.NoError(t, err)

	// Should have 2 nodes (module1 + external_lib)
	assert.Len(t, graph.Nodes, 2)
	// Should have 1 edge (only the external dependency)
	assert.Len(t, graph.Edges, 1)
	assert.True(t, graph.Edges[0].To == "ext_go_external_lib")
}

func TestDependencyVisualizer_RenderJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "visualizer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger := zaptest.NewLogger(t)
	outputPath := filepath.Join(tempDir, "test_graph.json")
	config := &VisualizerConfig{
		OutputFormat:    "json",
		OutputPath:      outputPath,
		IncludeExternal: true,
		IncludeInternal: true,
	}

	visualizer := NewDependencyVisualizer(logger, config)

	// Create simple graph
	graph := &DependencyGraph{
		Nodes: []*GraphNode{
			{ID: "node1", Label: "Node 1", Type: "module", Language: "go"},
			{ID: "node2", Label: "Node 2", Type: "external", Language: "go"},
		},
		Edges: []*GraphEdge{
			{From: "node1", To: "node2", Type: DependencyTypeImport, Strength: DependencyStrengthStrong},
		},
		Metadata: &GraphMetadata{
			Title:       "Test Graph",
			GeneratedAt: time.Now(),
			Repository:  "/test/repo",
		},
		Statistics: &GraphStatistics{
			NodesByType: map[string]int{"module": 1, "external": 1},
		},
	}

	err = visualizer.RenderGraph(graph)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Verify content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Node 1")
	assert.Contains(t, string(data), "Node 2")
	assert.Contains(t, string(data), "Test Graph")
}

func TestDependencyVisualizer_RenderDOT(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "visualizer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger := zaptest.NewLogger(t)
	outputPath := filepath.Join(tempDir, "test_graph.dot")
	config := &VisualizerConfig{
		OutputFormat:    "dot",
		OutputPath:      outputPath,
		IncludeExternal: true,
		IncludeInternal: true,
		ShowLabels:      true,
		ClusterByType:   true,
		LayoutAlgorithm: "dot",
	}

	visualizer := NewDependencyVisualizer(logger, config)

	// Create graph with clusters
	graph := &DependencyGraph{
		Nodes: []*GraphNode{
			{ID: "module1", Label: "Module 1", Type: "module", Language: "go", Size: 10, Complexity: 2.5},
			{ID: "ext_go_lib", Label: "External Lib", Type: "external", Language: "go", External: true},
		},
		Edges: []*GraphEdge{
			{From: "module1", To: "ext_go_lib", Type: DependencyTypeImport, Strength: DependencyStrengthStrong, Weight: 2.0},
		},
		Metadata: &GraphMetadata{
			Title: "Test Dependency Graph",
		},
		Clusters: map[string][]string{
			"internal_go": {"module1"},
			"external_go": {"ext_go_lib"},
		},
	}

	err = visualizer.RenderGraph(graph)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Verify DOT content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "digraph dependencies")
	assert.Contains(t, content, "Module 1")
	assert.Contains(t, content, "External Lib")
	assert.Contains(t, content, "subgraph cluster_")
	assert.Contains(t, content, "->")
	assert.Contains(t, content, "Test Dependency Graph")
}

func TestDependencyVisualizer_RenderHTML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "visualizer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger := zaptest.NewLogger(t)
	outputPath := filepath.Join(tempDir, "test_graph.html")
	config := &VisualizerConfig{
		OutputFormat: "html",
		OutputPath:   outputPath,
		ShowLabels:   true,
	}

	visualizer := NewDependencyVisualizer(logger, config)

	// Create simple graph
	graph := &DependencyGraph{
		Nodes: []*GraphNode{
			{ID: "node1", Label: "Node 1", Type: "module", Language: "go"},
		},
		Edges: []*GraphEdge{},
		Metadata: &GraphMetadata{
			Title:      "Interactive Graph",
			Repository: "/test/repo",
			Languages:  []string{"go", "javascript"},
		},
		Statistics: &GraphStatistics{
			Density: 0.25,
		},
	}

	err = visualizer.RenderGraph(graph)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Verify HTML content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "<!DOCTYPE html>")
	assert.Contains(t, content, "Interactive Graph")
	assert.Contains(t, content, "d3js.org")
	assert.Contains(t, content, "/test/repo")
	assert.Contains(t, content, "Node 1")
}

func TestDependencyVisualizer_GetOutputPath(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with specified output path
	config1 := &VisualizerConfig{
		OutputPath: "/custom/path/graph.json",
	}
	visualizer1 := NewDependencyVisualizer(logger, config1)
	assert.Equal(t, "/custom/path/graph.json", visualizer1.getOutputPath("json"))

	// Test with automatic path generation
	config2 := &VisualizerConfig{
		OutputPath: "",
	}
	visualizer2 := NewDependencyVisualizer(logger, config2)
	autoPath := visualizer2.getOutputPath("dot")
	assert.Contains(t, autoPath, "dependency_graph_")
	assert.Contains(t, autoPath, ".dot")
}

func TestDependencyVisualizer_CalculateEdgeWeight(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VisualizerConfig{}
	visualizer := NewDependencyVisualizer(logger, config)

	tests := []struct {
		dep      *Dependency
		expected float64
	}{
		{
			dep:      &Dependency{Type: DependencyTypeImport, Strength: DependencyStrengthStrong},
			expected: 3.0,
		},
		{
			dep:      &Dependency{Type: DependencyTypeRequire, Strength: DependencyStrengthWeak},
			expected: 1.8, // 1.5 * 1.2
		},
		{
			dep:      &Dependency{Type: DependencyTypeInclude, Strength: DependencyStrengthOptional},
			expected: 0.4, // 0.5 * 0.8
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			weight := visualizer.calculateEdgeWeight(tt.dep)
			assert.InDelta(t, tt.expected, weight, 0.01)
		})
	}
}

func TestDependencyVisualizer_UnsupportedFormat(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &VisualizerConfig{
		OutputFormat: "unsupported",
	}

	visualizer := NewDependencyVisualizer(logger, config)
	graph := &DependencyGraph{}

	err := visualizer.RenderGraph(graph)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format")
}

// Helper function to check if a node exists in the slice
func nodeExists(nodes []*GraphNode, nodeID string) bool {
	for _, node := range nodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}
