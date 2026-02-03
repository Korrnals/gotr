// internal/models/data/tests.go
package data

import "encoding/json"

// Test — тест (конкретный тест-кейс в рамках тест-рана)
// Соответствует документации TestRail API v2
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests
type Test struct {
	ID            int64           `json:"id"`                      // The unique ID of the test
	CaseID        int64           `json:"case_id"`                 // The ID of the related test case
	RunID         int64           `json:"run_id"`                  // The ID of the test run this test belongs to
	StatusID      int64           `json:"status_id,omitempty"`     // The ID of the current test status
	Title         string          `json:"title"`                   // The title of the test (inherited from test case)
	AssignedTo    int64           `json:"assignedto_id,omitempty"` // The ID of the user this test is assigned to
	TypeID        int64           `json:"type_id,omitempty"`       // The ID of the test case type
	PriorityID    int64           `json:"priority_id,omitempty"`   // The ID of the test case priority
	Estimate      string          `json:"estimate,omitempty"`      // The estimated time (e.g., "30m")
	EstimateForecast string       `json:"estimate_forecast,omitempty"` // The forecast time
	Refs          string          `json:"refs,omitempty"`          // A comma-separated list of references
	MilestoneID   int64           `json:"milestone_id,omitempty"`  // The ID of the milestone
	CustomFields  json.RawMessage `json:"custom_fields,omitempty"` // Custom fields from the test case
}

// GetTestsResponse — ответ на get_tests (массив тестов)
type GetTestsResponse []Test

// GetTestResponse — ответ на get_test (один тест)
type GetTestResponse Test

// Request структуры для Tests API

// UpdateTestRequest — запрос для update_test
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests#updatetest
// Позволяет изменить статус и назначение теста
type UpdateTestRequest struct {
	StatusID   int64 `json:"status_id,omitempty"`     // The ID of the test status
	AssignedTo int64 `json:"assignedto_id,omitempty"` // The ID of the user to assign the test to
}
