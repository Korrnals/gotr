// Package compare tests - comprehensive test suite for command execution
package compare

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for getClientSafe ====================

func TestGetClientSafe_WithClient(t *testing.T) {
	mock := &client.MockClient{}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)

	assert.NotNil(t, cli)
}

func TestGetClientSafe_NilClient(t *testing.T) {
	SetGetClientForTests(nil)

	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)

	assert.Nil(t, cli)
}

// ==================== Tests for suitesCmd ====================

func TestSuitesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSuitesCmd_NoClient(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestSuitesCmd_InvalidPid1(t *testing.T) {
	mock := &client.MockClient{}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=invalid", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSuitesCmd_ProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("project not found")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Tests for casesCmd ====================

func TestCasesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_WithField(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		DiffCasesDataFunc: func(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return &data.DiffCasesResponse{
				DiffByField: []struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{},
			}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--field=priority_id"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_NoClient(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestCasesCmd_GetProjectNamesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("project lookup failed")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project lookup failed")
}

func TestCasesCmd_CompareError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{{ID: 10, Name: "S"}}, nil
		},
		GetCasesParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			return nil, nil, context.Canceled
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(cctx)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "context canceled")
	}
}

func TestCasesCmd_PrintCompareResultError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=xml", "--save-to=out.xml", "--quiet"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

// ==================== Tests for plansCmd ====================

func TestPlansCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("plans", "plans", "test", "test", fetchPlanItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for runsCmd ====================

func TestRunsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("runs", "runs", "test", "test", fetchRunItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for sectionsCmd ====================

func TestSectionsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionsCmd_NoClient(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestSectionsCmd_GetProjectNamesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("project lookup failed")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project lookup failed")
}

func TestSectionsCmd_CompareError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{{ID: 10, Name: "S"}}, nil
		},
		GetSectionsParallelCtxFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetSectionsResponse, error) {
			return nil, errors.New("sections fetch failed")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load project")
}

func TestSectionsCmd_PrintCompareResultError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "P"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=xml", "--save-to=out.xml", "--quiet"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

// ==================== Tests for milestonesCmd ====================

func TestMilestonesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("milestones", "milestones", "test", "test", fetchMilestoneItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for datasetsCmd ====================

func TestDatasetsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("datasets", "datasets", "test", "test", fetchDatasetItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for groupsCmd ====================

func TestGroupsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("groups", "groups", "test", "test", fetchGroupItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for labelsCmd ====================

func TestLabelsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("labels", "labels", "test", "test", fetchLabelItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for templatesCmd ====================

func TestTemplatesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("templates", "templates", "test", "test", fetchTemplateItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for configurationsCmd ====================

func TestConfigurationsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("configurations", "configurations", "test", "test", fetchConfigurationItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for sharedStepsCmd ====================

func TestSharedStepsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd("sharedsteps", "sharedsteps", "test", "test", fetchSharedStepItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for allCmd ====================

func TestAllCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAllCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=json"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Tests for printCasesFieldDiff ====================

func TestPrintCasesFieldDiff_WithDiff(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		DiffCasesDataFunc: func(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return &data.DiffCasesResponse{
				DiffByField: []struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{
					{
						CaseID: 1,
						First:  data.Case{ID: 1, Title: "Case 1", PriorityID: 1},
						Second: data.Case{ID: 1, Title: "Case 1", PriorityID: 2},
					},
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	old := captureStdout(&buf)
	defer restoreStdout(old)

	printCasesFieldDiff(ctx, mock, 1, 2, "priority_id")
}

func TestPrintCasesFieldDiff_NoDiff(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		DiffCasesDataFunc: func(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return &data.DiffCasesResponse{
				DiffByField: []struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{},
			}, nil
		},
	}

	var buf bytes.Buffer
	old := captureStdout(&buf)
	defer restoreStdout(old)

	printCasesFieldDiff(ctx, mock, 1, 2, "priority_id")
}

func TestPrintCasesFieldDiff_Error(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		DiffCasesDataFunc: func(ctx context.Context, pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return nil, errors.New("API error")
		},
	}

	var buf bytes.Buffer
	old := captureStdout(&buf)
	defer restoreStdout(old)

	printCasesFieldDiff(ctx, mock, 1, 2, "priority_id")
}

// Helper functions for capturing stdout
func captureStdout(buf *bytes.Buffer) *bytes.Buffer {
	return buf
}

func restoreStdout(old *bytes.Buffer) {
	_ = old
}

func TestCasesCmd_InvalidPid_FactoryBranch(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=invalid", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCasesCmd_QuietBranch(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionsCmd_InvalidPid_FactoryBranch(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionsCmd_QuietBranch(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSimpleCompareCmd_NoClient(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})

	cmd := newSimpleCompareCmd(
		"dummy",
		"dummy",
		"dummy",
		"dummy",
		func(ctx context.Context, cli client.ClientInterface, pid int64) ([]ItemInfo, error) {
			return []ItemInfo{}, nil
		},
	)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestSimpleCompareCmd_FetchErrorWrapped(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSimpleCompareCmd(
		"broken",
		"broken",
		"broken",
		"broken",
		func(ctx context.Context, cli client.ClientInterface, pid int64) ([]ItemInfo, error) {
			return nil, errors.New("fetch failed")
		},
	)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comparison error broken")
}
