package result

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// ResultServiceInterface defines the interface for test result operations.
type ResultServiceInterface interface {
	ParseID(ctx context.Context, args []string, index int) (int64, error)
	PrintSuccess(ctx context.Context, cmd *cobra.Command, format string, args ...interface{})
	Output(ctx context.Context, cmd *cobra.Command, data interface{}) error
	AddForTest(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error)
	AddForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error)
	AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error)
	GetForTest(ctx context.Context, testID int64) (data.GetResultsResponse, error)
	GetForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error)
	GetForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error)
	GetRunsForProject(ctx context.Context, projectID int64) (data.GetRunsResponse, error)
	// AddBulkResults parses JSON and submits results (bulk operation).
	AddBulkResults(ctx context.Context, runID int64, fileData []byte) (interface{}, error)
}

// resultServiceWrapper wraps the result service for command handlers.
type resultServiceWrapper struct {
	svc *service.ResultService
}

// Compile-time check: resultServiceWrapper implements ResultServiceInterface.
var _ ResultServiceInterface = (*resultServiceWrapper)(nil)

// ParseID delegates ID parsing to the underlying result service.
func (w *resultServiceWrapper) ParseID(ctx context.Context, args []string, index int) (int64, error) {
	return w.svc.ParseID(ctx, args, index)
}

// PrintSuccess delegates success message formatting to the underlying result service.
func (w *resultServiceWrapper) PrintSuccess(ctx context.Context, cmd *cobra.Command, format string, args ...interface{}) {
	w.svc.PrintSuccess(ctx, cmd, format, args...)
}

// Output delegates command output formatting to the underlying result service.
func (w *resultServiceWrapper) Output(ctx context.Context, cmd *cobra.Command, v interface{}) error {
	return w.svc.Output(ctx, cmd, v)
}

// AddForTest delegates single-test result creation to the underlying result service.
func (w *resultServiceWrapper) AddForTest(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
	return w.svc.AddForTest(ctx, testID, req)
}

// AddForCase delegates case result creation to the underlying result service.
func (w *resultServiceWrapper) AddForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	return w.svc.AddForCase(ctx, runID, caseID, req)
}

// AddResults delegates bulk test result creation to the underlying result service.
func (w *resultServiceWrapper) AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	return w.svc.AddResults(ctx, runID, req)
}

// AddResultsForCases delegates bulk case result creation to the underlying result service.
func (w *resultServiceWrapper) AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	return w.svc.AddResultsForCases(ctx, runID, req)
}

// GetForTest delegates test result retrieval to the underlying result service.
func (w *resultServiceWrapper) GetForTest(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForTest(ctx, testID)
}

// GetForCase delegates run/case result retrieval to the underlying result service.
func (w *resultServiceWrapper) GetForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForCase(ctx, runID, caseID)
}

// GetForRun delegates run result retrieval to the underlying result service.
func (w *resultServiceWrapper) GetForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForRun(ctx, runID)
}

// GetRunsForProject delegates run listing to the underlying result service.
func (w *resultServiceWrapper) GetRunsForProject(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
	return w.svc.GetRunsForProject(ctx, projectID)
}

// AddBulkResults parses JSON and submits results (bulk operation).
func (w *resultServiceWrapper) AddBulkResults(ctx context.Context, runID int64, fileData []byte) (interface{}, error) {
	// Try parsing as an array with test_id
	var testResults []data.ResultEntry
	if err := json.Unmarshal(fileData, &testResults); err == nil && len(testResults) > 0 {
		req := &data.AddResultsRequest{Results: testResults}
		return w.svc.AddResults(ctx, runID, req)
	}

	// Try parsing as an array with case_id
	var caseResults []data.ResultForCaseEntry
	if err := json.Unmarshal(fileData, &caseResults); err == nil && len(caseResults) > 0 {
		req := &data.AddResultsForCasesRequest{Results: caseResults}
		return w.svc.AddResultsForCases(ctx, runID, req)
	}

	return nil, fmt.Errorf("failed to parse JSON file: expected array with test_id or case_id")
}

// newResultServiceFromInterface creates a result service from a client interface.
func newResultServiceFromInterface(cli client.ClientInterface) ResultServiceInterface {
	// Try to cast to *HTTPClient (not a mock)
	if httpClient, ok := cli.(*client.HTTPClient); ok {
		return &resultServiceWrapper{svc: service.NewResultService(httpClient)}
	}
	// For tests with a mock, use a special constructor
	return &resultServiceWrapper{svc: service.NewResultServiceFromInterface(cli)}
}
