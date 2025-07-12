package reposync

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

// renderDOT renders the graph in Graphviz DOT format
func (dv *DependencyVisualizer) renderDOT(graph *DependencyGraph) error {
	outputPath := dv.getOutputPath("dot")
	
	var sb strings.Builder
	
	// Write DOT header
	sb.WriteString("digraph dependencies {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [shape=box, fontname=\"Arial\", fontsize=10];\n")
	sb.WriteString("  edge [fontname=\"Arial\", fontsize=8];\n")
	sb.WriteString("  overlap=false;\n")
	sb.WriteString("  splines=true;\n")
	
	// Set layout algorithm
	if dv.config.LayoutAlgorithm != "" {
		sb.WriteString(fmt.Sprintf("  layout=%s;\n", dv.config.LayoutAlgorithm))
	}
	
	sb.WriteString("\n")
	
	// Add title
	sb.WriteString(fmt.Sprintf("  label=\"%s\";\n", graph.Metadata.Title))
	sb.WriteString("  labelloc=t;\n")
	sb.WriteString("  fontsize=16;\n")
	sb.WriteString("  fontname=\"Arial Bold\";\n\n")
	
	// Write clusters if enabled
	if dv.config.ClusterByType && len(graph.Clusters) > 0 {
		dv.writeDOTClusters(&sb, graph)
	} else {
		dv.writeDOTNodes(&sb, graph.Nodes)
	}
	
	// Write edges
	dv.writeDOTEdges(&sb, graph.Edges)
	
	// Write DOT footer
	sb.WriteString("}\n")
	
	// Write to file
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write DOT file: %w", err)
	}
	
	dv.logger.Info("Graph exported as DOT", 
		zap.String("path", outputPath),
		zap.Int("nodes", len(graph.Nodes)),
		zap.Int("edges", len(graph.Edges)))
	
	return nil
}

// writeDOTClusters writes clustered nodes to DOT format
func (dv *DependencyVisualizer) writeDOTClusters(sb *strings.Builder, graph *DependencyGraph) {
	clusterIndex := 0
	nodeWritten := make(map[string]bool)
	
	for clusterName, nodeIDs := range graph.Clusters {
		if len(nodeIDs) == 0 {
			continue
		}
		
		sb.WriteString(fmt.Sprintf("  subgraph cluster_%d {\n", clusterIndex))
		sb.WriteString(fmt.Sprintf("    label=\"%s\";\n", dv.formatClusterLabel(clusterName)))
		sb.WriteString("    style=filled;\n")
		sb.WriteString(fmt.Sprintf("    fillcolor=\"%s\";\n", dv.getClusterColor(clusterName)))
		sb.WriteString("    fontname=\"Arial Bold\";\n")
		sb.WriteString("    fontsize=12;\n\n")
		
		// Write nodes in this cluster
		for _, nodeID := range nodeIDs {
			for _, node := range graph.Nodes {
				if node.ID == nodeID {
					dv.writeDOTNode(sb, node, "    ")
					nodeWritten[nodeID] = true
					break
				}
			}
		}
		
		sb.WriteString("  }\n\n")
		clusterIndex++
	}
	
	// Write any remaining nodes that weren't in clusters
	for _, node := range graph.Nodes {
		if !nodeWritten[node.ID] {
			dv.writeDOTNode(sb, node, "  ")
		}
	}
}

// writeDOTNodes writes all nodes to DOT format
func (dv *DependencyVisualizer) writeDOTNodes(sb *strings.Builder, nodes []*GraphNode) {
	for _, node := range nodes {
		dv.writeDOTNode(sb, node, "  ")
	}
	sb.WriteString("\n")
}

// writeDOTNode writes a single node to DOT format
func (dv *DependencyVisualizer) writeDOTNode(sb *strings.Builder, node *GraphNode, indent string) {
	nodeID := dv.escapeDOTString(node.ID)
	label := dv.escapeDOTString(node.Label)
	
	// Build node attributes
	var attrs []string
	
	// Label
	if label != "" {
		attrs = append(attrs, fmt.Sprintf("label=\"%s\"", label))
	}
	
	// Shape based on node type
	shape := dv.getNodeShape(node)
	attrs = append(attrs, fmt.Sprintf("shape=%s", shape))
	
	// Color based on language and type
	color := dv.getNodeColor(node)
	attrs = append(attrs, fmt.Sprintf("fillcolor=\"%s\"", color))
	attrs = append(attrs, "style=filled")
	
	// Size based on complexity/size
	if node.Size > 0 || node.Complexity > 0 {
		width := 0.5 + (float64(node.Size)*0.1 + node.Complexity*0.2)
		if width > 3.0 {
			width = 3.0
		}
		attrs = append(attrs, fmt.Sprintf("width=%.2f", width))
	}
	
	// Font size based on importance
	fontSize := 10
	if importance, ok := node.Properties["importance"].(float64); ok {
		fontSize = 8 + int(importance*0.5)
		if fontSize > 16 {
			fontSize = 16
		}
		if fontSize < 8 {
			fontSize = 8
		}
	}
	attrs = append(attrs, fmt.Sprintf("fontsize=%d", fontSize))
	
	// External nodes get different border
	if node.External {
		attrs = append(attrs, "penwidth=2")
		attrs = append(attrs, "style=\"filled,dashed\"")
	}
	
	// Tooltip with additional information
	tooltip := dv.createNodeTooltip(node)
	if tooltip != "" {
		attrs = append(attrs, fmt.Sprintf("tooltip=\"%s\"", dv.escapeDOTString(tooltip)))
	}
	
	// Write the node
	sb.WriteString(fmt.Sprintf("%s\"%s\" [%s];\n", 
		indent, nodeID, strings.Join(attrs, ", ")))
}

// writeDOTEdges writes all edges to DOT format
func (dv *DependencyVisualizer) writeDOTEdges(sb *strings.Builder, edges []*GraphEdge) {
	sb.WriteString("  // Edges\n")
	
	for _, edge := range edges {
		fromID := dv.escapeDOTString(edge.From)
		toID := dv.escapeDOTString(edge.To)
		
		// Build edge attributes
		var attrs []string
		
		// Weight affects edge thickness
		penwidth := 1.0 + edge.Weight*0.5
		if penwidth > 5.0 {
			penwidth = 5.0
		}
		attrs = append(attrs, fmt.Sprintf("penwidth=%.1f", penwidth))
		
		// Color based on dependency type
		color := dv.getEdgeColor(edge)
		attrs = append(attrs, fmt.Sprintf("color=\"%s\"", color))
		
		// Style based on strength
		style := dv.getEdgeStyle(edge)
		if style != "solid" {
			attrs = append(attrs, fmt.Sprintf("style=%s", style))
		}
		
		// Arrow style based on type
		arrowhead := dv.getArrowHead(edge)
		if arrowhead != "normal" {
			attrs = append(attrs, fmt.Sprintf("arrowhead=%s", arrowhead))
		}
		
		// Label for edge type (optional)
		if dv.config.ShowLabels {
			label := string(edge.Type)
			attrs = append(attrs, fmt.Sprintf("label=\"%s\"", label))
			attrs = append(attrs, "fontsize=8")
		}
		
		// Tooltip
		tooltip := dv.createEdgeTooltip(edge)
		if tooltip != "" {
			attrs = append(attrs, fmt.Sprintf("tooltip=\"%s\"", dv.escapeDOTString(tooltip)))
		}
		
		// Write the edge
		sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [%s];\n", 
			fromID, toID, strings.Join(attrs, ", ")))
	}
}

// getNodeShape returns the shape for a node based on its type
func (dv *DependencyVisualizer) getNodeShape(node *GraphNode) string {
	switch node.Type {
	case "module":
		return "box"
	case "external":
		return "ellipse"
	case "file":
		return "note"
	default:
		return "box"
	}
}

// getNodeColor returns the color for a node based on language and type
func (dv *DependencyVisualizer) getNodeColor(node *GraphNode) string {
	if customColor, exists := dv.config.NodeStyles[node.Language]; exists {
		return customColor
	}
	
	// Default colors by language
	languageColors := map[string]string{
		"go":         "#00ADD8",
		"javascript": "#F7DF1E",
		"typescript": "#3178C6",
		"python":     "#3776AB",
		"java":       "#ED8B00",
		"rust":       "#000000",
		"cpp":        "#00599C",
		"csharp":     "#239120",
		"php":        "#777BB4",
		"ruby":       "#CC342D",
	}
	
	if color, exists := languageColors[node.Language]; exists {
		// Make external nodes lighter
		if node.External {
			return color + "80" // Add transparency
		}
		return color
	}
	
	// Default colors
	if node.External {
		return "#E0E0E0"
	}
	return "#90EE90"
}

// getClusterColor returns the color for a cluster
func (dv *DependencyVisualizer) getClusterColor(clusterName string) string {
	if strings.Contains(clusterName, "external") {
		return "#F0F0F0"
	}
	return "#E6F3FF"
}

// getEdgeColor returns the color for an edge based on its properties
func (dv *DependencyVisualizer) getEdgeColor(edge *GraphEdge) string {
	if customColor, exists := dv.config.EdgeStyles[string(edge.Type)]; exists {
		return customColor
	}
	
	switch edge.Strength {
	case DependencyStrengthStrong:
		return "#000000"
	case DependencyStrengthWeak:
		return "#666666"
	case DependencyStrengthOptional:
		return "#CCCCCC"
	default:
		return "#888888"
	}
}

// getEdgeStyle returns the style for an edge based on its strength
func (dv *DependencyVisualizer) getEdgeStyle(edge *GraphEdge) string {
	switch edge.Strength {
	case DependencyStrengthOptional:
		return "dashed"
	case DependencyStrengthWeak:
		return "dotted"
	default:
		return "solid"
	}
}

// getArrowHead returns the arrow head style for an edge
func (dv *DependencyVisualizer) getArrowHead(edge *GraphEdge) string {
	switch edge.Type {
	case DependencyTypeImport:
		return "normal"
	case DependencyTypeRequire:
		return "vee"
	case DependencyTypeInclude:
		return "diamond"
	case DependencyTypeInherit:
		return "empty"
	default:
		return "normal"
	}
}

// formatClusterLabel formats the label for a cluster
func (dv *DependencyVisualizer) formatClusterLabel(clusterName string) string {
	parts := strings.Split(clusterName, "_")
	if len(parts) >= 2 {
		scope := strings.Title(parts[0])
		language := strings.Title(parts[1])
		return fmt.Sprintf("%s (%s)", scope, language)
	}
	return strings.Title(clusterName)
}

// createNodeTooltip creates a tooltip for a node
func (dv *DependencyVisualizer) createNodeTooltip(node *GraphNode) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("Type: %s", node.Type))
	parts = append(parts, fmt.Sprintf("Language: %s", node.Language))
	
	if node.Size > 0 {
		if node.Type == "module" {
			parts = append(parts, fmt.Sprintf("Files: %d", node.Size))
		} else {
			parts = append(parts, fmt.Sprintf("Usage: %d", node.Size))
		}
	}
	
	if node.Complexity > 0 {
		parts = append(parts, fmt.Sprintf("Complexity: %.1f", node.Complexity))
	}
	
	if version, ok := node.Properties["version"].(string); ok && version != "" {
		parts = append(parts, fmt.Sprintf("Version: %s", version))
	}
	
	if description, ok := node.Properties["description"].(string); ok && description != "" {
		if len(description) > 50 {
			description = description[:50] + "..."
		}
		parts = append(parts, fmt.Sprintf("Description: %s", description))
	}
	
	return strings.Join(parts, "\\n")
}

// createEdgeTooltip creates a tooltip for an edge
func (dv *DependencyVisualizer) createEdgeTooltip(edge *GraphEdge) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("Type: %s", edge.Type))
	parts = append(parts, fmt.Sprintf("Strength: %s", edge.Strength))
	
	if language, ok := edge.Properties["language"].(string); ok {
		parts = append(parts, fmt.Sprintf("Language: %s", language))
	}
	
	if version, ok := edge.Properties["version"].(string); ok && version != "" {
		parts = append(parts, fmt.Sprintf("Version: %s", version))
	}
	
	return strings.Join(parts, "\\n")
}

// escapeDOTString escapes special characters for DOT format
func (dv *DependencyVisualizer) escapeDOTString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}