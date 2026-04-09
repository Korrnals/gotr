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

// ==================== Tests for get cases ====================

func TestCasesCmd_WithSuiteID(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			assert.Equal(t, int64(0), sectionID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case 1"},
				{ID: 2, Title: "Test Case 2"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_WithProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case 1"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_WithSectionID(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			assert.Equal(t, int64(100), sectionID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case in Section"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--section-id", "100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_AutoSelectSingleSuite(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Master Suite"},
			}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case 1"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"}) // Without --suite-id

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_AllSuites(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
				{ID: 20070, Name: "Suite 2"},
			}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			if suiteID == 20069 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			if suiteID == 20070 {
				return data.GetCasesResponse{{ID: 2, Title: "Case 2"}}, nil
			}
			return nil, fmt.Errorf("unexpected suite ID: %d", suiteID)
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--all-suites"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestCasesCmd_InvalidProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCasesCmd_NoSuites(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil // No suites
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no suites found")
}

func TestCasesCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCasesCmd_GetSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, fmt.Errorf("failed to get suites")
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get suites")
}

func TestCasesCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestCasesCmd_MultipleSuites_InteractiveSelect(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
				{ID: 20070, Name: "Suite 2"},
			}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20070), suiteID)
			return data.GetCasesResponse{{ID: 2, Title: "Case 2"}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})
	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_MultipleSuites_InteractiveSelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
				{ID: 20070, Name: "Suite 2"},
			}, nil
		},
	}

	// Empty queue forces MockPrompter.Select() error branch.
	p := interactive.NewMockPrompter()
	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select")
}

// ==================== Tests for get case ====================

func TestCaseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.Case{
				ID:    12345,
				Title: "Test Case",
			}, nil
		},
	}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid case ID")
}

func TestCaseCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCaseCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite", ProjectID: 30}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
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

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseCmd_NoArgs_Interactive_GetSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, fmt.Errorf("failed to get suites")
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get suites")
}

func TestCaseCmd_NoArgs_Interactive_NoSuites(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no suites found")
}

func TestCaseCmd_NoArgs_Interactive_SelectSuiteError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}, {ID: 20070, Name: "Other Suite"}}, nil
		},
	}

	// One select response for project, then missing response for suite selection.
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select")
}

func TestCaseCmd_NoArgs_Interactive_GetCasesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return nil, fmt.Errorf("failed to get cases")
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cases")
}

func TestCaseCmd_NoArgs_Interactive_NoCases(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cases found")
}

func TestCaseCmd_NoArgs_Interactive_SelectCaseError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20069, Name: "Master Suite"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 12345, Title: "Case 12345"}}, nil
		},
	}

	// Two select responses for project and suite, missing response for case selection.
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select case")
}

func TestCaseCommands_GlobalRegistrationVarsInitialized(t *testing.T) {
	assert.NotNil(t, casesCmd)
	assert.NotNil(t, caseCmd)
}

func TestCaseCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestCaseCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Tests for fetchCasesFromAllSuites ====================

func TestFetchCasesFromAllSuites_WithError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
				{ID: 20070, Name: "Suite 2"},
			}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			if suiteID == 20070 {
				return nil, fmt.Errorf("suite not found")
			}
			return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--all-suites"})

	err := cmd.Execute()
	// Error in one suite does not stop execution, just prints a message
	assert.NoError(t, err)
}

func TestCasesCmd_AllSuites_WithSectionID(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
			}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(100), sectionID)
			return data.GetCasesResponse{{ID: 1, Title: "Case in Section"}}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--all-suites", "--section-id", "100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_NilClient(t *testing.T) {
	// Test case when client is not initialized
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCasesCmd(nilClientFunc)
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestCaseCmd_NilClient(t *testing.T) {
	// Test case when client is not initialized
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCaseCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}
