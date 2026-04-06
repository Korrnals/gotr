// internal/models/data/templates.go
package data

// Template represents a test case template.
type Template struct {
	ID        int64  `json:"id"`         // Unique template ID
	Name      string `json:"name"`       // Template name
	IsDefault bool   `json:"is_default"` // Whether this is the default template
}

// GetTemplatesResponse is the response for get_templates.
type GetTemplatesResponse []Template
