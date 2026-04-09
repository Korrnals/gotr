package migration

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

// TestMigration_MigrateSuites verifies migration behavior for suite entities
func TestMigration_MigrateSuites(t *testing.T) {
	t.Run("Successful suites migration", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
				if projectID == 1 {
					return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
				}
				return data.GetSuitesResponse{}, nil
			},
			AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
				return &data.Suite{ID: 100, Name: req.Name}, nil
			},
		}

		m := setupTestMigration(t, mock)
		m.compareField = "Name" // Override field for test

		err := m.MigrateSuites(context.Background(), false)
		assert.NoError(t, err)

		id, exists := m.mapping.GetTargetBySource(10)
		assert.True(t, exists)
		assert.Equal(t, int64(100), id)
	})
}

// TestMigration_MigrateSharedSteps verifies shared steps migration
func TestMigration_MigrateSharedSteps(t *testing.T) {
	t.Run("Migration of only unused steps", func(t *testing.T) {
		mock := &MockClient{
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				if p == 1 { // source
					return data.GetSharedStepsResponse{
						{ID: 1, Title: "Used", CaseIDs: []int64{10}},
						{ID: 2, Title: "Free", CaseIDs: []int64{}},
					}, nil
				}
				// target — CHANGE 2 to 200 HERE
				return data.GetSharedStepsResponse{
					{ID: 200, Title: "Free", CaseIDs: []int64{}},
				}, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateSharedSteps(context.Background(), false)

		assert.NoError(t, err)
		// Verify mapping: step 2 (free) should be there
		id, exists := m.mapping.GetTargetBySource(2)
		assert.True(t, exists)
		assert.Equal(t, int64(200), id)
	})

	t.Run("Error fetching source cases after successful shared steps fetch", func(t *testing.T) {
		getCasesCalls := 0
		mock := &MockClient{
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				if p == 1 {
					return data.GetSharedStepsResponse{{ID: 1, Title: "S1"}}, nil
				}
				return data.GetSharedStepsResponse{}, nil
			},
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				getCasesCalls++
				return nil, errors.New("source_cases_for_shared_steps_error")
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateSharedSteps(context.Background(), false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source_cases_for_shared_steps_error")
		assert.Equal(t, 1, getCasesCalls)
	})
}

// TestMigration_MigrateSections verifies migration behavior for sections
func TestMigration_MigrateSections(t *testing.T) {
	t.Run("Error fetching sections data", func(t *testing.T) {
		mock := &MockClient{
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, errors.New("fail to fetch sections")
			},
		}
		m := setupTestMigration(t, mock)
		err := m.MigrateSections(context.Background(), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fail to fetch sections")
	})
}

// TestMigration_MigrateCases verifies test cases migration
func TestMigration_MigrateCases(t *testing.T) {
	t.Run("Successful cases migration", func(t *testing.T) {
		mock := &MockClient{
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				if p == 1 {
					return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
				}
				return data.GetCasesResponse{}, nil
			},
			AddCaseFunc: func(ctx context.Context, suiteID int64, req *data.AddCaseRequest) (*data.Case, error) {
				return &data.Case{ID: 100, Title: req.Title}, nil
			},
		}
		m := setupTestMigration(t, mock)
		err := m.MigrateCases(context.Background(), false)
		assert.NoError(t, err)
	})
}

// TestMigration_MigrateFull verifies sequential full migration (full flow)
func TestMigration_MigrateFull(t *testing.T) {
	t.Run("Stop on error at first stage (Suites)", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{}, errors.New("suites_critical_error")
			},
			// Important: stubs for remaining methods so they don't fail
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, nil
			},
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				return data.GetSharedStepsResponse{}, nil
			},
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) { return nil, nil },
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(context.Background(), false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "suites_critical_error")
	})

	t.Run("Stop on sections error — cases not started", func(t *testing.T) {
		calledCases := false

		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{}, nil
			},
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, errors.New("sections_error")
			},
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				calledCases = true
				return data.GetCasesResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(context.Background(), false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sections_error")
		assert.False(t, calledCases, "MigrateCases should not have been called")
	})

	t.Run("Stop on shared steps error — final cases stage not started", func(t *testing.T) {
		calledCases := false

		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{}, nil
			},
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, nil
			},
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				return nil, errors.New("shared_steps_error")
			},
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				calledCases = true
				return data.GetCasesResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(context.Background(), false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shared_steps_error")
		assert.False(t, calledCases, "MigrateCases should not have been called after shared steps error")
	})

	t.Run("Error at cases stage after previous stages succeeded", func(t *testing.T) {
		getCasesCalls := 0

		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{}, nil
			},
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, nil
			},
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				return data.GetSharedStepsResponse{}, nil
			},
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				getCasesCalls++
				if getCasesCalls == 1 {
					return data.GetCasesResponse{}, nil
				}
				return nil, errors.New("cases_stage_error")
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(context.Background(), false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cases_stage_error")
		assert.GreaterOrEqual(t, getCasesCalls, 2)
	})

	t.Run("DryRun does not create entities", func(t *testing.T) {
		addSuiteCalled := false
		addSectionCalled := false
		addSharedStepCalled := false
		addCaseCalled := false

		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{{ID: 1}}, nil
			},
			AddSuiteFunc: func(ctx context.Context, p int64, r *data.AddSuiteRequest) (*data.Suite, error) {
				addSuiteCalled = true
				return nil, nil
			},
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, nil
			},
			AddSectionFunc: func(ctx context.Context, p int64, r *data.AddSectionRequest) (*data.Section, error) {
				addSectionCalled = true
				return nil, nil
			},
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				return data.GetSharedStepsResponse{}, nil
			},
			AddSharedStepFunc: func(ctx context.Context, p int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
				addSharedStepCalled = true
				return nil, nil
			},
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				return data.GetCasesResponse{}, nil
			},
			AddCaseFunc: func(ctx context.Context, s int64, r *data.AddCaseRequest) (*data.Case, error) {
				addCaseCalled = true
				return nil, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(context.Background(), true)

		assert.NoError(t, err)
		assert.False(t, addSuiteCalled, "AddSuite should not be called in dryRun")
		assert.False(t, addSectionCalled, "AddSection should not be called in dryRun")
		assert.False(t, addSharedStepCalled, "AddSharedStep should not be called in dryRun")
		assert.False(t, addCaseCalled, "AddCase should not be called in dryRun")
	})

	t.Run("Successful full migration", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{{ID: 10}}, nil
			},
			AddSuiteFunc: func(ctx context.Context, p int64, r *data.AddSuiteRequest) (*data.Suite, error) {
				return &data.Suite{ID: 100}, nil
			},
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{{ID: 20}}, nil
			},
			AddSectionFunc: func(ctx context.Context, p int64, r *data.AddSectionRequest) (*data.Section, error) {
				return &data.Section{ID: 200}, nil
			},
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				return data.GetSharedStepsResponse{{ID: 30}}, nil
			},
			AddSharedStepFunc: func(ctx context.Context, p int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
				return &data.SharedStep{ID: 300}, nil
			},
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				return data.GetCasesResponse{{ID: 40}}, nil
			},
			AddCaseFunc: func(ctx context.Context, s int64, r *data.AddCaseRequest) (*data.Case, error) {
				return &data.Case{ID: 400}, nil
			},
		}

		m := setupTestMigration(t, mock)
		err := m.MigrateFull(context.Background(), false)

		assert.NoError(t, err)
		// Can verify mapping (if counter is added — check imported*)
	})
}
