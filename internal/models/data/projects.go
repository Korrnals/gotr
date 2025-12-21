// models/data/projects.go
package data

import "encoding/json"

// Project — основная структура для одного проекта (get_project / get_projects)
type Project struct {
	ID               int64  `json:"id"`
	Name             string `json:"name,omitempty"`
	Announcement     string `json:"announcement,omitempty"`
	ShowAnnouncement bool   `json:"show_announcement,omitempty"`
	SuiteMode        int    `json:"suite_mode,omitempty"` // 1 - single suite, 2 - multiple suites, 3 - baselines
	URL              string `json:"url,omitempty"`
	IsCompleted      bool   `json:"is_completed,omitempty"`
	CompletedOn      int64  `json:"completed_on,omitempty"`
	DefaultRoleID    int64  `json:"default_role_id,omitempty"`
	CreatedBy        int64  `json:"created_by,omitempty"`
	CreatedOn        int64  `json:"created_on,omitempty"`
	UpdatedBy        int64  `json:"updated_by,omitempty"`
	UpdatedOn        int64  `json:"updated_on,omitempty"`
	Users            []struct {
		ID            int64 `json:"id"`
		GlobalRoleID  int64 `json:"global_role_id,omitempty"`
		ProjectRoleID int64 `json:"project_role_id,omitempty"`
	} `json:"users,omitempty"`
	Groups []struct {
		ID   int64  `json:"id"`
		Name string `json:"name,omitempty"`
	} `json:"groups,omitempty"`
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// GetProjectsResponse — ответ для get_projects
type GetProjectsResponse struct {
	Pagination
	Projects []Project `json:"projects"`
}

// GetProjectResponse — ответ для get_project
type GetProjectResponse struct {
	Project
}

// AddProjectRequest — запрос для add_project
type AddProjectRequest struct {
	Name             string `json:"name"` // Обязательное — без omitempty
	Announcement     string `json:"announcement,omitempty"`
	ShowAnnouncement bool   `json:"show_announcement,omitempty"`
	SuiteMode        int    `json:"suite_mode,omitempty"` // Опционально
}

// UpdateProjectRequest — запрос для update_project
type UpdateProjectRequest struct {
	Name             string `json:"name,omitempty"`
	Announcement     string `json:"announcement,omitempty"`
	ShowAnnouncement bool   `json:"show_announcement,omitempty"`
	IsCompleted      bool   `json:"is_completed,omitempty"`
	SuiteMode        int    `json:"suite_mode,omitempty"`
}
