// internal/models/data/priorities.go
package data

// Priority represents a test case priority.
type Priority struct {
	ID        int64  `json:"id"`         // Unique priority ID
	Name      string `json:"name"`       // Priority name (e.g., "Critical", "High")
	ShortName string `json:"short_name"` // Short name
	Color     string `json:"color"`      // Color (HEX)
	IsDefault bool   `json:"is_default"` // Whether this is the default priority
	Priority  int    `json:"priority"`   // Numeric priority value (lower = higher priority)
}

// GetPrioritiesResponse is the response for get_priorities.
type GetPrioritiesResponse []Priority
