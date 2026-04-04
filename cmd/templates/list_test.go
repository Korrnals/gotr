package templates

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Template{
				{ID: 1, Name: "Test Case (Text)", IsDefault: true},
				{ID: 2, Name: "Test Case (Steps)", IsDefault: false},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты валидации ====================

func TestListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Template{{ID: 1, Name: "Test Case (Text)", IsDefault: true}}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(testhelper.GetClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestListCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id is required in non-interactive mode")
}

func TestListCmd_ResolveInteractiveError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("projects boom")
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты для outputResult ====================

func TestOutputResult_Stdout(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{
				{ID: 1, Name: "Test Case (Text)", IsDefault: true},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestOutputResult_ToFile(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{
				{ID: 1, Name: "Test Case (Text)", IsDefault: true},
				{ID: 2, Name: "Test Case (Steps)", IsDefault: false},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestOutputResult_MarshalError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("save", "", "")

	// Channel cannot be marshaled to JSON
	invalidData := make(chan int)

	err := output.OutputResult(cmd, invalidData, "templates")
	assert.Error(t, err)
}

func TestOutputResult_SaveToFile(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{
				{ID: 1, Name: "Test Case (Text)", IsDefault: true},
			}, nil
		},
	}

	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для Register ====================

func TestRegister(t *testing.T) {
	root := &cobra.Command{Use: "gotr"}
	mock := &client.MockClient{}

	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	// Verify templates command is registered
	templatesCmd, _, err := root.Find([]string{"templates"})
	assert.NoError(t, err)
	assert.NotNil(t, templatesCmd)
	assert.Equal(t, "templates", templatesCmd.Name())

	// Verify list subcommand is registered
	listCmd, _, err := root.Find([]string{"templates", "list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Name())
}

func TestRegister_Help(t *testing.T) {
	root := &cobra.Command{Use: "gotr"}
	mock := &client.MockClient{}

	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	// Verify calling templates without args shows help
	root.SetArgs([]string{"templates"})
	err := root.Execute()
	assert.NoError(t, err)
}
