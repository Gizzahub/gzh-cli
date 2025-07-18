package utils

import (
	"os"
	"path/filepath"
)

// FileExists checks if a file exists.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// DirExists checks if a directory exists.
func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(dir string) error {
	if !DirExists(dir) {
		return os.MkdirAll(dir, 0o755)
	}

	return nil
}

// GetDirectories returns a list of directory names in the given path.
func GetDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// GetFiles returns a list of file names in the given path.
func GetFiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// FindFileInParents searches for a file in the current directory and parent directories.
func FindFileInParents(startPath, filename string) (string, error) {
	current := startPath

	for {
		candidate := filepath.Join(current, filename)
		if FileExists(candidate) {
			return candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			break
		}

		current = parent
	}

	return "", os.ErrNotExist
}
