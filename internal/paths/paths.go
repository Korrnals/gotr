// internal/paths/paths.go
// Package paths provides centralized path management for gotr.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DirName is the base directory name for gotr.
	DirName = ".gotr"

	// Subdirectories
	ConfigDir   = "config"   // configuration
	LogsDir     = "logs"     // runtime logs
	SelftestDir = "selftest" // self-test reports
	CacheDir    = "cache"    // API cache
	ExportsDir  = "exports"  // user data exports
	TempDir     = "temp"     // temporary files
)

// BaseDir returns the path to ~/.gotr.
func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, DirName), nil
}

// ConfigDirPath returns the path to ~/.gotr/config.
func ConfigDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ConfigDir), nil
}

// LogsDirPath returns the path to ~/.gotr/logs.
func LogsDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, LogsDir), nil
}

// EnsureLogsDirPath returns ~/.gotr/logs and creates it when missing.
func EnsureLogsDirPath() (string, error) {
	dir, err := LogsDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create logs directory: %w", err)
	}
	return dir, nil
}

// SelftestDirPath returns the path to ~/.gotr/selftest.
func SelftestDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, SelftestDir), nil
}

// CacheDirPath returns the path to ~/.gotr/cache.
func CacheDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, CacheDir), nil
}

// ExportsDirPath returns the path to ~/.gotr/exports.
func ExportsDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ExportsDir), nil
}

// TempDirPath returns the path to ~/.gotr/temp.
func TempDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, TempDir), nil
}

// ConfigFile returns the path to the main config file ~/.gotr/config/default.yaml.
func ConfigFile() (string, error) {
	dir, err := ConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "default.yaml"), nil
}

// EnsureAllDirs creates all required directories.
func EnsureAllDirs() error {
	dirs := []func() (string, error){
		ConfigDirPath,
		LogsDirPath,
		SelftestDirPath,
		CacheDirPath,
		ExportsDirPath,
		TempDirPath,
	}

	for _, dirFunc := range dirs {
		dir, err := dirFunc()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}
	return nil
}

// EnsureDir creates a specific directory.
func EnsureDir(dirFunc func() (string, error)) error {
	dir, err := dirFunc()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}
