package result

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResultCmds_NilClientAdditional(t *testing.T) {
	nilClient := func(*cobra.Command) client.ClientInterface { return nil }

	t.Run("add-case", func(t *testing.T) {
		cmd := newAddCaseCmd(nilClient)
		cmd.SetArgs([]string{"10", "--case-id", "20", "--status-id", "1"})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP client not initialized")
	})

	t.Run("add-bulk", func(t *testing.T) {
		resultsFile := filepath.Join(t.TempDir(), "results.json")
		require.NoError(t, os.WriteFile(resultsFile, []byte(`[{"test_id":1,"status_id":1}]`), 0o600))

		cmd := newAddBulkCmd(nilClient)
		cmd.SetArgs([]string{"10", "--results-file", resultsFile})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP client not initialized")
	})

	t.Run("get-case", func(t *testing.T) {
		cmd := newGetCaseCmd(nilClient)
		cmd.SetArgs([]string{"10", "20"})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP client not initialized")
	})
}

func TestResultCmds_InteractiveResolveErrorsAdditional(t *testing.T) {
	getProjectsErrMock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, assert.AnError
		},
	}

	t.Run("add no args", func(t *testing.T) {
		cmd := newAddCmd(testhelper.GetClientForTests)
		cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{"--status-id", "1"})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("add-case no args", func(t *testing.T) {
		cmd := newAddCaseCmd(testhelper.GetClientForTests)
		cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{"--case-id", "2", "--status-id", "1"})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("get no args", func(t *testing.T) {
		cmd := newGetCmd(testhelper.GetClientForTests)
		cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("get-case no args", func(t *testing.T) {
		cmd := newGetCaseCmd(testhelper.GetClientForTests)
		cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		assert.Error(t, err)
	})
}

func TestResultCmds_InteractiveSelectionErrorsAfterRunAdditional(t *testing.T) {
	baseMock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 10, Name: "R10"}}, nil
		},
	}

	t.Run("add no tests in run", func(t *testing.T) {
		mock := *baseMock
		mock.GetTestsFunc = func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{}, nil
		}

		p := interactive.NewMockPrompter().
			WithSelectResponses(interactive.SelectResponse{Index: 0}).
			WithSelectResponses(interactive.SelectResponse{Index: 0})

		cmd := newAddCmd(testhelper.GetClientForTests)
		cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, &mock).Context(), p))
		cmd.SetArgs([]string{"--status-id", "1"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tests found")
	})

	t.Run("get no tests in run", func(t *testing.T) {
		mock := *baseMock
		mock.GetTestsFunc = func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{}, nil
		}

		p := interactive.NewMockPrompter().
			WithSelectResponses(interactive.SelectResponse{Index: 0}).
			WithSelectResponses(interactive.SelectResponse{Index: 0})

		cmd := newGetCmd(testhelper.GetClientForTests)
		cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, &mock).Context(), p))
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tests found")
	})

	t.Run("add-case no runs", func(t *testing.T) {
		mock := *baseMock
		mock.GetRunsFunc = func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{}, nil
		}

		p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
		cmd := newAddCaseCmd(testhelper.GetClientForTests)
		cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, &mock).Context(), p))
		cmd.SetArgs([]string{"--case-id", "2", "--status-id", "1"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no test runs found")
	})
}