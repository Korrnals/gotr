// internal/log/logger.go
// Package log provides centralized logging for gotr.
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Korrnals/gotr/internal/paths"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Config holds logger configuration.
type Config struct {
	Level       string // debug, info, warn, error
	JSONFormat  bool   // true = JSON, false = console
	LogDir      string // directory for log files
	Development bool   // development mode (line numbers, etc)
}

// DefaultConfig returns the default logger configuration.
func DefaultConfig() Config {
	logsDir, _ := paths.LogsDirPath()
	return Config{
		Level:       "info",
		JSONFormat:  false,
		LogDir:      logsDir,
		Development: false,
	}
}

// Init initializes the global logger.
func Init(cfg Config) error {
	var initErr error
	once.Do(func() {
		globalLogger, initErr = createLogger(cfg)
	})
	return initErr
}

// InitDefault initializes the logger with default settings.
func InitDefault() error {
	return Init(DefaultConfig())
}

// L returns the global logger.
func L() *zap.Logger {
	if globalLogger == nil {
		// Fallback to noop logger if not initialized
		return zap.NewNop()
	}
	return globalLogger
}

// Sync flushes logger buffers.
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// createLogger creates a zap.Logger with the given configuration.
func createLogger(cfg Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create log directory if needed
	if cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Select encoder
	var encoder zapcore.Encoder
	if cfg.JSONFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create writers
	var writers []zapcore.WriteSyncer

	// stdout
	writers = append(writers, zapcore.AddSync(os.Stdout))

	// Log file
	if cfg.LogDir != "" {
		logFile := filepath.Join(cfg.LogDir, fmt.Sprintf("gotr_%s.log", time.Now().Format("2006-01-02")))
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writers = append(writers, zapcore.AddSync(file))
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(writers...),
		level,
	)

	// Options
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.Development {
		options = append(options, zap.Development())
	}

	return zap.New(core, options...), nil
}

// WithField adds a field to the logger.
func WithField(key string, value interface{}) *zap.Logger {
	return L().With(zap.Any(key, value))
}

// WithFields adds multiple fields to the logger.
func WithFields(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return L().With(zapFields...)
}

// Debug logs at debug level.
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Info logs at info level.
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Warn logs at warn level.
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Error logs at error level.
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Fatal logs at fatal level and terminates the program.
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}
