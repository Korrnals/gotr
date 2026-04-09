// internal/models/data/configs.go
package data

// ConfigGroup represents a configuration group (e.g., "OS", "Browser").
type ConfigGroup struct {
	ID        int64    `json:"id"`         // Unique group ID
	Name      string   `json:"name"`       // Group name
	ProjectID int64    `json:"project_id"` // Project ID
	Configs   []Config `json:"configs"`    // List of configurations in the group
}

// Config represents a specific configuration (e.g., "Windows", "Chrome").
type Config struct {
	ID        int64  `json:"id"`         // Unique configuration ID
	Name      string `json:"name"`       // Configuration name
	GroupID   int64  `json:"group_id"`   // Group ID
	ProjectID int64  `json:"project_id"` // Project ID (optional)
}

// GetConfigsResponse is the response for get_configs (list of groups with configurations).
type GetConfigsResponse []ConfigGroup

// AddConfigGroupRequest is the request to create a configuration group.
type AddConfigGroupRequest struct {
	Name string `json:"name"` // Group name (required)
}

// AddConfigRequest is the request to create a configuration.
type AddConfigRequest struct {
	Name string `json:"name"` // Configuration name (required)
}

// UpdateConfigGroupRequest is the request to update a group.
type UpdateConfigGroupRequest struct {
	Name string `json:"name,omitempty"` // New group name
}

// UpdateConfigRequest is the request to update a configuration.
type UpdateConfigRequest struct {
	Name string `json:"name,omitempty"` // New configuration name
}
