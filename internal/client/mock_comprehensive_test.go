// internal/client/mock_comprehensive_test.go
// Comprehensive mock coverage expansion - focuses on nil paths and error coverage
package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// Test all remaining mock Suite/Section/SharedStep/Run/Result/Test methods
// to hit nil coverage paths that aren't in unit tests

func TestMock_Suites_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	suites, err := m.GetSuites(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, suites)

	suite, err := m.GetSuite(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, suite)

	suite, err = m.AddSuite(ctx, 1, nil)
	assert.NoError(t, err)
	// AddSuite returns a default Suite struct with ID: 999 when no function is set
	assert.NotNil(t, suite)
	assert.Equal(t, int64(999), suite.ID)

	suite, err = m.UpdateSuite(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, suite)

	err = m.DeleteSuite(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Suites_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetSuitesFunc: func(_ context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 1, Name: "s1"}}, nil
		},
		GetSuiteFunc: func(_ context.Context, suiteID int64) (*data.Suite, error) {
			return &data.Suite{ID: suiteID}, nil
		},
		AddSuiteFunc: func(_ context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			return &data.Suite{ID: 99}, nil
		},
		UpdateSuiteFunc: func(_ context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
			return &data.Suite{ID: suiteID}, nil
		},
		DeleteSuiteFunc: func(_ context.Context, suiteID int64) error {
			return errors.New("delete fail")
		},
	}
	ctx := context.Background()

	suites, _ := m.GetSuites(ctx, 1)
	assert.Len(t, suites, 1)

	suite, _ := m.GetSuite(ctx, 1)
	assert.Equal(t, int64(1), suite.ID)

	suite, _ = m.AddSuite(ctx, 1, &data.AddSuiteRequest{})
	assert.Equal(t, int64(99), suite.ID)

	suite, _ = m.UpdateSuite(ctx, 1, &data.UpdateSuiteRequest{})
	assert.Equal(t, int64(1), suite.ID)

	err := m.DeleteSuite(ctx, 1)
	assert.Error(t, err)
}

func TestMock_Sections_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	secs, err := m.GetSections(ctx, 1, 1)
	assert.NoError(t, err)
	assert.Nil(t, secs)

	sec, err := m.GetSection(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, sec)

	sec, err = m.AddSection(ctx, 1, nil)
	assert.NoError(t, err)
	// AddSection returns a default Section struct with ID: 999
	assert.NotNil(t, sec)
	assert.Equal(t, int64(999), sec.ID)

	sec, err = m.UpdateSection(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, sec)

	err = m.DeleteSection(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Sections_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetSectionsFunc: func(_ context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 1}}, nil
		},
		GetSectionFunc: func(_ context.Context, sectionID int64) (*data.Section, error) {
			return &data.Section{ID: sectionID}, nil
		},
		AddSectionFunc: func(_ context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
			return &data.Section{ID: 88}, nil
		},
		UpdateSectionFunc: func(_ context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
			return &data.Section{ID: sectionID}, nil
		},
		DeleteSectionFunc: func(_ context.Context, sectionID int64) error {
			return nil
		},
	}
	ctx := context.Background()

	secs, _ := m.GetSections(ctx, 1, 1)
	assert.Len(t, secs, 1)

	sec, _ := m.GetSection(ctx, 1)
	assert.Equal(t, int64(1), sec.ID)

	sec, _ = m.AddSection(ctx, 1, &data.AddSectionRequest{})
	assert.Equal(t, int64(88), sec.ID)

	sec, _ = m.UpdateSection(ctx, 1, &data.UpdateSectionRequest{})
	assert.Equal(t, int64(1), sec.ID)

	err := m.DeleteSection(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_SharedSteps_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	steps, err := m.GetSharedSteps(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, steps)

	step, err := m.GetSharedStep(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, step)

	step, err = m.AddSharedStep(ctx, 1, nil)
	assert.NoError(t, err)
	// AddSharedStep returns a default SharedStep struct with ID: 999
	assert.NotNil(t, step)
	assert.Equal(t, int64(999), step.ID)

	step, err = m.UpdateSharedStep(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, step)

	err = m.DeleteSharedStep(ctx, 1, 0)
	assert.NoError(t, err)

	hist, err := m.GetSharedStepHistory(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, hist)
}

func TestMock_SharedSteps_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetSharedStepsFunc: func(_ context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{{ID: 1}}, nil
		},
		GetSharedStepFunc: func(_ context.Context, stepID int64) (*data.SharedStep, error) {
			return &data.SharedStep{ID: stepID}, nil
		},
		AddSharedStepFunc: func(_ context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
			return &data.SharedStep{ID: 77}, nil
		},
		UpdateSharedStepFunc: func(_ context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
			return &data.SharedStep{ID: stepID}, nil
		},
		DeleteSharedStepFunc: func(_ context.Context, stepID int64, keepInCases int) error {
			return nil
		},
		GetSharedStepHistoryFunc: func(_ context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error) {
			return &data.GetSharedStepHistoryResponse{}, nil
		},
	}
	ctx := context.Background()

	steps, _ := m.GetSharedSteps(ctx, 1)
	assert.Len(t, steps, 1)

	step, _ := m.GetSharedStep(ctx, 1)
	assert.Equal(t, int64(1), step.ID)

	step, _ = m.AddSharedStep(ctx, 1, &data.AddSharedStepRequest{})
	assert.Equal(t, int64(77), step.ID)

	step, _ = m.UpdateSharedStep(ctx, 1, &data.UpdateSharedStepRequest{})
	assert.Equal(t, int64(1), step.ID)

	err := m.DeleteSharedStep(ctx, 1, 1)
	assert.NoError(t, err)

	hist, _ := m.GetSharedStepHistory(ctx, 1)
	assert.NotNil(t, hist)
}

func TestMock_Runs_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	runs, err := m.GetRuns(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, runs)

	run, err := m.GetRun(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, run)

	run, err = m.AddRun(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, run)

	run, err = m.UpdateRun(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, run)

	run, err = m.CloseRun(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, run)

	err = m.DeleteRun(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Runs_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetRunsFunc: func(_ context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 1, Name: "r1"}}, nil
		},
		GetRunFunc: func(_ context.Context, runID int64) (*data.Run, error) {
			return &data.Run{ID: runID, Name: "r"}, nil
		},
		AddRunFunc: func(_ context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			return &data.Run{ID: 66}, nil
		},
		UpdateRunFunc: func(_ context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			return &data.Run{ID: runID}, nil
		},
		CloseRunFunc: func(_ context.Context, runID int64) (*data.Run, error) {
			return &data.Run{ID: runID, IsCompleted: true}, nil
		},
		DeleteRunFunc: func(_ context.Context, runID int64) error {
			return nil
		},
	}
	ctx := context.Background()

	runs, _ := m.GetRuns(ctx, 1)
	assert.Len(t, runs, 1)

	run, _ := m.GetRun(ctx, 1)
	assert.Equal(t, int64(1), run.ID)

	run, _ = m.AddRun(ctx, 1, &data.AddRunRequest{})
	assert.Equal(t, int64(66), run.ID)

	run, _ = m.UpdateRun(ctx, 1, &data.UpdateRunRequest{})
	assert.Equal(t, int64(1), run.ID)

	run, _ = m.CloseRun(ctx, 1)
	assert.True(t, run.IsCompleted)

	err := m.DeleteRun(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Results_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	results, err := m.GetResults(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, results)

	results, err = m.GetResultsForRun(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, results)

	results, err = m.GetResultsForCase(ctx, 1, 1)
	assert.NoError(t, err)
	assert.Nil(t, results)

	result, err := m.AddResult(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)

	result, err = m.AddResultForCase(ctx, 1, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)

	results, err = m.AddResults(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, results)

	results, err = m.AddResultsForCases(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestMock_Results_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetResultsFunc: func(_ context.Context, testID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 1, StatusID: 1}}, nil
		},
		GetResultsForRunFunc: func(_ context.Context, runID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 1, StatusID: 1}}, nil
		},
		GetResultsForCaseFunc: func(_ context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 1, StatusID: 1}}, nil
		},
		AddResultFunc: func(_ context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			return &data.Result{ID: 55}, nil
		},
		AddResultForCaseFunc: func(_ context.Context, runID, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
			return &data.Result{ID: 56}, nil
		},
		AddResultsFunc: func(_ context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 57}}, nil
		},
		AddResultsForCasesFunc: func(_ context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{{ID: 58}}, nil
		},
	}
	ctx := context.Background()

	results, _ := m.GetResults(ctx, 1)
	assert.Len(t, results, 1)

	results, _ = m.GetResultsForRun(ctx, 1)
	assert.Len(t, results, 1)

	results, _ = m.GetResultsForCase(ctx, 1, 1)
	assert.Len(t, results, 1)

	result, _ := m.AddResult(ctx, 1, &data.AddResultRequest{})
	assert.Equal(t, int64(55), result.ID)

	result, _ = m.AddResultForCase(ctx, 1, 1, &data.AddResultRequest{})
	assert.Equal(t, int64(56), result.ID)

	results, _ = m.AddResults(ctx, 1, &data.AddResultsRequest{})
	assert.Len(t, results, 1)

	results, _ = m.AddResultsForCases(ctx, 1, &data.AddResultsForCasesRequest{})
	assert.Len(t, results, 1)
}

func TestMock_Tests_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	test, err := m.GetTest(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, test)

	tests, err := m.GetTests(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, tests)

	test, err = m.UpdateTest(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, test)
}

func TestMock_Tests_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetTestFunc: func(_ context.Context, testID int64) (*data.Test, error) {
			return &data.Test{ID: testID, Title: "t"}, nil
		},
		GetTestsFunc: func(_ context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{{ID: 1, Title: "t"}}, nil
		},
		UpdateTestFunc: func(_ context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
			return &data.Test{ID: testID}, nil
		},
	}
	ctx := context.Background()

	test, _ := m.GetTest(ctx, 1)
	assert.Equal(t, int64(1), test.ID)

	tests, _ := m.GetTests(ctx, 1, nil)
	assert.Len(t, tests, 1)

	test, _ = m.UpdateTest(ctx, 1, &data.UpdateTestRequest{})
	assert.Equal(t, int64(1), test.ID)
}

func TestMock_Milestones_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	ms, err := m.GetMilestone(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, ms)

	mss, err := m.GetMilestones(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, mss)

	ms, err = m.AddMilestone(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, ms)

	ms, err = m.UpdateMilestone(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, ms)

	err = m.DeleteMilestone(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Milestones_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetMilestoneFunc: func(_ context.Context, milestoneID int64) (*data.Milestone, error) {
			return &data.Milestone{ID: milestoneID, Name: "m"}, nil
		},
		GetMilestonesFunc: func(_ context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{{ID: 1, Name: "m"}}, nil
		},
		AddMilestoneFunc: func(_ context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
			return &data.Milestone{ID: 44}, nil
		},
		UpdateMilestoneFunc: func(_ context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
			return &data.Milestone{ID: milestoneID}, nil
		},
		DeleteMilestoneFunc: func(_ context.Context, milestoneID int64) error {
			return nil
		},
	}
	ctx := context.Background()

	ms, _ := m.GetMilestone(ctx, 1)
	assert.Equal(t, int64(1), ms.ID)

	mss, _ := m.GetMilestones(ctx, 1)
	assert.Len(t, mss, 1)

	ms, _ = m.AddMilestone(ctx, 1, &data.AddMilestoneRequest{})
	assert.Equal(t, int64(44), ms.ID)

	ms, _ = m.UpdateMilestone(ctx, 1, &data.UpdateMilestoneRequest{})
	assert.Equal(t, int64(1), ms.ID)

	err := m.DeleteMilestone(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Plans_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	plan, err := m.GetPlan(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, plan)

	plans, err := m.GetPlans(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, plans)

	plan, err = m.AddPlan(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, plan)

	plan, err = m.UpdatePlan(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, plan)

	plan, err = m.ClosePlan(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, plan)

	err = m.DeletePlan(ctx, 1)
	assert.NoError(t, err)

	plan, err = m.AddPlanEntry(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, plan)

	plan, err = m.UpdatePlanEntry(ctx, 1, "e1", nil)
	assert.NoError(t, err)
	assert.Nil(t, plan)

	err = m.DeletePlanEntry(ctx, 1, "e1")
	assert.NoError(t, err)
}

func TestMock_Plans_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetPlanFunc: func(_ context.Context, planID int64) (*data.Plan, error) {
			return &data.Plan{ID: planID, Name: "p"}, nil
		},
		GetPlansFunc: func(_ context.Context, projectID int64) (data.GetPlansResponse, error) {
			return data.GetPlansResponse{{ID: 1, Name: "p"}}, nil
		},
		AddPlanFunc: func(_ context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
			return &data.Plan{ID: 33}, nil
		},
		UpdatePlanFunc: func(_ context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
			return &data.Plan{ID: planID}, nil
		},
		ClosePlanFunc: func(_ context.Context, planID int64) (*data.Plan, error) {
			return &data.Plan{ID: planID, IsCompleted: true}, nil
		},
		DeletePlanFunc: func(_ context.Context, planID int64) error {
			return nil
		},
		AddPlanEntryFunc: func(_ context.Context, planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
			return &data.Plan{ID: planID}, nil
		},
		UpdatePlanEntryFunc: func(_ context.Context, planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
			return &data.Plan{ID: planID}, nil
		},
		DeletePlanEntryFunc: func(_ context.Context, planID int64, entryID string) error {
			return nil
		},
	}
	ctx := context.Background()

	plan, _ := m.GetPlan(ctx, 1)
	assert.Equal(t, int64(1), plan.ID)

	plans, _ := m.GetPlans(ctx, 1)
	assert.Len(t, plans, 1)

	plan, _ = m.AddPlan(ctx, 1, &data.AddPlanRequest{})
	assert.Equal(t, int64(33), plan.ID)

	plan, _ = m.UpdatePlan(ctx, 1, &data.UpdatePlanRequest{})
	assert.Equal(t, int64(1), plan.ID)

	plan, _ = m.ClosePlan(ctx, 1)
	assert.True(t, plan.IsCompleted)

	err := m.DeletePlan(ctx, 1)
	assert.NoError(t, err)

	plan, _ = m.AddPlanEntry(ctx, 1, &data.AddPlanEntryRequest{})
	assert.Equal(t, int64(1), plan.ID)

	plan, err = m.UpdatePlanEntry(ctx, 1, "e1", &data.UpdatePlanEntryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), plan.ID)

	err = m.DeletePlanEntry(ctx, 1, "e1")
	assert.NoError(t, err)
}

func TestMock_Attachments_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	err := m.DeleteAttachment(ctx, 1)
	assert.NoError(t, err)

	att, err := m.GetAttachment(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, att)

	atts, err := m.GetAttachmentsForCase(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, atts)

	atts, err = m.GetAttachmentsForPlan(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, atts)

	atts, err = m.GetAttachmentsForPlanEntry(ctx, 1, "e1")
	assert.NoError(t, err)
	assert.Nil(t, atts)

	atts, err = m.GetAttachmentsForRun(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, atts)

	atts, err = m.GetAttachmentsForTest(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, atts)
}

func TestMock_Attachments_WithFunctions(t *testing.T) {
	m := &MockClient{
		DeleteAttachmentFunc: func(_ context.Context, attachmentID int64) error {
			return nil
		},
		GetAttachmentFunc: func(_ context.Context, attachmentID int64) (*data.Attachment, error) {
			return &data.Attachment{ID: attachmentID, Name: "a"}, nil
		},
		GetAttachmentsForCaseFunc: func(_ context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{{ID: 1, Name: "a"}}, nil
		},
		GetAttachmentsForPlanFunc: func(_ context.Context, planID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{{ID: 1, Name: "a"}}, nil
		},
		GetAttachmentsForPlanEntryFunc: func(_ context.Context, planID int64, entryID string) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{{ID: 1, Name: "a"}}, nil
		},
		GetAttachmentsForRunFunc: func(_ context.Context, runID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{{ID: 1, Name: "a"}}, nil
		},
		GetAttachmentsForTestFunc: func(_ context.Context, testID int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{{ID: 1, Name: "a"}}, nil
		},
	}
	ctx := context.Background()

	err := m.DeleteAttachment(ctx, 1)
	assert.NoError(t, err)

	att, _ := m.GetAttachment(ctx, 1)
	assert.Equal(t, int64(1), att.ID)

	atts, _ := m.GetAttachmentsForCase(ctx, 1)
	assert.Len(t, atts, 1)

	atts, _ = m.GetAttachmentsForPlan(ctx, 1)
	assert.Len(t, atts, 1)

	atts, _ = m.GetAttachmentsForPlanEntry(ctx, 1, "e1")
	assert.Len(t, atts, 1)

	atts, _ = m.GetAttachmentsForRun(ctx, 1)
	assert.Len(t, atts, 1)

	atts, _ = m.GetAttachmentsForTest(ctx, 1)
	assert.Len(t, atts, 1)
}

func TestMock_Configs_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	configs, err := m.GetConfigs(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, configs)

	cg, err := m.AddConfigGroup(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, cg)

	cfg, err := m.AddConfig(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, cfg)

	cg, err = m.UpdateConfigGroup(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, cg)

	cfg, err = m.UpdateConfig(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, cfg)

	err = m.DeleteConfigGroup(ctx, 1)
	assert.NoError(t, err)

	err = m.DeleteConfig(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Configs_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetConfigsFunc: func(_ context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return data.GetConfigsResponse{{ID: 1, Name: "g"}}, nil
		},
		AddConfigGroupFunc: func(_ context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
			return &data.ConfigGroup{ID: 22}, nil
		},
		AddConfigFunc: func(_ context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
			return &data.Config{ID: 23}, nil
		},
		UpdateConfigGroupFunc: func(_ context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
			return &data.ConfigGroup{ID: groupID}, nil
		},
		UpdateConfigFunc: func(_ context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
			return &data.Config{ID: configID}, nil
		},
		DeleteConfigGroupFunc: func(_ context.Context, groupID int64) error {
			return nil
		},
		DeleteConfigFunc: func(_ context.Context, configID int64) error {
			return nil
		},
	}
	ctx := context.Background()

	configs, _ := m.GetConfigs(ctx, 1)
	assert.NotNil(t, configs)

	cg, _ := m.AddConfigGroup(ctx, 1, &data.AddConfigGroupRequest{})
	assert.Equal(t, int64(22), cg.ID)

	cfg, _ := m.AddConfig(ctx, 1, &data.AddConfigRequest{})
	assert.Equal(t, int64(23), cfg.ID)

	cg, _ = m.UpdateConfigGroup(ctx, 1, &data.UpdateConfigGroupRequest{})
	assert.Equal(t, int64(1), cg.ID)

	cfg, _ = m.UpdateConfig(ctx, 1, &data.UpdateConfigRequest{})
	assert.Equal(t, int64(1), cfg.ID)

	err := m.DeleteConfigGroup(ctx, 1)
	assert.NoError(t, err)

	err = m.DeleteConfig(ctx, 1)
	assert.NoError(t, err)
}

// Test AddAttachmentTo* methods null paths
func TestMock_AttachmentsAdd_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	att, err := m.AddAttachmentToCase(ctx, 1, "")
	assert.NoError(t, err)
	assert.Nil(t, att)

	att, err = m.AddAttachmentToPlan(ctx, 1, "")
	assert.NoError(t, err)
	assert.Nil(t, att)

	att, err = m.AddAttachmentToPlanEntry(ctx, 1, "e", "")
	assert.NoError(t, err)
	assert.Nil(t, att)

	att, err = m.AddAttachmentToResult(ctx, 1, "")
	assert.NoError(t, err)
	assert.Nil(t, att)

	att, err = m.AddAttachmentToRun(ctx, 1, "")
	assert.NoError(t, err)
	assert.Nil(t, att)
}

func TestMock_AttachmentsAdd_WithFunctions(t *testing.T) {
	m := &MockClient{
		AddAttachmentToCaseFunc: func(_ context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
			return &data.AttachmentResponse{AttachmentID: 11}, nil
		},
		AddAttachmentToPlanFunc: func(_ context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
			return &data.AttachmentResponse{AttachmentID: 12}, nil
		},
		AddAttachmentToPlanEntryFunc: func(_ context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
			return &data.AttachmentResponse{AttachmentID: 13}, nil
		},
		AddAttachmentToResultFunc: func(_ context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
			return &data.AttachmentResponse{AttachmentID: 14}, nil
		},
		AddAttachmentToRunFunc: func(_ context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
			return &data.AttachmentResponse{AttachmentID: 15}, nil
		},
	}
	ctx := context.Background()

	att, _ := m.AddAttachmentToCase(ctx, 1, "/tmp/file")
	assert.Equal(t, int64(11), att.AttachmentID)

	att, _ = m.AddAttachmentToPlan(ctx, 1, "/tmp/file")
	assert.Equal(t, int64(12), att.AttachmentID)

	att, _ = m.AddAttachmentToPlanEntry(ctx, 1, "e", "/tmp/file")
	assert.Equal(t, int64(13), att.AttachmentID)

	att, _ = m.AddAttachmentToResult(ctx, 1, "/tmp/file")
	assert.Equal(t, int64(14), att.AttachmentID)

	att, _ = m.AddAttachmentToRun(ctx, 1, "/tmp/file")
	assert.Equal(t, int64(15), att.AttachmentID)
}

func TestMock_Projects_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	projects, err := m.GetProjects(ctx)
	assert.NoError(t, err)
	assert.Nil(t, projects)

	project, err := m.GetProject(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, project)

	project, err = m.AddProject(ctx, nil)
	assert.NoError(t, err)
	assert.Nil(t, project)

	project, err = m.UpdateProject(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Nil(t, project)

	err = m.DeleteProject(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_Projects_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetProjectsFunc: func(_ context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "p1"}}, nil
		},
		GetProjectFunc: func(_ context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "p"}, nil
		},
		AddProjectFunc: func(_ context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: 2, Name: "new"}, nil
		},
		UpdateProjectFunc: func(_ context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "updated"}, nil
		},
		DeleteProjectFunc: func(_ context.Context, projectID int64) error {
			return nil
		},
	}
	ctx := context.Background()

	projects, _ := m.GetProjects(ctx)
	assert.Len(t, projects, 1)

	project, _ := m.GetProject(ctx, 1)
	assert.Equal(t, int64(1), project.ID)

	project, _ = m.AddProject(ctx, &data.AddProjectRequest{})
	assert.Equal(t, int64(2), project.ID)

	project, _ = m.UpdateProject(ctx, 1, &data.UpdateProjectRequest{})
	assert.Equal(t, int64(1), project.ID)

	err := m.DeleteProject(ctx, 1)
	assert.NoError(t, err)
}

func TestMock_UsersAndReports_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	users, err := m.GetUsers(ctx)
	assert.NoError(t, err)
	assert.Nil(t, users)

	users, err = m.GetUsersByProject(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, users)

	user, err := m.GetUser(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, user)

	user, err = m.GetUserByEmail(ctx, "u@example.com")
	assert.NoError(t, err)
	assert.Nil(t, user)

	user, err = m.AddUser(ctx, data.AddUserRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(999), user.ID)

	user, err = m.UpdateUser(ctx, 1, data.UpdateUserRequest{})
	assert.NoError(t, err)
	assert.Nil(t, user)

	priorities, err := m.GetPriorities(ctx)
	assert.NoError(t, err)
	assert.Nil(t, priorities)

	statuses, err := m.GetStatuses(ctx)
	assert.NoError(t, err)
	assert.Nil(t, statuses)

	templates, err := m.GetTemplates(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, templates)

	reports, err := m.GetReports(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, reports)

	reports, err = m.GetCrossProjectReports(ctx)
	assert.NoError(t, err)
	assert.Nil(t, reports)

	runReport, err := m.RunReport(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, runReport)

	runReport, err = m.RunCrossProjectReport(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, runReport)
}

func TestMock_UsersAndReports_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetUsersFunc: func(_ context.Context) (data.GetUsersResponse, error) {
			return data.GetUsersResponse{{ID: 1, Name: "u"}}, nil
		},
		GetUsersByProjectFunc: func(_ context.Context, projectID int64) (data.GetUsersResponse, error) {
			return data.GetUsersResponse{{ID: 2, Name: "up"}}, nil
		},
		GetUserFunc: func(_ context.Context, userID int64) (*data.User, error) {
			return &data.User{ID: userID, Name: "u"}, nil
		},
		GetUserByEmailFunc: func(_ context.Context, email string) (*data.User, error) {
			return &data.User{ID: 3, Email: email}, nil
		},
		AddUserFunc: func(_ context.Context, req data.AddUserRequest) (*data.User, error) {
			return &data.User{ID: 4}, nil
		},
		UpdateUserFunc: func(_ context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
			return &data.User{ID: userID}, nil
		},
		GetPrioritiesFunc: func(_ context.Context) (data.GetPrioritiesResponse, error) {
			return data.GetPrioritiesResponse{{ID: 1, Name: "p"}}, nil
		},
		GetStatusesFunc: func(_ context.Context) (data.GetStatusesResponse, error) {
			return data.GetStatusesResponse{{ID: 1, Name: "s"}}, nil
		},
		GetTemplatesFunc: func(_ context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return data.GetTemplatesResponse{{ID: 1, Name: "t"}}, nil
		},
		GetReportsFunc: func(_ context.Context, projectID int64) (data.GetReportsResponse, error) {
			return data.GetReportsResponse{{ID: 10, Name: "r"}}, nil
		},
		GetCrossProjectReportsFunc: func(_ context.Context) (data.GetReportsResponse, error) {
			return data.GetReportsResponse{{ID: 11, Name: "cr"}}, nil
		},
		RunReportFunc: func(_ context.Context, templateID int64) (*data.RunReportResponse, error) {
			return &data.RunReportResponse{ReportID: 12}, nil
		},
		RunCrossProjectReportFunc: func(_ context.Context, templateID int64) (*data.RunReportResponse, error) {
			return &data.RunReportResponse{ReportID: 13}, nil
		},
	}
	ctx := context.Background()

	users, _ := m.GetUsers(ctx)
	assert.Len(t, users, 1)

	users, _ = m.GetUsersByProject(ctx, 1)
	assert.Len(t, users, 1)

	user, _ := m.GetUser(ctx, 1)
	assert.Equal(t, int64(1), user.ID)

	user, _ = m.GetUserByEmail(ctx, "u@example.com")
	assert.Equal(t, int64(3), user.ID)

	user, _ = m.AddUser(ctx, data.AddUserRequest{})
	assert.Equal(t, int64(4), user.ID)

	user, _ = m.UpdateUser(ctx, 9, data.UpdateUserRequest{})
	assert.Equal(t, int64(9), user.ID)

	priorities, _ := m.GetPriorities(ctx)
	assert.Len(t, priorities, 1)

	statuses, _ := m.GetStatuses(ctx)
	assert.Len(t, statuses, 1)

	templates, _ := m.GetTemplates(ctx, 1)
	assert.Len(t, templates, 1)

	reports, _ := m.GetReports(ctx, 1)
	assert.Len(t, reports, 1)

	reports, _ = m.GetCrossProjectReports(ctx)
	assert.Len(t, reports, 1)

	runReport, _ := m.RunReport(ctx, 1)
	assert.Equal(t, int64(12), runReport.ReportID)

	runReport, _ = m.RunCrossProjectReport(ctx, 1)
	assert.Equal(t, int64(13), runReport.ReportID)
}

func TestMock_ExtendedAPIs_NilPaths(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	groups, err := m.GetGroups(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, groups)

	group, err := m.GetGroup(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, group)

	group, err = m.AddGroup(ctx, 1, "g", []int64{1})
	assert.NoError(t, err)
	assert.Nil(t, group)

	group, err = m.UpdateGroup(ctx, 1, "g", []int64{1})
	assert.NoError(t, err)
	assert.Nil(t, group)

	err = m.DeleteGroup(ctx, 1)
	assert.NoError(t, err)

	roles, err := m.GetRoles(ctx)
	assert.NoError(t, err)
	assert.Nil(t, roles)

	role, err := m.GetRole(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, role)

	resultFields, err := m.GetResultFields(ctx)
	assert.NoError(t, err)
	assert.Nil(t, resultFields)

	datasets, err := m.GetDatasets(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, datasets)

	dataset, err := m.GetDataset(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, dataset)

	dataset, err = m.AddDataset(ctx, 1, "d")
	assert.NoError(t, err)
	assert.Nil(t, dataset)

	dataset, err = m.UpdateDataset(ctx, 1, "d")
	assert.NoError(t, err)
	assert.Nil(t, dataset)

	err = m.DeleteDataset(ctx, 1)
	assert.NoError(t, err)

	vars, err := m.GetVariables(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, vars)

	variable, err := m.AddVariable(ctx, 1, "v")
	assert.NoError(t, err)
	assert.Nil(t, variable)

	variable, err = m.UpdateVariable(ctx, 1, "v")
	assert.NoError(t, err)
	assert.Nil(t, variable)

	err = m.DeleteVariable(ctx, 1)
	assert.NoError(t, err)

	bdd, err := m.GetBDD(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, bdd)

	bdd, err = m.AddBDD(ctx, 1, "content")
	assert.NoError(t, err)
	assert.Nil(t, bdd)

	labels, err := m.GetLabels(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, labels)

	label, err := m.GetLabel(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, label)

	label, err = m.UpdateLabel(ctx, 1, data.UpdateLabelRequest{})
	assert.NoError(t, err)
	assert.Nil(t, label)

	err = m.UpdateTestLabels(ctx, 1, []string{"a"})
	assert.NoError(t, err)

	err = m.UpdateTestsLabels(ctx, 1, []int64{1}, []string{"a"})
	assert.NoError(t, err)
}

func TestMock_ExtendedAPIs_WithFunctions(t *testing.T) {
	m := &MockClient{
		GetGroupsFunc: func(_ context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return data.GetGroupsResponse{{ID: 1, Name: "g"}}, nil
		},
		GetGroupFunc: func(_ context.Context, groupID int64) (*data.Group, error) {
			return &data.Group{ID: groupID, Name: "g"}, nil
		},
		AddGroupFunc: func(_ context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error) {
			return &data.Group{ID: 2, Name: name}, nil
		},
		UpdateGroupFunc: func(_ context.Context, groupID int64, name string, userIDs []int64) (*data.Group, error) {
			return &data.Group{ID: groupID, Name: name}, nil
		},
		DeleteGroupFunc: func(_ context.Context, groupID int64) error {
			return nil
		},
		GetRolesFunc: func(_ context.Context) (data.GetRolesResponse, error) {
			return data.GetRolesResponse{{ID: 1, Name: "r"}}, nil
		},
		GetRoleFunc: func(_ context.Context, roleID int64) (*data.Role, error) {
			return &data.Role{ID: roleID, Name: "r"}, nil
		},
		GetResultFieldsFunc: func(_ context.Context) (data.GetResultFieldsResponse, error) {
			return data.GetResultFieldsResponse{{ID: 1, Name: "f"}}, nil
		},
		GetDatasetsFunc: func(_ context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return data.GetDatasetsResponse{{ID: 1, Name: "d"}}, nil
		},
		GetDatasetFunc: func(_ context.Context, datasetID int64) (*data.Dataset, error) {
			return &data.Dataset{ID: datasetID, Name: "d"}, nil
		},
		AddDatasetFunc: func(_ context.Context, projectID int64, name string) (*data.Dataset, error) {
			return &data.Dataset{ID: 2, Name: name}, nil
		},
		UpdateDatasetFunc: func(_ context.Context, datasetID int64, name string) (*data.Dataset, error) {
			return &data.Dataset{ID: datasetID, Name: name}, nil
		},
		DeleteDatasetFunc: func(_ context.Context, datasetID int64) error {
			return nil
		},
		GetVariablesFunc: func(_ context.Context, datasetID int64) (data.GetVariablesResponse, error) {
			return data.GetVariablesResponse{{ID: 1, Name: "v"}}, nil
		},
		AddVariableFunc: func(_ context.Context, datasetID int64, name string) (*data.Variable, error) {
			return &data.Variable{ID: 2, Name: name}, nil
		},
		UpdateVariableFunc: func(_ context.Context, variableID int64, name string) (*data.Variable, error) {
			return &data.Variable{ID: variableID, Name: name}, nil
		},
		DeleteVariableFunc: func(_ context.Context, variableID int64) error {
			return nil
		},
		GetBDDFunc: func(_ context.Context, caseID int64) (*data.BDD, error) {
			return &data.BDD{ID: caseID}, nil
		},
		AddBDDFunc: func(_ context.Context, caseID int64, content string) (*data.BDD, error) {
			return &data.BDD{ID: caseID, Content: content}, nil
		},
		GetLabelsFunc: func(_ context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return data.GetLabelsResponse{{ID: 1, Name: "l"}}, nil
		},
		GetLabelFunc: func(_ context.Context, labelID int64) (*data.Label, error) {
			return &data.Label{ID: labelID, Name: "l"}, nil
		},
		UpdateLabelFunc: func(_ context.Context, labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
			return &data.Label{ID: labelID, Name: "u"}, nil
		},
		UpdateTestLabelsFunc: func(_ context.Context, testID int64, labels []string) error {
			return nil
		},
		UpdateTestsLabelsFunc: func(_ context.Context, runID int64, testIDs []int64, labels []string) error {
			return nil
		},
	}
	ctx := context.Background()

	groups, _ := m.GetGroups(ctx, 1)
	assert.Len(t, groups, 1)

	group, _ := m.GetGroup(ctx, 1)
	assert.Equal(t, int64(1), group.ID)

	group, _ = m.AddGroup(ctx, 1, "g2", []int64{1})
	assert.Equal(t, int64(2), group.ID)

	group, _ = m.UpdateGroup(ctx, 1, "g3", []int64{1})
	assert.Equal(t, "g3", group.Name)

	err := m.DeleteGroup(ctx, 1)
	assert.NoError(t, err)

	roles, _ := m.GetRoles(ctx)
	assert.Len(t, roles, 1)

	role, _ := m.GetRole(ctx, 1)
	assert.Equal(t, int64(1), role.ID)

	resultFields, _ := m.GetResultFields(ctx)
	assert.Len(t, resultFields, 1)

	datasets, _ := m.GetDatasets(ctx, 1)
	assert.Len(t, datasets, 1)

	dataset, _ := m.GetDataset(ctx, 1)
	assert.Equal(t, int64(1), dataset.ID)

	dataset, _ = m.AddDataset(ctx, 1, "d2")
	assert.Equal(t, "d2", dataset.Name)

	dataset, _ = m.UpdateDataset(ctx, 3, "d3")
	assert.Equal(t, int64(3), dataset.ID)

	err = m.DeleteDataset(ctx, 1)
	assert.NoError(t, err)

	vars, _ := m.GetVariables(ctx, 1)
	assert.Len(t, vars, 1)

	variable, _ := m.AddVariable(ctx, 1, "v2")
	assert.Equal(t, "v2", variable.Name)

	variable, _ = m.UpdateVariable(ctx, 5, "v3")
	assert.Equal(t, int64(5), variable.ID)

	err = m.DeleteVariable(ctx, 1)
	assert.NoError(t, err)

	bdd, _ := m.GetBDD(ctx, 1)
	assert.Equal(t, int64(1), bdd.ID)

	bdd, _ = m.AddBDD(ctx, 1, "Given")
	assert.Equal(t, "Given", bdd.Content)

	labels, _ := m.GetLabels(ctx, 1)
	assert.Len(t, labels, 1)

	label, _ := m.GetLabel(ctx, 1)
	assert.Equal(t, int64(1), label.ID)

	label, _ = m.UpdateLabel(ctx, 1, data.UpdateLabelRequest{})
	assert.Equal(t, "u", label.Name)

	err = m.UpdateTestLabels(ctx, 1, []string{"l1"})
	assert.NoError(t, err)

	err = m.UpdateTestsLabels(ctx, 1, []int64{1, 2}, []string{"l1"})
	assert.NoError(t, err)
}

func TestMock_ParallelCtx_NilAndCustomPaths(t *testing.T) {
	ctx := context.Background()

	m := &MockClient{}
	sections, err := m.GetSectionsParallelCtx(ctx, 1, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, sections)

	m.GetSectionsFunc = func(_ context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
		if suiteID == 1 {
			return data.GetSectionsResponse{{ID: 1}}, nil
		}
		if suiteID == 2 {
			return data.GetSectionsResponse{{ID: 2}}, nil
		}
		return nil, nil
	}
	sections, err = m.GetSectionsParallelCtx(ctx, 1, []int64{1, 2}, nil)
	assert.NoError(t, err)
	assert.Len(t, sections, 2)

	m.GetSectionsParallelCtxFunc = func(_ context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
		return data.GetSectionsResponse{{ID: 99}}, nil
	}
	sections, err = m.GetSectionsParallelCtx(ctx, 1, []int64{1}, &concurrency.ControllerConfig{MaxConcurrentSuites: 2})
	assert.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, int64(99), sections[0].ID)

	m2 := &MockClient{}
	cases, result, err := m2.GetCasesParallelCtx(ctx, 1, []int64{1, 2}, nil)
	assert.NoError(t, err)
	assert.Nil(t, cases)
	assert.NotNil(t, result)

	m2.GetCasesParallelCtxFunc = func(_ context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
		resp := data.GetCasesResponse{{ID: 42}}
		return resp, &concurrency.ExecutionResult{Cases: resp}, nil
	}
	cases, result, err = m2.GetCasesParallelCtx(ctx, 1, []int64{1}, &concurrency.ControllerConfig{MaxConcurrentSuites: 3})
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.NotNil(t, result)
	assert.Len(t, result.Cases, 1)
}

func TestMock_CasesAPITableDriven_NilAndCustomPaths(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T, m *MockClient, custom bool)
	}{
		{
			name: "GetCase",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.GetCaseFunc = func(_ context.Context, caseID int64) (*data.Case, error) {
						return &data.Case{ID: caseID, Title: "custom"}, nil
					}
				}
				got, err := m.GetCase(ctx, 11)
				assert.NoError(t, err)
				if custom {
					assert.NotNil(t, got)
					assert.Equal(t, int64(11), got.ID)
					return
				}
				assert.Nil(t, got)
			},
		},
		{
			name: "AddCase",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.AddCaseFunc = func(_ context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
						return &data.Case{ID: sectionID, Title: req.Title}, nil
					}
				}
				got, err := m.AddCase(ctx, 12, &data.AddCaseRequest{Title: "new"})
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if custom {
					assert.Equal(t, int64(12), got.ID)
					assert.Equal(t, "new", got.Title)
					return
				}
				assert.Equal(t, int64(999), got.ID)
			},
		},
		{
			name: "UpdateCase",
			run: func(t *testing.T, m *MockClient, custom bool) {
				title := "updated"
				if custom {
					m.UpdateCaseFunc = func(_ context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
						if req != nil && req.Title != nil {
							return &data.Case{ID: caseID, Title: *req.Title}, nil
						}
						return &data.Case{ID: caseID}, nil
					}
				}
				got, err := m.UpdateCase(ctx, 13, &data.UpdateCaseRequest{Title: &title})
				assert.NoError(t, err)
				if custom {
					assert.NotNil(t, got)
					assert.Equal(t, int64(13), got.ID)
					assert.Equal(t, "updated", got.Title)
					return
				}
				assert.Nil(t, got)
			},
		},
		{
			name: "DeleteCase",
			run: func(t *testing.T, m *MockClient, custom bool) {
				errExpected := errors.New("delete-case-error")
				if custom {
					m.DeleteCaseFunc = func(_ context.Context, caseID int64) error {
						assert.Equal(t, int64(14), caseID)
						return errExpected
					}
				}
				err := m.DeleteCase(ctx, 14)
				if custom {
					assert.ErrorIs(t, err, errExpected)
					return
				}
				assert.NoError(t, err)
			},
		},
		{
			name: "UpdateCases",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.UpdateCasesFunc = func(_ context.Context, suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
						resp := data.GetCasesResponse{{ID: suiteID}}
						return &resp, nil
					}
				}
				got, err := m.UpdateCases(ctx, 15, &data.UpdateCasesRequest{CaseIDs: []int64{1, 2}, PriorityID: 3})
				assert.NoError(t, err)
				if custom {
					assert.NotNil(t, got)
					assert.Len(t, *got, 1)
					assert.Equal(t, int64(15), (*got)[0].ID)
					return
				}
				assert.Nil(t, got)
			},
		},
		{
			name: "DeleteCases",
			run: func(t *testing.T, m *MockClient, custom bool) {
				errExpected := errors.New("delete-cases-error")
				if custom {
					m.DeleteCasesFunc = func(_ context.Context, suiteID int64, req *data.DeleteCasesRequest) error {
						assert.Equal(t, int64(16), suiteID)
						return errExpected
					}
				}
				err := m.DeleteCases(ctx, 16, &data.DeleteCasesRequest{CaseIDs: []int64{1, 2}})
				if custom {
					assert.ErrorIs(t, err, errExpected)
					return
				}
				assert.NoError(t, err)
			},
		},
		{
			name: "CopyCasesToSection",
			run: func(t *testing.T, m *MockClient, custom bool) {
				errExpected := errors.New("copy-cases-error")
				if custom {
					m.CopyCasesToSectionFunc = func(_ context.Context, sectionID int64, req *data.CopyCasesRequest) error {
						assert.Equal(t, int64(17), sectionID)
						return errExpected
					}
				}
				err := m.CopyCasesToSection(ctx, 17, &data.CopyCasesRequest{CaseIDs: []int64{1}})
				if custom {
					assert.ErrorIs(t, err, errExpected)
					return
				}
				assert.NoError(t, err)
			},
		},
		{
			name: "MoveCasesToSection",
			run: func(t *testing.T, m *MockClient, custom bool) {
				errExpected := errors.New("move-cases-error")
				if custom {
					m.MoveCasesToSectionFunc = func(_ context.Context, sectionID int64, req *data.MoveCasesRequest) error {
						assert.Equal(t, int64(18), sectionID)
						return errExpected
					}
				}
				err := m.MoveCasesToSection(ctx, 18, &data.MoveCasesRequest{CaseIDs: []int64{2}})
				if custom {
					assert.ErrorIs(t, err, errExpected)
					return
				}
				assert.NoError(t, err)
			},
		},
		{
			name: "GetHistoryForCase",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.GetHistoryForCaseFunc = func(_ context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error) {
						return &data.GetHistoryForCaseResponse{}, nil
					}
				}
				got, err := m.GetHistoryForCase(ctx, 19)
				assert.NoError(t, err)
				if custom {
					assert.NotNil(t, got)
					return
				}
				assert.Nil(t, got)
			},
		},
		{
			name: "GetCaseFields",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.GetCaseFieldsFunc = func(_ context.Context) (data.GetCaseFieldsResponse, error) {
						return data.GetCaseFieldsResponse{{ID: 1, Name: "field"}}, nil
					}
				}
				got, err := m.GetCaseFields(ctx)
				assert.NoError(t, err)
				if custom {
					assert.Len(t, got, 1)
					assert.Equal(t, "field", got[0].Name)
					return
				}
				assert.Nil(t, got)
			},
		},
		{
			name: "GetCaseTypes",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.GetCaseTypesFunc = func(_ context.Context) (data.GetCaseTypesResponse, error) {
						return data.GetCaseTypesResponse{{ID: 1, Name: "type"}}, nil
					}
				}
				got, err := m.GetCaseTypes(ctx)
				assert.NoError(t, err)
				if custom {
					assert.Len(t, got, 1)
					assert.Equal(t, "type", got[0].Name)
					return
				}
				assert.Nil(t, got)
			},
		},
		{
			name: "DiffCasesData",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.DiffCasesDataFunc = func(_ context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
						return &data.DiffCasesResponse{}, nil
					}
				}
				got, err := m.DiffCasesData(ctx, 1, 2, "title")
				assert.NoError(t, err)
				if custom {
					assert.NotNil(t, got)
					return
				}
				assert.Nil(t, got)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Run("default", func(t *testing.T) {
				tc.run(t, &MockClient{}, false)
			})
			t.Run("custom", func(t *testing.T) {
				tc.run(t, &MockClient{}, true)
			})
		})
	}
}

func TestMock_UsersReportsAndResultFields_TableDriven_NilAndCustomPaths(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T, m *MockClient, custom bool)
	}{
		{
			name: "GetUsersByProject",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.GetUsersByProjectFunc = func(_ context.Context, projectID int64) (data.GetUsersResponse, error) {
						return data.GetUsersResponse{{ID: 21, Name: "project-user"}}, nil
					}
				}
				users, err := m.GetUsersByProject(ctx, 100)
				assert.NoError(t, err)
				if custom {
					assert.Len(t, users, 1)
					assert.Equal(t, int64(21), users[0].ID)
					return
				}
				assert.Nil(t, users)
			},
		},
		{
			name: "AddUser",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.AddUserFunc = func(_ context.Context, req data.AddUserRequest) (*data.User, error) {
						return &data.User{ID: 22, Name: req.Name, Email: req.Email}, nil
					}
				}
				user, err := m.AddUser(ctx, data.AddUserRequest{Name: "U", Email: "u@example.com"})
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if custom {
					assert.Equal(t, int64(22), user.ID)
					assert.Equal(t, "U", user.Name)
					return
				}
				assert.Equal(t, int64(999), user.ID)
			},
		},
		{
			name: "UpdateUser",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.UpdateUserFunc = func(_ context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
						return &data.User{ID: userID, Name: req.Name}, nil
					}
				}
				user, err := m.UpdateUser(ctx, 23, data.UpdateUserRequest{Name: "updated"})
				assert.NoError(t, err)
				if custom {
					assert.NotNil(t, user)
					assert.Equal(t, int64(23), user.ID)
					return
				}
				assert.Nil(t, user)
			},
		},
		{
			name: "GetCrossProjectReports",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.GetCrossProjectReportsFunc = func(_ context.Context) (data.GetReportsResponse, error) {
						return data.GetReportsResponse{{ID: 24, Name: "cross"}}, nil
					}
				}
				reports, err := m.GetCrossProjectReports(ctx)
				assert.NoError(t, err)
				if custom {
					assert.Len(t, reports, 1)
					assert.Equal(t, int64(24), reports[0].ID)
					return
				}
				assert.Nil(t, reports)
			},
		},
		{
			name: "GetResultFields",
			run: func(t *testing.T, m *MockClient, custom bool) {
				if custom {
					m.GetResultFieldsFunc = func(_ context.Context) (data.GetResultFieldsResponse, error) {
						return data.GetResultFieldsResponse{{ID: 25, Name: "custom_result_field"}}, nil
					}
				}
				fields, err := m.GetResultFields(ctx)
				assert.NoError(t, err)
				if custom {
					assert.Len(t, fields, 1)
					assert.Equal(t, "custom_result_field", fields[0].Name)
					return
				}
				assert.Nil(t, fields)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Run("default", func(t *testing.T) {
				tc.run(t, &MockClient{}, false)
			})
			t.Run("custom", func(t *testing.T) {
				tc.run(t, &MockClient{}, true)
			})
		})
	}
}

func TestMock_GetSectionsParallelCtx_TableDriven_Paths(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "default-empty-suite-ids",
			run: func(t *testing.T) {
				m := &MockClient{}
				sections, err := m.GetSectionsParallelCtx(ctx, 1, nil, nil)
				assert.NoError(t, err)
				assert.Nil(t, sections)
			},
		},
		{
			name: "default-aggregate-success",
			run: func(t *testing.T) {
				m := &MockClient{
					GetSectionsFunc: func(_ context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
						return data.GetSectionsResponse{{ID: suiteID}}, nil
					},
				}
				sections, err := m.GetSectionsParallelCtx(ctx, 1, []int64{1, 2}, nil)
				assert.NoError(t, err)
				assert.Len(t, sections, 2)
			},
		},
		{
			name: "default-aggregate-with-error",
			run: func(t *testing.T) {
				errExpected := errors.New("sections-error")
				m := &MockClient{
					GetSectionsFunc: func(_ context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
						if suiteID == 2 {
							return nil, errExpected
						}
						return data.GetSectionsResponse{{ID: suiteID}}, nil
					},
				}
				sections, err := m.GetSectionsParallelCtx(ctx, 1, []int64{1, 2}, nil)
				assert.ErrorIs(t, err, errExpected)
				assert.Len(t, sections, 1)
				assert.Equal(t, int64(1), sections[0].ID)
			},
		},
		{
			name: "custom-func-path",
			run: func(t *testing.T) {
				m := &MockClient{
					GetSectionsParallelCtxFunc: func(_ context.Context, projectID int64, suiteIDs []int64, config *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
						assert.Equal(t, int64(1), projectID)
						assert.Equal(t, []int64{5}, suiteIDs)
						assert.NotNil(t, config)
						return data.GetSectionsResponse{{ID: 99}}, nil
					},
				}
				sections, err := m.GetSectionsParallelCtx(ctx, 1, []int64{5}, &concurrency.ControllerConfig{MaxConcurrentSuites: 2})
				assert.NoError(t, err)
				assert.Len(t, sections, 1)
				assert.Equal(t, int64(99), sections[0].ID)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, tc.run)
	}
}
