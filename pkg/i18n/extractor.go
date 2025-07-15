package i18n

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Extractor extracts translatable messages from source code
type Extractor struct {
	config        *ExtractorConfig
	messages      map[string]*ExtractedMessage
	functionNames []string
	patterns      []*regexp.Regexp
}

// ExtractorConfig holds configuration for message extraction
type ExtractorConfig struct {
	// SourceDirs are directories to scan for source files
	SourceDirs []string
	// OutputFile is where to write extracted messages
	OutputFile string
	// FilePatterns are file patterns to include (e.g., "*.go", "*.ts")
	FilePatterns []string
	// FunctionNames are function names to look for (e.g., "T", "Tn", "Tf")
	FunctionNames []string
	// KeyPatterns are regex patterns to extract message keys
	KeyPatterns []string
	// IncludeSource includes source location in output
	IncludeSource bool
	// SortOutput sorts messages by ID in output
	SortOutput bool
	// MergeExisting merges with existing translations
	MergeExisting bool
}

// ExtractedMessage represents a message found in source code
type ExtractedMessage struct {
	ID          string            `json:"id"`
	Message     string            `json:"message,omitempty"`
	Description string            `json:"description,omitempty"`
	Locations   []SourceLocation  `json:"locations,omitempty"`
	Context     string            `json:"context,omitempty"`
	Plural      map[string]string `json:"plural,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SourceLocation represents where a message was found
type SourceLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// DefaultExtractorConfig returns a default extractor configuration
func DefaultExtractorConfig() *ExtractorConfig {
	return &ExtractorConfig{
		SourceDirs: []string{"./cmd", "./pkg", "./internal"},
		OutputFile: "locales/messages.json",
		FilePatterns: []string{
			"*.go",
			"*.ts",
			"*.js",
			"*.tsx",
			"*.jsx",
		},
		FunctionNames: []string{
			"T",
			"Tn",
			"Tf",
			"MustT",
			"i18n.T",
			"i18n.Tn",
			"i18n.Tf",
			"manager.T",
			"manager.Tn",
			"manager.Tf",
		},
		KeyPatterns: []string{
			`T\s*\(\s*["']([^"']+)["']`,                  // T("key")
			`Tn\s*\(\s*["']([^"']+)["']`,                 // Tn("key", count)
			`Tf\s*\(\s*["']([^"']+)["']`,                 // Tf("key", args...)
			`MustT\s*\(\s*["']([^"']+)["']`,              // MustT("key")
			`i18n\.T\s*\(\s*["']([^"']+)["']`,            // i18n.T("key")
			`manager\.T\s*\(\s*["']([^"']+)["']`,         // manager.T("key")
			`const\s+\w+\s*=\s*["']([a-zA-Z0-9_.]+)["']`, // const MsgKey = "msg.key"
		},
		IncludeSource: true,
		SortOutput:    true,
		MergeExisting: true,
	}
}

// NewExtractor creates a new message extractor
func NewExtractor(config *ExtractorConfig) *Extractor {
	if config == nil {
		config = DefaultExtractorConfig()
	}

	extractor := &Extractor{
		config:        config,
		messages:      make(map[string]*ExtractedMessage),
		functionNames: config.FunctionNames,
	}

	// Compile regex patterns
	for _, pattern := range config.KeyPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			extractor.patterns = append(extractor.patterns, regex)
		}
	}

	return extractor
}

// Extract extracts messages from source files
func (e *Extractor) Extract() error {
	// Load existing messages if requested
	if e.config.MergeExisting {
		if err := e.loadExistingMessages(); err != nil {
			fmt.Printf("Warning: Could not load existing messages: %v\n", err)
		}
	}

	// Extract from each source directory
	for _, dir := range e.config.SourceDirs {
		if err := e.extractFromDirectory(dir); err != nil {
			return fmt.Errorf("failed to extract from directory %s: %w", dir, err)
		}
	}

	// Save extracted messages
	return e.saveMessages()
}

// extractFromDirectory extracts messages from all files in a directory
func (e *Extractor) extractFromDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches any pattern
		for _, pattern := range e.config.FilePatterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				return e.extractFromFile(path)
			}
		}

		return nil
	})
}

// extractFromFile extracts messages from a single file
func (e *Extractor) extractFromFile(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".go":
		return e.extractFromGoFile(filename)
	case ".ts", ".js", ".tsx", ".jsx":
		return e.extractFromJSFile(filename)
	default:
		return e.extractFromTextFile(filename)
	}
}

// extractFromGoFile extracts messages from Go source files using AST
func (e *Extractor) extractFromGoFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse Go file %s: %w", filename, err)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			e.extractFromCallExpr(fset, x, filename)
		case *ast.GenDecl:
			e.extractFromGenDecl(fset, x, filename)
		}
		return true
	})

	return nil
}

// extractFromCallExpr extracts messages from function calls
func (e *Extractor) extractFromCallExpr(fset *token.FileSet, call *ast.CallExpr, filename string) {
	// Get function name
	var funcName string
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		funcName = fun.Name
	case *ast.SelectorExpr:
		if x, ok := fun.X.(*ast.Ident); ok {
			funcName = x.Name + "." + fun.Sel.Name
		}
	default:
		return
	}

	// Check if this is a translation function
	var isTranslationFunc bool
	for _, name := range e.functionNames {
		if funcName == name || strings.HasSuffix(funcName, "."+name) {
			isTranslationFunc = true
			break
		}
	}

	if !isTranslationFunc {
		return
	}

	// Extract message ID from first argument
	if len(call.Args) == 0 {
		return
	}

	var messageID string
	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		// Remove quotes
		messageID = lit.Value[1 : len(lit.Value)-1]
	} else {
		return
	}

	// Get source location
	pos := fset.Position(call.Pos())
	location := SourceLocation{
		File:   filename,
		Line:   pos.Line,
		Column: pos.Column,
	}

	e.addMessage(messageID, "", location, funcName)
}

// extractFromGenDecl extracts message constants from declarations
func (e *Extractor) extractFromGenDecl(fset *token.FileSet, decl *ast.GenDecl, filename string) {
	if decl.Tok != token.CONST {
		return
	}

	for _, spec := range decl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for i, name := range valueSpec.Names {
				if i < len(valueSpec.Values) {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						// Check if this looks like a message key
						value := lit.Value[1 : len(lit.Value)-1] // Remove quotes
						if e.looksLikeMessageKey(value) {
							pos := fset.Position(lit.Pos())
							location := SourceLocation{
								File:   filename,
								Line:   pos.Line,
								Column: pos.Column,
							}
							e.addMessage(value, name.Name, location, "const")
						}
					}
				}
			}
		}
	}
}

// extractFromJSFile extracts messages from JavaScript/TypeScript files
func (e *Extractor) extractFromJSFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return e.extractWithRegex(string(content), filename)
}

// extractFromTextFile extracts messages from text files using regex
func (e *Extractor) extractFromTextFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return e.extractWithRegex(string(content), filename)
}

// extractWithRegex extracts messages using regex patterns
func (e *Extractor) extractWithRegex(content, filename string) error {
	lines := strings.Split(content, "\n")

	for _, pattern := range e.patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				messageID := match[1]

				// Find line number
				lineNum := 1
				for i, line := range lines {
					if strings.Contains(line, match[0]) {
						lineNum = i + 1
						break
					}
				}

				location := SourceLocation{
					File:   filename,
					Line:   lineNum,
					Column: 1,
				}

				e.addMessage(messageID, "", location, "regex")
			}
		}
	}

	return nil
}

// addMessage adds a message to the extracted messages
func (e *Extractor) addMessage(id, description string, location SourceLocation, context string) {
	if existing, exists := e.messages[id]; exists {
		// Add location to existing message
		existing.Locations = append(existing.Locations, location)
	} else {
		// Create new message
		e.messages[id] = &ExtractedMessage{
			ID:          id,
			Description: description,
			Locations:   []SourceLocation{location},
			Context:     context,
			Metadata:    make(map[string]string),
		}
	}
}

// looksLikeMessageKey checks if a string looks like a message key
func (e *Extractor) looksLikeMessageKey(s string) bool {
	// Simple heuristic: contains dots and starts with known prefixes
	if strings.Contains(s, ".") {
		prefixes := []string{"msg", "error", "cmd", "clone", "plugin", "docker"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(strings.ToLower(s), prefix) {
				return true
			}
		}
	}
	return false
}

// loadExistingMessages loads existing messages from output file
func (e *Extractor) loadExistingMessages() error {
	if _, err := os.Stat(e.config.OutputFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to load
	}

	data, err := os.ReadFile(e.config.OutputFile)
	if err != nil {
		return err
	}

	var existing map[string]*ExtractedMessage
	if err := json.Unmarshal(data, &existing); err != nil {
		return err
	}

	// Merge existing messages
	for id, msg := range existing {
		if _, exists := e.messages[id]; !exists {
			e.messages[id] = msg
		}
	}

	return nil
}

// saveMessages saves extracted messages to output file
func (e *Extractor) saveMessages() error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(e.config.OutputFile), 0o755); err != nil {
		return err
	}

	// Prepare output
	var output interface{}
	if e.config.SortOutput {
		// Sort messages by ID
		var sortedMessages []*ExtractedMessage
		var keys []string
		for id := range e.messages {
			keys = append(keys, id)
		}
		sort.Strings(keys)

		for _, id := range keys {
			msg := e.messages[id]
			if !e.config.IncludeSource {
				msg.Locations = nil
			}
			sortedMessages = append(sortedMessages, msg)
		}
		output = sortedMessages
	} else {
		if !e.config.IncludeSource {
			for _, msg := range e.messages {
				msg.Locations = nil
			}
		}
		output = e.messages
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(e.config.OutputFile, data, 0o644)
}

// GetMessages returns all extracted messages
func (e *Extractor) GetMessages() map[string]*ExtractedMessage {
	return e.messages
}

// GetMessageCount returns the number of extracted messages
func (e *Extractor) GetMessageCount() int {
	return len(e.messages)
}

// GenerateTemplate generates a template file for translators
func (e *Extractor) GenerateTemplate(lang string, outputFile string) error {
	bundle := &LocalizationBundle{
		Language: lang,
		Version:  "1.0.0",
		Messages: make(map[string]MessageConfig),
	}

	// Convert extracted messages to bundle format
	for id, msg := range e.messages {
		bundle.Messages[id] = MessageConfig{
			ID:          id,
			Description: msg.Description,
			Message:     "", // Empty for translators to fill
		}
	}

	// Save template
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, data, 0o644)
}

// ExtractFromSource is a convenience function to extract messages
func ExtractFromSource(sourceDirs []string, outputFile string) error {
	config := DefaultExtractorConfig()
	config.SourceDirs = sourceDirs
	config.OutputFile = outputFile

	extractor := NewExtractor(config)
	return extractor.Extract()
}
