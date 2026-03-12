// internal/client/sections.go
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

// GetSections — получает section для suite in project (поддерживает пагинацию)
// suite_id обязательно для multi-suite проектов
func (c *HTTPClient) GetSections(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
	endpoint := fmt.Sprintf("get_sections/%d", projectID)
	var baseQuery map[string]string
	if suiteID != 0 {
		baseQuery = map[string]string{"suite_id": fmt.Sprintf("%d", suiteID)}
	}
	sections, err := fetchAllPages[data.Section](ctx, c, endpoint, baseQuery, "sections")
	if err != nil {
		return nil, fmt.Errorf("request error GetSections for project %d, suite %d: %w", projectID, suiteID, err)
	}
	return data.GetSectionsResponse(sections), nil
}

// GetSection — получает одну секцию по ID
func (c *HTTPClient) GetSection(ctx context.Context, sectionID int64) (*data.Section, error) {
	endpoint := fmt.Sprintf("get_section/%d", sectionID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s getting section %d: %s", resp.Status, sectionID, string(body))
	}

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("decode error section %d: %w", sectionID, err)
	}

	return &section, nil
}

// AddSection — создаёт новую секцию in suite проекта
func (c *HTTPClient) AddSection(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddSectionRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_section/%d", projectID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddSection for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s creating section in project %d: %s", resp.Status, projectID, string(body))
	}

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("decode error created section: %w", err)
	}

	return &section, nil
}

// UpdateSection — обновляет секцию (name, description, parent_id для перемещения)
func (c *HTTPClient) UpdateSection(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error UpdateSectionRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_section/%d", sectionID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s updating section %d: %s", resp.Status, sectionID, string(body))
	}

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("decode error updated section %d: %w", sectionID, err)
	}

	return &section, nil
}

// DeleteSection — удаляет секцию (необратимо, удаляет cases/results)
func (c *HTTPClient) DeleteSection(ctx context.Context, sectionID int64) error {
	endpoint := fmt.Sprintf("delete_section/%d", sectionID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s deleting section %d: %s", resp.Status, sectionID, string(body))
	}

	return nil
}
