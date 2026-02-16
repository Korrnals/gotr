package templates

import (
	"fmt"
	"os"
	"path/filepath"
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
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "templates.json")

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
	cmd.SetArgs([]string{"1", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify file was created and contains expected content
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test Case (Text)")
	assert.Contains(t, string(content), "Test Case (Steps)")
}

func TestOutputResult_MarshalError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("save", "", "")

	// Channel cannot be marshaled to JSON
	invalidData := make(chan int)

	err := outputResult(cmd, invalidData)
	assert.Error(t, err)
}

func TestOutputResult_WriteFileError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("save", "", "")
	// Set output to a path that cannot be created (invalid directory)
	cmd.Flags().Set("output", "/nonexistent/dir/file.json")

	data := map[string]string{"key": "value"}

	err := outputResult(cmd, data)
	assert.Error(t, err)
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
