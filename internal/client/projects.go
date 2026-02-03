package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Korrnals/gotr/internal/models/data"
	"io"
	"net/http"
)

// GetProjects получает список всех проектов.
// Возвращает массив проектов (TestRail возвращает []Project напрямую).
func (c *HTTPClient) GetProjects() (data.GetProjectsResponse, error) {
	endpoint := "get_projects"
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetProjects: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetProjectsResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetProjects: %w", err)
	}
	return result, nil
}

// GetProject получает информацию о конкретном проекте по ID.
// ID — число (int64).
func (c *HTTPClient) GetProject(projectID int64) (*data.GetProjectResponse, error) {
	endpoint := fmt.Sprintf("get_project/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetProject %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetProject %d: %w", projectID, err)
	}
	return &result, nil
}

// AddProject создаёт новый проект.
// Принимает AddProjectRequest с обязательным Name.
// Возвращает созданный проект.
func (c *HTTPClient) AddProject(req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddProjectRequest: %w", err)
	}

	endpoint := "add_project"
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddProject: %w", err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа AddProject: %w", err)
	}
	return &result, nil
}

// UpdateProject обновляет существующий проект по ID.
// Поддерживает частичные обновления.
// Требует прав администратора.
func (c *HTTPClient) UpdateProject(projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
	endpoint := fmt.Sprintf("update_project/%d", projectID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateProjectRequest: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateProject %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа UpdateProject %d: %w", projectID, err)
	}
	return &result, nil
}

// DeleteProject удаляет проект по ID.
// Удаление необратимо — все данные проекта теряются.
// Требует прав администратора.
// Возвращает nil при успехе.
func (c *HTTPClient) DeleteProject(projectID int64) error {
	endpoint := fmt.Sprintf("delete_project/%d", projectID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteProject %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления проекта %d: %s, тело: %s", projectID, resp.Status, string(body))
	}
	return nil
}
