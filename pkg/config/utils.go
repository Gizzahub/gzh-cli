package config

import (
	"io"
	"os"
	"time"
)

// CreateDirectory creates a directory if it doesn't exist.
func CreateDirectory(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// WriteFile writes content to a file.
func WriteFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0o644)
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CopyFile copies a file from source to destination.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }() //nolint:errcheck // Deferred close

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }() //nolint:errcheck // Deferred close

	_, err = io.Copy(destFile, sourceFile)

	return err
}

// GenerateTimestamp generates a timestamp string for file naming.
func GenerateTimestamp() string {
	return time.Now().Format("20060102-150405")
}
