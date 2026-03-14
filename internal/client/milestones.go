// internal/client/milestones.go
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

// GetMilestone получает информацию о milestone по ID
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#getmilestone
func (c *HTTPClient) GetMilestone(ctx context.Context, milestoneID int64) (*data.Milestone, error) {
	endpoint := fmt.Sprintf("get_milestone/%d", milestoneID)

	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetMilestone for milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s getting milestone %d: %s",
			resp.Status, milestoneID, string(body))
	}

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("decode error milestone: %w", err)
	}

	return &milestone, nil
}

// GetMilestones получает список milestone for project (поддерживает пагинацию)
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#getmilestones
func (c *HTTPClient) GetMilestones(ctx context.Context, projectID int64) ([]data.Milestone, error) {
	endpoint := fmt.Sprintf("get_milestones/%d", projectID)
	milestones, err := fetchAllPages[data.Milestone](ctx, c, endpoint, nil, "milestones")
	if err != nil {
		return nil, fmt.Errorf("request error GetMilestones for project %d: %w", projectID, err)
	}
	return milestones, nil
}

// AddMilestone создает новый milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#addmilestone
func (c *HTTPClient) AddMilestone(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddMilestoneRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_milestone/%d", projectID)

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddMilestone for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s creating milestone for project %d: %s",
			resp.Status, projectID, string(body))
	}

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("decode error milestone: %w", err)
	}

	return &milestone, nil
}

// UpdateMilestone обновляет milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#updatemilestone
func (c *HTTPClient) UpdateMilestone(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error UpdateMilestoneRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_milestone/%d", milestoneID)

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateMilestone for milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s updating milestone %d: %s",
			resp.Status, milestoneID, string(body))
	}

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("decode error milestone: %w", err)
	}

	return &milestone, nil
}

// DeleteMilestone удаляет milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#deletemilestone
func (c *HTTPClient) DeleteMilestone(ctx context.Context, milestoneID int64) error {
	endpoint := fmt.Sprintf("delete_milestone/%d", milestoneID)

	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteMilestone for milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s deleting milestone %d: %s",
			resp.Status, milestoneID, string(body))
	}

	return nil
}
