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

// ==================== Functional Tests with Mock ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.Case{ID: 12345, Title: "Test Case", SectionID: 100}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	// The newGetCmd doesn't add save flag, so we need to add it for outputResult
	cmd.Flags().Bool("save", false, "")

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Validation Tests ====================

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs_Interactive(t *testing.T) {
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
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.Case{ID: 12345, Title: "Case 12345"}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newGetCmd(getClientForTests)
	base := setupTestCmd(t, mock)
	base.Flags().Bool("save", false, "")
	cmd.SetContext(interactive.WithPrompter(base.Context(), p))
	cmd.SetArgs([]string{})
	cmd.Flags().Bool("save", false, "")

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}
