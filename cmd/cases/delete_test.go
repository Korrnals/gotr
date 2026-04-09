package cases

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Dry Run Tests ====================

func TestDeleteCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Functional Tests with Mock ====================

func TestDeleteCmd_Success(t *testing.T) {
	deleteCalled := false
	mock := &client.MockClient{
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			assert.Equal(t, int64(12345), caseID)
			deleteCalled = true
			return nil
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, deleteCalled, "DeleteCase should have been called")
}

func TestDeleteCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			return fmt.Errorf("cannot delete: case has associated results")
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete")
}

// ==================== Validation Tests ====================

func TestDeleteCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_NoArgs_Interactive(t *testing.T) {
	deleteCalled := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 10", ProjectID: 1}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, int64(10), suiteID)
			return data.GetCasesResponse{{ID: 12345, Title: "Case 12345"}}, nil
		},
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			assert.Equal(t, int64(12345), caseID)
			deleteCalled = true
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, deleteCalled)
}

func TestDeleteCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	deleteCalled := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			deleteCalled = true
			return nil
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, deleteCalled)
}
