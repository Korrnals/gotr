// internal/client/mock_comprehensive_test.go
// Comprehensive mock coverage expansion - focuses on nil paths and error coverage
package client

import (
	"context"
	"errors"
	"testing"

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
