package result

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAddCmd_NoArgs_Interactive_RunResolveError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("failed to select project")
		},
	}

	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{"--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select project")
}

func TestAddCmd_NoArgs_Interactive_SelectTestError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 100, Name: "R1"}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{{ID: 10, CaseID: 20, Title: "T1"}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newAddCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select test")
}

func TestAddCaseCmd_NoArgs_NoPrompter(t *testing.T) {
	cmd := newAddCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"--case-id", "10", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run_id required in non-interactive mode")
}

func TestFieldsCmd_NilClient(t *testing.T) {
	cmd := newFieldsCmd(func(*cobra.Command) client.ClientInterface { return nil })
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestAddBulkCmd_FileReadSuccess_ParseErrorPath(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "broken.json")
	assert.NoError(t, os.WriteFile(filePath, []byte("not-json"), 0o600))

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"123", "--results-file", filePath})

	err := cmd.Execute()
	assert.Error(t, err)
}
