package log

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Korrnals/gotr/internal/paths"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func resetLoggerState(t *testing.T) {
	t.Helper()
	globalLogger = nil
	once = sync.Once{}
}

func TestDefaultConfig(t *testing.T) {
	resetLoggerState(t)

	logsDir, err := paths.LogsDirPath()
	if err != nil {
		t.Fatalf("LogsDirPath() error = %v", err)
	}

	cfg := DefaultConfig()
	if cfg.Level != "info" {
		t.Fatalf("expected default level info, got %q", cfg.Level)
	}
	if cfg.JSONFormat {
		t.Fatalf("expected JSONFormat=false by default")
	}
	if cfg.Development {
		t.Fatalf("expected Development=false by default")
	}
	if cfg.LogDir != logsDir {
		t.Fatalf("expected LogDir=%q, got %q", logsDir, cfg.LogDir)
	}
}

func TestCreateLoggerMkdirAllError(t *testing.T) {
	resetLoggerState(t)

	tmpDir := t.TempDir()
	asFile := filepath.Join(tmpDir, "already_file")
	if err := os.WriteFile(asFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := createLogger(Config{Level: "info", LogDir: asFile})
	if err == nil {
		t.Fatalf("expected createLogger error when LogDir points to file")
	}
}

func TestCreateLoggerAndWrappers(t *testing.T) {
	resetLoggerState(t)

	tmpDir := t.TempDir()
	cases := []Config{
		{
			Level:       "invalid-level",
			JSONFormat:  false,
			LogDir:      filepath.Join(tmpDir, "console"),
			Development: false,
		},
		{
			Level:       "debug",
			JSONFormat:  true,
			LogDir:      filepath.Join(tmpDir, "json-dev"),
			Development: true,
		},
	}

	for _, cfg := range cases {
		logger, err := createLogger(cfg)
		if err != nil {
			t.Fatalf("createLogger(%+v) error = %v", cfg, err)
		}
		if logger == nil {
			t.Fatalf("createLogger(%+v) returned nil logger", cfg)
		}

		globalLogger = logger
		if got := L(); got == nil {
			t.Fatalf("L() returned nil after initialization")
		}

		// Verify wrapper helpers execute without panic and return loggers.
		_ = WithField("k", "v")
		_ = WithFields(map[string]interface{}{"a": 1, "b": "x"})
		Debug("debug message", zap.String("test", "ok"))
		Info("info message", zap.String("test", "ok"))
		Warn("warn message", zap.String("test", "ok"))
		Error("error message", zap.String("test", "ok"))
	}
}

func TestLAndSyncWithoutInitialization(t *testing.T) {
	resetLoggerState(t)

	logger := L()
	if logger == nil {
		t.Fatalf("expected noop logger from L() when uninitialized")
	}

	// Noop logger should be safe for logging calls.
	logger.Info("noop")

	if err := Sync(); err != nil {
		t.Fatalf("Sync() expected nil for uninitialized logger, got %v", err)
	}
}

func TestInitRunsOnce(t *testing.T) {
	resetLoggerState(t)

	tmpDir := t.TempDir()
	cfg1 := Config{Level: "info", LogDir: filepath.Join(tmpDir, "first")}
	cfg2 := Config{Level: "debug", JSONFormat: true, LogDir: filepath.Join(tmpDir, "second"), Development: true}

	if err := Init(cfg1); err != nil {
		t.Fatalf("Init(cfg1) error = %v", err)
	}
	first := globalLogger
	if first == nil {
		t.Fatalf("globalLogger is nil after first Init")
	}

	if err := Init(cfg2); err != nil {
		t.Fatalf("Init(cfg2) error = %v", err)
	}
	if globalLogger != first {
		t.Fatalf("expected second Init to keep the original logger instance")
	}
}

func TestInitDefaultAndSync(t *testing.T) {
	resetLoggerState(t)

	if err := InitDefault(); err != nil {
		t.Fatalf("InitDefault() error = %v", err)
	}
	if globalLogger == nil {
		t.Fatalf("globalLogger is nil after InitDefault")
	}

	// Sync may return OS-dependent errors for stdout-backed writers.
	_ = Sync()
}

func TestFatal(t *testing.T) {
	resetLoggerState(t)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)
	globalLogger = zap.New(core, zap.WithFatalHook(zapcore.WriteThenPanic))

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from Fatal with WriteThenPanic hook")
		}
	}()

	Fatal("fatal test message")
}
