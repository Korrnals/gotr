// internal/paths/paths.go
// Package paths provides centralized path management for gotr.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DirName is the base directory name for gotr.
	DirName = ".gotr"

	// HomeEnvVar names the environment variable that overrides the default
	// gotr base directory. When set to a non-empty value it is used verbatim
	// as the base path instead of "$HOME/.gotr". Primarily intended for
	// tests and ephemeral CLI runs that must not touch the operator's
	// real ~/.gotr/.
	HomeEnvVar = "GOTR_HOME"

	// Subdirectories
	ConfigDir   = "config"   // configuration
	LogsDir     = "logs"     // runtime logs
	ReportsDir  = "reports"  // migration reports
	SelftestDir = "selftest" // self-test reports
	CacheDir    = "cache"    // API cache
	ExportsDir  = "exports"  // user data exports
	TempDir     = "temp"     // temporary files
	SnapsDir    = "snaps"    // pre-mutation snapshots

	// Categorized subdirectories under ExportsDir. Introduced in v3.3.0
	// to replace the flat ~/.gotr/exports/ layout.
	ExportsSnapsSubdir   = "snaps"   // portable snapshot bundles (.tar.gz)
	ExportsReportsSubdir = "reports" // exported migration report bundles (.zip / single files)
	ExportsAPISubdir     = "api"     // raw API responses saved via `gotr export <resource>`
	ExportsAllSubdir     = "all"     // full migration bundles containing all artefacts (snaps+reports+logs+manifests)
)

// BaseDir returns the path to ~/.gotr.
//
// If the GOTR_HOME environment variable is set to a non-empty value it is
// returned verbatim, allowing callers (notably test suites) to redirect
// every gotr-managed path to a sandbox.
func BaseDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv(HomeEnvVar)); override != "" {
		return override, nil
	}
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

// ReportsDirPath returns the path to ~/.gotr/reports.
func ReportsDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ReportsDir), nil
}

// EnsureLogsDirPath returns ~/.gotr/logs and creates it when missing.
func EnsureLogsDirPath() (string, error) {
	dir, err := LogsDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
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

// SnapsDirPath returns the path to ~/.gotr/snaps.
func SnapsDirPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, SnapsDir), nil
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
		ReportsDirPath,
		SelftestDirPath,
		CacheDirPath,
		ExportsDirPath,
		TempDirPath,
		SnapsDirPath,
	}

	for _, dirFunc := range dirs {
		dir, err := dirFunc()
		if err != nil {
			return fmt.Errorf("EnsureAllDirs: %w", err)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}
	return nil
}

// EnsureDir creates a specific directory.
func EnsureDir(dirFunc func() (string, error)) error {
	dir, err := dirFunc()
	if err != nil {
		return fmt.Errorf("EnsureDir: %w", err)
	}
	return os.MkdirAll(dir, 0o755)
}

// EnsureReportsDirPath returns ~/.gotr/reports and creates it when missing.
func EnsureReportsDirPath() (string, error) {
	dir, err := ReportsDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create reports directory: %w", err)
	}
	return dir, nil
}

// EnsureExportsDirPath returns ~/.gotr/exports and creates it when missing.
func EnsureExportsDirPath() (string, error) {
	dir, err := ExportsDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create exports directory: %w", err)
	}
	return dir, nil
}

// ExportsSnapsDirPath returns ~/.gotr/exports/snaps — the canonical v3.3.0
// destination for portable snapshot bundles produced by `gotr export snap`.
func ExportsSnapsDirPath() (string, error) {
	base, err := ExportsDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ExportsSnapsSubdir), nil
}

// ExportsReportsDirPath returns ~/.gotr/exports/reports — the canonical v3.3.0
// destination for exported migration report bundles and single-file report
// copies produced by `gotr export report`.
func ExportsReportsDirPath() (string, error) {
	base, err := ExportsDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ExportsReportsSubdir), nil
}

// ExportsAPIDirPath returns ~/.gotr/exports/api — the canonical v3.3.0
// destination for raw API responses saved via `gotr export <resource> --save`.
// Individual resource subdirectories live beneath this path
// (e.g. ~/.gotr/exports/api/cases/).
func ExportsAPIDirPath() (string, error) {
	base, err := ExportsDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ExportsAPISubdir), nil
}

// ExportsAllDirPath returns ~/.gotr/exports/all — destination for full
// migration bundles produced by `gotr export migration-archive` (default
// "everything" mode). Such archives contain snapshots, reports, logs and
// manifests, so they live in their own bucket separate from per-snap
// bundles in exports/snaps.
func ExportsAllDirPath() (string, error) {
	base, err := ExportsDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ExportsAllSubdir), nil
}

// EnsureExportsSnapsDirPath returns ~/.gotr/exports/snaps, creating it when missing.
func EnsureExportsSnapsDirPath() (string, error) {
	return ensureUnderExports(ExportsSnapsDirPath, "exports/snaps")
}

// EnsureExportsReportsDirPath returns ~/.gotr/exports/reports, creating it when missing.
func EnsureExportsReportsDirPath() (string, error) {
	return ensureUnderExports(ExportsReportsDirPath, "exports/reports")
}

// EnsureExportsAPIDirPath returns ~/.gotr/exports/api, creating it when missing.
func EnsureExportsAPIDirPath() (string, error) {
	return ensureUnderExports(ExportsAPIDirPath, "exports/api")
}

// EnsureExportsAllDirPath returns ~/.gotr/exports/all, creating it when missing.
func EnsureExportsAllDirPath() (string, error) {
	return ensureUnderExports(ExportsAllDirPath, "exports/all")
}

func ensureUnderExports(pathFn func() (string, error), label string) (string, error) {
	dir, err := pathFn()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create %s directory: %w", label, err)
	}
	return dir, nil
}

// RelToHome converts an absolute path to a portable ~/... form when the path
// is under the current user's home directory. Paths outside $HOME are returned
// unchanged. This is used to make exported manifests and PDF reports portable
// across machines where the absolute $HOME differs.
//
// Examples:
//
//	/home/alice/.gotr/reports/r.pdf  -> ~/.gotr/reports/r.pdf
//	/var/log/app.log                 -> /var/log/app.log
//	C:\Users\Bob\.gotr\s\x.json      -> ~/.gotr/s/x.json (on Windows)
func RelToHome(abs string) string {
	if abs == "" {
		return abs
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return abs
	}
	cleanAbs := filepath.Clean(abs)
	cleanHome := filepath.Clean(home)
	if cleanAbs == cleanHome {
		return "~"
	}
	// Use filepath.Rel to avoid false prefix matches (e.g. /home/al vs /home/alice).
	rel, err := filepath.Rel(cleanHome, cleanAbs)
	if err != nil {
		return abs
	}
	// If rel starts with ".." the path is outside $HOME — return as-is.
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return abs
	}
	// Always use forward slashes in the portable form.
	return "~/" + filepath.ToSlash(rel)
}

// ExpandHome is the inverse of RelToHome: it expands a leading "~/" (or bare
// "~") to the current user's home directory. Paths that do not begin with "~"
// are returned unchanged. This lets tooling accept portable paths from
// imported bundles and resolve them on the local machine.
func ExpandHome(p string) (string, error) {
	if p == "" {
		return p, nil
	}
	if p != "~" && !strings.HasPrefix(p, "~/") && !strings.HasPrefix(p, `~\`) {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	if p == "~" {
		return home, nil
	}
	// Strip the "~/" or "~\\" prefix and join with OS-native separator.
	return filepath.Join(home, filepath.FromSlash(p[2:])), nil
}
