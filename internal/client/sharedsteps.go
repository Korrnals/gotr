package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetSharedSteps получает список shared steps for project (поддерживает пагинацию).
func (c *HTTPClient) GetSharedSteps(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
	endpoint := fmt.Sprintf("get_shared_steps/%d", projectID)
	steps, err := fetchAllPages[data.SharedStep](ctx, c, endpoint, nil, "shared_steps")
	if err != nil {
		return nil, fmt.Errorf("request error GetSharedSteps project %d: %w", projectID, err)
	}
	return data.GetSharedStepsResponse(steps), nil
}

// GetSharedStep получает один shared step по ID.
func (c *HTTPClient) GetSharedStep(ctx context.Context, stepID int64) (*data.SharedStep, error) {
	endpoint := fmt.Sprintf("get_shared_step/%d", stepID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetSharedStep %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	var step data.SharedStep
	if err := json.NewDecoder(resp.Body).Decode(&step); err != nil {
		return nil, fmt.Errorf("decode error shared step %d: %w", stepID, err)
	}

	return &step, nil
}

// GetSharedStepHistory получает историю изменений shared step.
func (c *HTTPClient) GetSharedStepHistory(ctx context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error) {
	endpoint := fmt.Sprintf("get_shared_step_history/%d", stepID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetSharedStepHistory %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	var result data.GetSharedStepHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error shared step history %d: %w", stepID, err)
	}

	return &result, nil
}

// AddSharedStep создаёт новый shared step в указанном проекте.
// Требует Title в запросе.
func (c *HTTPClient) AddSharedStep(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
	endpoint := fmt.Sprintf("add_shared_step/%d", projectID)

	bodyBytes, _ := json.Marshal(req)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddSharedStep in project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.SharedStep
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error created shared step: %w", err)
	}

	return &result, nil
}

// UpdateSharedStep обновляет существующий shared step.
// Поддерживает частичные обновления.
func (c *HTTPClient) UpdateSharedStep(ctx context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
	bodyBytes, _ := json.Marshal(req)
	endpoint := fmt.Sprintf("update_shared_step/%d", stepID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateSharedStep %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	var result data.SharedStep
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error updated shared step %d: %w", stepID, err)
	}

	return &result, nil
}

// DeleteSharedStep удаляет shared step по ID.
// keepInCases: 1 — сохранить шаг в caseх, 0 — удалить полностью.
func (c *HTTPClient) DeleteSharedStep(ctx context.Context, stepID int64, keepInCases int) error {
	endpoint := fmt.Sprintf("delete_shared_step/%d", stepID)
	query := map[string]string{
		"keep_in_cases": fmt.Sprintf("%d", keepInCases),
	}

	resp, err := c.Post(ctx, endpoint, nil, query)
	if err != nil {
		return fmt.Errorf("request error DeleteSharedStep %d: %w", stepID, err)
	}
	defer resp.Body.Close()

	return nil
}
