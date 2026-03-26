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

// ==================== Функциональные тесты с моком ====================

func TestRunCrossProjectCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		RunCrossProjectReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			assert.Equal(t, int64(42), templateID)
			return &data.RunReportResponse{
				ReportID: 2000,
				URL:      "https://testrail.example.com/cross-report/2000",
				Status:   "pending",
			}, nil
		},
	}

	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRunCrossProjectCmd_Completed(t *testing.T) {
	mock := &client.MockClient{
		RunCrossProjectReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			return &data.RunReportResponse{
				ReportID: 2001,
				URL:      "https://testrail.example.com/cross-report/2001",
				Status:   "completed",
			}, nil
		},
	}

	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRunCrossProjectCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		RunCrossProjectReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			return nil, fmt.Errorf("template not found")
		},
	}

	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты валидации ====================

func TestRunCrossProjectCmd_InvalidTemplateID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRunCrossProjectCmd_NoArgs_Interactive(t *testing.T) {
	mock := &client.MockClient{
		GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
			return data.GetReportsResponse{{ID: 42, Name: "Cross Summary"}}, nil
		},
		RunCrossProjectReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			assert.Equal(t, int64(42), templateID)
			return &data.RunReportResponse{ReportID: 2000, Status: "pending"}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRunCrossProjectCmd_NoArgs_NonInteractive_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	niPrompter := interactive.NewNonInteractivePrompter()
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), niPrompter))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

func TestRunCrossProjectCmd_ZeroTemplateID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRunCrossProjectCmd_NegativeTemplateID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-1"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRunCrossProjectCmd_DryRun_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		RunCrossProjectReportFunc: func(ctx context.Context, templateID int64) (*data.RunReportResponse, error) {
			called = true
			return &data.RunReportResponse{ReportID: 2, Status: "pending"}, nil
		},
	}

	cmd := newRunCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.False(t, called)
}
