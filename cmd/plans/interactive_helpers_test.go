package plans

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolvePlanIDInteractive_Table(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantID      int64
		wantErrPart string
	}{
		{
			name: "success",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 1},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 10, Name: "P"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return data.GetPlansResponse{{ID: 100, Name: "Plan A"}, {ID: 200, Name: "Plan B"}}, nil
				},
			},
			wantID: 200,
		},
		{
			name: "get plans error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 77, Name: "P"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get plans for project 77",
		},
		{
			name: "no plans",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 55, Name: "P"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return data.GetPlansResponse{}, nil
				},
			},
			wantErrPart: "no plans found in project 55",
		},
		{
			name: "select plan error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return data.GetPlansResponse{{ID: 99, Name: "Only"}}, nil
				},
			},
			wantErrPart: "failed to select plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolvePlanIDInteractive(tt.ctx, tt.cli)
			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestResolvePlanEntryIDInteractive_Table(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantID      string
		wantErrPart string
	}{
		{
			name: "success",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: []data.PlanEntry{{ID: "e-1", Name: "E1"}, {ID: "e-2", Name: "E2"}}}, nil
				},
			},
			wantID: "e-2",
		},
		{
			name: "get plan error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get plan 500",
		},
		{
			name: "no entries",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: nil}, nil
				},
			},
			wantErrPart: "no entries found in plan 500",
		},
		{
			name: "select error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli: &client.MockClient{
				GetPlanFunc: func(ctx context.Context, planID int64) (*data.Plan, error) {
					return &data.Plan{Entries: []data.PlanEntry{{ID: "e-1", Name: "E1"}}}, nil
				},
			},
			wantErrPart: "failed to select plan entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolvePlanEntryIDInteractive(tt.ctx, tt.cli, 500)
			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}
