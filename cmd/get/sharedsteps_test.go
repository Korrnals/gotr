package get

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

// ==================== Tests for get sharedsteps ====================

func TestSharedStepsCmd_WithProjectID(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSharedStepsResponse{
				{ID: 1, Title: "Shared Step 1", ProjectID: 30},
				{ID: 2, Title: "Shared Step 2", ProjectID: 30},
			}, nil
		},
	}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepsCmd_WithProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSharedStepsResponse{
				{ID: 1, Title: "Shared Step 1", ProjectID: 30},
			}, nil
		},
	}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepsCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestSharedStepsCmd_InvalidProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSharedStepsCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSharedStepsCmd_NoArgs_Interactive_SelectProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("project lookup failed")
		},
	}

	p := interactive.NewMockPrompter()
	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project lookup failed")
}

// ==================== Tests for get sharedstep ====================

func TestSharedStepCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepFunc: func(ctx context.Context, stepID int64) (*data.SharedStep, error) {
			assert.Equal(t, int64(12345), stepID)
			return &data.SharedStep{
				ID:        12345,
				Title:     "Test Shared Step",
				ProjectID: 30,
			}, nil
		},
	}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepCmd_InvalidStepID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid step ID")
}

func TestSharedStepCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSharedStepCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSharedStepsResponse{{ID: 12345, Title: "Test Shared Step", ProjectID: 30}}, nil
		},
		GetSharedStepFunc: func(ctx context.Context, stepID int64) (*data.SharedStep, error) {
			assert.Equal(t, int64(12345), stepID)
			return &data.SharedStep{ID: 12345, Title: "Test Shared Step", ProjectID: 30}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestSharedStepsCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no projects found")
}

func TestSharedStepCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step_id required")
}

func TestSharedStepCmd_NoArgs_Interactive_EmptySharedSteps(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSharedStepsResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no shared steps found")
}

func TestSharedStepCmd_NoArgs_Interactive_GetSharedStepsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, fmt.Errorf("shared steps unavailable")
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get shared steps")
}

func TestSharedStepCmd_NoArgs_Interactive_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{{ID: 12345, Title: "Step", ProjectID: 30}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select shared step")
}

func TestSharedStepCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepFunc: func(ctx context.Context, stepID int64) (*data.SharedStep, error) {
			return nil, fmt.Errorf("shared step not found")
		},
	}

	cmd := newSharedStepCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSharedStepsCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSharedStepsCmd(nilClientFunc)
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestSharedStepCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSharedStepCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}
