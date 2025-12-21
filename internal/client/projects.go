package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gotr/internal/models/data"
	"io"
	"net/http"
)

// GetProjects получает список всех проектов.
// Не требует параметров, возвращает пагинированный список.
func (c *HTTPClient) GetProjects() (*data.GetProjectsResponse, error) {
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
	return &result, nil
}

// GetProject получает информацию о конкретном проекте по ID.
// ID должен быть валидным числом.
func (c *HTTPClient) GetProject(projectID string) (*data.GetProjectResponse, error) {
	endpoint := fmt.Sprintf("get_project/%s", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetProject %s: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа GetProject %s: %w", projectID, err)
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
// Поддерживает частичные обновления (UpdateProjectRequest).
// Требует прав администратора.
func (c *HTTPClient) UpdateProject(projectID string, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
	endpoint := fmt.Sprintf("update_project/%s", projectID)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateProjectRequest: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateProject %s: %w", projectID, err)
	}
	defer resp.Body.Close()

	var result data.GetProjectResponse
	if err := c.ReadJSONResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа UpdateProject %s: %w", projectID, err)
	}
	return &result, nil
}

// DeleteProject удаляет проект по ID.
// Удаление необратимо — все данные проекта (тесты, runs, результаты) теряются.
// Требует прав администратора.
// Возвращает nil при успехе (HTTP 200 OK).
func (c *HTTPClient) DeleteProject(projectID string) error {
	endpoint := fmt.Sprintf("delete_project/%s", projectID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteProject %s: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления проекта %s: %s, тело: %s", projectID, resp.Status, string(body))
	}

	return nil
}