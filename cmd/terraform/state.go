package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// StateCmd represents the state command
var StateCmd = &cobra.Command{
	Use:   "state",
	Short: "Manage Terraform state files",
	Long: `Manage Terraform state files with advanced operations.

Provides comprehensive state management including:
- State inspection and analysis
- Resource import and removal
- State migration between backends
- State backup and restore
- State validation and repair
- Resource address management

Examples:
  gz terraform state list
  gz terraform state show aws_instance.example
  gz terraform state mv aws_instance.old aws_instance.new
  gz terraform state rm aws_instance.unused
  gz terraform state import aws_instance.example i-1234567890abcdef0`,
	Run: runStateList,
}

var (
	stateOperation string
	stateAddress   string
	stateNewAddr   string
	stateID        string
	stateBackup    string
	stateOutput    string
	stateFormat    string
	stateDryRun    bool
	stateForce     bool
)

func init() {
	StateCmd.AddCommand(StateListCmd)
	StateCmd.AddCommand(StateShowCmd)
	StateCmd.AddCommand(StateMvCmd)
	StateCmd.AddCommand(StateRmCmd)
	StateCmd.AddCommand(StateImportCmd)
	StateCmd.AddCommand(StateBackupCmd)
	StateCmd.AddCommand(StateRestoreCmd)
	StateCmd.AddCommand(StateMigrateCmd)
}

// StateListCmd lists resources in the state
var StateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources in the state",
	Long:  `List all resources currently tracked in the Terraform state.`,
	Run:   runStateList,
}

// StateShowCmd shows detailed information about a resource
var StateShowCmd = &cobra.Command{
	Use:   "show [address]",
	Short: "Show detailed information about a resource",
	Long:  `Show detailed information about a specific resource in the state.`,
	Run:   runStateShow,
}

// StateMvCmd moves a resource in the state
var StateMvCmd = &cobra.Command{
	Use:   "mv [source] [destination]",
	Short: "Move a resource in the state",
	Long:  `Move a resource to a new address in the state.`,
	Run:   runStateMv,
}

// StateRmCmd removes a resource from the state
var StateRmCmd = &cobra.Command{
	Use:   "rm [address]",
	Short: "Remove a resource from the state",
	Long:  `Remove a resource from the state without destroying it.`,
	Run:   runStateRm,
}

// StateImportCmd imports a resource into the state
var StateImportCmd = &cobra.Command{
	Use:   "import [address] [id]",
	Short: "Import a resource into the state",
	Long:  `Import an existing resource into the Terraform state.`,
	Run:   runStateImport,
}

// StateBackupCmd creates a backup of the state
var StateBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the state",
	Long:  `Create a backup of the current Terraform state.`,
	Run:   runStateBackup,
}

// StateRestoreCmd restores state from a backup
var StateRestoreCmd = &cobra.Command{
	Use:   "restore [backup-file]",
	Short: "Restore state from a backup",
	Long:  `Restore Terraform state from a backup file.`,
	Run:   runStateRestore,
}

// StateMigrateCmd migrates state between backends
var StateMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate state between backends",
	Long:  `Migrate Terraform state from one backend to another.`,
	Run:   runStateMigrate,
}

func init() {
	// State list flags
	StateListCmd.Flags().StringVar(&stateFormat, "format", "table", "Output format (table, json)")

	// State show flags
	StateShowCmd.Flags().StringVar(&stateFormat, "format", "human", "Output format (human, json)")

	// State mv flags
	StateMvCmd.Flags().BoolVar(&stateDryRun, "dry-run", false, "Show what would be moved")
	StateMvCmd.Flags().StringVar(&stateBackup, "backup", "", "Backup file path")

	// State rm flags
	StateRmCmd.Flags().BoolVar(&stateDryRun, "dry-run", false, "Show what would be removed")
	StateRmCmd.Flags().BoolVar(&stateForce, "force", false, "Force removal without confirmation")
	StateRmCmd.Flags().StringVar(&stateBackup, "backup", "", "Backup file path")

	// State import flags
	StateImportCmd.Flags().BoolVar(&stateDryRun, "dry-run", false, "Show what would be imported")
	StateImportCmd.Flags().StringVar(&stateBackup, "backup", "", "Backup file path")

	// State backup flags
	StateBackupCmd.Flags().StringVarP(&stateOutput, "output", "o", "", "Backup file path")

	// State migrate flags
	StateMigrateCmd.Flags().BoolVar(&stateDryRun, "dry-run", false, "Show migration plan")
	StateMigrateCmd.Flags().BoolVar(&stateForce, "force", false, "Force migration")
}

// StateInfo represents information about terraform state
type StateInfo struct {
	Version          int                    `json:"version"`
	TerraformVersion string                 `json:"terraform_version"`
	Serial           int                    `json:"serial"`
	Resources        []StateResource        `json:"resources"`
	Outputs          map[string]interface{} `json:"outputs"`
	Timestamp        time.Time              `json:"timestamp"`
}

type StateResource struct {
	Address   string          `json:"address"`
	Mode      string          `json:"mode"`
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Provider  string          `json:"provider"`
	Instances []StateInstance `json:"instances"`
}

type StateInstance struct {
	SchemaVersion int                    `json:"schema_version"`
	Attributes    map[string]interface{} `json:"attributes"`
	Dependencies  []string               `json:"dependencies"`
}

func runStateList(cmd *cobra.Command, args []string) {
	fmt.Printf("📋 Listing Terraform state resources\n")

	// Get state information
	stateInfo, err := getStateInfo()
	if err != nil {
		fmt.Printf("❌ Failed to get state info: %v\n", err)
		os.Exit(1)
	}

	// Output based on format
	switch stateFormat {
	case "json":
		outputStateJSON(stateInfo)
	default:
		outputStateTable(stateInfo)
	}
}

func runStateShow(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("❌ Resource address is required\n")
		cmd.Help()
		os.Exit(1)
	}

	address := args[0]
	fmt.Printf("🔍 Showing state for resource: %s\n", address)

	// Run terraform state show
	execCmd := exec.Command("terraform", "state", "show", address)

	if stateFormat == "json" {
		execCmd.Args = append(execCmd.Args, "-json")
	}

	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		fmt.Printf("❌ Failed to show resource: %v\n", err)
		os.Exit(1)
	}
}

func runStateMv(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("❌ Source and destination addresses are required\n")
		cmd.Help()
		os.Exit(1)
	}

	source := args[0]
	destination := args[1]

	fmt.Printf("🔄 Moving resource from %s to %s\n", source, destination)

	if stateDryRun {
		fmt.Printf("🧪 Dry run mode - would move:\n")
		fmt.Printf("  From: %s\n", source)
		fmt.Printf("  To:   %s\n", destination)
		return
	}

	// Create backup if requested
	if stateBackup != "" {
		if err := createStateBackupFile(stateBackup); err != nil {
			fmt.Printf("⚠️ Failed to create backup: %v\n", err)
		} else {
			fmt.Printf("💾 State backed up to: %s\n", stateBackup)
		}
	}

	// Execute move
	execCmd := exec.Command("terraform", "state", "mv", source, destination)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		fmt.Printf("❌ Failed to move resource: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Resource moved successfully\n")
}

func runStateRm(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("❌ Resource address is required\n")
		cmd.Help()
		os.Exit(1)
	}

	address := args[0]

	fmt.Printf("🗑️ Removing resource from state: %s\n", address)

	if stateDryRun {
		fmt.Printf("🧪 Dry run mode - would remove: %s\n", address)
		return
	}

	// Confirm removal if not forced
	if !stateForce {
		if !confirmRemoval(address) {
			fmt.Printf("❌ Removal cancelled\n")
			os.Exit(1)
		}
	}

	// Create backup if requested
	if stateBackup != "" {
		if err := createStateBackupFile(stateBackup); err != nil {
			fmt.Printf("⚠️ Failed to create backup: %v\n", err)
		} else {
			fmt.Printf("💾 State backed up to: %s\n", stateBackup)
		}
	}

	// Execute removal
	execCmd := exec.Command("terraform", "state", "rm", address)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		fmt.Printf("❌ Failed to remove resource: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Resource removed from state successfully\n")
	fmt.Printf("⚠️ Note: The actual resource still exists in your cloud provider\n")
}

func runStateImport(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("❌ Resource address and ID are required\n")
		cmd.Help()
		os.Exit(1)
	}

	address := args[0]
	resourceID := args[1]

	fmt.Printf("📥 Importing resource: %s (ID: %s)\n", address, resourceID)

	if stateDryRun {
		fmt.Printf("🧪 Dry run mode - would import:\n")
		fmt.Printf("  Address: %s\n", address)
		fmt.Printf("  ID:      %s\n", resourceID)
		return
	}

	// Create backup if requested
	if stateBackup != "" {
		if err := createStateBackupFile(stateBackup); err != nil {
			fmt.Printf("⚠️ Failed to create backup: %v\n", err)
		} else {
			fmt.Printf("💾 State backed up to: %s\n", stateBackup)
		}
	}

	// Execute import
	execCmd := exec.Command("terraform", "import", address, resourceID)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		fmt.Printf("❌ Failed to import resource: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Resource imported successfully\n")
	fmt.Printf("📝 Next steps:\n")
	fmt.Printf("1. Add corresponding resource configuration to your .tf files\n")
	fmt.Printf("2. Run 'terraform plan' to verify configuration matches imported state\n")
	fmt.Printf("3. Adjust configuration as needed to eliminate plan differences\n")
}

func runStateBackup(cmd *cobra.Command, args []string) {
	// Generate backup filename if not provided
	backupFile := stateOutput
	if backupFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		backupFile = fmt.Sprintf("terraform.tfstate.backup.%s", timestamp)
	}

	fmt.Printf("💾 Creating state backup: %s\n", backupFile)

	if err := createStateBackupFile(backupFile); err != nil {
		fmt.Printf("❌ Failed to create backup: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ State backup created successfully\n")
	fmt.Printf("📁 Backup location: %s\n", backupFile)
}

func runStateRestore(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("❌ Backup file path is required\n")
		cmd.Help()
		os.Exit(1)
	}

	backupFile := args[0]

	fmt.Printf("🔄 Restoring state from backup: %s\n", backupFile)

	// Verify backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		fmt.Printf("❌ Backup file not found: %s\n", backupFile)
		os.Exit(1)
	}

	// Create backup of current state before restore
	currentBackup := fmt.Sprintf("terraform.tfstate.before-restore.%s", time.Now().Format("20060102-150405"))
	if err := createStateBackupFile(currentBackup); err != nil {
		fmt.Printf("⚠️ Failed to backup current state: %v\n", err)
	} else {
		fmt.Printf("💾 Current state backed up to: %s\n", currentBackup)
	}

	// Restore state
	if err := restoreStateFromFile(backupFile); err != nil {
		fmt.Printf("❌ Failed to restore state: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ State restored successfully\n")
	fmt.Printf("📝 Verify the restored state with 'terraform plan'\n")
}

func runStateMigrate(cmd *cobra.Command, args []string) {
	fmt.Printf("🚚 Starting state migration\n")

	if stateDryRun {
		fmt.Printf("🧪 Dry run mode - showing migration plan\n")
		return showMigrationPlan()
	}

	// Confirm migration if not forced
	if !stateForce {
		if !confirmMigration() {
			fmt.Printf("❌ Migration cancelled\n")
			os.Exit(1)
		}
	}

	// Create backup before migration
	backupFile := fmt.Sprintf("terraform.tfstate.pre-migration.%s", time.Now().Format("20060102-150405"))
	if err := createStateBackupFile(backupFile); err != nil {
		fmt.Printf("⚠️ Failed to create pre-migration backup: %v\n", err)
	} else {
		fmt.Printf("💾 Pre-migration backup created: %s\n", backupFile)
	}

	// Execute migration
	if err := executeMigration(); err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		fmt.Printf("🔄 Attempting to restore from backup: %s\n", backupFile)
		if restoreErr := restoreStateFromFile(backupFile); restoreErr != nil {
			fmt.Printf("❌ Failed to restore backup: %v\n", restoreErr)
		}
		os.Exit(1)
	}

	fmt.Printf("✅ State migration completed successfully\n")
}

func getStateInfo() (*StateInfo, error) {
	// Get state as JSON
	cmd := exec.Command("terraform", "show", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	var stateData map[string]interface{}
	if err := json.Unmarshal(output, &stateData); err != nil {
		return nil, fmt.Errorf("failed to parse state JSON: %w", err)
	}

	// Extract relevant information
	info := &StateInfo{
		Timestamp: time.Now(),
		Resources: []StateResource{},
	}

	// Extract basic info
	if values, ok := stateData["values"].(map[string]interface{}); ok {
		if rootModule, ok := values["root_module"].(map[string]interface{}); ok {
			if resources, ok := rootModule["resources"].([]interface{}); ok {
				for _, resource := range resources {
					if res, ok := resource.(map[string]interface{}); ok {
						stateRes := StateResource{
							Address: getStringValue(res, "address"),
							Mode:    getStringValue(res, "mode"),
							Type:    getStringValue(res, "type"),
							Name:    getStringValue(res, "name"),
						}
						info.Resources = append(info.Resources, stateRes)
					}
				}
			}
		}
	}

	return info, nil
}

func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func outputStateJSON(info *StateInfo) {
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Printf("❌ Failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

func outputStateTable(info *StateInfo) {
	if len(info.Resources) == 0 {
		fmt.Printf("📭 No resources found in state\n")
		return
	}

	fmt.Printf("\n📊 State Resources:\n")
	fmt.Printf("==================\n")
	fmt.Printf("%-60s %-15s %-20s\n", "Address", "Mode", "Type")
	fmt.Printf("%s\n", strings.Repeat("-", 95))

	for _, resource := range info.Resources {
		fmt.Printf("%-60s %-15s %-20s\n",
			truncateString(resource.Address, 60),
			resource.Mode,
			resource.Type)
	}

	fmt.Printf("\n📈 Summary: %d resources in state\n", len(info.Resources))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func confirmRemoval(address string) bool {
	fmt.Printf("⚠️ Are you sure you want to remove '%s' from state?\n", address)
	fmt.Printf("This will NOT destroy the actual resource.\n")
	fmt.Printf("Type 'yes' to confirm: ")

	var response string
	fmt.Scanln(&response)

	return strings.ToLower(response) == "yes"
}

func confirmMigration() bool {
	fmt.Printf("⚠️ Are you sure you want to migrate the state?\n")
	fmt.Printf("This will move your state to a new backend.\n")
	fmt.Printf("Type 'yes' to confirm: ")

	var response string
	fmt.Scanln(&response)

	return strings.ToLower(response) == "yes"
}

func createStateBackupFile(filename string) error {
	// Create directory if needed
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Get current state
	cmd := exec.Command("terraform", "state", "pull")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Write to backup file
	return os.WriteFile(filename, output, 0o644)
}

func restoreStateFromFile(filename string) error {
	// Read backup file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Push state
	cmd := exec.Command("terraform", "state", "push", filename)
	cmd.Stdin = strings.NewReader(string(data))
	return cmd.Run()
}

func showMigrationPlan() error {
	fmt.Printf("📋 Migration Plan:\n")
	fmt.Printf("=================\n")
	fmt.Printf("1. Current backend configuration will be analyzed\n")
	fmt.Printf("2. New backend configuration will be validated\n")
	fmt.Printf("3. State will be copied to new backend\n")
	fmt.Printf("4. Old state will remain as backup\n")
	fmt.Printf("5. Terraform will be reconfigured for new backend\n")

	// Show current backend info
	if backendInfo, err := getCurrentBackendInfo(); err == nil {
		fmt.Printf("\nCurrent Backend: %s\n", backendInfo)
	}

	return nil
}

func getCurrentBackendInfo() (string, error) {
	// Try to determine current backend from terraform init output
	cmd := exec.Command("terraform", "init", "-backend=false")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown", err
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "s3") {
		return "S3", nil
	} else if strings.Contains(outputStr, "gcs") {
		return "Google Cloud Storage", nil
	} else if strings.Contains(outputStr, "azurerm") {
		return "Azure Storage", nil
	}

	return "local", nil
}

func executeMigration() error {
	// This would implement the actual migration logic
	// For now, this is a placeholder that runs terraform init with migration

	fmt.Printf("🔄 Executing migration...\n")

	cmd := exec.Command("terraform", "init", "-migrate-state")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
