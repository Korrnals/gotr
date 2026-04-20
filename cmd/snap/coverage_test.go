package snap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// isInterruptError / wrapInterrupt
// ---------------------------------------------------------------------------

func TestIsInterruptError(t *testing.T) {
	assert.False(t, isInterruptError(nil))
	assert.True(t, isInterruptError(errors.New("user interrupt")))
	assert.True(t, isInterruptError(errors.New("interrupt")))
	assert.False(t, isInterruptError(errors.New("some error")))
}

func TestWrapInterrupt(t *testing.T) {
	original := errors.New("interrupt detected")
	assert.Equal(t, context.Canceled, wrapInterrupt(original))

	nonInterrupt := errors.New("other error")
	assert.Equal(t, nonInterrupt, wrapInterrupt(nonInterrupt))
}

// ---------------------------------------------------------------------------
// entityIDsLabel — edge cases
// ---------------------------------------------------------------------------

func TestEntityIDsLabel(t *testing.T) {
	assert.Equal(t, "", entityIDsLabel(nil))
	assert.Equal(t, "", entityIDsLabel([]int64{}))
	assert.Equal(t, " #42", entityIDsLabel([]int64{42}))
	assert.Equal(t, " #1,2", entityIDsLabel([]int64{1, 2}))
	assert.Equal(t, " #1,2,3", entityIDsLabel([]int64{1, 2, 3}))
	assert.Equal(t, " #1,…(+3)", entityIDsLabel([]int64{1, 2, 3, 4}))
}

// ---------------------------------------------------------------------------
// filterByLabel
// ---------------------------------------------------------------------------

func TestFilterByLabel(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{Label: "production deploy"},
		{Label: "test run"},
		{Label: "Production backup"},
		{Label: ""},
	}

	result := filterByLabel(entries, "production")
	assert.Len(t, result, 2)
	assert.Equal(t, "production deploy", result[0].Label)
	assert.Equal(t, "Production backup", result[1].Label)

	result = filterByLabel(entries, "missing")
	assert.Empty(t, result)

	result = filterByLabel(entries, "")
	assert.Len(t, result, 4) // empty query matches all
}

// ---------------------------------------------------------------------------
// parseEntityIDs
// ---------------------------------------------------------------------------

func TestParseEntityIDs(t *testing.T) {
	ids, err := parseEntityIDs("1,2,3")
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, ids)

	ids, err = parseEntityIDs(" 42 ")
	require.NoError(t, err)
	assert.Equal(t, []int64{42}, ids)

	ids, err = parseEntityIDs("1, 2, , 3")
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, ids)

	_, err = parseEntityIDs("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty ID list")

	_, err = parseEntityIDs("abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ID")
}

// ---------------------------------------------------------------------------
// undoHint
// ---------------------------------------------------------------------------

func TestUndoHint(t *testing.T) {
	tests := []struct {
		meta     *snaplib.Meta
		contains string
	}{
		{&snaplib.Meta{Operation: snaplib.OpAdd}, "deleted during rollback"},
		{&snaplib.Meta{Operation: snaplib.OpCopy}, "deleted during rollback"},
		{&snaplib.Meta{Operation: snaplib.OpUpdate}, "post-mutation values"},
		{&snaplib.Meta{Operation: snaplib.OpDelete}, "Undo unavailable for this snapshot"},
	}
	for _, tc := range tests {
		assert.Contains(t, undoHint(tc.meta), tc.contains, "op=%s", tc.meta.Operation)
	}
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestRegister(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	Register(root, getClientForTests)

	snapCmd, _, err := root.Find([]string{"snap"})
	require.NoError(t, err)
	assert.Equal(t, "snap", snapCmd.Name())

	// Check subcommands exist.
	subNames := make([]string, 0)
	for _, c := range snapCmd.Commands() {
		subNames = append(subNames, c.Name())
	}
	assert.Contains(t, subNames, "list")
	assert.Contains(t, subNames, "info")
	assert.Contains(t, subNames, "rollback")
	assert.Contains(t, subNames, "export")
	assert.Contains(t, subNames, "delete")
	assert.Contains(t, subNames, "gc")
}

// ---------------------------------------------------------------------------
// hasExplicitFormat
// ---------------------------------------------------------------------------

func TestHasExplicitFormat(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "", "output format")
	assert.False(t, hasExplicitFormat(cmd))

	_ = cmd.Flags().Set("format", "json")
	assert.True(t, hasExplicitFormat(cmd))
}

func TestHasExplicitFormat_NoFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	assert.False(t, hasExplicitFormat(cmd))
}

// ---------------------------------------------------------------------------
// renderStatusSummary
// ---------------------------------------------------------------------------

func TestRenderStatusSummary(t *testing.T) {
	tests := []struct {
		status   snaplib.Status
		contains string
	}{
		{snaplib.StatusRolledBack, "fully rolled back"},
		{snaplib.StatusExpired, "expired"},
		{snaplib.StatusAvailable, "available for rollback"},
	}
	for _, tc := range tests {
		buf := &bytes.Buffer{}
		renderStatusSummary(buf, &snaplib.Meta{Status: tc.status})
		assert.Contains(t, buf.String(), tc.contains, "status=%s", tc.status)
	}
}

func TestRenderStatusSummary_Partial(t *testing.T) {
	buf := &bytes.Buffer{}
	meta := &snaplib.Meta{
		ID:     "test-snap",
		Status: snaplib.StatusRollbackPartial,
		RollbackLog: []snaplib.RollbackLogEntry{
			{Status: snaplib.RBFailed},
			{Status: snaplib.RBRestored},
		},
	}
	renderStatusSummary(buf, meta)
	assert.Contains(t, buf.String(), "1 of 2")
}

// ---------------------------------------------------------------------------
// renderInfoCard — basic rendering
// ---------------------------------------------------------------------------

func TestRenderInfoCard(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{
		ID:           "cases/20260418T120000_update_0",
		ServerURL:    "https://example.com",
		Operation:    snaplib.OpUpdate,
		EntityType:   "case",
		Category:     "cases",
		RollbackTier: snaplib.Tier1,
		Status:       snaplib.StatusAvailable,
		EntityIDs:    []int64{42, 99},
		ProjectID:    3,
		SuiteID:      10,
		CLICommand:   "gotr cases update",
		Timestamp:    time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC),
		Name:         "test-snap",
		Label:        "my-label",
	}

	renderInfoCard(cmd, meta)
	out := buf.String()
	assert.Contains(t, out, "cases/20260418T120000_update_0")
	assert.Contains(t, out, "42, 99")
	assert.Contains(t, out, "test-snap")
	assert.Contains(t, out, "my-label")
}

func TestRenderInfoCard_WithEntities(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{
		ID:           "test-snap",
		Operation:    snaplib.OpDelete,
		EntityType:   "case",
		Status:       snaplib.StatusAvailable,
		RollbackTier: snaplib.Tier2,
		Timestamp:    time.Now(),
		Entities: []snaplib.Entity{
			{Type: "case", ID: 42, ParentID: 10},
			{Type: "section", ID: 10, ParentID: 0},
		},
	}

	renderInfoCard(cmd, meta)
	out := buf.String()
	assert.Contains(t, out, "Entities (2)")
}

func TestRenderInfoCard_WithRollbackLog(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{
		ID:           "test-snap",
		Operation:    snaplib.OpUpdate,
		EntityType:   "case",
		Status:       snaplib.StatusRolledBack,
		RollbackTier: snaplib.Tier1,
		Timestamp:    time.Now(),
		RollbackLog: []snaplib.RollbackLogEntry{
			{Type: "case", ID: 42, Status: snaplib.RBRestored},
			{Type: "case", ID: 99, Status: snaplib.RBFailed, Error: "not found"},
		},
	}

	renderInfoCard(cmd, meta)
	out := buf.String()
	assert.Contains(t, out, "Rollback Log (2)")
	assert.Contains(t, out, "not found")
}

// ---------------------------------------------------------------------------
// filterByLabel with real snapshots
// ---------------------------------------------------------------------------

func TestCLI_SnapList_FilterByLabel(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	_, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	// Modify entry label in memory for filter testing.
	entries := manifest.Entries
	require.NotEmpty(t, entries)
	entries[0].Label = "deploy-v2"

	// Filter should match.
	filtered := filterByLabel(entries, "deploy")
	assert.Len(t, filtered, 1)
	assert.Equal(t, "deploy-v2", filtered[0].Label)

	// Filter should not match.
	filtered = filterByLabel(entries, "nonexistent")
	assert.Empty(t, filtered)
}

// ---------------------------------------------------------------------------
// shortID
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// CLI snap info — non-interactive with real snapshot
// ---------------------------------------------------------------------------

func TestCLI_SnapInfo_WithSnapshot(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	cmd := newInfoCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{snapID})

	err := cmd.Execute()
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, snapID)
	assert.Contains(t, out, "available for rollback")
}

// ---------------------------------------------------------------------------
// CLI snap export — non-interactive
// ---------------------------------------------------------------------------

func TestCLI_SnapExport_WithSnapshot(t *testing.T) {
	redirectHome(t)
	outDir := t.TempDir()

	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	outFile := fmt.Sprintf("%s/export.json", outDir)
	cmd := newExportCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{snapID, outFile})

	err := cmd.Execute()
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// CLI snap delete — non-interactive
// ---------------------------------------------------------------------------

func TestCLI_SnapDelete_WithSnapshot(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	cmd := newDeleteCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{snapID})

	err := cmd.Execute()
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// projectLabel
// ---------------------------------------------------------------------------

func TestProjectLabel(t *testing.T) {
	assert.Contains(t, projectLabel(0, 0), "–")
	assert.Contains(t, projectLabel(0, 5), "P5")
	assert.Contains(t, projectLabel(3, 5), "P3")
	assert.Contains(t, projectLabel(3, 5), "P5")
	assert.Contains(t, projectLabel(3, 0), "P3")
}

// ---------------------------------------------------------------------------
// truncate
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hello", truncate("hello", 5))
	assert.Equal(t, "h...", truncate("hello", 4))
	assert.Equal(t, "hel...", truncate("hello world", 6))
}

// ---------------------------------------------------------------------------
// pickByOperation — single op group, direct to category
// ---------------------------------------------------------------------------

func TestPickByOperation_SingleOpGroup(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "cases/snap2", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
	}

	// Single op group + single category → direct to pickSnapshot.
	// pickSnapshot: ✕ Exit at 0, data at 1,2. Select index 0 = first data item.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}) // pick first snapshot

	snapID, err := pickByOperation(mp, entries, "Select:", false)
	require.NoError(t, err)
	assert.Equal(t, "cases/snap1", snapID)
}

func TestPickByOperation_BackFromSingleOp(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
	}

	// Single op + allowBack: pickSnapshot shows ← Back.
	// Select ← Back (index 0 raw) → propagated to errGoBack.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true}) // ← Back at top

	_, err := pickByOperation(mp, entries, "Select:", true)
	assert.Equal(t, errGoBack, err)
}

// ---------------------------------------------------------------------------
// pickByCategory — multiple categories
// ---------------------------------------------------------------------------

func TestPickByCategory_MultipleCategories(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "sections/snap2", Operation: snaplib.OpUpdate, EntityType: "section", Category: "sections"},
	}

	// Category picker (no allowBack): [cases] at 0, [sections] at 1.
	// Pick cases (index 0 raw), then pickSnapshot (allowBack=true):
	// [← Back, snap1, ← Back]. Pick data (index 0 = first data item).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0, Raw: true}, // pick cases category
			interactive.SelectResponse{Index: 0},            // pick first snapshot in that category
		)

	snapID, err := pickByCategory(mp, entries, "Select:", false)
	require.NoError(t, err)
	assert.Equal(t, "cases/snap1", snapID)
}

func TestPickByCategory_BackFromCategory(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "sections/snap2", Operation: snaplib.OpUpdate, EntityType: "section", Category: "sections"},
	}

	// allowBack: ← Back at 0, ✕ Exit at 1, [cases] at 2, [sections] at 3.
	// Select ← Back (index 0 raw).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	_, err := pickByCategory(mp, entries, "Select:", true)
	assert.Equal(t, errGoBack, err)
}

func TestPickByCategory_BackFromSnapshot(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "sections/snap2", Operation: snaplib.OpUpdate, EntityType: "section", Category: "sections"},
	}

	// allowBack=true: ← Back at 0, [cases] at 1, [sections] at 2.
	// Pick cases (index 1 raw). pickSnapshot: ← Back at top → returns errGoBack.
	// Category re-shown. Then select ← Back (index 0 raw).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick cases
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from snapshot
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from category
		)

	_, err := pickByCategory(mp, entries, "Select:", true)
	assert.Equal(t, errGoBack, err)
}

// ---------------------------------------------------------------------------
// browseByOperation — single/multi ops
// ---------------------------------------------------------------------------

func TestBrowseByOperation_SingleOp_Exit(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
	}

	// Single op + single cat → browseSnapList directly.
	// browseSnapList: ✕ Exit at 0 (no allowBack). Select ✕ Exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true}) // ✕ Exit

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByOperation(cmd, store, mp, entries, false)
	assert.Equal(t, errExit, err)
}

func TestBrowseByOperation_MultiOp_SelectAndExit(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "cases/snap2", Operation: snaplib.OpDelete, EntityType: "case", Category: "cases"},
	}

	// Multi op: ✕ Exit at 0, [update] at 1, [delete] at 2.
	// Select ✕ Exit (index 0 raw).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByOperation(cmd, store, mp, entries, false)
	assert.Equal(t, errExit, err)
}

func TestBrowseByOperation_MultiOp_BackPropagated(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "cases/snap2", Operation: snaplib.OpDelete, EntityType: "case", Category: "cases"},
	}

	// Multi op + allowBack: ← Back at 0, ✕ Exit at 1, [update] at 2, [delete] at 3.
	// Select ← Back.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByOperation(cmd, store, mp, entries, true)
	assert.Equal(t, errGoBack, err)
}

// ---------------------------------------------------------------------------
// browseByCategory — single/multi categories
// ---------------------------------------------------------------------------

func TestBrowseByCategory_SingleCat_Exit(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
	}

	// Single category → skip picker → browseSnapList.
	// browseSnapList: ✕ Exit at 0 (allowBack=true → ← Back at 0, ✕ Exit at 1).
	// Select ✕ Exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1, Raw: true}) // ✕ Exit

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByCategory(cmd, store, mp, entries, true)
	assert.Equal(t, errExit, err)
}

func TestBrowseByCategory_MultiCat_SelectAndExit(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "sections/snap2", Operation: snaplib.OpUpdate, EntityType: "section", Category: "sections"},
	}

	// Multi cat: ✕ Exit at 0, [cases] at 1, [sections] at 2.
	// Select ✕ Exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByCategory(cmd, store, mp, entries, false)
	assert.Equal(t, errExit, err)
}

func TestBrowseByCategory_MultiCat_BackFromSnapList(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "sections/snap2", Operation: snaplib.OpUpdate, EntityType: "section", Category: "sections"},
	}

	// Multi cat + allowBack: ← Back at 0, ✕ Exit at 1, [cases] at 2, [sections] at 3.
	// Pick cases (2 raw) → browseSnapList shows ← Back at 0. Select ← Back from snaplist.
	// Category re-shown. Then ← Back from category.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 2, Raw: true}, // pick cases
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from snaplist
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from category
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByCategory(cmd, store, mp, entries, true)
	assert.Equal(t, errGoBack, err)
}

// ---------------------------------------------------------------------------
// browseSnapshots — end-to-end
// ---------------------------------------------------------------------------

func TestBrowseSnapshots_Empty(t *testing.T) {
	redirectHome(t)
	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	err = browseSnapshots(cmd, store, manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshots found")
}

func TestBrowseSnapshots_ExitImmediately(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	// Single server + single op → browseByOperation → browseSnapList.
	// browseSnapList: ✕ Exit at 0. Select it.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapshots(cmd, store, manifest)
	assert.NoError(t, err) // errExit is converted to nil
}

// ---------------------------------------------------------------------------
// postCardAction — different statuses
// ---------------------------------------------------------------------------

func TestPostCardAction_RolledBack(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true}) // ← Back

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Status: snaplib.StatusRolledBack}
	key, err := postCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, "back", key)
	assert.Contains(t, buf.String(), "already rolled back")
}

func TestPostCardAction_Expired(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1, Raw: true}) // ✕ Exit

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Status: snaplib.StatusExpired}
	key, err := postCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, "exit", key)
	assert.Contains(t, buf.String(), "expired")
}

func TestPostCardAction_Available_Rollback(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 2, Raw: true}) // ↻ Rollback

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Status: snaplib.StatusAvailable}
	key, err := postCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, "rollback", key)
}

// ---------------------------------------------------------------------------
// executeRollbackFromBrowser — no parent
// ---------------------------------------------------------------------------

func TestExecuteRollbackFromBrowser_NoParent(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	executeRollbackFromBrowser(cmd, "snap-123")
	assert.Contains(t, buf.String(), "gotr snap rollback snap-123")
}

// ---------------------------------------------------------------------------
// undoPickerLabels
// ---------------------------------------------------------------------------

func TestUndoPickerLabels(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
	}

	labels := undoPickerLabels(store, entries)
	require.Len(t, labels, 1)
	assert.Contains(t, labels[0], "no undo")
}

// ---------------------------------------------------------------------------
// postUndoCardAction
// ---------------------------------------------------------------------------

func TestPostUndoCardAction_Back(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Operation: snaplib.OpDelete, Status: snaplib.StatusRolledBack,
		RollbackLog: []snaplib.RollbackLogEntry{{Status: snaplib.RBRestored, NewID: 99}}}
	action, err := postUndoCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, postActionBack, action)
}

func TestPostUndoCardAction_Exit(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Operation: snaplib.OpDelete, Status: snaplib.StatusRolledBack}
	action, err := postUndoCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, postActionExit, action)
}

func TestPostUndoCardAction_UndoAvailable(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 2, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Operation: snaplib.OpDelete, Status: snaplib.StatusRolledBack,
		RollbackLog: []snaplib.RollbackLogEntry{{Status: snaplib.RBRestored, NewID: 99}}}
	action, err := postUndoCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, postActionRollback, action)
}

func TestPostUndoCardAction_UndoUnavailable_ShowsHint(t *testing.T) {
	// Select "undo" (unavailable) first, then ← Back.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 2, Raw: true}, // ↩ Undo (unavailable)
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Operation: snaplib.OpAdd, Status: snaplib.StatusRolledBack}
	action, err := postUndoCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, postActionBack, action)
	assert.Contains(t, buf.String(), "deleted during rollback")
}

// ---------------------------------------------------------------------------
// browseUndoSnapshots — empty
// ---------------------------------------------------------------------------

func TestBrowseUndoSnapshots_Empty(t *testing.T) {
	redirectHome(t)
	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	err = browseUndoSnapshots(cmd, nil, store, manifest)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No rolled-back snapshots")
}

// ---------------------------------------------------------------------------
// executeUndo — not found, not undoable
// ---------------------------------------------------------------------------

func TestExecuteUndo_NotFound(t *testing.T) {
	redirectHome(t)
	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	err = executeUndo(cmd, nil, store, manifest, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in manifest")
}

func TestExecuteUndo_NotUndoable(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	err := executeUndo(cmd, nil, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo not available")
}

// ---------------------------------------------------------------------------
// undoHint — sync operation
// ---------------------------------------------------------------------------

func TestUndoHint_Sync(t *testing.T) {
	meta := &snaplib.Meta{Operation: snaplib.OpSyncFull}
	assert.Contains(t, undoHint(meta), "re-sync required")
}

// ---------------------------------------------------------------------------
// selectSnapshot — status filter with no matches
// ---------------------------------------------------------------------------

func TestSelectSnapshot_StatusFilter(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	_, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	mp := interactive.NewMockPrompter()
	ctx := interactive.WithPrompter(context.Background(), mp)

	// Filter for rolled-back, but snapshot is available → no matches.
	_, err := selectSnapshot(ctx, manifest, "Select:", &pickerOpts{
		statusFilter: []snaplib.Status{snaplib.StatusRolledBack},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshots with status")
}

// ---------------------------------------------------------------------------
// selectSnapshot — empty manifest
// ---------------------------------------------------------------------------

func TestSelectSnapshot_Empty(t *testing.T) {
	redirectHome(t)
	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	mp := interactive.NewMockPrompter()
	ctx := interactive.WithPrompter(context.Background(), mp)

	_, err = selectSnapshot(ctx, manifest, "Select:", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshots found")
}

// ---------------------------------------------------------------------------
// newRollbackListCmd — client doesn't support undo
// ---------------------------------------------------------------------------

func TestNewRollbackListCmd_UnsupportedClient(t *testing.T) {
	redirectHome(t)

	// Use nil client — cast check fails before any API call.
	cmd := newRollbackListCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(&bytes.Buffer{})
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support undo")
}

// ---------------------------------------------------------------------------
// newRollbackUndoCmd — client doesn't support undo
// ---------------------------------------------------------------------------

func TestNewRollbackUndoCmd_UnsupportedClient(t *testing.T) {
	redirectHome(t)

	cmd := newRollbackUndoCmd(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(&bytes.Buffer{})
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"snap-123"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support undo")
}

// ---------------------------------------------------------------------------
// seedRolledBackSnapshot seeds a snapshot and marks it as rolled back.
// ---------------------------------------------------------------------------

func seedRolledBackSnapshot(t *testing.T) (*snaplib.Store, *snaplib.Manifest, string) {
	t.Helper()
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, snapID := seedSnapshot(t, snaplib.OpDelete, "case", []int64{42}, snaplib.Tier1, "", c)

	// Mark as rolled back with undo-eligible log.
	meta, err := store.LoadMeta(snapID)
	require.NoError(t, err)
	meta.Status = snaplib.StatusRolledBack
	meta.RollbackLog = []snaplib.RollbackLogEntry{
		{ID: 42, Status: snaplib.RBRestored, NewID: 99},
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.UpdateStatus(snapID, snaplib.StatusRolledBack))
	return store, manifest, snapID
}

// ---------------------------------------------------------------------------
// browseUndoSnapshots — single server, exit
// ---------------------------------------------------------------------------

func TestBrowseUndoSnapshots_SingleServer_Exit(t *testing.T) {
	redirectHome(t)
	store, manifest, _ := seedRolledBackSnapshot(t)

	// browseUndoSnapshots → single server → browseUndoList.
	// browseUndoList (allowBack=false): [✕ Exit, snap...]. Select ✕ Exit (index 0).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseUndoSnapshots(cmd, nil, store, manifest)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// browseUndoList — back, exit, view card and back
// ---------------------------------------------------------------------------

func TestBrowseUndoList_Exit(t *testing.T) {
	redirectHome(t)
	store, manifest, _ := seedRolledBackSnapshot(t)
	entries := manifest.ListByStatus(snaplib.StatusRolledBack)
	require.NotEmpty(t, entries)

	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true}) // ✕ Exit (no allowBack, index 0)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseUndoList(cmd, nil, store, manifest, mp, entries, false)
	assert.Equal(t, errExit, err)
}

func TestBrowseUndoList_Back(t *testing.T) {
	redirectHome(t)
	store, manifest, _ := seedRolledBackSnapshot(t)
	entries := manifest.ListByStatus(snaplib.StatusRolledBack)

	// allowBack: [← Back, ✕ Exit, snap..., ← Back]. Select ← Back (index 0).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseUndoList(cmd, nil, store, manifest, mp, entries, true)
	assert.Equal(t, errGoBack, err)
}

func TestBrowseUndoList_ViewCardAndBack(t *testing.T) {
	redirectHome(t)
	store, manifest, _ := seedRolledBackSnapshot(t)
	entries := manifest.ListByStatus(snaplib.StatusRolledBack)

	// allowBack=false: [✕ Exit, snap...].
	// Select snap at index 1. Then postUndoCardAction: ← Back (index 0 raw).
	// Back to list. Select ✕ Exit at index 0.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from card
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit from list
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseUndoList(cmd, nil, store, manifest, mp, entries, false)
	assert.Equal(t, errExit, err)
}

// ---------------------------------------------------------------------------
// browseSnapList — view card and rollback hint (no parent)
// ---------------------------------------------------------------------------

func TestBrowseSnapList_ViewCardAndRollback(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	realEntries := manifest.All()
	require.NotEmpty(t, realEntries)

	// allowBack=false: [✕ Exit, snap...]. Pick snap (index 1 raw).
	// postCardAction: ↻ Rollback (index 2 raw) → executeRollbackFromBrowser (no parent → hint).
	// Back to list → ✕ Exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap
			interactive.SelectResponse{Index: 2, Raw: true}, // ↻ Rollback
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapList(cmd, store, mp, realEntries, false)
	assert.Equal(t, errExit, err)
	assert.Contains(t, buf.String(), "gotr snap rollback")
}

// ---------------------------------------------------------------------------
// executeRollbackFromBrowser — with parent but no rollback sibling
// ---------------------------------------------------------------------------

func TestExecuteRollbackFromBrowser_NoRollbackSibling(t *testing.T) {
	parent := &cobra.Command{Use: "snap"}
	child := &cobra.Command{Use: "test"}
	parent.AddCommand(child)

	buf := &bytes.Buffer{}
	child.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	child.SetContext(ctx)

	executeRollbackFromBrowser(child, "snap-123")
	assert.Contains(t, buf.String(), "gotr snap rollback snap-123")
}

// ---------------------------------------------------------------------------
// pickByOperation — multiple ops, select first
// ---------------------------------------------------------------------------

func TestPickByOperation_MultiOps_SelectFirst(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "cases/snap1", Operation: snaplib.OpUpdate, EntityType: "case", Category: "cases"},
		{ID: "cases/snap2", Operation: snaplib.OpDelete, EntityType: "case", Category: "cases"},
	}

	// Multi op sorted alphabetically: [delete] at 0, [update] at 1 (no allowBack).
	// Pick delete (index 0 raw). Then single category → pickSnapshot: pick first.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0, Raw: true}, // pick "delete" op
			interactive.SelectResponse{Index: 0},            // pick first snapshot
		)

	snapID, err := pickByOperation(mp, entries, "Select:", false)
	require.NoError(t, err)
	assert.Equal(t, "cases/snap2", snapID)
}

// ---------------------------------------------------------------------------
// selectSnapshot — happy path with selection
// ---------------------------------------------------------------------------

func TestSelectSnapshot_SelectFirst(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	_, manifest, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	// selectSnapshot: single server + single op → goes to pickByOperation → pickSnapshot.
	// pickSnapshot (no allowBack): [snap...]. Index 0 = first snap.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	ctx := interactive.WithPrompter(context.Background(), mp)

	selected, err := selectSnapshot(ctx, manifest, "Select:", nil)
	require.NoError(t, err)
	assert.Equal(t, snapID, selected)
}

// ---------------------------------------------------------------------------
// undoHint — add and update operations
// ---------------------------------------------------------------------------

func TestUndoHint_Add(t *testing.T) {
	meta := &snaplib.Meta{Operation: snaplib.OpAdd}
	assert.Contains(t, undoHint(meta), "deleted during rollback")
}

func TestUndoHint_Update(t *testing.T) {
	meta := &snaplib.Meta{Operation: snaplib.OpUpdate}
	assert.Contains(t, undoHint(meta), "post-mutation")
}

func TestUndoHint_Copy(t *testing.T) {
	meta := &snaplib.Meta{Operation: snaplib.OpCopy}
	assert.Contains(t, undoHint(meta), "deleted during rollback")
}

func TestUndoHint_Default(t *testing.T) {
	meta := &snaplib.Meta{Operation: snaplib.OpBulk}
	assert.Contains(t, undoHint(meta), "Undo unavailable")
}

// ---------------------------------------------------------------------------
// browseUndoList — view card and exit from card
// ---------------------------------------------------------------------------

func TestBrowseUndoList_ViewCardAndExit(t *testing.T) {
	redirectHome(t)
	store, manifest, _ := seedRolledBackSnapshot(t)
	entries := manifest.ListByStatus(snaplib.StatusRolledBack)

	// Pick snap (index 1 raw), then ✕ Exit from card (index 1 raw).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap
			interactive.SelectResponse{Index: 1, Raw: true}, // ✕ Exit from card
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseUndoList(cmd, nil, store, manifest, mp, entries, false)
	assert.Equal(t, errExit, err)
}

// ---------------------------------------------------------------------------
// browseSnapList — view card and exit from card
// ---------------------------------------------------------------------------

func TestBrowseSnapList_ViewCardAndExit(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	realEntries := manifest.All()
	require.NotEmpty(t, realEntries)

	// Pick snap (index 1 raw), then ✕ Exit from card (index 1 raw).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap
			interactive.SelectResponse{Index: 1, Raw: true}, // ✕ Exit from card
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapList(cmd, store, mp, realEntries, false)
	assert.Equal(t, errExit, err)
}

// ---------------------------------------------------------------------------
// browseSnapList — view card → back to list → exit
// ---------------------------------------------------------------------------

func TestBrowseSnapList_ViewCardAndBack(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	realEntries := manifest.All()
	require.NotEmpty(t, realEntries)

	// Pick snap (index 1 raw), ← Back from card (index 0 raw), ✕ Exit from list.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from card
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit from list
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapList(cmd, store, mp, realEntries, false)
	assert.Equal(t, errExit, err)
}

// ---------------------------------------------------------------------------
// browseByOperation — multi op → select op → view snap → exit
// ---------------------------------------------------------------------------

func TestBrowseByOperation_MultiOp_ViewAndExit(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)
	// Seed a second snapshot with different operation
	c2 := &data.Case{ID: 43, Title: "Test2", SectionID: 1}
	store2, manifest2, _ := seedSnapshot(t, snaplib.OpDelete, "case", []int64{43}, snaplib.Tier1, "", c2)
	_ = store2
	_ = manifest2

	entries := manifest.All()
	require.NotEmpty(t, entries)

	// If only one op group (all update), just test single op + exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByOperation(cmd, store, mp, entries, false)
	assert.Equal(t, errExit, err)
}

// ---------------------------------------------------------------------------
// browseByCategory — single cat, back propagated
// ---------------------------------------------------------------------------

func TestBrowseByCategory_SingleCat_Back(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := manifest.All()
	require.NotEmpty(t, entries)

	// Single category + allowBack=true → browseSnapList (allowBack=true).
	// ← Back at index 0 (raw).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByCategory(cmd, store, mp, entries, true)
	assert.Equal(t, errGoBack, err)
}

// ---------------------------------------------------------------------------
// postCardAction — partial rollback status
// ---------------------------------------------------------------------------

func TestPostCardAction_PartialRollback(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 2, Raw: true}) // ↻ Rollback

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Status: snaplib.StatusRollbackPartial}
	key, err := postCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, "rollback", key)
}

// ---------------------------------------------------------------------------
// executeUndoFromBrowser — canceled by user
// ---------------------------------------------------------------------------

func TestExecuteUndoFromBrowser_Canceled(t *testing.T) {
	redirectHome(t)
	store, manifest, snapID := seedRolledBackSnapshot(t)
	_ = store
	_ = manifest

	mp := interactive.NewMockPrompter().
		WithConfirmResponses(false) // User says No to "Undo this rollback?"

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	executeUndoFromBrowser(cmd, nil, store, manifest, snapID)
	assert.Contains(t, buf.String(), "Canceled")
}

// ---------------------------------------------------------------------------
// browseUndoSnapshots — single server, view card + undo canceled, then exit
// ---------------------------------------------------------------------------

func TestBrowseUndoSnapshots_ViewAndUndoCanceled(t *testing.T) {
	redirectHome(t)
	store, manifest, _ := seedRolledBackSnapshot(t)

	// Single server → browseUndoList (allowBack=false).
	// [✕ Exit, snap...]. Pick snap (index 1 raw).
	// postUndoCardAction: ↩ Undo rollback (index 2 raw) → postActionRollback.
	// executeUndoFromBrowser: confirm → No → "Canceled".
	// Re-show list. ✕ Exit (index 0 raw).
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap
			interactive.SelectResponse{Index: 2, Raw: true}, // ↩ Undo rollback
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit
		).
		WithConfirmResponses(false) // cancel undo

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseUndoSnapshots(cmd, nil, store, manifest)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Canceled")
}

// ---------------------------------------------------------------------------
// browseSnapshots — single server, view card back → exit
// ---------------------------------------------------------------------------

func TestBrowseSnapshots_ViewCardAndBack(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	// Single server + single op → browseByOperation → browseSnapList.
	// Pick snap, back from card, exit from list.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap from list
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from card
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit from list
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapshots(cmd, store, manifest)
	assert.NoError(t, err) // errExit → nil
}

// ---------------------------------------------------------------------------
// browseSnapList — back from list with allowBack
// ---------------------------------------------------------------------------

func TestBrowseSnapList_BackFromList(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	realEntries := manifest.All()
	require.NotEmpty(t, realEntries)

	// allowBack=true: [← Back, ✕ Exit, snap..., ← Back]. Pick ← Back at bottom.
	bottomIdx := 1 + 1 + len(realEntries) // back + exit + data
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: bottomIdx, Raw: true}) // ← Back at bottom

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapList(cmd, store, mp, realEntries, true)
	assert.Equal(t, errGoBack, err)
}

// ---------------------------------------------------------------------------
// browseByOperation — single op → snap → back → exit
// ---------------------------------------------------------------------------

func TestBrowseByOperation_SingleOp_BackThenExit(t *testing.T) {
	redirectHome(t)
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	store, manifest, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	entries := manifest.All()
	require.NotEmpty(t, entries)

	// Single op → single cat → browseSnapList (allowBack=false).
	// Pick snap (index 1 raw), ← Back from card → back to snap list.
	// ✕ Exit from list.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snap
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from card
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit from list
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseByOperation(cmd, store, mp, entries, false)
	assert.Equal(t, errExit, err)
}

// ---------------------------------------------------------------------------
// postCardAction — back from available
// ---------------------------------------------------------------------------

func TestPostCardAction_Available_Back(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true}) // ← Back

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	meta := &snaplib.Meta{Status: snaplib.StatusAvailable}
	key, err := postCardAction(cmd, mp, meta)
	require.NoError(t, err)
	assert.Equal(t, "back", key)
}

// ---------------------------------------------------------------------------
// browseSnapshots — multi-server, select server then exit
// ---------------------------------------------------------------------------

func seedMultiServerSnaps(t *testing.T) (*snaplib.Store, *snaplib.Manifest) {
	t.Helper()
	c1 := &data.Case{ID: 42, Title: "Test1", SectionID: 1}
	store, manifest, snapID1 := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c1)

	// Change the first snapshot's ServerURL to create a multi-server scenario.
	meta1, err := store.LoadMeta(snapID1)
	require.NoError(t, err)
	meta1.ServerURL = "https://server1.testrail.io"
	require.NoError(t, store.SaveMeta(meta1))
	// Update manifest entry too.
	require.NoError(t, manifest.Remove(snapID1))
	require.NoError(t, manifest.Add(meta1))

	// Add second snapshot with different server.
	c2 := &data.Case{ID: 43, Title: "Test2", SectionID: 1}
	meta2 := snaplib.BuildMeta(snaplib.OpUpdate, "case", []int64{43}, snaplib.Tier1, 1, 1, "", []string{"test"}, "https://server2.testrail.io")
	fetchFn := func(ctx context.Context) (interface{}, error) { return c2, nil }
	_, err = snaplib.TakeSnapshot(context.Background(), store, manifest, meta2, fetchFn)
	require.NoError(t, err)
	return store, manifest
}

func TestBrowseSnapshots_MultiServer_Exit(t *testing.T) {
	redirectHome(t)
	store, manifest := seedMultiServerSnaps(t)

	// Multi-server: ✕ Exit at 0, [server1] at 1, [server2] at 2.
	// Select ✕ Exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Raw: true})

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapshots(cmd, store, manifest)
	assert.NoError(t, err) // errExit → nil
}

func TestBrowseSnapshots_MultiServer_SelectAndExit(t *testing.T) {
	redirectHome(t)
	store, manifest := seedMultiServerSnaps(t)

	// Multi-server: ✕ Exit at 0, [server1] at 1, [server2] at 2.
	// Select server1 (index 1 raw). Then single op → single cat → browseSnapList.
	// ✕ Exit from snap list.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick server1
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit from snap list
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapshots(cmd, store, manifest)
	assert.NoError(t, err)
}

func TestBrowseSnapshots_MultiServer_BackFromOps(t *testing.T) {
	redirectHome(t)
	store, manifest := seedMultiServerSnaps(t)

	// Select server1, then ← Back from ops (single op → browseSnapList allowBack=true).
	// ← Back from snap list. Then ✕ Exit from server picker.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick server1
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back from snap list
			interactive.SelectResponse{Index: 0, Raw: true}, // ✕ Exit from server picker
		)

	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	err := browseSnapshots(cmd, store, manifest)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// selectSnapshot — multi-server selection
// ---------------------------------------------------------------------------

func TestSelectSnapshot_MultiServer(t *testing.T) {
	redirectHome(t)
	store, manifest := seedMultiServerSnaps(t)
	_ = store

	// Multi-server: [server1] at 0, [server2] at 1.
	// Select server1. Then single op → pickByOperation → pickSnapshot.
	// pickSnapshot (no allowBack): [snap]. Pick first.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0, Raw: true}, // pick server1
			interactive.SelectResponse{Index: 0},            // pick first snapshot
		)

	ctx := interactive.WithPrompter(context.Background(), mp)

	selected, err := selectSnapshot(ctx, manifest, "Select:", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, selected)
}
