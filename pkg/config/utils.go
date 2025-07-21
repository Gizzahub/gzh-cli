// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

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
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)

	return err
}

// GenerateTimestamp generates a timestamp string for file naming.
func GenerateTimestamp() string {
	return time.Now().Format("20060102-150405")
}
