package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileStore implements ChangeStore using local file storage.
type FileStore struct {
	basePath string
}

// NewFileStore creates a new file-based change store.
func NewFileStore(basePath string) (*FileStore, error) {
	// Create base directory if it doesn't exist
	err := os.MkdirAll(basePath, 0o755)
	if err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileStore{
		basePath: basePath,
	}, nil
}

// Store saves a change record to a file.
func (fs *FileStore) Store(ctx context.Context, record *ChangeRecord) error {
	// Create directory structure: basePath/year/month/day
	date := record.Timestamp
	dirPath := filepath.Join(fs.basePath,
		fmt.Sprintf("%d", date.Year()),
		fmt.Sprintf("%02d", date.Month()),
		fmt.Sprintf("%02d", date.Day()))

	err := os.MkdirAll(dirPath, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// File name format: changeID.json
	filePath := filepath.Join(dirPath, record.ID+".json")

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal change record: %w", err)
	}

	err = os.WriteFile(filePath, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write change record: %w", err)
	}

	return nil
}

// Get retrieves a change record by ID.
func (fs *FileStore) Get(ctx context.Context, id string) (*ChangeRecord, error) {
	// Search for the file across all directories
	filePath, err := fs.findRecordFile(id)
	if err != nil {
		return nil, fmt.Errorf("change record not found: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read change record: %w", err)
	}

	var record ChangeRecord

	err = json.Unmarshal(data, &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal change record: %w", err)
	}

	return &record, nil
}

// List retrieves change records based on filter criteria.
func (fs *FileStore) List(ctx context.Context, filter ChangeFilter) ([]*ChangeRecord, error) {
	var records []*ChangeRecord

	// Walk through all files and filter
	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err // Skip files that can't be read
		}

		var record ChangeRecord

		err = json.Unmarshal(data, &record)
		if err != nil {
			return err // Skip invalid files
		}

		// Apply filters
		if fs.matchesFilter(&record, filter) {
			records = append(records, &record)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort by timestamp (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	// Apply limit and offset
	if filter.Offset > 0 && filter.Offset < len(records) {
		records = records[filter.Offset:]
	}

	if filter.Limit > 0 && filter.Limit < len(records) {
		records = records[:filter.Limit]
	}

	return records, nil
}

// Delete removes a change record.
func (fs *FileStore) Delete(ctx context.Context, id string) error {
	filePath, err := fs.findRecordFile(id)
	if err != nil {
		return fmt.Errorf("change record not found: %w", err)
	}

	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete change record: %w", err)
	}

	return nil
}

// findRecordFile searches for a record file by ID across all directories.
func (fs *FileStore) findRecordFile(id string) (string, error) {
	var foundPath string

	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == id+".json" {
			foundPath = path
			return filepath.SkipDir // Stop searching
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("record file not found for ID: %s", id)
	}

	return foundPath, nil
}

// matchesFilter checks if a record matches the given filter criteria.
func (fs *FileStore) matchesFilter(record *ChangeRecord, filter ChangeFilter) bool {
	if filter.Organization != "" && record.Organization != filter.Organization {
		return false
	}

	if filter.Repository != "" && record.Repository != filter.Repository {
		return false
	}

	if filter.User != "" && record.User != filter.User {
		return false
	}

	if filter.Operation != "" && record.Operation != filter.Operation {
		return false
	}

	if filter.Category != "" && record.Category != filter.Category {
		return false
	}

	if !filter.Since.IsZero() && record.Timestamp.Before(filter.Since) {
		return false
	}

	if !filter.Until.IsZero() && record.Timestamp.After(filter.Until) {
		return false
	}

	return true
}

// GetStorePath returns the base storage path.
func (fs *FileStore) GetStorePath() string {
	return fs.basePath
}

// GetStats returns storage statistics.
func (fs *FileStore) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	totalFiles := 0
	totalSize := int64(0)

	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".json") {
			totalFiles++
			totalSize += info.Size()
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate stats: %w", err)
	}

	stats["total_records"] = totalFiles
	stats["total_size_bytes"] = totalSize
	stats["storage_path"] = fs.basePath

	return stats, nil
}
