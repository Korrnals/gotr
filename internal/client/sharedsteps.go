package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gotr/internal/models/data"
	"io"
	"net/http"
)

// GetSharedSteps получает список shared steps для проекта.
// Требует projectID.
func (c *HTTPClient) GetSharedSteps(projectID int64) (*data.GetSharedStepsResponse, error) {
	endpoint := fmt.Sprintf("get_shared_steps/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSharedSteps для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetSharedStepsResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetSharedSteps: %w", err)
	}
	return &result, nil
}

// GetSharedStep получает информацию о конкретном shared step по ID.
func (c *HTTPClient) GetSharedStep(sharedStepID int64) (*data.GetSharedStepResponse, error) {
	endpoint := fmt.Sprintf("get_shared_step/%d", sharedStepID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSharedStep %d: %w", sharedStepID, err)
	}
	defer resp.Body.Close()

	var result data.GetSharedStepResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetSharedStep %d: %w", sharedStepID, err)
	}
	return &result, nil
}

// GetSharedStepHistory получает историю изменений shared step.
func (c *HTTPClient) GetSharedStepHistory(sharedStepID int64) (*data.GetSharedStepHistoryResponse, error) {
	endpoint := fmt.Sprintf("get_shared_step_history/%d", sharedStepID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSharedStepHistory %d: %w", sharedStepID, err)
	}
	defer resp.Body.Close()

	var result data.GetSharedStepHistoryResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetSharedStepHistory %d: %w", sharedStepID, err)
	}
	return &result, nil
}

// AddSharedStep создаёт новый shared step.
// Требует Title.
func (c *HTTPClient) AddSharedStep(req *data.AddSharedStepRequest) (*data.GetSharedStepResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddSharedStepRequest: %w", err)
	}

	endpoint := "add_shared_step"
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddSharedStep: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetSharedStepResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа AddSharedStep: %w", err)
	}
	return &result, nil
}

// UpdateSharedStep обновляет существующий shared step.
// Поддерживает частичные обновления.
func (c *HTTPClient) UpdateSharedStep(sharedStepID int64, req *data.UpdateSharedStepRequest) (*data.GetSharedStepResponse, error) {
	endpoint := fmt.Sprintf("update_shared_step/%d", sharedStepID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateSharedStepRequest: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateSharedStep %d: %w", sharedStepID, err)
	}
	defer resp.Body.Close()

	var result data.GetSharedStepResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа UpdateSharedStep %d: %w", sharedStepID, err)
	}
	return &result, nil
}

// DeleteSharedStep удаляет shared step по ID.
// KeepInCases: 1 — сохранить step в кейсах, 0 — удалить полностью.
func (c *HTTPClient) DeleteSharedStep(sharedStepID int64, keepInCases int) error {
	endpoint := fmt.Sprintf("delete_shared_step/%d", sharedStepID)
	query := map[string]string{"keep_in_cases": fmt.Sprintf("%d", keepInCases)}

	resp, err := c.Post(endpoint, nil, query)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteSharedStep %d: %w", sharedStepID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления shared step %d: %s, тело: %s", sharedStepID, resp.Status, string(body))
	}
	return nil
}