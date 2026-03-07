package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Korrnals/gotr/internal/models/data"
	"io"
	"net/http"
)

// GetSuites получает список всех тест-сюит проекта.
func (c *HTTPClient) GetSuites(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
	endpoint := fmt.Sprintf("get_suites/%d", projectID)
	suites, err := fetchAllPages[data.Suite](ctx, c, endpoint, nil, "suites")
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSuites для проекта %d: %w", projectID, err)
	}
	return data.GetSuitesResponse(suites), nil
}

// GetSuite получает одну тест-сюиту по ID.
func (c *HTTPClient) GetSuite(ctx context.Context, suiteID int64) (*data.Suite, error) {
	endpoint := fmt.Sprintf("get_suite/%d", suiteID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSuite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении сюиты %d: %s", resp.Status, suiteID, string(body))
	}

	var suite data.Suite
	if err := json.NewDecoder(resp.Body).Decode(&suite); err != nil {
		return nil, fmt.Errorf("ошибка декодирования сюиты %d: %w", suiteID, err)
	}

	return &suite, nil
}

// AddSuite создаёт новую тест-сюиту в проекте.
func (c *HTTPClient) AddSuite(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddSuiteRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_suite/%d", projectID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddSuite для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при создании сюиты в проекте %d: %s", resp.Status, projectID, string(body))
	}

	var suite data.Suite
	if err := json.NewDecoder(resp.Body).Decode(&suite); err != nil {
		return nil, fmt.Errorf("ошибка декодирования созданной сюиты: %w", err)
	}

	return &suite, nil
}

// UpdateSuite обновляет существующую тест-сюиту.
func (c *HTTPClient) UpdateSuite(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateSuiteRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_suite/%d", suiteID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateSuite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при обновлении сюиты %d: %s", resp.Status, suiteID, string(body))
	}

	var suite data.Suite
	if err := json.NewDecoder(resp.Body).Decode(&suite); err != nil {
		return nil, fmt.Errorf("ошибка декодирования обновлённой сюиты %d: %w", suiteID, err)
	}

	return &suite, nil
}

// DeleteSuite удаляет тест-сюиту по ID.
// Удаление необратимо — все кейсы в сюите удаляются.
func (c *HTTPClient) DeleteSuite(ctx context.Context, suiteID int64) error {
	endpoint := fmt.Sprintf("delete_suite/%d", suiteID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteSuite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления сюиты %d: %s, тело: %s", suiteID, resp.Status, string(body))
	}

	return nil
}
