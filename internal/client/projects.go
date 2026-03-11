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

// GetProjects получает список всех проектов.
// Возвращает массив проектов (TestRail возвращает []Project напрямую).
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

// GetProject получает информацию о конкретном проекте по ID.
// ID — число (int64).
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

// AddProject создаёт новый проект.
// Принимает AddProjectRequest с обязательным Name.
// Возвращает созданный проект.
func (c *HTTPClient) AddProject(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error AddProjectRequest: %w", err)
	}

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

// UpdateProject обновляет существующий проект по ID.
// Поддерживает частичные обновления.
// Требует прав администратора.
func (c *HTTPClient) UpdateProject(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
	endpoint := fmt.Sprintf("update_project/%d", projectID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal error UpdateProjectRequest: %w", err)
	}

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

// DeleteProject удаляет проект по ID.
// Удаление необратимо — все данные проекта теряются.
// Требует прав администратора.
// Возвращает nil при успехе.
func (c *HTTPClient) DeleteProject(ctx context.Context, projectID int64) error {
	endpoint := fmt.Sprintf("delete_project/%d", projectID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteProject %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete error project %d: %s, body: %s", projectID, resp.Status, string(body))
	}
	return nil
}
