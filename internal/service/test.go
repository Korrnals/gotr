// internal/service/test.go
package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// TestService provides business logic for working with tests.
type TestService struct {
	client client.ClientInterface
}

// NewTestService creates a new service for working with tests.
func NewTestService(client client.ClientInterface) *TestService {
	return &TestService{client: client}
}

// Get retrieves a test by ID.
func (s *TestService) Get(ctx context.Context, testID int64) (*data.Test, error) {
	if testID <= 0 {
		return nil, fmt.Errorf("test ID must be a positive number")
	}

	test, err := s.client.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	return test, nil
}

// GetForRun retrieves the list of tests for a run.
func (s *TestService) GetForRun(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
	if runID <= 0 {
		return nil, fmt.Errorf("run ID must be a positive number")
	}

	tests, err := s.client.GetTests(ctx, runID, filters)
	if err != nil {
		return nil, err
	}

	return tests, nil
}

// Update updates a test.
func (s *TestService) Update(ctx context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
	if testID <= 0 {
		return nil, fmt.Errorf("test ID must be a positive number")
	}

	if req == nil {
		return nil, fmt.Errorf("request cannot be empty")
	}

	// Validation
	if req.StatusID < 0 {
		return nil, fmt.Errorf("status_id cannot be negative")
	}

	test, err := s.client.UpdateTest(ctx, testID, req)
	if err != nil {
		return nil, err
	}

	return test, nil
}

// ParseID parses an ID from command-line arguments.
func (s *TestService) ParseID(ctx context.Context, args []string, index int) (int64, error) {
	if len(args) <= index {
		return 0, fmt.Errorf("ID is required")
	}

	id, err := strconv.ParseInt(args[index], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID: %s", args[index])
	}

	if id <= 0 {
		return 0, fmt.Errorf("ID must be a positive number")
	}

	return id, nil
}

// PrintSuccess prints a success message.
func (s *TestService) PrintSuccess(ctx context.Context, cmd *cobra.Command, format string, args ...interface{}) {
	output.PrintSuccess(cmd, format, args...)
}

// Output renders the result as JSON.
func (s *TestService) Output(ctx context.Context, cmd *cobra.Command, data interface{}) error {
	return output.OutputResultWithFlags(cmd, data)
}
