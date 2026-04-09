package variables

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveVariableIDInteractive_Table(t *testing.T) {
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
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return data.GetDatasetsResponse{{ID: 10, Name: "D"}}, nil
				},
				GetVariablesFunc: func(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
					return data.GetVariablesResponse{{ID: 100, Name: "V1"}, {ID: 200, Name: "V2"}}, nil
				},
			},
			wantID: 200,
		},
		{
			name: "get variables error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return data.GetDatasetsResponse{{ID: 10, Name: "D"}}, nil
				},
				GetVariablesFunc: func(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get variables for dataset 10",
		},
		{
			name: "no variables",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return data.GetDatasetsResponse{{ID: 10, Name: "D"}}, nil
				},
				GetVariablesFunc: func(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
					return data.GetVariablesResponse{}, nil
				},
			},
			wantErrPart: "no variables found in dataset 10",
		},
		{
			name: "select variable error",
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
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return data.GetDatasetsResponse{{ID: 10, Name: "D"}}, nil
				},
				GetVariablesFunc: func(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
					return data.GetVariablesResponse{{ID: 100, Name: "V"}}, nil
				},
			},
			wantErrPart: "failed to select variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveVariableIDInteractive(tt.ctx, tt.cli)
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

func TestResolveDatasetIDInteractive_Table(t *testing.T) {
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
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
			},
			wantErrPart: "failed to select project",
		},
		{
			name: "get datasets error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 12, Name: "P"}}, nil
				},
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get datasets for project 12",
		},
		{
			name: "no datasets",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 13, Name: "P"}}, nil
				},
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return data.GetDatasetsResponse{}, nil
				},
			},
			wantErrPart: "no datasets found in project 13",
		},
		{
			name: "select dataset error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return data.GetDatasetsResponse{{ID: 100, Name: "D"}}, nil
				},
			},
			wantErrPart: "failed to select dataset",
		},
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
				GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
					return data.GetDatasetsResponse{{ID: 10, Name: "D1"}, {ID: 20, Name: "D2"}}, nil
				},
			},
			wantID: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveDatasetIDInteractive(tt.ctx, tt.cli)
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
