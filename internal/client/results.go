// internal/client/results.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// AddResult добавляет результат for test
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresult
func (c *HTTPClient) AddResult(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddResultRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_result/%d", testID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddResult for test %d: %w", testID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s adding result for test %d: %s",
			resp.Status, testID, string(body))
	}

	var result data.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error result: %w", err)
	}

	return &result, nil
}

// AddResultForCase добавляет результат для case в ране
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresultforcase
func (c *HTTPClient) AddResultForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddResultRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_result_for_case/%d/%d", runID, caseID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddResultForCase for run %d, case %d: %w", runID, caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s adding result for run %d, case %d: %s",
			resp.Status, runID, caseID, string(body))
	}

	var result data.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error result: %w", err)
	}

	return &result, nil
}

// AddResults добавляет результаты для нескольких тестов в ране (bulk)
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresults
func (c *HTTPClient) AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddResultsRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_results/%d", runID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddResults for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s bulk adding results for run %d: %s",
			resp.Status, runID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode error results: %w", err)
	}

	return results, nil
}

// AddResultsForCases добавляет результаты for cases in run (bulk)
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresultsforcases
func (c *HTTPClient) AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddResultsForCasesRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_results_for_cases/%d", runID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddResultsForCases for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s bulk adding results for cases in run %d: %s",
			resp.Status, runID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode error results: %w", err)
	}

	return results, nil
}

// GetResults получает результаты for test (поддерживает пагинацию)
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#getresults
func (c *HTTPClient) GetResults(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
	endpoint := fmt.Sprintf("get_results/%d", testID)
	results, err := fetchAllPages[data.Result](ctx, c, endpoint, nil, "results")
	if err != nil {
		return nil, fmt.Errorf("request error GetResults for test %d: %w", testID, err)
	}
	return data.GetResultsResponse(results), nil
}

// GetResultsForRun получает результаты for run (поддерживает пагинацию)
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#getresultsforrun
func (c *HTTPClient) GetResultsForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
	endpoint := fmt.Sprintf("get_results_for_run/%d", runID)
	results, err := fetchAllPages[data.Result](ctx, c, endpoint, nil, "results")
	if err != nil {
		return nil, fmt.Errorf("request error GetResultsForRun for run %d: %w", runID, err)
	}
	return data.GetResultsResponse(results), nil
}

// GetResultsForCase получает результаты для case в ране
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#getresultsforcase
func (c *HTTPClient) GetResultsForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
	endpoint := fmt.Sprintf("get_results_for_case/%d/%d", runID, caseID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetResultsForCase for run %d, case %d: %w", runID, caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s getting results for run %d, case %d: %s",
			resp.Status, runID, caseID, string(body))
	}

	var results data.GetResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode error results: %w", err)
	}

	return results, nil
}
