// internal/models/data/milestones.go
package data



// Milestone — майлстоун (версия релиза) в TestRail
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones
type Milestone struct {
	ID          int64     `json:"id"`                     // The unique ID of the milestone
	Name        string    `json:"name"`                   // The name of the milestone
	Description string    `json:"description,omitempty"`  // The description of the milestone
	ProjectID   int64     `json:"project_id"`             // The ID of the project this milestone belongs to
	DueOn       Timestamp `json:"due_on,omitempty"`       // The due date of the milestone
	StartOn     Timestamp `json:"start_on,omitempty"`     // The start date of the milestone
	StartedOn   Timestamp `json:"started_on,omitempty"`   // The date when milestone was started
	IsStarted   bool      `json:"is_started"`             // True if the milestone has been started
	IsCompleted bool      `json:"is_completed"`           // True if the milestone is completed
	CompletedOn Timestamp `json:"completed_on,omitempty"` // The date when milestone was completed
	URL         string    `json:"url,omitempty"`          // The address/URL of the milestone
}

// GetMilestonesResponse — ответ на get_milestones
type GetMilestonesResponse []Milestone

// GetMilestoneResponse — ответ на get_milestone
type GetMilestoneResponse Milestone

// AddMilestoneRequest — запрос на создание milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#addmilestone
type AddMilestoneRequest struct {
	Name        string `json:"name"`                  // The name of the milestone (required)
	Description string `json:"description,omitempty"` // The description of the milestone
	DueOn       string `json:"due_on,omitempty"`      // The due date of the milestone (timestamp)
	StartOn     string `json:"start_on,omitempty"`    // The start date of the milestone (timestamp)
}

// UpdateMilestoneRequest — запрос на обновление milestone
// https://support.testrail.com/hc/en-us/articles/7077721635988-Milestones#updatemilestone
type UpdateMilestoneRequest struct {
	Name        string `json:"name,omitempty"`        // The name of the milestone
	Description string `json:"description,omitempty"` // The description of the milestone
	DueOn       string `json:"due_on,omitempty"`      // The due date of the milestone (timestamp)
	StartOn     string `json:"start_on,omitempty"`    // The start date of the milestone (timestamp)
	IsCompleted bool   `json:"is_completed"`          // True if the milestone is completed
	IsStarted   bool   `json:"is_started"`            // True if the milestone has been started
}
