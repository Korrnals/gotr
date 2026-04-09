// internal/client/mock.go
// Universal mock client for testing.
package client

import (
	"context"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
)

// MockClient implements ClientInterface for testing.
type MockClient struct {
	// ProjectsAPI
	GetProjectsFunc   func(ctx context.Context) (data.GetProjectsResponse, error)
	GetProjectFunc    func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error)
	AddProjectFunc    func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error)
	UpdateProjectFunc func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error)
	DeleteProjectFunc func(ctx context.Context, projectID int64) error

	// CasesAPI
	GetCasesFunc           func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error)
	GetCasesPageFunc       func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error)
	GetCaseFunc            func(ctx context.Context, caseID int64) (*data.Case, error)
	AddCaseFunc            func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	UpdateCaseFunc         func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	DeleteCaseFunc         func(ctx context.Context, caseID int64) error
	UpdateCasesFunc        func(ctx context.Context, suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error)
	DeleteCasesFunc        func(ctx context.Context, suiteID int64, req *data.DeleteCasesRequest) error
	CopyCasesToSectionFunc func(ctx context.Context, sectionID int64, req *data.CopyCasesRequest) error
	MoveCasesToSectionFunc func(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error
	GetHistoryForCaseFunc  func(ctx context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error)
	GetCaseFieldsFunc      func(ctx context.Context) (data.GetCaseFieldsResponse, error)
	AddCaseFieldFunc       func(ctx context.Context, req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error)
	GetCaseTypesFunc       func(ctx context.Context) (data.GetCaseTypesResponse, error)
	DiffCasesDataFunc      func(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error)

	// ConcurrentAPI - Parallel methods (Stage 6.2)
	GetCasesParallelFunc          func(ctx context.Context, projectID int64, suiteIDs []int64, workers int) (map[int64]data.GetCasesResponse, error)
	GetCasesForSuitesParallelFunc func(ctx context.Context, projectID int64, suiteIDs []int64, workers int) (data.GetCasesResponse, error)
	GetCasesParallelCtxFunc       func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error)
	GetSuitesParallelFunc         func(ctx context.Context, projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error)

	// SuitesAPI
	GetSuitesFunc   func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error)
	GetSuiteFunc    func(ctx context.Context, suiteID int64) (*data.Suite, error)
	AddSuiteFunc    func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	UpdateSuiteFunc func(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error)
	DeleteSuiteFunc func(ctx context.Context, suiteID int64) error

	// ConcurrentAPI - Parallel methods (Stage 6.2)

	// SectionsAPI
	GetSectionsFunc            func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error)
	GetSectionsParallelCtxFunc func(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error)
	GetSectionFunc             func(ctx context.Context, sectionID int64) (*data.Section, error)
	AddSectionFunc             func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	UpdateSectionFunc          func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error)
	DeleteSectionFunc          func(ctx context.Context, sectionID int64) error

	// SharedStepsAPI
	GetSharedStepsFunc       func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error)
	GetSharedStepFunc        func(ctx context.Context, stepID int64) (*data.SharedStep, error)
	AddSharedStepFunc        func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	UpdateSharedStepFunc     func(ctx context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error)
	DeleteSharedStepFunc     func(ctx context.Context, stepID int64, keepInCases int) error
	GetSharedStepHistoryFunc func(ctx context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error)

	// RunsAPI
	GetRunsFunc   func(ctx context.Context, projectID int64) (data.GetRunsResponse, error)
	GetRunFunc    func(ctx context.Context, runID int64) (*data.Run, error)
	AddRunFunc    func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error)
	UpdateRunFunc func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	CloseRunFunc  func(ctx context.Context, runID int64) (*data.Run, error)
	DeleteRunFunc func(ctx context.Context, runID int64) error

	// ResultsAPI
	GetResultsFunc         func(ctx context.Context, testID int64) (data.GetResultsResponse, error)
	GetResultsForRunFunc   func(ctx context.Context, runID int64) (data.GetResultsResponse, error)
	GetResultsForCaseFunc  func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error)
	AddResultFunc          func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResultForCaseFunc   func(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResultsFunc         func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error)
	AddResultsForCasesFunc func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error)

	// TestsAPI
	GetTestFunc    func(ctx context.Context, testID int64) (*data.Test, error)
	GetTestsFunc   func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error)
	UpdateTestFunc func(ctx context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error)

	// MilestonesAPI
	GetMilestoneFunc    func(ctx context.Context, milestoneID int64) (*data.Milestone, error)
	GetMilestonesFunc   func(ctx context.Context, projectID int64) ([]data.Milestone, error)
	AddMilestoneFunc    func(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error)
	UpdateMilestoneFunc func(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error)
	DeleteMilestoneFunc func(ctx context.Context, milestoneID int64) error

	// PlansAPI
	GetPlanFunc         func(ctx context.Context, planID int64) (*data.Plan, error)
	GetPlansFunc        func(ctx context.Context, projectID int64) (data.GetPlansResponse, error)
	AddPlanFunc         func(ctx context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error)
	UpdatePlanFunc      func(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error)
	ClosePlanFunc       func(ctx context.Context, planID int64) (*data.Plan, error)
	DeletePlanFunc      func(ctx context.Context, planID int64) error
	AddPlanEntryFunc    func(ctx context.Context, planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error)
	UpdatePlanEntryFunc func(ctx context.Context, planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error)
	DeletePlanEntryFunc func(ctx context.Context, planID int64, entryID string) error

	// AttachmentsAPI
	AddAttachmentToCaseFunc        func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanFunc        func(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanEntryFunc   func(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToResultFunc      func(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToRunFunc         func(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error)
	DeleteAttachmentFunc           func(ctx context.Context, attachmentID int64) error
	GetAttachmentFunc              func(ctx context.Context, attachmentID int64) (*data.Attachment, error)
	GetAttachmentsForCaseFunc      func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForPlanFunc      func(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForPlanEntryFunc func(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error)
	GetAttachmentsForProjectFunc    func(ctx context.Context, projectID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForRunFunc       func(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForTestFunc      func(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error)

	// ConfigurationsAPI
	GetConfigsFunc        func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error)
	AddConfigGroupFunc    func(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error)
	AddConfigFunc         func(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error)
	UpdateConfigGroupFunc func(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error)
	UpdateConfigFunc      func(ctx context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error)
	DeleteConfigGroupFunc func(ctx context.Context, groupID int64) error
	DeleteConfigFunc      func(ctx context.Context, configID int64) error

	// UsersAPI
	GetUsersFunc          func(ctx context.Context) (data.GetUsersResponse, error)
	GetUsersByProjectFunc func(ctx context.Context, projectID int64) (data.GetUsersResponse, error)
	GetUserFunc           func(ctx context.Context, userID int64) (*data.User, error)
	GetUserByEmailFunc    func(ctx context.Context, email string) (*data.User, error)
	AddUserFunc           func(ctx context.Context, req data.AddUserRequest) (*data.User, error)
	UpdateUserFunc        func(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error)
	GetPrioritiesFunc     func(ctx context.Context) (data.GetPrioritiesResponse, error)
	GetStatusesFunc       func(ctx context.Context) (data.GetStatusesResponse, error)
	GetTemplatesFunc      func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error)

	// ReportsAPI
	GetReportsFunc             func(ctx context.Context, projectID int64) (data.GetReportsResponse, error)
	GetCrossProjectReportsFunc func(ctx context.Context) (data.GetReportsResponse, error)
	RunReportFunc              func(ctx context.Context, templateID int64) (*data.RunReportResponse, error)
	RunCrossProjectReportFunc  func(ctx context.Context, templateID int64) (*data.RunReportResponse, error)

	// ExtendedAPI - Groups
	GetGroupsFunc   func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error)
	GetGroupFunc    func(ctx context.Context, groupID int64) (*data.Group, error)
	AddGroupFunc    func(ctx context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error)
	UpdateGroupFunc func(ctx context.Context, groupID int64, name string, userIDs []int64) (*data.Group, error)
	DeleteGroupFunc func(ctx context.Context, groupID int64) error

	// ExtendedAPI - Roles
	GetRolesFunc func(ctx context.Context) (data.GetRolesResponse, error)
	GetRoleFunc  func(ctx context.Context, roleID int64) (*data.Role, error)

	// ExtendedAPI - ResultFields
	GetResultFieldsFunc func(ctx context.Context) (data.GetResultFieldsResponse, error)

	// ExtendedAPI - Datasets
	GetDatasetsFunc   func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error)
	GetDatasetFunc    func(ctx context.Context, datasetID int64) (*data.Dataset, error)
	AddDatasetFunc    func(ctx context.Context, projectID int64, name string) (*data.Dataset, error)
	UpdateDatasetFunc func(ctx context.Context, datasetID int64, name string) (*data.Dataset, error)
	DeleteDatasetFunc func(ctx context.Context, datasetID int64) error

	// ExtendedAPI - Variables
	GetVariablesFunc   func(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error)
	AddVariableFunc    func(ctx context.Context, datasetID int64, name string) (*data.Variable, error)
	UpdateVariableFunc func(ctx context.Context, variableID int64, name string) (*data.Variable, error)
	DeleteVariableFunc func(ctx context.Context, variableID int64) error

	// ExtendedAPI - BDDs
	GetBDDFunc func(ctx context.Context, caseID int64) (*data.BDD, error)
	AddBDDFunc func(ctx context.Context, caseID int64, content string) (*data.BDD, error)

	// ExtendedAPI - Labels
	GetLabelsFunc         func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error)
	GetLabelFunc          func(ctx context.Context, labelID int64) (*data.Label, error)
	UpdateLabelFunc       func(ctx context.Context, labelID int64, req data.UpdateLabelRequest) (*data.Label, error)
	UpdateTestLabelsFunc  func(ctx context.Context, testID int64, labels []string) error
	UpdateTestsLabelsFunc func(ctx context.Context, runID int64, testIDs []int64, labels []string) error
}

// Compile-time check: MockClient must implement ClientInterface.
var _ ClientInterface = (*MockClient)(nil)

// ---------------------------------------------------------------------------
// ProjectsAPI
// ---------------------------------------------------------------------------
// GetProjects calls the configured mock implementation when it is set.
func (m *MockClient) GetProjects(ctx context.Context) (data.GetProjectsResponse, error) {
	if m.GetProjectsFunc != nil {
		return m.GetProjectsFunc(ctx)
	}
	return nil, nil
}

// GetProject calls the configured mock implementation when it is set.
func (m *MockClient) GetProject(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, projectID)
	}
	return nil, nil
}

// AddProject calls the configured mock implementation when it is set.
func (m *MockClient) AddProject(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
	if m.AddProjectFunc != nil {
		return m.AddProjectFunc(ctx, req)
	}
	return nil, nil
}

// UpdateProject calls the configured mock implementation when it is set.
func (m *MockClient) UpdateProject(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
	if m.UpdateProjectFunc != nil {
		return m.UpdateProjectFunc(ctx, projectID, req)
	}
	return nil, nil
}

// DeleteProject calls the configured mock implementation when it is set.
func (m *MockClient) DeleteProject(ctx context.Context, projectID int64) error {
	if m.DeleteProjectFunc != nil {
		return m.DeleteProjectFunc(ctx, projectID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// CasesAPI
// ---------------------------------------------------------------------------
// GetCases calls the configured mock implementation when it is set.
func (m *MockClient) GetCases(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
	if m.GetCasesFunc != nil {
		return m.GetCasesFunc(ctx, projectID, suiteID, sectionID)
	}
	return nil, nil
}

// GetCasesPage calls the configured mock implementation when it is set.
func (m *MockClient) GetCasesPage(ctx context.Context, projectID, suiteID int64, offset, limit int) (data.GetCasesResponse, error) {
	if m.GetCasesPageFunc != nil {
		return m.GetCasesPageFunc(ctx, projectID, suiteID, offset, limit)
	}
	return nil, nil
}

// GetCase calls the configured mock implementation when it is set.
func (m *MockClient) GetCase(ctx context.Context, caseID int64) (*data.Case, error) {
	if m.GetCaseFunc != nil {
		return m.GetCaseFunc(ctx, caseID)
	}
	return nil, nil
}

// AddCase calls the configured mock implementation when it is set.
func (m *MockClient) AddCase(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
	if m.AddCaseFunc != nil {
		return m.AddCaseFunc(ctx, sectionID, req)
	}
	return &data.Case{ID: 999}, nil
}

// UpdateCase calls the configured mock implementation when it is set.
func (m *MockClient) UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
	if m.UpdateCaseFunc != nil {
		return m.UpdateCaseFunc(ctx, caseID, req)
	}
	return nil, nil
}

// DeleteCase calls the configured mock implementation when it is set.
func (m *MockClient) DeleteCase(ctx context.Context, caseID int64) error {
	if m.DeleteCaseFunc != nil {
		return m.DeleteCaseFunc(ctx, caseID)
	}
	return nil
}

// UpdateCases calls the configured mock implementation when it is set.
func (m *MockClient) UpdateCases(ctx context.Context, suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
	if m.UpdateCasesFunc != nil {
		return m.UpdateCasesFunc(ctx, suiteID, req)
	}
	return nil, nil
}

// DeleteCases calls the configured mock implementation when it is set.
func (m *MockClient) DeleteCases(ctx context.Context, suiteID int64, req *data.DeleteCasesRequest) error {
	if m.DeleteCasesFunc != nil {
		return m.DeleteCasesFunc(ctx, suiteID, req)
	}
	return nil
}

// CopyCasesToSection calls the configured mock implementation when it is set.
func (m *MockClient) CopyCasesToSection(ctx context.Context, sectionID int64, req *data.CopyCasesRequest) error {
	if m.CopyCasesToSectionFunc != nil {
		return m.CopyCasesToSectionFunc(ctx, sectionID, req)
	}
	return nil
}

// MoveCasesToSection calls the configured mock implementation when it is set.
func (m *MockClient) MoveCasesToSection(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error {
	if m.MoveCasesToSectionFunc != nil {
		return m.MoveCasesToSectionFunc(ctx, sectionID, req)
	}
	return nil
}

// GetHistoryForCase calls the configured mock implementation when it is set.
func (m *MockClient) GetHistoryForCase(ctx context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error) {
	if m.GetHistoryForCaseFunc != nil {
		return m.GetHistoryForCaseFunc(ctx, caseID)
	}
	return nil, nil
}

// GetCaseFields calls the configured mock implementation when it is set.
func (m *MockClient) GetCaseFields(ctx context.Context) (data.GetCaseFieldsResponse, error) {
	if m.GetCaseFieldsFunc != nil {
		return m.GetCaseFieldsFunc(ctx)
	}
	return nil, nil
}

// AddCaseField calls the configured mock implementation when it is set.
func (m *MockClient) AddCaseField(ctx context.Context, req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error) {
	if m.AddCaseFieldFunc != nil {
		return m.AddCaseFieldFunc(ctx, req)
	}
	return nil, nil
}

// GetCaseTypes calls the configured mock implementation when it is set.
func (m *MockClient) GetCaseTypes(ctx context.Context) (data.GetCaseTypesResponse, error) {
	if m.GetCaseTypesFunc != nil {
		return m.GetCaseTypesFunc(ctx)
	}
	return nil, nil
}

// DiffCasesData calls the configured mock implementation when it is set.
func (m *MockClient) DiffCasesData(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
	if m.DiffCasesDataFunc != nil {
		return m.DiffCasesDataFunc(ctx, pid1, pid2, field)
	}
	return nil, nil
}

// GetCasesParallel calls the configured mock implementation when it is set.
func (m *MockClient) GetCasesParallel(ctx context.Context, projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetCasesResponse, error) {
	if m.GetCasesParallelFunc != nil {
		return m.GetCasesParallelFunc(ctx, projectID, suiteIDs, workers)
	}
	// Default: sequential fallback
	results := make(map[int64]data.GetCasesResponse)
	for _, sid := range suiteIDs {
		cases, err := m.GetCases(ctx, projectID, sid, 0)
		if err != nil {
			return results, err
		}
		results[sid] = cases
		if monitor != nil {
			monitor.Increment()
		}
	}
	return results, nil
}

// GetSuitesParallel calls the configured mock implementation when it is set.
func (m *MockClient) GetSuitesParallel(ctx context.Context, projectIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetSuitesResponse, error) {
	results := make(map[int64]data.GetSuitesResponse)
	for _, pid := range projectIDs {
		suites, err := m.GetSuites(ctx, pid)
		if err != nil {
			return results, err
		}
		results[pid] = suites
		if monitor != nil {
			monitor.Increment()
		}
	}
	return results, nil
}

// GetCasesForSuitesParallel calls the configured mock implementation when it is set.
func (m *MockClient) GetCasesForSuitesParallel(ctx context.Context, projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (data.GetCasesResponse, error) {
	if m.GetCasesForSuitesParallelFunc != nil {
		return m.GetCasesForSuitesParallelFunc(ctx, projectID, suiteIDs, workers)
	}
	// Default: use GetCasesParallel and flatten
	results, err := m.GetCasesParallel(ctx, projectID, suiteIDs, workers, monitor)
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
// GetSuites calls the configured mock implementation when it is set.
func (m *MockClient) GetSuites(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
	if m.GetSuitesFunc != nil {
		return m.GetSuitesFunc(ctx, projectID)
	}
	return nil, nil
}

// GetSuite calls the configured mock implementation when it is set.
func (m *MockClient) GetSuite(ctx context.Context, suiteID int64) (*data.Suite, error) {
	if m.GetSuiteFunc != nil {
		return m.GetSuiteFunc(ctx, suiteID)
	}
	return nil, nil
}

// AddSuite calls the configured mock implementation when it is set.
func (m *MockClient) AddSuite(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	if m.AddSuiteFunc != nil {
		return m.AddSuiteFunc(ctx, projectID, req)
	}
	return &data.Suite{ID: 999}, nil
}

// UpdateSuite calls the configured mock implementation when it is set.
func (m *MockClient) UpdateSuite(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
	if m.UpdateSuiteFunc != nil {
		return m.UpdateSuiteFunc(ctx, suiteID, req)
	}
	return nil, nil
}

// DeleteSuite calls the configured mock implementation when it is set.
func (m *MockClient) DeleteSuite(ctx context.Context, suiteID int64) error {
	if m.DeleteSuiteFunc != nil {
		return m.DeleteSuiteFunc(ctx, suiteID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// SectionsAPI
// ---------------------------------------------------------------------------
// GetSections calls the configured mock implementation when it is set.
func (m *MockClient) GetSections(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
	if m.GetSectionsFunc != nil {
		return m.GetSectionsFunc(ctx, projectID, suiteID)
	}
	return nil, nil
}

// GetSectionsParallelCtx calls the configured mock implementation when it is set.
func (m *MockClient) GetSectionsParallelCtx(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
	if m.GetSectionsParallelCtxFunc != nil {
		return m.GetSectionsParallelCtxFunc(ctx, projectID, suiteIDs, config)
	}

	if len(suiteIDs) == 0 {
		return m.GetSections(ctx, projectID, 0)
	}

	all := make(data.GetSectionsResponse, 0)
	for _, suiteID := range suiteIDs {
		sections, err := m.GetSections(ctx, projectID, suiteID)
		if err != nil {
			return all, err
		}
		all = append(all, sections...)
	}

	return all, nil
}

// GetSection calls the configured mock implementation when it is set.
func (m *MockClient) GetSection(ctx context.Context, sectionID int64) (*data.Section, error) {
	if m.GetSectionFunc != nil {
		return m.GetSectionFunc(ctx, sectionID)
	}
	return nil, nil
}

// AddSection calls the configured mock implementation when it is set.
func (m *MockClient) AddSection(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	if m.AddSectionFunc != nil {
		return m.AddSectionFunc(ctx, projectID, req)
	}
	return &data.Section{ID: 999}, nil
}

// UpdateSection calls the configured mock implementation when it is set.
func (m *MockClient) UpdateSection(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
	if m.UpdateSectionFunc != nil {
		return m.UpdateSectionFunc(ctx, sectionID, req)
	}
	return nil, nil
}

// DeleteSection calls the configured mock implementation when it is set.
func (m *MockClient) DeleteSection(ctx context.Context, sectionID int64) error {
	if m.DeleteSectionFunc != nil {
		return m.DeleteSectionFunc(ctx, sectionID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// SharedStepsAPI
// ---------------------------------------------------------------------------
// GetSharedSteps calls the configured mock implementation when it is set.
func (m *MockClient) GetSharedSteps(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
	if m.GetSharedStepsFunc != nil {
		return m.GetSharedStepsFunc(ctx, projectID)
	}
	return nil, nil
}

// GetSharedStep calls the configured mock implementation when it is set.
func (m *MockClient) GetSharedStep(ctx context.Context, stepID int64) (*data.SharedStep, error) {
	if m.GetSharedStepFunc != nil {
		return m.GetSharedStepFunc(ctx, stepID)
	}
	return nil, nil
}

// AddSharedStep calls the configured mock implementation when it is set.
func (m *MockClient) AddSharedStep(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
	if m.AddSharedStepFunc != nil {
		return m.AddSharedStepFunc(ctx, projectID, req)
	}
	return &data.SharedStep{ID: 999}, nil
}

// UpdateSharedStep calls the configured mock implementation when it is set.
func (m *MockClient) UpdateSharedStep(ctx context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
	if m.UpdateSharedStepFunc != nil {
		return m.UpdateSharedStepFunc(ctx, stepID, req)
	}
	return nil, nil
}

// DeleteSharedStep calls the configured mock implementation when it is set.
func (m *MockClient) DeleteSharedStep(ctx context.Context, stepID int64, keepInCases int) error {
	if m.DeleteSharedStepFunc != nil {
		return m.DeleteSharedStepFunc(ctx, stepID, keepInCases)
	}
	return nil
}

// GetSharedStepHistory calls the configured mock implementation when it is set.
func (m *MockClient) GetSharedStepHistory(ctx context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error) {
	if m.GetSharedStepHistoryFunc != nil {
		return m.GetSharedStepHistoryFunc(ctx, stepID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// RunsAPI
// ---------------------------------------------------------------------------
// GetRuns calls the configured mock implementation when it is set.
func (m *MockClient) GetRuns(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
	if m.GetRunsFunc != nil {
		return m.GetRunsFunc(ctx, projectID)
	}
	return nil, nil
}

// GetRun calls the configured mock implementation when it is set.
func (m *MockClient) GetRun(ctx context.Context, runID int64) (*data.Run, error) {
	if m.GetRunFunc != nil {
		return m.GetRunFunc(ctx, runID)
	}
	return nil, nil
}

// AddRun calls the configured mock implementation when it is set.
func (m *MockClient) AddRun(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	if m.AddRunFunc != nil {
		return m.AddRunFunc(ctx, projectID, req)
	}
	return nil, nil
}

// UpdateRun calls the configured mock implementation when it is set.
func (m *MockClient) UpdateRun(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	if m.UpdateRunFunc != nil {
		return m.UpdateRunFunc(ctx, runID, req)
	}
	return nil, nil
}

// CloseRun calls the configured mock implementation when it is set.
func (m *MockClient) CloseRun(ctx context.Context, runID int64) (*data.Run, error) {
	if m.CloseRunFunc != nil {
		return m.CloseRunFunc(ctx, runID)
	}
	return nil, nil
}

// DeleteRun calls the configured mock implementation when it is set.
func (m *MockClient) DeleteRun(ctx context.Context, runID int64) error {
	if m.DeleteRunFunc != nil {
		return m.DeleteRunFunc(ctx, runID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ResultsAPI
// ---------------------------------------------------------------------------
// GetResults calls the configured mock implementation when it is set.
func (m *MockClient) GetResults(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
	if m.GetResultsFunc != nil {
		return m.GetResultsFunc(ctx, testID)
	}
	return nil, nil
}

// GetResultsForRun calls the configured mock implementation when it is set.
func (m *MockClient) GetResultsForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
	if m.GetResultsForRunFunc != nil {
		return m.GetResultsForRunFunc(ctx, runID)
	}
	return nil, nil
}

// GetResultsForCase calls the configured mock implementation when it is set.
func (m *MockClient) GetResultsForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
	if m.GetResultsForCaseFunc != nil {
		return m.GetResultsForCaseFunc(ctx, runID, caseID)
	}
	return nil, nil
}

// AddResult calls the configured mock implementation when it is set.
func (m *MockClient) AddResult(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
	if m.AddResultFunc != nil {
		return m.AddResultFunc(ctx, testID, req)
	}
	return nil, nil
}

// AddResultForCase calls the configured mock implementation when it is set.
func (m *MockClient) AddResultForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
	if m.AddResultForCaseFunc != nil {
		return m.AddResultForCaseFunc(ctx, runID, caseID, req)
	}
	return nil, nil
}

// AddResults calls the configured mock implementation when it is set.
func (m *MockClient) AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
	if m.AddResultsFunc != nil {
		return m.AddResultsFunc(ctx, runID, req)
	}
	return nil, nil
}

// AddResultsForCases calls the configured mock implementation when it is set.
func (m *MockClient) AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
	if m.AddResultsForCasesFunc != nil {
		return m.AddResultsForCasesFunc(ctx, runID, req)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// TestsAPI
// ---------------------------------------------------------------------------
// GetTest calls the configured mock implementation when it is set.
func (m *MockClient) GetTest(ctx context.Context, testID int64) (*data.Test, error) {
	if m.GetTestFunc != nil {
		return m.GetTestFunc(ctx, testID)
	}
	return nil, nil
}

// GetTests calls the configured mock implementation when it is set.
func (m *MockClient) GetTests(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
	if m.GetTestsFunc != nil {
		return m.GetTestsFunc(ctx, runID, filters)
	}
	return nil, nil
}

// UpdateTest calls the configured mock implementation when it is set.
func (m *MockClient) UpdateTest(ctx context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
	if m.UpdateTestFunc != nil {
		return m.UpdateTestFunc(ctx, testID, req)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// MilestonesAPI
// ---------------------------------------------------------------------------
// GetMilestone calls the configured mock implementation when it is set.
func (m *MockClient) GetMilestone(ctx context.Context, milestoneID int64) (*data.Milestone, error) {
	if m.GetMilestoneFunc != nil {
		return m.GetMilestoneFunc(ctx, milestoneID)
	}
	return nil, nil
}

// GetMilestones calls the configured mock implementation when it is set.
func (m *MockClient) GetMilestones(ctx context.Context, projectID int64) ([]data.Milestone, error) {
	if m.GetMilestonesFunc != nil {
		return m.GetMilestonesFunc(ctx, projectID)
	}
	return nil, nil
}

// AddMilestone calls the configured mock implementation when it is set.
func (m *MockClient) AddMilestone(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
	if m.AddMilestoneFunc != nil {
		return m.AddMilestoneFunc(ctx, projectID, req)
	}
	return nil, nil
}

// UpdateMilestone calls the configured mock implementation when it is set.
func (m *MockClient) UpdateMilestone(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
	if m.UpdateMilestoneFunc != nil {
		return m.UpdateMilestoneFunc(ctx, milestoneID, req)
	}
	return nil, nil
}

// DeleteMilestone calls the configured mock implementation when it is set.
func (m *MockClient) DeleteMilestone(ctx context.Context, milestoneID int64) error {
	if m.DeleteMilestoneFunc != nil {
		return m.DeleteMilestoneFunc(ctx, milestoneID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// PlansAPI
// ---------------------------------------------------------------------------
// GetPlan calls the configured mock implementation when it is set.
func (m *MockClient) GetPlan(ctx context.Context, planID int64) (*data.Plan, error) {
	if m.GetPlanFunc != nil {
		return m.GetPlanFunc(ctx, planID)
	}
	return nil, nil
}

// GetPlans calls the configured mock implementation when it is set.
func (m *MockClient) GetPlans(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
	if m.GetPlansFunc != nil {
		return m.GetPlansFunc(ctx, projectID)
	}
	return nil, nil
}

// AddPlan calls the configured mock implementation when it is set.
func (m *MockClient) AddPlan(ctx context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
	if m.AddPlanFunc != nil {
		return m.AddPlanFunc(ctx, projectID, req)
	}
	return nil, nil
}

// UpdatePlan calls the configured mock implementation when it is set.
func (m *MockClient) UpdatePlan(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
	if m.UpdatePlanFunc != nil {
		return m.UpdatePlanFunc(ctx, planID, req)
	}
	return nil, nil
}

// ClosePlan calls the configured mock implementation when it is set.
func (m *MockClient) ClosePlan(ctx context.Context, planID int64) (*data.Plan, error) {
	if m.ClosePlanFunc != nil {
		return m.ClosePlanFunc(ctx, planID)
	}
	return nil, nil
}

// DeletePlan calls the configured mock implementation when it is set.
func (m *MockClient) DeletePlan(ctx context.Context, planID int64) error {
	if m.DeletePlanFunc != nil {
		return m.DeletePlanFunc(ctx, planID)
	}
	return nil
}

// AddPlanEntry calls the configured mock implementation when it is set.
func (m *MockClient) AddPlanEntry(ctx context.Context, planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
	if m.AddPlanEntryFunc != nil {
		return m.AddPlanEntryFunc(ctx, planID, req)
	}
	return nil, nil
}

// UpdatePlanEntry calls the configured mock implementation when it is set.
func (m *MockClient) UpdatePlanEntry(ctx context.Context, planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
	if m.UpdatePlanEntryFunc != nil {
		return m.UpdatePlanEntryFunc(ctx, planID, entryID, req)
	}
	return nil, nil
}

// DeletePlanEntry calls the configured mock implementation when it is set.
func (m *MockClient) DeletePlanEntry(ctx context.Context, planID int64, entryID string) error {
	if m.DeletePlanEntryFunc != nil {
		return m.DeletePlanEntryFunc(ctx, planID, entryID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// AttachmentsAPI
// ---------------------------------------------------------------------------
// AddAttachmentToCase calls the configured mock implementation when it is set.
func (m *MockClient) AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToCaseFunc != nil {
		return m.AddAttachmentToCaseFunc(ctx, caseID, filePath)
	}
	return nil, nil
}

// AddAttachmentToPlan calls the configured mock implementation when it is set.
func (m *MockClient) AddAttachmentToPlan(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToPlanFunc != nil {
		return m.AddAttachmentToPlanFunc(ctx, planID, filePath)
	}
	return nil, nil
}

// AddAttachmentToPlanEntry calls the configured mock implementation when it is set.
func (m *MockClient) AddAttachmentToPlanEntry(ctx context.Context, planID int64, entryID, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToPlanEntryFunc != nil {
		return m.AddAttachmentToPlanEntryFunc(ctx, planID, entryID, filePath)
	}
	return nil, nil
}

// AddAttachmentToResult calls the configured mock implementation when it is set.
func (m *MockClient) AddAttachmentToResult(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToResultFunc != nil {
		return m.AddAttachmentToResultFunc(ctx, resultID, filePath)
	}
	return nil, nil
}

// AddAttachmentToRun calls the configured mock implementation when it is set.
func (m *MockClient) AddAttachmentToRun(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
	if m.AddAttachmentToRunFunc != nil {
		return m.AddAttachmentToRunFunc(ctx, runID, filePath)
	}
	return nil, nil
}

// DeleteAttachment calls the configured mock implementation when it is set.
func (m *MockClient) DeleteAttachment(ctx context.Context, attachmentID int64) error {
	if m.DeleteAttachmentFunc != nil {
		return m.DeleteAttachmentFunc(ctx, attachmentID)
	}
	return nil
}

// GetAttachment calls the configured mock implementation when it is set.
func (m *MockClient) GetAttachment(ctx context.Context, attachmentID int64) (*data.Attachment, error) {
	if m.GetAttachmentFunc != nil {
		return m.GetAttachmentFunc(ctx, attachmentID)
	}
	return nil, nil
}

// GetAttachmentsForCase calls the configured mock implementation when it is set.
func (m *MockClient) GetAttachmentsForCase(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForCaseFunc != nil {
		return m.GetAttachmentsForCaseFunc(ctx, caseID)
	}
	return nil, nil
}

// GetAttachmentsForPlan calls the configured mock implementation when it is set.
func (m *MockClient) GetAttachmentsForPlan(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForPlanFunc != nil {
		return m.GetAttachmentsForPlanFunc(ctx, planID)
	}
	return nil, nil
}

// GetAttachmentsForPlanEntry calls the configured mock implementation when it is set.
func (m *MockClient) GetAttachmentsForPlanEntry(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForPlanEntryFunc != nil {
		return m.GetAttachmentsForPlanEntryFunc(ctx, planID, entryID)
	}
	return nil, nil
}

// GetAttachmentsForRun calls the configured mock implementation when it is set.
func (m *MockClient) GetAttachmentsForRun(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForRunFunc != nil {
		return m.GetAttachmentsForRunFunc(ctx, runID)
	}
	return nil, nil
}

// GetAttachmentsForProject calls the configured mock implementation when it is set.
func (m *MockClient) GetAttachmentsForProject(ctx context.Context, projectID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForProjectFunc != nil {
		return m.GetAttachmentsForProjectFunc(ctx, projectID)
	}
	return nil, nil
}

// GetAttachmentsForTest calls the configured mock implementation when it is set.
func (m *MockClient) GetAttachmentsForTest(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error) {
	if m.GetAttachmentsForTestFunc != nil {
		return m.GetAttachmentsForTestFunc(ctx, testID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ConfigurationsAPI
// ---------------------------------------------------------------------------
// GetConfigs calls the configured mock implementation when it is set.
func (m *MockClient) GetConfigs(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
	if m.GetConfigsFunc != nil {
		return m.GetConfigsFunc(ctx, projectID)
	}
	return nil, nil
}

// AddConfigGroup calls the configured mock implementation when it is set.
func (m *MockClient) AddConfigGroup(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
	if m.AddConfigGroupFunc != nil {
		return m.AddConfigGroupFunc(ctx, projectID, req)
	}
	return nil, nil
}

// AddConfig calls the configured mock implementation when it is set.
func (m *MockClient) AddConfig(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
	if m.AddConfigFunc != nil {
		return m.AddConfigFunc(ctx, groupID, req)
	}
	return nil, nil
}

// UpdateConfigGroup calls the configured mock implementation when it is set.
func (m *MockClient) UpdateConfigGroup(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
	if m.UpdateConfigGroupFunc != nil {
		return m.UpdateConfigGroupFunc(ctx, groupID, req)
	}
	return nil, nil
}

// UpdateConfig calls the configured mock implementation when it is set.
func (m *MockClient) UpdateConfig(ctx context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
	if m.UpdateConfigFunc != nil {
		return m.UpdateConfigFunc(ctx, configID, req)
	}
	return nil, nil
}

// DeleteConfigGroup calls the configured mock implementation when it is set.
func (m *MockClient) DeleteConfigGroup(ctx context.Context, groupID int64) error {
	if m.DeleteConfigGroupFunc != nil {
		return m.DeleteConfigGroupFunc(ctx, groupID)
	}
	return nil
}

// DeleteConfig calls the configured mock implementation when it is set.
func (m *MockClient) DeleteConfig(ctx context.Context, configID int64) error {
	if m.DeleteConfigFunc != nil {
		return m.DeleteConfigFunc(ctx, configID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// UsersAPI
// ---------------------------------------------------------------------------
// GetUsers calls the configured mock implementation when it is set.
func (m *MockClient) GetUsers(ctx context.Context) (data.GetUsersResponse, error) {
	if m.GetUsersFunc != nil {
		return m.GetUsersFunc(ctx)
	}
	return nil, nil
}

// GetUsersByProject calls the configured mock implementation when it is set.
func (m *MockClient) GetUsersByProject(ctx context.Context, projectID int64) (data.GetUsersResponse, error) {
	if m.GetUsersByProjectFunc != nil {
		return m.GetUsersByProjectFunc(ctx, projectID)
	}
	return nil, nil
}

// GetUser calls the configured mock implementation when it is set.
func (m *MockClient) GetUser(ctx context.Context, userID int64) (*data.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return nil, nil
}

// GetUserByEmail calls the configured mock implementation when it is set.
func (m *MockClient) GetUserByEmail(ctx context.Context, email string) (*data.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

// AddUser calls the configured mock implementation when it is set.
func (m *MockClient) AddUser(ctx context.Context, req data.AddUserRequest) (*data.User, error) {
	if m.AddUserFunc != nil {
		return m.AddUserFunc(ctx, req)
	}
	return &data.User{ID: 999}, nil
}

// UpdateUser calls the configured mock implementation when it is set.
func (m *MockClient) UpdateUser(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, userID, req)
	}
	return nil, nil
}

// GetPriorities calls the configured mock implementation when it is set.
func (m *MockClient) GetPriorities(ctx context.Context) (data.GetPrioritiesResponse, error) {
	if m.GetPrioritiesFunc != nil {
		return m.GetPrioritiesFunc(ctx)
	}
	return nil, nil
}

// GetStatuses calls the configured mock implementation when it is set.
func (m *MockClient) GetStatuses(ctx context.Context) (data.GetStatusesResponse, error) {
	if m.GetStatusesFunc != nil {
		return m.GetStatusesFunc(ctx)
	}
	return nil, nil
}

// GetTemplates calls the configured mock implementation when it is set.
func (m *MockClient) GetTemplates(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
	if m.GetTemplatesFunc != nil {
		return m.GetTemplatesFunc(ctx, projectID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ReportsAPI
// ---------------------------------------------------------------------------
// GetReports calls the configured mock implementation when it is set.
func (m *MockClient) GetReports(ctx context.Context, projectID int64) (data.GetReportsResponse, error) {
	if m.GetReportsFunc != nil {
		return m.GetReportsFunc(ctx, projectID)
	}
	return nil, nil
}

// GetCrossProjectReports calls the configured mock implementation when it is set.
func (m *MockClient) GetCrossProjectReports(ctx context.Context) (data.GetReportsResponse, error) {
	if m.GetCrossProjectReportsFunc != nil {
		return m.GetCrossProjectReportsFunc(ctx)
	}
	return nil, nil
}

// RunReport calls the configured mock implementation when it is set.
func (m *MockClient) RunReport(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
	if m.RunReportFunc != nil {
		return m.RunReportFunc(ctx, templateID)
	}
	return nil, nil
}

// RunCrossProjectReport calls the configured mock implementation when it is set.
func (m *MockClient) RunCrossProjectReport(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
	if m.RunCrossProjectReportFunc != nil {
		return m.RunCrossProjectReportFunc(ctx, templateID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Groups
// ---------------------------------------------------------------------------
// GetGroups calls the configured mock implementation when it is set.
func (m *MockClient) GetGroups(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
	if m.GetGroupsFunc != nil {
		return m.GetGroupsFunc(ctx, projectID)
	}
	return nil, nil
}

// GetGroup calls the configured mock implementation when it is set.
func (m *MockClient) GetGroup(ctx context.Context, groupID int64) (*data.Group, error) {
	if m.GetGroupFunc != nil {
		return m.GetGroupFunc(ctx, groupID)
	}
	return nil, nil
}

// AddGroup calls the configured mock implementation when it is set.
func (m *MockClient) AddGroup(ctx context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error) {
	if m.AddGroupFunc != nil {
		return m.AddGroupFunc(ctx, projectID, name, userIDs)
	}
	return nil, nil
}

// UpdateGroup calls the configured mock implementation when it is set.
func (m *MockClient) UpdateGroup(ctx context.Context, groupID int64, name string, userIDs []int64) (*data.Group, error) {
	if m.UpdateGroupFunc != nil {
		return m.UpdateGroupFunc(ctx, groupID, name, userIDs)
	}
	return nil, nil
}

// DeleteGroup calls the configured mock implementation when it is set.
func (m *MockClient) DeleteGroup(ctx context.Context, groupID int64) error {
	if m.DeleteGroupFunc != nil {
		return m.DeleteGroupFunc(ctx, groupID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Roles
// ---------------------------------------------------------------------------
// GetRoles calls the configured mock implementation when it is set.
func (m *MockClient) GetRoles(ctx context.Context) (data.GetRolesResponse, error) {
	if m.GetRolesFunc != nil {
		return m.GetRolesFunc(ctx)
	}
	return nil, nil
}

// GetRole calls the configured mock implementation when it is set.
func (m *MockClient) GetRole(ctx context.Context, roleID int64) (*data.Role, error) {
	if m.GetRoleFunc != nil {
		return m.GetRoleFunc(ctx, roleID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - ResultFields
// ---------------------------------------------------------------------------
// GetResultFields calls the configured mock implementation when it is set.
func (m *MockClient) GetResultFields(ctx context.Context) (data.GetResultFieldsResponse, error) {
	if m.GetResultFieldsFunc != nil {
		return m.GetResultFieldsFunc(ctx)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Datasets
// ---------------------------------------------------------------------------
// GetDatasets calls the configured mock implementation when it is set.
func (m *MockClient) GetDatasets(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
	if m.GetDatasetsFunc != nil {
		return m.GetDatasetsFunc(ctx, projectID)
	}
	return nil, nil
}

// GetDataset calls the configured mock implementation when it is set.
func (m *MockClient) GetDataset(ctx context.Context, datasetID int64) (*data.Dataset, error) {
	if m.GetDatasetFunc != nil {
		return m.GetDatasetFunc(ctx, datasetID)
	}
	return nil, nil
}

// AddDataset calls the configured mock implementation when it is set.
func (m *MockClient) AddDataset(ctx context.Context, projectID int64, name string) (*data.Dataset, error) {
	if m.AddDatasetFunc != nil {
		return m.AddDatasetFunc(ctx, projectID, name)
	}
	return nil, nil
}

// UpdateDataset calls the configured mock implementation when it is set.
func (m *MockClient) UpdateDataset(ctx context.Context, datasetID int64, name string) (*data.Dataset, error) {
	if m.UpdateDatasetFunc != nil {
		return m.UpdateDatasetFunc(ctx, datasetID, name)
	}
	return nil, nil
}

// DeleteDataset calls the configured mock implementation when it is set.
func (m *MockClient) DeleteDataset(ctx context.Context, datasetID int64) error {
	if m.DeleteDatasetFunc != nil {
		return m.DeleteDatasetFunc(ctx, datasetID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Variables
// ---------------------------------------------------------------------------
// GetVariables calls the configured mock implementation when it is set.
func (m *MockClient) GetVariables(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
	if m.GetVariablesFunc != nil {
		return m.GetVariablesFunc(ctx, datasetID)
	}
	return nil, nil
}

// AddVariable calls the configured mock implementation when it is set.
func (m *MockClient) AddVariable(ctx context.Context, datasetID int64, name string) (*data.Variable, error) {
	if m.AddVariableFunc != nil {
		return m.AddVariableFunc(ctx, datasetID, name)
	}
	return nil, nil
}

// UpdateVariable calls the configured mock implementation when it is set.
func (m *MockClient) UpdateVariable(ctx context.Context, variableID int64, name string) (*data.Variable, error) {
	if m.UpdateVariableFunc != nil {
		return m.UpdateVariableFunc(ctx, variableID, name)
	}
	return nil, nil
}

// DeleteVariable calls the configured mock implementation when it is set.
func (m *MockClient) DeleteVariable(ctx context.Context, variableID int64) error {
	if m.DeleteVariableFunc != nil {
		return m.DeleteVariableFunc(ctx, variableID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - BDDs
// ---------------------------------------------------------------------------
// GetBDD calls the configured mock implementation when it is set.
func (m *MockClient) GetBDD(ctx context.Context, caseID int64) (*data.BDD, error) {
	if m.GetBDDFunc != nil {
		return m.GetBDDFunc(ctx, caseID)
	}
	return nil, nil
}

// AddBDD calls the configured mock implementation when it is set.
func (m *MockClient) AddBDD(ctx context.Context, caseID int64, content string) (*data.BDD, error) {
	if m.AddBDDFunc != nil {
		return m.AddBDDFunc(ctx, caseID, content)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ExtendedAPI - Labels
// ---------------------------------------------------------------------------
// GetLabels calls the configured mock implementation when it is set.
func (m *MockClient) GetLabels(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
	if m.GetLabelsFunc != nil {
		return m.GetLabelsFunc(ctx, projectID)
	}
	return nil, nil
}

// GetLabel calls the configured mock implementation when it is set.
func (m *MockClient) GetLabel(ctx context.Context, labelID int64) (*data.Label, error) {
	if m.GetLabelFunc != nil {
		return m.GetLabelFunc(ctx, labelID)
	}
	return nil, nil
}

// UpdateLabel calls the configured mock implementation when it is set.
func (m *MockClient) UpdateLabel(ctx context.Context, labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
	if m.UpdateLabelFunc != nil {
		return m.UpdateLabelFunc(ctx, labelID, req)
	}
	return nil, nil
}

// UpdateTestLabels calls the configured mock implementation when it is set.
func (m *MockClient) UpdateTestLabels(ctx context.Context, testID int64, labels []string) error {
	if m.UpdateTestLabelsFunc != nil {
		return m.UpdateTestLabelsFunc(ctx, testID, labels)
	}
	return nil
}

// UpdateTestsLabels calls the configured mock implementation when it is set.
func (m *MockClient) UpdateTestsLabels(ctx context.Context, runID int64, testIDs []int64, labels []string) error {
	if m.UpdateTestsLabelsFunc != nil {
		return m.UpdateTestsLabelsFunc(ctx, runID, testIDs, labels)
	}
	return nil
}

// GetCasesParallelCtx — mock implementation for Stage 6.7
// GetCasesParallelCtx calls the configured mock implementation when it is set.
func (m *MockClient) GetCasesParallelCtx(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
	if m.GetCasesParallelCtxFunc != nil {
		return m.GetCasesParallelCtxFunc(ctx, projectID, suiteIDs, config)
	}
	// Default: delegate to GetCasesForSuitesParallel
	workers := 5
	if config != nil && config.MaxConcurrentSuites > 0 {
		workers = config.MaxConcurrentSuites
	}
	cases, err := m.GetCasesForSuitesParallel(ctx, projectID, suiteIDs, workers, nil)
	result := &concurrency.ExecutionResult{
		Cases: cases,
	}
	return cases, result, err
}
