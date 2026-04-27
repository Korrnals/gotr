package migration

import (
	"context"
	"sync"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImportCases_PreservesSourceOrderViaMoveCasesToSection guards against the
// regression: cases imported concurrently
// landed in the destination suite in a non-deterministic order because TestRail
// records cases in the order add_case calls arrive.
//
// The fix issues a single move_cases_to_section per destination section after
// the parallel phase, with case IDs sorted by their original index in the
// source slice.
func TestImportCases_PreservesSourceOrderViaMoveCasesToSection(t *testing.T) {
	const dstSectionID int64 = 200

	// 6 cases all destined for the same dst section, in deliberate source order.
	filtered := data.GetCasesResponse{
		{ID: 1, Title: "case-001", SectionID: 20},
		{ID: 2, Title: "case-002", SectionID: 20},
		{ID: 3, Title: "case-003", SectionID: 20},
		{ID: 4, Title: "case-004", SectionID: 20},
		{ID: 5, Title: "case-005", SectionID: 20},
		{ID: 6, Title: "case-006", SectionID: 20},
	}

	var moveMu sync.Mutex
	var moveCalls []*data.MoveCasesRequest
	var moveSections []int64

	mock := &MockClient{
		// Source/destination section listings — used by resolveSectionMapByName.
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 20, Name: "S"}}, nil
		},
		// AddCase deterministically assigns newID = srcID*10 so we can verify
		// the post-import reorder request carries IDs in source order even
		// though the goroutines themselves complete in arbitrary order.
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			// Title encodes source order: case-001 .. case-006.
			// Map back to srcID via Title to avoid coupling to call order.
			var srcID int64
			switch req.Title {
			case "case-001":
				srcID = 1
			case "case-002":
				srcID = 2
			case "case-003":
				srcID = 3
			case "case-004":
				srcID = 4
			case "case-005":
				srcID = 5
			case "case-006":
				srcID = 6
			}
			return &data.Case{ID: srcID * 10, Title: req.Title}, nil
		},
		MoveCasesToSectionFunc: func(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error {
			moveMu.Lock()
			defer moveMu.Unlock()
			moveSections = append(moveSections, sectionID)
			cp := *req
			cp.CaseIDs = append([]int64(nil), req.CaseIDs...)
			moveCalls = append(moveCalls, &cp)
			return nil
		},
	}

	m := setupTestMigration(t, mock)
	// Pre-populate section mapping so resolveDestinationSectionID returns dstSectionID.
	m.mapping.AddPair(20, dstSectionID, "existing")

	require.NoError(t, m.ImportCases(context.Background(), filtered, false))

	require.Len(t, moveCalls, 1, "exactly one move_cases_to_section call expected (single dst section)")
	assert.Equal(t, dstSectionID, moveSections[0])
	assert.Equal(t, []int64{10, 20, 30, 40, 50, 60}, moveCalls[0].CaseIDs,
		"case IDs in move_cases_to_section must follow the source order")
}

// TestImportCasesReport_PreservesSourceOrder mirrors the regression check for
// the report variant used by sync_full and the migration report writer.
func TestImportCasesReport_PreservesSourceOrder(t *testing.T) {
	const dstSectionID int64 = 200

	filtered := data.GetCasesResponse{
		{ID: 11, Title: "x-001", SectionID: 21},
		{ID: 12, Title: "x-002", SectionID: 21},
		{ID: 13, Title: "x-003", SectionID: 21},
	}

	var moveMu sync.Mutex
	var moveCalls []*data.MoveCasesRequest

	mock := &MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 21, Name: "S2"}}, nil
		},
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			switch req.Title {
			case "x-001":
				return &data.Case{ID: 1001, Title: req.Title}, nil
			case "x-002":
				return &data.Case{ID: 1002, Title: req.Title}, nil
			case "x-003":
				return &data.Case{ID: 1003, Title: req.Title}, nil
			}
			return &data.Case{ID: 0}, nil
		},
		MoveCasesToSectionFunc: func(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error {
			moveMu.Lock()
			defer moveMu.Unlock()
			cp := *req
			cp.CaseIDs = append([]int64(nil), req.CaseIDs...)
			moveCalls = append(moveCalls, &cp)
			return nil
		},
	}

	m := setupTestMigration(t, mock)
	m.mapping.AddPair(21, dstSectionID, "existing")

	createdIDs, errs, err := m.ImportCasesReport(context.Background(), filtered, false)
	require.NoError(t, err)
	assert.Empty(t, errs)
	assert.Len(t, createdIDs, 3)

	require.Len(t, moveCalls, 1)
	assert.Equal(t, []int64{1001, 1002, 1003}, moveCalls[0].CaseIDs)
}

// TestImportCases_SkipsReorderForSingleCaseSection ensures we don't issue a
// pointless move_cases_to_section for sections that received only one case.
func TestImportCases_SkipsReorderForSingleCaseSection(t *testing.T) {
	filtered := data.GetCasesResponse{
		{ID: 1, Title: "solo", SectionID: 20},
	}

	moveCalled := false
	mock := &MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 20, Name: "S"}}, nil
		},
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			return &data.Case{ID: 999, Title: req.Title}, nil
		},
		MoveCasesToSectionFunc: func(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error {
			moveCalled = true
			return nil
		},
	}

	m := setupTestMigration(t, mock)
	m.mapping.AddPair(20, 200, "existing")

	require.NoError(t, m.ImportCases(context.Background(), filtered, false))
	assert.False(t, moveCalled, "move_cases_to_section must not be called for a 1-case section")
}

// TestImportCases_ReorderErrorIsNonFatal ensures that a failing
// move_cases_to_section does not abort the migration: the data is already
// imported, ordering is a UX concern.
func TestImportCases_ReorderErrorIsNonFatal(t *testing.T) {
	filtered := data.GetCasesResponse{
		{ID: 1, Title: "a", SectionID: 20},
		{ID: 2, Title: "b", SectionID: 20},
	}

	mock := &MockClient{
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 20, Name: "S"}}, nil
		},
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			if req.Title == "a" {
				return &data.Case{ID: 1}, nil
			}
			return &data.Case{ID: 2}, nil
		},
		MoveCasesToSectionFunc: func(ctx context.Context, sectionID int64, req *data.MoveCasesRequest) error {
			return assert.AnError
		},
	}

	m := setupTestMigration(t, mock)
	m.mapping.AddPair(20, 200, "existing")

	// Import must still succeed even though reorder failed.
	assert.NoError(t, m.ImportCases(context.Background(), filtered, false))
}
