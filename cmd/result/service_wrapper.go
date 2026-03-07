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

func (w *resultServiceWrapper) ParseID(ctx context.Context, args []string, index int) (int64, error) {
	return w.svc.ParseID(ctx, args, index)
}

func (w *resultServiceWrapper) PrintSuccess(ctx context.Context, cmd *cobra.Command, format string, args ...interface{}) {
	w.svc.PrintSuccess(ctx, cmd, format, args...)
}

func (w *resultServiceWrapper) Output(ctx context.Context, cmd *cobra.Command, data interface{}) error {
	return w.svc.Output(ctx, cmd, data)
}

func (w *resultServiceWrapper) AddForTest(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
	return w.svc.AddForTest(ctx, testID, req)
}

func (w *resultServiceWrapper) AddForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	return w.svc.AddForCase(ctx, runID, caseID, req)
}

func (w *resultServiceWrapper) AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	return w.svc.AddResults(ctx, runID, req)
}

func (w *resultServiceWrapper) AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	return w.svc.AddResultsForCases(ctx, runID, req)
}

func (w *resultServiceWrapper) GetForTest(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForTest(ctx, testID)
}

func (w *resultServiceWrapper) GetForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForCase(ctx, runID, caseID)
}

func (w *resultServiceWrapper) GetForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForRun(ctx, runID)
}

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

	return nil, fmt.Errorf("не удалось распарсить JSON файл: ожидается массив с test_id или case_id")
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
