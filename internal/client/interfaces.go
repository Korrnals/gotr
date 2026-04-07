// internal/client/interfaces.go
// Full interfaces for the TestRail API HTTP client.
package client

import (
	"context"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
)

// ProgressMonitor defines an interface for tracking operation progress.
type ProgressMonitor interface {
	Increment()
	IncrementBy(n int)
}

// ProjectsAPI — project operations.
type ProjectsAPI interface {
	GetProjects(ctx context.Context) (data.GetProjectsResponse, error)
	GetProject(ctx context.Context, projectID int64) (*data.GetProjectResponse, error)
	AddProject(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error)
	UpdateProject(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error)
	DeleteProject(ctx context.Context, projectID int64) error
}

// CasesAPI — test case operations.
type CasesAPI interface {
	GetCases(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error)
	GetCasesPage(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error)
	GetCase(ctx context.Context, caseID int64) (*data.Case, error)
	AddCase(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	DeleteCase(ctx context.Context, caseID int64) error
	UpdateCases(ctx context.Context, suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error)
	DeleteCases(ctx context.Context, suiteID int64, req *data.DeleteCasesRequest) error
	CopyCasesToSection(ctx context.Context, sectionID int64, req *data.CopyCasesRequest) error
	MoveCasesToSection(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error
	GetHistoryForCase(ctx context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error)
	GetCaseFields(ctx context.Context) (data.GetCaseFieldsResponse, error)
	AddCaseField(ctx context.Context, req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error)
	GetCaseTypes(ctx context.Context) (data.GetCaseTypesResponse, error)
	DiffCasesData(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error)

	// Parallel methods (Stage 6.2, 6.7)
	// GetCasesParallel fetches cases from multiple suites in parallel (Legacy, use GetCasesForSuitesParallel).
	GetCasesParallel(ctx context.Context, projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetCasesResponse, error)
	// GetSuitesParallel fetches suites from multiple projects in parallel.
	GetSuitesParallel(ctx context.Context, projectIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetSuitesResponse, error)
	// GetCasesForSuitesParallel fetches all cases for a list of suites within one project.
	// Uses recursive parallelization (Stage 6.7) for maximum performance
	GetCasesForSuitesParallel(ctx context.Context, projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (data.GetCasesResponse, error)
	// GetCasesParallelCtx fetches cases from multiple suites with full control (Stage 6.7).
	// Uses streaming parallelization via concurrency.Controller.
	// Progress reporting: set config.Reporter (implements concurrency.ProgressReporter).
	GetCasesParallelCtx(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error)
}

// SuitesAPI — test suite operations.
type SuitesAPI interface {
	GetSuites(ctx context.Context, projectID int64) (data.GetSuitesResponse, error)
	GetSuite(ctx context.Context, suiteID int64) (*data.Suite, error)
	AddSuite(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	UpdateSuite(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error)
	DeleteSuite(ctx context.Context, suiteID int64) error
}

// SectionsAPI — section operations.
type SectionsAPI interface {
	GetSections(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error)
	// GetSectionsParallelCtx gets sections for multiple suites using shared concurrency runtime controls.
	GetSectionsParallelCtx(ctx context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error)
	GetSection(ctx context.Context, sectionID int64) (*data.Section, error)
	AddSection(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	UpdateSection(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error)
	DeleteSection(ctx context.Context, sectionID int64) error
}

// SharedStepsAPI — shared step operations.
type SharedStepsAPI interface {
	GetSharedSteps(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error)
	GetSharedStep(ctx context.Context, stepID int64) (*data.SharedStep, error)
	AddSharedStep(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	UpdateSharedStep(ctx context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error)
	DeleteSharedStep(ctx context.Context, stepID int64, keepInCases int) error
	GetSharedStepHistory(ctx context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error)
}

// RunsAPI — test run operations.
type RunsAPI interface {
	GetRuns(ctx context.Context, projectID int64) (data.GetRunsResponse, error)
	GetRun(ctx context.Context, runID int64) (*data.Run, error)
	AddRun(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error)
	UpdateRun(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	CloseRun(ctx context.Context, runID int64) (*data.Run, error)
	DeleteRun(ctx context.Context, runID int64) error
}

// ResultsAPI — test result operations.
type ResultsAPI interface {
	GetResults(ctx context.Context, testID int64) (data.GetResultsResponse, error)
	GetResultsForRun(ctx context.Context, runID int64) (data.GetResultsResponse, error)
	GetResultsForCase(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error)
	AddResult(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResultForCase(ctx context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResults(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error)
	AddResultsForCases(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error)
}

// TestsAPI — test operations.
type TestsAPI interface {
	GetTest(ctx context.Context, testID int64) (*data.Test, error)
	GetTests(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error)
	UpdateTest(ctx context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error)
}

// MilestonesAPI — milestone operations.
type MilestonesAPI interface {
	GetMilestone(ctx context.Context, milestoneID int64) (*data.Milestone, error)
	GetMilestones(ctx context.Context, projectID int64) ([]data.Milestone, error)
	AddMilestone(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error)
	UpdateMilestone(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error)
	DeleteMilestone(ctx context.Context, milestoneID int64) error
}

// PlansAPI — test plan operations.
type PlansAPI interface {
	GetPlan(ctx context.Context, planID int64) (*data.Plan, error)
	GetPlans(ctx context.Context, projectID int64) (data.GetPlansResponse, error)
	AddPlan(ctx context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error)
	UpdatePlan(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error)
	ClosePlan(ctx context.Context, planID int64) (*data.Plan, error)
	DeletePlan(ctx context.Context, planID int64) error
	AddPlanEntry(ctx context.Context, planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error)
	UpdatePlanEntry(ctx context.Context, planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error)
	DeletePlanEntry(ctx context.Context, planID int64, entryID string) error
}

// AttachmentsAPI — attachment operations.
type AttachmentsAPI interface {
	AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlan(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanEntry(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToResult(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToRun(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error)
	DeleteAttachment(ctx context.Context, attachmentID int64) error
	GetAttachment(ctx context.Context, attachmentID int64) (*data.Attachment, error)
	GetAttachmentsForCase(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForPlan(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForPlanEntry(ctx context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error)
	GetAttachmentsForProject(ctx context.Context, projectID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForRun(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error)
	GetAttachmentsForTest(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error)
}

// ConfigurationsAPI — configuration operations.
type ConfigurationsAPI interface {
	GetConfigs(ctx context.Context, projectID int64) (data.GetConfigsResponse, error)
	AddConfigGroup(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error)
	AddConfig(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error)
	UpdateConfigGroup(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error)
	UpdateConfig(ctx context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error)
	DeleteConfigGroup(ctx context.Context, groupID int64) error
	DeleteConfig(ctx context.Context, configID int64) error
}

// UsersAPI — user and reference data operations.
type UsersAPI interface {
	GetUsers(ctx context.Context) (data.GetUsersResponse, error)
	GetUsersByProject(ctx context.Context, projectID int64) (data.GetUsersResponse, error)
	GetUser(ctx context.Context, userID int64) (*data.User, error)
	GetUserByEmail(ctx context.Context, email string) (*data.User, error)
	AddUser(ctx context.Context, req data.AddUserRequest) (*data.User, error)
	UpdateUser(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error)
	GetPriorities(ctx context.Context) (data.GetPrioritiesResponse, error)
	GetStatuses(ctx context.Context) (data.GetStatusesResponse, error)
	GetTemplates(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error)
}

// ReportsAPI — report operations.
type ReportsAPI interface {
	GetReports(ctx context.Context, projectID int64) (data.GetReportsResponse, error)
	GetCrossProjectReports(ctx context.Context) (data.GetReportsResponse, error)
	RunReport(ctx context.Context, templateID int64) (*data.RunReportResponse, error)
	RunCrossProjectReport(ctx context.Context, templateID int64) (*data.RunReportResponse, error)
}

// GroupsAPI — group operations.
type GroupsAPI interface {
	GetGroups(ctx context.Context, projectID int64) (data.GetGroupsResponse, error)
	GetGroup(ctx context.Context, groupID int64) (*data.Group, error)
	AddGroup(ctx context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error)
	UpdateGroup(ctx context.Context, groupID int64, name string, userIDs []int64) (*data.Group, error)
	DeleteGroup(ctx context.Context, groupID int64) error
}

// RolesAPI — role operations.
type RolesAPI interface {
	GetRoles(ctx context.Context) (data.GetRolesResponse, error)
	GetRole(ctx context.Context, roleID int64) (*data.Role, error)
}

// ResultFieldsAPI — result field operations.
type ResultFieldsAPI interface {
	GetResultFields(ctx context.Context) (data.GetResultFieldsResponse, error)
}

// DatasetsAPI — dataset operations.
type DatasetsAPI interface {
	GetDatasets(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error)
	GetDataset(ctx context.Context, datasetID int64) (*data.Dataset, error)
	AddDataset(ctx context.Context, projectID int64, name string) (*data.Dataset, error)
	UpdateDataset(ctx context.Context, datasetID int64, name string) (*data.Dataset, error)
	DeleteDataset(ctx context.Context, datasetID int64) error
}

// VariablesAPI — variable operations.
type VariablesAPI interface {
	GetVariables(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error)
	AddVariable(ctx context.Context, datasetID int64, name string) (*data.Variable, error)
	UpdateVariable(ctx context.Context, variableID int64, name string) (*data.Variable, error)
	DeleteVariable(ctx context.Context, variableID int64) error
}

// BDDsAPI — BDD scenario operations.
type BDDsAPI interface {
	GetBDD(ctx context.Context, caseID int64) (*data.BDD, error)
	AddBDD(ctx context.Context, caseID int64, content string) (*data.BDD, error)
}

// LabelsAPI — label operations.
type LabelsAPI interface {
	GetLabels(ctx context.Context, projectID int64) (data.GetLabelsResponse, error)
	GetLabel(ctx context.Context, labelID int64) (*data.Label, error)
	UpdateLabel(ctx context.Context, labelID int64, req data.UpdateLabelRequest) (*data.Label, error)
	UpdateTestLabels(ctx context.Context, testID int64, labels []string) error
	UpdateTestsLabels(ctx context.Context, runID int64, testIDs []int64, labels []string) error
}

// ExtendedAPI — extended APIs (Groups, Roles, ResultFields, Datasets, Variables, BDDs, Labels).
type ExtendedAPI interface {
	GroupsAPI
	RolesAPI
	ResultFieldsAPI
	DatasetsAPI
	VariablesAPI
	BDDsAPI
	LabelsAPI
}

// ClientInterface is the complete TestRail API client interface.
type ClientInterface interface {
	ProjectsAPI
	CasesAPI
	SuitesAPI
	SectionsAPI
	SharedStepsAPI
	RunsAPI
	ResultsAPI
	TestsAPI
	MilestonesAPI
	PlansAPI
	AttachmentsAPI
	ConfigurationsAPI
	UsersAPI
	ReportsAPI
	ExtendedAPI
}

// Compile-time check: HTTPClient must implement ClientInterface.
var _ ClientInterface = (*HTTPClient)(nil)
