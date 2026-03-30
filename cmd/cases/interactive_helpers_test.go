package cases

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveCaseIDInteractive_GetSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, assert.AnError
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	id, err := resolveCaseIDInteractive(ctx, mock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get suites")
	assert.Zero(t, id)
}

func TestResolveCaseIDInteractive_NoSuites(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	id, err := resolveCaseIDInteractive(ctx, mock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no suites found in project 1")
	assert.Zero(t, id)
}

func TestResolveCaseIDInteractive_GetCasesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return nil, assert.AnError
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	id, err := resolveCaseIDInteractive(ctx, mock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cases")
	assert.Zero(t, id)
}

func TestResolveCaseIDInteractive_NoCases(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	id, err := resolveCaseIDInteractive(ctx, mock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cases found in suite 10")
	assert.Zero(t, id)
}

func TestResolveCaseIDInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{
				{ID: 100, Title: "Case A"},
				{ID: 200, Title: "Case B"},
			}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 1})
	ctx := interactive.WithPrompter(context.Background(), p)

	id, err := resolveCaseIDInteractive(ctx, mock)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), id)
}

func TestSelectCaseID_NonInteractive(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cases := data.GetCasesResponse{{ID: 100, Title: "Case A"}}

	id, err := selectCaseID(ctx, cases)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select case")
	assert.Zero(t, id)
}
