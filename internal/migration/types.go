// internal/migration/types.go
package migration

import (
	"fmt"
	"gotr/internal/models/data"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Migration — контекст миграции (клиент, параметры, mapping, логгер)
type Migration struct {
	Client        ClientInterface // интерфейс клиента
	srcProject    int64
	srcSuite      int64
	dstProject    int64
	dstSuite      int64
	compareField  string
	importedCases int // количество успешно импортированных кейсов

	mapping *SharedStepMapping // mapping shared steps (из mapping.go)
	logger  *zap.SugaredLogger
}

// ClientInterface — интерфейс для API-клиента (для mock в тестах)
type ClientInterface interface {
	AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	AddCase(suiteID int64, req *data.AddCaseRequest) (*data.Case, error)
	AddSection(projectID int64, req *data.AddSectionRequest) (*data.Section, error)

	GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error)
	GetSuites(projectID int64) (data.GetSuitesResponse, error)
	GetCases(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error)
	GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error)
}

// NewMigration — конструктор с zap-логгером
func NewMigration(client ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*Migration, error) {
	// Создаём директорию для логов
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию лога %s: %w", logDir, err)
	}

	// Файл для JSON-лога
	logFile := filepath.Join(logDir, fmt.Sprintf("migration_%s.json", time.Now().Format("2006-01-02_15-04-05")))
	fileWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Энкодеры
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// Cores
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)
	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(fileWriter), zap.InfoLevel)

	core := zapcore.NewTee(consoleCore, fileCore)
	logger := zap.New(core, zap.AddCaller()).Sugar()

	m := &Migration{
		Client:        client,
		srcProject:    srcProject,
		srcSuite:      srcSuite,
		dstProject:    dstProject,
		dstSuite:      dstSuite,
		compareField:  compareField,
		importedCases: 0,
		mapping:       NewSharedStepMapping(srcProject, dstProject), // из mapping.go
		logger:        logger,
	}

	m.logger.Info("Создан объект миграции", "log_file", logFile)
	return m, nil
}

// Close — завершает работу миграции, сбрасывая логи на диск
func (m *Migration) Close() error {
	if m.logger != nil {
		_ = m.logger.Sync() // Сброс буфера zap
	}
	return nil
}

// Mapping возвращает простую map[sourceID]=targetID для внешнего использования
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
