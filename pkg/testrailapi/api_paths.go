// Package testrailapi provides a complete structured representation of all TestRail API v2 endpoints.
// Grouped by resources for easy access and reuse (e.g., in CLI tools like gotr).
//
// Usage:
//
//	api := testrailapi.New()
//	allPaths := api.Paths() // For -list-paths flag
//	casesPaths := api.Cases.Paths() // Specific resource only
package testrailapi

// APIPath represents a single TestRail API v2 endpoint.
type APIPath struct {
	Method      string            // HTTP method: GET, POST, etc.
	URI         string            // Full URI (e.g., index.php?/api/v2/get_case/{case_id})
	Description string            // Short description
	Params      map[string]string // Path or required params (name -> desc); nil if none
}

// TestRailAPI root structure with all resources.
type TestRailAPI struct {
	Attachments    *Attachments
	Cases          *Cases
	CaseFields     *CaseFields
	CaseTypes      *CaseTypes
	Configurations *Configurations
	Milestones     *Milestones
	Plans          *Plans
	Priorities     *Priorities
	Projects       *Projects
	Results        *Results
	ResultFields   *ResultFields
	Runs           *Runs
	Sections       *Sections
	Statuses       *Statuses
	Suites         *Suites
	Tests          *Tests
	Users          *Users
	Reports        *Reports
	Roles          *Roles
	Templates      *Templates
	Groups         *Groups
	SharedSteps    *SharedSteps
	Variables      *Variables
	Labels         *Labels
	Datasets       *Datasets
	BDDs           *BDDs
}

// New initializes a fully populated TestRailAPI.
func New() *TestRailAPI {
	return &TestRailAPI{
		Attachments:    &Attachments{},
		Cases:          &Cases{},
		CaseFields:     &CaseFields{},
		CaseTypes:      &CaseTypes{},
		Configurations: &Configurations{},
		Milestones:     &Milestones{},
		Plans:          &Plans{},
		Priorities:     &Priorities{},
		Projects:       &Projects{},
		Results:        &Results{},
		ResultFields:   &ResultFields{},
		Runs:           &Runs{},
		Sections:       &Sections{},
		Statuses:       &Statuses{},
		Suites:         &Suites{},
		Tests:          &Tests{},
		Users:          &Users{},
		Reports:        &Reports{},
		Roles:          &Roles{},
		Templates:      &Templates{},
		Groups:         &Groups{},
		SharedSteps:    &SharedSteps{},
		Variables:      &Variables{},
		Labels:         &Labels{},
		Datasets:       &Datasets{},
		BDDs:           &BDDs{},
	}
}

// Paths returns all endpoints (perfect for listing in CLI).
func (api *TestRailAPI) Paths() []APIPath {
	var all []APIPath
	all = append(all, api.Attachments.Paths()...)
	all = append(all, api.Cases.Paths()...)
	all = append(all, api.CaseFields.Paths()...)
	all = append(all, api.CaseTypes.Paths()...)
	all = append(all, api.Configurations.Paths()...)
	all = append(all, api.Milestones.Paths()...)
	all = append(all, api.Plans.Paths()...)
	all = append(all, api.Priorities.Paths()...)
	all = append(all, api.Projects.Paths()...)
	all = append(all, api.Results.Paths()...)
	all = append(all, api.ResultFields.Paths()...)
	all = append(all, api.Runs.Paths()...)
	all = append(all, api.Sections.Paths()...)
	all = append(all, api.Statuses.Paths()...)
	all = append(all, api.Suites.Paths()...)
	all = append(all, api.Tests.Paths()...)
	all = append(all, api.Users.Paths()...)
	all = append(all, api.Reports.Paths()...)
	all = append(all, api.Roles.Paths()...)
	all = append(all, api.Templates.Paths()...)
	all = append(all, api.SharedSteps.Paths()...)
	all = append(all, api.Variables.Paths()...)
	all = append(all, api.Labels.Paths()...)
	all = append(all, api.Datasets.Paths()...)
	all = append(all, api.BDDs.Paths()...)
	return all
}

// --- Resources ---

// Attachments lists API endpoints for the corresponding TestRail resource.
type Attachments struct{}

// Paths returns API endpoints for Attachments.
func (r *Attachments) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_attachment/{attachment_id}", Description: "Get single attachment", Params: map[string]string{"attachment_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_attachments_for_case/{case_id}", Description: "Get attachments for case", Params: map[string]string{"case_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_attachments_for_plan/{plan_id}", Description: "Get attachments for plan", Params: map[string]string{"plan_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_attachments_for_plan_entry/{plan_id}/{entry_id}", Description: "Get attachments for plan entry"},
		{Method: "GET", URI: "index.php?/api/v2/get_attachments_for_run/{run_id}", Description: "Get attachments for run", Params: map[string]string{"run_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_attachments_for_test/{test_id}", Description: "Get attachments for test", Params: map[string]string{"test_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_attachments_for_project/{project_id}", Description: "Get attachments for project", Params: map[string]string{"project_id": "Required"}},
		{Method: "POST", URI: "index.php?/api/v2/add_attachment_to_case/{case_id}", Description: "Add attachment to case"},
		{Method: "POST", URI: "index.php?/api/v2/add_attachment_to_plan/{plan_id}", Description: "Add attachment to plan"},
		{Method: "POST", URI: "index.php?/api/v2/add_attachment_to_plan_entry/{plan_id}/{entry_id}", Description: "Add attachment to plan entry"},
		{Method: "POST", URI: "index.php?/api/v2/add_attachment_to_result/{result_id}", Description: "Add attachment to result"},
		{Method: "POST", URI: "index.php?/api/v2/add_attachment_to_run/{run_id}", Description: "Add attachment to run"},
		{Method: "POST", URI: "index.php?/api/v2/delete_attachment/{attachment_id}", Description: "Delete attachment", Params: map[string]string{"attachment_id": "Required"}},
	}
}

// Cases lists API endpoints for the corresponding TestRail resource.
type Cases struct{}

// Paths returns API endpoints for Cases.
func (r *Cases) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_case/{case_id}", Description: "Get single test case", Params: map[string]string{"case_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_cases/{project_id}", Description: "Get list of test cases for project (supports filters like suite_id, section_id)"},
		{Method: "GET", URI: "index.php?/api/v2/get_history_for_case/{case_id}", Description: "Get edit history for test case"},
		{Method: "POST", URI: "index.php?/api/v2/add_case/{section_id}", Description: "Create new test case"},
		{Method: "POST", URI: "index.php?/api/v2/copy_cases_to_section/{section_id}", Description: "Copy cases to another section"},
		{Method: "POST", URI: "index.php?/api/v2/move_cases_to_section/{section_id}", Description: "Move cases to another section"},
		{Method: "POST", URI: "index.php?/api/v2/update_case/{case_id}", Description: "Update existing test case"},
		{Method: "POST", URI: "index.php?/api/v2/update_cases/{suite_id}", Description: "Bulk update test cases"},
		{Method: "POST", URI: "index.php?/api/v2/delete_case/{case_id}", Description: "Delete test case"},
		{Method: "POST", URI: "index.php?/api/v2/delete_cases/{suite_id}", Description: "Bulk delete test cases"},
	}
}

// CaseFields lists API endpoints for the corresponding TestRail resource.
type CaseFields struct{}

// Paths returns API endpoints for CaseFields.
func (r *CaseFields) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_case_fields", Description: "Get list of available case fields"},
		{Method: "POST", URI: "index.php?/api/v2/add_case_field", Description: "Create new custom case field"},
	}
}

// CaseTypes lists API endpoints for the corresponding TestRail resource.
type CaseTypes struct{}

// Paths returns API endpoints for CaseTypes.
func (r *CaseTypes) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_case_types", Description: "Get list of available case types"},
	}
}

// Configurations lists API endpoints for the corresponding TestRail resource.
type Configurations struct{}

// Paths returns API endpoints for Configurations.
func (r *Configurations) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_configs/{project_id}", Description: "Get configurations for project"},
		{Method: "POST", URI: "index.php?/api/v2/add_config_group/{project_id}", Description: "Add config group"},
		{Method: "POST", URI: "index.php?/api/v2/add_config/{config_group_id}", Description: "Add config to group"},
		{Method: "POST", URI: "index.php?/api/v2/update_config_group/{config_group_id}", Description: "Update config group"},
		{Method: "POST", URI: "index.php?/api/v2/update_config/{config_id}", Description: "Update config"},
		{Method: "POST", URI: "index.php?/api/v2/delete_config_group/{config_group_id}", Description: "Delete config group"},
		{Method: "POST", URI: "index.php?/api/v2/delete_config/{config_id}", Description: "Delete config"},
	}
}

// Milestones lists API endpoints for the corresponding TestRail resource.
type Milestones struct{}

// Paths returns API endpoints for Milestones.
func (r *Milestones) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_milestone/{milestone_id}", Description: "Get single milestone"},
		{Method: "GET", URI: "index.php?/api/v2/get_milestones/{project_id}", Description: "Get milestones for project"},
		{Method: "POST", URI: "index.php?/api/v2/add_milestone/{project_id}", Description: "Create milestone"},
		{Method: "POST", URI: "index.php?/api/v2/update_milestone/{milestone_id}", Description: "Update milestone"},
		{Method: "POST", URI: "index.php?/api/v2/delete_milestone/{milestone_id}", Description: "Delete milestone"},
	}
}

// Plans lists API endpoints for the corresponding TestRail resource.
type Plans struct{}

// Paths returns API endpoints for Plans.
func (r *Plans) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_plan/{plan_id}", Description: "Get single test plan"},
		{Method: "GET", URI: "index.php?/api/v2/get_plans/{project_id}", Description: "Get test plans for project"},
		{Method: "POST", URI: "index.php?/api/v2/add_plan/{project_id}", Description: "Create test plan"},
		{Method: "POST", URI: "index.php?/api/v2/add_plan_entry/{plan_id}", Description: "Add entry to plan"},
		{Method: "POST", URI: "index.php?/api/v2/update_plan/{plan_id}", Description: "Update test plan"},
		{Method: "POST", URI: "index.php?/api/v2/update_plan_entry/{plan_id}/{entry_id}", Description: "Update plan entry"},
		{Method: "POST", URI: "index.php?/api/v2/close_plan/{plan_id}", Description: "Close test plan"},
		{Method: "POST", URI: "index.php?/api/v2/delete_plan/{plan_id}", Description: "Delete test plan"},
		{Method: "POST", URI: "index.php?/api/v2/delete_plan_entry/{plan_id}/{entry_id}", Description: "Delete plan entry"},
	}
}

// Priorities lists API endpoints for the corresponding TestRail resource.
type Priorities struct{}

// Paths returns API endpoints for Priorities.
func (r *Priorities) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_priorities", Description: "Get list of available priorities"},
	}
}

// Projects lists API endpoints for the corresponding TestRail resource.
type Projects struct{}

// Paths returns API endpoints for Projects.
func (r *Projects) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_project/{project_id}", Description: "Get single project"},
		{Method: "GET", URI: "index.php?/api/v2/get_projects", Description: "Get all projects (with filters)"},
		{Method: "POST", URI: "index.php?/api/v2/add_project", Description: "Create new project"},
		{Method: "POST", URI: "index.php?/api/v2/update_project/{project_id}", Description: "Update project"},
		{Method: "POST", URI: "index.php?/api/v2/delete_project/{project_id}", Description: "Delete (archive) project"},
	}
}

// Results lists API endpoints for the corresponding TestRail resource.
type Results struct{}

// Paths returns API endpoints for Results.
func (r *Results) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_results/{test_id}", Description: "Get results for test"},
		{Method: "GET", URI: "index.php?/api/v2/get_results_for_case/{run_id}/{case_id}", Description: "Get results for case in run"},
		{Method: "GET", URI: "index.php?/api/v2/get_results_for_run/{run_id}", Description: "Get results for run"},
		{Method: "POST", URI: "index.php?/api/v2/add_result/{test_id}", Description: "Add result for test"},
		{Method: "POST", URI: "index.php?/api/v2/add_result_for_case/{run_id}/{case_id}", Description: "Add result for case in run"},
		{Method: "POST", URI: "index.php?/api/v2/add_results/{run_id}", Description: "Bulk add results for run"},
		{Method: "POST", URI: "index.php?/api/v2/add_results_for_cases/{run_id}", Description: "Bulk add results for cases"},
	}
}

// ResultFields lists API endpoints for the corresponding TestRail resource.
type ResultFields struct{}

// Paths returns API endpoints for ResultFields.
func (r *ResultFields) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_result_fields", Description: "Get list of available result fields"},
	}
}

// Runs lists API endpoints for the corresponding TestRail resource.
type Runs struct{}

// Paths returns API endpoints for Runs.
func (r *Runs) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_run/{run_id}", Description: "Get single test run"},
		{Method: "GET", URI: "index.php?/api/v2/get_runs/{project_id}", Description: "Get test runs for project"},
		{Method: "POST", URI: "index.php?/api/v2/add_run/{project_id}", Description: "Create test run"},
		{Method: "POST", URI: "index.php?/api/v2/update_run/{run_id}", Description: "Update test run"},
		{Method: "POST", URI: "index.php?/api/v2/close_run/{run_id}", Description: "Close test run"},
		{Method: "POST", URI: "index.php?/api/v2/delete_run/{run_id}", Description: "Delete test run"},
	}
}

// Sections lists API endpoints for the corresponding TestRail resource.
type Sections struct{}

// Paths returns API endpoints for Sections.
func (r *Sections) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_section/{section_id}", Description: "Get single section"},
		{Method: "GET", URI: "index.php?/api/v2/get_sections/{project_id}&suite_id={suite_id}", Description: "Get sections for project/suite"},
		{Method: "POST", URI: "index.php?/api/v2/add_section/{project_id}", Description: "Create section"},
		{Method: "POST", URI: "index.php?/api/v2/update_section/{section_id}", Description: "Update section"},
		{Method: "POST", URI: "index.php?/api/v2/delete_section/{section_id}", Description: "Delete section"},
	}
}

// Statuses lists API endpoints for the corresponding TestRail resource.
type Statuses struct{}

// Paths returns API endpoints for Statuses.
func (r *Statuses) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_statuses", Description: "Get list of available statuses"},
	}
}

// Suites lists API endpoints for the corresponding TestRail resource.
type Suites struct{}

// Paths returns API endpoints for Suites.
func (r *Suites) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_suite/{suite_id}", Description: "Get single test suite"},
		{Method: "GET", URI: "index.php?/api/v2/get_suites/{project_id}", Description: "Get test suites for project"},
		{Method: "POST", URI: "index.php?/api/v2/add_suite/{project_id}", Description: "Create test suite"},
		{Method: "POST", URI: "index.php?/api/v2/update_suite/{suite_id}", Description: "Update test suite"},
		{Method: "POST", URI: "index.php?/api/v2/delete_suite/{suite_id}", Description: "Delete test suite"},
	}
}

// Tests lists API endpoints for the corresponding TestRail resource.
type Tests struct{}

// Paths returns API endpoints for Tests.
func (r *Tests) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_test/{test_id}", Description: "Get single test"},
		{Method: "GET", URI: "index.php?/api/v2/get_tests/{run_id}", Description: "Get tests for run (supports status_id filter)"},
		{Method: "POST", URI: "index.php?/api/v2/update_test/{test_id}", Description: "Update test (e.g., status)"},
	}
}

// Users lists API endpoints for the corresponding TestRail resource.
type Users struct{}

// Paths returns API endpoints for Users.
func (r *Users) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_user/{user_id}", Description: "Get user by ID"},
		{Method: "GET", URI: "index.php?/api/v2/get_user_by_email&email={email}", Description: "Get user by email"},
		{Method: "GET", URI: "index.php?/api/v2/get_users", Description: "Get all users"},
		{Method: "GET", URI: "index.php?/api/v2/get_users/{project_id}", Description: "Get users for project", Params: map[string]string{"project_id": "Required"}},
		{Method: "POST", URI: "index.php?/api/v2/add_user", Description: "Add new user"},
		{Method: "POST", URI: "index.php?/api/v2/update_user/{user_id}", Description: "Update user", Params: map[string]string{"user_id": "Required"}},
	}
}

// Reports lists API endpoints for the corresponding TestRail resource.
type Reports struct{}

// Paths returns API endpoints for Reports.
func (r *Reports) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_reports/{project_id}", Description: "Get available reports for project"},
		{Method: "GET", URI: "index.php?/api/v2/get_cross_project_reports", Description: "Get cross-project reports"},
		{Method: "GET", URI: "index.php?/api/v2/run_report/{report_template_id}", Description: "Run single-project report"},
		{Method: "GET", URI: "index.php?/api/v2/run_cross_project_report/{report_template_id}", Description: "Run cross-project report"},
	}
}

// Roles lists API endpoints for the corresponding TestRail resource.
type Roles struct{}

// Paths returns API endpoints for Roles.
func (r *Roles) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_roles", Description: "Get list of available roles"},
		{Method: "GET", URI: "index.php?/api/v2/get_role/{role_id}", Description: "Get single role"},
		// Other methods (add/update/delete) may be admin-only and are not always publicly documented
	}
}

// Templates lists API endpoints for the corresponding TestRail resource.
type Templates struct{}

// Paths returns API endpoints for Templates.
func (r *Templates) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_templates/{project_id}", Description: "Get list of available templates for project", Params: map[string]string{"project_id": "Required"}},
	}
}

// Groups lists API endpoints for the corresponding TestRail resource.
type Groups struct{}

// Paths returns API endpoints for Groups.
func (r *Groups) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_group/{group_id}", Description: "Get single group", Params: map[string]string{"group_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_groups/{project_id}", Description: "Get groups for project", Params: map[string]string{"project_id": "Required"}},
		{Method: "POST", URI: "index.php?/api/v2/add_group/{project_id}", Description: "Add new group"},
		{Method: "POST", URI: "index.php?/api/v2/update_group/{group_id}", Description: "Update group"},
		{Method: "POST", URI: "index.php?/api/v2/delete_group/{group_id}", Description: "Delete group"},
	}
}

// SharedSteps lists API endpoints for the corresponding TestRail resource.
type SharedSteps struct{}

// Paths returns API endpoints for SharedSteps.
func (r *SharedSteps) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_shared_step/{shared_step_id}", Description: "Get single shared step", Params: map[string]string{"shared_step_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_shared_step_history/{shared_step_id}", Description: "Get history for shared step"},
		{Method: "GET", URI: "index.php?/api/v2/get_shared_steps/{project_id}", Description: "Get shared steps for project", Params: map[string]string{"project_id": "Required"}},
		{Method: "POST", URI: "index.php?/api/v2/add_shared_step/{project_id}", Description: "Add new shared step"},
		{Method: "POST", URI: "index.php?/api/v2/update_shared_step/{shared_step_id}", Description: "Update shared step"},
		{Method: "POST", URI: "index.php?/api/v2/delete_shared_step/{shared_step_id}", Description: "Delete shared step"},
	}
}

// Variables lists API endpoints for the corresponding TestRail resource.
type Variables struct{}

// Paths returns API endpoints for Variables.
func (r *Variables) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_variables/{dataset_id}", Description: "Get variables for dataset"},
		{Method: "POST", URI: "index.php?/api/v2/add_variable/{dataset_id}", Description: "Add new variable"},
		{Method: "POST", URI: "index.php?/api/v2/update_variable/{variable_id}", Description: "Update variable"},
		{Method: "POST", URI: "index.php?/api/v2/delete_variable/{variable_id}", Description: "Delete variable"},
	}
}

// Labels lists API endpoints for the corresponding TestRail resource.
type Labels struct{}

// Paths returns API endpoints for Labels.
func (r *Labels) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_label/{label_id}", Description: "Get single label", Params: map[string]string{"label_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_labels/{project_id}", Description: "Get labels for project", Params: map[string]string{"project_id": "Required"}},
		{Method: "POST", URI: "index.php?/api/v2/update_label/{label_id}", Description: "Update label", Params: map[string]string{"label_id": "Required"}},
		{Method: "POST", URI: "index.php?/api/v2/update_test_labels/{test_id}", Description: "Update labels for test"},
		{Method: "POST", URI: "index.php?/api/v2/update_tests_labels/{run_id}", Description: "Bulk update labels for tests in run"},
	}
}

// Datasets lists API endpoints for the corresponding TestRail resource.
type Datasets struct{}

// Paths returns API endpoints for Datasets.
func (r *Datasets) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_dataset/{dataset_id}", Description: "Get single dataset", Params: map[string]string{"dataset_id": "Required"}},
		{Method: "GET", URI: "index.php?/api/v2/get_datasets/{project_id}", Description: "Get datasets for project"},
		{Method: "POST", URI: "index.php?/api/v2/add_dataset/{project_id}", Description: "Add new dataset"},
		{Method: "POST", URI: "index.php?/api/v2/update_dataset/{dataset_id}", Description: "Update dataset"},
		{Method: "POST", URI: "index.php?/api/v2/delete_dataset/{dataset_id}", Description: "Delete dataset"},
	}
}

// BDDs lists API endpoints for the corresponding TestRail resource.
type BDDs struct{}

// Paths returns API endpoints for BDDs.
func (r *BDDs) Paths() []APIPath {
	return []APIPath{
		{Method: "GET", URI: "index.php?/api/v2/get_bdd/{case_id}", Description: "Export BDD scenario as .feature file", Params: map[string]string{"case_id": "Required"}},
		{Method: "POST", URI: "index.php?/api/v2/add_bdd/{case_id}", Description: "Import BDD scenario from .feature file"},
	}
}
