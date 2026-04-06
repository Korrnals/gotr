// internal/client/milestones.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetMilestone fetches milestone info by ID.
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#getmilestone
func (c *HTTPClient) GetMilestone(ctx context.Context, milestoneID int64) (*data.Milestone, error) {
	endpoint := fmt.Sprintf("get_milestone/%d", milestoneID)

	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetMilestone for milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("decode error milestone: %w", err)
	}

	return &milestone, nil
}

// GetMilestones fetches the milestone list for a project (with pagination).
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#getmilestones
func (c *HTTPClient) GetMilestones(ctx context.Context, projectID int64) ([]data.Milestone, error) {
	endpoint := fmt.Sprintf("get_milestones/%d", projectID)
	milestones, err := fetchAllPages[data.Milestone](ctx, c, endpoint, nil, "milestones")
	if err != nil {
		return nil, fmt.Errorf("request error GetMilestones for project %d: %w", projectID, err)
	}
	return milestones, nil
}

// AddMilestone creates a new milestone.
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#addmilestone
func (c *HTTPClient) AddMilestone(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}

	bodyBytes, _ := json.Marshal(req)
	endpoint := fmt.Sprintf("add_milestone/%d", projectID)

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddMilestone for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("decode error milestone: %w", err)
	}

	return &milestone, nil
}

// UpdateMilestone updates a milestone.
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#updatemilestone
func (c *HTTPClient) UpdateMilestone(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}

	bodyBytes, _ := json.Marshal(req)
	endpoint := fmt.Sprintf("update_milestone/%d", milestoneID)

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateMilestone for milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	var milestone data.Milestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, fmt.Errorf("decode error milestone: %w", err)
	}

	return &milestone, nil
}

// DeleteMilestone deletes a milestone.
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#deletemilestone
func (c *HTTPClient) DeleteMilestone(ctx context.Context, milestoneID int64) error {
	endpoint := fmt.Sprintf("delete_milestone/%d", milestoneID)

	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteMilestone for milestone %d: %w", milestoneID, err)
	}
	defer resp.Body.Close()

	return nil
}
