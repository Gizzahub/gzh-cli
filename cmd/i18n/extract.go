package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/i18n"
	"github.com/spf13/cobra"
)

// ExtractCmd represents the extract command
var ExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract translatable messages from source code",
	Long: `Extract translatable messages from source code files and generate message catalogs.

This command scans Go, TypeScript, JavaScript source files for translatable strings
and generates message catalog files that can be used for translation.

Examples:
  gz i18n extract --source ./cmd,./pkg --output locales/messages.json
  gz i18n extract --patterns "*.go,*.ts" --functions "T,Tn,Tf"
  gz i18n extract --merge --sort`,
	Run: runExtract,
}

var (
	sourceDirs        []string
	outputFile        string
	filePatterns      []string
	functionNames     []string
	keyPatterns       []string
	includeSource     bool
	sortOutput        bool
	mergeExisting     bool
	generateTemplates bool
	templateLangs     []string
	verbose           bool
)

func init() {
	ExtractCmd.Flags().StringSliceVarP(&sourceDirs, "source", "s", []string{"./cmd", "./pkg", "./internal"}, "Source directories to scan")
	ExtractCmd.Flags().StringVarP(&outputFile, "output", "o", "locales/messages.json", "Output file for extracted messages")
	ExtractCmd.Flags().StringSliceVar(&filePatterns, "patterns", []string{"*.go", "*.ts", "*.js"}, "File patterns to include")
	ExtractCmd.Flags().StringSliceVar(&functionNames, "functions", []string{"T", "Tn", "Tf", "MustT"}, "Translation function names to look for")
	ExtractCmd.Flags().StringSliceVar(&keyPatterns, "key-patterns", nil, "Custom regex patterns for extracting message keys")
	ExtractCmd.Flags().BoolVar(&includeSource, "include-source", true, "Include source location information")
	ExtractCmd.Flags().BoolVar(&sortOutput, "sort", true, "Sort messages by ID in output")
	ExtractCmd.Flags().BoolVar(&mergeExisting, "merge", true, "Merge with existing messages")
	ExtractCmd.Flags().BoolVar(&generateTemplates, "templates", false, "Generate translation templates")
	ExtractCmd.Flags().StringSliceVar(&templateLangs, "template-langs", []string{"ko", "ja", "zh"}, "Languages for template generation")
	ExtractCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func runExtract(cmd *cobra.Command, args []string) {
	if verbose {
		fmt.Println("🔍 Extracting translatable messages...")
		fmt.Printf("📁 Source directories: %s\n", strings.Join(sourceDirs, ", "))
		fmt.Printf("📄 File patterns: %s\n", strings.Join(filePatterns, ", "))
		fmt.Printf("🔧 Function names: %s\n", strings.Join(functionNames, ", "))
		fmt.Printf("💾 Output file: %s\n", outputFile)
	}

	// Validate source directories
	for _, dir := range sourceDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("❌ Source directory does not exist: %s\n", dir)
			os.Exit(1)
		}
	}

	// Create extractor configuration
	config := &i18n.ExtractorConfig{
		SourceDirs:    sourceDirs,
		OutputFile:    outputFile,
		FilePatterns:  filePatterns,
		FunctionNames: functionNames,
		KeyPatterns:   keyPatterns,
		IncludeSource: includeSource,
		SortOutput:    sortOutput,
		MergeExisting: mergeExisting,
	}

	// Add default key patterns if none specified
	if len(keyPatterns) == 0 {
		config.KeyPatterns = []string{
			`T\s*\(\s*["']([^"']+)["']`,
			`Tn\s*\(\s*["']([^"']+)["']`,
			`Tf\s*\(\s*["']([^"']+)["']`,
			`MustT\s*\(\s*["']([^"']+)["']`,
			`i18n\.T\s*\(\s*["']([^"']+)["']`,
			`const\s+Msg\w+\s*=\s*["']([a-zA-Z0-9_.]+)["']`,
		}
	}

	// Create extractor
	extractor := i18n.NewExtractor(config)

	// Extract messages
	if err := extractor.Extract(); err != nil {
		fmt.Printf("❌ Failed to extract messages: %v\n", err)
		os.Exit(1)
	}

	messageCount := extractor.GetMessageCount()
	fmt.Printf("✅ Extracted %d messages to %s\n", messageCount, outputFile)

	// Generate translation templates if requested
	if generateTemplates {
		fmt.Println("📝 Generating translation templates...")

		for _, lang := range templateLangs {
			templateFile := filepath.Join(filepath.Dir(outputFile), fmt.Sprintf("template_%s.json", lang))

			if err := extractor.GenerateTemplate(lang, templateFile); err != nil {
				fmt.Printf("❌ Failed to generate template for %s: %v\n", lang, err)
				continue
			}

			fmt.Printf("✅ Generated template for %s: %s\n", lang, templateFile)
		}
	}

	// Show summary
	if verbose {
		fmt.Println("\n📊 Extraction Summary:")
		messages := extractor.GetMessages()

		// Count by context
		contextCount := make(map[string]int)
		for _, msg := range messages {
			contextCount[msg.Context]++
		}

		for context, count := range contextCount {
			fmt.Printf("  %s: %d messages\n", context, count)
		}

		// Show sample messages
		fmt.Println("\n📋 Sample Messages:")
		count := 0
		for id, msg := range messages {
			if count >= 5 {
				break
			}
			fmt.Printf("  %s (%s)\n", id, msg.Context)
			if len(msg.Locations) > 0 {
				fmt.Printf("    └─ %s:%d\n", msg.Locations[0].File, msg.Locations[0].Line)
			}
			count++
		}

		if len(messages) > 5 {
			fmt.Printf("  ... and %d more\n", len(messages)-5)
		}
	}

	fmt.Println("\n💡 Next steps:")
	fmt.Println("  1. Review extracted messages in", outputFile)
	fmt.Println("  2. Translate messages using the generated templates")
	fmt.Println("  3. Initialize i18n in your application with: gz i18n init")
}
