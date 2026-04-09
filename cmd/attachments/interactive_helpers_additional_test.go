package attachments

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveCaseIDInteractive_MissingBranches(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantErrPart string
	}{
		{
			name: "get suites error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 11, Name: "P"}}, nil
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
					return data.GetProjectsResponse{{ID: 12, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{}, nil
				},
			},
			wantErrPart: "no suites found in project 12",
		},
		{
			name: "select suite error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
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
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 2, Name: "S"}}, nil
				},
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get cases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveCaseIDInteractive(tt.ctx, tt.cli)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrPart)
		})
	}
}

func TestResolveAttachmentIDInteractive_Table(t *testing.T) {
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
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 1},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 2, Name: "S"}}, nil
				},
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{{ID: 3, Title: "C"}}, nil
				},
				GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
					return data.GetAttachmentsResponse{{ID: 7, Name: "A1"}, {ID: 8, Name: "A2"}}, nil
				},
			},
			wantID: 8,
		},
		{
			name: "attachments error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 10, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 20, Name: "S"}}, nil
				},
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{{ID: 30, Title: "C"}}, nil
				},
				GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get attachments for case 30",
		},
		{
			name: "no attachments",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 2, Name: "S"}}, nil
				},
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{{ID: 9, Title: "C"}}, nil
				},
				GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
					return data.GetAttachmentsResponse{}, nil
				},
			},
			wantErrPart: "no attachments found in case 9",
		},
		{
			name: "select attachment error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
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
				GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{{ID: 3, Title: "C"}}, nil
				},
				GetAttachmentsForCaseFunc: func(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
					return data.GetAttachmentsResponse{{ID: 4, Name: "A"}}, nil
				},
			},
			wantErrPart: "failed to select attachment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveAttachmentIDInteractive(tt.ctx, tt.cli)
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

func TestResolveTestIDInteractive_Table(t *testing.T) {
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
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 10, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 101, CaseID: 1, Title: "T1"}, {ID: 202, CaseID: 2, Title: "T2"}}, nil
				},
			},
			wantID: 202,
		},
		{
			name: "get tests error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 11, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get tests for run 11",
		},
		{
			name: "no tests",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 12, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{}, nil
				},
			},
			wantErrPart: "no tests found in run 12",
		},
		{
			name: "select test error",
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
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 13, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 9, CaseID: 9, Title: "T"}}, nil
				},
			},
			wantErrPart: "failed to select test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveTestIDInteractive(tt.ctx, tt.cli)
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

func TestResolveRunIDInteractive_Table(t *testing.T) {
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
			name: "get runs error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 10, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get runs for project 10",
		},
		{
			name: "no runs",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 11, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{}, nil
				},
			},
			wantErrPart: "no runs found in project 11",
		},
		{
			name: "select run error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 22, Name: "R"}}, nil
				},
			},
			wantErrPart: "failed to select run",
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
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 50, Name: "R1"}, {ID: 60, Name: "R2"}}, nil
				},
			},
			wantID: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveRunIDInteractive(tt.ctx, tt.cli)
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

func TestResolvePlanIDInteractive_Table(t *testing.T) {
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
			name: "get plans error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 7, Name: "P"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get plans for project 7",
		},
		{
			name: "no plans",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 8, Name: "P"}}, nil
				},
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return data.GetPlansResponse{}, nil
				},
			},
			wantErrPart: "no plans found in project 8",
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
					return data.GetPlansResponse{{ID: 31, Name: "Plan"}}, nil
				},
			},
			wantErrPart: "failed to select plan",
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
				GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
					return data.GetPlansResponse{{ID: 70, Name: "P1"}, {ID: 80, Name: "P2"}}, nil
				},
			},
			wantID: 80,
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

func TestResolveResultIDInteractive_Table(t *testing.T) {
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
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 1},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 10, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 100, CaseID: 1, Title: "T"}}, nil
				},
				GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
					return data.GetResultsResponse{{ID: 5, StatusID: 1, TestID: testID}, {ID: 6, StatusID: 5, TestID: testID}}, nil
				},
			},
			wantID: 6,
		},
		{
			name: "get results error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 10, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 42, CaseID: 1, Title: "T"}}, nil
				},
				GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
					return nil, assert.AnError
				},
			},
			wantErrPart: "failed to get results for test 42",
		},
		{
			name: "no results",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 10, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 77, CaseID: 1, Title: "T"}}, nil
				},
				GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
					return data.GetResultsResponse{}, nil
				},
			},
			wantErrPart: "no results found in test 77",
		},
		{
			name: "select result error",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 0},
					interactive.SelectResponse{Index: 9},
				)),
			cli: &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
				},
				GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
					return data.GetRunsResponse{{ID: 10, Name: "R"}}, nil
				},
				GetTestsFunc: func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
					return []data.Test{{ID: 100, CaseID: 1, Title: "T"}}, nil
				},
				GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
					return data.GetResultsResponse{{ID: 9, StatusID: 1, TestID: testID}}, nil
				},
			},
			wantErrPart: "failed to select result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveResultIDInteractive(tt.ctx, tt.cli)
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
