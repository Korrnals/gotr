// internal/models/data/attachments.go
package data

// Attachment represents a file attachment in TestRail.
type Attachment struct {
	ID          int64  `json:"id"`                    // Unique attachment ID
	Name        string `json:"name"`                  // File name
	Filename    string `json:"filename"`              // File name (alternative field)
	Size        int64  `json:"size"`                  // File size in bytes
	ContentType string `json:"content_type"`          // MIME content type
	CreatedOn   int64  `json:"created_on"`            // Creation timestamp
	ProjectID   int64  `json:"project_id"`            // Project ID
	CaseID      int64  `json:"case_id"`               // Case ID (if attached to a case)
	RunID       int64  `json:"run_id"`                // Run ID (if attached to a run)
	PlanID      int64  `json:"plan_id"`               // Plan ID (if attached to a plan)
	EntryID     string `json:"entry_id,omitempty"`    // Plan entry ID (if attached to a plan entry)
	TestID      int64  `json:"test_id,omitempty"`     // Test ID (if attached to a test)
	ResultID    int64  `json:"result_id"`             // Result ID (if attached to a result)
	UserID      int64  `json:"user_id"`               // ID of the user who uploaded the file
	EntityType  string `json:"entity_type,omitempty"` // TestRail 7.1+ cloud format: "case"|"run"|"plan"|"result"|...
	EntityID    string `json:"entity_id,omitempty"`   // TestRail 7.1+ cloud format: parent entity ID as string
}

// InferredEntityType returns the entity kind this attachment is bound to.
// Prefers the explicit EntityType field (TestRail 7.1+ cloud format) and
// falls back to inspecting the populated *_id fields. Returns one of
// "case", "run", "plan", "plan_entry", "result", "test", or "" when unknown.
func (a *Attachment) InferredEntityType() string {
	if a.EntityType != "" {
		return a.EntityType
	}
	switch {
	case a.ResultID != 0:
		return "result"
	case a.TestID != 0:
		return "test"
	case a.EntryID != "":
		return "plan_entry"
	case a.PlanID != 0:
		return "plan"
	case a.RunID != 0:
		return "run"
	case a.CaseID != 0:
		return "case"
	default:
		return ""
	}
}

// AttachmentResponse is the API response for adding an attachment.
type AttachmentResponse struct {
	AttachmentID int64  `json:"attachment_id"` // ID of the created attachment
	URL          string `json:"url"`           // URL for accessing the attachment
	Name         string `json:"name"`          // File name
	Size         int64  `json:"size"`          // File size
}

// GetAttachmentsResponse is the response for retrieving a list of attachments.
type GetAttachmentsResponse []Attachment
