// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type openOptions struct {
	verbose    bool
	background bool
	wait       bool
}

// newIDEOpenCmd creates the IDE open subcommand
func newIDEOpenCmd() *cobra.Command {
	o := &openOptions{}

	cmd := &cobra.Command{
		Use:   "open <ide-name> [path]",
		Short: "Open an IDE with specified project path",
		Long: `Open an IDE application with the specified project path.

The command accepts an IDE name (or alias) and an optional path argument.
If no path is provided, the current directory is used.

Supported IDE names and aliases:
- JetBrains IDEs: pycharm, idea, webstorm, goland, clion, etc.
- VS Code family: code, vscode, cursor, codium
- Other editors: vim, nvim, emacs, subl

Path resolution:
- "." or empty: current directory
- Relative path: resolved from current directory
- Absolute path: used as-is
- Project name: searches in common project directories

Examples:
  # Open PyCharm in current directory
  gz ide open pycharm

  # Open VS Code in current directory (explicit)
  gz ide open code .

  # Open Cursor in specific directory
  gz ide open cursor ~/projects/myapp

  # Open GoLand in relative directory
  gz ide open goland ../other-project

  # Open IDE in background (don't wait for exit)
  gz ide open pycharm --background`,
		Args: cobra.RangeArgs(1, 2),
		RunE: o.runOpen,
	}

	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Show detailed execution information")
	cmd.Flags().BoolVar(&o.background, "background", false, "Run IDE in background (don't wait for exit)")
	cmd.Flags().BoolVar(&o.wait, "wait", false, "Wait for IDE to exit (opposite of background)")

	return cmd
}

func (o *openOptions) runOpen(cmd *cobra.Command, args []string) error {
	ideNameOrAlias := args[0]

	// Determine target path
	var targetPath string
	if len(args) > 1 {
		targetPath = args[1]
	} else {
		targetPath = "."
	}

	// Resolve target path
	resolvedPath, err := o.resolvePath(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve target path '%s': %w", targetPath, err)
	}

	// Check if target path exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return fmt.Errorf("target path does not exist: %s", resolvedPath)
	}

	// Detect IDEs
	detector := NewIDEDetector()
	ides, err := detector.DetectIDEs(true) // Use cache for opening
	if err != nil {
		return fmt.Errorf("failed to detect IDEs: %w", err)
	}

	// Find IDE by name or alias
	targetIDE := detector.FindIDEByAlias(ides, ideNameOrAlias)
	if targetIDE == nil {
		return fmt.Errorf("IDE '%s' not found. Available IDEs: %s",
			ideNameOrAlias, o.getAvailableIDENames(ides))
	}

	// Verify IDE executable exists
	if _, err := os.Stat(targetIDE.Executable); os.IsNotExist(err) {
		return fmt.Errorf("IDE executable not found: %s", targetIDE.Executable)
	}

	if o.verbose {
		fmt.Printf("🚀 Opening %s...\n", targetIDE.Name)
		fmt.Printf("   Executable: %s\n", targetIDE.Executable)
		fmt.Printf("   Target path: %s\n", resolvedPath)
		fmt.Printf("   Absolute path: %s\n", resolvedPath)
		if o.background {
			fmt.Printf("   Running in background\n")
		}
		fmt.Println()
	}

	// Execute IDE
	return o.executeIDE(targetIDE, resolvedPath)
}

func (o *openOptions) resolvePath(path string) (string, error) {
	// Handle special cases
	if path == "." || path == "" {
		return os.Getwd()
	}

	// Expand tilde
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[2:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func (o *openOptions) executeIDE(ide *IDE, targetPath string) error {
	// Prepare command arguments based on IDE type
	args := o.prepareIDEArgs(ide, targetPath)

	// Create command
	cmd := exec.Command(ide.Executable, args...)

	// Set working directory to target path if it's a directory
	if info, err := os.Stat(targetPath); err == nil && info.IsDir() {
		cmd.Dir = targetPath
	}

	// Determine execution strategy
	shouldRunInBackground := o.shouldRunInBackground(ide)

	if shouldRunInBackground {
		return o.executeInBackground(cmd, ide)
	} else {
		return o.executeAndWait(cmd, ide)
	}
}

func (o *openOptions) prepareIDEArgs(ide *IDE, targetPath string) []string {
	switch ide.Type {
	case "jetbrains":
		return o.prepareJetBrainsArgs(ide, targetPath)
	case "vscode":
		return o.prepareVSCodeArgs(ide, targetPath)
	default:
		return o.prepareGenericArgs(ide, targetPath)
	}
}

func (o *openOptions) prepareJetBrainsArgs(ide *IDE, targetPath string) []string {
	// JetBrains IDEs generally accept a project path as the last argument
	return []string{targetPath}
}

func (o *openOptions) prepareVSCodeArgs(ide *IDE, targetPath string) []string {
	// VS Code family accepts project paths with no special flags
	return []string{targetPath}
}

func (o *openOptions) prepareGenericArgs(ide *IDE, targetPath string) []string {
	// For generic editors, try common patterns
	switch {
	case strings.Contains(ide.Name, "Sublime"):
		return []string{"--project", targetPath}
	case strings.Contains(ide.Name, "Vim") || strings.Contains(ide.Name, "vim"):
		// For vim-like editors, open the directory or specific file
		return []string{targetPath}
	case strings.Contains(ide.Name, "Emacs"):
		return []string{targetPath}
	default:
		// Default: just pass the path
		return []string{targetPath}
	}
}

func (o *openOptions) getAvailableIDENames(ides []IDE) string {
	var names []string

	for _, ide := range ides {
		// Add main name
		names = append(names, ide.Name)

		// Add unique aliases
		for _, alias := range ide.Aliases {
			// Check if alias is not already in names
			found := false
			for _, existing := range names {
				if strings.EqualFold(existing, alias) {
					found = true
					break
				}
			}
			if !found {
				names = append(names, alias)
			}
		}
	}

	if len(names) == 0 {
		return "none found - run 'gz ide scan' first"
	}

	// Limit to first 10 to avoid overwhelming output
	if len(names) > 10 {
		names = names[:10]
		return strings.Join(names, ", ") + "... (and more)"
	}

	return strings.Join(names, ", ")
}

// shouldRunInBackground determines if IDE should run in background
func (o *openOptions) shouldRunInBackground(ide *IDE) bool {
	// Explicit flags take priority
	if o.background {
		return true
	}
	if o.wait {
		return false
	}

	// Default behavior: GUI apps run in background, CLI apps wait
	switch ide.Type {
	case "jetbrains", "vscode":
		return true // GUI apps
	case "other":
		// Check if it's a GUI or CLI editor
		switch {
		case strings.Contains(ide.Name, "Sublime"):
			return true // GUI
		case strings.Contains(ide.Name, "Vim"), strings.Contains(ide.Name, "vim"):
			return false // CLI
		case strings.Contains(ide.Name, "Emacs"):
			return false // Usually CLI (can be GUI but often used in terminal)
		default:
			return true // Default to background for unknown GUI apps
		}
	default:
		return true
	}
}

// executeInBackground starts IDE in background
func (o *openOptions) executeInBackground(cmd *exec.Cmd, ide *IDE) error {
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", ide.Name, err)
	}

	if o.verbose {
		fmt.Printf("✅ %s started successfully (PID: %d)\n", ide.Name, cmd.Process.Pid)
	} else {
		fmt.Printf("✅ %s opened successfully\n", ide.Name)
	}

	return nil
}

// executeAndWait runs IDE and waits for completion
func (o *openOptions) executeAndWait(cmd *exec.Cmd, ide *IDE) error {
	if o.verbose {
		fmt.Printf("⏳ Starting %s and waiting for exit...\n", ide.Name)
	}

	if err := cmd.Run(); err != nil {
		// Provide more specific error messages
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s exited with status %d", ide.Name, exitError.ExitCode())
		}
		return fmt.Errorf("failed to execute %s: %w", ide.Name, err)
	}

	if o.verbose {
		fmt.Printf("✅ %s exited successfully\n", ide.Name)
	}

	return nil
}
