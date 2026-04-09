package reports

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Functional tests with mock ====================

func TestRunCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		RunReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
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
		RunReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
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
		RunReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			return nil, fmt.Errorf("template not found")
		},
	}

	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Validation tests ====================

func TestRunCmd_InvalidTemplateID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRunCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetReportsFunc: func(ctx context.Context, projectID int64) (data.GetReportsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetReportsResponse{{ID: 42, Name: "Summary Report"}}, nil
		},
		RunReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			assert.Equal(t, int64(42), templateID)
			return &data.RunReportResponse{ReportID: 1000, Status: "pending"}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRunCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCmd(testhelper.GetClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestRunCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template_id is required in non-interactive mode")
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

func TestRunCmd_DryRun_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		RunReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			called = true
			return &data.RunReportResponse{ReportID: 1, Status: "pending"}, nil
		},
	}

	cmd := newRunCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.False(t, called)
}
