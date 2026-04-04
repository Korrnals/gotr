package cmd

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupDeleteTest(t *testing.T, mock *client.MockClient) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{
		Use:   deleteCmd.Use,
		Short: deleteCmd.Short,
		Long:  deleteCmd.Long,
		RunE:  runDelete,
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	cmd.Flags().Bool("soft", false, "Мягкое удаление (где поддерживается)")

	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)

	return cmd
}

func TestDelete_Project_WithID_Success(t *testing.T) {
	called := false
	mock := &client.MockClient{
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			assert.Equal(t, int64(77), projectID)
			return nil
		},
	}

	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"project", "77"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDelete_Project_AutoSelectID_Success(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 11, Name: "Project 11"}}, nil
		},
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			assert.Equal(t, int64(11), projectID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{"project"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSelectCaseID(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})
	cases := data.GetCasesResponse{
		{ID: 100, Title: "Case A"},
		{ID: 200, Title: "Case B"},
	}

	id, err := selectCaseID(context.Background(), p, cases)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), id)
}

func TestSelectSharedStepID(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	steps := data.GetSharedStepsResponse{
		{ID: 555, Title: "Step A"},
	}

	id, err := selectSharedStepID(p, steps)
	assert.NoError(t, err)
	assert.Equal(t, int64(555), id)
}

func TestRunDeleteDryRun_SwitchEndpoints(t *testing.T) {
	dr := output.NewDryRunPrinter("delete")

	assert.NoError(t, runDeleteDryRun(dr, "project", 1))
	assert.NoError(t, runDeleteDryRun(dr, "suite", 2))
	assert.NoError(t, runDeleteDryRun(dr, "section", 3))
	assert.NoError(t, runDeleteDryRun(dr, "case", 4))
	assert.NoError(t, runDeleteDryRun(dr, "run", 5))
	assert.NoError(t, runDeleteDryRun(dr, "shared-step", 6))

	err := runDeleteDryRun(dr, "unknown", 1)
	assert.Error(t, err)
}

func TestResolveDeleteID_SuiteAndRun(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 20, Name: "R1"}}, nil
		},
	}
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0})

	suiteID, err := resolveDeleteID(context.Background(), p, mock, "suite")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), suiteID)

	runID, err := resolveDeleteID(context.Background(), p, mock, "run")
	assert.NoError(t, err)
	assert.Equal(t, int64(20), runID)
}

func TestResolveDeleteID_SectionCaseSharedStep(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 30, Name: "SEC1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 40, Title: "CASE1"}}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{{ID: 50, Title: "STEP1"}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
			interactive.SelectResponse{Index: 0},
		)

	sectionID, err := resolveDeleteID(context.Background(), p, mock, "section")
	assert.NoError(t, err)
	assert.Equal(t, int64(30), sectionID)

	caseID, err := resolveDeleteID(context.Background(), p, mock, "case")
	assert.NoError(t, err)
	assert.Equal(t, int64(40), caseID)

	stepID, err := resolveDeleteID(context.Background(), p, mock, "shared-step")
	assert.NoError(t, err)
	assert.Equal(t, int64(50), stepID)
}

func TestDelete_AutoSelectEndpointAndSuite_Success(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 5, Name: "Project 5"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(5), projectID)
			return data.GetSuitesResponse{{ID: 700, Name: "Suite 700"}}, nil
		},
		DeleteSuiteFunc: func(ctx context.Context, suiteID int64) error {
			called = true
			assert.Equal(t, int64(700), suiteID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDelete_NonInteractive_NoArgs_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			return nil
		},
	}

	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

func TestDelete_DryRun_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
			called = true
			return nil
		},
	}

	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"project", "77", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestDelete_InvalidID_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"project", "abc"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ID")
}

// ============= LAYER 1 AGGRESSIVE EXPANSION =============

func TestDelete_AllEndpointsWithID_Success(t *testing.T) {
	endpoints := []string{"project", "suite", "section", "case", "run", "shared-step"}
	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			called := false
			mock := &client.MockClient{
				DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
					called = true
					return nil
				},
				DeleteSuiteFunc: func(ctx context.Context, suiteID int64) error {
					called = true
					return nil
				},
				DeleteSectionFunc: func(ctx context.Context, sectionID int64) error {
					called = true
					return nil
				},
				DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
					called = true
					return nil
				},
				DeleteRunFunc: func(ctx context.Context, runID int64) error {
					called = true
					return nil
				},
				DeleteSharedStepFunc: func(ctx context.Context, stepID int64, keep int) error {
					called = true
					return nil
				},
			}

			cmd := setupDeleteTest(t, mock)
			cmd.SetArgs([]string{endpoint, "999"})

			err := cmd.Execute()
			assert.NoError(t, err, "endpoint %s failed", endpoint)
			assert.True(t, called, "API not called for %s", endpoint)
		})
	}
}

func TestDelete_Suite_Full_Interactive_Flow(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
				{ID: 2, Name: "Project 2"},
			}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 2 {
				return data.GetSuitesResponse{{ID: 100, Name: "Suite 100"}}, nil
			}
			return nil, nil
		},
		DeleteSuiteFunc: func(ctx context.Context, suiteID int64) error {
			called = true
			assert.Equal(t, int64(100), suiteID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1}). // select Project 2
		WithSelectResponses(interactive.SelectResponse{Index: 0})   // select Suite 100

	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{"suite"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDelete_Section_Full_Interactive_Flow(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 10"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{
				{ID: 20, Name: "Section A"},
				{ID: 21, Name: "Section B"},
			}, nil
		},
		DeleteSectionFunc: func(ctx context.Context, sectionID int64) error {
			called = true
			assert.Equal(t, int64(21), sectionID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}). // select suite
		WithSelectResponses(interactive.SelectResponse{Index: 1})   // select section B

	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{"section"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDelete_Case_Full_Interactive_Flow(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{
				{ID: 100, Title: "Case 1"},
				{ID: 101, Title: "Case 2"},
				{ID: 102, Title: "Case 3"},
			}, nil
		},
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			called = true
			assert.Equal(t, int64(101), caseID)
			return nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithSelectResponses(interactive.SelectResponse{Index: 1}) // select Case 2

	cmd := setupDeleteTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{"case"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestDelete_DryRun_AllEndpoints(t *testing.T) {
	endpoints := []string{"project", "suite", "section", "case", "run", "shared-step"}
	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			called := false
			mock := &client.MockClient{
				DeleteProjectFunc: func(ctx context.Context, projectID int64) error {
					called = true
					return nil
				},
				DeleteSuiteFunc: func(ctx context.Context, suiteID int64) error {
					called = true
					return nil
				},
				DeleteSectionFunc: func(ctx context.Context, sectionID int64) error {
					called = true
					return nil
				},
				DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
					called = true
					return nil
				},
				DeleteRunFunc: func(ctx context.Context, runID int64) error {
					called = true
					return nil
				},
				DeleteSharedStepFunc: func(ctx context.Context, stepID int64, keep int) error {
					called = true
					return nil
				},
			}

			cmd := setupDeleteTest(t, mock)
			cmd.SetArgs([]string{endpoint, "999", "--dry-run"})

			err := cmd.Execute()
			assert.NoError(t, err)
			assert.False(t, called, "should not call API in dry-run for %s", endpoint)
		})
	}
}

func TestDelete_UnsupportedEndpoint_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"unsupported-endpoint", "123"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported endpoint")
}

func TestDelete_ZeroID_Error(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupDeleteTest(t, mock)
	cmd.SetArgs([]string{"project", "0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ID")
}

func TestSelectCaseID_EmptyList_Error(t *testing.T) {
	p := interactive.NewMockPrompter()
	cases := data.GetCasesResponse{}

	id, err := selectCaseID(context.Background(), p, cases)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cases found")
	assert.Equal(t, int64(0), id)
}

func TestSelectCaseID_MultipleOptions(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 2})
	cases := data.GetCasesResponse{
		{ID: 100, Title: "Case 1"},
		{ID: 200, Title: "Case 2"},
		{ID: 300, Title: "Case 3"},
		{ID: 400, Title: "Case 4"},
	}

	id, err := selectCaseID(context.Background(), p, cases)
	assert.NoError(t, err)
	assert.Equal(t, int64(300), id)
}

func TestSelectSharedStepID_MultipleOptions(t *testing.T) {
	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})
	steps := data.GetSharedStepsResponse{
		{ID: 555, Title: "Step A"},
		{ID: 666, Title: "Step B"},
		{ID: 777, Title: "Step C"},
	}

	id, err := selectSharedStepID(p, steps)
	assert.NoError(t, err)
	assert.Equal(t, int64(666), id)
}

func TestSelectSharedStepID_EmptyList_Error(t *testing.T) {
	p := interactive.NewMockPrompter()
	steps := data.GetSharedStepsResponse{}

	id, err := selectSharedStepID(p, steps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no shared steps found")
	assert.Equal(t, int64(0), id)
}

func TestResolveDeleteID_UnsupportedEndpoint(t *testing.T) {
	id, err := resolveDeleteID(context.Background(), interactive.NewMockPrompter(), &client.MockClient{}, "weird")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported endpoint")
	assert.Zero(t, id)
}

func TestResolveDeleteID_ErrorBranches(t *testing.T) {
	t.Run("suite get suites error", func(t *testing.T) {
		mock := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
			},
			GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
				return nil, assert.AnError
			},
		}

		p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
		id, err := resolveDeleteID(context.Background(), p, mock, "suite")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get suites for project 1")
		assert.Zero(t, id)
	})

	t.Run("section get sections error", func(t *testing.T) {
		mock := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
			},
			GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
			},
			GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
				return nil, assert.AnError
			},
		}

		p := interactive.NewMockPrompter().
			WithSelectResponses(interactive.SelectResponse{Index: 0}).
			WithSelectResponses(interactive.SelectResponse{Index: 0})
		id, err := resolveDeleteID(context.Background(), p, mock, "section")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get sections for project 1 suite 10")
		assert.Zero(t, id)
	})

	t.Run("case get cases error", func(t *testing.T) {
		mock := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
			},
			GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{{ID: 10, Name: "S1"}}, nil
			},
			GetCasesFunc: func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
				return nil, assert.AnError
			},
		}

		p := interactive.NewMockPrompter().
			WithSelectResponses(interactive.SelectResponse{Index: 0}).
			WithSelectResponses(interactive.SelectResponse{Index: 0})
		id, err := resolveDeleteID(context.Background(), p, mock, "case")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get cases for project 1 suite 10")
		assert.Zero(t, id)
	})

	t.Run("run get runs error", func(t *testing.T) {
		mock := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
			},
			GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
				return nil, assert.AnError
			},
		}

		p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
		id, err := resolveDeleteID(context.Background(), p, mock, "run")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get runs for project 1")
		assert.Zero(t, id)
	})

	t.Run("shared-step get steps error", func(t *testing.T) {
		mock := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
			},
			GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
				return nil, assert.AnError
			},
		}

		p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
		id, err := resolveDeleteID(context.Background(), p, mock, "shared-step")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get shared steps for project 1")
		assert.Zero(t, id)
	})
}

func TestSelectSharedStepID_SelectError(t *testing.T) {
	p := interactive.NewNonInteractivePrompter()
	steps := data.GetSharedStepsResponse{{ID: 101, Title: "Step X"}}

	id, err := selectSharedStepID(p, steps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select shared step")
	assert.Zero(t, id)
}

// ============= LAYER 2: delete.go missing branches =============

func TestDelete_NoArgs_NoPrompter_Error(t *testing.T) {
cmd := setupDeleteTest(t, &client.MockClient{})
// No interactive.WithPrompter in context → !HasPrompterInContext = true
cmd.SetArgs([]string{})

err := cmd.Execute()
assert.Error(t, err)
assert.Contains(t, err.Error(), "endpoint and id required")
}

func TestDelete_Interactive_ResolveIDError(t *testing.T) {
mock := &client.MockClient{
GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
return nil, assert.AnError
},
}
// Interactive mode, endpoint from prompter, then resolveDeleteID fails
p := interactive.NewMockPrompter().
WithSelectResponses(interactive.SelectResponse{Index: 0}) // select endpoint "project"

cmd := setupDeleteTest(t, mock)
cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
cmd.SetArgs([]string{"project"}) // supply endpoint, id=0 → resolveDeleteID

err := cmd.Execute()
assert.Error(t, err)
}

func TestSelectCaseID_SelectError(t *testing.T) {
p := interactive.NewNonInteractivePrompter()
cases := data.GetCasesResponse{{ID: 100, Title: "Case 1"}}

id, err := selectCaseID(context.Background(), p, cases)
assert.Error(t, err)
assert.Contains(t, err.Error(), "failed to select case")
assert.Zero(t, id)
}

func TestResolveDeleteID_SelectProjectErrors(t *testing.T) {
endpoints := []string{"suite", "section", "case", "run", "shared-step"}
for _, ep := range endpoints {
ep := ep
t.Run(ep+" get projects error", func(t *testing.T) {
mock := &client.MockClient{
GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
return nil, assert.AnError
},
}
p := interactive.NewMockPrompter()
id, err := resolveDeleteID(context.Background(), p, mock, ep)
assert.Error(t, err)
assert.Zero(t, id)
})
}
}

func TestResolveDeleteID_SelectSuiteForProjectErrors(t *testing.T) {
endpoints := []string{"section", "case"}
for _, ep := range endpoints {
ep := ep
t.Run(ep+" get suites for project error", func(t *testing.T) {
mock := &client.MockClient{
GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
},
GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
return nil, assert.AnError
},
}
p := interactive.NewMockPrompter().
WithSelectResponses(interactive.SelectResponse{Index: 0})
id, err := resolveDeleteID(context.Background(), p, mock, ep)
assert.Error(t, err)
assert.Zero(t, id)
})
}
}
