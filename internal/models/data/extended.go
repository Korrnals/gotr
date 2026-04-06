// internal/models/data/extended.go
// Extended APIs: Groups, Roles, ResultFields, Variables, Datasets, BDDs, Labels

package data

// Group represents a user group.
type Group struct {
	ID        int64   `json:"id"`         // Unique group ID
	Name      string  `json:"name"`       // Group name
	UserIDs   []int64 `json:"user_ids"`   // User IDs in the group
	ProjectID int64   `json:"project_id"` // Project ID
}

// GetGroupsResponse is the response for get_groups.
type GetGroupsResponse []Group

// Role represents a user role.
type Role struct {
	ID   int64  `json:"id"`   // Unique role ID
	Name string `json:"name"` // Role name
}

// GetRolesResponse is the response for get_roles.
type GetRolesResponse []Role

// ResultField represents a result field definition.
type ResultField struct {
	ID         int64  `json:"id"`          // Unique field ID
	Name       string `json:"name"`        // Field name
	SystemName string `json:"system_name"` // System name
	TypeID     int    `json:"type_id"`     // Field type
	IsActive   bool   `json:"is_active"`   // Whether the field is active
}

// GetResultFieldsResponse is the response for get_result_fields.
type GetResultFieldsResponse []ResultField

// Dataset represents a data set for parameterized tests.
type Dataset struct {
	ID        int64  `json:"id"`         // Unique dataset ID
	Name      string `json:"name"`       // Dataset name
	ProjectID int64  `json:"project_id"` // Project ID
}

// GetDatasetsResponse is the response for get_datasets.
type GetDatasetsResponse []Dataset

// Variable represents a variable in a dataset.
type Variable struct {
	ID        int64  `json:"id"`         // Unique variable ID
	Name      string `json:"name"`       // Variable name
	DatasetID int64  `json:"dataset_id"` // Dataset ID
}

// GetVariablesResponse is the response for get_variables.
type GetVariablesResponse []Variable

// BDD represents a BDD scenario for a case.
type BDD struct {
	ID      int64  `json:"id"`      // Unique ID
	CaseID  int64  `json:"case_id"` // Case ID
	Content string `json:"content"` // BDD scenario content
}

// UpdateLabelsRequest is the request to update test labels.
type UpdateLabelsRequest struct {
	Labels []string `json:"labels"` // List of labels
}

// UpdateTestsLabelsRequest is the request to update labels for multiple tests.
type UpdateTestsLabelsRequest struct {
	TestIDs []int64  `json:"test_ids"` // Test IDs
	Labels  []string `json:"labels"`   // List of labels
}

// UpdateLabelRequest is the request to update a label.
type UpdateLabelRequest struct {
	ProjectID int64  `json:"project_id"` // Project ID
	Title     string `json:"title"`      // Label title (max 20 characters)
}

// GetLabelsResponse is the response for get_labels.
type GetLabelsResponse []Label
