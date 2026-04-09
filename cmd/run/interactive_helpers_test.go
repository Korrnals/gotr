package run

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRunID_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		args        []string
		wantID      int64
		wantErrPart string
	}{
		{
			name:   "args parse success",
			ctx:    context.Background(),
			cli:    &client.MockClient{},
			args:   []string{"123"},
			wantID: 123,
		},
		{
			name:        "args parse error",
			ctx:         context.Background(),
			cli:         &client.MockClient{},
			args:        []string{"bad"},
			wantErrPart: "invalid syntax",
		},
		{
			name: "interactive project selection error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
				},
			},
			wantErrPart: "failed to select project",
		},
		{
			name: "interactive get runs error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get runs for project 1",
		},
		{
			name: "interactive select run error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 10, Name: "R10"}}, nil
				},
			},
			wantErrPart: "failed to select run",
		},
		{
			name: "interactive success",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 11, Name: "R11"}}, nil
				},
			},
			wantID: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveRunID(tt.ctx, tt.cli, tt.args)
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
