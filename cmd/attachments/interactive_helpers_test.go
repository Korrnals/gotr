package attachments

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePlanEntryIDInteractive_GetPlanError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	cli := &client.MockClient{
		GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
			return nil, assert.AnError
		},
	}

	id, err := resolvePlanEntryIDInteractive(ctx, cli, 100)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get plan 100")
	assert.Empty(t, id)
}

func TestResolvePlanAndEntryIDInteractive_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantPlanID  int64
		wantEntryID string
		wantErrPart string
	}{
		{
			name: "resolve plan id error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
				},
			},
			wantErrPart: "failed to select project",
		},
		{
			name: "resolve plan entry id error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 7, Name: "P7"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return data.GetPlansResponse{{ID: 70, Name: "Plan 70"}}, nil
				},
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get plan 70",
		},
		{
			name: "success",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 5, Name: "P5"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return data.GetPlansResponse{{ID: 50, Name: "Plan 50"}}, nil
				},
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: []data.PlanEntry{{ID: "entry-1", Name: "Entry 1"}}}, nil
				},
			},
			wantPlanID:  50,
			wantEntryID: "entry-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planID, entryID, err := resolvePlanAndEntryIDInteractive(tt.ctx, tt.cli)
			if tt.wantErrPart != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPlanID, planID)
			assert.Equal(t, tt.wantEntryID, entryID)
		})
	}
}

func TestResolvePlanEntryIDInteractive_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantID      string
		wantErr     string
		wantErrPart string
	}{
		{
			name: "non-interactive",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: []data.PlanEntry{{ID: "e-1", Name: "Entry 1"}}}, nil
				},
			},
			wantErrPart: "failed to select plan entry",
		},
		{
			name: "valid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: []data.PlanEntry{{ID: "e-42", Name: "Entry 42"}}}, nil
				},
			},
			wantID: "e-42",
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 3})),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: []data.PlanEntry{{ID: "e-1", Name: "Entry 1"}}}, nil
				},
			},
			wantErrPart: "failed to select plan entry",
		},
		{
			name: "empty input",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: nil}, nil
				},
			},
			wantErrPart: "no plan entries found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePlanEntryIDInteractive(tt.ctx, tt.cli, 100)
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

func TestResolveCaseIDInteractive_TableDriven(t *testing.T) {
	baseClient := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20, Name: "S"}}, nil
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
				GetSuitesFunc:   baseClient.GetSuitesFunc,
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{{ID: 30, Title: "Case 30"}}, nil
				},
			},
			wantID: 30,
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 2},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetSuitesFunc:   baseClient.GetSuitesFunc,
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{{ID: 30, Title: "Case 30"}}, nil
				},
			},
			wantErrPart: "failed to select case",
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
				GetSuitesFunc:   baseClient.GetSuitesFunc,
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return nil, nil
				},
			},
			wantErrPart: "no cases found in suite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCaseIDInteractive(tt.ctx, tt.cli)
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