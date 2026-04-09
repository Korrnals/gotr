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

// ==================== Tests for get case-history ====================

func TestCaseHistoryCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetHistoryForCaseFunc: func(ctx context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.GetHistoryForCaseResponse{
				History: []struct {
					ID        int64         "json:\"id\""
					TypeID    int64         "json:\"type_id\""
					CreatedOn int64         "json:\"created_on\""
					UserID    int64         "json:\"user_id\""
					Changes   []data.Change "json:\"changes\""
				}{
					{
						ID:        1,
						TypeID:    1,
						CreatedOn: 1234567890,
						UserID:    5,
						Changes: []data.Change{
							{Field: "title", OldText: "Old Title", NewText: "New Title"},
						},
					},
				},
			}, nil
		},
	}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseHistoryCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid case ID")
}

func TestCaseHistoryCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCaseHistoryCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite", ProjectID: 30}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			return data.GetCasesResponse{{ID: 12345, Title: "Case 12345"}}, nil
		},
		GetHistoryForCaseFunc: func(ctx context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.GetHistoryForCaseResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseHistoryCmd_NoArgs_Interactive_GetSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, fmt.Errorf("failed to get suites")
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get suites")
}

func TestCaseHistoryCmd_NoArgs_Interactive_NoSuites(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no suites found")
}

func TestCaseHistoryCmd_NoArgs_Interactive_SelectSuiteError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}, {ID: 20070, Name: "Other Suite"}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select")
}

func TestCaseHistoryCmd_NoArgs_Interactive_GetCasesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return nil, fmt.Errorf("failed to get cases")
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cases")
}

func TestCaseHistoryCmd_NoArgs_Interactive_NoCases(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cases found")
}

func TestCaseHistoryCmd_NoArgs_Interactive_SelectCaseError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 12345, Title: "Case 12345"}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select case")
}

func TestCaseHistoryCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestCaseHistoryCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetHistoryForCaseFunc: func(ctx context.Context, caseID int64) (*data.GetHistoryForCaseResponse, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Tests for get sharedstep-history ====================

func TestSharedStepHistoryCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepHistoryFunc: func(ctx context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error) {
			assert.Equal(t, int64(56789), stepID)
			return &data.GetSharedStepHistoryResponse{
				History: []struct {
					ID                   int64       "json:\"id\""
					Timestamp            int64       "json:\"timestamp\""
					UserID               int64       "json:\"user_id\""
					CustomStepsSeparated []data.Step "json:\"custom_steps_separated,omitempty\""
					Title                string      "json:\"title,omitempty\""
				}{
					{
						ID:        1,
						Timestamp: 1234567890,
						UserID:    5,
						Title:     "Updated Step Title",
					},
				},
			}, nil
		},
	}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"56789"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepHistoryCmd_InvalidStepID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid step ID")
}

func TestSharedStepHistoryCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSharedStepHistoryCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSharedStepsResponse{{ID: 56789, Title: "Shared Step 56789", ProjectID: 30}}, nil
		},
		GetSharedStepHistoryFunc: func(ctx context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error) {
			assert.Equal(t, int64(56789), stepID)
			return &data.GetSharedStepHistoryResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepHistoryCmd_NoArgs_Interactive_GetStepsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, fmt.Errorf("failed to get shared steps")
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get shared steps")
}

func TestSharedStepHistoryCmd_NoArgs_Interactive_NoSteps(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no shared steps found")
}

func TestSharedStepHistoryCmd_NoArgs_Interactive_SelectStepError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{{ID: 56789, Title: "Shared Step"}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select shared step")
}

func TestSharedStepHistoryCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestSharedStepHistoryCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepHistoryFunc: func(ctx context.Context, stepID int64) (*data.GetSharedStepHistoryResponse, error) {
			return nil, fmt.Errorf("shared step not found")
		},
	}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCaseHistoryCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCaseHistoryCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestSharedStepHistoryCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSharedStepHistoryCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestHistoryCommands_GlobalRegistrationVarsInitialized(t *testing.T) {
	assert.NotNil(t, caseHistoryCmd)
	assert.NotNil(t, sharedStepHistoryCmd)
}
