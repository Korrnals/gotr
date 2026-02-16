// internal/service/run.go
// Сервис для работы с test runs
package service

import (
	"errors"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runClientInterface определяет методы клиента, необходимые для RunService
type runClientInterface interface {
	GetRun(runID int64) (*data.Run, error)
	GetRuns(projectID int64) (data.GetRunsResponse, error)
	AddRun(projectID int64, req *data.AddRunRequest) (*data.Run, error)
	UpdateRun(runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	CloseRun(runID int64) (*data.Run, error)
	DeleteRun(runID int64) error
}

// RunService предоставляет методы для работы с test runs
type RunService struct {
	client runClientInterface
}

// NewRunService создаёт новый сервис для работы с runs
func NewRunService(client *client.HTTPClient) *RunService {
	return &RunService{client: client}
}

// NewRunServiceFromInterface создаёт сервис из клиента-интерфейса (для тестов)
func NewRunServiceFromInterface(cli client.ClientInterface) *RunService {
	return &RunService{client: cli}
}

// Get получает информацию о test run по ID
func (s *RunService) Get(runID int64) (*data.Run, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	return s.client.GetRun(runID)
}

// GetByProject получает список runs для проекта
func (s *RunService) GetByProject(projectID int64) (data.GetRunsResponse, error) {
	if err := s.validateID(projectID, "project_id"); err != nil {
		return nil, err
	}
	return s.client.GetRuns(projectID)
}

// Create создаёт новый test run с валидацией параметров
func (s *RunService) Create(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	log.L().Info("creating test run",
		zap.Int64("project_id", projectID),
		zap.String("name", req.Name),
		zap.Int64("suite_id", req.SuiteID),
	)

	if err := s.validateID(projectID, "project_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "project_id"), zap.Error(err))
		return nil, err
	}
	if err := s.validateCreateRequest(req); err != nil {
		log.L().Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("валидация запроса: %w", err)
	}

	run, err := s.client.AddRun(projectID, req)
	if err != nil {
		log.L().Error("failed to create run", zap.Int64("project_id", projectID), zap.Error(err))
		return nil, err
	}

	log.L().Info("test run created",
		zap.Int64("run_id", run.ID),
		zap.Int64("project_id", run.ProjectID),
		zap.String("name", run.Name),
	)
	return run, nil
}

// Update обновляет существующий test run с валидацией
func (s *RunService) Update(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	log.L().Info("updating test run", zap.Int64("run_id", runID))

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return nil, err
	}

	run, err := s.client.UpdateRun(runID, req)
	if err != nil {
		log.L().Error("failed to update run", zap.Int64("run_id", runID), zap.Error(err))
		return nil, err
	}

	log.L().Info("test run updated",
		zap.Int64("run_id", run.ID),
		zap.String("name", run.Name),
	)
	return run, nil
}

// Close закрывает test run с валидацией ID
func (s *RunService) Close(runID int64) (*data.Run, error) {
	log.L().Info("closing test run", zap.Int64("run_id", runID))

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return nil, err
	}

	run, err := s.client.CloseRun(runID)
	if err != nil {
		log.L().Error("failed to close run", zap.Int64("run_id", runID), zap.Error(err))
		return nil, err
	}

	log.L().Info("test run closed",
		zap.Int64("run_id", run.ID),
		zap.Int64("completed_on", run.CompletedOn),
	)
	return run, nil
}

// Delete удаляет test run с валидацией ID
func (s *RunService) Delete(runID int64) error {
	log.L().Warn("deleting test run", zap.Int64("run_id", runID))

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return err
	}

	if err := s.client.DeleteRun(runID); err != nil {
		log.L().Error("failed to delete run", zap.Int64("run_id", runID), zap.Error(err))
		return err
	}

	log.L().Warn("test run deleted", zap.Int64("run_id", runID))
	return nil
}

// Output выводит результат в JSON и сохраняет в файл (если указан --output)
func (s *RunService) Output(cmd *cobra.Command, data interface{}) error {
	return utils.OutputResult(cmd, data)
}

// PrintSuccess выводит сообщение об успехе
func (s *RunService) PrintSuccess(cmd *cobra.Command, format string, args ...interface{}) {
	utils.PrintSuccess(cmd, format, args...)
}

// ParseID парсит ID из аргументов команды
func (s *RunService) ParseID(args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("отсутствует аргумент с ID на позиции %d", index)
	}
	return utils.ParseID(args[index])
}

// validateID проверяет что ID положительный
func (s *RunService) validateID(id int64, fieldName string) error {
	if id <= 0 {
		return fmt.Errorf("%s должен быть положительным числом, получено: %d", fieldName, id)
	}
	return nil
}

// validateCreateRequest валидирует параметры создания run
func (s *RunService) validateCreateRequest(req *data.AddRunRequest) error {
	if req == nil {
		return errors.New("запрос не может быть nil")
	}
	if req.Name == "" {
		return errors.New("название run (name) обязательно")
	}
	if req.SuiteID <= 0 {
		return errors.New("suite_id должен быть положительным числом")
	}
	return nil
}
