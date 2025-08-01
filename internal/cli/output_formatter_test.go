// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTableData implements TableData interface for testing.
type TestTableData struct {
	headers []string
	rows    [][]string
}

func (t *TestTableData) GetHeaders() []string {
	return t.headers
}

func (t *TestTableData) GetRows() [][]string {
	return t.rows
}

func TestNewOutputFormatter(t *testing.T) {
	formatter := NewOutputFormatter("json")

	assert.NotNil(t, formatter)
	assert.Equal(t, "json", formatter.format)
	assert.NotNil(t, formatter.writer)
}

func TestNewOutputFormatterWithWriter(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("yaml", buffer)

	assert.NotNil(t, formatter)
	assert.Equal(t, "yaml", formatter.format)
	assert.Equal(t, buffer, formatter.writer)
}

func TestOutputFormatter_FormatOutput_JSON(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	testData := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
		"active":  true,
	}

	err := formatter.FormatOutput(testData)
	assert.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "\"name\": \"test\"")
	assert.Contains(t, output, "\"version\": \"1.0.0\"")
	assert.Contains(t, output, "\"active\": true")

	// Check for proper JSON indentation
	assert.Contains(t, output, "  \"name\"") // Should have 2-space indentation
}

func TestOutputFormatter_FormatOutput_YAML(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("yaml", buffer)

	testData := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
		"active":  true,
	}

	err := formatter.FormatOutput(testData)
	assert.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "name: test")
	assert.Contains(t, output, "version: 1.0.0")
	assert.Contains(t, output, "active: true")
}

func TestOutputFormatter_FormatOutput_Table_UnsupportedType(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("table", buffer)

	testData := "simple string"

	err := formatter.FormatOutput(testData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table format not implemented for this data type")
}

func TestOutputFormatter_FormatOutput_UnsupportedFormat(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("xml", buffer)

	testData := map[string]string{"key": "value"}

	err := formatter.FormatOutput(testData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format: xml")
}

func TestOutputFormatter_FormatTable(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("table", buffer)

	testData := &TestTableData{
		headers: []string{"Name", "Version", "Status"},
		rows: [][]string{
			{"app1", "1.0.0", "active"},
			{"application-with-long-name", "2.1.0", "inactive"},
			{"app3", "0.5.0", "pending"},
		},
	}

	err := formatter.FormatTable(testData)
	assert.NoError(t, err)

	output := buffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header + separator + 3 data rows = 5 lines
	assert.Len(t, lines, 5)

	// Check header
	assert.Contains(t, lines[0], "Name")
	assert.Contains(t, lines[0], "Version")
	assert.Contains(t, lines[0], "Status")

	// Check separator (should contain dashes)
	assert.Contains(t, lines[1], "---")

	// Check data rows
	assert.Contains(t, lines[2], "app1")
	assert.Contains(t, lines[2], "1.0.0")
	assert.Contains(t, lines[2], "active")

	assert.Contains(t, lines[3], "application-with-long-name")
	assert.Contains(t, lines[3], "2.1.0")
	assert.Contains(t, lines[3], "inactive")

	// Verify column alignment (longest name should determine column width)
	separatorLine := lines[1]

	// The separator should have appropriate spacing for the longest column content
	assert.True(t, len(separatorLine) >= len("application-with-long-name"))
}

func TestOutputFormatter_FormatTable_EmptyData(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("table", buffer)

	testData := &TestTableData{
		headers: []string{"Col1", "Col2"},
		rows:    [][]string{},
	}

	err := formatter.FormatTable(testData)
	assert.NoError(t, err)

	output := buffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header + separator only = 2 lines
	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "Col1")
	assert.Contains(t, lines[0], "Col2")
	assert.Contains(t, lines[1], "---")
}

func TestOutputFormatter_FormatTable_NonTableFormat(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	testData := &TestTableData{
		headers: []string{"Name", "Version"},
		rows: [][]string{
			{"app1", "1.0.0"},
		},
	}

	err := formatter.FormatTable(testData)
	assert.NoError(t, err)

	// Should fall back to FormatOutput for non-table formats
	output := buffer.String()
	// The output should be valid JSON (even if just "{}" for empty struct)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "{")
}

func TestOutputFormatter_FormatTable_UnevenRows(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("table", buffer)

	testData := &TestTableData{
		headers: []string{"Col1", "Col2", "Col3"},
		rows: [][]string{
			{"short", "medium-length", "very-very-long-content"},
			{"a", "b"},               // Missing third column
			{"x", "y", "z", "extra"}, // Extra column
		},
	}

	err := formatter.FormatTable(testData)
	assert.NoError(t, err)

	output := buffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should handle uneven rows gracefully
	assert.Len(t, lines, 5) // header + separator + 3 rows

	// All lines should be present
	for i, line := range lines {
		assert.NotEmpty(t, line, "Line %d should not be empty", i)
	}
}

func TestOutputFormatter_PrintSuccess(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	formatter.PrintSuccess("Operation completed successfully")

	output := buffer.String()
	assert.Contains(t, output, "✅ Operation completed successfully")
	assert.Contains(t, output, "\n") // Should end with newline
}

func TestOutputFormatter_PrintError(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	formatter.PrintError("Something went wrong")

	output := buffer.String()
	assert.Contains(t, output, "❌ Something went wrong")
	assert.Contains(t, output, "\n") // Should end with newline
}

func TestOutputFormatter_PrintWarning(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	formatter.PrintWarning("This is a warning")

	output := buffer.String()
	assert.Contains(t, output, "⚠️  This is a warning")
	assert.Contains(t, output, "\n") // Should end with newline
}

func TestOutputFormatter_PrintInfo(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	formatter.PrintInfo("Informational message")

	output := buffer.String()
	assert.Contains(t, output, "ℹ️  Informational message")
	assert.Contains(t, output, "\n") // Should end with newline
}

func TestOutputFormatter_PrintVerbose(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	tests := []struct {
		name     string
		verbose  bool
		message  string
		expected string
	}{
		{
			name:     "verbose_enabled",
			verbose:  true,
			message:  "Debug information",
			expected: "🔍 Debug information\n",
		},
		{
			name:     "verbose_disabled",
			verbose:  false,
			message:  "Debug information",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer.Reset()
			formatter.PrintVerbose(test.verbose, test.message)

			output := buffer.String()
			assert.Equal(t, test.expected, output)
		})
	}
}

func TestOutputFormatter_printRow(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("table", buffer)

	cells := []string{"short", "medium-length", "very-very-long"}
	colWidths := []int{10, 15, 20}

	formatter.printRow(cells, colWidths)

	output := buffer.String()

	// Should contain all cells
	assert.Contains(t, output, "short")
	assert.Contains(t, output, "medium-length")
	assert.Contains(t, output, "very-very-long")

	// Should have proper spacing (2 spaces between columns)
	assert.Contains(t, output, "  ")

	// Should end with newline
	assert.Contains(t, output, "\n")
}

func TestOutputFormatter_printRow_FewerCellsThanWidths(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("table", buffer)

	cells := []string{"cell1", "cell2"}
	colWidths := []int{10, 10, 10, 10} // More widths than cells

	formatter.printRow(cells, colWidths)

	output := buffer.String()
	assert.Contains(t, output, "cell1")
	assert.Contains(t, output, "cell2")

	// Should handle gracefully without panicking
	assert.NotEmpty(t, output)
}

func TestOutputFormatter_printSeparator(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("table", buffer)

	colWidths := []int{5, 10, 15}

	formatter.printSeparator(colWidths)

	output := buffer.String()

	// Should contain dashes for each column width
	assert.Contains(t, output, "-----")           // 5 dashes
	assert.Contains(t, output, "----------")      // 10 dashes
	assert.Contains(t, output, "---------------") // 15 dashes

	// Should have spacing between columns
	assert.Contains(t, output, "  ")

	// Should end with newline
	assert.Contains(t, output, "\n")
}

func TestOutputFormatter_ComplexJSONStructure(t *testing.T) {
	buffer := &bytes.Buffer{}
	formatter := NewOutputFormatterWithWriter("json", buffer)

	complexData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":    "test-app",
			"version": "1.2.3",
		},
		"config": map[string]interface{}{
			"enabled": true,
			"ports":   []int{80, 443, 8080},
		},
		"tags": []string{"web", "api", "production"},
	}

	err := formatter.FormatOutput(complexData)
	assert.NoError(t, err)

	output := buffer.String()

	// Verify nested structure is properly formatted
	assert.Contains(t, output, "\"metadata\"")
	assert.Contains(t, output, "\"config\"")
	assert.Contains(t, output, "\"ports\"")
	assert.Contains(t, output, "\"tags\"")

	// Verify arrays are formatted correctly
	assert.Contains(t, output, "[")
	assert.Contains(t, output, "]")

	// Verify proper indentation (should be pretty-printed)
	lines := strings.Split(output, "\n")
	indentedLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "  ") {
			indentedLines++
		}
	}
	assert.Greater(t, indentedLines, 0, "JSON should have indented lines")
}

func TestTableData_Interface(t *testing.T) {
	testData := &TestTableData{
		headers: []string{"A", "B", "C"},
		rows: [][]string{
			{"1", "2", "3"},
			{"4", "5", "6"},
		},
	}

	// Verify it implements the interface
	var _ TableData = testData

	assert.Equal(t, []string{"A", "B", "C"}, testData.GetHeaders())
	assert.Equal(t, [][]string{{"1", "2", "3"}, {"4", "5", "6"}}, testData.GetRows())
}

func TestOutputFormatter_Integration(t *testing.T) {
	// Test a complete workflow with different formats
	testData := map[string]interface{}{
		"name":   "integration-test",
		"status": "success",
		"count":  42,
	}

	formats := []string{"json", "yaml"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			formatter := NewOutputFormatterWithWriter(format, buffer)

			// Test data output
			err := formatter.FormatOutput(testData)
			require.NoError(t, err)
			assert.NotEmpty(t, buffer.String())

			// Test message outputs
			buffer.Reset()
			formatter.PrintSuccess("Success message")
			formatter.PrintError("Error message")
			formatter.PrintWarning("Warning message")
			formatter.PrintInfo("Info message")
			formatter.PrintVerbose(true, "Verbose message")

			output := buffer.String()
			assert.Contains(t, output, "✅")
			assert.Contains(t, output, "❌")
			assert.Contains(t, output, "⚠️")
			assert.Contains(t, output, "ℹ️")
			assert.Contains(t, output, "🔍")
		})
	}
}
