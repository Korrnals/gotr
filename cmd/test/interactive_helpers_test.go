package test

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRunIDInteractive_TableDriven(t *testing.T) {
	baseClient := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
	}

	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantID      int64
		wantErrPart string
	}{
		{
			name: "non-interactive",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli:  baseClient,
			wantErrPart: "failed to select project",
		},
		{
			name: "valid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
				},
			},
			wantID: 101,
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 2},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
				},
			},
			wantErrPart: "failed to select run",
		},
		{
			name: "empty input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return nil, nil
				},
			},
			wantErrPart: "no runs found for project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveRunIDInteractive(tt.ctx, tt.cli)
			if tt.wantErrPart != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestResolveTestIDInteractive_TableDriven(t *testing.T) {
	baseClient := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
	}

	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantID      int64
		wantErrPart string
	}{
		{
			name: "non-interactive",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli:  baseClient,
			wantErrPart: "failed to select project",
		},
		{
			name: "valid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 9001, Title: "T1"}}, nil
				},
			},
			wantID: 9001,
		},
		{
			name: "invalid run input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 4},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 9001, Title: "T1"}}, nil
				},
			},
			wantErrPart: "failed to select run",
		},
		{
			name: "invalid test input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 9001, Title: "T1"}}, nil
				},
			},
			wantErrPart: "failed to select test",
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTestIDInteractive(tt.ctx, tt.cli)
			if tt.wantErrPart != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestResolveRunIDInteractive_GetRunsError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(),
		interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0}))

	cli := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, assert.AnError
		},
	}

	_, err := resolveRunIDInteractive(ctx, cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get runs for project")
}

func TestResolveTestIDInteractive_GetTestsError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(),
		interactive.NewMockPrompter().WithSelectResponses(
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
		))

	cli := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
		},
		GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return nil, assert.AnError
		},
	}

	_, err := resolveTestIDInteractive(ctx, cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tests for run")
}

func TestResolveTestIDInteractive_GetRunsError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(),
		interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0}))

	cli := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, assert.AnError
		},
	}

	_, err := resolveTestIDInteractive(ctx, cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get runs for project")
}

func TestResolveTestIDInteractive_NoRuns(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(),
		interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0}))

	cli := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{}, nil
		},
	}

	_, err := resolveTestIDInteractive(ctx, cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no runs found for project 1")
}
func TestResolveTestIDInteractive_NoTests(t *testing.T) {
ctx := interactive.WithPrompter(context.Background(),
interactive.NewMockPrompter().WithSelectResponses(
interactive.SelectResponse{Index: 0},
interactive.SelectResponse{Index: 0},
))

cli := &client.MockClient{
GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
},
GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
},
GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
return nil, nil
},
}

_, err := resolveTestIDInteractive(ctx, cli)
require.Error(t, err)
assert.Contains(t, err.Error(), "no tests found for run 101")
}

func TestResolveTestIDInteractive_SelectsChosenTest(t *testing.T) {
ctx := interactive.WithPrompter(context.Background(),
interactive.NewMockPrompter().WithSelectResponses(
interactive.SelectResponse{Index: 0},
interactive.SelectResponse{Index: 0},
interactive.SelectResponse{Index: 1},
))

cli := &client.MockClient{
GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
},
GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
},
GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
return []data.Test{{ID: 9001, Title: "T1"}, {ID: 9002, Title: "T2"}}, nil
},
}

got, err := resolveTestIDInteractive(ctx, cli)
require.NoError(t, err)
assert.Equal(t, int64(9002), got)
}
