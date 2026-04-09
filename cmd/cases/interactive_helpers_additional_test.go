package cases

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveProjectAndSuiteInteractive_Table(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantProject int64
		wantSuite   int64
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
					return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 10, Name: "S1"}, {ID: 20, Name: "S2"}}, nil
				},
			},
			wantProject: 1,
			wantSuite:   20,
		},
		{
			name: "get suites error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 2, Name: "P2"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get suites",
		},
		{
			name: "no suites",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 3, Name: "P3"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{}, nil
				},
			},
			wantErrPart: "no suites found in project 3",
		},
		{
			name: "suite select error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
				},
			},
			wantErrPart: "failed to select suite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectID, suiteID, err := resolveProjectAndSuiteInteractive(tt.ctx, tt.cli)
			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantProject, projectID)
			assert.Equal(t, tt.wantSuite, suiteID)
		})
	}
}

func TestResolveSectionIDInteractive_Table(t *testing.T) {
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
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 1},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 11, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 22, Name: "S"}}, nil
				},
				GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
					return data.GetSectionsResponse{{ID: 101, Name: "Root"}, {ID: 202, Name: "Child"}}, nil
				},
			},
			wantID: 202,
		},
		{
			name: "get sections error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 9, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 8, Name: "S"}}, nil
				},
				GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get sections",
		},
		{
			name: "no sections",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 4, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 5, Name: "S"}}, nil
				},
				GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
					return data.GetSectionsResponse{}, nil
				},
			},
			wantErrPart: "no sections found in suite 5",
		},
		{
			name: "select section error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 2, Name: "S"}}, nil
				},
				GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
					return data.GetSectionsResponse{{ID: 3, Name: "Sec"}}, nil
				},
			},
			wantErrPart: "failed to select section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveSectionIDInteractive(tt.ctx, tt.cli)
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
