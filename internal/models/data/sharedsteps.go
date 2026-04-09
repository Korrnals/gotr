// models/data/sharedsteps.go
package data

import (
	"encoding/json"
)

// SharedStep is the primary structure for a shared step.
type SharedStep struct {
	ID                   int64   `json:"id"`
	Title                string  `json:"title,omitempty"`
	ProjectID            int64   `json:"project_id,omitempty"`
	CreatedBy            int64   `json:"created_by,omitempty"`
	CreatedOn            int64   `json:"created_on,omitempty"`
	UpdatedBy            int64   `json:"updated_by,omitempty"`
	UpdatedOn            int64   `json:"updated_on,omitempty"`
	CustomStepsSeparated []Step  `json:"custom_steps_separated,omitempty"` // Step from shared.go
	CaseIDs              []int64 `json:"case_ids,omitempty"`               // Cases using this shared step
	// For unknown extensions
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// GetSharedStepsResponse is the response for get_shared_steps (array of steps).
type GetSharedStepsResponse []SharedStep

// GetSharedStepResponse is the response for get_shared_step (a single step).
type GetSharedStepResponse SharedStep

// GetSharedStepHistoryResponse is the response for get_shared_step_history.
type GetSharedStepHistoryResponse struct {
	Pagination
	History []struct {
		ID                   int64  `json:"id"`
		Timestamp            int64  `json:"timestamp"`
		UserID               int64  `json:"user_id"`
		CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"`
		Title                string `json:"title,omitempty"`
	} `json:"step_history"`
}

// Request structs

// AddSharedStepRequest is the request for add_shared_step.
type AddSharedStepRequest struct {
	Title                string `json:"title"` // Required
	CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"`
}

// UpdateSharedStepRequest is the request for update_shared_step.
type UpdateSharedStepRequest struct {
	Title                string `json:"title,omitempty"`
	CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"`
}

// DeleteSharedStepRequest is the request for delete_shared_step.
type DeleteSharedStepRequest struct {
	KeepInCases int `json:"keep_in_cases"` // 1 = keep in cases, 0 = remove
}
