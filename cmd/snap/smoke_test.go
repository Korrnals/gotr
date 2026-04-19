package snap

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
// Test infrastructure
// ---------------------------------------------------------------------------

type testCtxKey string

const testClientKey testCtxKey = "httpClient"

func getClientForTests(cmd *cobra.Command) client.ClientInterface {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if c, ok := cmd.Context().Value(testClientKey).(client.ClientInterface); ok {
		return c
	}
	return nil
}

// redirectHome sets HOME to a temp dir so snap store goes to $HOME/.gotr/snaps/.
// Returns the home path.
func redirectHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

// seedSnapshot creates a snapshot in the redirected store via the library.
// Returns store, manifest, snapID.
func seedSnapshot(t *testing.T, op snaplib.Operation, entityType string, entityIDs []int64, tier snaplib.Tier, customName string, fetchData interface{}) (*snaplib.Store, *snaplib.Manifest, string) {
	t.Helper()

	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	meta := snaplib.BuildMeta(op, entityType, entityIDs, tier, 1, 1, customName, []string{"test"}, "")
	var fetchFn func(context.Context) (interface{}, error)
	if fetchData != nil {
		fetchFn = func(ctx context.Context) (interface{}, error) {
			return fetchData, nil
		}
	}

	snap, err := snaplib.TakeSnapshot(context.Background(), store, manifest, meta, fetchFn)
	require.NoError(t, err)
	return store, manifest, snap.Meta.ID
}

// ---------------------------------------------------------------------------
// CLI Smoke: snap list
// ---------------------------------------------------------------------------

func TestCLI_SnapList_Empty(t *testing.T) {
	redirectHome(t)

	_, err := snaplib.NewStore()
	require.NoError(t, err)

	cmd := newListCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "No snapshots found")
}

func TestCLI_SnapList_WithEntries(t *testing.T) {
	redirectHome(t)

	// Seed two snapshots.
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)
	seedSnapshot(t, snaplib.OpDelete, "case", []int64{99}, snaplib.Tier2, "", c)

	cmd := newListCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "update")
	assert.Contains(t, out, "delete")
	assert.Contains(t, out, "case")
	assert.Contains(t, out, "cases")
}

// ---------------------------------------------------------------------------
// CLI Smoke: snap info
// ---------------------------------------------------------------------------

func TestCLI_SnapInfo(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 42, Title: "Info test", SectionID: 1, PriorityID: 3}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	cmd := newInfoCmd()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{snapID})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	// Output is now a table card. Verify key fields are present.
	out := captured.String()
	assert.Contains(t, out, snapID)
	assert.Contains(t, out, "update")
	assert.Contains(t, out, "case")
	assert.Contains(t, out, "available")
}

func TestCLI_SnapInfo_NotFound(t *testing.T) {
	redirectHome(t)
	_, _ = snaplib.NewStore() // ensure dir

	cmd := newInfoCmd()
	cmd.SetArgs([]string{"nonexistent/snap"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// CLI Smoke: snap rollback
// ---------------------------------------------------------------------------

// nonInteractiveCtx returns a context with non-interactive prompter and mock client.
func nonInteractiveCtx(mock client.ClientInterface) context.Context {
	ctx := context.WithValue(context.Background(), testClientKey, mock)
	ctx = interactive.WithPrompter(ctx, &interactive.NonInteractivePrompter{})
	return ctx
}

func TestCLI_SnapRollback_Update(t *testing.T) {
	redirectHome(t)

	original := &data.Case{ID: 42, Title: "Original", SectionID: 1, PriorityID: 3, TypeID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", original)

	var capturedReq *data.UpdateCaseRequest
	mock := &client.MockClient{
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(42), caseID)
			capturedReq = req
			return &data.Case{ID: 42, Title: "Original"}, nil
		},
	}

	cmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd.SetContext(ctx)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{snapID})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	// Verify API was called with original values.
	require.NotNil(t, capturedReq)
	assert.Equal(t, "Original", *capturedReq.Title)
	assert.Equal(t, int64(3), *capturedReq.PriorityID)

	// Verify status changed in manifest.
	store, _ := snaplib.NewStore()
	manifest, _ := snaplib.LoadManifest(store)
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, snaplib.StatusRolledBack, entry.Status)
}

func TestCLI_SnapRollback_Delete(t *testing.T) {
	redirectHome(t)

	original := &data.Case{ID: 99, Title: "Deleted case", SectionID: 20, PriorityID: 2, TypeID: 3}
	_, _, snapID := seedSnapshot(t, snaplib.OpDelete, "case", []int64{99}, snaplib.Tier2, "", original)

	mock := &client.MockClient{
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(20), sectionID)
			assert.Equal(t, "Deleted case", req.Title)
			return &data.Case{ID: 1001, Title: req.Title, SectionID: sectionID}, nil
		},
	}

	cmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd.SetContext(ctx)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{snapID})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "re-created")
}

func TestCLI_SnapRollback_Add(t *testing.T) {
	redirectHome(t)

	// Seed add snapshot + finalize with createdID=500.
	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	meta := snaplib.BuildMeta(snaplib.OpAdd, "case", nil, snaplib.Tier2, 1, 1, "", []string{"cases", "add"}, "")
	snap, err := snaplib.TakeSnapshot(context.Background(), store, manifest, meta, nil)
	require.NoError(t, err)
	require.NoError(t, snap.FinalizeAdd(500))
	snapID := snap.Meta.ID

	var deletedID int64
	mock := &client.MockClient{
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			deletedID = caseID
			return nil
		},
	}

	cmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd.SetContext(ctx)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{snapID})
	err = cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Equal(t, int64(500), deletedID)
	assert.Contains(t, captured.String(), "deleted")
}

func TestCLI_SnapRollback_NotFound(t *testing.T) {
	redirectHome(t)
	_, _ = snaplib.NewStore()

	mock := &client.MockClient{}
	cmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"nonexistent/snap"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCLI_SnapRollback_AlreadyRolledBack(t *testing.T) {
	redirectHome(t)

	original := &data.Case{ID: 42, Title: "Done", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", original)

	mock := &client.MockClient{
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return &data.Case{ID: caseID}, nil
		},
	}

	// First rollback.
	cmd1 := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd1.SetContext(ctx)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	cmd1.SetArgs([]string{snapID})
	err := cmd1.Execute()
	w.Close()
	os.Stdout = old
	require.NoError(t, err)

	// Second rollback → error.
	cmd2 := newRollbackCmd(getClientForTests)
	cmd2.SetContext(ctx)
	cmd2.SetArgs([]string{snapID})
	err = cmd2.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rolled_back")
}

// ---------------------------------------------------------------------------
// CLI Smoke: snap delete
// ---------------------------------------------------------------------------

func TestCLI_SnapDelete(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 10, Title: "ToDelete", SectionID: 1}
	store, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{10}, snaplib.Tier1, "", c)

	assert.True(t, store.Exists(snapID))

	cmd := newDeleteCmd()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{snapID})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "deleted")

	// Verify files removed.
	assert.False(t, store.Exists(snapID))

	// Manifest entry removed.
	manifest, _ := snaplib.LoadManifest(store)
	assert.Nil(t, manifest.Find(snapID))
}

// ---------------------------------------------------------------------------
// CLI Smoke: snap gc
// ---------------------------------------------------------------------------

func TestCLI_SnapGC_NoOrphans(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 1, Title: "Tracked", SectionID: 1}
	seedSnapshot(t, snaplib.OpUpdate, "case", []int64{1}, snaplib.Tier1, "", c)

	cmd := newGCCmd()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "No orphaned snapshots")
}

func TestCLI_SnapGC_CleansOrphans(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 1, Title: "Tracked", SectionID: 1}
	store, _, _ := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{1}, snaplib.Tier1, "", c)

	// Create an orphan on disk (no manifest entry).
	orphanMeta := &snaplib.Meta{
		ID:        "cases/orphan_cli_test",
		Category:  "cases",
		Operation: snaplib.OpUpdate,
		Status:    snaplib.StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(orphanMeta))
	assert.True(t, store.Exists("cases/orphan_cli_test"))

	cmd := newGCCmd()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "Cleaned 1")
	assert.False(t, store.Exists("cases/orphan_cli_test"))
}

// ---------------------------------------------------------------------------
// CLI Smoke: Full cycle — list → rollback → list (status changed)
// ---------------------------------------------------------------------------

func TestCLI_FullCycle_ListRollbackList(t *testing.T) {
	redirectHome(t)

	original := &data.Case{ID: 42, Title: "Cycle test", SectionID: 5, PriorityID: 2}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", original)

	// 1. List — should show "available".
	store, _ := snaplib.NewStore()
	manifest, _ := snaplib.LoadManifest(store)
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, snaplib.StatusAvailable, entry.Status)

	// 2. Rollback.
	mock := &client.MockClient{
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return &data.Case{ID: caseID}, nil
		},
	}
	rbCmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	rbCmd.SetContext(ctx)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	rbCmd.SetArgs([]string{snapID})
	err := rbCmd.Execute()
	w.Close()
	os.Stdout = old
	require.NoError(t, err)

	// 3. Reload manifest — status must be "rolled_back".
	store2, _ := snaplib.NewStore()
	manifest2, _ := snaplib.LoadManifest(store2)
	entry2 := manifest2.Find(snapID)
	require.NotNil(t, entry2)
	assert.Equal(t, snaplib.StatusRolledBack, entry2.Status)
}

// ---------------------------------------------------------------------------
// CLI Smoke: rollback with API error produces user-friendly message
// ---------------------------------------------------------------------------

func TestCLI_SnapRollback_APIError(t *testing.T) {
	redirectHome(t)

	original := &data.Case{ID: 55, Title: "Error test", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{55}, snaplib.Tier1, "", original)

	mock := &client.MockClient{
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return nil, fmt.Errorf("API returned 503: service unavailable")
		},
	}

	cmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{snapID})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
}

// ---------------------------------------------------------------------------
// CLI Interactive: snapshot selection picker
// ---------------------------------------------------------------------------

// interactiveCtx returns a context with MockPrompter and mock client.
func interactiveCtx(mock client.ClientInterface, prompter *interactive.MockPrompter) context.Context {
	ctx := context.WithValue(context.Background(), testClientKey, mock)
	ctx = interactive.WithPrompter(ctx, prompter)
	return ctx
}

func TestCLI_SnapInfo_Interactive(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 42, Title: "Interactive info", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	// browseSnapshots: single server + single op → direct to snapshot picker.
	// Options: ✕ Exit(0), [1] snapshot(1). Select index 1 to view card.
	// On next loop iteration mock exhausts → graceful exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1, Raw: true})

	cmd := newInfoCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, snapID)
	assert.Contains(t, out, "update")
	assert.Contains(t, out, "case")
}

func TestCLI_SnapDelete_Interactive(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 99, Title: "Delete interactive", SectionID: 1}
	seedSnapshot(t, snaplib.OpDelete, "case", []int64{99}, snaplib.Tier2, "", c)

	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newDeleteCmd()
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "deleted")
}

func TestCLI_SnapRollback_Interactive(t *testing.T) {
	redirectHome(t)

	original := &data.Case{ID: 42, Title: "Original", SectionID: 1, PriorityID: 2}
	_, _, _ = seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", original)

	mock := &client.MockClient{
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			return &data.Case{ID: 42, Title: "Changed", SectionID: 1, PriorityID: 3}, nil
		},
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return &data.Case{ID: 42, Title: "Original", SectionID: 1, PriorityID: 2}, nil
		},
	}

	// MockPrompter: select snapshot (index 0 — header in prompt), then confirm rollback.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithConfirmResponses(true)

	cmd := newRollbackCmd(getClientForTests)
	ctx := interactiveCtx(mock, mp)
	cmd.SetContext(ctx)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// No args — interactive selection.
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "Rollback complete")
}

func TestCLI_SnapInfo_NonInteractive_NoArgs(t *testing.T) {
	redirectHome(t)
	_, _ = snaplib.NewStore()

	cmd := newInfoCmd()
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive")
}

func TestCLI_SnapExport_Interactive(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 77, Title: "Export me", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{77}, snaplib.Tier1, "", c)

	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithInputResponses("", "")

	cmd := newExportCmd()
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	// Use a temp working directory so the default output file lands there.
	workDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(workDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// No args — interactive picker, default output filename.
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "Exported snapshot")

	// Verify default file was created.
	expectedFile := "snapshot_" + sanitizeFilename(snapID) + ".json"
	assert.FileExists(t, expectedFile)
}

// ---------------------------------------------------------------------------
// Phase 7: groupByServer unit tests
// ---------------------------------------------------------------------------

func TestGroupByServer_MultipleServers(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "a", ServerURL: "https://b.testrail.io", Operation: snaplib.OpUpdate},
		{ID: "b", ServerURL: "https://a.testrail.io", Operation: snaplib.OpDelete},
		{ID: "c", ServerURL: "https://b.testrail.io", Operation: snaplib.OpAdd},
		{ID: "d", ServerURL: "", Operation: snaplib.OpUpdate},
	}

	groups := groupByServer(entries)
	require.Len(t, groups, 3)

	// Sorted by URL: "" < "https://a..." < "https://b..."
	assert.Equal(t, "", groups[0].URL)
	assert.Len(t, groups[0].Entries, 1)

	assert.Equal(t, "https://a.testrail.io", groups[1].URL)
	assert.Len(t, groups[1].Entries, 1)

	assert.Equal(t, "https://b.testrail.io", groups[2].URL)
	assert.Len(t, groups[2].Entries, 2)
}

func TestGroupByServer_SingleServer(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "a", ServerURL: "https://x.testrail.io"},
		{ID: "b", ServerURL: "https://x.testrail.io"},
	}

	groups := groupByServer(entries)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0].Entries, 2)
}

// ---------------------------------------------------------------------------
// Phase 7: formatEntryLabel unit tests
// ---------------------------------------------------------------------------

func TestFormatPickerLabel_WithName(t *testing.T) {
	e := snaplib.ManifestEntry{
		Operation:    snaplib.OpUpdate,
		EntityType:   "case",
		Name:         "my-snap",
		Status:       snaplib.StatusAvailable,
		RollbackTier: snaplib.Tier1,
		EntityIDs:    []int64{42},
		Timestamp:    time.Date(2026, 4, 14, 7, 38, 0, 0, time.UTC),
	}
	label := formatPickerLabel(1, e)
	assert.Contains(t, label, "[1]")
	assert.Contains(t, label, "update case")
	assert.Contains(t, label, "#42")
	assert.Contains(t, label, `"my-snap"`)
	assert.Contains(t, label, "T1")
	assert.Contains(t, label, "2026-04-14 07:38")
}

func TestFormatPickerLabel_WithoutName(t *testing.T) {
	e := snaplib.ManifestEntry{
		Operation:    snaplib.OpDelete,
		EntityType:   "section",
		Status:       snaplib.StatusRolledBack,
		RollbackTier: snaplib.Tier2,
		Timestamp:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	label := formatPickerLabel(3, e)
	assert.Contains(t, label, "[3]")
	assert.Contains(t, label, "delete section")
	assert.NotContains(t, label, `"`)
	assert.Contains(t, label, "T2")
}

// ---------------------------------------------------------------------------
// Phase 7: listTable table output
// ---------------------------------------------------------------------------

func TestCLI_SnapList_TableOutput(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)

	cmd := newListCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	// Table should include header columns.
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "SERVER")
	assert.Contains(t, out, "OP")
	assert.Contains(t, out, "STATUS")
	assert.Contains(t, out, "update")
}

// ---------------------------------------------------------------------------
// Phase 7: listInteractive two-level picker
// ---------------------------------------------------------------------------

func TestCLI_SnapList_Interactive_MultiServer(t *testing.T) {
	redirectHome(t)

	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	// Seed two snapshots with different server URLs.
	meta1 := snaplib.BuildMeta(snaplib.OpUpdate, "case", []int64{1}, snaplib.Tier1, 1, 1, "", []string{"test"}, "https://server-a.testrail.io")
	_, err = snaplib.TakeSnapshot(context.Background(), store, manifest, meta1, func(ctx context.Context) (interface{}, error) {
		return &data.Case{ID: 1, Title: "Case A", SectionID: 1}, nil
	})
	require.NoError(t, err)

	meta2 := snaplib.BuildMeta(snaplib.OpDelete, "case", []int64{2}, snaplib.Tier2, 1, 1, "", []string{"test"}, "https://server-b.testrail.io")
	_, err = snaplib.TakeSnapshot(context.Background(), store, manifest, meta2, func(ctx context.Context) (interface{}, error) {
		return &data.Case{ID: 2, Title: "Case B", SectionID: 1}, nil
	})
	require.NoError(t, err)

	// Mock: browseSnapshots — select server (index 1: skip ✕ Exit at 0),
	// then snapshot (index 2: skip ← Back at 0, ✕ Exit at 1).
	// After viewing card, mock exhausts → graceful exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true},
			interactive.SelectResponse{Index: 2, Raw: true},
		)

	cmd := newListCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	// list shows info card via browseSnapshots.
	assert.Contains(t, out, "Snapshot Info")
	assert.Contains(t, out, "update")
}

// ---------------------------------------------------------------------------
// Phase 7: renderInfoCard output verification
// ---------------------------------------------------------------------------

func TestCLI_SnapInfo_CardContainsFields(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 42, Title: "Card test", SectionID: 1, PriorityID: 3}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "my-named-snap", c)

	cmd := newInfoCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), &interactive.NonInteractivePrompter{})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{snapID})
	err := cmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Snapshot Info")
	assert.Contains(t, out, snapID)
	assert.Contains(t, out, "update case")
	assert.Contains(t, out, "T1 (full rollback)")
	assert.Contains(t, out, "available")
	assert.Contains(t, out, "my-named-snap")
	assert.Contains(t, out, "42")
}

// ---------------------------------------------------------------------------
// Phase 7: printRollbackHeader output
// ---------------------------------------------------------------------------

func TestPrintRollbackHeader(t *testing.T) {
	entry := &snaplib.ManifestEntry{
		ID:           "cases/test_snap",
		ServerURL:    "https://my.testrail.io",
		Operation:    snaplib.OpUpdate,
		EntityType:   "case",
		RollbackTier: snaplib.Tier1,
	}

	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	printRollbackHeader(cmd, entry)
	out := buf.String()

	assert.Contains(t, out, "https://my.testrail.io")
	assert.Contains(t, out, "cases/test_snap")
	assert.Contains(t, out, "update case")
	assert.Contains(t, out, "T1")
}

func TestPrintRollbackHeader_UnknownServer(t *testing.T) {
	entry := &snaplib.ManifestEntry{
		ID:        "cases/test_snap",
		ServerURL: "",
	}

	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	printRollbackHeader(cmd, entry)
	assert.Contains(t, buf.String(), "(legacy)")
}

// ---------------------------------------------------------------------------
// Phase 7: selectSnapshot server-aware picker options
// ---------------------------------------------------------------------------

func TestSelectSnapshot_ShowsSnapshotInOptions(t *testing.T) {
	redirectHome(t)

	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	meta := snaplib.BuildMeta(snaplib.OpUpdate, "case", []int64{1}, snaplib.Tier1, 1, 1, "", []string{"test"}, "https://demo.testrail.io")
	_, err = snaplib.TakeSnapshot(context.Background(), store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return &data.Case{ID: 1, Title: "Test", SectionID: 1}, nil
	})
	require.NoError(t, err)

	// Mock: select index 0 (no back for single server/op, header is in prompt text).
	mp := &captureSelectPrompter{index: 0}

	ctx := interactive.WithPrompter(context.Background(), mp)

	snapID, err := selectSnapshot(ctx, manifest, "Pick:", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, snapID)

	// Single server, single op → 1 data option.
	require.Len(t, mp.options, 1)
	assert.Contains(t, mp.options[0], "update case")
	assert.Contains(t, mp.options[0], "#1")
	// Prompt contains snapshot count.
	assert.Contains(t, mp.message, "1 snapshots")
}

// captureSelectPrompter records options passed to Select.
type captureSelectPrompter struct {
	index   int
	message string
	options []string
}

func (c *captureSelectPrompter) Input(message, defaultVal string) (string, error) {
	return defaultVal, nil
}
func (c *captureSelectPrompter) Confirm(message string, def bool) (bool, error) {
	return def, nil
}
func (c *captureSelectPrompter) Select(message string, options []string) (int, string, error) {
	c.message = message
	c.options = options
	if c.index >= len(options) {
		return 0, "", fmt.Errorf("index out of range")
	}
	return c.index, options[c.index], nil
}
func (c *captureSelectPrompter) MultilineInput(message, defVal string) (string, error) {
	return defVal, nil
}

// ---------------------------------------------------------------------------
// Phase 7: export interactive prompt with custom values
// ---------------------------------------------------------------------------

func TestCLI_SnapExport_Interactive_CustomPath(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 88, Title: "Custom export", SectionID: 1}
	seedSnapshot(t, snaplib.OpUpdate, "case", []int64{88}, snaplib.Tier1, "", c)

	workDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(workDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}).
		WithInputResponses("custom_export.json", ".")

	cmd := newExportCmd()
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "Exported snapshot")
	assert.FileExists(t, "custom_export.json")
}

// ---------------------------------------------------------------------------
// Phase 7: tierLabel and humanSize helpers
// ---------------------------------------------------------------------------

func TestTierLabel(t *testing.T) {
	assert.Equal(t, "T1 (full rollback)", tierLabel(snaplib.Tier1))
	assert.Equal(t, "T2 (new ID on rollback)", tierLabel(snaplib.Tier2))
	assert.Equal(t, "T3 (info only)", tierLabel(snaplib.Tier3))
	assert.Equal(t, "T0", tierLabel(snaplib.Tier(0)))
}

func TestHumanSize(t *testing.T) {
	assert.Equal(t, "0 B", humanSize(0))
	assert.Equal(t, "512 B", humanSize(512))
	assert.Contains(t, humanSize(1024), "KB")
	assert.Contains(t, humanSize(1024*1024), "MB")
}

// ---------------------------------------------------------------------------
// Phase 7: sanitizeFilename
// ---------------------------------------------------------------------------

func TestSanitizeFilename(t *testing.T) {
	assert.Equal(t, "cases_snap_1", sanitizeFilename("cases/snap_1"))
	assert.Equal(t, "no_slashes", sanitizeFilename("no_slashes"))
	assert.Equal(t, "a_b_c", sanitizeFilename("a/b/c"))
}

// ---------------------------------------------------------------------------
// serverLabel URL normalization
// ---------------------------------------------------------------------------

func TestServerLabel_URLNormalization(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "(legacy)"},
		{"https://testrail.komus.net/index.php?/api/v2/", "https://testrail.komus.net"},
		{"https://demo.testrail.io/api/v2", "https://demo.testrail.io"},
		{"https://my.testrail.io", "https://my.testrail.io"},
		{"test", "test"}, // non-URL string preserved as-is
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, serverLabel(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// groupByOperation
// ---------------------------------------------------------------------------

func TestGroupByOperation(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{ID: "a", Operation: snaplib.OpUpdate},
		{ID: "b", Operation: snaplib.OpDelete},
		{ID: "c", Operation: snaplib.OpUpdate},
		{ID: "d", Operation: snaplib.OpAdd},
	}

	groups := groupByOperation(entries)
	require.Len(t, groups, 3)

	// Sorted: add < delete < update
	assert.Equal(t, "add", groups[0].Label)
	assert.Len(t, groups[0].Entries, 1)
	assert.Equal(t, "delete", groups[1].Label)
	assert.Len(t, groups[1].Entries, 1)
	assert.Equal(t, "update", groups[2].Label)
	assert.Len(t, groups[2].Entries, 2)
}

// ---------------------------------------------------------------------------
// alignedPickerLabels
// ---------------------------------------------------------------------------

func TestAlignedPickerLabels(t *testing.T) {
	entries := []snaplib.ManifestEntry{
		{
			ID: "cases/20260101T000000_update_42",
			Operation: snaplib.OpUpdate, EntityType: "case",
			EntityIDs: []int64{42}, Status: snaplib.StatusAvailable,
			RollbackTier: snaplib.Tier1, Category: "cases",
			Timestamp:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID: "sections/20260102T000000_delete_7",
			Operation: snaplib.OpDelete, EntityType: "section",
			Name: "my-snap", Status: snaplib.StatusRolledBack,
			RollbackTier: snaplib.Tier2, Category: "sections",
			Timestamp:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	header, labels := alignedPickerLabels(entries)
	require.Len(t, labels, 2)
	// Header has column titles.
	assert.Contains(t, header, "OPERATION")
	assert.Contains(t, header, "IDS")
	assert.Contains(t, header, "CATEGORY")
	assert.Contains(t, header, "STATUS")
	assert.Contains(t, header, "SNAPSHOT ID")
	// Data rows.
	assert.Contains(t, labels[0], "update case")
	assert.Contains(t, labels[0], "#42")
	assert.Contains(t, labels[0], "cases")
	assert.Contains(t, labels[0], "20260101T000000_update_42")
	assert.Contains(t, labels[1], "delete section")
	assert.Contains(t, labels[1], `"my-snap"`)
	assert.Contains(t, labels[1], "sections")
	assert.Contains(t, labels[1], "–") // no IDs → dash
	assert.Contains(t, labels[1], "20260102T000000_delete_7")
	// IDS separated with │ from OPERATION.
	assert.Contains(t, labels[0], "│ #42")
	assert.Contains(t, labels[1], "│ –")
}

func TestShortID(t *testing.T) {
	assert.Equal(t, "20260413T194322_add_bulk_0", shortID("cases/20260413T194322_add_bulk_0"))
	assert.Equal(t, "snap_1", shortID("sections/snap_1"))
	assert.Equal(t, "no_slash", shortID("no_slash"))
}

// ---------------------------------------------------------------------------
// Post-card action menu
// ---------------------------------------------------------------------------

func TestCLI_SnapInfo_Interactive_BackAfterCard(t *testing.T) {
	redirectHome(t)

	c := &data.Case{ID: 77, Title: "Back after card", SectionID: 1}
	_, _, snapID := seedSnapshot(t, snaplib.OpUpdate, "case", []int64{77}, snaplib.Tier1, "", c)

	// browseSnapshots: single server + single op → direct to snapshot picker.
	// Select index 1 (skip ✕ Exit at 0) to view card.
	// Post-card menu: select 0 = ← Back → loops to snapshot list.
	// Mock then exhausts on next snapshot picker → graceful exit.
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 1, Raw: true}, // pick snapshot
			interactive.SelectResponse{Index: 0, Raw: true}, // ← Back (post-card)
		)

	cmd := newInfoCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, snapID)
	assert.Contains(t, out, "Snapshot Info")
}
