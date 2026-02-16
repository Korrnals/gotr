package reports

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestRunCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		RunReportFunc: func(templateID int64) (*data.RunReportResponse, error) {
			assert.Equal(t, int64(42), templateID)
			return &data.RunReportResponse{
				ReportID: 1000,
				URL:      "https://testrail.example.com/report/1000",
				Status:   "pending",
			}, nil
		},
	}

	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRunCmd_Completed(t *testing.T) {
	mock := &client.MockClient{
		RunReportFunc: func(templateID int64) (*data.RunReportResponse, error) {
			return &data.RunReportResponse{
				ReportID: 1001,
				URL:      "https://testrail.example.com/report/1001",
				Status:   "completed",
			}, nil
		},
	}

	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRunCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		RunReportFunc: func(templateID int64) (*data.RunReportResponse, error) {
			return nil, fmt.Errorf("template not found")
		},
	}

	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты валидации ====================

func TestRunCmd_InvalidTemplateID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRunCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRunCmd_ZeroTemplateID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRunCmd_NegativeTemplateID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-1"})

	err := cmd.Execute()
	assert.Error(t, err)
}
