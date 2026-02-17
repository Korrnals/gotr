// internal/client/mock.go
// Универсальный mock-клиент для тестирования
package client

import (
	"github.com/Korrnals/gotr/internal/models/data"
)

// MockClient реализует ClientInterface для тестирования
type MockClient struct {
	// ProjectsAPI
	GetProjectsFunc   func() (data.GetProjectsResponse, error)
	GetProjectFunc    func(projectID int64) (*data.GetProjectResponse, error)
	AddProjectFunc    func(req *data.AddProjectRequest) (*data.GetProjectResponse, error)
	UpdateProjectFunc func(projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error)
	DeleteProjectFunc func(projectID int64) error

	// CasesAPI
	GetCasesFunc           func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error)
	GetCaseFunc            func(caseID int64) (*data.Case, error)
	AddCaseFunc            func(sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	UpdateCaseFunc         func(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	DeleteCaseFunc         func(caseID int64) error
	UpdateCasesFunc        func(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error)
	DeleteCasesFunc        func(suiteID int64, req *data.DeleteCasesRequest) error
	CopyCasesToSectionFunc func(sectionID int64, req *data.CopyCasesRequest) error
	MoveCasesToSectionFunc func(sectionID int64, req *data.MoveCasesRequest) error
	GetHistoryForCaseFunc  func(caseID int64) (*data.GetHistoryForCaseResponse, error)
	GetCaseFieldsFunc      func() (data.GetCaseFieldsResponse, error)
	AddCaseFieldFunc       func(req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error)
	GetCaseTypesFunc       func() (data.GetCaseTypesResponse, error)
	DiffCasesDataFunc      func(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error)

	// ConcurrentAPI - Parallel methods (Stage 6.2)
	GetCasesParallelFunc          func(projectID int64, suiteIDs []int64, workers int) (map[int64]data.GetCasesResponse, error)
	GetCasesForSuitesParallelFunc func(projectID int64, suiteIDs []int64, workers int) (data.GetCasesResponse, error)
	GetSuitesParallelFunc         func(projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error)

	// SuitesAPI
	GetSuitesFunc   func(projectID int64) (data.GetSuitesResponse, error)
	GetSuiteFunc    func(suiteID int64) (*data.Suite, error)
	AddSuiteFunc    func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	UpdateSuiteFunc func(suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error)
	DeleteSuiteFunc func(suiteID int64) error

	// ConcurrentAPI - Parallel methods (Stage 6.2)

	// SectionsAPI
	GetSectionsFunc   func(projectID, suiteID int64) (data.GetSectionsResponse, error)
	GetSectionFunc    func(sectionID int64) (*data.Section, error)
	AddSectionFunc    func(projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	UpdateSectionFunc func(sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error)
	DeleteSectionFunc func(sectionID int64) error

	// SharedStepsAPI
	GetSharedStepsFunc       func(projectID int64) (data.GetSharedStepsResponse, error)
	GetSharedStepFunc        func(stepID int64) (*data.SharedStep, error)
	AddSharedStepFunc        func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	UpdateSharedStepFunc     func(stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error)
	DeleteSharedStepFunc     func(stepID int64, keepInCases int) error
	GetSharedStepHistoryFunc func(stepID int64) (*data.GetSharedStepHistoryResponse, error)

	// RunsAPI
	GetRunsFunc   func(projectID int64) (data.GetRunsResponse, error)
	GetRunFunc    func(runID int64) (*data.Run, error)
	AddRunFunc    func(projectID int64, req *data.AddRunRequest) (*data.Run, error)
	UpdateRunFunc func(runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	CloseRunFunc  func(runID int64) (*data.Run, error)
	DeleteRunFunc func(runID int64) error

	// ResultsAPI
	GetResultsFunc         func(testID int64) (data.GetResultsResponse, error)
	GetResultsForRunFunc   func(runID int64) (data.GetResultsResponse, error)
	GetResultsForCaseFunc  func(runID, caseID int64) (data.GetResultsResponse, error)
	AddResultFunc          func(testID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResultForCaseFunc   func(runID, caseID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResultsFunc         func(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error)
	AddResultsForCasesFunc func(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error)

	// TestsAPI
	GetTestFunc    func(testID int64) (*data.Test, error)
	GetTestsFunc   func(runID int64, filters map[string]string) ([]data.Test, error)
	UpdateTestFunc func(testID int64, req *data.UpdateTestRequest) (*data.Test, error)

	// MilestonesAPI
	GetMilestoneFunc    func(milestoneID int64) (*data.Milestone, error)
	GetMilestonesFunc   func(projectID int64) ([]data.Milestone, error)
	AddMilestoneFunc    func(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error)
	UpdateMilestoneFunc func(milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error)
	DeleteMilestoneFunc func(milestoneID int64) error

	// PlansAPI
	GetPlanFunc         func(planID int64) (*data.Plan, error)
	GetPlansFunc        func(projectID int64) (data.GetPlansResponse, error)
	AddPlanFunc         func(projectID int64, req *data.AddPlanRequest) (*data.Plan, error)
	UpdatePlanFunc      func(planID int64, req *data.UpdatePlanRequest) (*data.Plan, error)
	ClosePlanFunc       func(planID int64) (*data.Plan, error)
	DeletePlanFunc      func(planID int64) error
	AddPlanEntryFunc    func(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error)
	UpdatePlanEntryFunc func(planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error)
	DeletePlanEntryFunc func(planID int64, entryID string) error

	// AttachmentsAPI
	AddAttachmentToCaseFunc       func(caseID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanFunc       func(planID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanEntryFunc  func(planID int64, entryID string, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToResultFunc     func(resultID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToRunFunc        func(runID int64, filePath string) (*data.AttachmentResponse, error)
	DeleteAttachmentFunc          func(attachmentID int64) error
	GetAttachmentFunc             func(attachmentID int64) (*data.Attachment, error)
	GetAttachmentsForCaseFunc     func(caseID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForPlanFunc     func(planID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForPlanEntryFunc func(planID int64, entryID string) (data.GetAttachmentsResponse, error)
	GetAttachmentsForRunFunc      func(runID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForTestFunc     func(testID int64) (data.GetAttachmentsResponse, error)

	// ConfigurationsAPI
	GetConfigsFunc          func(projectID int64) (data.GetConfigsResponse, error)
	AddConfigGroupFunc      func(projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error)
	AddConfigFunc           func(groupID int64, req *data.AddConfigRequest) (*data.Config, error)
	UpdateConfigGroupFunc   func(groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error)
	UpdateConfigFunc        func(configID int64, req *data.UpdateConfigRequest) (*data.Config, error)
	DeleteConfigGroupFunc   func(groupID int64) error
	DeleteConfigFunc        func(configID int64) error

	// UsersAPI
	GetUsersFunc         func() (data.GetUsersResponse, error)
	GetUsersByProjectFunc func(projectID int64) (data.GetUsersResponse, error)
	GetUserFunc          func(userID int64) (*data.User, error)
	GetUserByEmailFunc   func(email string) (*data.User, error)
	AddUserFunc          func(req data.AddUserRequest) (*data.User, error)
	UpdateUserFunc       func(userID int64, req data.UpdateUserRequest) (*data.User, error)
	GetPrioritiesFunc    func() (data.GetPrioritiesResponse, error)
	GetStatusesFunc      func() (data.GetStatusesResponse, error)
	GetTemplatesFunc     func(projectID int64) (data.GetTemplatesResponse, error)

	// ReportsAPI
	GetReportsFunc              func(projectID int64) (data.GetReportsResponse, error)
	GetCrossProjectReportsFunc  func() (data.GetReportsResponse, error)
	RunReportFunc               func(templateID int64) (*data.RunReportResponse, error)
	RunCrossProjectReportFunc   func(templateID int64) (*data.RunReportResponse, error)

	// ExtendedAPI - Groups
	GetGroupsFunc       func(projectID int64) (data.GetGroupsResponse, error)
	GetGroupFunc        func(groupID int64) (*data.Group, error)
	AddGroupFunc        func(projectID int64, name string, userIDs []int64) (*data.Group, error)
	UpdateGroupFunc     func(groupID int64, name string, userIDs []int64) (*data.Group, error)
	DeleteGroupFunc     func(groupID int64) error

	// ExtendedAPI - Roles
	GetRolesFunc        func() (data.GetRolesResponse, error)
	GetRoleFunc         func(roleID int64) (*data.Role, error)

	// ExtendedAPI - ResultFields
	GetResultFieldsFunc func() (data.GetResultFieldsResponse, error)

	// ExtendedAPI - Datasets
	GetDatasetsFunc     func(projectID int64) (data.GetDatasetsResponse, error)
	GetDatasetFunc      func(datasetID int64) (*data.Dataset, error)
	AddDatasetFunc      func(projectID int64, name string) (*data.Dataset, error)
	UpdateDatasetFunc   func(datasetID int64, name string) (*data.Dataset, error)
	DeleteDatasetFunc   func(datasetID int64) error

	// ExtendedAPI - Variables
	GetVariablesFunc    func(datasetID int64) (data.GetVariablesResponse, error)
	AddVariableFunc     func(datasetID int64, name string) (*data.Variable, error)
	UpdateVariableFunc  func(variableID int64, name string) (*data.Variable, error)
	DeleteVariableFunc  func(variableID int64) error

	// ExtendedAPI - BDDs
	GetBDDFunc          func(caseID int64) (*data.BDD, error)
	AddBDDFunc          func(caseID int64, content string) (*data.BDD, error)

	// ExtendedAPI - Labels
	GetLabelsFunc         func(projectID int64) (data.GetLabelsResponse, error)
	GetLabelFunc          func(labelID int64) (*data.Label, error)
	UpdateLabelFunc       func(labelID int64, req data.UpdateLabelRequest) (*data.Label, error)
	UpdateTestLabelsFunc  func(testID int64, labels []string) error
	UpdateTestsLabelsFunc func(runID int64, testIDs []int64, labels []string) error
}

// Проверка, что MockClient реализует ClientInterface
var _ ClientInterface = (*MockClient)(nil)

// ---------------------------------------------------------------------------
// ProjectsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetProjects() (data.GetProjectsResponse, error) {
	if m.GetProjectsFunc != nil {
		return m.GetProjectsFunc()
	}
	return nil, nil
}

func (m *MockClient) GetProject(projectID int64) (*data.GetProjectResponse, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) AddProject(req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
	if m.AddProjectFunc != nil {
		return m.AddProjectFunc(req)
	}
	return nil, nil
}

func (m *MockClient) UpdateProject(projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
	if m.UpdateProjectFunc != nil {
		return m.UpdateProjectFunc(projectID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteProject(projectID int64) error {
	if m.DeleteProjectFunc != nil {
		return m.DeleteProjectFunc(projectID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// CasesAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetCases(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
	if m.GetCasesFunc != nil {
		return m.GetCasesFunc(projectID, suiteID, sectionID)
	}
	return nil, nil
}

func (m *MockClient) GetCase(caseID int64) (*data.Case, error) {
	if m.GetCaseFunc != nil {
		return m.GetCaseFunc(caseID)
	}
	return nil, nil
}

func (m *MockClient) AddCase(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
	if m.AddCaseFunc != nil {
		return m.AddCaseFunc(sectionID, req)
	}
	return &data.Case{ID: 999}, nil
}

func (m *MockClient) UpdateCase(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
	if m.UpdateCaseFunc != nil {
		return m.UpdateCaseFunc(caseID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteCase(caseID int64) error {
	if m.DeleteCaseFunc != nil {
		return m.DeleteCaseFunc(caseID)
	}
	return nil
}

func (m *MockClient) UpdateCases(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
	if m.UpdateCasesFunc != nil {
		return m.UpdateCasesFunc(suiteID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteCases(suiteID int64, req *data.DeleteCasesRequest) error {
	if m.DeleteCasesFunc != nil {
		return m.DeleteCasesFunc(suiteID, req)
	}
	return nil
}

func (m *MockClient) CopyCasesToSection(sectionID int64, req *data.CopyCasesRequest) error {
	if m.CopyCasesToSectionFunc != nil {
		return m.CopyCasesToSectionFunc(sectionID, req)
	}
	return nil
}

func (m *MockClient) MoveCasesToSection(sectionID int64, req *data.MoveCasesRequest) error {
	if m.MoveCasesToSectionFunc != nil {
		return m.MoveCasesToSectionFunc(sectionID, req)
	}
	return nil
}

func (m *MockClient) GetHistoryForCase(caseID int64) (*data.GetHistoryForCaseResponse, error) {
	if m.GetHistoryForCaseFunc != nil {
		return m.GetHistoryForCaseFunc(caseID)
	}
	return nil, nil
}

func (m *MockClient) GetCaseFields() (data.GetCaseFieldsResponse, error) {
	if m.GetCaseFieldsFunc != nil {
		return m.GetCaseFieldsFunc()
	}
	return nil, nil
}

func (m *MockClient) AddCaseField(req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error) {
	if m.AddCaseFieldFunc != nil {
		return m.AddCaseFieldFunc(req)
	}
	return nil, nil
}

func (m *MockClient) GetCaseTypes() (data.GetCaseTypesResponse, error) {
	if m.GetCaseTypesFunc != nil {
		return m.GetCaseTypesFunc()
	}
	return nil, nil
}

func (m *MockClient) DiffCasesData(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
	if m.DiffCasesDataFunc != nil {
		return m.DiffCasesDataFunc(pid1, pid2, field)
	}
	return nil, nil
}

// GetCasesParallel — mock implementation
func (m *MockClient) GetCasesParallel(projectID int64, suiteIDs []int64, workers int) (map[int64]data.GetCasesResponse, error) {
	if m.GetCasesParallelFunc != nil {
		return m.GetCasesParallelFunc(projectID, suiteIDs, workers)
	}
	// Default: sequential fallback
	results := make(map[int64]data.GetCasesResponse)
	for _, sid := range suiteIDs {
		cases, err := m.GetCases(projectID, sid, 0)
		if err != nil {
			return results, err
		}
		results[sid] = cases
	}
	return results, nil
}

// GetCasesForSuitesParallel — mock implementation
func (m *MockClient) GetCasesForSuitesParallel(projectID int64, suiteIDs []int64, workers int) (data.GetCasesResponse, error) {
	if m.GetCasesForSuitesParallelFunc != nil {
		return m.GetCasesForSuitesParallelFunc(projectID, suiteIDs, workers)
	}
	// Default: use GetCasesParallel and flatten
	results, err := m.GetCasesParallel(projectID, suiteIDs, workers)
	if err != nil {
		return nil, err
	}
	var allCases data.GetCasesResponse
	for _, cases := range results {
		allCases = append(allCases, cases...)
	}
	return allCases, nil
}

// ---------------------------------------------------------------------------
// SuitesAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetSuites(projectID int64) (data.GetSuitesResponse, error) {
	if m.GetSuitesFunc != nil {
		return m.GetSuitesFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetSuite(suiteID int64) (*data.Suite, error) {
	if m.GetSuiteFunc != nil {
		return m.GetSuiteFunc(suiteID)
	}
	return nil, nil
}

func (m *MockClient) AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	if m.AddSuiteFunc != nil {
		return m.AddSuiteFunc(projectID, req)
	}
	return &data.Suite{ID: 999}, nil
}

func (m *MockClient) UpdateSuite(suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
	if m.UpdateSuiteFunc != nil {
		return m.UpdateSuiteFunc(suiteID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteSuite(suiteID int64) error {
	if m.DeleteSuiteFunc != nil {
		return m.DeleteSuiteFunc(suiteID)
	}
	return nil
}

// GetSuitesParallel — mock implementation
func (m *MockClient) GetSuitesParallel(projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error) {
	if m.GetSuitesParallelFunc != nil {
		return m.GetSuitesParallelFunc(projectIDs, workers)
	}
	// Default: sequential fallback
	results := make(map[int64]data.GetSuitesResponse)
	for _, pid := range projectIDs {
		suites, err := m.GetSuites(pid)
		if err != nil {
			return results, err
		}
		results[pid] = suites
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// SectionsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error) {
	if m.GetSectionsFunc != nil {
		return m.GetSectionsFunc(projectID, suiteID)
	}
	return nil, nil
}

func (m *MockClient) GetSection(sectionID int64) (*data.Section, error) {
	if m.GetSectionFunc != nil {
		return m.GetSectionFunc(sectionID)
	}
	return nil, nil
}

func (m *MockClient) AddSection(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	if m.AddSectionFunc != nil {
		return m.AddSectionFunc(projectID, req)
	}
	return &data.Section{ID: 999}, nil
}

func (m *MockClient) UpdateSection(sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
	if m.UpdateSectionFunc != nil {
		return m.UpdateSectionFunc(sectionID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteSection(sectionID int64) error {
	if m.DeleteSectionFunc != nil {
		return m.DeleteSectionFunc(sectionID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// SharedStepsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error) {
	if m.GetSharedStepsFunc != nil {
		return m.GetSharedStepsFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetSharedStep(stepID int64) (*data.SharedStep, error) {
	if m.GetSharedStepFunc != nil {
		return m.GetSharedStepFunc(stepID)
	}
	return nil, nil
}

func (m *MockClient) AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
	if m.AddSharedStepFunc != nil {
		return m.AddSharedStepFunc(projectID, req)
	}
	return &data.SharedStep{ID: 999}, nil
}

func (m *MockClient) UpdateSharedStep(stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
	if m.UpdateSharedStepFunc != nil {
		return m.UpdateSharedStepFunc(stepID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteSharedStep(stepID int64, keepInCases int) error {
	if m.DeleteSharedStepFunc != nil {
		return m.DeleteSharedStepFunc(stepID, keepInCases)
	}
	return nil
}

func (m *MockClient) GetSharedStepHistory(stepID int64) (*data.GetSharedStepHistoryResponse, error) {
	if m.GetSharedStepHistoryFunc != nil {
		return m.GetSharedStepHistoryFunc(stepID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// RunsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetRuns(projectID int64) (data.GetRunsResponse, error) {
	if m.GetRunsFunc != nil {
		return m.GetRunsFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetRun(runID int64) (*data.Run, error) {
	if m.GetRunFunc != nil {
		return m.GetRunFunc(runID)
	}
	return nil, nil
}

func (m *MockClient) AddRun(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	if m.AddRunFunc != nil {
		return m.AddRunFunc(projectID, req)
	}
	return nil, nil
}

func (m *MockClient) UpdateRun(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	if m.UpdateRunFunc != nil {
		return m.UpdateRunFunc(runID, req)
	}
	return nil, nil
}

func (m *MockClient) CloseRun(runID int64) (*data.Run, error) {
	if m.CloseRunFunc != nil {
		return m.CloseRunFunc(runID)
	}
	return nil, nil
}

func (m *MockClient) DeleteRun(runID int64) error {
	if m.DeleteRunFunc != nil {
		return m.DeleteRunFunc(runID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ResultsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetResults(testID int64) (data.GetResultsResponse, error) {
	if m.GetResultsFunc != nil {
		return m.GetResultsFunc(testID)
	}
	return nil, nil
}

func (m *MockClient) GetResultsForRun(runID int64) (data.GetResultsResponse, error) {
	if m.GetResultsForRunFunc != nil {
		return m.GetResultsForRunFunc(runID)
	}
	return nil, nil
}

func (m *MockClient) GetResultsForCase(runID, caseID int64) (data.GetResultsResponse, error) {
	if m.GetResultsForCaseFunc != nil {
		return m.GetResultsForCaseFunc(runID, caseID)
	}
	return nil, nil
}

func (m *MockClient) AddResult(testID int64, req *data.AddResultRequest) (*data.Result, error) {
	if m.AddResultFunc != nil {
		return m.AddResultFunc(testID, req)
	}
	return nil, nil
}

func (m *MockClient) AddResultForCase(runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	if m.AddResultForCaseFunc != nil {
		return m.AddResultForCaseFunc(runID, caseID, req)
	}
	return nil, nil
}

func (m *MockClient) AddResults(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	if m.AddResultsFunc != nil {
		return m.AddResultsFunc(runID, req)
	}
	return nil, nil
}

func (m *MockClient) AddResultsForCases(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	if m.AddResultsForCasesFunc != nil {
		return m.AddResultsForCasesFunc(runID, req)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// TestsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetTest(testID int64) (*data.Test, error) {
	if m.GetTestFunc != nil {
		return m.GetTestFunc(testID)
	}
	return nil, nil
}

func (m *MockClient) GetTests(runID int64, filters map[string]string) ([]data.Test, error) {
	if m.GetTestsFunc != nil {
		return m.GetTestsFunc(runID, filters)
	}
	return nil, nil
}

func (m *MockClient) UpdateTest(testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
	if m.UpdateTestFunc != nil {
		return m.UpdateTestFunc(testID, req)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// MilestonesAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetMilestone(milestoneID int64) (*data.Milestone, error) {
	if m.GetMilestoneFunc != nil {
		return m.GetMilestoneFunc(milestoneID)
	}
	return nil, nil
}

func (m *MockClient) GetMilestones(projectID int64) ([]data.Milestone, error) {
	if m.GetMilestonesFunc != nil {
		return m.GetMilestonesFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) AddMilestone(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
	if m.AddMilestoneFunc != nil {
		return m.AddMilestoneFunc(projectID, req)
	}
	return nil, nil
}

func (m *MockClient) UpdateMilestone(milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
	if m.UpdateMilestoneFunc != nil {
		return m.UpdateMilestoneFunc(milestoneID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteMilestone(milestoneID int64) error {
	if m.DeleteMilestoneFunc != nil {
		return m.DeleteMilestoneFunc(milestoneID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// PlansAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetPlan(planID int64) (*data.Plan, error) {
	if m.GetPlanFunc != nil {
		return m.GetPlanFunc(planID)
	}
	return nil, nil
}

func (m *MockClient) GetPlans(projectID int64) (data.GetPlansResponse, error) {
	if m.GetPlansFunc != nil {
		return m.GetPlansFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) AddPlan(projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
	if m.AddPlanFunc != nil {
		return m.AddPlanFunc(projectID, req)
	}
	return nil, nil
}

func (m *MockClient) UpdatePlan(planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
	if m.UpdatePlanFunc != nil {
		return m.UpdatePlanFunc(planID, req)
	}
	return nil, nil
}

func (m *MockClient) ClosePlan(planID int64) (*data.Plan, error) {
	if m.ClosePlanFunc != nil {
		return m.ClosePlanFunc(planID)
	}
	return nil, nil
}

func (m *MockClient) DeletePlan(planID int64) error {
	if m.DeletePlanFunc != nil {
		return m.DeletePlanFunc(planID)
	}
	return nil
}

func (m *MockClient) AddPlanEntry(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
	if m.AddPlanEntryFunc != nil {
		return m.AddPlanEntryFunc(planID, req)
	}
	return nil, nil
}

func (m *MockClient) UpdatePlanEntry(planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
	if m.UpdatePlanEntryFunc != nil {
		return m.UpdatePlanEntryFunc(planID, entryID, req)
	}
	return nil, nil
}

func (m *MockClient) DeletePlanEntry(planID int64, entryID string) error {
	if m.DeletePlanEntryFunc != nil {
		return m.DeletePlanEntryFunc(planID, entryID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// AttachmentsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) AddAttachmentToCase(caseID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToCaseFunc != nil {
		return m.AddAttachmentToCaseFunc(caseID, filePath)
	}
	return nil, nil
}

func (m *MockClient) AddAttachmentToPlan(planID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToPlanFunc != nil {
		return m.AddAttachmentToPlanFunc(planID, filePath)
	}
	return nil, nil
}

func (m *MockClient) AddAttachmentToPlanEntry(planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToPlanEntryFunc != nil {
		return m.AddAttachmentToPlanEntryFunc(planID, entryID, filePath)
	}
	return nil, nil
}

func (m *MockClient) AddAttachmentToResult(resultID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToResultFunc != nil {
		return m.AddAttachmentToResultFunc(resultID, filePath)
	}
	return nil, nil
}

func (m *MockClient) AddAttachmentToRun(runID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToRunFunc != nil {
		return m.AddAttachmentToRunFunc(runID, filePath)
	}
	return nil, nil
}

func (m *MockClient) DeleteAttachment(attachmentID int64) error {
	if m.DeleteAttachmentFunc != nil {
		return m.DeleteAttachmentFunc(attachmentID)
	}
	return nil
}

func (m *MockClient) GetAttachment(attachmentID int64) (*data.Attachment, error) {
	if m.GetAttachmentFunc != nil {
		return m.GetAttachmentFunc(attachmentID)
	}
	return nil, nil
}

func (m *MockClient) GetAttachmentsForCase(caseID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForCaseFunc != nil {
		return m.GetAttachmentsForCaseFunc(caseID)
	}
	return nil, nil
}

func (m *MockClient) GetAttachmentsForPlan(planID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForPlanFunc != nil {
		return m.GetAttachmentsForPlanFunc(planID)
	}
	return nil, nil
}

func (m *MockClient) GetAttachmentsForPlanEntry(planID int64, entryID string) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForPlanEntryFunc != nil {
		return m.GetAttachmentsForPlanEntryFunc(planID, entryID)
	}
	return nil, nil
}

func (m *MockClient) GetAttachmentsForRun(runID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForRunFunc != nil {
		return m.GetAttachmentsForRunFunc(runID)
	}
	return nil, nil
}

func (m *MockClient) GetAttachmentsForTest(testID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForTestFunc != nil {
		return m.GetAttachmentsForTestFunc(testID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ConfigurationsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetConfigs(projectID int64) (data.GetConfigsResponse, error) {
	if m.GetConfigsFunc != nil {
		return m.GetConfigsFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) AddConfigGroup(projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
	if m.AddConfigGroupFunc != nil {
		return m.AddConfigGroupFunc(projectID, req)
	}
	return nil, nil
}

func (m *MockClient) AddConfig(groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
	if m.AddConfigFunc != nil {
		return m.AddConfigFunc(groupID, req)
	}
	return nil, nil
}

func (m *MockClient) UpdateConfigGroup(groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
	if m.UpdateConfigGroupFunc != nil {
		return m.UpdateConfigGroupFunc(groupID, req)
	}
	return nil, nil
}

func (m *MockClient) UpdateConfig(configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
	if m.UpdateConfigFunc != nil {
		return m.UpdateConfigFunc(configID, req)
	}
	return nil, nil
}

func (m *MockClient) DeleteConfigGroup(groupID int64) error {
	if m.DeleteConfigGroupFunc != nil {
		return m.DeleteConfigGroupFunc(groupID)
	}
	return nil
}

func (m *MockClient) DeleteConfig(configID int64) error {
	if m.DeleteConfigFunc != nil {
		return m.DeleteConfigFunc(configID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// UsersAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetUsers() (data.GetUsersResponse, error) {
	if m.GetUsersFunc != nil {
		return m.GetUsersFunc()
	}
	return nil, nil
}

func (m *MockClient) GetUsersByProject(projectID int64) (data.GetUsersResponse, error) {
	if m.GetUsersByProjectFunc != nil {
		return m.GetUsersByProjectFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetUser(userID int64) (*data.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(userID)
	}
	return nil, nil
}

func (m *MockClient) GetUserByEmail(email string) (*data.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, nil
}

func (m *MockClient) AddUser(req data.AddUserRequest) (*data.User, error) {
	if m.AddUserFunc != nil {
		return m.AddUserFunc(req)
	}
	return &data.User{ID: 999}, nil
}

func (m *MockClient) UpdateUser(userID int64, req data.UpdateUserRequest) (*data.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(userID, req)
	}
	return nil, nil
}

func (m *MockClient) GetPriorities() (data.GetPrioritiesResponse, error) {
	if m.GetPrioritiesFunc != nil {
		return m.GetPrioritiesFunc()
	}
	return nil, nil
}

func (m *MockClient) GetStatuses() (data.GetStatusesResponse, error) {
	if m.GetStatusesFunc != nil {
		return m.GetStatusesFunc()
	}
	return nil, nil
}

func (m *MockClient) GetTemplates(projectID int64) (data.GetTemplatesResponse, error) {
	if m.GetTemplatesFunc != nil {
		return m.GetTemplatesFunc(projectID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ReportsAPI
// ---------------------------------------------------------------------------
func (m *MockClient) GetReports(projectID int64) (data.GetReportsResponse, error) {
	if m.GetReportsFunc != nil {
		return m.GetReportsFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetCrossProjectReports() (data.GetReportsResponse, error) {
	if m.GetCrossProjectReportsFunc != nil {
		return m.GetCrossProjectReportsFunc()
	}
	return nil, nil
}

func (m *MockClient) RunReport(templateID int64) (*data.RunReportResponse, error) {
	if m.RunReportFunc != nil {
		return m.RunReportFunc(templateID)
	}
	return nil, nil
}

func (m *MockClient) RunCrossProjectReport(templateID int64) (*data.RunReportResponse, error) {
	if m.RunCrossProjectReportFunc != nil {
		return m.RunCrossProjectReportFunc(templateID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Groups
// ---------------------------------------------------------------------------
func (m *MockClient) GetGroups(projectID int64) (data.GetGroupsResponse, error) {
	if m.GetGroupsFunc != nil {
		return m.GetGroupsFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetGroup(groupID int64) (*data.Group, error) {
	if m.GetGroupFunc != nil {
		return m.GetGroupFunc(groupID)
	}
	return nil, nil
}

func (m *MockClient) AddGroup(projectID int64, name string, userIDs []int64) (*data.Group, error) {
	if m.AddGroupFunc != nil {
		return m.AddGroupFunc(projectID, name, userIDs)
	}
	return nil, nil
}

func (m *MockClient) UpdateGroup(groupID int64, name string, userIDs []int64) (*data.Group, error) {
	if m.UpdateGroupFunc != nil {
		return m.UpdateGroupFunc(groupID, name, userIDs)
	}
	return nil, nil
}

func (m *MockClient) DeleteGroup(groupID int64) error {
	if m.DeleteGroupFunc != nil {
		return m.DeleteGroupFunc(groupID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Roles
// ---------------------------------------------------------------------------
func (m *MockClient) GetRoles() (data.GetRolesResponse, error) {
	if m.GetRolesFunc != nil {
		return m.GetRolesFunc()
	}
	return nil, nil
}

func (m *MockClient) GetRole(roleID int64) (*data.Role, error) {
	if m.GetRoleFunc != nil {
		return m.GetRoleFunc(roleID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - ResultFields
// ---------------------------------------------------------------------------
func (m *MockClient) GetResultFields() (data.GetResultFieldsResponse, error) {
	if m.GetResultFieldsFunc != nil {
		return m.GetResultFieldsFunc()
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Datasets
// ---------------------------------------------------------------------------
func (m *MockClient) GetDatasets(projectID int64) (data.GetDatasetsResponse, error) {
	if m.GetDatasetsFunc != nil {
		return m.GetDatasetsFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetDataset(datasetID int64) (*data.Dataset, error) {
	if m.GetDatasetFunc != nil {
		return m.GetDatasetFunc(datasetID)
	}
	return nil, nil
}

func (m *MockClient) AddDataset(projectID int64, name string) (*data.Dataset, error) {
	if m.AddDatasetFunc != nil {
		return m.AddDatasetFunc(projectID, name)
	}
	return nil, nil
}

func (m *MockClient) UpdateDataset(datasetID int64, name string) (*data.Dataset, error) {
	if m.UpdateDatasetFunc != nil {
		return m.UpdateDatasetFunc(datasetID, name)
	}
	return nil, nil
}

func (m *MockClient) DeleteDataset(datasetID int64) error {
	if m.DeleteDatasetFunc != nil {
		return m.DeleteDatasetFunc(datasetID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Variables
// ---------------------------------------------------------------------------
func (m *MockClient) GetVariables(datasetID int64) (data.GetVariablesResponse, error) {
	if m.GetVariablesFunc != nil {
		return m.GetVariablesFunc(datasetID)
	}
	return nil, nil
}

func (m *MockClient) AddVariable(datasetID int64, name string) (*data.Variable, error) {
	if m.AddVariableFunc != nil {
		return m.AddVariableFunc(datasetID, name)
	}
	return nil, nil
}

func (m *MockClient) UpdateVariable(variableID int64, name string) (*data.Variable, error) {
	if m.UpdateVariableFunc != nil {
		return m.UpdateVariableFunc(variableID, name)
	}
	return nil, nil
}

func (m *MockClient) DeleteVariable(variableID int64) error {
	if m.DeleteVariableFunc != nil {
		return m.DeleteVariableFunc(variableID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - BDDs
// ---------------------------------------------------------------------------
func (m *MockClient) GetBDD(caseID int64) (*data.BDD, error) {
	if m.GetBDDFunc != nil {
		return m.GetBDDFunc(caseID)
	}
	return nil, nil
}

func (m *MockClient) AddBDD(caseID int64, content string) (*data.BDD, error) {
	if m.AddBDDFunc != nil {
		return m.AddBDDFunc(caseID, content)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Labels
// ---------------------------------------------------------------------------
func (m *MockClient) GetLabels(projectID int64) (data.GetLabelsResponse, error) {
	if m.GetLabelsFunc != nil {
		return m.GetLabelsFunc(projectID)
	}
	return nil, nil
}

func (m *MockClient) GetLabel(labelID int64) (*data.Label, error) {
	if m.GetLabelFunc != nil {
		return m.GetLabelFunc(labelID)
	}
	return nil, nil
}

func (m *MockClient) UpdateLabel(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
	if m.UpdateLabelFunc != nil {
		return m.UpdateLabelFunc(labelID, req)
	}
	return nil, nil
}

func (m *MockClient) UpdateTestLabels(testID int64, labels []string) error {
	if m.UpdateTestLabelsFunc != nil {
		return m.UpdateTestLabelsFunc(testID, labels)
	}
	return nil
}

func (m *MockClient) UpdateTestsLabels(runID int64, testIDs []int64, labels []string) error {
	if m.UpdateTestsLabelsFunc != nil {
		return m.UpdateTestsLabelsFunc(runID, testIDs, labels)
	}
	return nil
}
