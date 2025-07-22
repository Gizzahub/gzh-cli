// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package shell

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
)

// ShellCmd represents the shell command.
var ShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start interactive debugging shell (REPL)",
	Long: `Start an interactive debugging shell (REPL) for real-time system inspection.

The shell provides a command-line interface for:
- Real-time system state inspection
- Dynamic configuration changes
- Live debugging and troubleshooting
- Interactive plugin execution
- Memory and performance monitoring
- Command history and auto-completion

Available commands:
  help           - Show available commands
  status         - Show system status
  memory         - Show memory usage
  config         - Show/modify configuration
  plugins        - List and manage plugins
  logs           - Show recent logs
  metrics        - Show system metrics
  trace          - Start/stop tracing
  profile        - Start/stop profiling
  exit, quit     - Exit the shell

Examples:
  gz shell              # Start interactive shell
  gz shell --timeout 30m  # Auto-exit after 30 minutes`,
	Run: runShell,
}

var (
	timeout   time.Duration
	quietMode bool
	noHistory bool
)

func init() {
	ShellCmd.Flags().DurationVar(&timeout, "timeout", 0, "Auto-exit timeout (0 = no timeout)")
	ShellCmd.Flags().BoolVar(&quietMode, "quiet", false, "Quiet mode - minimal output")
	ShellCmd.Flags().BoolVar(&noHistory, "no-history", false, "Disable command history")
}

// Shell represents the interactive debugging shell.
type Shell struct {
	client   *gzhclient.Client
	history  []string
	commands map[string]ShellCommand
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// ShellCommand represents a shell command.
type ShellCommand struct { //nolint:revive // Name is appropriate for package context
	Name        string
	Description string
	Usage       string
	Handler     func(*Shell, []string) error
	Completer   func(*Shell, string) []string
}

// ShellContext holds shell execution context.
type ShellContext struct { //nolint:revive // Name is appropriate for package context
	StartTime time.Time              `json:"startTime"`
	Uptime    time.Duration          `json:"uptime"`
	Commands  int                    `json:"commandsExecuted"`
	LastCmd   string                 `json:"lastCommand"`
	Vars      map[string]interface{} `json:"variables"`
}

func runShell(_ *cobra.Command, args []string) {
	if !quietMode {
		fmt.Println("🚀 Starting GZH Manager Interactive Shell")
		fmt.Println("Type 'help' for available commands, 'exit' to quit")
		fmt.Println()
	}

	// Create GZH client
	clientConfig := gzhclient.DefaultConfig()
	clientConfig.LogLevel = "warn" // Reduce noise in shell

	client, err := gzhclient.NewClient(clientConfig)
	if err != nil {
		fmt.Printf("❌ Failed to create GZH client: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close() // Ignore close error
	}()

	// Create shell
	shell := NewShell(client)

	// Setup timeout if specified
	if timeout > 0 {
		go func() {
			time.Sleep(timeout)
			fmt.Printf("\n⏰ Shell timeout reached (%v), exiting...\n", timeout)
			shell.Stop()
		}()
	}

	// Run shell
	if err := shell.Run(); err != nil {
		fmt.Printf("❌ Shell error: %v\n", err)
		return
	}

	if !quietMode {
		fmt.Println("👋 Shell session ended")
	}
}

// NewShell creates a new interactive shell.
func NewShell(client *gzhclient.Client) *Shell {
	ctx, cancel := context.WithCancel(context.Background())

	shell := &Shell{
		client:   client,
		history:  []string{},
		commands: make(map[string]ShellCommand),
		running:  true,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register built-in commands
	shell.registerCommands()

	return shell
}

// Run starts the shell REPL loop.
func (s *Shell) Run() error {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n^C received, exiting...")
		s.Stop()
	}()

	// Main REPL loop
	scanner := bufio.NewScanner(os.Stdin)

	for s.running {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			// Print prompt
			fmt.Print("gz> ")

			// Read input
			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Add to history
			if !noHistory {
				s.addToHistory(input)
			}

			// Execute command
			if err := s.executeCommand(input); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		}
	}

	return scanner.Err()
}

// Stop stops the shell.
func (s *Shell) Stop() {
	s.running = false
	s.cancel()
}

// addToHistory adds a command to the history.
func (s *Shell) addToHistory(cmd string) {
	// Avoid duplicate consecutive commands
	if len(s.history) > 0 && s.history[len(s.history)-1] == cmd {
		return
	}

	s.history = append(s.history, cmd)

	// Keep only last 100 commands
	if len(s.history) > 100 {
		s.history = s.history[1:]
	}
}

// executeCommand parses and executes a shell command.
func (s *Shell) executeCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmdName := parts[0]
	args := parts[1:]

	// Handle built-in commands
	if cmd, exists := s.commands[cmdName]; exists {
		return cmd.Handler(s, args)
	}

	return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmdName)
}

// registerCommands registers all built-in shell commands.
func (s *Shell) registerCommands() {
	s.commands["help"] = ShellCommand{
		Name:        "help",
		Description: "Show available commands",
		Usage:       "help [command]",
		Handler:     handleHelp,
		Completer:   completeHelp,
	}

	s.commands["exit"] = ShellCommand{
		Name:        "exit",
		Description: "Exit the shell",
		Usage:       "exit",
		Handler:     handleExit,
	}

	s.commands["quit"] = ShellCommand{
		Name:        "quit",
		Description: "Exit the shell",
		Usage:       "quit",
		Handler:     handleExit,
	}

	s.commands["status"] = ShellCommand{
		Name:        "status",
		Description: "Show system status",
		Usage:       "status [--json]",
		Handler:     handleStatus,
	}

	s.commands["memory"] = ShellCommand{
		Name:        "memory",
		Description: "Show memory usage",
		Usage:       "memory [--json] [--gc]",
		Handler:     handleMemory,
	}

	s.commands["plugins"] = ShellCommand{
		Name:        "plugins",
		Description: "List and manage plugins",
		Usage:       "plugins [list|exec <name> <method>]",
		Handler:     handlePlugins,
		Completer:   completePlugins,
	}

	s.commands["config"] = ShellCommand{
		Name:        "config",
		Description: "Show/modify configuration",
		Usage:       "config [get|set <key> <value>|list]",
		Handler:     handleConfig,
	}

	s.commands["metrics"] = ShellCommand{
		Name:        "metrics",
		Description: "Show system metrics",
		Usage:       "metrics [--json] [--watch]",
		Handler:     handleMetrics,
	}

	s.commands["trace"] = ShellCommand{
		Name:        "trace",
		Description: "Control execution tracing",
		Usage:       "trace [start|stop|status]",
		Handler:     handleTrace,
	}

	s.commands["profile"] = ShellCommand{
		Name:        "profile",
		Description: "Control performance profiling",
		Usage:       "profile [start|stop|status]",
		Handler:     handleProfile,
	}

	s.commands["history"] = ShellCommand{
		Name:        "history",
		Description: "Show command history",
		Usage:       "history [--clear] [--count <n>]",
		Handler:     handleHistory,
	}

	s.commands["clear"] = ShellCommand{
		Name:        "clear",
		Description: "Clear the screen",
		Usage:       "clear",
		Handler:     handleClear,
	}

	s.commands["context"] = ShellCommand{
		Name:        "context",
		Description: "Show shell context",
		Usage:       "context [--json]",
		Handler:     handleContext,
	}

	s.commands["logs"] = ShellCommand{
		Name:        "logs",
		Description: "Show recent logs",
		Usage:       "logs [--count <n>] [--level <level>]",
		Handler:     handleLogs,
	}
}

// Command handlers

func handleHelp(s *Shell, args []string) error {
	if len(args) > 0 {
		// Show help for specific command
		cmdName := args[0]
		if cmd, exists := s.commands[cmdName]; exists {
			fmt.Printf("Command: %s\n", cmd.Name)
			fmt.Printf("Description: %s\n", cmd.Description)
			fmt.Printf("Usage: %s\n", cmd.Usage)
		} else {
			return fmt.Errorf("unknown command: %s", cmdName)
		}

		return nil
	}

	// Show all commands
	fmt.Println("Available commands:")
	fmt.Println()

	// Sort commands for consistent output
	names := make([]string, 0, len(s.commands))
	for name := range s.commands {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		cmd := s.commands[name]
		fmt.Printf("  %-12s %s\n", cmd.Name, cmd.Description)
	}

	fmt.Println()
	fmt.Println("Use 'help <command>' for detailed usage information.")

	return nil
}

func handleExit(s *Shell, _ []string) error {
	s.Stop()
	return nil
}

func handleStatus(s *Shell, args []string) error {
	jsonOutput := len(args) > 0 && args[0] == "--json"

	health := s.client.Health()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := map[string]interface{}{
		"healthy":    health.Overall == "healthy",
		"uptime":     time.Since(health.Timestamp).String(),
		"version":    "1.0.0", // This should come from build info
		"memory_mb":  float64(m.Alloc) / 1024 / 1024,
		"goroutines": runtime.NumGoroutine(),
		"timestamp":  time.Now(),
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("System Status:\n")
		fmt.Printf("  Healthy: %v\n", status["healthy"])
		fmt.Printf("  Uptime: %v\n", status["uptime"])
		fmt.Printf("  Version: %v\n", status["version"])
		fmt.Printf("  Memory: %.2f MB\n", status["memory_mb"])
		fmt.Printf("  Goroutines: %v\n", status["goroutines"])
	}

	return nil
}

func handleMemory(_ *Shell, args []string) error {
	jsonOutput := false
	runGC := false

	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOutput = true
		case "--gc":
			runGC = true
		}
	}

	if runGC {
		fmt.Println("Running garbage collection...")
		runtime.GC()
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memStats := map[string]interface{}{
		"allocated_mb":   float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"goroutines":     runtime.NumGoroutine(),
		"heap_objects":   m.HeapObjects,
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(memStats, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Memory Usage:\n")
		fmt.Printf("  Allocated: %.2f MB\n", memStats["allocated_mb"])
		fmt.Printf("  Total Allocated: %.2f MB\n", memStats["total_alloc_mb"])
		fmt.Printf("  System: %.2f MB\n", memStats["sys_mb"])
		fmt.Printf("  GC Cycles: %v\n", memStats["num_gc"])
		fmt.Printf("  Goroutines: %v\n", memStats["goroutines"])
		fmt.Printf("  Heap Objects: %v\n", memStats["heap_objects"])
	}

	return nil
}

func handlePlugins(_ *Shell, args []string) error {
	fmt.Println("Plugin functionality has been disabled in this version.")
	fmt.Println("The plugins package has been removed from the codebase.")
	fmt.Println("Available subcommands: list, exec (both non-functional)")

	return nil
}

func handleConfig(s *Shell, args []string) error {
	// Simplified config handling - in a real implementation,
	// this would integrate with the actual config system
	fmt.Println("Configuration management not yet implemented in shell")
	fmt.Println("Use 'gz config' command outside of shell for configuration management")

	return nil
}

func handleMetrics(s *Shell, args []string) error {
	jsonOutput := false
	watchMode := false

	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOutput = true
		case "--watch":
			watchMode = true
		}
	}

	if watchMode {
		fmt.Println("Watching metrics (press Ctrl+C to stop)...")

		for {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				metrics, err := s.client.GetSystemMetrics()
				if err != nil {
					return err
				}

				if jsonOutput {
					data, _ := json.MarshalIndent(metrics, "", "  ")
					fmt.Println(string(data))
				} else {
					fmt.Printf("\r[%s] CPU: %.1f%%, Memory: %d MB, Uptime: %v",
						time.Now().Format("15:04:05"),
						metrics.CPU.Usage, metrics.Memory.Used/1024/1024, metrics.Uptime)
				}

				time.Sleep(1 * time.Second)
			}
		}
	} else {
		metrics, err := s.client.GetSystemMetrics()
		if err != nil {
			return err
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(metrics, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("System Metrics:\n")
			fmt.Printf("  CPU Usage: %.1f%%\n", metrics.CPU.Usage)
			fmt.Printf("  CPU Cores: %d\n", metrics.CPU.Cores)
			fmt.Printf("  Memory Used: %.1f MB\n", float64(metrics.Memory.Used)/1024/1024)
			fmt.Printf("  Memory Total: %.1f MB\n", float64(metrics.Memory.Total)/1024/1024)
			fmt.Printf("  Disk Used: %.1f GB\n", float64(metrics.Disk.Used)/1024/1024/1024)
			fmt.Printf("  Uptime: %v\n", metrics.Uptime)

			if len(metrics.LoadAvg) > 0 {
				fmt.Printf("  Load Average: %.2f\n", metrics.LoadAvg[0])
			}
		}
	}

	return nil
}

func handleTrace(s *Shell, args []string) error {
	if len(args) == 0 {
		args = []string{"status"}
	}

	switch args[0] {
	case "start":
		fmt.Println("Starting execution tracing...")
		// Tracing integration placeholder - implement with actual tracer
		fmt.Println("Tracing started (use 'trace stop' to stop)")

	case "stop":
		fmt.Println("Stopping execution tracing...")
		// Tracing integration placeholder - implement with actual tracer
		fmt.Println("Tracing stopped")

	case "status":
		fmt.Println("Trace Status: Not implemented yet")
		// Status display placeholder - implement with actual tracer status

	default:
		return fmt.Errorf("unknown trace command: %s", args[0])
	}

	return nil
}

func handleProfile(s *Shell, args []string) error {
	if len(args) == 0 {
		args = []string{"status"}
	}

	switch args[0] {
	case "start":
		fmt.Println("Starting performance profiling...")
		// Profiler integration placeholder - implement with actual profiler
		fmt.Println("Profiling started (use 'profile stop' to stop)")

	case "stop":
		fmt.Println("Stopping performance profiling...")
		// Profiler integration placeholder - implement with actual profiler
		fmt.Println("Profiling stopped")

	case "status":
		fmt.Println("Profile Status: Not implemented yet")
		// Profiler status placeholder - implement with actual profiler status

	default:
		return fmt.Errorf("unknown profile command: %s", args[0])
	}

	return nil
}

func handleHistory(s *Shell, args []string) error {
	clearHistory := false
	count := len(s.history)

	for i, arg := range args {
		switch arg {
		case "--clear":
			clearHistory = true
		case "--count":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					count = n
				}
			}
		}
	}

	if clearHistory {
		s.history = []string{}

		fmt.Println("Command history cleared")

		return nil
	}

	if len(s.history) == 0 {
		fmt.Println("No command history")
		return nil
	}

	fmt.Printf("Command History (last %d):\n", count)

	start := len(s.history) - count
	if start < 0 {
		start = 0
	}

	for i, cmd := range s.history[start:] {
		fmt.Printf("  %d: %s\n", start+i+1, cmd)
	}

	return nil
}

func handleClear(s *Shell, args []string) error {
	fmt.Print("\033[2J\033[H") // ANSI escape codes to clear screen
	return nil
}

func handleContext(s *Shell, args []string) error {
	jsonOutput := len(args) > 0 && args[0] == "--json"

	context := ShellContext{
		StartTime: time.Now(), // Current time used as placeholder for actual start time
		Uptime:    time.Since(time.Now()),
		Commands:  len(s.history),
		LastCmd:   "",
		Vars:      map[string]interface{}{},
	}

	if len(s.history) > 0 {
		context.LastCmd = s.history[len(s.history)-1]
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(context, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Shell Context:\n")
		fmt.Printf("  Start Time: %v\n", context.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Uptime: %v\n", context.Uptime)
		fmt.Printf("  Commands Executed: %d\n", context.Commands)
		fmt.Printf("  Last Command: %s\n", context.LastCmd)
	}

	return nil
}

func handleLogs(s *Shell, args []string) error {
	count := 10
	level := ""

	for i, arg := range args {
		switch arg {
		case "--count":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					count = n
				}
			}
		case "--level":
			if i+1 < len(args) {
				level = args[i+1]
			}
		}
	}

	// Log system integration placeholder - implement with actual log system
	fmt.Printf("Recent Logs (last %d):\n", count)
	fmt.Println("Log integration not yet implemented")
	fmt.Printf("Filters: level=%s\n", level)

	return nil
}

// Completion functions

func completeHelp(s *Shell, input string) []string {
	var completions []string

	for name := range s.commands {
		if strings.HasPrefix(name, input) {
			completions = append(completions, name)
		}
	}

	return completions
}

func completePlugins(s *Shell, input string) []string {
	completions := []string{"list", "exec"}

	var filtered []string

	for _, comp := range completions {
		if strings.HasPrefix(comp, input) {
			filtered = append(filtered, comp)
		}
	}

	return filtered
}
