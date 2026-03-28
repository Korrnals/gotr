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
