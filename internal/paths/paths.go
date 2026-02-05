// internal/paths/paths.go
// Централизованное управление путями gotr
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DirName базовая директория gotr
	DirName = ".gotr"

	// Поддиректории
	ConfigDir   = "config"   // Конфигурация
	LogsDir     = "logs"     // Логи работы
	SelftestDir = "selftest" // Отчёты самотестирования
	CacheDir    = "cache"    // Кэш API
	ExportsDir  = "exports"  // Экспорт пользовательских данных
	TempDir     = "temp"     // Временные файлы
)

// BaseDir возвращает путь к ~/.testrail
func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, DirName), nil
}

// ConfigDirPath возвращает путь к ~/.testrail/config
func ConfigDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ConfigDir), nil
}

// LogsDirPath возвращает путь к ~/.testrail/logs
func LogsDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, LogsDir), nil
}

// SelftestDirPath возвращает путь к ~/.testrail/selftest
func SelftestDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, SelftestDir), nil
}

// CacheDirPath возвращает путь к ~/.testrail/cache
func CacheDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, CacheDir), nil
}

// ExportsDirPath возвращает путь к ~/.testrail/exports
func ExportsDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ExportsDir), nil
}

// TempDirPath возвращает путь к ~/.testrail/temp
func TempDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, TempDir), nil
}

// ConfigFile возвращает путь к основному конфигу ~/.gotr/config/default.yaml
func ConfigFile() (string, error) {
	dir, err := ConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "default.yaml"), nil
}

// EnsureAllDirs создаёт все необходимые директории
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

// EnsureDir создаёт конкретную директорию
func EnsureDir(dirFunc func() (string, error)) error {
	dir, err := dirFunc()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}
