package variables

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetVariablesFunc: func(datasetID int64) (data.GetVariablesResponse, error) {
			assert.Equal(t, int64(123), datasetID)
			return []data.Variable{
				{ID: 1, Name: "username", DatasetID: 123},
				{ID: 2, Name: "password", DatasetID: 123},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetVariablesFunc: func(datasetID int64) (data.GetVariablesResponse, error) {
			return []data.Variable{}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"456"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetVariablesFunc: func(datasetID int64) (data.GetVariablesResponse, error) {
			return nil, fmt.Errorf("датасет не найден")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_WithOutputFile(t *testing.T) {
	mock := &client.MockClient{
		GetVariablesFunc: func(datasetID int64) (data.GetVariablesResponse, error) {
			return []data.Variable{{ID: 1, Name: "email"}}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "vars.json")

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"123", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "email")
}

func TestListCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
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
	ctx := context.WithValue(context.Background(), "other_key", "value")
	cmd.SetContext(ctx)

	result := getClientForTests(cmd)
	assert.Nil(t, result)
}

func TestOutputResult_JSONError(t *testing.T) {
	badData := make(chan int)

	cmd := &cobra.Command{}
	cmd.Flags().String("save", "", "")

	err := outputResult(cmd, badData)
	assert.Error(t, err)
}

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	varsCmd, _, err := root.Find([]string{"variables"})
	assert.NoError(t, err)
	assert.NotNil(t, varsCmd)

	// Check all subcommands exist
	listCmd, _, _ := root.Find([]string{"variables", "list"})
	assert.NotNil(t, listCmd)

	addCmd, _, _ := root.Find([]string{"variables", "add"})
	assert.NotNil(t, addCmd)

	updateCmd, _, _ := root.Find([]string{"variables", "update"})
	assert.NotNil(t, updateCmd)

	deleteCmd, _, _ := root.Find([]string{"variables", "delete"})
	assert.NotNil(t, deleteCmd)
}
