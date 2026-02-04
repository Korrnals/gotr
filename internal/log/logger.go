// internal/log/logger.go
// Пакет для централизованного логирования gotr
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Config конфигурация логгера
type Config struct {
	Level       string // debug, info, warn, error
	JSONFormat  bool   // true = JSON, false = console
	LogDir      string // директория для логов
	Development bool   // development mode (line numbers, etc)
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		Level:       "info",
		JSONFormat:  false,
		LogDir:      filepath.Join(homeDir, ".testrail", "logs"),
		Development: false,
	}
}

// Init инициализирует глобальный логгер
func Init(cfg Config) error {
	var initErr error
	once.Do(func() {
		globalLogger, initErr = createLogger(cfg)
	})
	return initErr
}

// InitDefault инициализирует логгер с настройками по умолчанию
func InitDefault() error {
	return Init(DefaultConfig())
}

// L возвращает глобальный логгер
func L() *zap.Logger {
	if globalLogger == nil {
		// Fallback на noop logger если не инициализирован
		return zap.NewNop()
	}
	return globalLogger
}

// Sync сбрасывает буферы логгера
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// createLogger создаёт zap.Logger с заданной конфигурацией
func createLogger(cfg Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Создаём директорию для логов если нужно
	if cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
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

	// Выбираем encoder
	var encoder zapcore.Encoder
	if cfg.JSONFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Создаём writers
	var writers []zapcore.WriteSyncer

	//stdout
	writers = append(writers, zapcore.AddSync(os.Stdout))

	// Файл логов
	if cfg.LogDir != "" {
		logFile := filepath.Join(cfg.LogDir, fmt.Sprintf("gotr_%s.log", time.Now().Format("2006-01-02")))
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writers = append(writers, zapcore.AddSync(file))
	}

	// Создаём core
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(writers...),
		level,
	)

	// Опции
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.Development {
		options = append(options, zap.Development())
	}

	return zap.New(core, options...), nil
}

// WithField добавляет поле к логгеру
func WithField(key string, value interface{}) *zap.Logger {
	return L().With(zap.Any(key, value))
}

// WithFields добавляет несколько полей к логгеру
func WithFields(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return L().With(zapFields...)
}

// Debug логирует на уровне debug
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Info логирует на уровне info
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Warn логирует на уровне warn
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Error логирует на уровне error
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Fatal логирует на уровне fatal и завершает программу
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}
