package reposync

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"os"
	"time"

	"go.uber.org/zap"
)

// renderHTML renders the graph as an interactive HTML page
func (dv *DependencyVisualizer) renderHTML(graph *DependencyGraph) error {
	outputPath := dv.getOutputPath("html")

	// Prepare template data
	templateData := struct {
		Graph     *DependencyGraph
		Config    *VisualizerConfig
		Timestamp string
		GraphJSON string
	}{
		Graph:     graph,
		Config:    dv.config,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Serialize graph to JSON for JavaScript
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return fmt.Errorf("failed to marshal graph for HTML: %w", err)
	}
	templateData.GraphJSON = string(graphJSON)

	// Create HTML template
	tmpl, err := template.New("dependency_graph").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	dv.logger.Info("Graph exported as HTML",
		zap.String("path", outputPath),
		zap.Int("nodes", len(graph.Nodes)),
		zap.Int("edges", len(graph.Edges)))

	return nil
}

// renderSVG renders the graph as SVG (basic implementation)
func (dv *DependencyVisualizer) renderSVG(graph *DependencyGraph) error {
	outputPath := dv.getOutputPath("svg")

	// For now, generate a simple force-directed layout
	layout := dv.calculateForceDirectedLayout(graph)

	// Create SVG content
	svg := dv.generateSVG(graph, layout)

	if err := os.WriteFile(outputPath, []byte(svg), 0644); err != nil {
		return fmt.Errorf("failed to write SVG file: %w", err)
	}

	dv.logger.Info("Graph exported as SVG", zap.String("path", outputPath))
	return nil
}

// calculateForceDirectedLayout calculates node positions using a simple force-directed algorithm
func (dv *DependencyVisualizer) calculateForceDirectedLayout(graph *DependencyGraph) map[string]*NodePosition {
	positions := make(map[string]*NodePosition)

	// Initialize random positions
	for i, node := range graph.Nodes {
		angle := float64(i) * 2.0 * 3.14159 / float64(len(graph.Nodes))
		radius := 200.0
		positions[node.ID] = &NodePosition{
			X: 400 + radius*math.Cos(angle),
			Y: 300 + radius*math.Sin(angle),
		}
	}

	// Simple force-directed iterations (simplified)
	for iter := 0; iter < 50; iter++ {
		forces := make(map[string]*NodePosition)

		// Initialize forces
		for nodeID := range positions {
			forces[nodeID] = &NodePosition{X: 0, Y: 0}
		}

		// Repulsive forces between all nodes
		for _, node1 := range graph.Nodes {
			for _, node2 := range graph.Nodes {
				if node1.ID == node2.ID {
					continue
				}

				pos1 := positions[node1.ID]
				pos2 := positions[node2.ID]

				dx := pos1.X - pos2.X
				dy := pos1.Y - pos2.Y
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist > 0 {
					repulsion := 5000.0 / (dist * dist)
					forces[node1.ID].X += repulsion * dx / dist
					forces[node1.ID].Y += repulsion * dy / dist
				}
			}
		}

		// Attractive forces for connected nodes
		for _, edge := range graph.Edges {
			pos1 := positions[edge.From]
			pos2 := positions[edge.To]

			dx := pos2.X - pos1.X
			dy := pos2.Y - pos1.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > 0 {
				attraction := dist * 0.01 * edge.Weight
				forces[edge.From].X += attraction * dx / dist
				forces[edge.From].Y += attraction * dy / dist
				forces[edge.To].X -= attraction * dx / dist
				forces[edge.To].Y -= attraction * dy / dist
			}
		}

		// Apply forces
		for nodeID, force := range forces {
			positions[nodeID].X += force.X * 0.1
			positions[nodeID].Y += force.Y * 0.1
		}
	}

	return positions
}

// generateSVG generates SVG content for the graph
func (dv *DependencyVisualizer) generateSVG(graph *DependencyGraph, layout map[string]*NodePosition) string {
	svg := `<?xml version="1.0" encoding="UTF-8"?>
<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <marker id="arrowhead" markerWidth="10" markerHeight="7" 
            refX="10" refY="3.5" orient="auto">
      <polygon points="0 0, 10 3.5, 0 7" fill="#666"/>
    </marker>
  </defs>
  <rect width="100%" height="100%" fill="white"/>
`

	// Draw edges first (so they appear behind nodes)
	for _, edge := range graph.Edges {
		fromPos := layout[edge.From]
		toPos := layout[edge.To]

		if fromPos != nil && toPos != nil {
			strokeWidth := 1.0 + edge.Weight*0.5
			color := "#666"

			switch edge.Strength {
			case DependencyStrengthStrong:
				color = "#000"
			case DependencyStrengthWeak:
				color = "#999"
			case DependencyStrengthOptional:
				color = "#CCC"
			}

			svg += fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" 
            stroke="%s" stroke-width="%.1f" marker-end="url(#arrowhead)"/>
`, fromPos.X, fromPos.Y, toPos.X, toPos.Y, color, strokeWidth)
		}
	}

	// Draw nodes
	for _, node := range graph.Nodes {
		pos := layout[node.ID]
		if pos == nil {
			continue
		}

		radius := 10.0 + float64(node.Size)*0.5
		if radius > 30 {
			radius = 30
		}

		color := "#90EE90"
		if node.External {
			color = "#FFB6C1"
		}

		// Language-specific colors
		switch node.Language {
		case "go":
			color = "#00ADD8"
		case "javascript":
			color = "#F7DF1E"
		case "typescript":
			color = "#3178C6"
		case "python":
			color = "#3776AB"
		}

		svg += fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" 
          stroke="#333" stroke-width="1">
    <title>%s (%s)</title>
  </circle>
`, pos.X, pos.Y, radius, color, node.Label, node.Language)

		// Add label
		if dv.config.ShowLabels && node.Label != "" {
			svg += fmt.Sprintf(`  <text x="%.1f" y="%.1f" text-anchor="middle" 
            font-family="Arial" font-size="10" dy="0.3em">%s</text>
`, pos.X, pos.Y+radius+12, node.Label)
		}
	}

	// Add title
	svg += fmt.Sprintf(`  <text x="400" y="30" text-anchor="middle" 
          font-family="Arial" font-size="16" font-weight="bold">%s</text>
`, graph.Metadata.Title)

	svg += "</svg>"
	return svg
}

// HTML template for interactive dependency graph
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Graph.Metadata.Title}}</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .graph-container {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .controls {
            padding: 20px;
            border-bottom: 1px solid #eee;
        }
        .control-group {
            display: inline-block;
            margin-right: 20px;
        }
        .control-group label {
            font-weight: bold;
            margin-right: 5px;
        }
        #graph {
            width: 100%;
            height: 600px;
            border: 1px solid #ddd;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stat-value {
            font-size: 2em;
            font-weight: bold;
            color: #007bff;
        }
        .legend {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-top: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .legend-item {
            display: inline-block;
            margin-right: 20px;
            margin-bottom: 10px;
        }
        .legend-color {
            display: inline-block;
            width: 20px;
            height: 20px;
            border-radius: 3px;
            margin-right: 5px;
            vertical-align: middle;
        }
        .tooltip {
            position: absolute;
            padding: 10px;
            background: rgba(0, 0, 0, 0.8);
            color: white;
            border-radius: 4px;
            pointer-events: none;
            font-size: 12px;
            z-index: 1000;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Graph.Metadata.Title}}</h1>
            <p>{{.Graph.Metadata.Description}}</p>
            <p><strong>Repository:</strong> {{.Graph.Metadata.Repository}}</p>
            <p><strong>Generated:</strong> {{.Timestamp}}</p>
            <p><strong>Languages:</strong> {{range $i, $lang := .Graph.Metadata.Languages}}{{if $i}}, {{end}}{{$lang}}{{end}}</p>
        </div>

        <div class="graph-container">
            <div class="controls">
                <div class="control-group">
                    <label>Layout:</label>
                    <select id="layoutSelect">
                        <option value="force">Force Directed</option>
                        <option value="circular">Circular</option>
                        <option value="hierarchical">Hierarchical</option>
                    </select>
                </div>
                <div class="control-group">
                    <label>Node Size:</label>
                    <select id="nodeSizeSelect">
                        <option value="constant">Constant</option>
                        <option value="size">By Size</option>
                        <option value="complexity">By Complexity</option>
                    </select>
                </div>
                <div class="control-group">
                    <label>Show Labels:</label>
                    <input type="checkbox" id="showLabels" checked>
                </div>
                <div class="control-group">
                    <button onclick="resetZoom()">Reset Zoom</button>
                </div>
            </div>
            <svg id="graph"></svg>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">{{.Graph.Metadata.TotalNodes}}</div>
                <div>Total Nodes</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{{.Graph.Metadata.TotalEdges}}</div>
                <div>Total Edges</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{{printf "%.2f" .Graph.Statistics.Density}}</div>
                <div>Graph Density</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{{.Graph.Statistics.MaxDepth}}</div>
                <div>Max Depth</div>
            </div>
        </div>

        <div class="legend">
            <h3>Legend</h3>
            <div class="legend-item">
                <span class="legend-color" style="background-color: #00ADD8;"></span>
                Go
            </div>
            <div class="legend-item">
                <span class="legend-color" style="background-color: #F7DF1E;"></span>
                JavaScript
            </div>
            <div class="legend-item">
                <span class="legend-color" style="background-color: #3178C6;"></span>
                TypeScript
            </div>
            <div class="legend-item">
                <span class="legend-color" style="background-color: #3776AB;"></span>
                Python
            </div>
            <div class="legend-item">
                <span class="legend-color" style="background-color: #E0E0E0;"></span>
                External Dependencies
            </div>
        </div>
    </div>

    <div class="tooltip" id="tooltip" style="display: none;"></div>

    <script>
        // Graph data from Go template
        const graphData = {{.GraphJSON}};
        
        // D3.js visualization code
        const width = 1160;
        const height = 600;
        
        const svg = d3.select("#graph")
            .attr("width", width)
            .attr("height", height);
            
        const g = svg.append("g");
        
        // Zoom behavior
        const zoom = d3.zoom()
            .scaleExtent([0.1, 4])
            .on("zoom", (event) => {
                g.attr("transform", event.transform);
            });
            
        svg.call(zoom);
        
        // Tooltip
        const tooltip = d3.select("#tooltip");
        
        let simulation;
        
        function initializeGraph() {
            // Clear previous content
            g.selectAll("*").remove();
            
            // Create links and nodes data
            const nodes = graphData.nodes.map(d => ({...d}));
            const links = graphData.edges.map(d => ({
                source: d.from,
                target: d.to,
                ...d
            }));
            
            // Create simulation
            simulation = d3.forceSimulation(nodes)
                .force("link", d3.forceLink(links).id(d => d.id).distance(100))
                .force("charge", d3.forceManyBody().strength(-300))
                .force("center", d3.forceCenter(width / 2, height / 2));
            
            // Create links
            const link = g.append("g")
                .selectAll("line")
                .data(links)
                .enter().append("line")
                .attr("stroke", d => getEdgeColor(d))
                .attr("stroke-width", d => Math.max(1, d.weight))
                .attr("stroke-opacity", 0.6);
            
            // Create nodes
            const node = g.append("g")
                .selectAll("circle")
                .data(nodes)
                .enter().append("circle")
                .attr("r", d => getNodeRadius(d))
                .attr("fill", d => getNodeColor(d))
                .attr("stroke", "#333")
                .attr("stroke-width", 1.5)
                .on("mouseover", showTooltip)
                .on("mouseout", hideTooltip)
                .call(d3.drag()
                    .on("start", dragstarted)
                    .on("drag", dragged)
                    .on("end", dragended));
            
            // Create labels
            const labels = g.append("g")
                .selectAll("text")
                .data(nodes)
                .enter().append("text")
                .text(d => d.label)
                .attr("font-size", 10)
                .attr("text-anchor", "middle")
                .attr("dy", ".35em")
                .style("pointer-events", "none")
                .style("display", document.getElementById("showLabels").checked ? "block" : "none");
            
            // Update positions on simulation tick
            simulation.on("tick", () => {
                link
                    .attr("x1", d => d.source.x)
                    .attr("y1", d => d.source.y)
                    .attr("x2", d => d.target.x)
                    .attr("y2", d => d.target.y);
                
                node
                    .attr("cx", d => d.x)
                    .attr("cy", d => d.y);
                    
                labels
                    .attr("x", d => d.x)
                    .attr("y", d => d.y);
            });
        }
        
        function getNodeColor(d) {
            const colors = {
                "go": "#00ADD8",
                "javascript": "#F7DF1E", 
                "typescript": "#3178C6",
                "python": "#3776AB",
                "java": "#ED8B00"
            };
            
            if (d.external) {
                return "#E0E0E0";
            }
            
            return colors[d.language] || "#90EE90";
        }
        
        function getNodeRadius(d) {
            const sizeMode = document.getElementById("nodeSizeSelect").value;
            switch (sizeMode) {
                case "size":
                    return Math.max(5, Math.min(20, d.size));
                case "complexity":
                    return Math.max(5, Math.min(20, d.complexity * 5));
                default:
                    return 8;
            }
        }
        
        function getEdgeColor(d) {
            switch (d.strength) {
                case "strong": return "#000";
                case "weak": return "#666";
                case "optional": return "#CCC";
                default: return "#888";
            }
        }
        
        function showTooltip(event, d) {
            const tooltipContent = ` + "`" + `
                <strong>${d.label}</strong><br>
                Type: ${d.type}<br>
                Language: ${d.language}<br>
                ${d.external ? 'External' : 'Internal'}<br>
                ${d.size ? 'Size: ' + d.size + '<br>' : ''}
                ${d.complexity ? 'Complexity: ' + d.complexity.toFixed(1) + '<br>' : ''}
                ${d.properties && d.properties.version ? 'Version: ' + d.properties.version : ''}
            ` + "`" + `;
            
            tooltip
                .style("display", "block")
                .html(tooltipContent)
                .style("left", (event.pageX + 10) + "px")
                .style("top", (event.pageY - 10) + "px");
        }
        
        function hideTooltip() {
            tooltip.style("display", "none");
        }
        
        function dragstarted(event, d) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        }
        
        function dragged(event, d) {
            d.fx = event.x;
            d.fy = event.y;
        }
        
        function dragended(event, d) {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
        }
        
        function resetZoom() {
            svg.transition().duration(750).call(
                zoom.transform,
                d3.zoomIdentity
            );
        }
        
        // Event listeners
        document.getElementById("showLabels").addEventListener("change", function(e) {
            g.selectAll("text").style("display", e.target.checked ? "block" : "none");
        });
        
        document.getElementById("nodeSizeSelect").addEventListener("change", function() {
            g.selectAll("circle").attr("r", d => getNodeRadius(d));
        });
        
        document.getElementById("layoutSelect").addEventListener("change", function(e) {
            const layout = e.target.value;
            if (layout === "circular") {
                // Implement circular layout
                const nodes = simulation.nodes();
                nodes.forEach((d, i) => {
                    const angle = (i / nodes.length) * 2 * Math.PI;
                    const radius = 200;
                    d.fx = width/2 + radius * Math.cos(angle);
                    d.fy = height/2 + radius * Math.sin(angle);
                });
                simulation.alpha(0.3).restart();
            } else {
                // Reset to force-directed
                simulation.nodes().forEach(d => {
                    d.fx = null;
                    d.fy = null;
                });
                simulation.alpha(0.3).restart();
            }
        });
        
        // Initialize the graph
        initializeGraph();
    </script>
</body>
</html>`
