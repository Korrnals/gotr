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

func TestProjectCmd_ZeroProjectID(t *testing.T) {
	cmd := newProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project_id")
}

func TestSectionGetCmd_NoArgs_Interactive_NoSections(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P10"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no sections found")
}

func TestSectionGetCmd_NoArgs_Interactive_SelectSectionError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P10"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 1, Name: "S1"}}, nil
		},
	}

	// Only first select (project) is available; section select should fail.
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select section")
}

func TestProjectCmd_NoArgs_Interactive_SelectProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("project lookup failed")
		},
	}
	cmd := newProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project lookup failed")
}
