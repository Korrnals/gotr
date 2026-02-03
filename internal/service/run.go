// internal/service/run.go
// Сервис для работы с test runs
package service

import (
	"errors"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
)

// RunService предоставляет методы для работы с test runs
type RunService struct {
	client *client.HTTPClient
}

// NewRunService создаёт новый сервис для работы с runs
func NewRunService(client *client.HTTPClient) *RunService {
	return &RunService{client: client}
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
	if err := s.validateID(projectID, "project_id"); err != nil {
		return nil, err
	}
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("валидация запроса: %w", err)
	}
	return s.client.AddRun(projectID, req)
}

// Update обновляет существующий test run с валидацией
func (s *RunService) Update(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	return s.client.UpdateRun(runID, req)
}

// Close закрывает test run с валидацией ID
func (s *RunService) Close(runID int64) (*data.Run, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	return s.client.CloseRun(runID)
}

// Delete удаляет test run с валидацией ID
func (s *RunService) Delete(runID int64) error {
	if err := s.validateID(runID, "run_id"); err != nil {
		return err
	}
	return s.client.DeleteRun(runID)
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
