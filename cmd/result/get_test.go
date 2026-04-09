package result

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for result get ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), testID)
			return data.GetResultsResponse{
				{ID: 1, TestID: testID, StatusID: 1, Comment: "Test passed"},
				{ID: 2, TestID: testID, StatusID: 5, Comment: "Test failed"},
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_InvalidTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test")
}

func TestGetCmd_MissingTestID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("test not found")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetCmd_EmptyResults(t *testing.T) {
	mock := &client.MockClient{
		GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			return data.GetResultsResponse{}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NilClient(t *testing.T) {
	cmd := newGetCmd(func(cmd *cobra.Command) client.ClientInterface { return nil })
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client")
}

// ==================== Tests for result get-case ====================

func TestGetCaseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForCaseFunc: func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(100), runID)
			assert.Equal(t, int64(200), caseID)
			return data.GetResultsResponse{
				{ID: 1, TestID: 1, StatusID: 1, Comment: "Passed"},
			}, nil
		},
	}

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "200"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCaseCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "200"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run")
}

func TestGetCaseCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case")
}

func TestGetCaseCmd_MissingArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCaseCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForCaseFunc: func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("case not found in run")
		},
	}

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetCaseCmd_ZeroIDs(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForCaseFunc: func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("invalid id")
		},
	}

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 100, Name: "Run 100"}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{{ID: 12345, CaseID: 200, RunID: runID, Title: "Test 12345"}}, nil
		},
		GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), testID)
			return data.GetResultsResponse{{ID: 1, TestID: testID, StatusID: 1}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCaseCmd_MissingCaseID_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			return []data.Test{{ID: 12345, CaseID: 200, RunID: runID, Title: "Case 200"}}, nil
		},
		GetResultsForCaseFunc: func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(100), runID)
			assert.Equal(t, int64(200), caseID)
			return data.GetResultsResponse{{ID: 1, TestID: 12345, StatusID: 1}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCaseCmd_MissingCaseID_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{{ID: 12345, CaseID: 200, RunID: runID, Title: "Case 200"}}, nil
		},
	}
	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCaseCmd_NoArgs_Interactive_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetRunsResponse{{ID: 100, Name: "Run 100", ProjectID: 30}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			return []data.Test{{ID: 12345, CaseID: 200, RunID: runID, Title: "Case 200"}}, nil
		},
		GetResultsForCaseFunc: func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(100), runID)
			assert.Equal(t, int64(200), caseID)
			return data.GetResultsResponse{{ID: 1, TestID: 12345, StatusID: 1}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}). // project
		WithSelectResponses(interactive.SelectResponse{Index: 0}). // run
		WithSelectResponses(interactive.SelectResponse{Index: 0})  // case

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestResolveResultRunID_GetRunsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, fmt.Errorf("runs unavailable")
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	runID, err := resolveResultRunID(ctx, mock)
	assert.Error(t, err)
	assert.Equal(t, int64(0), runID)
	assert.Contains(t, err.Error(), "failed to get runs list")
}

func TestResolveResultRunID_NoRuns(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	runID, err := resolveResultRunID(ctx, mock)
	assert.Error(t, err)
	assert.Equal(t, int64(0), runID)
	assert.Contains(t, err.Error(), "no test runs found")
}

func TestResolveResultRunID_SelectRunError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 100, Name: "Run 100", ProjectID: 30}}, nil
		},
	}

	// One selection is consumed by SelectProject, second SelectRun fails.
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	runID, err := resolveResultRunID(ctx, mock)
	assert.Error(t, err)
	assert.Equal(t, int64(0), runID)
	assert.Contains(t, err.Error(), "select")
}

func TestSelectTestIDForRun_GetTestsError(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return nil, fmt.Errorf("tests unavailable")
		},
	}

	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	testID, err := selectTestIDForRun(ctx, mock, 100)
	assert.Error(t, err)
	assert.Equal(t, int64(0), testID)
	assert.Contains(t, err.Error(), "failed to get tests")
}

func TestSelectTestIDForRun_NoTests(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{}, nil
		},
	}

	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	testID, err := selectTestIDForRun(ctx, mock, 100)
	assert.Error(t, err)
	assert.Equal(t, int64(0), testID)
	assert.Contains(t, err.Error(), "no tests found")
}

func TestSelectTestIDForRun_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{{ID: 12345, CaseID: 200, RunID: runID, Title: "Test 12345"}}, nil
		},
	}

	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	testID, err := selectTestIDForRun(ctx, mock, 100)
	assert.Error(t, err)
	assert.Equal(t, int64(0), testID)
	assert.Contains(t, err.Error(), "failed to select test")
}

func TestSelectCaseIDForRun_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{{ID: 12345, CaseID: 200, RunID: runID, Title: "Case 200"}}, nil
		},
	}

	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	caseID, err := selectCaseIDForRun(ctx, mock, 100)
	assert.Error(t, err)
	assert.Equal(t, int64(0), caseID)
	assert.Contains(t, err.Error(), "failed to select case")
}

func TestSelectCaseIDForRun_GetTestsError(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return nil, fmt.Errorf("tests unavailable")
		},
	}

	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	caseID, err := selectCaseIDForRun(ctx, mock, 100)
	assert.Error(t, err)
	assert.Equal(t, int64(0), caseID)
	assert.Contains(t, err.Error(), "failed to get tests")
}

func TestSelectCaseIDForRun_NoTests(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{}, nil
		},
	}

	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	caseID, err := selectCaseIDForRun(ctx, mock, 100)
	assert.Error(t, err)
	assert.Equal(t, int64(0), caseID)
	assert.Contains(t, err.Error(), "no tests found")
}

func TestSelectCaseIDForRun_SuccessDedupAndSorted(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{
				{ID: 3001, CaseID: 300, RunID: runID, Title: "Third"},
				{ID: 1001, CaseID: 100, RunID: runID, Title: "First"},
				{ID: 1002, CaseID: 100, RunID: runID, Title: "First duplicate"},
				{ID: 2001, CaseID: 200, RunID: runID, Title: "Second"},
			}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})
	ctx := interactive.WithPrompter(context.Background(), p)
	caseID, err := selectCaseIDForRun(ctx, mock, 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), caseID)
}
