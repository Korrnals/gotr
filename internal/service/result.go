// internal/service/result.go
// Service for working with test results.
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"go.uber.org/zap"
)

// resultClientInterface defines the client methods required by ResultService.
type resultClientInterface interface {
	GetResults(ctx context.Context, testID int64) (data.GetResultsResponse, error)
	GetResultsForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error)
	GetResultsForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error)
	GetRuns(ctx context.Context, projectID int64) (data.GetRunsResponse, error)
	AddResult(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResultForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error)
	AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error)
}

// ResultService provides methods for working with test results.
type ResultService struct {
	client resultClientInterface
}

// NewResultService creates a new service for working with test results.
func NewResultService(c client.ClientInterface) *ResultService {
	return &ResultService{client: c}
}

// GetForTest retrieves results for a test ID.
func (s *ResultService) GetForTest(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
	if err := s.validateID(testID, "test_id"); err != nil {
		return nil, err
	}
	return s.client.GetResults(ctx, testID)
}

// GetForCase retrieves results for a case in a run.
func (s *ResultService) GetForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	if err := s.validateID(caseID, "case_id"); err != nil {
		return nil, err
	}
	return s.client.GetResultsForCase(ctx, runID, caseID)
}

// GetForRun retrieves results for a run ID.
func (s *ResultService) GetForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	return s.client.GetResultsForRun(ctx, runID)
}

// GetRunsForProject retrieves the list of runs for a project (for interactive selection).
func (s *ResultService) GetRunsForProject(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
	if err := s.validateID(projectID, "project_id"); err != nil {
		return nil, err
	}
	return s.client.GetRuns(ctx, projectID)
}

// AddForTest adds a result for a test with validation.
func (s *ResultService) AddForTest(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
	var statusID int64
	if req != nil {
		statusID = req.StatusID
	}

	log.L().Info("adding result for test",
		zap.Int64("test_id", testID),
		zap.Int64("status_id", statusID),
	)

	if err := s.validateID(testID, "test_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "test_id"), zap.Error(err))
		return nil, err
	}
	if err := s.validateAddResultRequest(req); err != nil {
		log.L().Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("request validation: %w", err)
	}

	result, err := s.client.AddResult(ctx, testID, req)
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

// AddForCase adds a result for a case in a run with validation.
func (s *ResultService) AddForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	var statusID int64
	if req != nil {
		statusID = req.StatusID
	}

	log.L().Info("adding result for case",
		zap.Int64("run_id", runID),
		zap.Int64("case_id", caseID),
		zap.Int64("status_id", statusID),
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
		return nil, fmt.Errorf("request validation: %w", err)
	}

	result, err := s.client.AddResultForCase(ctx, runID, caseID, req)
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

// AddResults adds multiple results (bulk) with validation.
func (s *ResultService) AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	count := 0
	if req != nil {
		count = len(req.Results)
	}

	log.L().Info("adding bulk results",
		zap.Int64("run_id", runID),
		zap.Int("count", count),
	)

	if err := s.validateID(runID, "run_id"); err != nil {
		log.L().Error("validation failed", zap.String("field", "run_id"), zap.Error(err))
		return nil, err
	}
	if err := s.validateAddResultsRequest(req); err != nil {
		log.L().Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("request validation: %w", err)
	}

	results, err := s.client.AddResults(ctx, runID, req)
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

// AddResultsForCases adds results for cases (bulk) with validation.
func (s *ResultService) AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	if err := s.validateAddResultsForCasesRequest(req); err != nil {
		return nil, fmt.Errorf("request validation: %w", err)
	}
	return s.client.AddResultsForCases(ctx, runID, req)
}

// ParseID parses an ID from command arguments.
func (s *ResultService) ParseID(ctx context.Context, args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("missing ID argument at position %d", index)
	}
	return strconv.ParseInt(args[index], 10, 64)
}

// validateID checks that the ID is a positive number.
func (s *ResultService) validateID(id int64, fieldName string) error {
	if id <= 0 {
		return fmt.Errorf("%s must be a positive number, got: %d", fieldName, id)
	}
	return nil
}

// validateAddResultRequest validates a request to add a result.
func (s *ResultService) validateAddResultRequest(req *data.AddResultRequest) error {
	if req == nil {
		return errors.New("запрос не может быть nil")
	}
	if req.StatusID <= 0 {
		return errors.New("status_id обязателен и должен быть положительным числом")
	}
	return nil
}

// validateAddResultsRequest validates a bulk results request.
func (s *ResultService) validateAddResultsRequest(req *data.AddResultsRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}
	if len(req.Results) == 0 {
		return errors.New("results list cannot be empty")
	}
	for i, r := range req.Results {
		if r.StatusID <= 0 {
			return fmt.Errorf("result[%d]: status_id must be positive", i)
		}
	}
	return nil
}

// validateAddResultsForCasesRequest validates bulk request for cases
func (s *ResultService) validateAddResultsForCasesRequest(req *data.AddResultsForCasesRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}
	if len(req.Results) == 0 {
		return errors.New("results list cannot be empty")
	}
	for i, r := range req.Results {
		if r.CaseID <= 0 {
			return fmt.Errorf("result[%d]: case_id must be positive", i)
		}
		if r.StatusID <= 0 {
			return fmt.Errorf("result[%d]: status_id must be positive", i)
		}
	}
	return nil
}
