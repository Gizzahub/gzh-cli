package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuditHistory represents a historical audit record
type AuditHistory struct {
	Timestamp    time.Time                   `json:"timestamp"`
	Organization string                      `json:"organization"`
	Summary      AuditSummary                `json:"summary"`
	PolicyStats  map[string]PolicyStatistics `json:"policy_stats"`
}

// AuditSummary provides overall compliance statistics
type AuditSummary struct {
	TotalRepositories     int     `json:"total_repositories"`
	CompliantRepositories int     `json:"compliant_repositories"`
	CompliancePercentage  float64 `json:"compliance_percentage"`
	TotalViolations       int     `json:"total_violations"`
	CriticalViolations    int     `json:"critical_violations"`
}

// PolicyStatistics tracks statistics for a specific policy
type PolicyStatistics struct {
	PolicyName           string  `json:"policy_name"`
	ViolationCount       int     `json:"violation_count"`
	CompliantRepos       int     `json:"compliant_repos"`
	ViolatingRepos       int     `json:"violating_repos"`
	CompliancePercentage float64 `json:"compliance_percentage"`
}

// AuditStore interface for storing and retrieving audit data
type AuditStore interface {
	SaveAuditResult(history *AuditHistory) error
	GetHistoricalData(org string, duration time.Duration) ([]AuditHistory, error)
	GetPolicyTrends(org, policy string, duration time.Duration) ([]PolicyStatistics, error)
}

// FileBasedAuditStore implements AuditStore using file system
type FileBasedAuditStore struct {
	basePath string
}

// NewFileBasedAuditStore creates a new file-based audit store
func NewFileBasedAuditStore(basePath string) (*FileBasedAuditStore, error) {
	if basePath == "" {
		// Default to user's config directory
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user config dir: %w", err)
		}
		basePath = filepath.Join(configDir, "gzh-manager", "audit-history")
	}

	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create audit history directory: %w", err)
	}

	return &FileBasedAuditStore{
		basePath: basePath,
	}, nil
}

// SaveAuditResult saves an audit result to the file system
func (s *FileBasedAuditStore) SaveAuditResult(history *AuditHistory) error {
	// Create organization directory
	orgPath := filepath.Join(s.basePath, history.Organization)
	if err := os.MkdirAll(orgPath, 0o755); err != nil {
		return fmt.Errorf("failed to create organization directory: %w", err)
	}

	// Create filename based on date
	filename := filepath.Join(orgPath, history.Timestamp.Format("2006-01-02")+".jsonl")

	// Open file in append mode
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	// Write JSON line
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(history); err != nil {
		return fmt.Errorf("failed to write audit record: %w", err)
	}

	return nil
}

// GetHistoricalData retrieves historical audit data for an organization
func (s *FileBasedAuditStore) GetHistoricalData(org string, duration time.Duration) ([]AuditHistory, error) {
	orgPath := filepath.Join(s.basePath, org)

	// Check if organization directory exists
	if _, err := os.Stat(orgPath); os.IsNotExist(err) {
		return []AuditHistory{}, nil // Return empty slice if no history exists
	}

	startDate := time.Now().Add(-duration)
	var allHistory []AuditHistory

	// Read files for the specified duration
	for d := startDate; !d.After(time.Now()); d = d.AddDate(0, 0, 1) {
		filename := filepath.Join(orgPath, d.Format("2006-01-02")+".jsonl")

		// Skip if file doesn't exist
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			continue
		}

		// Read file
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit file %s: %w", filename, err)
		}
		defer file.Close()

		// Decode JSON lines
		decoder := json.NewDecoder(file)
		for decoder.More() {
			var history AuditHistory
			if err := decoder.Decode(&history); err != nil {
				// Log error but continue reading
				continue
			}
			allHistory = append(allHistory, history)
		}
	}

	return allHistory, nil
}

// GetPolicyTrends retrieves policy-specific trends
func (s *FileBasedAuditStore) GetPolicyTrends(org, policy string, duration time.Duration) ([]PolicyStatistics, error) {
	history, err := s.GetHistoricalData(org, duration)
	if err != nil {
		return nil, err
	}

	var policyTrends []PolicyStatistics
	for _, h := range history {
		if stats, exists := h.PolicyStats[policy]; exists {
			policyTrends = append(policyTrends, stats)
		}
	}

	return policyTrends, nil
}
