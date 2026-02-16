package result

import (
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// ResultServiceInterface определяет интерфейс для операций с результатами
type ResultServiceInterface interface {
	ParseID(args []string, index int) (int64, error)
	PrintSuccess(cmd *cobra.Command, format string, args ...interface{})
	Output(cmd *cobra.Command, data interface{}) error
	AddForTest(testID int64, req *data.AddResultRequest) (*data.Result, error)
	AddForCase(runID, caseID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResults(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error)
	AddResultsForCases(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error)
	GetForTest(testID int64) (data.GetResultsResponse, error)
	GetForCase(runID, caseID int64) (data.GetResultsResponse, error)
	GetForRun(runID int64) (data.GetResultsResponse, error)
	GetRunsForProject(projectID int64) (data.GetRunsResponse, error)
	// AddBulkResults парсит JSON и добавляет результаты (bulk операция)
	AddBulkResults(runID int64, fileData []byte) (interface{}, error)
}

// resultServiceWrapper оборачивает сервис для работы с результатами
type resultServiceWrapper struct {
	svc *service.ResultService
}

// Проверка что resultServiceWrapper реализует ResultServiceInterface
var _ ResultServiceInterface = (*resultServiceWrapper)(nil)

func (w *resultServiceWrapper) ParseID(args []string, index int) (int64, error) {
	return w.svc.ParseID(args, index)
}

func (w *resultServiceWrapper) PrintSuccess(cmd *cobra.Command, format string, args ...interface{}) {
	w.svc.PrintSuccess(cmd, format, args...)
}

func (w *resultServiceWrapper) Output(cmd *cobra.Command, data interface{}) error {
	return w.svc.Output(cmd, data)
}

func (w *resultServiceWrapper) AddForTest(testID int64, req *data.AddResultRequest) (*data.Result, error) {
	return w.svc.AddForTest(testID, req)
}

func (w *resultServiceWrapper) AddForCase(runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	return w.svc.AddForCase(runID, caseID, req)
}

func (w *resultServiceWrapper) AddResults(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	return w.svc.AddResults(runID, req)
}

func (w *resultServiceWrapper) AddResultsForCases(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	return w.svc.AddResultsForCases(runID, req)
}

func (w *resultServiceWrapper) GetForTest(testID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForTest(testID)
}

func (w *resultServiceWrapper) GetForCase(runID, caseID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForCase(runID, caseID)
}

func (w *resultServiceWrapper) GetForRun(runID int64) (data.GetResultsResponse, error) {
	return w.svc.GetForRun(runID)
}

func (w *resultServiceWrapper) GetRunsForProject(projectID int64) (data.GetRunsResponse, error) {
	return w.svc.GetRunsForProject(projectID)
}

// AddBulkResults парсит JSON и добавляет результаты (bulk операция)
func (w *resultServiceWrapper) AddBulkResults(runID int64, fileData []byte) (interface{}, error) {
	// Пробуем как массив с test_id
	var testResults []data.ResultEntry
	if err := json.Unmarshal(fileData, &testResults); err == nil && len(testResults) > 0 {
		req := &data.AddResultsRequest{Results: testResults}
		return w.svc.AddResults(runID, req)
	}

	// Пробуем как массив с case_id
	var caseResults []data.ResultForCaseEntry
	if err := json.Unmarshal(fileData, &caseResults); err == nil && len(caseResults) > 0 {
		req := &data.AddResultsForCasesRequest{Results: caseResults}
		return w.svc.AddResultsForCases(runID, req)
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
