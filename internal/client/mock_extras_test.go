// internal/client/mock_extras_test.go
// Targeted tests for the remaining 0%-coverage mock methods:
//
//	GetCasesPage, AddCaseField, DeleteSection, DeleteSharedStep, GetCasesParallelCtx
package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GetCasesPage
// ---------------------------------------------------------------------------

func TestMock_GetCasesPage_nil(t *testing.T) {
	m := &MockClient{}
	got, err := m.GetCasesPage(context.Background(), 30, 1, 0, 250)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestMock_GetCasesPage_func(t *testing.T) {
	want := data.GetCasesResponse{{ID: 7, Title: "T"}}
	m := &MockClient{
		GetCasesPageFunc: func(_ context.Context, projectID int64, suiteID int64, offset int, limit int) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(1), suiteID)
			assert.Equal(t, 0, offset)
			assert.Equal(t, 250, limit)
			return want, nil
		},
	}
	got, err := m.GetCasesPage(context.Background(), 30, 1, 0, 250)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMock_GetCasesPage_error(t *testing.T) {
	boom := errors.New("page error")
	m := &MockClient{
		GetCasesPageFunc: func(_ context.Context, _ int64, _ int64, _ int, _ int) (data.GetCasesResponse, error) {
			return nil, boom
		},
	}
	_, err := m.GetCasesPage(context.Background(), 30, 1, 0, 250)
	assert.ErrorIs(t, err, boom)
}

// ---------------------------------------------------------------------------
// AddCaseField
// ---------------------------------------------------------------------------

func TestMock_AddCaseField_nil(t *testing.T) {
	m := &MockClient{}
	got, err := m.AddCaseField(context.Background(), &data.AddCaseFieldRequest{Name: "F"})
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestMock_AddCaseField_func(t *testing.T) {
	want := &data.AddCaseFieldResponse{ID: 42, Name: "F"}
	m := &MockClient{
		AddCaseFieldFunc: func(_ context.Context, req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error) {
			assert.Equal(t, "F", req.Name)
			return want, nil
		},
	}
	got, err := m.AddCaseField(context.Background(), &data.AddCaseFieldRequest{Name: "F"})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMock_AddCaseField_error(t *testing.T) {
	boom := errors.New("add case field error")
	m := &MockClient{
		AddCaseFieldFunc: func(_ context.Context, _ *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error) {
			return nil, boom
		},
	}
	_, err := m.AddCaseField(context.Background(), &data.AddCaseFieldRequest{})
	assert.ErrorIs(t, err, boom)
}

// ---------------------------------------------------------------------------
// DeleteSection
// ---------------------------------------------------------------------------

func TestMock_DeleteSection_nil(t *testing.T) {
	m := &MockClient{}
	err := m.DeleteSection(context.Background(), 99)
	assert.NoError(t, err)
}

func TestMock_DeleteSection_func(t *testing.T) {
	m := &MockClient{
		DeleteSectionFunc: func(_ context.Context, sectionID int64) error {
			assert.Equal(t, int64(55), sectionID)
			return nil
		},
	}
	err := m.DeleteSection(context.Background(), 55)
	assert.NoError(t, err)
}

func TestMock_DeleteSection_error(t *testing.T) {
	boom := errors.New("del section error")
	m := &MockClient{
		DeleteSectionFunc: func(_ context.Context, _ int64) error {
			return boom
		},
	}
	err := m.DeleteSection(context.Background(), 55)
	assert.ErrorIs(t, err, boom)
}

// ---------------------------------------------------------------------------
// DeleteSharedStep
// ---------------------------------------------------------------------------

func TestMock_DeleteSharedStep_nil(t *testing.T) {
	m := &MockClient{}
	err := m.DeleteSharedStep(context.Background(), 10, 0)
	assert.NoError(t, err)
}

func TestMock_DeleteSharedStep_func(t *testing.T) {
	m := &MockClient{
		DeleteSharedStepFunc: func(_ context.Context, stepID int64, keepInCases int) error {
			assert.Equal(t, int64(10), stepID)
			assert.Equal(t, 1, keepInCases)
			return nil
		},
	}
	err := m.DeleteSharedStep(context.Background(), 10, 1)
	assert.NoError(t, err)
}

func TestMock_DeleteSharedStep_error(t *testing.T) {
	boom := errors.New("del shared step error")
	m := &MockClient{
		DeleteSharedStepFunc: func(_ context.Context, _ int64, _ int) error {
			return boom
		},
	}
	err := m.DeleteSharedStep(context.Background(), 10, 0)
	assert.ErrorIs(t, err, boom)
}

// ---------------------------------------------------------------------------
// GetCasesParallelCtx
// ---------------------------------------------------------------------------

func TestMock_GetCasesParallelCtx_nil_func_default_path(t *testing.T) {
	// No GetCasesParallelCtxFunc, no GetCasesFunc → falls through to default delegate path
	m := &MockClient{}
	cases, result, err := m.GetCasesParallelCtx(context.Background(), 30, []int64{1, 2}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, cases)
}

func TestMock_GetCasesParallelCtx_success_func(t *testing.T) {
	want := data.GetCasesResponse{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}}
	wantResult := &concurrency.ExecutionResult{Cases: want}
	m := &MockClient{
		GetCasesParallelCtxFunc: func(_ context.Context, projectID int64, suiteIDs []int64, cfg *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, []int64{1, 2}, suiteIDs)
			return want, wantResult, nil
		},
	}
	cases, result, err := m.GetCasesParallelCtx(context.Background(), 30, []int64{1, 2}, &concurrency.ControllerConfig{MaxConcurrentSuites: 3})
	require.NoError(t, err)
	assert.Equal(t, want, cases)
	assert.Equal(t, wantResult, result)
}

func TestMock_GetCasesParallelCtx_error_func(t *testing.T) {
	boom := errors.New("ctx parallel error")
	m := &MockClient{
		GetCasesParallelCtxFunc: func(_ context.Context, _ int64, _ []int64, _ *concurrency.ControllerConfig) (data.GetCasesResponse, *concurrency.ExecutionResult, error) {
			return nil, nil, boom
		},
	}
	_, _, err := m.GetCasesParallelCtx(context.Background(), 30, []int64{1}, nil)
	assert.ErrorIs(t, err, boom)
}

func TestMock_GetCasesParallelCtx_with_config_workers(t *testing.T) {
	// exercises the config.MaxConcurrentSuites branch in the default path
	m := &MockClient{
		GetCasesFunc: func(_ context.Context, _ int64, suiteID int64, _ int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: suiteID}}, nil
		},
	}
	cfg := &concurrency.ControllerConfig{MaxConcurrentSuites: 2}
	cases, result, err := m.GetCasesParallelCtx(context.Background(), 30, []int64{4, 5}, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, cases, 2)
}

func TestMock_DefaultBranches_ProjectsAndExtended(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()

	projects, err := m.GetProjects(ctx)
	assert.NoError(t, err)
	assert.Nil(t, projects)

	project, err := m.GetProject(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, project)

	addedProject, err := m.AddProject(ctx, &data.AddProjectRequest{Name: "p"})
	assert.NoError(t, err)
	assert.Nil(t, addedProject)

	updatedProject, err := m.UpdateProject(ctx, 1, &data.UpdateProjectRequest{Name: "p2"})
	assert.NoError(t, err)
	assert.Nil(t, updatedProject)

	assert.NoError(t, m.DeleteProject(ctx, 1))

	gotCase, err := m.GetCase(ctx, 11)
	assert.NoError(t, err)
	assert.Nil(t, gotCase)

	addedCase, err := m.AddCase(ctx, 10, &data.AddCaseRequest{Title: "x"})
	assert.NoError(t, err)
	assert.NotNil(t, addedCase)
	assert.Equal(t, int64(999), addedCase.ID)

	updatedCase, err := m.UpdateCase(ctx, 11, &data.UpdateCaseRequest{})
	assert.NoError(t, err)
	assert.Nil(t, updatedCase)
	assert.NoError(t, m.DeleteCase(ctx, 11))

	updatedCases, err := m.UpdateCases(ctx, 7, &data.UpdateCasesRequest{})
	assert.NoError(t, err)
	assert.Nil(t, updatedCases)
	assert.NoError(t, m.DeleteCases(ctx, 7, &data.DeleteCasesRequest{}))
	assert.NoError(t, m.CopyCasesToSection(ctx, 7, &data.CopyCasesRequest{}))
	assert.NoError(t, m.MoveCasesToSection(ctx, 7, &data.MoveCasesRequest{}))

	history, err := m.GetHistoryForCase(ctx, 11)
	assert.NoError(t, err)
	assert.Nil(t, history)

	fields, err := m.GetCaseFields(ctx)
	assert.NoError(t, err)
	assert.Nil(t, fields)

	types, err := m.GetCaseTypes(ctx)
	assert.NoError(t, err)
	assert.Nil(t, types)

	diff, err := m.DiffCasesData(ctx, 1, 2, "title")
	assert.NoError(t, err)
	assert.Nil(t, diff)

	groupList, err := m.GetGroups(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, groupList)

	group, err := m.GetGroup(ctx, 2)
	assert.NoError(t, err)
	assert.Nil(t, group)

	newGroup, err := m.AddGroup(ctx, 1, "g", []int64{1, 2})
	assert.NoError(t, err)
	assert.Nil(t, newGroup)

	updatedGroup, err := m.UpdateGroup(ctx, 2, "g2", []int64{3})
	assert.NoError(t, err)
	assert.Nil(t, updatedGroup)
	assert.NoError(t, m.DeleteGroup(ctx, 2))

	roles, err := m.GetRoles(ctx)
	assert.NoError(t, err)
	assert.Nil(t, roles)

	role, err := m.GetRole(ctx, 3)
	assert.NoError(t, err)
	assert.Nil(t, role)

	datasets, err := m.GetDatasets(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, datasets)

	dataset, err := m.GetDataset(ctx, 5)
	assert.NoError(t, err)
	assert.Nil(t, dataset)

	addedDataset, err := m.AddDataset(ctx, 1, "d")
	assert.NoError(t, err)
	assert.Nil(t, addedDataset)

	updatedDataset, err := m.UpdateDataset(ctx, 5, "d2")
	assert.NoError(t, err)
	assert.Nil(t, updatedDataset)
	assert.NoError(t, m.DeleteDataset(ctx, 5))

	variables, err := m.GetVariables(ctx, 6)
	assert.NoError(t, err)
	assert.Nil(t, variables)

	addedVariable, err := m.AddVariable(ctx, 6, "v")
	assert.NoError(t, err)
	assert.Nil(t, addedVariable)

	updatedVariable, err := m.UpdateVariable(ctx, 7, "v2")
	assert.NoError(t, err)
	assert.Nil(t, updatedVariable)
	assert.NoError(t, m.DeleteVariable(ctx, 7))

	bdd, err := m.GetBDD(ctx, 11)
	assert.NoError(t, err)
	assert.Nil(t, bdd)

	addedBDD, err := m.AddBDD(ctx, 11, "feature")
	assert.NoError(t, err)
	assert.Nil(t, addedBDD)

	labels, err := m.GetLabels(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, labels)

	label, err := m.GetLabel(ctx, 8)
	assert.NoError(t, err)
	assert.Nil(t, label)

	updatedLabel, err := m.UpdateLabel(ctx, 8, data.UpdateLabelRequest{ProjectID: 1, Title: "x"})
	assert.NoError(t, err)
	assert.Nil(t, updatedLabel)

	assert.NoError(t, m.UpdateTestLabels(ctx, 9, []string{"a"}))
	assert.NoError(t, m.UpdateTestsLabels(ctx, 10, []int64{1, 2}, []string{"b"}))
}

func TestMock_FunctionBranches_ProjectsAndExtended(t *testing.T) {
	ctx := context.Background()
	m := &MockClient{
		GetProjectsFunc: func(context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P"}}, nil
		},
		GetProjectFunc: func(context.Context, int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: 1, Name: "P"}, nil
		},
		AddProjectFunc: func(context.Context, *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: 2, Name: "P2"}, nil
		},
		UpdateProjectFunc: func(context.Context, int64, *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: 2, Name: "P2"}, nil
		},
		DeleteProjectFunc: func(context.Context, int64) error { return nil },

		GetGroupFunc: func(context.Context, int64) (*data.Group, error) { return &data.Group{ID: 1, Name: "G"}, nil },
		AddGroupFunc: func(context.Context, int64, string, []int64) (*data.Group, error) { return &data.Group{ID: 2, Name: "G2"}, nil },
		UpdateGroupFunc: func(context.Context, int64, string, []int64) (*data.Group, error) { return &data.Group{ID: 2, Name: "G2"}, nil },
		DeleteGroupFunc: func(context.Context, int64) error { return nil },

		GetRoleFunc: func(context.Context, int64) (*data.Role, error) { return &data.Role{ID: 1, Name: "R"}, nil },

		GetDatasetFunc:    func(context.Context, int64) (*data.Dataset, error) { return &data.Dataset{ID: 1, Name: "D"}, nil },
		AddDatasetFunc:    func(context.Context, int64, string) (*data.Dataset, error) { return &data.Dataset{ID: 2, Name: "D2"}, nil },
		UpdateDatasetFunc: func(context.Context, int64, string) (*data.Dataset, error) { return &data.Dataset{ID: 2, Name: "D2"}, nil },
		DeleteDatasetFunc: func(context.Context, int64) error { return nil },

		AddVariableFunc:    func(context.Context, int64, string) (*data.Variable, error) { return &data.Variable{ID: 1, Name: "V"}, nil },
		UpdateVariableFunc: func(context.Context, int64, string) (*data.Variable, error) { return &data.Variable{ID: 2, Name: "V2"}, nil },
		DeleteVariableFunc: func(context.Context, int64) error { return nil },

		AddBDDFunc: func(context.Context, int64, string) (*data.BDD, error) { return &data.BDD{ID: 1, Content: "bdd"}, nil },

		GetLabelsFunc:         func(context.Context, int64) (data.GetLabelsResponse, error) { return data.GetLabelsResponse{{ID: 1, Name: "L"}}, nil },
		GetLabelFunc:          func(context.Context, int64) (*data.Label, error) { return &data.Label{ID: 2, Name: "L2"}, nil },
		UpdateLabelFunc:       func(context.Context, int64, data.UpdateLabelRequest) (*data.Label, error) { return &data.Label{ID: 2, Name: "L2"}, nil },
		UpdateTestLabelsFunc:  func(context.Context, int64, []string) error { return nil },
		UpdateTestsLabelsFunc: func(context.Context, int64, []int64, []string) error { return nil },
	}

	projects, err := m.GetProjects(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, 1)

	project, err := m.GetProject(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, project)

	addedProject, err := m.AddProject(ctx, &data.AddProjectRequest{Name: "P2"})
	require.NoError(t, err)
	assert.NotNil(t, addedProject)

	updatedProject, err := m.UpdateProject(ctx, 1, &data.UpdateProjectRequest{Name: "P3"})
	require.NoError(t, err)
	assert.NotNil(t, updatedProject)
	require.NoError(t, m.DeleteProject(ctx, 1))

	group, err := m.GetGroup(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, group)
	addedGroup, err := m.AddGroup(ctx, 1, "G2", []int64{1})
	require.NoError(t, err)
	assert.NotNil(t, addedGroup)
	updatedGroup, err := m.UpdateGroup(ctx, 2, "G3", []int64{2})
	require.NoError(t, err)
	assert.NotNil(t, updatedGroup)
	require.NoError(t, m.DeleteGroup(ctx, 2))

	role, err := m.GetRole(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, role)

	dataset, err := m.GetDataset(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, dataset)
	addedDataset, err := m.AddDataset(ctx, 1, "D2")
	require.NoError(t, err)
	assert.NotNil(t, addedDataset)
	updatedDataset, err := m.UpdateDataset(ctx, 1, "D3")
	require.NoError(t, err)
	assert.NotNil(t, updatedDataset)
	require.NoError(t, m.DeleteDataset(ctx, 1))

	addedVar, err := m.AddVariable(ctx, 1, "V")
	require.NoError(t, err)
	assert.NotNil(t, addedVar)
	updatedVar, err := m.UpdateVariable(ctx, 1, "V2")
	require.NoError(t, err)
	assert.NotNil(t, updatedVar)
	require.NoError(t, m.DeleteVariable(ctx, 1))

	bdd, err := m.AddBDD(ctx, 1, "bdd")
	require.NoError(t, err)
	assert.NotNil(t, bdd)

	labels, err := m.GetLabels(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, labels, 1)
	label, err := m.GetLabel(ctx, 2)
	require.NoError(t, err)
	assert.NotNil(t, label)
	updatedLabel, err := m.UpdateLabel(ctx, 2, data.UpdateLabelRequest{ProjectID: 1, Title: "L3"})
	require.NoError(t, err)
	assert.NotNil(t, updatedLabel)
	require.NoError(t, m.UpdateTestLabels(ctx, 1, []string{"a"}))
	require.NoError(t, m.UpdateTestsLabels(ctx, 1, []int64{1}, []string{"a"}))
}
