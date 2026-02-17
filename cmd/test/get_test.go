package test

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(testID int64) (*data.Test, error) {
			assert.Equal(t, int64(12345), testID)
			return &data.Test{
				ID:       testID,
				CaseID:   100,
				RunID:    200,
				Title:    "Test Case Title",
				StatusID: 1,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(testID int64) (*data.Test, error) {
			assert.Equal(t, int64(12345), testID)
			return &data.Test{
				ID:       testID,
				CaseID:   100,
				RunID:    200,
				Title:    "Test Case Title",
				StatusID: 1,
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NegativeID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"-1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(testID int64) (*data.Test, error) {
			return nil, fmt.Errorf("тест не найден")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тест не найден")
}

func TestGetCmd_NilClient(t *testing.T) {
	cmd := newGetCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не инициализирован")
}

func TestGetCmd_InvalidOutputPath(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(testID int64) (*data.Test, error) {
			return &data.Test{ID: testID}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--output", "/nonexistent/dir/test.json"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_WithSaveEnabled(t *testing.T) {
	mock := &client.MockClient{
		GetTestFunc: func(testID int64) (*data.Test, error) {
			return &data.Test{ID: testID}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	testCmd := testhelper.SetupTestCmd(t, mock)
	cmd.SetContext(testCmd.Context())
	cmd.SetArgs([]string{"12345", "--save"})

	// This tests the save path - we just verify it works
	err := cmd.Execute()
	assert.NoError(t, err)
}
