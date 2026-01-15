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
// Возвращает все шаги (с пагинацией, если она есть).
func (c *HTTPClient) GetSharedSteps(projectID int64) ([]data.SharedStep, error) {
	endpoint := fmt.Sprintf("get_shared_steps/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSharedSteps проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s для проекта %d: %s", resp.Status, projectID, string(body))
	}

	var steps []data.SharedStep
	if err := json.NewDecoder(resp.Body).Decode(&steps); err != nil {
		return nil, fmt.Errorf("ошибка декодирования shared steps проекта %d: %w", projectID, err)
	}

	return steps, nil
}

// GetSharedStep получает один shared step по ID.
func (c *HTTPClient) GetSharedStep(stepID int64) (*data.SharedStep, error) {
	endpoint := fmt.Sprintf("get_shared_step/%d", stepID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSharedStep %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении shared step %d: %s",
			resp.Status, stepID, string(body))
	}

	var step data.SharedStep
	if err := json.NewDecoder(resp.Body).Decode(&step); err != nil {
		return nil, fmt.Errorf("ошибка декодирования shared step %d: %w", stepID, err)
	}

	return &step, nil
}

// GetSharedStepHistory получает историю изменений shared step.
func (c *HTTPClient) GetSharedStepHistory(stepID int64) (*data.GetSharedStepHistoryResponse, error) {
	endpoint := fmt.Sprintf("get_shared_step_history/%d", stepID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSharedStepHistory %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении истории shared step %d: %s",
			resp.Status, stepID, string(body))
	}

	var result data.GetSharedStepHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования истории shared step %d: %w", stepID, err)
	}

	return &result, nil
}

// AddSharedStep создаёт новый shared step.
// Требует Title в запросе.
func (c *HTTPClient) AddSharedStep(req *data.AddSharedStepRequest) (*data.SharedStep, error) {
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при создании shared step: %s", resp.Status, string(body))
	}

	var result data.SharedStep
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования созданного shared step: %w", err)
	}

	return &result, nil
}

// UpdateSharedStep обновляет существующий shared step.
// Поддерживает частичные обновления.
func (c *HTTPClient) UpdateSharedStep(stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateSharedStepRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_shared_step/%d", stepID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateSharedStep %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при обновлении shared step %d: %s",
			resp.Status, stepID, string(body))
	}

	var result data.SharedStep
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования обновлённого shared step %d: %w", stepID, err)
	}

	return &result, nil
}

// DeleteSharedStep удаляет shared step по ID.
// keepInCases: 1 — сохранить шаг в кейсах, 0 — удалить полностью.
func (c *HTTPClient) DeleteSharedStep(stepID int64, keepInCases int) error {
	endpoint := fmt.Sprintf("delete_shared_step/%d", stepID)
	query := map[string]string{
		"keep_in_cases": fmt.Sprintf("%d", keepInCases),
	}

	resp, err := c.Post(endpoint, nil, query)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteSharedStep %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления shared step %d: %s, тело: %s",
			stepID, resp.Status, string(body))
	}

	return nil
}
