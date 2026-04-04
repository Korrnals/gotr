package labels

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLabelIDInteractive_TableDriven(t *testing.T) {
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
				GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
					return data.GetLabelsResponse{{ID: 501, Name: "L501"}}, nil
				},
			},
			wantID: 501,
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 5},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
					return data.GetLabelsResponse{{ID: 501, Name: "L501"}}, nil
				},
			},
			wantErrPart: "failed to select label",
		},
		{
			name: "empty input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
					return nil, nil
				},
			},
			wantErrPart: "no labels found",
		},
		{
			name: "get labels error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get labels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveLabelIDInteractive(tt.ctx, tt.cli)
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
					return []data.Test{{ID: 801, Title: "T801"}}, nil
				},
			},
			wantID: 801,
		},
		{
			name: "get runs error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get runs",
		},
		{
			name: "get tests error",
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
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get tests",
		},
		{
			name: "select run error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 5},
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
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 3},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 101, Name: "R101"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 801, Title: "T801"}}, nil
				},
			},
			wantErrPart: "failed to select test",
		},
		{
			name: "empty input",
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
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return nil, nil
				},
			},
			wantErrPart: "no tests found",
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