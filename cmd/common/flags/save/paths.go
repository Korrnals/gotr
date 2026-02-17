package save

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetExportsDir returns the exports directory path for a given resource.
// Pattern: ~/.gotr/exports/{resource}/
func GetExportsDir(resource string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".gotr", "exports", resource), nil
}

// EnsureDir ensures that the given directory exists, creating it if necessary.
// Creates parent directories as needed with permissions 0755.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", path, err)
	}
	return nil
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetHomeDir returns the user's home directory.
// Wrapper around os.UserHomeDir for testability.
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}
