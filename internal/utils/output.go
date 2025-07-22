// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// OutputFormatter provides consistent output formatting.
type OutputFormatter struct {
	useColor bool
	verbose  bool
}

// NewOutputFormatter creates a new output formatter.
func NewOutputFormatter(useColor, verbose bool) *OutputFormatter {
	return &OutputFormatter{
		useColor: useColor,
		verbose:  verbose,
	}
}

// Success prints a success message.
func (of *OutputFormatter) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if of.useColor {
		color.Green("✅ " + msg)
	} else {
		fmt.Println("✅ " + msg)
	}
}

// Error prints an error message.
func (of *OutputFormatter) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if of.useColor {
		color.Red("❌ " + msg)
	} else {
		fmt.Fprintln(os.Stderr, "❌ "+msg)
	}
}

// Warning prints a warning message.
func (of *OutputFormatter) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if of.useColor {
		color.Yellow("⚠️  " + msg)
	} else {
		fmt.Println("⚠️  " + msg)
	}
}

// Info prints an info message.
func (of *OutputFormatter) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if of.useColor {
		color.Cyan("ℹ️  " + msg)
	} else {
		fmt.Println("ℹ️  " + msg)
	}
}

// Debug prints a debug message if verbose is enabled.
func (of *OutputFormatter) Debug(format string, args ...interface{}) {
	if of.verbose {
		msg := fmt.Sprintf(format, args...)
		if of.useColor {
			color.HiBlack("🔍 " + msg)
		} else {
			fmt.Println("🔍 " + msg)
		}
	}
}

// ProgressBar creates a new progress bar.
func (of *OutputFormatter) ProgressBar(max int, description string) *progressbar.ProgressBar {
	options := []progressbar.Option{
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	}

	if !of.useColor {
		options = append(options, progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	}

	return progressbar.NewOptions(max, options...)
}

// Table formats data as a simple table.
func (of *OutputFormatter) Table(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	headerLine := ""
	separator := ""

	for i, h := range headers {
		headerLine += fmt.Sprintf("%-*s  ", widths[i], h)
		separator += strings.Repeat("-", widths[i]) + "  "
	}

	if of.useColor {
		color.Set(color.Bold)
		fmt.Println(headerLine)
		color.Unset()
	} else {
		fmt.Println(headerLine)
	}

	fmt.Println(separator)

	// Print rows
	for _, row := range rows {
		rowLine := ""

		for i, cell := range row {
			if i < len(widths) {
				rowLine += fmt.Sprintf("%-*s  ", widths[i], cell)
			}
		}

		fmt.Println(rowLine)
	}
}
