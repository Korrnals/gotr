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

func TestDeleteCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteRunFunc: func(runID int64) error {
			assert.Equal(t, int64(12345), runID)
			return nil
		},
	}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestDeleteCmd_InvalidRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_ZeroRunID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		DeleteRunFunc: func(runID int64) error {
			return fmt.Errorf("run not found")
		},
	}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newDeleteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_NilClient(t *testing.T) {
	// Test when getClient returns nil
	cmd := newDeleteCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestNewRunServiceFromInterface_WithMockClient(t *testing.T) {
	// Test the else branch where mock client is used
	mock := &client.MockClient{}
	wrapper := newRunServiceFromInterface(mock)
	
	assert.NotNil(t, wrapper)
	assert.NotNil(t, wrapper.svc)
}

func TestRunServiceWrapper_Methods(t *testing.T) {
	mock := &client.MockClient{
		DeleteRunFunc: func(runID int64) error {
			return nil
		},
		GetRunFunc: func(runID int64) (*data.Run, error) {
			return &data.Run{ID: runID, Name: "Test"}, nil
		},
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 1, Name: "Run 1"}}, nil
		},
		AddRunFunc: func(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			return &data.Run{ID: 1, Name: req.Name}, nil
		},
		UpdateRunFunc: func(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			return &data.Run{ID: runID}, nil
		},
		CloseRunFunc: func(runID int64) (*data.Run, error) {
			return &data.Run{ID: runID, IsCompleted: true}, nil
		},
	}
	
	wrapper := newRunServiceFromInterface(mock)
	
	// Test Delete
	err := wrapper.Delete(12345)
	assert.NoError(t, err)
	
	// Test ParseID
	id, err := wrapper.ParseID([]string{"123"}, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
	
	// Test PrintSuccess - just verify it doesn't panic
	cmd := &cobra.Command{}
	wrapper.PrintSuccess(cmd, "Test message")
	
	// Test Create
	req := &data.AddRunRequest{Name: "Test Run", SuiteID: 100}
	run, err := wrapper.Create(30, req)
	assert.NoError(t, err)
	assert.NotNil(t, run)
	
	// Test Output
	err = wrapper.Output(cmd, run)
	assert.NoError(t, err)
	
	// Test Close
	closedRun, err := wrapper.Close(12345)
	assert.NoError(t, err)
	assert.NotNil(t, closedRun)
	assert.True(t, closedRun.IsCompleted)
	
	// Test Update
	updateReq := &data.UpdateRunRequest{}
	updatedRun, err := wrapper.Update(12345, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, updatedRun)
	
	// Test Get
	gotRun, err := wrapper.Get(12345)
	assert.NoError(t, err)
	assert.NotNil(t, gotRun)
	
	// Test GetByProject
	runs, err := wrapper.GetByProject(30)
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
}
