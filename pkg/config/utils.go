package config

import (
	"os"
	"path/filepath"
)

// CreateDirectory creates a directory if it doesn't exist
func CreateDirectory(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// WriteFile writes content to a file
func WriteFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}