// models/data/suites.go
package data

// Suite represents a test suite.
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

// GetSuitesResponse is the response for get_suites (direct array of suites).
type GetSuitesResponse []Suite

// GetSuiteResponse is the response for get_suite (a single suite).
type GetSuiteResponse Suite

// AddSuiteRequest is the request for add_suite.
type AddSuiteRequest struct {
	Name        string `json:"name"` // Required
	Description string `json:"description,omitempty"`
}

// UpdateSuiteRequest is the request for update_suite.
type UpdateSuiteRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsCompleted bool   `json:"is_completed,omitempty"`
}
