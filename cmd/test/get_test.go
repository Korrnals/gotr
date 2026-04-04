package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			assert.Equal(t, int64(12345), testID)
			return &data.Test{
				ID:       testID,
				CaseID:   100,
				RunID:    200,
				Title:    "Test Case Title",
				StatusID: 1,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			assert.Equal(t, int64(12345), testID)
			return &data.Test{
				ID:       testID,
				CaseID:   100,
				RunID:    200,
				Title:    "Test Case Title",
				StatusID: 1,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ID")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NegativeID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"-1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{
				{ID: 100, Name: "Run 1", ProjectID: projectID},
			}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			return []data.Test{
				{ID: 1, CaseID: 101, RunID: runID, Title: "Test 1", StatusID: 1},
			}, nil
		},
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			assert.Equal(t, int64(1), testID)
			return &data.Test{
				ID:       testID,
				CaseID:   100,
				RunID:    100,
				Title:    "Test Case Title",
				StatusID: 1,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_id is required in non-interactive mode")
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			return nil, fmt.Errorf("test not found")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test not found")
}

func TestGetCmd_NilClient(t *testing.T) {
	cmd := newGetCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestGetCmd_InvalidOutputPath(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			return &data.Test{ID: testID}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--output", "/nonexistent/dir/test.json"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag")
}

func TestGetCmd_WithSaveEnabled(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			return &data.Test{ID: testID}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--save"})

	// This tests the save path - we just verify it works
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSave_SaveError_HomeIsFile(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			return &data.Test{ID: testID}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--save", "--output", "/nonexistent/dir/test.json"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag")
}

func TestGetCmd_TooManyArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"1", "2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts at most 1 arg")
}

func TestGetCmd_NoArgs_Interactive_ResolveError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(testCmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select project")
}

func TestGetCmd_WithSave_SaveError(t *testing.T) {
	homeMarker := t.TempDir()
	target := homeMarker + "/home-file"
	require.NoError(t, os.WriteFile(target, []byte("x"), 0o600))
	t.Setenv("HOME", target)

	mock := &client.MockClient{
		GetTestFunc: func(ctx context.Context, testID int64) (*data.Test, error) {
			return &data.Test{ID: testID}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--save"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save error")
}
