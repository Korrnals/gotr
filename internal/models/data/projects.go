package data

// Структура для 'get_project'
type GetProject struct {
	ID               int    `json:"id"`
	Announcement     string `json:"announcement"`
	CompletedOn      int    `json:"completed_on"`
	DefaultRoleID    int    `json:"default_role_id"`
	DefaultRole      string `json:"default_role"`
	IsCompleted      bool   `json:"is_completed"`
	Name             string `json:"name"`
	ShowAnnouncement bool   `json:"show_announcement"`
	SuiteMode        int    `json:"suite_mode"`
	URL              string `json:"url"`
	Users            []struct {
		ID            int `json:"id"`
		GlobalRoleID  any `json:"global_role_id"`
		GlobalRole    any `json:"global_role"`
		ProjectRoleID any `json:"project_role_id"`
		ProjectRole   any `json:"project_role"`
	} `json:"users"`
	Groups []any `json:"groups"`
}

// Структура для 'get_projects'
type GetProjectsResponse struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Size   int `json:"size"`
	Links  struct {
		Next any `json:"next"`
		Prev any `json:"prev"`
	} `json:"_links"`
	Projects []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"projects"`
}

// Структура для 'add_project'
type AddProjectRequest struct {
	Name             string `json:"name"`
	Announcement     string `json:"announcement"`
	ShowAnnouncement bool   `json:"show_announcement"`
}

// Структура для 'update_project'
type UpdateProjectResponse struct {
	Name             string `json:"name"`
	Announcement     string `json:"announcement"`
	ShowAnnouncement bool   `json:"show_announcement"`
	DefaultRoleID    int    `json:"default_role_id"`
	IsCompleted      bool   `json:"is_completed"`
	Users            []struct {
		UserID int `json:"user_id"`
		RoleID any `json:"role_id"`
	} `json:"users"`
	Groups []any `json:"groups"`
}
