package bdds

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCaseIDInteractive_TableDriven(t *testing.T) {
	baseClient := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 11, Name: "S11"}}, nil
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
					return data.GetCasesResponse{{ID: 1010, Title: "C1010"}}, nil
				},
			},
			wantID: 1010,
		},
		{
			name: "select suite error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetSuitesFunc:   baseClient.GetSuitesFunc,
			},
			wantErrPart: "failed to select suite",
		},
		{
			name: "get cases error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetSuitesFunc:   baseClient.GetSuitesFunc,
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get cases for project 1 suite 11",
		},
		{
			name: "invalid input",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 6},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: baseClient.GetProjectsFunc,
				GetSuitesFunc:   baseClient.GetSuitesFunc,
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{{ID: 1010, Title: "C1010"}}, nil
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
			wantErrPart: "no cases found in project",
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