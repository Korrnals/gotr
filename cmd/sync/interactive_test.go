package sync

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for interactive.SelectProject ====================

func TestSelectProject_Success_FirstProject(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
				{ID: 2, Name: "Project 2"},
			}, nil
		},
	}
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})

	id, err := interactive.SelectProject(ctx, p, mock, "Test prompt:")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestSelectProject_Success_SecondProject(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "Alpha"},
				{ID: 20, Name: "Beta"},
				{ID: 30, Name: "Gamma"},
			}, nil
		},
	}
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})

	id, err := interactive.SelectProject(ctx, p, mock, "Select project:")
	assert.NoError(t, err)
	assert.Equal(t, int64(20), id)
}

func TestSelectProject_GetProjectsError(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("failed to fetch projects")
		},
	}
	p := interactive.NewMockPrompter()

	id, err := interactive.SelectProject(ctx, p, mock, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get projects list")
	assert.Equal(t, int64(0), id)
}

func TestSelectProject_NoProjects(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{}, nil
		},
	}
	p := interactive.NewMockPrompter()

	id, err := interactive.SelectProject(ctx, p, mock, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no projects found")
	assert.Equal(t, int64(0), id)
}

func TestSelectProject_SelectQueueExhausted(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
				{ID: 2, Name: "Project 2"},
			}, nil
		},
	}
	// No queued responses — should return error
	p := interactive.NewMockPrompter()

	id, err := interactive.SelectProject(ctx, p, mock, "Test prompt:")
	assert.Error(t, err)
	assert.Equal(t, int64(0), id)
}

// ==================== Tests for interactive.SelectSuiteForProject ====================

func TestSelectSuiteForProject_Success(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetSuitesResponse{
				{ID: 10, Name: "Suite 1"},
				{ID: 20, Name: "Suite 2"},
			}, nil
		},
	}
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})

	id, err := interactive.SelectSuiteForProject(ctx, p, mock, 1, "Test prompt:")
	assert.NoError(t, err)
	assert.Equal(t, int64(20), id)
}

func TestSelectSuiteForProject_GetSuitesError(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, fmt.Errorf("failed to fetch suites")
		},
	}
	p := interactive.NewMockPrompter()

	id, err := interactive.SelectSuiteForProject(ctx, p, mock, 1, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get suites")
	assert.Equal(t, int64(0), id)
}

func TestSelectSuiteForProject_NoSuites(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
	}
	p := interactive.NewMockPrompter()

	id, err := interactive.SelectSuiteForProject(ctx, p, mock, 1, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no suites found")
	assert.Equal(t, int64(0), id)
}

func TestSelectSuiteForProject_SelectQueueExhausted(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 10, Name: "Suite 1"},
				{ID: 20, Name: "Suite 2"},
			}, nil
		},
	}
	// No queued responses — should return error
	p := interactive.NewMockPrompter()

	id, err := interactive.SelectSuiteForProject(ctx, p, mock, 1, "Test prompt:")
	assert.Error(t, err)
	assert.Equal(t, int64(0), id)
}
