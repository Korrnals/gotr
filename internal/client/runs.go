// internal/client/runs.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetRun получает информацию о тест-ране
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#getrun
func (c *HTTPClient) GetRun(runID int64) (*data.Run, error) {
	endpoint := fmt.Sprintf("get_run/%d", runID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetRun для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении рана %d: %s",
			resp.Status, runID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("ошибка декодирования рана %d: %w", runID, err)
	}

	return &run, nil
}

// GetRuns получает список тест-ранов проекта
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#getruns
func (c *HTTPClient) GetRuns(projectID int64) (data.GetRunsResponse, error) {
	endpoint := fmt.Sprintf("get_runs/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetRuns для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении ранов проекта %d: %s",
			resp.Status, projectID, string(body))
	}

	var runs data.GetRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ранов: %w", err)
	}

	return runs, nil
}

// AddRun создаёт новый тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#addrun
func (c *HTTPClient) AddRun(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddRunRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_run/%d", projectID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddRun для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при создании рана в проекте %d: %s",
			resp.Status, projectID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("ошибка декодирования созданного рана: %w", err)
	}

	return &run, nil
}

// UpdateRun обновляет существующий тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#updaterun
func (c *HTTPClient) UpdateRun(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateRunRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_run/%d", runID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateRun для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при обновлении рана %d: %s",
			resp.Status, runID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("ошибка декодирования обновлённого рана %d: %w", runID, err)
	}

	return &run, nil
}

// CloseRun закрывает тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#closerun
func (c *HTTPClient) CloseRun(runID int64) (*data.Run, error) {
	endpoint := fmt.Sprintf("close_run/%d", runID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса CloseRun для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при закрытии рана %d: %s",
			resp.Status, runID, string(body))
	}

	var run data.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("ошибка декодирования закрытого рана %d: %w", runID, err)
	}

	return &run, nil
}

// DeleteRun удаляет тест-ран
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#deleterun
func (c *HTTPClient) DeleteRun(runID int64) error {
	endpoint := fmt.Sprintf("delete_run/%d", runID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteRun для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления рана %d: %s, тело: %s", runID, resp.Status, string(body))
	}

	return nil
}
