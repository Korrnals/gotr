package reports

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Functional tests with mock ====================

func TestListCrossProjectCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
			return data.GetReportsResponse{
				{ID: 100, Name: "Cross Project Summary", Description: "Summary across projects"},
				{ID: 200, Name: "Cross Project Coverage", Description: "Coverage across projects"},
			}, nil
		},
	}

	cmd := newListCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCrossProjectCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
			return data.GetReportsResponse{}, nil
		},
	}

	cmd := newListCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCrossProjectCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
			return nil, fmt.Errorf("failed to fetch cross-project reports")
		},
	}

	cmd := newListCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch cross-project reports")
}

func TestListCrossProjectCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
			return data.GetReportsResponse{
				{ID: 100, Name: "Cross Project Report", Description: "Test"},
			}, nil
		},
	}

	cmd := newListCrossProjectCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
