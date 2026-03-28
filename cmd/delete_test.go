package cmd

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupDeleteTest(t *testing.T, mock *client.MockClient) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{
		Use:   deleteCmd.Use,
		Short: deleteCmd.Short,
		Long:  deleteCmd.Long,
		RunE:  runDelete,
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	cmd.Flags().Bool("soft", false, "Мягкое удаление (где поддерживается)")

	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)

	return cmd
}

func TestDelete_Project_WithID_Success(t *testing.T) {
	called := false
	mock := &client.MockClient{
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			assert.Equal(t, int64(77), projectID)
			return nil
		},
	}

	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"project", "77"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDelete_Project_AutoSelectID_Success(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 11, Name: "Project 11"}}, nil
		},
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			assert.Equal(t, int64(11), projectID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{"project"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSelectCaseID(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})
	cases := data.GetCasesResponse{
		{ID: 100, Title: "Case A"},
		{ID: 200, Title: "Case B"},
	}

	id, err := selectCaseID(context.Background(), p, cases)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), id)
}

func TestSelectSharedStepID(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	steps := data.GetSharedStepsResponse{
		{ID: 555, Title: "Step A"},
	}

	id, err := selectSharedStepID(p, steps)
	assert.NoError(t, err)
	assert.Equal(t, int64(555), id)
}

func TestRunDeleteDryRun_SwitchEndpoints(t *testing.T) {
	dr := output.NewDryRunPrinter("delete")

	assert.NoError(t, runDeleteDryRun(dr, "project", 1))
	assert.NoError(t, runDeleteDryRun(dr, "suite", 2))
	assert.NoError(t, runDeleteDryRun(dr, "section", 3))
	assert.NoError(t, runDeleteDryRun(dr, "case", 4))
	assert.NoError(t, runDeleteDryRun(dr, "run", 5))
	assert.NoError(t, runDeleteDryRun(dr, "shared-step", 6))

	err := runDeleteDryRun(dr, "unknown", 1)
	assert.Error(t, err)
}

func TestResolveDeleteID_SuiteAndRun(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 20, Name: "R1"}}, nil
		},
	}
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0})

	suiteID, err := resolveDeleteID(context.Background(), p, mock, "suite")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), suiteID)

	runID, err := resolveDeleteID(context.Background(), p, mock, "run")
	assert.NoError(t, err)
	assert.Equal(t, int64(20), runID)
}

func TestResolveDeleteID_SectionCaseSharedStep(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 30, Name: "SEC1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 40, Title: "CASE1"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{{ID: 50, Title: "STEP1"}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
		)

	sectionID, err := resolveDeleteID(context.Background(), p, mock, "section")
	assert.NoError(t, err)
	assert.Equal(t, int64(30), sectionID)

	caseID, err := resolveDeleteID(context.Background(), p, mock, "case")
	assert.NoError(t, err)
	assert.Equal(t, int64(40), caseID)

	stepID, err := resolveDeleteID(context.Background(), p, mock, "shared-step")
	assert.NoError(t, err)
	assert.Equal(t, int64(50), stepID)
}

func TestDelete_AutoSelectEndpointAndSuite_Success(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 5, Name: "Project 5"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(5), projectID)
			return data.GetSuitesResponse{{ID: 700, Name: "Suite 700"}}, nil
		},
		DeleteSuiteFunc: func(ctx context.Context, suiteID int64) error {
			called = true
			assert.Equal(t, int64(700), suiteID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDelete_NonInteractive_NoArgs_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			return nil
		},
	}

	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

func TestDelete_DryRun_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			return nil
		},
	}

	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"project", "77", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestDelete_InvalidID_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"project", "abc"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ID")
}
