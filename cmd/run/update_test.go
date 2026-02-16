package run

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Equal(t, "Updated Run Name", *req.Name)
			assert.Equal(t, "Updated description", *req.Description)
			return &data.Run{ID: runID, Name: "Updated Run Name", Description: "Updated description"}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--name", "Updated Run Name", "--description", "Updated description"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_WithMilestoneAndAssignedTo(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Equal(t, int64(100), *req.MilestoneID)
			assert.Equal(t, int64(5), *req.AssignedTo)
			return &data.Run{ID: runID, Name: "Test"}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--milestone-id", "100", "--assigned-to", "5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			return nil, fmt.Errorf("run is completed and cannot be modified")
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be modified")
}

func TestUpdateCmd_NilClient(t *testing.T) {
	// Test when getClient returns nil
	cmd := newUpdateCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"12345", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestUpdateCmd_WithCaseIDs(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Equal(t, []int64{100, 200, 300}, req.CaseIDs)
			return &data.Run{ID: runID}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--case-ids", "100,200,300"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_WithIncludeAll(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			assert.NotNil(t, req.IncludeAll)
			assert.True(t, *req.IncludeAll)
			return &data.Run{ID: runID}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--include-all"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_WithIncludeAllFalse(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(12345), runID)
			assert.NotNil(t, req.IncludeAll)
			assert.False(t, *req.IncludeAll)
			return &data.Run{ID: runID}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--include-all=false"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
