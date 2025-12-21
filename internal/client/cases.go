package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gotr/internal/models/data"
	"io"
	"net/http"
)

// GetCases получает список кейсов для проекта (пагинированный).
// Требует projectID, поддерживает фильтры (suite_id, section_id и т.д.).
func (c *HTTPClient) GetCases(projectID int64) (*data.GetCasesResponse, error) {
	endpoint := fmt.Sprintf("get_cases/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetCases для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetCasesResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetCases: %w", err)
	}
	return &result, nil
}

// GetCase получает информацию о конкретном кейсе по ID.
func (c *HTTPClient) GetCase(caseID int64) (*data.GetCaseResponse, error) {
	endpoint := fmt.Sprintf("get_case/%d", caseID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var result data.GetCaseResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetCase %d: %w", caseID, err)
	}
	return &result, nil
}

// GetHistoryForCase получает историю изменений кейса.
func (c *HTTPClient) GetHistoryForCase(caseID int64) (*data.GetHistoryForCaseResponse, error) {
	endpoint := fmt.Sprintf("get_history_for_case/%d", caseID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetHistoryForCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var result data.GetHistoryForCaseResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetHistoryForCase %d: %w", caseID, err)
	}
	return &result, nil
}

// AddCase создаёт новый кейс в секции.
// Требует sectionID и Title.
func (c *HTTPClient) AddCase(sectionID int64, req *data.AddCaseRequest) (*data.GetCaseResponse, error) {
	endpoint := fmt.Sprintf("add_case/%d", sectionID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddCaseRequest: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddCase: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetCaseResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа AddCase: %w", err)
	}
	return &result, nil
}

// UpdateCase обновляет существующий кейс.
// Поддерживает частичные обновления.
func (c *HTTPClient) UpdateCase(caseID int64, req *data.UpdateCaseRequest) (*data.GetCaseResponse, error) {
	endpoint := fmt.Sprintf("update_case/%d", caseID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateCaseRequest: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var result data.GetCaseResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа UpdateCase %d: %w", caseID, err)
	}
	return &result, nil
}

// UpdateCases — bulk-обновление кейсов в suite.
// Требует suiteID и список caseIDs.
func (c *HTTPClient) UpdateCases(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
	endpoint := fmt.Sprintf("update_cases/%d", suiteID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateCasesRequest: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateCases: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetCasesResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа UpdateCases: %w", err)
	}
	return &result, nil
}

// DeleteCase удаляет кейс по ID.
// Удаление необратимо.
func (c *HTTPClient) DeleteCase(caseID int64) error {
	endpoint := fmt.Sprintf("delete_case/%d", caseID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления кейса %d: %s, тело: %s", caseID, resp.Status, string(body))
	}
	return nil
}

// DeleteCases — bulk-удаление кейсов в suite.
func (c *HTTPClient) DeleteCases(suiteID int64, req *data.DeleteCasesRequest) error {
	endpoint := fmt.Sprintf("delete_cases/%d", suiteID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга DeleteCasesRequest: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteCases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка bulk-удаления кейсов в suite %d: %s, тело: %s", suiteID, resp.Status, string(body))
	}
	return nil
}

// GetCaseTypes получает список всех типов кейсов.
func (c *HTTPClient) GetCaseTypes() (*data.GetCaseTypesResponse, error) {
	endpoint := "get_case_types"
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetCaseTypes: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetCaseTypesResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetCaseTypes: %w", err)
	}
	return &result, nil
}

// GetCaseFields получает список всех полей кейсов.
func (c *HTTPClient) GetCaseFields() (*data.GetCaseFieldsResponse, error) {
	endpoint := "get_case_fields"
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetCaseFields: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetCaseFieldsResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetCaseFields: %w", err)
	}
	return &result, nil
}

// AddCaseField создаёт новое поле кейса.
func (c *HTTPClient) AddCaseField(req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddCaseFieldRequest: %w", err)
	}

	endpoint := "add_case_field"
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddCaseField: %w", err)
	}
	defer resp.Body.Close()

	var result data.AddCaseFieldResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа AddCaseField: %w", err)
	}
	return &result, nil
}