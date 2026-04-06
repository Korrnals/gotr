package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetProjects fetches the list of all projects.
// Returns an array of projects (TestRail returns []Project directly).
func (c *HTTPClient) GetProjects(ctx context.Context) (data.GetProjectsResponse, error) {
	endpoint := "get_projects"
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetProjects: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetProjectsResponse
	if err := c.ReadJSONResponse(ctx, resp, &result); err != nil {
		return nil, fmt.Errorf("decode error response GetProjects: %w", err)
	}
	return result, nil
}

// GetProject fetches a specific project by ID.
func (c *HTTPClient) GetProject(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
	endpoint := fmt.Sprintf("get_project/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetProject %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(ctx, resp, &result); err != nil {
		return nil, fmt.Errorf("decode error response GetProject %d: %w", projectID, err)
	}
	return &result, nil
}

// AddProject creates a new project.
// Requires Name in the AddProjectRequest.
// Returns the created project.
func (c *HTTPClient) AddProject(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
	bodyBytes, _ := json.Marshal(req)
	endpoint := "add_project"
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddProject: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(ctx, resp, &result); err != nil {
		return nil, fmt.Errorf("decode error response AddProject: %w", err)
	}
	return &result, nil
}

// UpdateProject updates an existing project by ID.
// Supports partial updates.
// Requires admin permissions.
func (c *HTTPClient) UpdateProject(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
	endpoint := fmt.Sprintf("update_project/%d", projectID)
	bodyBytes, _ := json.Marshal(req)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateProject %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(ctx, resp, &result); err != nil {
		return nil, fmt.Errorf("decode error response UpdateProject %d: %w", projectID, err)
	}
	return &result, nil
}

// DeleteProject deletes a project by ID.
// This is irreversible — all project data will be lost.
// Requires admin permissions.
// Returns nil on success.
func (c *HTTPClient) DeleteProject(ctx context.Context, projectID int64) error {
	endpoint := fmt.Sprintf("delete_project/%d", projectID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteProject %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	return nil
}
