// internal/service/run.go
// Сервис для работы с test runs
package service

import (
	"context"
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
	GetRun(ctx context.Context, runID int64) (*data.Run, error)
	GetRuns(ctx context.Context, projectID int64) (data.GetRunsResponse, error)
	AddRun(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error)
	UpdateRun(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	CloseRun(ctx context.Context, runID int64) (*data.Run, error)
	DeleteRun(ctx context.Context, runID int64) error
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
func (s *RunService) Get(ctx context.Context, runID int64) (*data.Run, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	return s.client.GetRun(ctx, runID)
}

// GetByProject получает список runs for project
func (s *RunService) GetByProject(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
	if err := s.validateID(projectID, "project_id"); err != nil {
		return nil, err
	}
	return s.client.GetRuns(ctx, projectID)
}

// Create создаёт новый test run с валидацией параметров
func (s *RunService) Create(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
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
		return nil, fmt.Errorf("request validation: %w", err)
	}

	run, err := s.client.AddRun(ctx, projectID, req)
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
func (s *RunService) Update(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	log.L().Info("updating test run", zap.Int64("run_id", runID))

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return nil, err
	}

	run, err := s.client.UpdateRun(ctx, runID, req)
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
func (s *RunService) Close(ctx context.Context, runID int64) (*data.Run, error) {
	log.L().Info("closing test run", zap.Int64("run_id", runID))

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return nil, err
	}

	run, err := s.client.CloseRun(ctx, runID)
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
func (s *RunService) Delete(ctx context.Context, runID int64) error {
	log.L().Warn("deleting test run", zap.Int64("run_id", runID))

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return err
	}

	if err := s.client.DeleteRun(ctx, runID); err != nil {
		log.L().Error("failed to delete run", zap.Int64("run_id", runID), zap.Error(err))
		return err
	}

	log.L().Warn("test run deleted", zap.Int64("run_id", runID))
	return nil
}

// Output выводит результат в JSON и сохраняет в файл (если указан --output)
func (s *RunService) Output(ctx context.Context, cmd *cobra.Command, data interface{}) error {
	return utils.OutputResult(cmd, data)
}

// PrintSuccess выводит сообщение об успехе
func (s *RunService) PrintSuccess(ctx context.Context, cmd *cobra.Command, format string, args ...interface{}) {
	utils.PrintSuccess(cmd, format, args...)
}

// ParseID парсит ID из аргументов команды
func (s *RunService) ParseID(ctx context.Context, args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("missing ID argument at position %d", index)
	}
	return utils.ParseID(args[index])
}

// validateID проверяет что ID положительный
func (s *RunService) validateID(id int64, fieldName string) error {
	if id <= 0 {
		return fmt.Errorf("%s must be a positive number, got: %d", fieldName, id)
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
