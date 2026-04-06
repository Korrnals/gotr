// internal/models/data/sections.go
package data

// Section represents a TestRail section.
type Section struct {
	Depth        int64  `json:"depth,omitempty"`         // Nesting depth
	Description  string `json:"description,omitempty"`   // Section description
	DisplayOrder int64  `json:"display_order,omitempty"` // Display order
	ID           int64  `json:"id"`                      // Unique ID (required)
	Name         string `json:"name"`                    // Section name (required)
	ParentID     int64  `json:"parent_id,omitempty"`     // Parent section ID (0 if none)
	SuiteID      int64  `json:"suite_id,omitempty"`      // Suite ID (may be 0 for baseline)
}

// GetSectionsResponse is the response for get_sections (array of sections).
type GetSectionsResponse []Section

// GetSectionResponse is the response for get_section (a single section).
type GetSectionResponse Section

// AddSectionRequest is the request for add_section (suite_id is required).
type AddSectionRequest struct {
	Name        string `json:"name"` // Required
	Description string `json:"description,omitempty"`
	SuiteID     int64  `json:"suite_id"` // Required
	ParentID    int64  `json:"parent_id,omitempty"`
}

// UpdateSectionRequest is the request for update_section (for modifying/moving).
type UpdateSectionRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    int64  `json:"parent_id,omitempty"`
}
