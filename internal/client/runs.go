// internal/client/runs.go
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

// GetRun получает информацию о тест-ране
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#getrun
func (c *HTTPClient) GetRun(ctx context.Context, runID int64) (*data.Run, error) {
	endpoint := fmt.Sprintf("get_run/%d", runID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetRun for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s getting run %d: %s",
			resp.Status, runID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("decode error run %d: %w", runID, err)
	}

	return &run, nil
}

// GetRuns получает список тест-ранов проекта (поддерживает пагинацию)
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#getruns
func (c *HTTPClient) GetRuns(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
	endpoint := fmt.Sprintf("get_runs/%d", projectID)
	runs, err := fetchAllPages[data.Run](ctx, c, endpoint, nil, "runs")
	if err != nil {
		return nil, fmt.Errorf("request error GetRuns for project %d: %w", projectID, err)
	}
	return data.GetRunsResponse(runs), nil
}

// AddRun создаёт новый тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#addrun
func (c *HTTPClient) AddRun(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddRunRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_run/%d", projectID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddRun for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s creating run in project %d: %s",
			resp.Status, projectID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("decode error created run: %w", err)
	}

	return &run, nil
}

// UpdateRun обновляет существующий тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#updaterun
func (c *HTTPClient) UpdateRun(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error UpdateRunRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_run/%d", runID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateRun for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s updating run %d: %s",
			resp.Status, runID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("decode error updated run %d: %w", runID, err)
	}

	return &run, nil
}

// CloseRun закрывает тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#closerun
func (c *HTTPClient) CloseRun(ctx context.Context, runID int64) (*data.Run, error) {
	endpoint := fmt.Sprintf("close_run/%d", runID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("request error CloseRun for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s closing run %d: %s",
			resp.Status, runID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("decode error closed run %d: %w", runID, err)
	}

	return &run, nil
}

// DeleteRun удаляет тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#deleterun
func (c *HTTPClient) DeleteRun(ctx context.Context, runID int64) error {
	endpoint := fmt.Sprintf("delete_run/%d", runID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteRun for run %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete error run %d: %s, body: %s", runID, resp.Status, string(body))
	}

	return nil
}
