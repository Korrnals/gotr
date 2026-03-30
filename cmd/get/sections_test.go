package get

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для get sections section ====================

func TestSectionGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSectionFunc: func(ctx context.Context, sectionID int64) (*data.Section, error) {
			assert.Equal(t, int64(100), sectionID)
			return &data.Section{
				ID:      100,
				Name:    "Test Section",
				SuiteID: 20069,
				Depth:   0,
			}, nil
		},
	}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionGetCmd_InvalidSectionID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section_id")
}

func TestSectionGetCmd_ZeroSectionID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section_id")
}

func TestSectionGetCmd_NegativeSectionID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-5"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionGetCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(0), suiteID)
			return data.GetSectionsResponse{{ID: 100, Name: "Test Section", SuiteID: 20069}}, nil
		},
		GetSectionFunc: func(ctx context.Context, sectionID int64) (*data.Section, error) {
			assert.Equal(t, int64(100), sectionID)
			return &data.Section{ID: 100, Name: "Test Section", SuiteID: 20069}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestSectionGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSectionFunc: func(ctx context.Context, sectionID int64) (*data.Section, error) {
			return nil, fmt.Errorf("section not found")
		},
	}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты для get sections list ====================

func TestSectionsListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(0), suiteID)
			return data.GetSectionsResponse{
				{ID: 1, Name: "Section 1", SuiteID: 20069},
				{ID: 2, Name: "Section 2", SuiteID: 20069},
			}, nil
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionsListCmd_WithSuiteID(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			return data.GetSectionsResponse{
				{ID: 1, Name: "Section 1", SuiteID: 20069},
			}, nil
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionsListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestSectionsListCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestSectionsListCmd_NegativeProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-10"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionsListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSectionsResponse{{ID: 1, Name: "Section 1", SuiteID: 20069}}, nil
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionsListCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestSectionsListCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSectionGetCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section_id required")
}

func TestSectionGetCmd_NoArgs_Interactive_GetSectionsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 30, Name: "Project 30"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return nil, fmt.Errorf("sections unavailable")
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get sections list")
}

func TestSelectSectionID_SelectError(t *testing.T) {
	sections := data.GetSectionsResponse{{ID: 10, Name: "S-10"}}
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())

	sectionID, err := selectSectionID(ctx, sections)
	assert.Error(t, err)
	assert.Equal(t, int64(0), sectionID)
	assert.Contains(t, err.Error(), "failed to select section")
}
