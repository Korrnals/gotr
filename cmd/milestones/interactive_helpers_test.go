package milestones

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveMilestoneIDInteractive_Table(t *testing.T) {
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
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
					return []data.Milestone{{ID: 10, Name: "M1"}, {ID: 20, Name: "M2"}}, nil
				},
			},
			wantID: 20,
		},
		{
			name: "get milestones error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 42, Name: "P"}}, nil
				},
				GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get milestones for project 42",
		},
		{
			name: "no milestones",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 66, Name: "P"}}, nil
				},
				GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
					return []data.Milestone{}, nil
				},
			},
			wantErrPart: "no milestones found in project 66",
		},
		{
			name: "select milestone error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
					return []data.Milestone{{ID: 7, Name: "M"}}, nil
				},
			},
			wantErrPart: "failed to select milestone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveMilestoneIDInteractive(tt.ctx, tt.cli)
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
