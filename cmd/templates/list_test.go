package templates

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
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
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
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
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
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

func TestListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты для outputResult ====================

func TestOutputResult_Stdout(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
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
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
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

	err := outputResult(cmd, invalidData)
	assert.Error(t, err)
}

func TestOutputResult_SaveToFile(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
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
