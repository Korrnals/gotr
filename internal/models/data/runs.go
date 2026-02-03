// internal/models/data/runs.go
package data

import "encoding/json"

// Run — тест-ран (набор тестов для выполнения)
// Соответствует документации TestRail API v2
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs
type Run struct {
	ID              int64           `json:"id"`                          // The unique ID of the test run
	SuiteID         int64           `json:"suite_id"`                    // The ID of the test suite this run belongs to
	Name            string          `json:"name"`                        // The name of the test run
	Description     string          `json:"description,omitempty"`       // The description of the test run
	MilestoneID     int64           `json:"milestone_id,omitempty"`      // The ID of the milestone this run is associated with
	AssignedTo      int64           `json:"assignedto_id,omitempty"`     // The ID of the user this run is assigned to
	ProjectID       int64           `json:"project_id"`                  // The ID of the project this run belongs to
	PlanID          int64           `json:"plan_id,omitempty"`           // The ID of the test plan this run belongs to (if any)
	ConfigIDs       []int64         `json:"config_ids,omitempty"`        // Array of configuration IDs used for this run
	Config          string          `json:"config,omitempty"`            // The configuration string (human-readable)
	PassedCount     int             `json:"passed_count"`                // The number of passed tests
	BlockedCount    int             `json:"blocked_count"`               // The number of blocked tests
	UntestedCount   int             `json:"untested_count"`              // The number of untested tests
	RetestCount     int             `json:"retest_count"`                // The number of retest tests
	FailedCount     int             `json:"failed_count"`                // The number of failed tests
	CustomStatusCount map[string]int `json:"custom_status_count,omitempty"` // Custom status counts
	CreatedBy       int64           `json:"created_by"`                  // The ID of the user who created the run
	CreatedOn       int64           `json:"created_on"`                  // The date/time when the run was created (Unix timestamp)
	UpdatedBy       int64           `json:"updated_by,omitempty"`        // The ID of the user who last updated the run
	UpdatedOn       int64           `json:"updated_on,omitempty"`        // The date/time when the run was last updated
	IsCompleted     bool            `json:"is_completed"`                // True if the test run was closed
	CompletedOn     int64           `json:"completed_on,omitempty"`      // The date/time when the run was closed (Unix timestamp)
	URL             string          `json:"url,omitempty"`               // The address/URL of the test run
	IncludeAll      bool            `json:"include_all"`                 // True if the run includes all test cases
	CaseIDs         []int64         `json:"case_ids,omitempty"`          // Array of case IDs included in this run (if include_all is false)
	CustomFields    json.RawMessage `json:"custom_fields,omitempty"`     // Custom fields (varies by project configuration)
}

// GetRunsResponse — ответ на get_runs (массив ранов)
type GetRunsResponse []Run

// GetRunResponse — ответ на get_run (один ран)
type GetRunResponse Run

// Request структуры для Runs API

// AddRunRequest — запрос для add_run
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#addrun
type AddRunRequest struct {
	Name        string  `json:"name"`                    // The name of the test run (required)
	Description string  `json:"description,omitempty"`   // The description of the test run
	SuiteID     int64   `json:"suite_id"`                // The ID of the test suite to create the run from (required)
	MilestoneID int64   `json:"milestone_id,omitempty"`  // The ID of the milestone to link the run to
	AssignedTo  int64   `json:"assignedto_id,omitempty"` // The ID of the user the test run should be assigned to
	IncludeAll  bool    `json:"include_all,omitempty"`   // True to include all test cases, false otherwise
	CaseIDs     []int64 `json:"case_ids,omitempty"`      // Array of case IDs to include (if include_all is false)
	ConfigIDs   []int64 `json:"config_ids,omitempty"`    // Array of configuration IDs for the test run
	Refs        string  `json:"refs,omitempty"`          // A string of references/requirements
}

// UpdateRunRequest — запрос для update_run
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#updaterun
type UpdateRunRequest struct {
	Name        *string `json:"name,omitempty"`          // The name of the test run
	Description *string `json:"description,omitempty"`   // The description of the test run
	MilestoneID *int64  `json:"milestone_id,omitempty"`  // The ID of the milestone
	AssignedTo  *int64  `json:"assignedto_id,omitempty"` // The ID of the user to assign to
	IncludeAll  *bool   `json:"include_all,omitempty"`   // True to include all test cases
	CaseIDs     []int64 `json:"case_ids,omitempty"`      // Array of case IDs to include
	Refs        *string `json:"refs,omitempty"`          // A string of references/requirements
}

// CloseRunRequest — запрос для close_run
// https://support.testrail.com/hc/en-us/articles/7077816294684-Runs#closerun
// Этот endpoint не требует тела запроса (POST с пустым телом)
type CloseRunRequest struct {
	// Пустая структура — close_run не требует параметров в body
}
