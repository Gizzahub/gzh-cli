package ide

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewIDECmd(t *testing.T) {
	cmd := NewIDECmd()

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
	opts := defaultIDEOptions()

	// Test valid JetBrains product names
	validProducts := []string{
		"IntelliJIdea2023.2",
		"PyCharm2024.1",
		"WebStorm2023.3",
		"PhpStorm2024.2",
		"CLion2023.1",
		"GoLand2024.1",
		"DataGrip2023.3",
		"Rider2024.1",
		"AndroidStudio2023.2",
	}

	for _, product := range validProducts {
		assert.True(t, opts.isJetBrainsProduct(product), "Should recognize %s as JetBrains product", product)
	}

	// Test invalid product names
	invalidProducts := []string{
		"VSCode",
		"Eclipse",
		"SublimeText",
		"Atom",
		"Random2023.1",
	}

	for _, product := range invalidProducts {
		assert.False(t, opts.isJetBrainsProduct(product), "Should not recognize %s as JetBrains product", product)
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
	opts := defaultIDEOptions()

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.tmp")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o644)

	ignoredFiles := []string{
		filepath.Join(tmpDir, "file.tmp"),
		filepath.Join(tmpDir, "file~"),
		filepath.Join(tmpDir, "file.swp"),
		filepath.Join(tmpDir, ".DS_Store"),
		filepath.Join(tmpDir, "Thumbs.db"),
		filepath.Join(tmpDir, "file.lock"),
		filepath.Join(tmpDir, "file.log"),
		filepath.Join(tmpDir, "___jb_temp"),
	}

	for _, file := range ignoredFiles {
		event := fsnotify.Event{Name: file, Op: fsnotify.Write}
		assert.True(t, opts.shouldIgnoreEvent(event), "Should ignore %s", file)
	}

	// Test files that should not be ignored
	normalFiles := []string{
		filepath.Join(tmpDir, "config.xml"),
		filepath.Join(tmpDir, "settings.json"),
		filepath.Join(tmpDir, "filetypes.xml"),
	}

	for _, file := range normalFiles {
		event := fsnotify.Event{Name: file, Op: fsnotify.Write}
		assert.False(t, opts.shouldIgnoreEvent(event), "Should not ignore %s", file)
	}
}

func TestIsSyncProblematicFile(t *testing.T) {
	opts := defaultIDEOptions()

	problematicFiles := []string{
		"/path/to/filetypes.xml",
		"/path/to/settingsSync/options/filetypes.xml",
		"/path/to/workspace.xml",
	}

	for _, file := range problematicFiles {
		assert.True(t, opts.isSyncProblematicFile(file), "Should identify %s as problematic", file)
	}

	normalFiles := []string{
		"/path/to/colors.xml",
		"/path/to/keymap.xml",
		"/path/to/other.xml",
	}

	for _, file := range normalFiles {
		assert.False(t, opts.isSyncProblematicFile(file), "Should not identify %s as problematic", file)
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
	cmd := NewIDECmd()

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
	cmd := NewIDECmd()

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
