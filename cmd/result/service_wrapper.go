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

// ResultServiceInterface определяет интерфейс для операций с результатами
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
	// AddBulkResults парсит JSON и добавляет результаты (bulk операция)
	AddBulkResults(ctx context.Context, runID int64, fileData []byte) (interface{}, error)
}

// resultServiceWrapper оборачивает сервис для работы с результатами
type resultServiceWrapper struct {
	svc *service.ResultService
}

// Проверка что resultServiceWrapper реализует ResultServiceInterface
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
func (w *resultServiceWrapper) Output(ctx context.Context, cmd *cobra.Command, data interface{}) error {
	return w.svc.Output(ctx, cmd, data)
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

// AddBulkResults парсит JSON и добавляет результаты (bulk операция)
func (w *resultServiceWrapper) AddBulkResults(ctx context.Context, runID int64, fileData []byte) (interface{}, error) {
	// Пробуем как массив с test_id
	var testResults []data.ResultEntry
	if err := json.Unmarshal(fileData, &testResults); err == nil && len(testResults) > 0 {
		req := &data.AddResultsRequest{Results: testResults}
		return w.svc.AddResults(ctx, runID, req)
	}

	// Пробуем как массив с case_id
	var caseResults []data.ResultForCaseEntry
	if err := json.Unmarshal(fileData, &caseResults); err == nil && len(caseResults) > 0 {
		req := &data.AddResultsForCasesRequest{Results: caseResults}
		return w.svc.AddResultsForCases(ctx, runID, req)
	}

	return nil, fmt.Errorf("failed to parse JSON file: expected array with test_id or case_id")
}

// newResultServiceFromInterface создаёт сервис из клиента-интерфейса
func newResultServiceFromInterface(cli client.ClientInterface) ResultServiceInterface {
	// Пытаемся привести к *HTTPClient, если это не mock
	if httpClient, ok := cli.(*client.HTTPClient); ok {
		return &resultServiceWrapper{svc: service.NewResultService(httpClient)}
	}
	// Для тестов с mock - используем специальный конструктор
	return &resultServiceWrapper{svc: service.NewResultServiceFromInterface(cli)}
}
