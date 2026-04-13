package snap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
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

	meta := snaplib.BuildMeta(op, entityType, entityIDs, tier, 1, 1, customName, []string{"test"})
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

// execCmd builds a cobra command, sets args, captures stdout, executes.
func execCmd(t *testing.T, cmd *cobra.Command, args []string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// ---------------------------------------------------------------------------
// CLI Smoke: snap list
// ---------------------------------------------------------------------------

func TestCLI_SnapList_Empty(t *testing.T) {
	redirectHome(t)

	cmd := newListCmd()
	// Override stdout to capture.
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	// Create store dir so it doesn't error.
	_, err := snaplib.NewStore()
	require.NoError(t, err)

	// Redirect os.Stdout for this test (list writes to os.Stdout directly).
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	captured.ReadFrom(r)
	os.Stdout = old

	assert.Contains(t, captured.String(), "No snapshots found")
}

func TestCLI_SnapList_WithEntries(t *testing.T) {
	redirectHome(t)

	// Seed two snapshots.
	c := &data.Case{ID: 42, Title: "Test", SectionID: 1}
	seedSnapshot(t, snaplib.OpUpdate, "case", []int64{42}, snaplib.Tier1, "", c)
	seedSnapshot(t, snaplib.OpDelete, "case", []int64{99}, snaplib.Tier2, "", c)

	cmd := newListCmd()

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

	out := captured.String()
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

	// Output is JSON — parse it.
	var meta snaplib.Meta
	require.NoError(t, json.Unmarshal(captured.Bytes(), &meta))
	assert.Equal(t, snapID, meta.ID)
	assert.Equal(t, snaplib.OpUpdate, meta.Operation)
	assert.Equal(t, "case", meta.EntityType)
	assert.Equal(t, snaplib.StatusAvailable, meta.Status)
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
	ctx := context.WithValue(context.Background(), testClientKey, mock)
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
	ctx := context.WithValue(context.Background(), testClientKey, mock)
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

	meta := snaplib.BuildMeta(snaplib.OpAdd, "case", nil, snaplib.Tier2, 1, 1, "", []string{"cases", "add"})
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
	ctx := context.WithValue(context.Background(), testClientKey, mock)
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
	ctx := context.WithValue(context.Background(), testClientKey, mock)
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
	ctx := context.WithValue(context.Background(), testClientKey, mock)
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
	ctx := context.WithValue(context.Background(), testClientKey, mock)
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
	ctx := context.WithValue(context.Background(), testClientKey, mock)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{snapID})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
}
