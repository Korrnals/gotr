// models/data/suites.go
package data

// Suite — структура тест-сюиты
type Suite struct {
	ID          int64  `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ProjectID   int64  `json:"project_id,omitempty"`
	IsMaster    bool   `json:"is_master,omitempty"`
	IsBaseline  bool   `json:"is_baseline,omitempty"`
	IsCompleted bool   `json:"is_completed,omitempty"`
	CompletedOn int64  `json:"completed_on,omitempty"`
	URL         string `json:"url,omitempty"`
}

// GetSuitesResponse — ответ на get_suites (массив сюит напрямую)
type GetSuitesResponse []Suite

// GetSuiteResponse — ответ на get_suite (одна сюита напрямую)
type GetSuiteResponse Suite

// AddSuiteRequest — запрос для add_suite
type AddSuiteRequest struct {
	Name        string `json:"name"` // обязательно
	Description string `json:"description,omitempty"`
}

// UpdateSuiteRequest — запрос для update_suite
type UpdateSuiteRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsCompleted bool   `json:"is_completed,omitempty"`
}
