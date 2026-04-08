package bdds

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Login\n  Given user is on login page"), 0o644)
	assert.NoError(t, err)

	mock := &client.MockClient{
		AddBDDFunc: func(ctx context.Context, caseID int64, content string) (*data.BDD, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Contains(t, content, "Feature: Login")
			return &data.BDD{ID: 1, CaseID: caseID, Content: content}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--file", bddFile})

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Test"), 0o644)
	assert.NoError(t, err)

	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--file", bddFile, "--dry-run"})

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_MissingContent(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestAddCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--file", "/nonexistent/file.feature"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestAddCmd_ClientError(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Test"), 0o644)
	assert.NoError(t, err)

	mock := &client.MockClient{
		AddBDDFunc: func(ctx context.Context, caseID int64, content string) (*data.BDD, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "--file", bddFile})

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case not found")
}

func TestAddCmd_NoArgs_Interactive(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Login\n  Given user is on login page"), 0o644)
	assert.NoError(t, err)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetSuitesResponse{{ID: 2, Name: "Suite 1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, int64(2), suiteID)
			return data.GetCasesResponse{{ID: 12345, Title: "Case 1"}}, nil
		},
		AddBDDFunc: func(ctx context.Context, caseID int64, content string) (*data.BDD, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Contains(t, content, "Feature: Login")
			return &data.BDD{ID: 1, CaseID: caseID, Content: content}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{"--file", bddFile})

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Login"), 0o644)
	assert.NoError(t, err)

	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{"--file", bddFile})

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// ============= LAYER 2: !HasPrompterInContext branch =============

func TestAddCmd_NoArgs_NoPrompterInContext_Error(t *testing.T) {
tmpDir := t.TempDir()
bddFile := filepath.Join(tmpDir, "scenario.feature")
err := os.WriteFile(bddFile, []byte("Feature: Login"), 0o644)
assert.NoError(t, err)

mock := &client.MockClient{}
cmd := newAddCmd(getClientForTests)
// No interactive.WithPrompter → HasPrompterInContext = false
cmd.SetContext(setupTestCmd(t, mock).Context())
cmd.SetArgs([]string{"--file", bddFile})

err = cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestAddCmd_NoArgs_Interactive_GetProjectsError(t *testing.T) {
tmpDir := t.TempDir()
bddFile := filepath.Join(tmpDir, "scenario.feature")
err := os.WriteFile(bddFile, []byte("Feature: Error Test"), 0o644)
assert.NoError(t, err)

mock := &client.MockClient{
GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
return nil, fmt.Errorf("projects unavailable")
},
}
p := interactive.NewMockPrompter()
cmd := newAddCmd(getClientForTests)
cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
cmd.SetArgs([]string{"--file", bddFile})

err = cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "projects unavailable")
}
