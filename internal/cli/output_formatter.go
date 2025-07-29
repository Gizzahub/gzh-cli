// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// OutputFormatter provides consistent output formatting across commands.
type OutputFormatter struct {
	writer io.Writer
	format string
}

// NewOutputFormatter creates a new output formatter.
func NewOutputFormatter(format string) *OutputFormatter {
	return &OutputFormatter{
		writer: os.Stdout,
		format: format,
	}
}

// NewOutputFormatterWithWriter creates a new output formatter with custom writer.
func NewOutputFormatterWithWriter(format string, writer io.Writer) *OutputFormatter {
	return &OutputFormatter{
		writer: writer,
		format: format,
	}
}

// FormatOutput formats and outputs data in the specified format.
func (f *OutputFormatter) FormatOutput(data interface{}) error {
	switch f.format {
	case "json":
		return f.outputJSON(data)
	case "yaml":
		return f.outputYAML(data)
	case "table":
		return f.outputTable(data)
	default:
		return fmt.Errorf("unsupported output format: %s", f.format)
	}
}

// outputJSON outputs data in JSON format.
func (f *OutputFormatter) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputYAML outputs data in YAML format.
func (f *OutputFormatter) outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// outputTable outputs data in table format (placeholder - specific tables should implement their own).
func (f *OutputFormatter) outputTable(data interface{}) error {
	// This is a generic fallback - specific commands should implement their own table formatting
	return fmt.Errorf("table format not implemented for this data type")
}

// TableData represents data that can be formatted as a table.
type TableData interface {
	GetHeaders() []string
	GetRows() [][]string
}

// FormatTable formats table data with consistent styling.
func (f *OutputFormatter) FormatTable(data TableData) error {
	if f.format != "table" {
		return f.FormatOutput(data)
	}

	headers := data.GetHeaders()
	rows := data.GetRows()

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	f.printRow(headers, colWidths)
	f.printSeparator(colWidths)

	// Print rows
	for _, row := range rows {
		f.printRow(row, colWidths)
	}

	return nil
}

// printRow prints a table row with proper spacing.
func (f *OutputFormatter) printRow(cells []string, colWidths []int) {
	var parts []string
	for i, cell := range cells {
		if i < len(colWidths) {
			parts = append(parts, fmt.Sprintf("%-*s", colWidths[i], cell))
		}
	}
	fmt.Fprintln(f.writer, strings.Join(parts, "  "))
}

// printSeparator prints a separator line.
func (f *OutputFormatter) printSeparator(colWidths []int) {
	var parts []string
	for _, width := range colWidths {
		parts = append(parts, strings.Repeat("-", width))
	}
	fmt.Fprintln(f.writer, strings.Join(parts, "  "))
}

// PrintSuccess prints a success message.
func (f *OutputFormatter) PrintSuccess(message string) {
	fmt.Fprintf(f.writer, "✅ %s\n", message)
}

// PrintError prints an error message.
func (f *OutputFormatter) PrintError(message string) {
	fmt.Fprintf(f.writer, "❌ %s\n", message)
}

// PrintWarning prints a warning message.
func (f *OutputFormatter) PrintWarning(message string) {
	fmt.Fprintf(f.writer, "⚠️  %s\n", message)
}

// PrintInfo prints an info message.
func (f *OutputFormatter) PrintInfo(message string) {
	fmt.Fprintf(f.writer, "ℹ️  %s\n", message)
}

// PrintVerbose prints a verbose message if verbose mode is enabled.
func (f *OutputFormatter) PrintVerbose(verbose bool, message string) {
	if verbose {
		fmt.Fprintf(f.writer, "🔍 %s\n", message)
	}
}
