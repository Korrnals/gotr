package bdds

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetBDDFunc: func(ctx context.Context, caseID int64) (*data.BDD, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.BDD{
				ID:      1,
				CaseID:  caseID,
				Content: "Feature: Login\n  Scenario: Successful login",
			}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetBDDFunc: func(ctx context.Context, caseID int64) (*data.BDD, error) {
			return &data.BDD{ID: 1, CaseID: caseID, Content: "Feature: Test"}, nil
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetBDDFunc: func(ctx context.Context, caseID int64) (*data.BDD, error) {
			return nil, fmt.Errorf("BDD not found")
		},
	}

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

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
			return data.GetSuitesResponse{{ID: 2, Name: "Suite 1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, int64(2), suiteID)
			return data.GetCasesResponse{{ID: 12345, Title: "Case 1"}}, nil
		},
		GetBDDFunc: func(ctx context.Context, caseID int64) (*data.BDD, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.BDD{ID: 1, CaseID: caseID, Content: "Feature: Test"}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(getClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetClientForTests_NilCmd(t *testing.T) {
	result := getClientForTests(nil)
	assert.Nil(t, result)
}

func TestGetClientForTests_NilContext(t *testing.T) {
	cmd := &cobra.Command{}
	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_NoMockInContext(t *testing.T) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), testContextKey("other_key"), "value")
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestOutputResult_JSONError(t *testing.T) {
	badData := make(chan int)

	cmd := &cobra.Command{}
	cmd.Flags().String("save", "", "")

	err := output.OutputResult(cmd, badData, "bdds")
	assert.Error(t, err)
}

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	bddsCmd, _, err := root.Find([]string{"bdds"})
	assert.NoError(t, err)
	assert.NotNil(t, bddsCmd)

	getCmd, _, _ := root.Find([]string{"bdds", "get"})
	assert.NotNil(t, getCmd)

	addCmd, _, _ := root.Find([]string{"bdds", "add"})
	assert.NotNil(t, addCmd)
}

// ============= LAYER 2: !HasPrompterInContext branch + resolveCaseIDInteractive error =============

func TestGetCmd_NoArgs_NoPrompterInContext_Error(t *testing.T) {
mock := &client.MockClient{}
cmd := newGetCmd(getClientForTests)
// No interactive.WithPrompter → HasPrompterInContext = false
cmd.SetContext(setupTestCmd(t, mock).Context())
cmd.SetArgs([]string{})

err := cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestGetCmd_NoArgs_Interactive_GetProjectsError(t *testing.T) {
mock := &client.MockClient{
GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
return nil, fmt.Errorf("projects boom")
},
}
p := interactive.NewMockPrompter()
cmd := newGetCmd(getClientForTests)
cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
cmd.SetArgs([]string{})

err := cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "projects boom")
}
