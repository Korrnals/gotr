// models/data/cases.go
package data

import (
	"encoding/json"
)

// Case is the primary structure for a single test case (used in get_case and get_cases).
type Case struct {
	ID                   int64   `json:"id"`
	Title                string  `json:"title,omitempty"`
	SectionID            int64   `json:"section_id,omitempty"`
	TemplateID           int64   `json:"template_id,omitempty"`
	TypeID               int64   `json:"type_id,omitempty"`
	PriorityID           int64   `json:"priority_id,omitempty"`
	MilestoneID          int64   `json:"milestone_id,omitempty"`
	Refs                 string  `json:"refs,omitempty"`
	CreatedBy            int64   `json:"created_by,omitempty"`
	CreatedOn            int64   `json:"created_on,omitempty"`
	UpdatedBy            int64   `json:"updated_by,omitempty"`
	UpdatedOn            int64   `json:"updated_on,omitempty"`
	Estimate             string  `json:"estimate,omitempty"`
	EstimateForecast     string  `json:"estimate_forecast,omitempty"`
	SuiteID              int64   `json:"suite_id,omitempty"`
	DisplayOrder         int     `json:"display_order,omitempty"`
	IsDeleted            int     `json:"is_deleted,omitempty"` // 1 = deleted, 0 = not deleted
	CustomAutomationType int64   `json:"custom_automation_type,omitempty"`
	CustomPreconds       string  `json:"custom_preconds,omitempty"`
	CustomSteps          string  `json:"custom_steps,omitempty"`
	CustomExpected       string  `json:"custom_expected,omitempty"`
	CustomStepsSeparated []Step  `json:"custom_steps_separated,omitempty"` // Structured steps (Step from shared.go)
	CustomMission        string  `json:"custom_mission,omitempty"`
	CustomGoals          string  `json:"custom_goals,omitempty"`
	Labels               []Label `json:"labels,omitempty"` // Label from shared.go
	// For fully custom fields (if TestRail returns an unknown field)
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// GetCasesResponse is the response for get_cases (paginated list of cases).
type GetCasesResponse []Case

// PaginatedCasesResponse is the wrapper format for TestRail API v2 (6.7+).
// Newer versions return:
//
//	{"offset": 0, "limit": 250, "size": 5000, "_links": {...}, "cases": [...]}
//
// Older versions return a flat []Case array.
// Used to quickly obtain Size (total count) with a single API call.
type PaginatedCasesResponse struct {
	Pagination
	Cases []Case `json:"cases"`
}

// GetCaseResponse is the response for get_case (a single case).
type GetCaseResponse Case

// GetHistoryForCaseResponse is the response for get_history_for_case.
type GetHistoryForCaseResponse struct {
	Pagination
	History []struct {
		ID        int64    `json:"id"`
		TypeID    int64    `json:"type_id"`
		CreatedOn int64    `json:"created_on"`
		UserID    int64    `json:"user_id"`
		Changes   []Change `json:"changes"`
	} `json:"history"`
}

// Change represents a single change entry in case history.
type Change struct {
	TypeID   int64  `json:"type_id"`
	OldText  string `json:"old_text,omitempty"`
	NewText  string `json:"new_text,omitempty"`
	Field    string `json:"field,omitempty"`
	OldValue int64  `json:"old_value,omitempty"`
	NewValue int64  `json:"new_value,omitempty"`
}

// Request structs //

// AddCaseRequest is the request for add_case.
type AddCaseRequest struct {
	Title                string `json:"title"`      // Required
	SectionID            int64  `json:"section_id"` // Explicit section assignment
	TypeID               int64  `json:"type_id"`
	PriorityID           int64  `json:"priority_id"`
	Estimate             string `json:"estimate,omitempty"`
	CustomPreconds       string `json:"custom_preconds,omitempty"`
	CustomSteps          string `json:"custom_steps,omitempty"`           // Text-format steps (alternative to CustomStepsSeparated)
	CustomExpected       string `json:"custom_expected,omitempty"`        // Expected result (text format)
	CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"` // Structured steps
	Refs                 string `json:"refs,omitempty"`
	MilestoneID          int64  `json:"milestone_id,omitempty"`
	TemplateID           int64  `json:"template_id,omitempty"`
}

// UpdateCaseRequest is the request for update_case (partial updates).
// Pointer fields distinguish "not set" from "empty value".
type UpdateCaseRequest struct {
	Title                *string `json:"title,omitempty"`
	TypeID               *int64  `json:"type_id,omitempty"` // Change the case type
	PriorityID           *int64  `json:"priority_id,omitempty"`
	Estimate             *string `json:"estimate,omitempty"`
	CustomPreconds       *string `json:"custom_preconds,omitempty"`
	CustomSteps          *string `json:"custom_steps,omitempty"`           // Text-format steps
	CustomExpected       *string `json:"custom_expected,omitempty"`        // Expected result
	CustomStepsSeparated []Step  `json:"custom_steps_separated,omitempty"` // Step from shared.go
	Refs                 *string `json:"refs,omitempty"`
	MilestoneID          *int64  `json:"milestone_id,omitempty"`
	SuiteID              *int64  `json:"suite_id,omitempty"`    // Move between suites
	SectionID            *int64  `json:"section_id,omitempty"`  // Move between sections
	TemplateID           *int64  `json:"template_id,omitempty"` // Change the template
}

// CopyCasesRequest is the request for copy_cases_to_section.
type CopyCasesRequest struct {
	CaseIDs []int64 `json:"case_ids"` // IDs of cases to copy
}

// MoveCasesRequest is the request for move_cases_to_section.
type MoveCasesRequest struct {
	CaseIDs []int64 `json:"case_ids"`           // IDs of cases to move
	SuiteID int64   `json:"suite_id,omitempty"` // Target suite ID (for cross-suite moves)
}

// UpdateCasesRequest is the request for bulk update_cases.
type UpdateCasesRequest struct {
	CaseIDs    []int64 `json:"case_ids"`
	PriorityID int64   `json:"priority_id,omitempty"`
	Estimate   string  `json:"estimate,omitempty"`
}

// DeleteCasesRequest is the request for delete_cases.
type DeleteCasesRequest struct {
	CaseIDs []int64 `json:"case_ids"`
}

// GetCaseTypesResponse is the response for get_case_types.
type GetCaseTypesResponse []struct {
	ID        int64  `json:"id"`
	IsDefault bool   `json:"is_default"`
	Name      string `json:"name"`
}

// GetCaseFieldsResponse is the response for get_case_fields.
type GetCaseFieldsResponse []struct {
	Configs []struct {
		Context struct {
			IsGlobal   bool    `json:"is_global"`
			ProjectIDs []int64 `json:"project_ids,omitempty"`
		} `json:"context"`
		ID      string `json:"id"`
		Options struct {
			DefaultValue string `json:"default_value,omitempty"`
			Format       string `json:"format,omitempty"`
			IsRequired   bool   `json:"is_required"`
			Rows         string `json:"rows,omitempty"`
		} `json:"options"`
	} `json:"configs"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
	ID           int64  `json:"id"`
	Label        string `json:"label"`
	Name         string `json:"name"`
	SystemName   string `json:"system_name"`
	TypeID       int64  `json:"type_id"`
}

// AddCaseFieldRequest is the request for add_case_field.
type AddCaseFieldRequest struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Configs     []struct {
		Context struct {
			IsGlobal   bool    `json:"is_global"`
			ProjectIDs []int64 `json:"project_ids,omitempty"`
		} `json:"context"`
		Options struct {
			IsRequired bool   `json:"is_required"`
			Items      string `json:"items,omitempty"`
		} `json:"options"`
	} `json:"configs"`
	IncludeAll bool `json:"include_all"`
}

// AddCaseFieldResponse is the response for add_case_field.
type AddCaseFieldResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SystemName   string `json:"system_name"`
	EntityID     int64  `json:"entity_id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	TypeID       int64  `json:"type_id"`
	LocationID   int64  `json:"location_id"`
	DisplayOrder int    `json:"display_order"`
	Configs      string `json:"configs"`
	IsMulti      int    `json:"is_multi"`
	IsActive     int    `json:"is_active"`
	StatusID     int    `json:"status_id"`
	IsSystem     int    `json:"is_system"`
	IncludeAll   int    `json:"include_all"`
	TemplateIDs  []any  `json:"template_ids"`
}

// DiffCasesResponse is the result of comparing cases between two projects.
type DiffCasesResponse struct {
	OnlyInFirst  []Case `json:"only_in_first"`  // Present only in pid1
	OnlyInSecond []Case `json:"only_in_second"` // Present only in pid2
	DiffByField  []struct {
		CaseID int64 `json:"case_id"`
		First  Case  `json:"first"`
		Second Case  `json:"second"`
	} `json:"diff_by_field"` // Differ by field value
}
