// internal/service/migration/types.go
package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Migration holds the migration context: client, parameters, mapping, and logger.
type Migration struct {
	Client        client.ClientInterface // API client interface
	srcProject    int64
	srcSuite      int64
	dstProject    int64
	dstSuite      int64
	compareField  string
	importedCases int // number of successfully imported entities (cases or sections)
	failedImports int // number of entities that could not be imported (e.g. unresolved section/parent)

	mapping  *SharedStepMapping // shared step ID mapping (see mapping.go)
	logger   *zap.SugaredLogger
	logFile  *os.File // log file handle, closed in Close()

	lastFilteredSteps data.GetSharedStepsResponse // filtered shared steps from last MigrateSharedSteps run
	lastFilterStats   FilterStats                 // statistics from the last Filter* call
}

// NewMigration creates a new Migration instance with a zap logger.
func NewMigration(cli client.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*Migration, error) {
	// Create directory for log files
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// JSON log file path
	logFile := filepath.Join(logDir, fmt.Sprintf("migration_%s.json", time.Now().Format("2006-01-02_15-04-05")))
	fileWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	// Encoders
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// Cores: console shows only warnings+, file gets everything from info+
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.WarnLevel)
	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(fileWriter), zap.DebugLevel)

	core := zapcore.NewTee(consoleCore, fileCore)
	logger := zap.New(core, zap.AddCaller()).Sugar()

	m := &Migration{
		Client:        cli,
		srcProject:    srcProject,
		srcSuite:      srcSuite,
		dstProject:    dstProject,
		dstSuite:      dstSuite,
		compareField:  compareField,
		importedCases: 0,
		mapping:       NewSharedStepMapping(srcProject, dstProject), // from mapping.go
		logger:        logger,
		logFile:       fileWriter,
	}

	m.logger.Info("Migration object created", "log_file", logFile)
	return m, nil
}

// Close shuts down the migration, flushing log buffers to disk.
func (m *Migration) Close() error {
	if m.logger != nil {
		if err := m.logger.Sync(); err != nil && !isSyncIgnorable(err) {
			fmt.Fprintf(os.Stderr, "warning: failed to flush migration log: %v\n", err)
		}
	}
	if m.logFile != nil {
		if err := m.logFile.Close(); err != nil {
			return fmt.Errorf("failed to close migration log file: %w", err)
		}
	}
	return nil
}

// isSyncIgnorable returns true for errors expected when syncing non-file
// descriptors (stdout/stderr). These are harmless and should be suppressed.
func isSyncIgnorable(err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, syscall.EINVAL) || errors.Is(pathErr.Err, syscall.ENOTSUP)
	}
	return false
}

// FilteredSharedSteps returns the filtered shared steps from the last MigrateSharedSteps run.
func (m *Migration) FilteredSharedSteps() data.GetSharedStepsResponse {
	return m.lastFilteredSteps
}

// Mapping returns a simple map[sourceID]=targetID for external use
func (m *Migration) Mapping() map[int64]int64 {
	res := make(map[int64]int64)
	if m.mapping == nil {
		return res
	}
	for k, v := range m.mapping.index {
		res[k] = v
	}
	return res
}

// ImportedCount returns the number of entities successfully imported in the
// last ImportCases/ImportSections call (cumulative across calls on this
// Migration instance).
func (m *Migration) ImportedCount() int { return m.importedCases }

// FailedCount returns the number of entities that failed to import because of
// unresolved scope (missing destination section, missing mapped parent) or API
// errors. This counter makes otherwise-silent skips visible in the summary.
func (m *Migration) FailedCount() int { return m.failedImports }
