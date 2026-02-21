// internal/models/data/plans.go
package data

// Plan — тест-план в TestRail
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans
type Plan struct {
	ID          int64       `json:"id"`                    // The unique ID of the test plan
	Name        string      `json:"name"`                  // The name of the test plan
	Description string      `json:"description,omitempty"` // The description of the test plan
	MilestoneID int64       `json:"milestone_id"`          // The ID of the milestone this test plan is associated with
	AssignedTo  int64       `json:"assignedto,omitempty"`  // The ID of the user the test plan is assigned to
	ProjectID   int64       `json:"project_id"`            // The ID of the project this test plan belongs to
	IsCompleted bool        `json:"is_completed"`          // True if the test plan is marked as completed
	CompletedOn Timestamp   `json:"completed_on,omitempty"` // The date/time when the test plan was closed
	PassedCount int         `json:"passed_count"`          // The number of tests in the test plan marked as passed
	BlockedCount int        `json:"blocked_count"`         // The number of tests in the test plan marked as blocked
	UntestedCount int      `json:"untested_count"`        // The number of tests in the test plan marked as untested
	RetestCount int         `json:"retest_count"`          // The number of tests in the test plan marked as retest
	FailedCount int         `json:"failed_count"`          // The number of tests in the test plan marked as failed
	CustomStatusCount map[string]int `json:"custom_status_count,omitempty"` // Custom statuses count
	Entries     []PlanEntry `json:"entries,omitempty"`     // An array of test plan entries (runs)
	URL         string      `json:"url,omitempty"`         // The address/URL of the test plan
	CreatedOn   Timestamp   `json:"created_on,omitempty"`  // The date/time when the test plan was created
	CreatedBy   int64       `json:"created_by,omitempty"`  // The ID of the user who created the test plan
}

// PlanEntry — запись (entry) в тест-плане
// Каждый entry представляет собой конфигурацию рана в плане
type PlanEntry struct {
	ID          string    `json:"id"`                    // The unique ID of the test plan entry
	Name        string    `json:"name"`                  // The name of the test plan entry
	Description string    `json:"description,omitempty"` // The description of the test plan entry
	SuiteID     int64     `json:"suite_id"`              // The ID of the test suite for this entry
	AssignedTo  int64     `json:"assignedto,omitempty"`  // The ID of the user the entry is assigned to
	IncludeAll  bool      `json:"include_all"`           // True if the entry includes all test cases
	CaseIDs     []int64   `json:"case_ids,omitempty"`    // Array of case IDs to include (if include_all is false)
	ConfigIDs   []int64   `json:"config_ids,omitempty"`  // Array of configuration IDs to use
	Runs        []Run     `json:"runs,omitempty"`        // Array of test runs generated from this entry
	CreatedOn   Timestamp `json:"created_on,omitempty"`  // The date/time when the entry was created
	CreatedBy   int64     `json:"created_by,omitempty"`  // The ID of the user who created the entry
}

// GetPlansResponse — ответ на get_plans
type GetPlansResponse []Plan

// GetPlanResponse — ответ на get_plan
type GetPlanResponse Plan

// AddPlanRequest — запрос на создание плана
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#addplan
type AddPlanRequest struct {
	Name        string           `json:"name"`                  // The name of the test plan (required)
	Description string           `json:"description,omitempty"` // The description of the test plan
	MilestoneID int64            `json:"milestone_id,omitempty"` // The ID of the milestone to link to
	Entries     []PlanEntryInput `json:"entries,omitempty"`     // An array of test plan entries to create
}

// PlanEntryInput — входные данные для создания entry в плане
type PlanEntryInput struct {
	Name        string  `json:"name"`                  // The name of the test plan entry (required)
	Description string  `json:"description,omitempty"` // The description of the test plan entry
	SuiteID     int64   `json:"suite_id"`              // The ID of the test suite (required)
	AssignedTo  int64   `json:"assignedto,omitempty"`  // The ID of the user the entry is assigned to
	IncludeAll  bool    `json:"include_all"`           // True to include all test cases
	CaseIDs     []int64 `json:"case_ids,omitempty"`    // Array of case IDs to include
	ConfigIDs   []int64 `json:"config_ids,omitempty"`  // Array of configuration IDs
	Runs        []RunInput `json:"runs,omitempty"`     // Array of configurations for multiple runs
}

// RunInput — входные данные для создания run в entry
type RunInput struct {
	ConfigIDs   []int64 `json:"config_ids,omitempty"`  // Array of configuration IDs for this run
	AssignedTo  int64   `json:"assignedto,omitempty"`  // The ID of the user the run is assigned to
	Description string  `json:"description,omitempty"` // The description of the run
	CaseIDs     []int64 `json:"case_ids,omitempty"`    // Array of case IDs for this run
}

// UpdatePlanRequest — запрос на обновление плана
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#updateplan
type UpdatePlanRequest struct {
	Name        string `json:"name,omitempty"`        // The name of the test plan
	Description string `json:"description,omitempty"` // The description of the test plan
	MilestoneID int64  `json:"milestone_id,omitempty"` // The ID of the milestone to link to
}

// AddPlanEntryRequest — запрос на добавление entry в существующий план
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#addplanentry
type AddPlanEntryRequest struct {
	Name        string  `json:"name"`                  // The name of the test plan entry (required)
	Description string  `json:"description,omitempty"` // The description of the test plan entry
	SuiteID     int64   `json:"suite_id"`              // The ID of the test suite (required)
	AssignedTo  int64   `json:"assignedto,omitempty"`  // The ID of the user the entry is assigned to
	IncludeAll  bool    `json:"include_all"`           // True to include all test cases
	CaseIDs     []int64 `json:"case_ids,omitempty"`    // Array of case IDs to include
	ConfigIDs   []int64 `json:"config_ids,omitempty"`  // Array of configuration IDs
}

// UpdatePlanEntryRequest — запрос на обновление entry в плане
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#updateplanentry
type UpdatePlanEntryRequest struct {
	Name        string  `json:"name,omitempty"`        // The name of the test plan entry
	Description string  `json:"description,omitempty"` // The description of the test plan entry
	AssignedTo  int64   `json:"assignedto,omitempty"`  // The ID of the user the entry is assigned to
	IncludeAll  bool    `json:"include_all"`           // True to include all test cases
	CaseIDs     []int64 `json:"case_ids,omitempty"`    // Array of case IDs to include
}

// ClosePlanRequest — запрос на закрытие плана (обычно пустой)
type ClosePlanRequest struct{}
