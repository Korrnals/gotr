// internal/client/sections.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gotr/internal/models/data"
)

// GetSections — получает секции для suite в проекте (suite_id обязательно)
func (c *HTTPClient) GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error) {
	endpoint := fmt.Sprintf("get_sections/%d&suite_id=%d", projectID, suiteID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSections для проекта %d, suite %d: %w", projectID, suiteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении секций проекта %d, suite %d: %s", resp.Status, projectID, suiteID, string(body))
	}

	var sections data.GetSectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&sections); err != nil {
		return nil, fmt.Errorf("ошибка декодирования секций проекта %d, suite %d: %w", projectID, suiteID, err)
	}

	return sections, nil
}

// GetSection — получает одну секцию по ID
func (c *HTTPClient) GetSection(sectionID int64) (*data.Section, error) {
	endpoint := fmt.Sprintf("get_section/%d", sectionID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении секции %d: %s", resp.Status, sectionID, string(body))
	}

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("ошибка декодирования секции %d: %w", sectionID, err)
	}

	return &section, nil
}

// AddSection — создаёт новую секцию в suite проекта
func (c *HTTPClient) AddSection(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddSectionRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_section/%d", projectID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddSection для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при создании секции в проекте %d: %s", resp.Status, projectID, string(body))
	}

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("ошибка декодирования созданной секции: %w", err)
	}

	return &section, nil
}

// UpdateSection — обновляет секцию (name, description, parent_id для перемещения)
func (c *HTTPClient) UpdateSection(sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateSectionRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_section/%d", sectionID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при обновлении секции %d: %s", resp.Status, sectionID, string(body))
	}

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("ошибка декодирования обновлённой секции %d: %w", sectionID, err)
	}

	return &section, nil
}

// DeleteSection — удаляет секцию (необратимо, удаляет cases/results)
func (c *HTTPClient) DeleteSection(sectionID int64) error {
	endpoint := fmt.Sprintf("delete_section/%d", sectionID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API вернул %s при удалении секции %d: %s", resp.Status, sectionID, string(body))
	}

	return nil
}
