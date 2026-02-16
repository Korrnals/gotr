package tests

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			assert.Empty(t, filters)
			return []data.Test{
				{ID: 1, CaseID: 101, RunID: runID, Title: "Test 1", StatusID: 1},
				{ID: 2, CaseID: 102, RunID: runID, Title: "Test 2", StatusID: 5},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_WithStatusFilter(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(runID int64, filters map[string]string) ([]data.Test, error) {
			assert.Equal(t, int64(100), runID)
			assert.Equal(t, "5", filters["status_id"])
			return []data.Test{
				{ID: 2, CaseID: 102, RunID: runID, Title: "Failed Test", StatusID: 5},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100", "--status-id", "5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "должен быть положительным числом")
}

func TestListCmd_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(runID int64, filters map[string]string) ([]data.Test, error) {
			return nil, fmt.Errorf("ран не найден")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ран не найден")
}

func TestListCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{
				{ID: 1, CaseID: 101, RunID: runID, Title: "Test 1", StatusID: 1},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetTestsFunc: func(runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
