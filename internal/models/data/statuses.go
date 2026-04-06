// internal/models/data/statuses.go
package data

// Status represents a test result status.
type Status struct {
	ID          int64  `json:"id"`           // Unique status ID
	Name        string `json:"name"`         // Status name (e.g., "Passed", "Failed")
	Label       string `json:"label"`        // Display label
	ColorBright int64  `json:"color_bright"` // Bright color
	ColorDark   int64  `json:"color_dark"`   // Dark color
	ColorMedium int64  `json:"color_medium"` // Medium color
	IsFinal     bool   `json:"is_final"`     // Whether this is a final status
	IsSystem    bool   `json:"is_system"`    // Whether this is a system status
	IsUntested  bool   `json:"is_untested"` // Whether this is the "untested" status
}

// GetStatusesResponse is the response for get_statuses.
type GetStatusesResponse []Status
