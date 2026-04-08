// internal/service/run.go
// Service for working with test runs.
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

// runClientInterface defines the client methods required by RunService.
type runClientInterface interface {
	GetRun(ctx context.Context, runID int64) (*data.Run, error)
	GetRuns(ctx context.Context, projectID int64) (data.GetRunsResponse, error)
	AddRun(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error)
	UpdateRun(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	CloseRun(ctx context.Context, runID int64) (*data.Run, error)
	DeleteRun(ctx context.Context, runID int64) error
}

// RunService provides methods for working with test runs.
type RunService struct {
	client runClientInterface
}

// NewRunService creates a new service for working with test runs.
func NewRunService(c client.ClientInterface) *RunService {
	return &RunService{client: c}
}

// Get retrieves information about a test run by ID.
func (s *RunService) Get(ctx context.Context, runID int64) (*data.Run, error) {
	if err := s.validateID(runID, "run_id"); err != nil {
		return nil, err
	}
	return s.client.GetRun(ctx, runID)
}

// GetByProject retrieves the list of runs for a project.
func (s *RunService) GetByProject(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
	if err := s.validateID(projectID, "project_id"); err != nil {
		return nil, err
	}
	return s.client.GetRuns(ctx, projectID)
}

// Create creates a new test run with parameter validation.
func (s *RunService) Create(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	var name string
	var suiteID int64
	if req != nil {
		name = req.Name
		suiteID = req.SuiteID
	}

	log.L().Info("creating test run",
		zap.Int64("project_id", projectID),
		zap.String("name", name),
		zap.Int64("suite_id", suiteID),
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

// Update updates an existing test run with validation.
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

// Close closes a test run with ID validation.
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

// Delete deletes a test run with ID validation.
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

// ParseID parses an ID from command arguments.
func (s *RunService) ParseID(ctx context.Context, args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("missing ID argument at position %d", index)
	}
	return strconv.ParseInt(args[index], 10, 64)
}

// validateID checks that the ID is a positive number.
func (s *RunService) validateID(id int64, fieldName string) error {
	if id <= 0 {
		return fmt.Errorf("%s must be a positive number, got: %d", fieldName, id)
	}
	return nil
}

// validateCreateRequest validates the parameters for creating a run.
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
