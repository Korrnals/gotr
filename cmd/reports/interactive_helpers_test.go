package reports

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveReportTemplateIDInteractive_TableDriven(t *testing.T) {
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
				GetReportsFunc: func(ctx context.Context, projectID int64) (data.GetReportsResponse, error) {
					return data.GetReportsResponse{{ID: 100, Name: "R100"}}, nil
				},
			},
			wantID: 100,
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 7},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetReportsFunc: func(ctx context.Context, projectID int64) (data.GetReportsResponse, error) {
					return data.GetReportsResponse{{ID: 100, Name: "R100"}}, nil
				},
			},
			wantErrPart: "failed to select report template",
		},
		{
			name: "empty input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetReportsFunc: func(ctx context.Context, projectID int64) (data.GetReportsResponse, error) {
					return nil, nil
				},
			},
			wantErrPart: "no report templates found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveReportTemplateIDInteractive(tt.ctx, tt.cli)
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

func TestResolveCrossProjectReportTemplateIDInteractive_TableDriven(t *testing.T) {
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
			cli: &client.MockClient{
				GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
					return data.GetReportsResponse{{ID: 50, Name: "CR50"}}, nil
				},
			},
			wantErrPart: "failed to select report template",
		},
		{
			name: "valid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
					return data.GetReportsResponse{{ID: 50, Name: "CR50"}}, nil
				},
			},
			wantID: 50,
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 4})),
			cli: &client.MockClient{
				GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
					return data.GetReportsResponse{{ID: 50, Name: "CR50"}}, nil
				},
			},
			wantErrPart: "failed to select report template",
		},
		{
			name: "empty input",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{
				GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
					return nil, nil
				},
			},
			wantErrPart: "no cross-project report templates found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCrossProjectReportTemplateIDInteractive(tt.ctx, tt.cli)
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

func TestResolveReportTemplateIDInteractive_GetReportsError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(),
		interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0}))

	cli := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetReportsFunc: func(ctx context.Context, projectID int64) (data.GetReportsResponse, error) {
			return nil, assert.AnError
		},
	}

	_, err := resolveReportTemplateIDInteractive(ctx, cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list reports for project")
}

func TestResolveCrossProjectReportTemplateIDInteractive_GetReportsError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	cli := &client.MockClient{
		GetCrossProjectReportsFunc: func(ctx context.Context) (data.GetReportsResponse, error) {
			return nil, assert.AnError
		},
	}

	_, err := resolveCrossProjectReportTemplateIDInteractive(ctx, cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list cross-project reports")
}