package snap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

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
		ID:          "cases/20260418T120000_update_0",
		ServerURL:   "https://example.com",
		Operation:   snaplib.OpUpdate,
		EntityType:  "case",
		Category:    "cases",
		RollbackTier: snaplib.Tier1,
		Status:      snaplib.StatusAvailable,
		EntityIDs:   []int64{42, 99},
		ProjectID:   3,
		SuiteID:     10,
		CLICommand:  "gotr cases update",
		Timestamp:   time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC),
		Name:        "test-snap",
		Label:       "my-label",
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
		ID:          "test-snap",
		Operation:   snaplib.OpDelete,
		EntityType:  "case",
		Status:      snaplib.StatusAvailable,
		RollbackTier: snaplib.Tier2,
		Timestamp:   time.Now(),
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
		ID:          "test-snap",
		Operation:   snaplib.OpUpdate,
		EntityType:  "case",
		Status:      snaplib.StatusRolledBack,
		RollbackTier: snaplib.Tier1,
		Timestamp:   time.Now(),
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
