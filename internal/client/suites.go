package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gotr/internal/models/data"
	"io"
	"net/http"
)

// GetSuites получает список всех тест-сюит проекта.
func (c *HTTPClient) GetSuites(projectID int64) (data.GetSuitesResponse, error) {
	endpoint := fmt.Sprintf("get_suites/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSuites для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении сюит проекта %d: %s", resp.Status, projectID, string(body))
	}

	var suites data.GetSuitesResponse
	if err := json.NewDecoder(resp.Body).Decode(&suites); err != nil {
		return nil, fmt.Errorf("ошибка декодирования сюит проекта %d: %w", projectID, err)
	}

	return suites, nil
}

// GetSuite получает одну тест-сюиту по ID.
func (c *HTTPClient) GetSuite(suiteID int64) (*data.Suite, error) {
	endpoint := fmt.Sprintf("get_suite/%d", suiteID)
	resp, err := c.Get(endpoint, nil)
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
func (c *HTTPClient) AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddSuiteRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_suite/%d", projectID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
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
func (c *HTTPClient) UpdateSuite(suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateSuiteRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_suite/%d", suiteID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
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
func (c *HTTPClient) DeleteSuite(suiteID int64) error {
	endpoint := fmt.Sprintf("delete_suite/%d", suiteID)
	resp, err := c.Post(endpoint, nil, nil)
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
