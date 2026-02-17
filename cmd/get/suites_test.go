package get

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для get suites ====================

func TestSuitesCmd_WithProjectID(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Master Suite", ProjectID: 30},
				{ID: 20070, Name: "Regression Suite", ProjectID: 30},
			}, nil
		},
	}

	cmd := newSuitesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSuitesCmd_WithProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Master Suite", ProjectID: 30},
			}, nil
		},
	}

	cmd := newSuitesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSuitesCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
	}

	cmd := newSuitesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSuitesCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSuitesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID проекта")
}

func TestSuitesCmd_InvalidProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSuitesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSuitesCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newSuitesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты для get suite ====================

func TestSuiteCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSuiteFunc: func(suiteID int64) (*data.Suite, error) {
			assert.Equal(t, int64(20069), suiteID)
			return &data.Suite{
				ID:          20069,
				Name:        "Master Suite",
				Description: "Main test suite",
				ProjectID:   30,
				IsMaster:    true,
			}, nil
		},
	}

	cmd := newSuiteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"20069"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSuiteCmd_InvalidSuiteID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSuiteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID сюиты")
}

func TestSuiteCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSuiteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSuiteCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSuiteFunc: func(suiteID int64) (*data.Suite, error) {
			return nil, fmt.Errorf("suite not found")
		},
	}

	cmd := newSuiteCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSuitesCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSuitesCmd(nilClientFunc)
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestSuiteCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSuiteCmd(nilClientFunc)
	cmd.SetArgs([]string{"20069"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}
