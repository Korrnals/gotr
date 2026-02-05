// internal/service/result.go
// Сервис для работы с результатами тестов
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

// ResultService предоставляет методы для работы с результатами тестов
type ResultService struct {
	client *client.HTTPClient
}

// NewResultService создаёт новый сервис для работы с результатами
func NewResultService(client *client.HTTPClient) *ResultService {
	return &ResultService{client: client}
}

// GetForTest получает результаты для test ID
func (s *ResultService) GetForTest(testID int64) (data.GetResultsResponse, error) {
	if err := s.validateID(testID, "test_id"); err != nil {
		return nil, err
	}
	return s.client.GetResults(testID)
}

// GetForCase получает результаты для case в run
func (s *ResultService) GetForCase(runID, caseID int64) (data.GetResultsResponse, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	if err := s.validateID(caseID, "case_id"); err != nil {
		return nil, err
	}
	return s.client.GetResultsForCase(runID, caseID)
}

// GetForRun получает результаты для run ID
func (s *ResultService) GetForRun(runID int64) (data.GetResultsResponse, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	return s.client.GetResultsForRun(runID)
}

// GetRunsForProject получает список runs для проекта (для интерактивного выбора)
func (s *ResultService) GetRunsForProject(projectID int64) (data.GetRunsResponse, error) {
	if err := s.validateID(projectID, "project_id"); err != nil {
		return nil, err
	}
	return s.client.GetRuns(projectID)
}

// AddForTest добавляет результат для test с валидацией
func (s *ResultService) AddForTest(testID int64, req *data.AddResultRequest) (*data.Result, error) {
	log.L().Info("adding result for test",
		zap.Int64("test_id", testID),
		zap.Int64("status_id", req.StatusID),
	)

	if err := s.validateID(testID, "test_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "test_id"), zap.Error(err))
		return nil, err
	}
	if err := s.validateAddResultRequest(req); err != nil {
		log.L().Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("валидация запроса: %w", err)
	}

	result, err := s.client.AddResult(testID, req)
	if err != nil {
		log.L().Error("failed to add result", zap.Int64("test_id", testID), zap.Error(err))
		return nil, err
	}

	log.L().Info("result added",
		zap.Int64("result_id", result.ID),
		zap.Int64("test_id", result.TestID),
		zap.Int64("status_id", result.StatusID),
	)
	return result, nil
}

// AddForCase добавляет результат для case в run с валидацией
func (s *ResultService) AddForCase(runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	log.L().Info("adding result for case",
		zap.Int64("run_id", runID),
		zap.Int64("case_id", caseID),
		zap.Int64("status_id", req.StatusID),
	)

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return nil, err
	}
	if err := s.validateID(caseID, "case_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "case_id"), zap.Error(err))
		return nil, err
	}
	if err := s.validateAddResultRequest(req); err != nil {
		log.L().Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("валидация запроса: %w", err)
	}

	result, err := s.client.AddResultForCase(runID, caseID, req)
	if err != nil {
		log.L().Error("failed to add result", zap.Int64("run_id", runID), zap.Int64("case_id", caseID), zap.Error(err))
		return nil, err
	}

	log.L().Info("result added",
		zap.Int64("result_id", result.ID),
		zap.Int64("run_id", runID),
		zap.Int64("case_id", caseID),
		zap.Int64("status_id", result.StatusID),
	)
	return result, nil
}

// AddResults добавляет несколько результатов (bulk) с валидацией
func (s *ResultService) AddResults(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	log.L().Info("adding bulk results",
		zap.Int64("run_id", runID),
		zap.Int("count", len(req.Results)),
	)

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return nil, err
	}
	if err := s.validateAddResultsRequest(req); err != nil {
		log.L().Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("валидация запроса: %w", err)
	}

	results, err := s.client.AddResults(runID, req)
	if err != nil {
		log.L().Error("failed to add bulk results", zap.Int64("run_id", runID), zap.Error(err))
		return nil, err
	}

	log.L().Info("bulk results added",
		zap.Int64("run_id", runID),
		zap.Int("count", len(results)),
	)
	return results, nil
}

// AddResultsForCases добавляет результаты для кейсов (bulk) с валидацией
func (s *ResultService) AddResultsForCases(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	if err := s.validateAddResultsForCasesRequest(req); err != nil {
		return nil, fmt.Errorf("валидация запроса: %w", err)
	}
	return s.client.AddResultsForCases(runID, req)
}

// Output выводит результат в JSON и сохраняет в файл
func (s *ResultService) Output(cmd *cobra.Command, data interface{}) error {
	return utils.OutputResult(cmd, data)
}

// PrintSuccess выводит сообщение об успехе
func (s *ResultService) PrintSuccess(cmd *cobra.Command, format string, args ...interface{}) {
	utils.PrintSuccess(cmd, format, args...)
}

// ParseID парсит ID из аргументов
func (s *ResultService) ParseID(args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("отсутствует аргумент с ID на позиции %d", index)
	}
	return utils.ParseID(args[index])
}

// validateID проверяет что ID положительный
func (s *ResultService) validateID(id int64, fieldName string) error {
	if id <= 0 {
		return fmt.Errorf("%s должен быть положительным числом, получено: %d", fieldName, id)
	}
	return nil
}

// validateAddResultRequest валидирует запрос добавления результата
func (s *ResultService) validateAddResultRequest(req *data.AddResultRequest) error {
	if req == nil {
		return errors.New("запрос не может быть nil")
	}
	if req.StatusID <= 0 {
		return errors.New("status_id обязателен и должен быть положительным числом")
	}
	return nil
}

// validateAddResultsRequest валидирует bulk запрос
func (s *ResultService) validateAddResultsRequest(req *data.AddResultsRequest) error {
	if req == nil {
		return errors.New("запрос не может быть nil")
	}
	if len(req.Results) == 0 {
		return errors.New("список результатов не может быть пустым")
	}
	for i, r := range req.Results {
		if r.StatusID <= 0 {
			return fmt.Errorf("result[%d]: status_id должен быть положительным", i)
		}
	}
	return nil
}

// validateAddResultsForCasesRequest валидирует bulk запрос для кейсов
func (s *ResultService) validateAddResultsForCasesRequest(req *data.AddResultsForCasesRequest) error {
	if req == nil {
		return errors.New("запрос не может быть nil")
	}
	if len(req.Results) == 0 {
		return errors.New("список результатов не может быть пустым")
	}
	for i, r := range req.Results {
		if r.CaseID <= 0 {
			return fmt.Errorf("result[%d]: case_id должен быть положительным", i)
		}
		if r.StatusID <= 0 {
			return fmt.Errorf("result[%d]: status_id должен быть положительным", i)
		}
	}
	return nil
}
