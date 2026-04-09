// internal/models/data/reports.go
package data

// ReportTemplate represents a report template.
type ReportTemplate struct {
	ID          int64  `json:"id"`          // Unique template ID
	Name        string `json:"name"`        // Template name
	Description string `json:"description"` // Description
}

// Report represents a generated report.
type Report struct {
	ID        int64  `json:"id"`         // Report ID
	Name      string `json:"name"`       // Name
	URL       string `json:"url"`        // Download URL
	Status    string `json:"status"`     // Status (completed, pending, error)
	CreatedOn int64  `json:"created_on"` // Creation timestamp
}

// GetReportsResponse is the response for get_reports.
type GetReportsResponse []ReportTemplate

// RunReportResponse is the response for run_report.
type RunReportResponse struct {
	ReportID int64  `json:"report_id"`
	URL      string `json:"url"`
	Status   string `json:"status"`
}
