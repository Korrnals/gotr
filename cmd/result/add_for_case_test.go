package result

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddForCaseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddResultForCaseFunc: func(runID int64, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Equal(t, int64(678), caseID)
			assert.Equal(t, int64(1), req.StatusID)
			assert.Equal(t, "Test passed", req.Comment)
			return &data.Result{ID: 1, TestID: 100, StatusID: 1}, nil
		},
	}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--case-id", "678", "--status-id", "1", "--comment", "Test passed"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddForCaseCmd_MissingStatusID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--case-id", "678"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddForCaseCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--case-id", "678", "--status-id", "1", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddForCaseCmd_WithAllFlags(t *testing.T) {
	mock := &client.MockClient{
		AddResultForCaseFunc: func(runID int64, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Equal(t, int64(678), caseID)
			assert.Equal(t, int64(5), req.StatusID)
			assert.Equal(t, "Found bug", req.Comment)
			assert.Equal(t, "v1.2.3", req.Version)
			assert.Equal(t, "1m30s", req.Elapsed)
			assert.Equal(t, "BUG-123", req.Defects)
			assert.Equal(t, int64(10), req.AssignedTo)
			return &data.Result{ID: 1}, nil
		},
	}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{
		"12345",
		"--case-id", "678",
		"--status-id", "5",
		"--comment", "Found bug",
		"--version", "v1.2.3",
		"--elapsed", "1m30s",
		"--defects", "BUG-123",
		"--assigned-to", "10",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddForCaseCmd_InvalidIDs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--case-id", "678", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddForCaseCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		AddResultForCaseFunc: func(runID int64, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
			return nil, fmt.Errorf("run is closed")
		},
	}

	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--case-id", "678", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}
