// internal/client/interfaces.go
// Полные интерфейсы для HTTP клиента TestRail API
package client

import (
	"github.com/Korrnals/gotr/internal/models/data"
)

// ProjectsAPI — операции с проектами
type ProjectsAPI interface {
	GetProjects() (data.GetProjectsResponse, error)
	GetProject(projectID int64) (*data.GetProjectResponse, error)
	AddProject(req *data.AddProjectRequest) (*data.GetProjectResponse, error)
	UpdateProject(projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error)
	DeleteProject(projectID int64) error
}

// CasesAPI — операции с кейсами
type CasesAPI interface {
	GetCases(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error)
	GetCase(caseID int64) (*data.Case, error)
	AddCase(sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	UpdateCase(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	DeleteCase(caseID int64) error
	UpdateCases(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error)
	DeleteCases(suiteID int64, req *data.DeleteCasesRequest) error
	CopyCasesToSection(sectionID int64, req *data.CopyCasesRequest) error
	MoveCasesToSection(sectionID int64, req *data.MoveCasesRequest) error
	GetHistoryForCase(caseID int64) (*data.GetHistoryForCaseResponse, error)
	GetCaseFields() (data.GetCaseFieldsResponse, error)
	AddCaseField(req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error)
	GetCaseTypes() (data.GetCaseTypesResponse, error)
	DiffCasesData(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error)
}

// SuitesAPI — операции с сьютами
type SuitesAPI interface {
	GetSuites(projectID int64) (data.GetSuitesResponse, error)
	GetSuite(suiteID int64) (*data.Suite, error)
	AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	UpdateSuite(suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error)
	DeleteSuite(suiteID int64) error
}

// SectionsAPI — операции с секциями
type SectionsAPI interface {
	GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error)
	GetSection(sectionID int64) (*data.Section, error)
	AddSection(projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	UpdateSection(sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error)
	DeleteSection(sectionID int64) error
}

// SharedStepsAPI — операции с shared steps
type SharedStepsAPI interface {
	GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error)
	GetSharedStep(stepID int64) (*data.SharedStep, error)
	AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	UpdateSharedStep(stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error)
	DeleteSharedStep(stepID int64, keepInCases int) error
	GetSharedStepHistory(stepID int64) (*data.GetSharedStepHistoryResponse, error)
}

// RunsAPI — операции с test runs
type RunsAPI interface {
	GetRuns(projectID int64) (data.GetRunsResponse, error)
	GetRun(runID int64) (*data.Run, error)
	AddRun(projectID int64, req *data.AddRunRequest) (*data.Run, error)
	UpdateRun(runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	CloseRun(runID int64) (*data.Run, error)
	DeleteRun(runID int64) error
}

// ResultsAPI — операции с результатами тестов
type ResultsAPI interface {
	GetResults(testID int64) (data.GetResultsResponse, error)
	GetResultsForRun(runID int64) (data.GetResultsResponse, error)
	GetResultsForCase(runID, caseID int64) (data.GetResultsResponse, error)
	AddResult(testID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResultForCase(runID, caseID int64, req *data.AddResultRequest) (*data.Result, error)
	AddResults(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error)
	AddResultsForCases(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error)
}

// TestsAPI — операции с тестами
type TestsAPI interface {
	GetTest(testID int64) (*data.Test, error)
	GetTests(runID int64, filters map[string]string) ([]data.Test, error)
	UpdateTest(testID int64, req *data.UpdateTestRequest) (*data.Test, error)
}

// MilestonesAPI — операции с milestones
type MilestonesAPI interface {
	GetMilestone(milestoneID int64) (*data.Milestone, error)
	GetMilestones(projectID int64) ([]data.Milestone, error)
	AddMilestone(projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error)
	UpdateMilestone(milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error)
	DeleteMilestone(milestoneID int64) error
}

// PlansAPI — операции с тест-планами
type PlansAPI interface {
	GetPlan(planID int64) (*data.Plan, error)
	GetPlans(projectID int64) (data.GetPlansResponse, error)
	AddPlan(projectID int64, req *data.AddPlanRequest) (*data.Plan, error)
	UpdatePlan(planID int64, req *data.UpdatePlanRequest) (*data.Plan, error)
	ClosePlan(planID int64) (*data.Plan, error)
	DeletePlan(planID int64) error
	AddPlanEntry(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error)
	UpdatePlanEntry(planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error)
	DeletePlanEntry(planID int64, entryID string) error
}

// ClientInterface — полный интерфейс клиента TestRail API
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
}

// Проверка, что HTTPClient реализует ClientInterface
var _ ClientInterface = (*HTTPClient)(nil)
