package configurations

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveGroupIDInteractive_TableDriven(t *testing.T) {
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
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return data.GetConfigsResponse{{ID: 10, Name: "G10"}}, nil
				},
			},
			wantID: 10,
		},
		{
			name: "get configs error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get configuration groups for project 1",
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 3},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return data.GetConfigsResponse{{ID: 10, Name: "G10"}}, nil
				},
			},
			wantErrPart: "failed to select configuration group",
		},
		{
			name: "empty input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return nil, nil
				},
			},
			wantErrPart: "no configuration groups found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveGroupIDInteractive(tt.ctx, tt.cli)
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

func TestResolveConfigIDInteractive_TableDriven(t *testing.T) {
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
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return data.GetConfigsResponse{{
						ID:   10,
						Name: "G10",
						Configs: []data.Config{
							{ID: 100, Name: "C100"},
						},
					}}, nil
				},
			},
			wantID: 100,
		},
		{
			name: "get configs error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get configuration groups",
		},
		{
			name: "no groups",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return data.GetConfigsResponse{}, nil
				},
			},
			wantErrPart: "no configuration groups found",
		},
		{
			name: "select group error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 4},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return data.GetConfigsResponse{{
						ID:   10,
						Name: "G10",
						Configs: []data.Config{
							{ID: 100, Name: "C100"},
						},
					}}, nil
				},
			},
			wantErrPart: "failed to select configuration group",
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 5},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return data.GetConfigsResponse{{
						ID:   10,
						Name: "G10",
						Configs: []data.Config{
							{ID: 100, Name: "C100"},
						},
					}}, nil
				},
			},
			wantErrPart: "failed to select configuration",
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
				GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
					return data.GetConfigsResponse{{
						ID:      10,
						Name:    "G10",
						Configs: nil,
					}}, nil
				},
			},
			wantErrPart: "no configurations found in group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveConfigIDInteractive(tt.ctx, tt.cli)
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

func TestSelectGroupID_TableDriven(t *testing.T) {
	groups := data.GetConfigsResponse{
		{ID: 10, Name: "Group 10"},
	}

	tests := []struct {
		name        string
		ctx         context.Context
		groups      data.GetConfigsResponse
		wantID      int64
		wantErrPart string
	}{
		{
			name: "non-interactive",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			groups: groups,
			wantErrPart: "failed to select configuration group",
		},
		{
			name: "valid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			groups: groups,
			wantID: 10,
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 5})),
			groups: groups,
			wantErrPart: "failed to select configuration group",
		},
		{
			name: "empty input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter()),
			groups:      nil,
			wantErrPart: "select options list is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectGroupID(tt.ctx, tt.groups)
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