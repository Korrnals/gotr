package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestFunc: func(testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
			assert.Equal(t, int64(12345), testID)
			assert.Equal(t, int64(1), req.StatusID)
			return &data.Test{ID: testID, StatusID: 1}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--status-id", "1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_WithAssignedTo(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestFunc: func(testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
			assert.Equal(t, int64(5), req.AssignedTo)
			return &data.Test{ID: testID, AssignedTo: 5}, nil
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--assigned-to", "5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_WithOutput(t *testing.T) {
	t.Skip("TODO: fix output file test")
	mock := &client.MockClient{
		UpdateTestFunc: func(testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
			return &data.Test{ID: testID, StatusID: 1}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.json")

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--status-id", "1", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "12345")
}

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--status-id", "1", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"invalid", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestFunc: func(testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
			return nil, fmt.Errorf("тест не найден")
		},
	}

	cmd := newUpdateCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"99999", "--status-id", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тест не найден")
}

func TestGetClientForTests_NilCmd(t *testing.T) {
	result := testhelper.GetClientForTests(nil)
	assert.Nil(t, result)
}

func TestGetClientForTests_NilContext(t *testing.T) {
	cmd := &cobra.Command{}
	result := testhelper.GetClientForTests(cmd)
	assert.Nil(t, result)
}

func TestGetClientForTests_NoMockInContext(t *testing.T) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), "other_key", "value")
	cmd.SetContext(ctx)

	result := testhelper.GetClientForTests(cmd)
	assert.Nil(t, result)
}

func TestOutputResult_JSONError(t *testing.T) {
	badData := make(chan int)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", false, "")

	err := outputResult(cmd, badData, time.Now())
	assert.Error(t, err)
}

func TestRegister(t *testing.T) {
	root := &cobra.Command{}
	Register(root, func(cmd *cobra.Command) client.ClientInterface {
		return &client.MockClient{}
	})

	testsCmd, _, err := root.Find([]string{"tests"})
	assert.NoError(t, err)
	assert.NotNil(t, testsCmd)

	updateCmd, _, _ := root.Find([]string{"tests", "update"})
	assert.NotNil(t, updateCmd)
}

// ==================== Additional Tests for Output Functions ====================

func TestPrintJSON_Success(t *testing.T) {
	cmd := &cobra.Command{}
	data := map[string]string{"key": "value"}

	err := printJSON(cmd, data, time.Now())
	assert.NoError(t, err)
}

func TestPrintJSON_MarshalError(t *testing.T) {
	cmd := &cobra.Command{}
	// Channel cannot be marshaled to JSON
	invalidData := make(chan int)

	err := printJSON(cmd, invalidData, time.Now())
	assert.Error(t, err)
}

func TestOutputResult_WithSaveFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", true, "")

	data := map[string]string{"key": "value"}
	err := outputResult(cmd, data, time.Now())
	assert.NoError(t, err)
}

func TestOutputResult_WithoutSaveFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", false, "")

	data := map[string]string{"key": "value"}
	err := outputResult(cmd, data, time.Now())
	assert.NoError(t, err)
}
