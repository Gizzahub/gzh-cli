// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// 저장 경로의 실패 처리는 외부 API를 늘리지 않고 단위 테스트에서만 주입한다.
var (
	sshSaveOpenFile  = os.OpenFile
	sshSaveOpen      = os.Open
	sshSaveRename    = os.Rename
	sshSaveRemoveAll = os.RemoveAll
	sshSaveChmod     = os.Chmod
	sshSaveMkdirDir  = os.Mkdir
	sshSaveMkdirTemp = os.MkdirTemp
	sshSaveCopyBytes = io.Copy
	sshSaveClose     = func(closer io.Closer) error { return closer.Close() }
	sshSaveWrite     = func(writer io.Writer, data []byte) (int, error) { return writer.Write(data) }
)

// EnhancedSSHOptions represents options for enhanced SSH commands.
type EnhancedSSHOptions struct {
	Name          string
	Description   string
	ConfigPath    string
	StorePath     string
	Force         bool
	ListAll       bool
	IncludeKeys   bool
	IncludePublic bool
}

// EnhancedSSHMetadata represents metadata for saved SSH configurations.
type EnhancedSSHMetadata struct {
	Description  string    `json:"description"`
	SavedAt      time.Time `json:"saved_at"`
	SourcePath   string    `json:"source_path"`
	IncludeFiles []string  `json:"include_files"`
	PrivateKeys  []string  `json:"private_keys"`
	PublicKeys   []string  `json:"public_keys"`
	HasIncludes  bool      `json:"has_includes"`
	HasKeys      bool      `json:"has_keys"`
}

// EnhancedSSHCommand provides enhanced SSH configuration management.
type EnhancedSSHCommand struct{}

// NewEnhancedSSHCommand creates a new enhanced SSH command instance.
func NewEnhancedSSHCommand() *EnhancedSSHCommand {
	return &EnhancedSSHCommand{}
}

// DefaultEnhancedOptions returns default options for enhanced SSH commands.
func (c *EnhancedSSHCommand) DefaultEnhancedOptions() *EnhancedSSHOptions {
	homeDir, _ := os.UserHomeDir()

	return &EnhancedSSHOptions{
		ConfigPath:    filepath.Join(homeDir, ".ssh", "config"),
		StorePath:     filepath.Join(homeDir, ".gz", "ssh-configs"),
		IncludeKeys:   true,
		IncludePublic: true,
	}
}

// CreateEnhancedSaveCommand creates the enhanced save command.
func (c *EnhancedSSHCommand) CreateEnhancedSaveCommand() *cobra.Command {
	opts := c.DefaultEnhancedOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current SSH configuration with includes and keys",
		Long: `Save current SSH configuration including:
- Main SSH config file
- All files referenced by Include directives
- All private keys referenced by IdentityFile directives
- Corresponding public keys (optional)

The configuration is published as one normalized, private snapshot.
Include files are stored with Load-compatible names; their source tree
layout is not preserved.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.SaveEnhancedConfig(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description for the saved configuration")
	cmd.Flags().StringVar(&opts.ConfigPath, "config-path", opts.ConfigPath, "Path to SSH config file")
	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to store saved configurations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVar(&opts.IncludeKeys, "include-keys", opts.IncludeKeys, "Include private keys")
	cmd.Flags().BoolVar(&opts.IncludePublic, "include-public", opts.IncludePublic, "Include public keys")

	cmd.MarkFlagRequired("name")

	return cmd
}

// CreateEnhancedLoadCommand creates the enhanced load command.
func (c *EnhancedSSHCommand) CreateEnhancedLoadCommand() *cobra.Command {
	opts := c.DefaultEnhancedOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load saved SSH configuration with includes and keys",
		Long: `Load saved SSH configuration including:
- Main SSH config file
- All included configuration files
- All private and public keys

The configuration is restored from the normalized snapshot layout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.LoadEnhancedConfig(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of the configuration to load (required)")
	cmd.Flags().StringVar(&opts.ConfigPath, "config-path", opts.ConfigPath, "Path to SSH config file")
	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to stored configurations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Replace an observed regular file; symlink targets are not followed")

	cmd.MarkFlagRequired("name")

	return cmd
}

// CreateEnhancedListCommand creates the enhanced list command.
func (c *EnhancedSSHCommand) CreateEnhancedListCommand() *cobra.Command {
	opts := c.DefaultEnhancedOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved SSH configurations with details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ListEnhancedConfigs(opts)
		},
	}

	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to stored configurations")
	cmd.Flags().BoolVarP(&opts.ListAll, "all", "a", false, "Show detailed information for all configurations")

	return cmd
}

// SaveEnhancedConfig saves the SSH configuration with includes and keys.
func (c *EnhancedSSHCommand) SaveEnhancedConfig(opts *EnhancedSSHOptions) (err error) {
	if err := validateSSHSaveName(opts.Name); err != nil {
		return err
	}
	if _, err := sshSaveRegularSource(opts.ConfigPath); err != nil {
		return fmt.Errorf("SSH config file not found or not regular at %s: %w", opts.ConfigPath, err)
	}
	parsed, err := NewSSHConfigParser(opts.ConfigPath).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse SSH configuration: %w", err)
	}
	if err := ensureSSHStore(opts.StorePath); err != nil {
		return err
	}
	configDir := filepath.Join(opts.StorePath, opts.Name)
	if err := checkSSHSaveDestination(configDir, opts.Force); err != nil {
		return err
	}

	includes, err := sshSaveDedupe(parsed.IncludeFiles, "include")
	if err != nil {
		return err
	}
	privateSources, publicSources := parsed.PrivateKeys, parsed.PublicKeys
	if !opts.IncludeKeys {
		privateSources = nil
	}
	if !opts.IncludePublic {
		publicSources = nil
	}
	privateKeys, publicKeys, err := sshSaveKeyPlan(privateSources, publicSources)
	if err != nil {
		return err
	}

	stage, err := sshSaveMkdirTemp(opts.StorePath, ".gzh-ssh-stage-")
	if err != nil {
		return fmt.Errorf("failed to create SSH staging directory: %w", err)
	}
	if err := sshSaveChmod(stage, 0o700); err != nil {
		cleanupErr := sshSaveRemoveAll(stage)
		return errors.Join(fmt.Errorf("failed to secure staging directory %s (%s): %w", stage, sshSaveObservedState(stage, configDir, "", ""), err), sshSaveCleanupError("stage", stage, cleanupErr))
	}
	disposition := sshSaveStageCleanup
	var publish sshSavePublishResult
	defer func() {
		if disposition == sshSaveStageCleanup {
			if cleanupErr := sshSaveRemoveAll(stage); cleanupErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to remove unpublished stage %s (stage=%s final=%s): %w", stage, stage, configDir, cleanupErr))
			}
			if publish.recovery != nil {
				err = errors.Join(err, fmt.Errorf("post-cleanup observed snapshot state: %s", sshSaveObservedState(publish.recovery.stage, publish.recovery.final, publish.recovery.backup, publish.recovery.wrapper)))
			}
		}
	}()
	if err := sshSaveMkdir(filepath.Join(stage, "includes"), len(includes) > 0); err != nil {
		return err
	}
	if err := sshSaveMkdir(filepath.Join(stage, "keys"), len(privateKeys)+len(publicKeys) > 0); err != nil {
		return err
	}
	if err := sshSaveCopy(opts.ConfigPath, filepath.Join(stage, "config"), 0o600); err != nil {
		return fmt.Errorf("failed to stage main config: %w", err)
	}
	for i, source := range includes {
		if err := sshSaveCopy(source, filepath.Join(stage, "includes", fmt.Sprintf("include_%d_%s", i, filepath.Base(source))), 0o600); err != nil {
			return fmt.Errorf("failed to stage include %s: %w", source, err)
		}
	}
	for _, source := range privateKeys {
		if err := sshSaveCopy(source, filepath.Join(stage, "keys", filepath.Base(source)), 0o600); err != nil {
			return fmt.Errorf("failed to stage private key %s: %w", source, err)
		}
	}
	for _, source := range publicKeys {
		if err := sshSaveCopy(source, filepath.Join(stage, "keys", filepath.Base(source)), 0o644); err != nil {
			return fmt.Errorf("failed to stage public key %s: %w", source, err)
		}
	}
	metadata := EnhancedSSHMetadata{Description: opts.Description, SavedAt: time.Now(), SourcePath: opts.ConfigPath, IncludeFiles: includes, PrivateKeys: privateKeys, PublicKeys: publicKeys, HasIncludes: len(includes) > 0, HasKeys: len(privateKeys)+len(publicKeys) > 0}
	//gosec:disable G117 -- PrivateKeys에는 키 내용이 아닌 선택되어 저장된 원본 경로만 기록한다.
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := sshSaveBytes(filepath.Join(stage, "metadata.json"), append(metadataBytes, '\n'), 0o600); err != nil {
		return fmt.Errorf("failed to stage metadata: %w", err)
	}
	publish, err = sshSavePublish(stage, configDir, opts.Force)
	disposition = publish.disposition
	if err != nil {
		return err
	}

	fmt.Printf("✅ SSH configuration '%s' saved successfully\n", opts.Name)
	if opts.Description != "" {
		fmt.Printf("   Description: %s\n", opts.Description)
	}
	fmt.Printf("   Main config: %s\n", parsed.MainConfigPath)
	fmt.Printf("   Include files: %d\n", len(includes))
	if opts.IncludeKeys {
		fmt.Printf("   Private keys: %d\n", len(privateKeys))
	}
	if opts.IncludePublic {
		fmt.Printf("   Public keys: %d\n", len(publicKeys))
	}
	fmt.Printf("   Saved to: %s\n", configDir)

	return nil
}

func validateSSHSaveName(name string) error {
	if name == "" {
		return fmt.Errorf("configuration name is required")
	}
	if filepath.Base(name) != name || name == "." || name == ".." || strings.ContainsAny(name, `/\\`) {
		return fmt.Errorf("configuration name must be one basename")
	}
	return nil
}

func ensureSSHStore(store string) error {
	info, err := os.Lstat(store)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(store, 0o700); err != nil {
			return fmt.Errorf("failed to create store directory: %w", err)
		}
		return sshSaveChmod(store, 0o700)
	}
	if err != nil {
		return fmt.Errorf("failed to inspect store directory: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("store path is not a real directory: %s", store)
	}
	return nil
}

func checkSSHSaveDestination(final string, force bool) error {
	info, err := os.Lstat(final)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to inspect destination %s: %w", final, err)
	}
	if !force {
		return fmt.Errorf("configuration '%s' already exists (use --force to overwrite)", filepath.Base(final))
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("--force only replaces a real directory at %s", final)
	}
	return nil
}

func sshSaveRegularSource(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("source is not a regular file: %s", path)
	}
	return resolved, nil
}

// sshSaveDedupe는 Load가 복원할 basename을 기준으로 충돌을 막는다. 같은 실제
// source라도 다른 basename이면 둘 다 복원해야 하므로, 같은 fold-name인 경우만
// 동일 inode를 최초 항목으로 합친다.
func sshSaveDedupe(paths []string, kind string) ([]string, error) {
	result := make([]string, 0, len(paths))
	type savedName struct {
		name string
		info os.FileInfo
	}
	byName := []savedName{}
	for _, path := range paths {
		_, err := sshSaveRegularSource(path)
		if err != nil {
			return nil, fmt.Errorf("invalid %s source %s: %w", kind, path, err)
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s source %s: %w", kind, path, err)
		}
		name := filepath.Base(path)
		matched := false
		for _, prior := range byName {
			if strings.EqualFold(prior.name, name) {
				matched = true
				if !os.SameFile(prior.info, info) {
					return nil, fmt.Errorf("%s basename collision for %q", kind, name)
				}
				break
			}
		}
		if matched {
			continue
		}
		byName = append(byName, savedName{name: name, info: info})
		result = append(result, path)
	}
	sort.Strings(result)
	return result, nil
}

// sshSaveKeyPlan은 keys/ 한 디렉터리의 실제 게시 대상을 하나의 계획으로 만든다.
// 동일 inode는 private 0600을 우선해 한 번만 저장하고, 다른 inode의 대소문자
// 무시 basename 충돌은 Load 시 모호해지므로 거부한다.
func sshSaveKeyPlan(privateSources, publicSources []string) ([]string, []string, error) {
	privateKeys, err := sshSaveDedupe(privateSources, "key")
	if err != nil {
		return nil, nil, err
	}
	publicKeys, err := sshSaveDedupe(publicSources, "key")
	if err != nil {
		return nil, nil, err
	}
	privateInfo := make([]struct {
		name string
		info os.FileInfo
	}, 0, len(privateKeys))
	for _, path := range privateKeys {
		info, err := os.Stat(path)
		if err != nil {
			return nil, nil, err
		}
		privateInfo = append(privateInfo, struct {
			name string
			info os.FileInfo
		}{name: filepath.Base(path), info: info})
	}
	actualPublic := make([]string, 0, len(publicKeys))
	for _, path := range publicKeys {
		info, err := os.Stat(path)
		if err != nil {
			return nil, nil, err
		}
		duplicatePrivate := false
		for _, old := range privateInfo {
			if strings.EqualFold(old.name, filepath.Base(path)) {
				if os.SameFile(old.info, info) {
					duplicatePrivate = true
					break
				}
				return nil, nil, fmt.Errorf("key basename collision for %q", filepath.Base(path))
			}
		}
		if !duplicatePrivate {
			actualPublic = append(actualPublic, path)
		}
	}
	return privateKeys, actualPublic, nil
}

func sshSaveMkdir(path string, needed bool) error {
	if !needed {
		return nil
	}
	if err := sshSaveMkdirDir(path, 0o700); err != nil {
		return fmt.Errorf("failed to create snapshot directory %s: %w", path, err)
	}
	if err := sshSaveChmod(path, 0o700); err != nil {
		return fmt.Errorf("failed to secure snapshot directory %s: %w", path, err)
	}
	return nil
}

func sshSaveCopy(source, destination string, mode os.FileMode) (err error) {
	if _, err = sshSaveRegularSource(source); err != nil {
		return err
	}
	src, err := sshSaveOpen(source)
	if err != nil {
		return fmt.Errorf("failed to open source %s: %w", source, err)
	}
	srcClosed := false
	dst, err := sshSaveOpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		closeErr := sshSaveClose(src)
		srcClosed = true
		if closeErr != nil {
			return errors.Join(fmt.Errorf("failed to create destination %s: %w", destination, err), fmt.Errorf("failed to close source %s: %w", source, closeErr))
		}
		return fmt.Errorf("failed to create destination %s: %w", destination, err)
	}
	dstClosed := false
	defer func() {
		var cleanup []error
		if !srcClosed {
			if closeErr := sshSaveClose(src); closeErr != nil {
				cleanup = append(cleanup, fmt.Errorf("failed to close source %s: %w", source, closeErr))
			}
		}
		if !dstClosed {
			if closeErr := sshSaveClose(dst); closeErr != nil {
				cleanup = append(cleanup, fmt.Errorf("failed to close destination %s: %w", destination, closeErr))
			}
		}
		if len(cleanup) > 0 {
			err = errors.Join(append([]error{err}, cleanup...)...)
		}
	}()
	if _, err = sshSaveCopyBytes(dst, src); err != nil {
		return fmt.Errorf("failed to copy %s: %w", source, err)
	}
	if closeErr := sshSaveClose(src); closeErr != nil {
		srcClosed = true
		return fmt.Errorf("failed to close source %s: %w", source, closeErr)
	}
	srcClosed = true
	if err = sshSaveChmod(destination, mode); err != nil {
		return fmt.Errorf("failed to set mode on %s: %w", destination, err)
	}
	if closeErr := sshSaveClose(dst); closeErr != nil {
		dstClosed = true
		return fmt.Errorf("failed to close destination %s: %w", destination, closeErr)
	}
	dstClosed = true
	return nil
}

func sshSaveBytes(destination string, data []byte, mode os.FileMode) error {
	file, err := sshSaveOpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create metadata %s: %w", destination, err)
	}
	if _, err := sshSaveWrite(file, data); err != nil {
		closeErr := sshSaveClose(file)
		if closeErr != nil {
			return errors.Join(fmt.Errorf("failed to write metadata %s: %w", destination, err), fmt.Errorf("failed to close metadata %s: %w", destination, closeErr))
		}
		return fmt.Errorf("failed to write metadata %s: %w", destination, err)
	}
	if err := sshSaveChmod(destination, mode); err != nil {
		closeErr := sshSaveClose(file)
		if closeErr != nil {
			return errors.Join(fmt.Errorf("failed to set mode on metadata %s: %w", destination, err), fmt.Errorf("failed to close metadata %s: %w", destination, closeErr))
		}
		return fmt.Errorf("failed to set mode on metadata %s: %w", destination, err)
	}
	if err := sshSaveClose(file); err != nil {
		return fmt.Errorf("failed to close metadata %s: %w", destination, err)
	}
	return nil
}

type sshSaveStageDisposition uint8

const (
	sshSaveStageCleanup sshSaveStageDisposition = iota
	sshSaveStageRetained
	sshSaveStageCommitted
)

type sshSaveRecoveryPaths struct{ stage, final, backup, wrapper string }

type sshSavePublishResult struct {
	disposition sshSaveStageDisposition
	recovery    *sshSaveRecoveryPaths
}

func sshSavePublish(stage, final string, force bool) (sshSavePublishResult, error) {
	if !force {
		if err := sshSaveRename(stage, final); err != nil {
			return sshSavePublishResult{}, fmt.Errorf("failed to publish staged snapshot %s to %s: %w", stage, final, err)
		}
		return sshSavePublishResult{disposition: sshSaveStageCommitted}, nil
	}
	info, err := os.Lstat(final)
	if os.IsNotExist(err) {
		if err := sshSaveRename(stage, final); err != nil {
			return sshSavePublishResult{}, fmt.Errorf("failed to publish staged snapshot %s to %s: %w", stage, final, err)
		}
		return sshSavePublishResult{disposition: sshSaveStageCommitted}, nil
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return sshSavePublishResult{}, fmt.Errorf("force destination changed or is not a real directory: %s", final)
	}
	wrapper, err := sshSaveMkdirTemp(filepath.Dir(final), ".gzh-ssh-backup-")
	if err != nil {
		return sshSavePublishResult{}, fmt.Errorf("failed to create backup wrapper beside %s: %w", final, err)
	}
	if err := sshSaveChmod(wrapper, 0o700); err != nil {
		cleanupErr := sshSaveRemoveAll(wrapper)
		return sshSavePublishResult{}, errors.Join(
			fmt.Errorf("failed to secure backup wrapper %s (%s): %w", wrapper, sshSaveObservedState(stage, final, "", wrapper), err),
			sshSaveCleanupError("backup wrapper", wrapper, cleanupErr),
		)
	}
	backup := filepath.Join(wrapper, filepath.Base(final))
	if err := sshSaveRename(final, backup); err != nil {
		wrapperErr := sshSaveRemoveAll(wrapper)
		state := sshSaveObservedState(stage, final, backup, wrapper)
		return sshSavePublishResult{}, errors.Join(
			fmt.Errorf("failed to move old snapshot %s to backup %s (%s): %w", final, backup, state, err),
			sshSaveCleanupError("backup wrapper", wrapper, wrapperErr),
		)
	}
	if err := sshSaveRename(stage, final); err != nil {
		rollbackErr := sshSaveRename(backup, final)
		if rollbackErr != nil {
			state := sshSaveObservedState(stage, final, backup, wrapper)
			return sshSavePublishResult{disposition: sshSaveStageRetained}, errors.Join(
				fmt.Errorf("failed to publish stage %s to final %s; backup may require recovery (stage retained; %s): %w", stage, final, state, err),
				fmt.Errorf("failed to rollback backup %s to final %s (%s): %w", backup, final, state, rollbackErr),
			)
		}
		wrapperErr := sshSaveRemoveAll(wrapper)
		return sshSavePublishResult{recovery: &sshSaveRecoveryPaths{stage, final, backup, wrapper}}, errors.Join(
			fmt.Errorf("failed to publish stage %s to final %s; old snapshot restored: %w", stage, final, err),
			sshSaveCleanupError("backup wrapper", wrapper, wrapperErr),
		)
	}
	if err := sshSaveRemoveAll(wrapper); err != nil {
		return sshSavePublishResult{disposition: sshSaveStageCommitted}, fmt.Errorf("snapshot committed; backup cleanup pending (%s): %w", sshSaveObservedState(stage, final, backup, wrapper), err)
	}
	return sshSavePublishResult{disposition: sshSaveStageCommitted}, nil
}

func sshSaveCleanupError(kind, path string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to remove %s %s: %w", kind, path, err)
}

// sshSaveObservedState reports the post-operation view without making recovery
// claims when an OS cleanup can have removed only part of a directory tree.
func sshSaveObservedState(stage, final, backup, wrapper string) string {
	state := func(path string) string {
		if _, err := os.Lstat(path); err == nil {
			return "exists"
		} else if os.IsNotExist(err) {
			return "missing"
		}
		return "unknown"
	}
	parts := []string{}
	for _, item := range []struct{ name, path string }{{"stage", stage}, {"final", final}, {"backup", backup}, {"wrapper", wrapper}} {
		if item.path != "" {
			parts = append(parts, fmt.Sprintf("%s=%s(%s)", item.name, item.path, state(item.path)))
		}
	}
	return strings.Join(parts, " ")
}

// LoadEnhancedConfig loads a saved SSH configuration.
func (c *EnhancedSSHCommand) LoadEnhancedConfig(opts *EnhancedSSHOptions) error {
	// Validate inputs
	if opts.Name == "" {
		return fmt.Errorf("configuration name is required")
	}

	// Check if saved config exists
	configDir := filepath.Join(opts.StorePath, opts.Name)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fmt.Errorf("configuration '%s' not found", opts.Name)
	}

	// Load metadata
	metadataFile := filepath.Join(configDir, "metadata.json")
	metadata, err := c.loadEnhancedMetadata(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}
	if err := preflightEnhancedLoad(configDir, opts.ConfigPath, opts.Force); err != nil {
		return err
	}

	// Create target directory
	if err := os.MkdirAll(filepath.Dir(opts.ConfigPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load main config file
	mainConfigSrc := filepath.Join(configDir, "config")
	if err := c.restoreEnhancedFile(mainConfigSrc, opts.ConfigPath, 0o600, opts.Force); err != nil {
		return fmt.Errorf("failed to load main config: %w", err)
	}

	loadedFiles := 1

	// Load include files (this is tricky - we need to restore them to their original paths)
	includeDir := filepath.Join(configDir, "includes")
	if _, err := os.Stat(includeDir); err == nil {
		entries, err := os.ReadDir(includeDir)
		if err != nil {
			return fmt.Errorf("failed to read includes directory: %w", err)
		}

		configD := filepath.Join(filepath.Dir(opts.ConfigPath), "config.d")
		hasIncludeFiles := false
		for _, entry := range entries {
			if !entry.IsDir() {
				hasIncludeFiles = true
				break
			}
		}
		if hasIncludeFiles {
			// Include 파일은 대상 설정 디렉터리 아래에 복원한다.
			if err := os.MkdirAll(configD, 0o755); err != nil {
				return fmt.Errorf("failed to create includes directory: %w", err)
			}
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			srcPath := filepath.Join(includeDir, entry.Name())
			// 저장 시 추가한 접두사를 제거한다.
			originalName := sshLoadIncludeName(entry.Name())
			destPath := filepath.Join(configD, originalName)
			if err := c.restoreEnhancedFile(srcPath, destPath, 0o600, opts.Force); err != nil {
				return fmt.Errorf("failed to load include file %s: %w", entry.Name(), err)
			}
			loadedFiles++
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect includes directory: %w", err)
	}

	// Load keys
	keysDir := filepath.Join(configDir, "keys")
	if _, err := os.Stat(keysDir); err == nil {
		sshKeysDir := filepath.Dir(opts.ConfigPath)
		entries, err := os.ReadDir(keysDir)
		if err != nil {
			return fmt.Errorf("failed to read keys directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			srcPath := filepath.Join(keysDir, entry.Name())
			destPath := filepath.Join(sshKeysDir, entry.Name())
			mode := sshLoadKeyMode(entry.Name(), metadata)
			if err := c.restoreEnhancedFile(srcPath, destPath, mode, opts.Force); err != nil {
				return fmt.Errorf("failed to load key %s: %w", entry.Name(), err)
			}
			loadedFiles++
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect keys directory: %w", err)
	}

	// Print summary
	fmt.Printf("✅ SSH configuration '%s' loaded successfully\n", opts.Name)
	if metadata.Description != "" {
		fmt.Printf("   Description: %s\n", metadata.Description)
	}
	fmt.Printf("   Originally saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Files restored: %d\n", loadedFiles)
	fmt.Printf("   Loaded to: %s\n", opts.ConfigPath)

	return nil
}

func sshLoadKeyMode(name string, metadata *EnhancedSSHMetadata) os.FileMode {
	for _, path := range metadata.PrivateKeys {
		if strings.EqualFold(filepath.Base(path), name) {
			return 0o600
		}
	}
	for _, path := range metadata.PublicKeys {
		if strings.EqualFold(filepath.Base(path), name) {
			return 0o644
		}
	}
	// 레거시 metadata에 역할 목록이 모두 없을 때만 과거 suffix 규칙을 사용한다.
	if len(metadata.PrivateKeys) == 0 && len(metadata.PublicKeys) == 0 && strings.HasSuffix(name, ".pub") {
		return 0o644
	}
	return 0o600
}

// preflightEnhancedLoad은 복원 전에 최종 경로 전체를 계산해, snapshot 내부의
// 이름 충돌이나 no-force 대상 충돌 때문에 일부만 복원되는 것을 막는다.
func preflightEnhancedLoad(configDir, configPath string, force bool) error {
	destinations := []string{}
	reserved := []string{}
	collides := func(paths []string, path string) (string, bool) {
		for _, old := range paths {
			if strings.EqualFold(old, path) {
				return old, true
			}
		}
		return "", false
	}
	addDestination := func(path string) error {
		if old, ok := collides(destinations, path); ok {
			return fmt.Errorf("snapshot restore destination collision: %s and %s", old, path)
		}
		if old, ok := collides(reserved, path); ok {
			return fmt.Errorf("snapshot restore destination collision: reserved %s and %s", old, path)
		}
		destinations = append(destinations, path)
		return nil
	}
	addReserved := func(path string) error {
		if old, ok := collides(destinations, path); ok {
			return fmt.Errorf("snapshot restore destination collision: %s and reserved %s", old, path)
		}
		if old, ok := collides(reserved, path); ok {
			return fmt.Errorf("snapshot restore destination collision: reserved %s and %s", old, path)
		}
		reserved = append(reserved, path)
		return nil
	}
	if err := addDestination(configPath); err != nil {
		return err
	}
	includeTarget := filepath.Join(filepath.Dir(configPath), "config.d")
	for _, spec := range []struct {
		dir, target string
		include     bool
	}{
		{filepath.Join(configDir, "includes"), includeTarget, true},
		{filepath.Join(configDir, "keys"), filepath.Dir(configPath), false},
	} {
		entries, err := os.ReadDir(spec.dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to inspect snapshot directory %s: %w", spec.dir, err)
		}
		hasFiles := false
		for _, entry := range entries {
			if !entry.IsDir() {
				hasFiles = true
				break
			}
		}
		if spec.include && hasFiles {
			if err := addReserved(includeTarget); err != nil {
				return err
			}
			if info, lstatErr := os.Lstat(includeTarget); lstatErr == nil {
				if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
					return fmt.Errorf("restore include container is not a real directory: %s", includeTarget)
				}
			} else if !os.IsNotExist(lstatErr) {
				return fmt.Errorf("failed to inspect restore include container %s: %w", includeTarget, lstatErr)
			}
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if spec.include {
				name = sshLoadIncludeName(name)
			}
			if err := addDestination(filepath.Join(spec.target, name)); err != nil {
				return err
			}
		}
	}
	for _, path := range destinations {
		if info, err := os.Lstat(path); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return fmt.Errorf("restore destination is not a regular file: %s", path)
			}
			if !force {
				return fmt.Errorf("restore destination already exists: %s", path)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to inspect restore destination %s: %w", path, err)
		}
	}
	return nil
}

func sshLoadIncludeName(name string) string {
	name = strings.TrimPrefix(name, "include_")
	if index := strings.Index(name, "_"); index > 0 {
		return name[index+1:]
	}
	return name
}

// restoreEnhancedFile은 저장된 SSH 파일 하나를 최종 대상에 복원한다.
// 사용자 소유이고 비적대적인 부모 경로를 전제로, 복사 전 관찰한 최종 entry만
// Lstat으로 검증한다. 심볼릭 링크 대상은 따라가지 않지만 동시 entry 교체는 이
// 배치의 위협 모델 밖이며, source 검증·조상 no-follow·다중 파일 롤백도 범위 밖이다.
func (c *EnhancedSSHCommand) restoreEnhancedFile(src, dst string, mode os.FileMode, force bool) (err error) {
	if mode != 0o600 && mode != 0o644 {
		return fmt.Errorf("unsupported restore mode %04o", mode)
	}

	if info, lstatErr := os.Lstat(dst); lstatErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlink destination %s", dst)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing non-regular destination %s", dst)
		}
		if !force {
			return fmt.Errorf("destination already exists at %s (use --force to replace a regular file)", dst)
		}
	} else if !os.IsNotExist(lstatErr) {
		return fmt.Errorf("failed to inspect destination %s: %w", dst, lstatErr)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source %s: %w", src, err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(dst), ".gzh-ssh-restore-*")
	if err != nil {
		if closeErr := sourceFile.Close(); closeErr != nil {
			return errors.Join(
				fmt.Errorf("failed to create restore temp for %s: %w", dst, err),
				fmt.Errorf("failed to close source %s: %w", src, closeErr),
			)
		}
		return fmt.Errorf("failed to create restore temp for %s: %w", dst, err)
	}
	tempPath := tempFile.Name()
	published := false
	sourceClosed := false
	tempClosed := false
	defer func() {
		var cleanupErrs []error
		if !sourceClosed {
			if closeErr := sourceFile.Close(); closeErr != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("failed to close source %s: %w", src, closeErr))
			}
		}
		if !tempClosed {
			if closeErr := tempFile.Close(); closeErr != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("failed to close restore temp %s: %w", tempPath, closeErr))
			}
		}
		if !published {
			if removeErr := os.Remove(tempPath); removeErr != nil && !os.IsNotExist(removeErr) {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("failed to remove restore temp %s: %w", tempPath, removeErr))
			}
		}
		if len(cleanupErrs) > 0 {
			err = errors.Join(append([]error{err}, cleanupErrs...)...)
		}
	}()

	if err = tempFile.Chmod(0o600); err != nil {
		return fmt.Errorf("failed to secure restore temp %s: %w", tempPath, err)
	}
	if _, err = io.Copy(tempFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy %s to restore temp: %w", src, err)
	}
	if closeErr := sourceFile.Close(); closeErr != nil {
		sourceClosed = true
		return fmt.Errorf("failed to close source %s: %w", src, closeErr)
	}
	sourceClosed = true
	if err = tempFile.Chmod(mode); err != nil {
		return fmt.Errorf("failed to set restore mode for %s: %w", dst, err)
	}
	if err = tempFile.Close(); err != nil {
		tempClosed = true
		return fmt.Errorf("failed to close restore temp %s: %w", tempPath, err)
	}
	tempClosed = true

	if force {
		if err = os.Rename(tempPath, dst); err != nil {
			return fmt.Errorf("failed to replace destination %s: %w", dst, err)
		}
		published = true
		return nil
	}

	if err = os.Link(tempPath, dst); err != nil {
		return fmt.Errorf("failed to publish destination %s without replacement: %w", dst, err)
	}
	published = true
	if err = os.Remove(tempPath); err != nil {
		return fmt.Errorf("published destination %s but failed to remove restore temp %s: %w", dst, tempPath, err)
	}

	return nil
}

// ListEnhancedConfigs lists saved SSH configurations.
func (c *EnhancedSSHCommand) ListEnhancedConfigs(opts *EnhancedSSHOptions) error {
	// Check if store directory exists
	if _, err := os.Stat(opts.StorePath); os.IsNotExist(err) {
		fmt.Printf("No SSH configurations found (store directory doesn't exist)\n")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(opts.StorePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for directories
	var configs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has metadata.json
			metadataPath := filepath.Join(opts.StorePath, entry.Name(), "metadata.json")
			if _, err := os.Stat(metadataPath); err == nil {
				configs = append(configs, entry.Name())
			}
		}
	}

	if len(configs) == 0 {
		fmt.Printf("No SSH configurations found\n")
		return nil
	}

	fmt.Printf("Saved SSH configurations:\n\n")

	for _, configName := range configs {
		if opts.ListAll {
			c.printDetailedEnhancedConfig(opts.StorePath, configName)
		} else {
			fmt.Printf("  • %s\n", configName)
		}
	}

	if !opts.ListAll {
		fmt.Printf("\nUse --all to show detailed information\n")
	}

	return nil
}

// printDetailedEnhancedConfig prints detailed information about a configuration.
func (c *EnhancedSSHCommand) printDetailedEnhancedConfig(storePath, configName string) {
	fmt.Printf("  📁 %s\n", configName)

	metadataFile := filepath.Join(storePath, configName, "metadata.json")
	if metadata, err := c.loadEnhancedMetadata(metadataFile); err == nil {
		if metadata.Description != "" {
			fmt.Printf("     Description: %s\n", metadata.Description)
		}
		fmt.Printf("     Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		if metadata.SourcePath != "" {
			fmt.Printf("     Source: %s\n", metadata.SourcePath)
		}
		if metadata.HasIncludes {
			fmt.Printf("     Include files: %d\n", len(metadata.IncludeFiles))
		}
		if metadata.HasKeys {
			fmt.Printf("     Private keys: %d\n", len(metadata.PrivateKeys))
			fmt.Printf("     Public keys: %d\n", len(metadata.PublicKeys))
		}
	}

	configDir := filepath.Join(storePath, configName)
	if entries, err := os.ReadDir(configDir); err == nil {
		totalSize := int64(0)
		for _, entry := range entries {
			if entry.Name() != "metadata.json" {
				if info, err := entry.Info(); err == nil {
					totalSize += info.Size()
				}
			}
		}
		fmt.Printf("     Total size: %d bytes\n", totalSize)
	}

	fmt.Println()
}

// saveEnhancedMetadata saves enhanced metadata to a JSON file.
func (c *EnhancedSSHCommand) saveEnhancedMetadata(filename string, metadata EnhancedSSHMetadata) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	//gosec:disable G117 -- PrivateKeys는 키 내용이 아니라 원본 파일 경로만 포함한다.
	return encoder.Encode(metadata)
}

// loadEnhancedMetadata loads enhanced metadata from a JSON file.
func (c *EnhancedSSHCommand) loadEnhancedMetadata(filename string) (*EnhancedSSHMetadata, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata EnhancedSSHMetadata
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&metadata)
	return &metadata, err
}
