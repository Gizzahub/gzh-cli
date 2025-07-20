//nolint:testpackage // White-box testing needed for internal function access
package ide

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewIDECmd(t *testing.T) {
	cmd := NewIDECmd(context.Background())

	assert.Equal(t, "ide", cmd.Use)
	assert.Equal(t, "Monitor and manage IDE configuration changes", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	// Verify subcommands exist
	var monitorCmd, listCmd, fixSyncCmd *cobra.Command

	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "monitor":
			monitorCmd = subcmd
		case "list":
			listCmd = subcmd
		case "fix-sync":
			fixSyncCmd = subcmd
		}
	}

	assert.NotNil(t, monitorCmd)
	assert.Equal(t, "monitor", monitorCmd.Use)
	assert.Equal(t, "Monitor JetBrains settings for changes", monitorCmd.Short)

	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
	assert.Equal(t, "List detected JetBrains IDE installations", listCmd.Short)

	assert.NotNil(t, fixSyncCmd)
	assert.Equal(t, "fix-sync", fixSyncCmd.Use)
	assert.Equal(t, "Fix JetBrains settings synchronization issues", fixSyncCmd.Short)
}

func TestDefaultIDEOptions(t *testing.T) {
	opts := defaultIDEOptions()

	assert.True(t, opts.recursive)
	assert.False(t, opts.verbose)
	assert.False(t, opts.daemon)
	assert.False(t, opts.fixSync)
	assert.NotEmpty(t, opts.logPath)
	assert.Contains(t, opts.excludePaths, ".git")
	assert.Contains(t, opts.excludePaths, "node_modules")
	assert.Contains(t, opts.excludePaths, ".idea/shelf")
}

func TestGetJetBrainsBasePaths(t *testing.T) {
	opts := defaultIDEOptions()
	paths := opts.getJetBrainsBasePaths()

	switch runtime.GOOS {
	case "linux":
		assert.Len(t, paths, 1)
		assert.Contains(t, paths[0], ".config/JetBrains")
	case "darwin":
		assert.Len(t, paths, 1)
		assert.Contains(t, paths[0], "Library/Application Support/JetBrains")
	case "windows":
		assert.Len(t, paths, 1)
		assert.Contains(t, paths[0], "JetBrains")
	default:
		assert.Len(t, paths, 0)
	}
}

func TestIsJetBrainsProduct(t *testing.T) {
	tests := []struct {
		name     string
		product  string
		expected bool
	}{
		{"IntelliJ IDEA", "IntelliJIdea2023.2", true},
		{"PyCharm", "PyCharm2024.1", true},
		{"WebStorm", "WebStorm2023.3", true},
		{"PhpStorm", "PhpStorm2024.2", true},
		{"CLion", "CLion2023.1", true},
		{"GoLand", "GoLand2024.1", true},
		{"DataGrip", "DataGrip2023.3", true},
		{"Rider", "Rider2024.1", true},
		{"AndroidStudio", "AndroidStudio2023.2", true},
		{"VSCode", "VSCode", false},
		{"Eclipse", "Eclipse", false},
		{"SublimeText", "SublimeText", false},
		{"Atom", "Atom", false},
		{"Random product", "Random2023.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultIDEOptions()
			result := opts.isJetBrainsProduct(tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatProductName(t *testing.T) {
	opts := defaultIDEOptions()

	testCases := []struct {
		input    string
		expected string
	}{
		{"IntelliJIdea2023.2", "IntelliJIdea 2023.2"},
		{"PyCharm2024.1", "PyCharm 2024.1"},
		{"WebStorm2023.3.1", "WebStorm 2023.3.1"},
		{"ProductNameOnly", "ProductNameOnly"},
		{"GoLand2024.1.1", "GoLand 2024.1.1"},
	}

	for _, tc := range testCases {
		result := opts.formatProductName(tc.input)
		assert.Equal(t, tc.expected, result, "Format product name for %s", tc.input)
	}
}

func TestShouldIgnoreEvent(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		filename     string
		shouldIgnore bool
	}{
		{"temp file", "file.tmp", true},
		{"backup file", "file~", true},
		{"swap file", "file.swp", true},
		{"macOS DS_Store", ".DS_Store", true},
		{"Windows thumbs", "Thumbs.db", true},
		{"lock file", "file.lock", true},
		{"log file", "file.log", true},
		{"JetBrains temp", "___jb_temp", true},
		{"config XML", "config.xml", false},
		{"settings JSON", "settings.json", false},
		{"filetypes XML", "filetypes.xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultIDEOptions()
			file := filepath.Join(tmpDir, tt.filename)
			event := fsnotify.Event{Name: file, Op: fsnotify.Write}
			result := opts.shouldIgnoreEvent(event)
			assert.Equal(t, tt.shouldIgnore, result)
		})
	}
}

func TestIsSyncProblematicFile(t *testing.T) {
	tests := []struct {
		name          string
		filepath      string
		isProblematic bool
	}{
		{"filetypes XML", "/path/to/filetypes.xml", true},
		{"sync filetypes XML", "/path/to/settingsSync/options/filetypes.xml", true},
		{"workspace XML", "/path/to/workspace.xml", true},
		{"colors XML", "/path/to/colors.xml", false},
		{"keymap XML", "/path/to/keymap.xml", false},
		{"other XML", "/path/to/other.xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultIDEOptions()
			result := opts.isSyncProblematicFile(tt.filepath)
			assert.Equal(t, tt.isProblematic, result)
		})
	}
}

func TestApplyFiletypesXMLFixes(t *testing.T) {
	opts := defaultIDEOptions()

	// Test content with duplicate lines
	content := `<component name="FileTypeManager">
  <mapping ext="txt" type="PLAIN_TEXT" />
  <mapping ext="js" type="JavaScript" />
  <mapping ext="txt" type="PLAIN_TEXT" />
  <mapping ext="py" type="Python" />
</component>`

	expected := `<component name="FileTypeManager">
  <mapping ext="txt" type="PLAIN_TEXT" />
  <mapping ext="js" type="JavaScript" />
  <mapping ext="py" type="Python" />
</component>`

	result := opts.applyFiletypesXMLFixes(content)
	assert.Equal(t, expected, result)
}

func TestGetRelativePath(t *testing.T) {
	opts := defaultIDEOptions()
	homeDir, _ := os.UserHomeDir()

	testCases := []struct {
		input    string
		expected string
	}{
		{filepath.Join(homeDir, "test", "file.xml"), "~/test/file.xml"},
		{"/etc/hosts", "/etc/hosts"}, // Outside home, should return full path
	}

	for _, tc := range testCases {
		result := opts.getRelativePath(tc.input)
		assert.Equal(t, tc.expected, result, "Relative path for %s", tc.input)
	}
}

func TestFormatSize(t *testing.T) {
	opts := defaultIDEOptions()

	testCases := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range testCases {
		result := opts.formatSize(tc.bytes)
		assert.Equal(t, tc.expected, result, "Format size for %d bytes", tc.bytes)
	}
}

func TestCopyFile(t *testing.T) {
	opts := defaultIDEOptions()
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")
	content := "test content"

	// Write source file
	err := os.WriteFile(srcFile, []byte(content), 0o644)
	assert.NoError(t, err)

	// Test copy
	err = opts.copyFile(srcFile, dstFile)
	assert.NoError(t, err)

	// Verify copy
	dstContent, err := os.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, content, string(dstContent))
}

func TestIDECmdStructure(t *testing.T) {
	cmd := NewIDECmd(context.Background())

	// Test that the command has proper structure
	assert.NotNil(t, cmd.Use)
	assert.NotNil(t, cmd.Short)
	assert.NotNil(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Test that examples are included in Long description
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "gz ide monitor")
	assert.Contains(t, cmd.Long, "gz ide list")
	assert.Contains(t, cmd.Long, "gz ide fix-sync")
}

func TestIDECmdHelpContent(t *testing.T) {
	cmd := NewIDECmd(context.Background())

	// Verify help content mentions key features
	longDesc := cmd.Long
	assert.Contains(t, longDesc, "Monitor and manage IDE configuration changes")
	assert.Contains(t, longDesc, "JetBrains products")
	assert.Contains(t, longDesc, "Real-time monitoring")
	assert.Contains(t, longDesc, "Cross-platform support")
	assert.Contains(t, longDesc, "Settings synchronization")

	// Verify supported IDEs are mentioned
	assert.Contains(t, longDesc, "IntelliJ IDEA")
	assert.Contains(t, longDesc, "PyCharm")
	assert.Contains(t, longDesc, "WebStorm")
	assert.Contains(t, longDesc, "GoLand")
}
