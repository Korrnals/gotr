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
    Attachments      *Attachments
    Cases            *Cases
    CaseFields       *CaseFields
    CaseTypes        *CaseTypes
    Configurations   *Configurations
    Milestones       *Milestones
    Plans            *Plans
    Priorities       *Priorities
    Projects         *Projects
    Results          *Results
    ResultFields     *ResultFields
    Runs             *Runs
    Sections         *Sections
    Statuses         *Statuses
    Suites           *Suites
    Tests            *Tests
    Users            *Users
    Reports          *Reports
    Roles            *Roles
    Templates        *Templates
    Groups           *Groups
    SharedSteps      *SharedSteps
    Variables        *Variables
    Labels           *Labels
    Datasets         *Datasets
    BDDs             *BDDs
}

// New initializes a fully populated TestRailAPI.
func New() *TestRailAPI {
    return &TestRailAPI{
        Attachments:      &Attachments{},
        Cases:            &Cases{},
        CaseFields:       &CaseFields{},
        CaseTypes:        &CaseTypes{},
        Configurations:   &Configurations{},
        Milestones:       &Milestones{},
        Plans:            &Plans{},
        Priorities:       &Priorities{},
        Projects:         &Projects{},
        Results:          &Results{},
        ResultFields:     &ResultFields{},
        Runs:             &Runs{},
        Sections:         &Sections{},
        Statuses:         &Statuses{},
        Suites:           &Suites{},
        Tests:            &Tests{},
        Users:            &Users{},
        Reports:          &Reports{},
        Roles:            &Roles{},
        Templates:        &Templates{},
        Groups:           &Groups{},
        SharedSteps:      &SharedSteps{},
        Variables:        &Variables{},
        Labels:           &Labels{},
        Datasets:         &Datasets{},
        BDDs:             &BDDs{},
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

type Attachments struct{}

func (a *Attachments) Paths() []APIPath {
    return []APIPath{
        {Method: "POST", URI: "index.php?/api/v2/add_attachment_to_case/{case_id}", Description: "Add attachment to case"},
        {Method: "POST", URI: "index.php?/api/v2/add_attachment_to_plan/{plan_id}", Description: "Add attachment to plan"},
        {Method: "POST", URI: "index.php?/api/v2/add_attachment_to_plan_entry/{plan_id}/{entry_id}", Description: "Add attachment to plan entry"},
        {Method: "POST", URI: "index.php?/api/v2/add_attachment_to_result/{result_id}", Description: "Add attachment to result"},
        {Method: "POST", URI: "index.php?/api/v2/add_attachment_to_run/{run_id}", Description: "Add attachment to run"},
    }
}

type Cases struct{}

func (c *Cases) Paths() []APIPath {
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

type CaseFields struct{}

func (cf *CaseFields) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_case_fields", Description: "Get list of available case fields"},
        {Method: "POST", URI: "index.php?/api/v2/add_case_field", Description: "Create new custom case field"},
    }
}

type CaseTypes struct{}

func (ct *CaseTypes) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_case_types", Description: "Get list of available case types"},
    }
}

type Configurations struct{}

func (cfg *Configurations) Paths() []APIPath {
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

type Milestones struct{}

func (m *Milestones) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_milestone/{milestone_id}", Description: "Get single milestone"},
        {Method: "GET", URI: "index.php?/api/v2/get_milestones/{project_id}", Description: "Get milestones for project"},
        {Method: "POST", URI: "index.php?/api/v2/add_milestone/{project_id}", Description: "Create milestone"},
        {Method: "POST", URI: "index.php?/api/v2/update_milestone/{milestone_id}", Description: "Update milestone"},
        {Method: "POST", URI: "index.php?/api/v2/delete_milestone/{milestone_id}", Description: "Delete milestone"},
    }
}

type Plans struct{}

func (p *Plans) Paths() []APIPath {
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

type Priorities struct{}

func (pr *Priorities) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_priorities", Description: "Get list of available priorities"},
    }
}

type Projects struct{}

func (p *Projects) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_project/{project_id}", Description: "Get single project"},
        {Method: "GET", URI: "index.php?/api/v2/get_projects", Description: "Get all projects (with filters)"},
        {Method: "POST", URI: "index.php?/api/v2/add_project", Description: "Create new project"},
        {Method: "POST", URI: "index.php?/api/v2/update_project/{project_id}", Description: "Update project"},
        {Method: "POST", URI: "index.php?/api/v2/delete_project/{project_id}", Description: "Delete (archive) project"},
    }
}

type Results struct{}

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

type ResultFields struct{}

func (rf *ResultFields) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_result_fields", Description: "Get list of available result fields"},
    }
}

type Runs struct{}

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

type Sections struct{}

func (s *Sections) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_section/{section_id}", Description: "Get single section"},
        {Method: "GET", URI: "index.php?/api/v2/get_sections/{project_id}&suite_id={suite_id}", Description: "Get sections for project/suite"},
        {Method: "POST", URI: "index.php?/api/v2/add_section/{project_id}", Description: "Create section"},
        {Method: "POST", URI: "index.php?/api/v2/update_section/{section_id}", Description: "Update section"},
        {Method: "POST", URI: "index.php?/api/v2/delete_section/{section_id}", Description: "Delete section"},
    }
}

type Statuses struct{}

func (s *Statuses) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_statuses", Description: "Get list of available statuses"},
    }
}

type Suites struct{}

func (s *Suites) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_suite/{suite_id}", Description: "Get single test suite"},
        {Method: "GET", URI: "index.php?/api/v2/get_suites/{project_id}", Description: "Get test suites for project"},
        {Method: "POST", URI: "index.php?/api/v2/add_suite/{project_id}", Description: "Create test suite"},
        {Method: "POST", URI: "index.php?/api/v2/update_suite/{suite_id}", Description: "Update test suite"},
        {Method: "POST", URI: "index.php?/api/v2/delete_suite/{suite_id}", Description: "Delete test suite"},
    }
}

type Tests struct{}

func (t *Tests) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_test/{test_id}", Description: "Get single test"},
        {Method: "GET", URI: "index.php?/api/v2/get_tests/{run_id}", Description: "Get tests for run (supports status_id filter)"},
        {Method: "POST", URI: "index.php?/api/v2/update_test/{test_id}", Description: "Update test (e.g., status)"},
    }
}

type Users struct{}

func (u *Users) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_user/{user_id}", Description: "Get user by ID"},
        {Method: "GET", URI: "index.php?/api/v2/get_user_by_email&email={email}", Description: "Get user by email"},
        {Method: "GET", URI: "index.php?/api/v2/get_users", Description: "Get all users"},
    }
}

type Reports struct{}

func (r *Reports) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_reports/{project_id}", Description: "Get available reports for project"},
        {Method: "GET", URI: "index.php?/api/v2/run_report/{report_template_id}", Description: "Run single-project report"},
        {Method: "GET", URI: "index.php?/api/v2/run_cross_project_report/{report_template_id}", Description: "Run cross-project report"},
    }
}

type Roles struct{}

func (r *Roles) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_roles", Description: "Get list of available roles"},
        {Method: "GET", URI: "index.php?/api/v2/get_role/{role_id}", Description: "Get single role"},
        // Другие методы (add/update/delete) могут быть только для админов и не всегда документированы публично
    }
}

type Templates struct{}

func (t *Templates) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_templates/{project_id}", Description: "Get list of available templates for project", Params: map[string]string{"project_id": "Required"}},
    }
}

type Groups struct{}

func (g *Groups) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_group/{group_id}", Description: "Get single group", Params: map[string]string{"group_id": "Required"}},
        {Method: "GET", URI: "index.php?/api/v2/get_groups/{project_id}", Description: "Get groups for project", Params: map[string]string{"project_id": "Required"}},
        {Method: "POST", URI: "index.php?/api/v2/add_group/{project_id}", Description: "Add new group"},
        {Method: "POST", URI: "index.php?/api/v2/update_group/{group_id}", Description: "Update group"},
        {Method: "POST", URI: "index.php?/api/v2/delete_group/{group_id}", Description: "Delete group"},
    }
}

type SharedSteps struct{}

func (s *SharedSteps) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_shared_step/{shared_step_id}", Description: "Get single shared step", Params: map[string]string{"shared_step_id": "Required"}},
        {Method: "GET", URI: "index.php?/api/v2/get_shared_step_history/{shared_step_id}", Description: "Get history for shared step"},
        {Method: "GET", URI: "index.php?/api/v2/get_shared_steps/{project_id}", Description: "Get shared steps for project", Params: map[string]string{"project_id": "Required"}},
        {Method: "POST", URI: "index.php?/api/v2/add_shared_step/{project_id}", Description: "Add new shared step"},
        {Method: "POST", URI: "index.php?/api/v2/update_shared_step/{shared_step_id}", Description: "Update shared step"},
        {Method: "POST", URI: "index.php?/api/v2/delete_shared_step/{shared_step_id}", Description: "Delete shared step"},
    }
}

type Variables struct{}

func (v *Variables) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_variables/{dataset_id}", Description: "Get variables for dataset"},
        {Method: "POST", URI: "index.php?/api/v2/add_variable/{dataset_id}", Description: "Add new variable"},
        {Method: "POST", URI: "index.php?/api/v2/update_variable/{variable_id}", Description: "Update variable"},
        {Method: "POST", URI: "index.php?/api/v2/delete_variable/{variable_id}", Description: "Delete variable"},
    }
}

type Labels struct{}

func (l *Labels) Paths() []APIPath {
    return []APIPath{
        // Labels часто обновляются через тесты/результаты, но есть bulk-методы
        {Method: "POST", URI: "index.php?/api/v2/update_test_labels/{test_id}", Description: "Update labels for test"},
        {Method: "POST", URI: "index.php?/api/v2/update_tests_labels/{run_id}", Description: "Bulk update labels for tests in run"},
        // Если появятся dedicated endpoints — добавим
    }
}

type Datasets struct{}

func (d *Datasets) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_dataset/{dataset_id}", Description: "Get single dataset", Params: map[string]string{"dataset_id": "Required"}},
        {Method: "GET", URI: "index.php?/api/v2/get_datasets/{project_id}", Description: "Get datasets for project"},
        {Method: "POST", URI: "index.php?/api/v2/add_dataset/{project_id}", Description: "Add new dataset"},
        {Method: "POST", URI: "index.php?/api/v2/update_dataset/{dataset_id}", Description: "Update dataset"},
        {Method: "POST", URI: "index.php?/api/v2/delete_dataset/{dataset_id}", Description: "Delete dataset"},
    }
}

type BDDs struct{}

func (b *BDDs) Paths() []APIPath {
    return []APIPath{
        {Method: "GET", URI: "index.php?/api/v2/get_bdd/{case_id}", Description: "Export BDD scenario as .feature file", Params: map[string]string{"case_id": "Required"}},
        {Method: "POST", URI: "index.php?/api/v2/add_bdd/{case_id}", Description: "Import BDD scenario from .feature file"},
    }
}
