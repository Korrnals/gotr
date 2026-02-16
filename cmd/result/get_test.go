package result

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для result get ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetResultsFunc: func(testID int64) (data.GetResultsResponse, error) {
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
		GetResultsFunc: func(testID int64) (data.GetResultsResponse, error) {
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
		GetResultsFunc: func(testID int64) (data.GetResultsResponse, error) {
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
	assert.Contains(t, err.Error(), "клиент")
}

// ==================== Тесты для result get-case ====================

func TestGetCaseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForCaseFunc: func(runID, caseID int64) (data.GetResultsResponse, error) {
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
		GetResultsForCaseFunc: func(runID, caseID int64) (data.GetResultsResponse, error) {
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
		GetResultsForCaseFunc: func(runID, caseID int64) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("invalid id")
		},
	}

	cmd := newGetCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

