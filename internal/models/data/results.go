// internal/models/data/results.go
package data

import "encoding/json"

// Result — результат выполнения теста
// Соответствует документации TestRail API v2
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results
type Result struct {
	ID           int64           `json:"id"`                         // The unique ID of the test result
	TestID       int64           `json:"test_id"`                    // The ID of the test this result belongs to
	StatusID     int64           `json:"status_id"`                  // The ID of the test status
	CreatedBy    int64           `json:"created_by"`                 // The ID of the user who created the test result
	CreatedOn    int64           `json:"created_on"`                 // The date/time when the test result was created (Unix timestamp)
	AssignedTo   int64           `json:"assignedto_id,omitempty"`    // The ID of the user the test is assigned to
	Comment      string          `json:"comment,omitempty"`          // A comment related to the test result
	Version      string          `json:"version,omitempty"`          // The version or build tested against
	Elapsed      string          `json:"elapsed,omitempty"`          // The amount of time it took to execute the test (e.g., "30m" or "1h 30m")
	Defects      string          `json:"defects,omitempty"`          // A comma-separated list of defects linked to the test result
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`    // Custom fields (varies by project configuration)
}

// GetResultsResponse — ответ на get_results (массив результатов)
type GetResultsResponse []Result

// GetResultResponse — ответ на get_result (один результат)
type GetResultResponse Result

// Request структуры для Results API

// AddResultRequest — запрос для add_result
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresult
type AddResultRequest struct {
	StatusID   int64  `json:"status_id"`                  // The ID of the test status (required)
	Comment    string `json:"comment,omitempty"`          // A comment for the test result
	Version    string `json:"version,omitempty"`          // The version or build tested against
	Elapsed    string `json:"elapsed,omitempty"`          // The time it took to execute, e.g., "30m" or "1h 30m"
	Defects    string `json:"defects,omitempty"`          // A comma-separated list of defects to link
	AssignedTo int64  `json:"assignedto_id,omitempty""`   // The ID of the user the test should be assigned to
	// Custom fields are added dynamically via json.RawMessage if needed
}

// AddResultForCaseRequest — запрос для add_result_for_case
// Использует те же поля что и AddResultRequest
type AddResultForCaseRequest AddResultRequest

// AddResultsRequest — запрос для bulk add_results
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresults
type AddResultsRequest struct {
	Results []ResultEntry `json:"results"` // Array of result entries
}

// ResultEntry — отдельная запись результата для bulk операций
type ResultEntry struct {
	TestID     int64  `json:"test_id"`                    // The ID of the test
	StatusID   int64  `json:"status_id"`                  // The ID of the test status
	Comment    string `json:"comment,omitempty"`          // A comment for the test result
	Version    string `json:"version,omitempty"`          // The version or build tested against
	Elapsed    string `json:"elapsed,omitempty"`          // The time it took to execute
	Defects    string `json:"defects,omitempty"`          // A comma-separated list of defects
	AssignedTo int64  `json:"assignedto_id,omitempty"`    // The ID of the user to assign to
}

// AddResultsForCasesRequest — запрос для add_results_for_cases
// https://support.testrail.com/hc/en-us/articles/7077874763156-Results#addresultsforcases
type AddResultsForCasesRequest struct {
	Results []ResultForCaseEntry `json:"results"` // Array of result entries with case_id
}

// ResultForCaseEntry — отдельная запись результата для кейса (используется с run_id)
type ResultForCaseEntry struct {
	CaseID     int64  `json:"case_id"`                    // The ID of the test case
	StatusID   int64  `json:"status_id"`                  // The ID of the test status
	Comment    string `json:"comment,omitempty"`          // A comment for the test result
	Version    string `json:"version,omitempty"`          // The version or build tested against
	Elapsed    string `json:"elapsed,omitempty"`          // The time it took to execute
	Defects    string `json:"defects,omitempty"`          // A comma-separated list of defects
	AssignedTo int64  `json:"assignedto_id,omitempty"`    // The ID of the user to assign to
}
