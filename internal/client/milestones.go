// internal/client/milestones.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetMilestone получает информацию о milestone по ID
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#getmilestone
func (c *HTTPClient) GetMilestone(milestoneID int64) (*data.Milestone, error) {
	endpoint := fmt.Sprintf("get_milestone/%d", milestoneID)
	
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetMilestone для milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении milestone %d: %s",
			resp.Status, milestoneID, string(body))
	}

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("ошибка декодирования milestone: %w", err)
	}

	return &milestone, nil
}

// GetMilestones получает список milestone для проекта
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#getmilestones
func (c *HTTPClient) GetMilestones(projectID int64) ([]data.Milestone, error) {
	endpoint := fmt.Sprintf("get_milestones/%d", projectID)
	
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetMilestones для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении milestone для проекта %d: %s",
			resp.Status, projectID, string(body))
	}

	var milestones []data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestones); err != nil {
		return nil, fmt.Errorf("ошибка декодирования списка milestone: %w", err)
	}

	return milestones, nil
}

// AddMilestone создает новый milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#addmilestone
func (c *HTTPClient) AddMilestone(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
	if req == nil {
		return nil, fmt.Errorf("тело запроса обязательно")
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddMilestoneRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_milestone/%d", projectID)
	
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddMilestone для проекта %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при создании milestone для проекта %d: %s",
			resp.Status, projectID, string(body))
	}

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("ошибка декодирования milestone: %w", err)
	}

	return &milestone, nil
}

// UpdateMilestone обновляет milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#updatemilestone
func (c *HTTPClient) UpdateMilestone(milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
	if req == nil {
		return nil, fmt.Errorf("тело запроса обязательно")
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateMilestoneRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_milestone/%d", milestoneID)
	
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateMilestone для milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при обновлении milestone %d: %s",
			resp.Status, milestoneID, string(body))
	}

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("ошибка декодирования milestone: %w", err)
	}

	return &milestone, nil
}

// DeleteMilestone удаляет milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#deletemilestone
func (c *HTTPClient) DeleteMilestone(milestoneID int64) error {
	endpoint := fmt.Sprintf("delete_milestone/%d", milestoneID)
	
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteMilestone для milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API вернул %s при удалении milestone %d: %s",
			resp.Status, milestoneID, string(body))
	}

	return nil
}
