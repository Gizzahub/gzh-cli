package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Solution represents a potential solution for an error
type Solution struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Steps       []SolutionStep    `json:"steps"`
	Commands    []Command         `json:"commands,omitempty"`
	Links       []Link            `json:"links,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Priority    int               `json:"priority"`   // Higher = more important
	Confidence  float64           `json:"confidence"` // 0.0-1.0
	Context     map[string]string `json:"context,omitempty"`
	Automated   bool              `json:"automated"` // Can be auto-applied
}

// SolutionStep represents a single step in a solution
type SolutionStep struct {
	Order        int    `json:"order"`
	Action       string `json:"action"`
	Description  string `json:"description"`
	Command      string `json:"command,omitempty"`
	Verification string `json:"verification,omitempty"`
}

// Command represents an executable command
type Command struct {
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Description string   `json:"description"`
	Args        []string `json:"args,omitempty"`
	WorkingDir  string   `json:"working_dir,omitempty"`
	Dangerous   bool     `json:"dangerous"` // Requires confirmation
}

// Link represents a documentation or FAQ link
type Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"` // "docs", "faq", "tutorial", "forum"
}

// SolutionMatcher defines how to match errors to solutions
type SolutionMatcher struct {
	ErrorCode    *ErrorCode `json:"error_code,omitempty"`
	DomainRegex  string     `json:"domain_regex,omitempty"`
	MessageRegex string     `json:"message_regex,omitempty"`
	ContextKeys  []string   `json:"context_keys,omitempty"`
	Priority     int        `json:"priority"`
}

// SolutionDatabase manages solutions for different error types
type SolutionDatabase struct {
	mu        sync.RWMutex
	solutions map[string][]Solution // Keyed by error code string
	matchers  []SolutionMatcher
	providers []SolutionProvider
}

// SolutionProvider interface for external solution sources
type SolutionProvider interface {
	GetSolutions(ctx context.Context, err *UserError) ([]Solution, error)
	GetProviderInfo() ProviderInfo
}

// ProviderInfo describes a solution provider
type ProviderInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

// SolutionEngine provides intelligent error resolution suggestions
type SolutionEngine struct {
	database     *SolutionDatabase
	learningMode bool
	metrics      SolutionMetrics
	mu           sync.RWMutex
}

// SolutionMetrics tracks solution effectiveness
type SolutionMetrics struct {
	TotalSuggestions int64              `json:"total_suggestions"`
	AppliedSolutions int64              `json:"applied_solutions"`
	SuccessfulFixes  int64              `json:"successful_fixes"`
	SolutionRatings  map[string]float64 `json:"solution_ratings"`
	PopularSolutions map[string]int64   `json:"popular_solutions"`
	ErrorFrequency   map[string]int64   `json:"error_frequency"`
	LastUpdated      time.Time          `json:"last_updated"`
}

// NewSolutionDatabase creates a new solution database
func NewSolutionDatabase() *SolutionDatabase {
	return &SolutionDatabase{
		solutions: make(map[string][]Solution),
		matchers:  make([]SolutionMatcher, 0),
		providers: make([]SolutionProvider, 0),
	}
}

// NewSolutionEngine creates a new solution engine
func NewSolutionEngine() *SolutionEngine {
	engine := &SolutionEngine{
		database:     NewSolutionDatabase(),
		learningMode: true,
		metrics: SolutionMetrics{
			SolutionRatings:  make(map[string]float64),
			PopularSolutions: make(map[string]int64),
			ErrorFrequency:   make(map[string]int64),
			LastUpdated:      time.Now(),
		},
	}

	// Initialize with default solutions
	engine.initializeDefaultSolutions()

	return engine
}

// AddSolution adds a solution to the database
func (db *SolutionDatabase) AddSolution(errorCode string, solution Solution) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.solutions[errorCode] == nil {
		db.solutions[errorCode] = make([]Solution, 0)
	}

	db.solutions[errorCode] = append(db.solutions[errorCode], solution)

	// Sort by priority (higher first)
	sort.Slice(db.solutions[errorCode], func(i, j int) bool {
		return db.solutions[errorCode][i].Priority > db.solutions[errorCode][j].Priority
	})
}

// AddMatcher adds a solution matcher
func (db *SolutionDatabase) AddMatcher(matcher SolutionMatcher) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.matchers = append(db.matchers, matcher)

	// Sort by priority (higher first)
	sort.Slice(db.matchers, func(i, j int) bool {
		return db.matchers[i].Priority > db.matchers[j].Priority
	})
}

// RegisterProvider registers a solution provider
func (db *SolutionDatabase) RegisterProvider(provider SolutionProvider) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.providers = append(db.providers, provider)
}

// GetSolutions returns solutions for a given error
func (engine *SolutionEngine) GetSolutions(ctx context.Context, userErr *UserError) ([]Solution, error) {
	if userErr == nil {
		return nil, fmt.Errorf("error cannot be nil")
	}

	engine.mu.Lock()
	engine.metrics.TotalSuggestions++
	engine.metrics.ErrorFrequency[userErr.Code.String()]++
	engine.mu.Unlock()

	var allSolutions []Solution

	// Get solutions from database
	solutions := engine.getSolutionsFromDB(userErr)
	allSolutions = append(allSolutions, solutions...)

	// Get solutions from external providers
	for _, provider := range engine.database.providers {
		if providerSolutions, err := provider.GetSolutions(ctx, userErr); err == nil {
			allSolutions = append(allSolutions, providerSolutions...)
		}
	}

	// Apply matching logic
	filteredSolutions := engine.filterSolutions(userErr, allSolutions)

	// Sort by confidence and priority
	sort.Slice(filteredSolutions, func(i, j int) bool {
		if filteredSolutions[i].Confidence != filteredSolutions[j].Confidence {
			return filteredSolutions[i].Confidence > filteredSolutions[j].Confidence
		}
		return filteredSolutions[i].Priority > filteredSolutions[j].Priority
	})

	// Apply learning from metrics
	engine.applySolutionLearning(filteredSolutions)

	return filteredSolutions, nil
}

// getSolutionsFromDB retrieves solutions from the internal database
func (engine *SolutionEngine) getSolutionsFromDB(userErr *UserError) []Solution {
	engine.database.mu.RLock()
	defer engine.database.mu.RUnlock()

	// Direct code match
	if solutions, exists := engine.database.solutions[userErr.Code.String()]; exists {
		return solutions
	}

	// Try domain-category matches
	domainKey := fmt.Sprintf("%s_*", strings.ToUpper(userErr.Code.Domain))
	if solutions, exists := engine.database.solutions[domainKey]; exists {
		return solutions
	}

	return nil
}

// filterSolutions applies matchers to filter and score solutions
func (engine *SolutionEngine) filterSolutions(userErr *UserError, solutions []Solution) []Solution {
	var filtered []Solution

	for _, solution := range solutions {
		if engine.matchesSolution(userErr, solution) {
			// Calculate confidence based on various factors
			confidence := engine.calculateConfidence(userErr, solution)
			solution.Confidence = confidence
			filtered = append(filtered, solution)
		}
	}

	return filtered
}

// matchesSolution checks if a solution matches the error
func (engine *SolutionEngine) matchesSolution(userErr *UserError, solution Solution) bool {
	engine.database.mu.RLock()
	defer engine.database.mu.RUnlock()

	// Check matchers
	for _, matcher := range engine.database.matchers {
		if engine.matcherMatches(userErr, matcher) {
			return true
		}
	}

	// Check context matching
	if len(solution.Context) > 0 {
		for key, expectedValue := range solution.Context {
			if contextValue, exists := userErr.Context[key]; exists {
				if matched, _ := regexp.MatchString(expectedValue, fmt.Sprintf("%v", contextValue)); matched {
					return true
				}
			}
		}
	}

	return true // Default to including solution
}

// matcherMatches checks if a matcher applies to the error
func (engine *SolutionEngine) matcherMatches(userErr *UserError, matcher SolutionMatcher) bool {
	// Check error code match
	if matcher.ErrorCode != nil {
		if userErr.Code.Domain == matcher.ErrorCode.Domain &&
			userErr.Code.Category == matcher.ErrorCode.Category &&
			userErr.Code.Code == matcher.ErrorCode.Code {
			return true
		}
	}

	// Check domain regex
	if matcher.DomainRegex != "" {
		if matched, _ := regexp.MatchString(matcher.DomainRegex, userErr.Code.Domain); matched {
			return true
		}
	}

	// Check message regex
	if matcher.MessageRegex != "" {
		if matched, _ := regexp.MatchString(matcher.MessageRegex, userErr.Message); matched {
			return true
		}
	}

	// Check context keys
	if len(matcher.ContextKeys) > 0 {
		for _, key := range matcher.ContextKeys {
			if _, exists := userErr.Context[key]; exists {
				return true
			}
		}
	}

	return false
}

// calculateConfidence calculates solution confidence based on various factors
func (engine *SolutionEngine) calculateConfidence(userErr *UserError, solution Solution) float64 {
	confidence := solution.Confidence
	if confidence == 0 {
		confidence = 0.5 // Default confidence
	}

	// Boost confidence based on historical success
	engine.mu.RLock()
	if rating, exists := engine.metrics.SolutionRatings[solution.ID]; exists {
		confidence = (confidence + rating) / 2
	}

	// Boost popular solutions
	if popularity, exists := engine.metrics.PopularSolutions[solution.ID]; exists && popularity > 5 {
		confidence += 0.1
	}
	engine.mu.RUnlock()

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// applySolutionLearning applies machine learning insights to solutions
func (engine *SolutionEngine) applySolutionLearning(solutions []Solution) {
	if !engine.learningMode {
		return
	}

	// This would integrate with ML models in a real implementation
	// For now, we adjust based on simple heuristics

	engine.mu.Lock()
	engine.metrics.LastUpdated = time.Now()
	engine.mu.Unlock()
}

// RecordSolutionFeedback records user feedback on solution effectiveness
func (engine *SolutionEngine) RecordSolutionFeedback(solutionID string, rating float64, successful bool) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	// Update ratings with exponential moving average
	currentRating := engine.metrics.SolutionRatings[solutionID]
	alpha := 0.3
	engine.metrics.SolutionRatings[solutionID] = alpha*rating + (1-alpha)*currentRating

	// Track popularity
	engine.metrics.PopularSolutions[solutionID]++

	// Track success
	if successful {
		engine.metrics.SuccessfulFixes++
	}

	engine.metrics.LastUpdated = time.Now()
}

// ApplySolution attempts to automatically apply a solution
func (engine *SolutionEngine) ApplySolution(ctx context.Context, solution Solution) error {
	if !solution.Automated {
		return fmt.Errorf("solution %s is not automated", solution.ID)
	}

	engine.mu.Lock()
	engine.metrics.AppliedSolutions++
	engine.mu.Unlock()

	// Execute commands in order
	for _, cmd := range solution.Commands {
		if err := engine.executeCommand(ctx, cmd); err != nil {
			return fmt.Errorf("failed to execute command %s: %w", cmd.Name, err)
		}
	}

	return nil
}

// executeCommand executes a solution command
func (engine *SolutionEngine) executeCommand(ctx context.Context, cmd Command) error {
	if cmd.Dangerous {
		return fmt.Errorf("dangerous command %s requires manual confirmation", cmd.Name)
	}

	// This would execute the actual command in a real implementation
	// For now, we just simulate
	fmt.Printf("Executing: %s\n", cmd.Command)

	return nil
}

// LoadSolutionsFromFile loads solutions from a JSON file
func (engine *SolutionEngine) LoadSolutionsFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read solutions file: %w", err)
	}

	var solutions []Solution
	if err := json.Unmarshal(data, &solutions); err != nil {
		return fmt.Errorf("failed to parse solutions file: %w", err)
	}

	for _, solution := range solutions {
		// Determine error code from solution tags or context
		errorCode := engine.deriveErrorCodeFromSolution(solution)
		engine.database.AddSolution(errorCode, solution)
	}

	return nil
}

// LoadSolutionsFromDirectory loads all solution files from a directory
func (engine *SolutionEngine) LoadSolutionsFromDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			if err := engine.LoadSolutionsFromFile(path); err != nil {
				return fmt.Errorf("failed to load %s: %w", path, err)
			}
		}

		return nil
	})
}

// deriveErrorCodeFromSolution derives error code from solution metadata
func (engine *SolutionEngine) deriveErrorCodeFromSolution(solution Solution) string {
	// Try to extract from context
	if domain, ok := solution.Context["domain"]; ok {
		if category, ok := solution.Context["category"]; ok {
			if code, ok := solution.Context["code"]; ok {
				return fmt.Sprintf("%s_%s_%s",
					strings.ToUpper(domain),
					strings.ToUpper(category),
					strings.ToUpper(code))
			}
		}
	}

	// Try to extract from tags
	for _, tag := range solution.Tags {
		if strings.Contains(tag, "_") {
			parts := strings.Split(tag, "_")
			if len(parts) >= 3 {
				return strings.ToUpper(tag)
			}
		}
	}

	// Default to generic
	return "GENERIC_SOLUTION"
}

// GetMetrics returns solution engine metrics
func (engine *SolutionEngine) GetMetrics() SolutionMetrics {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	return engine.metrics
}

// initializeDefaultSolutions adds built-in solutions
func (engine *SolutionEngine) initializeDefaultSolutions() {
	// GitHub authentication solutions
	engine.database.AddSolution("GITHUB_AUTH_INVALID_TOKEN", Solution{
		ID:          "github-token-fix",
		Title:       "Fix GitHub Authentication",
		Description: "Resolve GitHub token authentication issues",
		Priority:    10,
		Confidence:  0.9,
		Steps: []SolutionStep{
			{Order: 1, Action: "Check token", Description: "Verify GITHUB_TOKEN environment variable is set"},
			{Order: 2, Action: "Validate permissions", Description: "Ensure token has repo and admin:org permissions"},
			{Order: 3, Action: "Test token", Description: "Test token with GitHub API", Command: "curl -H \"Authorization: token $GITHUB_TOKEN\" https://api.github.com/user"},
		},
		Links: []Link{
			{Title: "GitHub Token Documentation", URL: "https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token", Type: "docs"},
			{Title: "Token Permissions Guide", URL: "https://docs.github.com/en/developers/apps/building-oauth-apps/scopes-for-oauth-apps", Type: "docs"},
		},
		Tags: []string{"github", "authentication", "token"},
	})

	// Network timeout solutions
	engine.database.AddSolution("NETWORK_TIMEOUT_OPERATION_TIMEOUT", Solution{
		ID:          "network-timeout-fix",
		Title:       "Fix Network Timeout Issues",
		Description: "Resolve network timeout problems",
		Priority:    8,
		Confidence:  0.8,
		Steps: []SolutionStep{
			{Order: 1, Action: "Check connectivity", Description: "Test internet connection"},
			{Order: 2, Action: "Increase timeout", Description: "Use --timeout flag to increase timeout"},
			{Order: 3, Action: "Check firewall", Description: "Verify firewall settings allow connections"},
			{Order: 4, Action: "Try alternative DNS", Description: "Use 8.8.8.8 or 1.1.1.1 as DNS"},
		},
		Commands: []Command{
			{Name: "ping-test", Command: "ping -c 4 8.8.8.8", Description: "Test basic connectivity"},
			{Name: "dns-test", Command: "nslookup github.com", Description: "Test DNS resolution"},
		},
		Links: []Link{
			{Title: "Network Troubleshooting Guide", URL: "https://github.com/gizzahub/gzh-manager-go/wiki/Network-Troubleshooting", Type: "docs"},
		},
		Tags: []string{"network", "timeout", "connectivity"},
	})

	// File permission solutions
	engine.database.AddSolution("FILE_PERMISSION_ACCESS_DENIED", Solution{
		ID:          "file-permission-fix",
		Title:       "Fix File Permission Issues",
		Description: "Resolve file access permission problems",
		Priority:    9,
		Confidence:  0.85,
		Steps: []SolutionStep{
			{Order: 1, Action: "Check permissions", Description: "Check file/directory permissions", Command: "ls -la"},
			{Order: 2, Action: "Fix ownership", Description: "Change file ownership if needed", Command: "sudo chown $USER:$USER"},
			{Order: 3, Action: "Set permissions", Description: "Set appropriate permissions", Command: "chmod 755"},
		},
		Commands: []Command{
			{Name: "check-perms", Command: "ls -la", Description: "List permissions"},
			{Name: "fix-perms", Command: "chmod 755", Description: "Set read/write/execute permissions", Dangerous: true},
		},
		Tags: []string{"file", "permission", "access"},
	})

	// Configuration errors
	engine.database.AddSolution("CONFIG_VALIDATION_INVALID_FIELD", Solution{
		ID:          "config-validation-fix",
		Title:       "Fix Configuration Validation",
		Description: "Resolve configuration validation errors",
		Priority:    7,
		Confidence:  0.9,
		Steps: []SolutionStep{
			{Order: 1, Action: "Validate syntax", Description: "Check YAML/JSON syntax"},
			{Order: 2, Action: "Check schema", Description: "Validate against configuration schema"},
			{Order: 3, Action: "Use examples", Description: "Refer to example configurations"},
		},
		Commands: []Command{
			{Name: "validate-config", Command: "gz config validate", Description: "Validate configuration file"},
			{Name: "show-schema", Command: "gz config schema", Description: "Show configuration schema"},
		},
		Links: []Link{
			{Title: "Configuration Guide", URL: "https://github.com/gizzahub/gzh-manager-go/wiki/Configuration", Type: "docs"},
			{Title: "Configuration Examples", URL: "https://github.com/gizzahub/gzh-manager-go/tree/main/samples", Type: "tutorial"},
		},
		Tags:      []string{"config", "validation", "syntax"},
		Automated: true,
	})
}
