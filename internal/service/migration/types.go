// internal/service/migration/types.go
package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Korrnals/gotr/internal/client"
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
	importedCases int // number of successfully imported cases

	mapping  *SharedStepMapping // shared step ID mapping (see mapping.go)
	logger   *zap.SugaredLogger
	logFile  *os.File // log file handle, closed in Close()
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

	// Cores
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)
	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(fileWriter), zap.InfoLevel)

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
		if err := m.logger.Sync(); err != nil {
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
