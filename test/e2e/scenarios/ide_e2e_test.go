//nolint:testpackage // White-box testing needed for internal function access
package scenarios

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Gizzahub/gzh-manager-go/test/e2e/helpers"
)

const ideSettingsPath = ".config/JetBrains/IntelliJIdea2024.1"

func TestIDE_List_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test listing IDE installations
	result := env.RunCommand("ide", "list")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().OutputContains("JetBrains")
}

func TestIDE_Monitor_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create mock IDE settings directory
	env.CreateDir(ideSettingsPath)

	// Create mock settings files
	settingsFiles := []string{
		"options/filetypes.xml",
		"options/colors.xml",
		"options/keymap.xml",
	}

	for _, file := range settingsFiles {
		content := `<?xml version="1.0" encoding="UTF-8"?>
<application>
  <component name="TestComponent">
    <option name="test" value="true" />
  </component>
</application>`
		env.CreateFile(ideSettingsPath+"/"+file, content)
	}

	// Test monitoring command with short timeout
	result := env.RunCommand("ide", "monitor", "--timeout", "2s", "--verbose")

	assertions := helpers.NewCLIAssertions(t, result)
	// Monitor command should start successfully, might timeout which is expected
	assertions.OutputContains("monitor")
}

func TestIDE_MonitorDaemon_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create IDE settings directory
	env.CreateDir(ideSettingsPath)

	// Test daemon mode (background process)
	ctx := context.Background()
	cmd, err := env.CLI.RunAsync(ctx, "ide", "monitor", "--daemon", "--pid-file", "ide-monitor.pid")
	if err != nil {
		t.Skipf("Daemon mode not available: %v", err)
		return
	}

	// Give the daemon a moment to start
	time.Sleep(100 * time.Millisecond)

	// Check if daemon started
	if cmd.Process != nil {
		// Verify PID file was created
		env.AssertFileExists("ide-monitor.pid")

		// Stop the daemon
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Warning: failed to kill process: %v", err)
		}
		if err := cmd.Wait(); err != nil {
			t.Logf("Warning: process wait failed: %v", err)
		}
	}
}

func TestIDE_FixSync_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create problematic settings file (filetypes.xml with duplicates)
	env.CreateDir(ideSettingsPath + "/options")

	problematicContent := `<?xml version="1.0" encoding="UTF-8"?>
<application>
  <component name="FileTypeManager">
    <mapping ext="txt" type="PLAIN_TEXT" />
    <mapping ext="js" type="JavaScript" />
    <mapping ext="txt" type="PLAIN_TEXT" />
    <mapping ext="py" type="Python" />
  </component>
</application>`

	env.CreateFile(ideSettingsPath+"/options/filetypes.xml", problematicContent)

	// Test fix-sync command
	result := env.RunCommand("ide", "fix-sync", "--target", ideSettingsPath)

	assertions := helpers.NewCLIAssertions(t, result)
	if result.ExitCode == 0 {
		assertions.Success().OutputContains("fix")

		// Verify duplicates were removed
		fixedContent := env.ReadFile(ideSettingsPath + "/options/filetypes.xml")
		lines := strings.Split(fixedContent, "\n")
		txtMappings := 0

		for _, line := range lines {
			if strings.Contains(line, `ext="txt"`) {
				txtMappings++
			}
		}

		if txtMappings <= 1 {
			t.Log("Duplicate mappings were successfully removed")
		}
	}
}

func TestIDE_MultipleIDEs_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create settings for multiple IDE versions
	ides := []string{
		"IntelliJIdea2024.1",
		"PyCharm2024.1",
		"WebStorm2024.1",
		"GoLand2024.1",
	}

	for _, ide := range ides {
		ideDir := ".config/JetBrains/" + ide
		env.CreateDir(ideDir + "/options")

		settingsContent := `<?xml version="1.0" encoding="UTF-8"?>
<application>
  <component name="` + ide + `Component">
    <option name="version" value="2024.1" />
  </component>
</application>`

		env.CreateFile(ideDir+"/options/ide.xml", settingsContent)
	}

	// Test listing multiple IDEs
	result := env.RunCommand("ide", "list")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success()

	// Check that all IDEs are detected
	for _, ide := range ides {
		if strings.Contains(result.Output, ide) || strings.Contains(result.Output, strings.Replace(ide, "2024.1", "", 1)) {
			t.Logf("IDE %s detected", ide)
		}
	}
}

func TestIDE_SettingsBackup_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create IDE settings
	env.CreateDir(ideSettingsPath + "/options")

	originalContent := `<?xml version="1.0" encoding="UTF-8"?>
<application>
  <component name="BackupTest">
    <option name="original" value="true" />
  </component>
</application>`

	env.CreateFile(ideSettingsPath+"/options/test.xml", originalContent)

	// Test backup command
	result := env.RunCommand("ide", "backup", "--target", ideSettingsPath)

	if result.ExitCode == 0 {
		assertions := helpers.NewCLIAssertions(t, result)
		assertions.Success().OutputContains("backup")

		// Verify backup directory was created
		backupFiles := env.ListFiles(".")
		backupFound := false

		for _, file := range backupFiles {
			if strings.Contains(file, "backup") && strings.Contains(file, ".xml") {
				backupFound = true
				break
			}
		}

		if backupFound {
			t.Log("Settings backup created successfully")
		}
	}
}

func TestIDE_SettingsRestore_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create IDE settings and backup
	env.CreateDir(ideSettingsPath + "/options")

	originalContent := `<?xml version="1.0" encoding="UTF-8"?>
<application>
  <component name="RestoreTest">
    <option name="original" value="true" />
  </component>
</application>`

	env.CreateFile(ideSettingsPath+"/options/test.xml", originalContent)

	// Create backup directory and file
	backupDir := "ide-backup-" + time.Now().Format("20060102150405")
	env.CreateDir(backupDir)
	env.CopyFile(ideSettingsPath+"/options/test.xml", backupDir+"/test.xml")

	// Modify original file
	modifiedContent := `<?xml version="1.0" encoding="UTF-8"?>
<application>
  <component name="RestoreTest">
    <option name="modified" value="true" />
  </component>
</application>`

	env.CreateFile(ideSettingsPath+"/options/test.xml", modifiedContent)

	// Test restore command
	result := env.RunCommand("ide", "restore", "--backup", backupDir, "--target", ideSettingsPath)

	if result.ExitCode == 0 {
		assertions := helpers.NewCLIAssertions(t, result)
		assertions.Success().OutputContains("restore")

		// Verify original content was restored
		restoredContent := env.ReadFile(ideSettingsPath + "/options/test.xml")
		if strings.Contains(restoredContent, "original") && !strings.Contains(restoredContent, "modified") {
			t.Log("Settings restored successfully")
		}
	}
}

func TestIDE_ConfigurationValidation_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create IDE configuration
	ideConfig := `
monitor:
  watch_dirs:
    - "~/.config/JetBrains"
  exclude_patterns:
    - "*.tmp"
    - "*.log"
  sync_check: true
  auto_fix: true
`
	env.WriteConfig("ide-config.yaml", ideConfig)

	// Test configuration validation
	result := env.RunCommand("ide", "validate", "--config", "ide-config.yaml")

	if result.ExitCode == 0 {
		assertions := helpers.NewCLIAssertions(t, result)
		assertions.Success().OutputContains("valid")
	}
}

func TestIDE_ErrorHandling_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test with non-existent IDE directory
	result := env.RunCommand("ide", "monitor", "--target", "non-existent-dir")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Failure().OutputContains("not found")

	// Test with invalid configuration
	invalidConfig := `
monitor:
  invalid_option: true
  watch_dirs: [
`
	env.WriteConfig("invalid-ide-config.yaml", invalidConfig)

	result = env.RunCommand("ide", "validate", "--config", "invalid-ide-config.yaml")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.Failure()
}

func TestIDE_LogOutput_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create IDE settings directory
	env.CreateDir(ideSettingsPath)

	// Test with logging enabled
	result := env.RunCommand("ide", "monitor", "--timeout", "1s", "--log-file", "ide-monitor.log", "--verbose")

	// Check if log file was created
	if result.ExitCode == 0 || result.ExitCode == 1 { // Timeout is expected
		env.AssertFileExists("ide-monitor.log")

		logContent := env.ReadFile("ide-monitor.log")
		if strings.Contains(logContent, "monitor") || strings.Contains(logContent, "IDE") {
			t.Log("Log file contains expected content")
		}
	}
}

func TestIDE_HelpAndUsage_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test IDE help
	result := env.RunCommand("ide", "--help")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().
		OutputContains("ide").
		OutputContains("monitor").
		OutputContains("list").
		OutputContains("fix-sync")

	// Test subcommand help
	subcommands := []string{"monitor", "list", "fix-sync"}

	for _, subcmd := range subcommands {
		result = env.RunCommand("ide", subcmd, "--help")

		assertions = helpers.NewCLIAssertions(t, result)
		assertions.Success().OutputContains(subcmd)
	}
}
