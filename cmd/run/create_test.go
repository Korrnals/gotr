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

func TestCreateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, "Smoke Tests", req.Name)
			assert.Equal(t, int64(20069), req.SuiteID)
			return &data.Run{ID: 123, Name: req.Name}, nil
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Smoke Tests"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCreateCmd_WithDescription(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, "Regression", req.Name)
			assert.Equal(t, "Full regression suite", req.Description)
			return &data.Run{ID: 124}, nil
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Regression", "--description", "Full regression suite"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCreateCmd_WithAssignedTo(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(5), req.AssignedTo)
			return &data.Run{ID: 125}, nil
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Test", "--assigned-to", "5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCreateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Test", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCreateCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--suite-id", "20069", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCreateCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "--suite-id", "20069", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCreateCmd_NilClient(t *testing.T) {
	// Test when getClient returns nil
	cmd := newCreateCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestCreateCmd_WithCaseIDs(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, "Test with cases", req.Name)
			assert.Equal(t, []int64{123, 456, 789}, req.CaseIDs)
			return &data.Run{ID: 126, Name: req.Name}, nil
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Test with cases", "--case-ids", "123,456,789"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCreateCmd_WithConfigIDs(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, []int64{1, 2}, req.ConfigIDs)
			return &data.Run{ID: 127}, nil
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Test", "--config-ids", "1,2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCreateCmd_WithMilestoneID(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(100), req.MilestoneID)
			return &data.Run{ID: 128}, nil
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Test", "--milestone-id", "100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCreateCmd_WithIncludeAllFalse(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(30), projectID)
			assert.False(t, req.IncludeAll)
			return &data.Run{ID: 129}, nil
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--name", "Test", "--include-all=false"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
