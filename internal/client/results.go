// internal/client/results.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// AddResult добавляет результат для теста
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresult
func (c *HTTPClient) AddResult(testID int64, req *data.AddResultRequest) (*data.Result, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddResultRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_result/%d", testID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddResult для теста %d: %w", testID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при добавлении результата для теста %d: %s",
			resp.Status, testID, string(body))
	}

	var result data.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// AddResultForCase добавляет результат для кейса в ране
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresultforcase
func (c *HTTPClient) AddResultForCase(runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddResultRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_result_for_case/%d/%d", runID, caseID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddResultForCase для рана %d, кейса %d: %w", runID, caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при добавлении результата для рана %d, кейса %d: %s",
			resp.Status, runID, caseID, string(body))
	}

	var result data.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// AddResults добавляет результаты для нескольких тестов в ране (bulk)
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresults
func (c *HTTPClient) AddResults(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddResultsRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_results/%d", runID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddResults для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при bulk-добавлении результатов для рана %d: %s",
			resp.Status, runID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	return results, nil
}

// AddResultsForCases добавляет результаты для кейсов в ране (bulk)
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresultsforcases
func (c *HTTPClient) AddResultsForCases(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddResultsForCasesRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_results_for_cases/%d", runID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddResultsForCases для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при bulk-добавлении результатов для кейсов в ране %d: %s",
			resp.Status, runID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	return results, nil
}

// GetResults получает результаты для теста
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#getresults
func (c *HTTPClient) GetResults(testID int64) (data.GetResultsResponse, error) {
	endpoint := fmt.Sprintf("get_results/%d", testID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetResults для теста %d: %w", testID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении результатов для теста %d: %s",
			resp.Status, testID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	return results, nil
}

// GetResultsForRun получает результаты для рана
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#getresultsforrun
func (c *HTTPClient) GetResultsForRun(runID int64) (data.GetResultsResponse, error) {
	endpoint := fmt.Sprintf("get_results_for_run/%d", runID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetResultsForRun для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении результатов для рана %d: %s",
			resp.Status, runID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	return results, nil
}

// GetResultsForCase получает результаты для кейса в ране
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#getresultsforcase
func (c *HTTPClient) GetResultsForCase(runID, caseID int64) (data.GetResultsResponse, error) {
	endpoint := fmt.Sprintf("get_results_for_case/%d/%d", runID, caseID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetResultsForCase для рана %d, кейса %d: %w", runID, caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении результатов для рана %d, кейса %d: %s",
			resp.Status, runID, caseID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	return results, nil
}
