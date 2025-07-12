// Package i18n provides internationalization commands for GZH Manager
package i18n

import (
	"github.com/spf13/cobra"
)

// I18nCmd represents the i18n command
var I18nCmd = &cobra.Command{
	Use:   "i18n",
	Short: "Internationalization management commands",
	Long: `Internationalization (i18n) management commands for GZH Manager.

This command group provides tools for managing translations, extracting
translatable messages from source code, and validating translation files.

Available subcommands:
  init     - Initialize i18n configuration and locale files
  extract  - Extract translatable messages from source code
  validate - Validate translation files and check for issues
  serve    - Start translation management server (if available)

Examples:
  gz i18n init --languages en,ko,ja
  gz i18n extract --source ./cmd,./pkg
  gz i18n validate --locales ./locales`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Add subcommands
	I18nCmd.AddCommand(InitCmd)
	I18nCmd.AddCommand(ExtractCmd)
	I18nCmd.AddCommand(ValidateCmd)
}
