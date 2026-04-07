package sync

// wave6g_target_coverage_test.go — targeted tests to push cmd/sync coverage to 92%+
// Covers: interactive‑selection error paths, FetchData errors, empty‑filtered paths,
//         Confirm errors, Import errors, and migration‑init errors.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service/migration"

	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────
// sync_cases.go – uncovered branches
// ─────────────────────────────────────────────────────────────────

// TestWave6G_Cases_SrcSuiteZero_SelectError covers the
// "srcSuite==0 → SelectSuiteForProject fails" error path (line ~75).
func TestWave6G_Cases_SrcSuiteZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	// srcSuite intentionally left at 0 → triggers interactive SelectSuiteForProject
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	// MockPrompter with no Select responses → "mock select queue exhausted" on suite select
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Cases_DstProjectZero_SelectError covers the
// "dstProject==0 → SelectProject fails" error path (line ~83).
func TestWave6G_Cases_DstProjectZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 2, Name: "P2"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	// dstProject left at 0 → triggers interactive SelectProject
	cmd.Flags().Set("dst-suite", "20")
	p := interactive.NewMockPrompter() // no Select responses → error
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Cases_DstSuiteZero_SelectError covers the
// "dstSuite==0 → SelectSuiteForProject fails" error path (line ~91).
func TestWave6G_Cases_DstSuiteZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20, Name: "Suite 2"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	// dstSuite left at 0 → triggers interactive SelectSuiteForProject for dstProject
	p := interactive.NewMockPrompter() // no Select responses → error
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Cases_FetchCasesError covers the FetchCasesData error paths
// (inner callback `if err != nil` ≈ line 138 and outer ≈ line 149).
func TestWave6G_Cases_FetchCasesError(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return nil, errors.New("api_cases_error")
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_cases_error")
}

// TestWave6G_Cases_MatchesBranch covers the "matches" branch where a source case
// already has a matching title in the destination (line ~167).
func TestWave6G_Cases_MatchesBranch(t *testing.T) {
	// Both src and dst return a case with the same title → FilterCases excludes it
	// from filtered → it ends up in matches slice → line 167 is hit.
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: projectID*10 + suiteID, Title: "SharedTitle"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "true") // skip confirm

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────
// sync_full.go – uncovered branches
// ─────────────────────────────────────────────────────────────────

// TestWave6G_Full_SrcSuiteZero_SelectError covers srcSuite==0 interactive error
// in the full migration command (line ~68).
func TestWave6G_Full_SrcSuiteZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	// srcSuite left at 0 → follows interactive path
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	p := interactive.NewMockPrompter() // no Select → exhausted error
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Full_DstProjectZero_SelectError covers dstProject==0 path (line ~76).
func TestWave6G_Full_DstProjectZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 2, Name: "P2"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	// dstProject left at 0 → interactive SelectProject called
	cmd.Flags().Set("dst-suite", "20")
	p := interactive.NewMockPrompter() // no Select → error
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Full_DstSuiteZero_SelectError covers dstSuite==0 path (line ~82).
func TestWave6G_Full_DstSuiteZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20, Name: "Suite 2"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	// dstSuite left at 0
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Full_MigrateSharedStepsError covers the error branch after
// runSyncStatus wrapping MigrateSharedSteps (line ~113).
func TestWave6G_Full_MigrateSharedStepsError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, errors.New("shared_steps_fetch_error")
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("approve", "true") // auto-approve to pass straight through

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shared_steps_fetch_error")
}

// ─────────────────────────────────────────────────────────────────
// sync_sections.go – uncovered branches
// ─────────────────────────────────────────────────────────────────

// TestWave6G_Sections_SrcSuiteZero_SelectError covers srcSuite==0 path (line ~71).
func TestWave6G_Sections_SrcSuiteZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	// srcSuite left at 0
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Sections_DstProjectZero_SelectError covers dstProject==0 path (line ~79).
func TestWave6G_Sections_DstProjectZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 2, Name: "P2"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	// dstProject left at 0
	cmd.Flags().Set("dst-suite", "20")
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_Sections_FetchSectionsError covers FetchSectionsData error path
// (inner callback ~line 104 and outer ~line 115).
func TestWave6G_Sections_FetchSectionsError(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return nil, errors.New("sections_fetch_error")
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sections_fetch_error")
}

// TestWave6G_Sections_NoNewSections covers the "len(filtered)==0 → No new sections"
// path — source and destination share the same section name.
func TestWave6G_Sections_NoNewSections(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: projectID * 100, Name: "Login"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	// dryRun=false so we reach the len(filtered)==0 gate
	cmd.Flags().Set("dry-run", "false")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

// TestWave6G_Sections_ConfirmError covers the Confirm error path when autoApprove
// is false and Confirm exhausts its mock queue (line ~135).
func TestWave6G_Sections_ConfirmError(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "Unique Section"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
		AddSectionFunc: func(ctx context.Context, projectID int64, r *data.AddSectionRequest) (*data.Section, error) {
			return &data.Section{ID: 200, Name: r.Name}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	// autoApprove=false (default) → p.Confirm called; empty queue → error
	p := interactive.NewMockPrompter() // no confirm responses
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}

// TestWave6G_Sections_ConfirmDeclined covers the !ok branch when the user
// declines the import confirmation (line ~136-138 in sections).
func TestWave6G_Sections_ConfirmDeclined(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return data.GetSectionsResponse{{ID: 11, Name: "New Section"}}, nil
			}
			return data.GetSectionsResponse{}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	// false → user declines → ui.Cancelled + return nil
	p := interactive.NewMockPrompter().WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

// TestWave6G_Sections_MigrationInitError covers newMigration() error path (line ~89).
func TestWave6G_Sections_MigrationInitError(t *testing.T) {
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = func(_ client.ClientInterface, _, _, _, _ int64, _, _ string) (*migration.Migration, error) {
		return nil, errors.New("migration_init_failed")
	}

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, &client.MockClient{})
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration_init_failed")
}

// ─────────────────────────────────────────────────────────────────
// sync_shared_steps.go – uncovered branches
// ─────────────────────────────────────────────────────────────────

// TestWave6G_SharedSteps_SrcProjectZero_SelectError covers srcProject==0 error
// path (line ~66).
func TestWave6G_SharedSteps_SrcProjectZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "P1"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	// srcProject left at 0 → must select interactively; no Select queued → error
	cmd.Flags().Set("dst-project", "2")
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_SharedSteps_SrcSuiteConfirmYes_SelectError covers the path where
// srcSuite==0 and user confirms "Specify source suite?" → SelectSuiteForProject
// fails (line ~71).
func TestWave6G_SharedSteps_SrcSuiteConfirmYes_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 10, Name: "Suite 1"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	// srcSuite left at 0 → Confirm("Specify source suite?") queued as true
	// → SelectSuiteForProject called → p.Select exhausted → error at line ~71
	cmd.Flags().Set("dst-project", "2")
	p := interactive.NewMockPrompter().WithConfirmResponses(true) // specifySuite=true
	// no Select responses → Select exhausted
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_SharedSteps_DstProjectZero_SelectError covers dstProject==0 path
// (line ~80).
func TestWave6G_SharedSteps_DstProjectZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 2, Name: "P2"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	// dstProject left at 0
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_SharedSteps_FetchSharedStepsError covers FetchSharedStepsData error
// (inner callback ~line 106 and outer ~line 117).
func TestWave6G_SharedSteps_FetchSharedStepsError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, errors.New("shared_steps_api_error")
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shared_steps_api_error")
}

// TestWave6G_SharedSteps_GetCasesError covers GetCases error when loading source
// cases after shared steps are fetched (line ~128 and ~132).
func TestWave6G_SharedSteps_GetCasesError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{{ID: 1, Title: "Step A"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return nil, errors.New("cases_for_sharedsteps_error")
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cases_for_sharedsteps_error")
}

// TestWave6G_SharedSteps_NoNewSharedSteps covers the "filtered empty → No new shared
// steps" path — same step title in source and destination.
// Also covers loop body (line 132) since GetCasesFunc returns a case.
func TestWave6G_SharedSteps_NoNewSharedSteps(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{{ID: projectID * 10, Title: "Shared Login"}}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			// Return a case so the loop body at line ~132 is covered.
			return data.GetCasesResponse{{ID: 1, Title: "SomeCase"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "false")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

// TestWave6G_SharedSteps_ConfirmError covers the Confirm error path when the
// confirm queue is exhausted (line analogous to sections ~135).
func TestWave6G_SharedSteps_ConfirmError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "UniqueStep"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			return &data.SharedStep{ID: 100, Title: r.Title}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	// autoApprove=false (default) + empty confirm queue → exhausted error
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}

// TestWave6G_SharedSteps_ConfirmDeclined covers the !ok branch when the user
// declines the import confirmation.
func TestWave6G_SharedSteps_ConfirmDeclined(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "UniqueStep"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	p := interactive.NewMockPrompter().WithConfirmResponses(false) // decline
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────
// sync_suites.go – uncovered branches
// ─────────────────────────────────────────────────────────────────

// TestWave6G_Suites_MigrationInitError covers newMigration() error (line ~60).
func TestWave6G_Suites_MigrationInitError(t *testing.T) {
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = func(_ client.ClientInterface, _, _, _, _ int64, _, _ string) (*migration.Migration, error) {
		return nil, errors.New("suites_migration_init_error")
	}

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, &client.MockClient{})
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suites_migration_init_error")
}

// TestWave6G_Suites_FetchSuitesError covers FetchSuitesData error paths
// (inner callback ~line 74 and outer ~line 85).
func TestWave6G_Suites_FetchSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, errors.New("suites_fetch_error")
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suites_fetch_error")
}

// TestWave6G_Suites_NoNewSuites covers the "len(filtered)==0 → No new suites" path
// (~line 103-106) — source and destination share the same suite name.
func TestWave6G_Suites_NoNewSuites(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: projectID * 10, Name: "Same Suite"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "false") // reach len==0 gate without dry-run short-circuit

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

// TestWave6G_Suites_ConfirmError covers the Confirm error branch (~line 103 confirm
// error) when the mock confirm queue is exhausted.
func TestWave6G_Suites_ConfirmError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Unique Suite"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, r *data.AddSuiteRequest) (*data.Suite, error) {
			return &data.Suite{ID: 100, Name: r.Name}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	// autoApprove=false (default), filtered non-empty, empty confirm queue → error
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}

// TestWave6G_Suites_ConfirmDeclined covers the !ok path when the user declines
// the import confirmation (line ~115-118 equivalent in suites).
func TestWave6G_Suites_ConfirmDeclined(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return data.GetSuitesResponse{{ID: 10, Name: "Unique Suite"}}, nil
			}
			return data.GetSuitesResponse{}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	p := interactive.NewMockPrompter().WithConfirmResponses(false) // decline
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────
// Additional targeted tests for remaining uncovered blocks
// ─────────────────────────────────────────────────────────────────

// TestWave6G_Sections_DstSuiteZero_SelectError covers the dstSuite==0 interactive
// error path in sections (~line 79).
func TestWave6G_Sections_DstSuiteZero_SelectError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 20, Name: "Suite 2"}}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	// dstSuite left at 0 → triggers SelectSuiteForProject
	p := interactive.NewMockPrompter() // no Select → exhausted error
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

// TestWave6G_SharedSteps_SrcSuiteConfirmError covers the p.Confirm error when
// srcSuite==0 and the confirm queue is empty (line ~66).
func TestWave6G_SharedSteps_SrcSuiteConfirmError(t *testing.T) {
	mock := &client.MockClient{}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	// srcSuite left at 0 → p.Confirm("Specify source suite?") called
	// empty confirm queue → "mock confirm queue exhausted" error
	cmd.Flags().Set("dst-project", "2")
	p := interactive.NewMockPrompter() // no confirm responses
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}

// TestWave6G_SharedSteps_MigrationInitError covers newMigration() error for
// shared-steps command (~line 92).
func TestWave6G_SharedSteps_MigrationInitError(t *testing.T) {
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = func(_ client.ClientInterface, _, _, _, _ int64, _, _ string) (*migration.Migration, error) {
		return nil, errors.New("sharedsteps_migration_init_error")
	}

	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, &client.MockClient{})
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sharedsteps_migration_init_error")
}

// TestWave6G_Full_MigrationInitError covers newMigration() error for the full
// migration command (~line 86).
func TestWave6G_Full_MigrationInitError(t *testing.T) {
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = func(_ client.ClientInterface, _, _, _, _ int64, _, _ string) (*migration.Migration, error) {
		return nil, errors.New("full_migration_init_error")
	}

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, &client.MockClient{})
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "full_migration_init_error")
}

// TestWave6G_Full_MigrateCasesError covers the MigrateCases error path (~line 113).
// SharedSteps fetches nothing so MigrateSharedSteps succeeds trivially, then
// GetCasesFunc errors → runSyncStatus returns error at the MigrateCases phase.
func TestWave6G_Full_MigrateCasesError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return data.GetSharedStepsResponse{}, nil // empty → migrate succeeds
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			// MigrateSharedSteps calls GetCases(srcProject=1,...) → succeed.
			// MigrateCases FetchCasesData calls GetCases(dstProject=2,...) → error.
			if projectID == 2 {
				return nil, errors.New("cases_step2_error")
			}
			return data.GetCasesResponse{}, nil
		},
	}
	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("approve", "true")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cases_step2_error")
}

// ─────────────────────────────────────────────────────────────────
// EnsureLogsDirPath error paths — covers the if-err-return-err
// block after paths.EnsureLogsDirPath() in each sync subcommand.
// Trigger: set HOME to a regular file so os.MkdirAll fails.
// ─────────────────────────────────────────────────────────────────

// blockHomeDir sets HOME to a regular file so paths.EnsureLogsDirPath fails.
func blockHomeDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", blocker)
}

// TestWave6G_Cases_EnsureLogsDirPath_Error covers sync_cases.go ~line 97.
func TestWave6G_Cases_EnsureLogsDirPath_Error(t *testing.T) {
	blockHomeDir(t)
	mock := &client.MockClient{}
	resetCasesFlags()
	cmd := casesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logs directory")
}

// TestWave6G_Full_EnsureLogsDirPath_Error covers sync_full.go ~line 81.
func TestWave6G_Full_EnsureLogsDirPath_Error(t *testing.T) {
	blockHomeDir(t)
	mock := &client.MockClient{}
	resetFullFlags()
	cmd := fullCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logs directory")
}

// TestWave6G_Sections_EnsureLogsDirPath_Error covers sync_sections.go ~line 84.
func TestWave6G_Sections_EnsureLogsDirPath_Error(t *testing.T) {
	blockHomeDir(t)
	mock := &client.MockClient{}
	resetSectionsFlags()
	cmd := sectionsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logs directory")
}

// TestWave6G_SharedSteps_EnsureLogsDirPath_Error covers sync_shared_steps.go ~line 86.
func TestWave6G_SharedSteps_EnsureLogsDirPath_Error(t *testing.T) {
	blockHomeDir(t)
	mock := &client.MockClient{}
	resetSharedStepsFlags()
	cmd := sharedStepsCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logs directory")
}

// TestWave6G_Suites_EnsureLogsDirPath_Error covers sync_suites.go ~line 55.
func TestWave6G_Suites_EnsureLogsDirPath_Error(t *testing.T) {
	blockHomeDir(t)
	mock := &client.MockClient{}
	resetSuitesFlags()
	cmd := suitesCmd
	SetTestClient(cmd, mock)
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logs directory")
}

