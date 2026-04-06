package groups

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveGroupIDInteractive_Table(t *testing.T) {
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
				GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
					return data.GetGroupsResponse{{ID: 101, Name: "G1"}, {ID: 202, Name: "G2"}}, nil
				},
			},
			wantID: 202,
		},
		{
			name: "get groups error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 9, Name: "P"}}, nil
				},
				GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get groups for project 9",
		},
		{
			name: "no groups",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 8, Name: "P"}}, nil
				},
				GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
					return data.GetGroupsResponse{}, nil
				},
			},
			wantErrPart: "no groups found in project 8",
		},
		{
			name: "select group error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
					return data.GetGroupsResponse{{ID: 7, Name: "G"}}, nil
				},
			},
			wantErrPart: "failed to select group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveGroupIDInteractive(tt.ctx, tt.cli)
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
