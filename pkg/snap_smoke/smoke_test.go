//go:build smoke

package snap_smoke

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test suite setup
//
// Default mode: starts an in-memory FakeTestRail (no external deps).
// Real-server mode: set GOTR_SMOKE_URL + GOTR_SMOKE_USER + GOTR_SMOKE_KEY +
//
//	GOTR_SMOKE_PROJECT to hit a live TestRail instance instead.
//
// ---------------------------------------------------------------------------

var (
	smokeCfg     *SmokeConfig
	smokeCli     *client.HTTPClient
	smokeSectID  int64
	smokeSetupOK bool
	fakeServer   *FakeTestRail // non-nil when using the built-in mock
)

func TestMain(m *testing.M) {
	cfg, err := LoadConfig()
	if err != nil {
		// No env configured — spin up FakeTestRail.
		fake := NewFakeTestRail()
		fakeServer = fake

		cfg = &SmokeConfig{
			BaseURL:   fake.URL(),
			Username:  "smoke@test.local",
			APIKey:    "fakekey",
			ProjectID: 3,
			SuiteID:   1,
		}

		// Pre-seed a section so testSection() can find it.
		fake.SeedSection(100, "[smoke] snap-rollback-tests", cfg.SuiteID)

		fmt.Fprintln(os.Stderr, "smoke: using built-in FakeTestRail at", fake.URL())
	} else {
		fmt.Fprintln(os.Stderr, "smoke: using real server at", cfg.BaseURL)
	}

	smokeCfg = cfg

	cli, err := NewClient(cfg)
	if err != nil {
		os.Stderr.WriteString("smoke: client creation failed — " + err.Error() + "\n")
		os.Exit(1)
	}
	smokeCli = cli
	smokeSetupOK = true

	code := m.Run()

	if fakeServer != nil {
		fakeServer.Close()
	}
	os.Exit(code)
}

func ensureSection(t *testing.T) int64 {
	t.Helper()
	if smokeSectID != 0 {
		return smokeSectID
	}
	smokeSectID = testSection(t, smokeCli, smokeCfg.ProjectID, smokeCfg.SuiteID)
	return smokeSectID
}

// ---------------------------------------------------------------------------
// Smoke 1: Update → Snapshot → Rollback → Original restored
// ---------------------------------------------------------------------------

func TestSmoke_UpdateRollback(t *testing.T) {
	ctx := context.Background()
	sectionID := ensureSection(t)

	// 1. Create a test case.
	c := testCase(t, smokeCli, sectionID, "update-rollback")
	t.Logf("Original case: ID=%d Title=%q Priority=%d", c.ID, c.Title, c.PriorityID)

	// 2. Set HOME to temp dir for isolated snap store.
	home := t.TempDir()
	t.Setenv("HOME", home)

	// 3. Take snapshot before update.
	store, err := snap.NewStore()
	require.NoError(t, err)
	manifest, err := snap.LoadManifest(store)
	require.NoError(t, err)

	meta := snap.BuildMeta(snap.OpUpdate, "case", []int64{c.ID}, snap.Tier1, smokeCfg.ProjectID, smokeCfg.SuiteID, "", []string{"cases", "update"})
	snapObj, err := snap.TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return smokeCli.GetCase(ctx, c.ID)
	})
	require.NoError(t, err)
	snapID := snapObj.Meta.ID
	t.Logf("Snapshot taken: %s", snapID)

	// 4. Mutate the case on the real server.
	newTitle := "[smoke] update-rollback MUTATED"
	newPriority := int64(1)
	_, err = smokeCli.UpdateCase(ctx, c.ID, &data.UpdateCaseRequest{
		Title:      &newTitle,
		PriorityID: &newPriority,
	})
	require.NoError(t, err)
	t.Logf("Case mutated: title=%q priority=%d", newTitle, newPriority)

	// Verify mutation.
	mutated, err := smokeCli.GetCase(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, newTitle, mutated.Title)
	assert.Equal(t, int64(1), mutated.PriorityID)

	// 5. Rollback.
	result, err := snap.Rollback(ctx, smokeCli, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	t.Logf("Rollback result: %s", result.Message)

	// 6. Verify original state on remote server.
	restored, err := smokeCli.GetCase(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.Title, restored.Title, "Title should be restored")
	assert.Equal(t, c.PriorityID, restored.PriorityID, "Priority should be restored")
	t.Logf("Verified: case %d restored to original state", c.ID)
}

// ---------------------------------------------------------------------------
// Smoke 2: Delete → Snapshot → Delete → Rollback → Re-created (new ID)
// ---------------------------------------------------------------------------

func TestSmoke_DeleteRollback(t *testing.T) {
	ctx := context.Background()
	sectionID := ensureSection(t)

	// 1. Create case.
	c := testCase(t, smokeCli, sectionID, "delete-rollback")
	origID := c.ID
	t.Logf("Original case: ID=%d", origID)

	// 2. Isolated snap store.
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := snap.NewStore()
	require.NoError(t, err)
	manifest, err := snap.LoadManifest(store)
	require.NoError(t, err)

	// 3. Snapshot before delete.
	meta := snap.BuildMeta(snap.OpDelete, "case", []int64{origID}, snap.Tier2, smokeCfg.ProjectID, smokeCfg.SuiteID, "", []string{"cases", "delete"}, "")
	snapObj, err := snap.TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return smokeCli.GetCase(ctx, origID)
	})
	require.NoError(t, err)
	snapID := snapObj.Meta.ID
	t.Logf("Snapshot taken: %s", snapID)

	// 4. Delete the case on the real server.
	require.NoError(t, smokeCli.DeleteCase(ctx, origID))
	t.Logf("Case %d deleted", origID)

	// 5. Rollback → re-creates with a new ID (Tier 2).
	result, err := snap.Rollback(ctx, smokeCli, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotZero(t, result.NewEntityID)
	t.Logf("Rollback result: %s (new ID=%d)", result.Message, result.NewEntityID)

	// 6. Verify the re-created case.
	recreated, err := smokeCli.GetCase(ctx, result.NewEntityID)
	require.NoError(t, err)
	assert.Equal(t, c.Title, recreated.Title)
	t.Logf("Verified: case re-created as ID=%d with original title", result.NewEntityID)

	// 7. Cleanup: delete the re-created case (testCase cleanup targets origID which is gone).
	t.Cleanup(func() {
		if err := smokeCli.DeleteCase(context.Background(), result.NewEntityID); err != nil {
			t.Logf("cleanup: delete re-created case %d failed: %v", result.NewEntityID, err)
		} else {
			t.Logf("cleanup: deleted re-created case %d", result.NewEntityID)
		}
	})
}

// ---------------------------------------------------------------------------
// Smoke 3: Add → Snapshot + FinalizeAdd → Rollback → Created case deleted
// ---------------------------------------------------------------------------

func TestSmoke_AddRollback(t *testing.T) {
	ctx := context.Background()
	sectionID := ensureSection(t)

	// 1. Isolated snap store.
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := snap.NewStore()
	require.NoError(t, err)
	manifest, err := snap.LoadManifest(store)
	require.NoError(t, err)

	// 2. Snapshot for add (no fetchFn).
	meta := snap.BuildMeta(snap.OpAdd, "case", nil, snap.Tier2, smokeCfg.ProjectID, smokeCfg.SuiteID, "", []string{"cases", "add"}, "")
	snapObj, err := snap.TakeSnapshot(ctx, store, manifest, meta, nil)
	require.NoError(t, err)
	snapID := snapObj.Meta.ID

	// 3. Create case on the real server.
	created, err := smokeCli.AddCase(ctx, sectionID, &data.AddCaseRequest{
		Title:      "[smoke] add-rollback",
		PriorityID: 2,
		TypeID:     1,
	})
	require.NoError(t, err)
	t.Logf("Created case ID=%d", created.ID)

	// 4. FinalizeAdd.
	require.NoError(t, snapObj.FinalizeAdd(created.ID))
	t.Logf("FinalizeAdd: createdID=%d recorded in snapshot", created.ID)

	// 5. Rollback → deletes the created case.
	result, err := snap.Rollback(ctx, smokeCli, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	t.Logf("Rollback result: %s", result.Message)

	// 6. Verify case was deleted.
	_, err = smokeCli.GetCase(ctx, created.ID)
	assert.Error(t, err, "Case should be deleted after rollback")
	t.Logf("Verified: case %d no longer accessible (deleted by rollback)", created.ID)
}

// ---------------------------------------------------------------------------
// Smoke 4: Custom name → Snapshot → List → Info → Delete cycle
// ---------------------------------------------------------------------------

func TestSmoke_SnapManagementCycle(t *testing.T) {
	ctx := context.Background()
	sectionID := ensureSection(t)

	// 1. Create case.
	c := testCase(t, smokeCli, sectionID, "management-cycle")

	// 2. Isolated snap store with custom name.
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := snap.NewStore()
	require.NoError(t, err)
	manifest, err := snap.LoadManifest(store)
	require.NoError(t, err)

	meta := snap.BuildMeta(snap.OpUpdate, "case", []int64{c.ID}, snap.Tier1, smokeCfg.ProjectID, smokeCfg.SuiteID, "before-smoke-test", []string{"cases", "update"}, "")
	snapObj, err := snap.TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return smokeCli.GetCase(ctx, c.ID)
	})
	require.NoError(t, err)
	snapID := snapObj.Meta.ID
	t.Logf("Snapshot with custom name: %s", snapID)

	// 3. List — should show 1 entry.
	entries := manifest.All()
	assert.Len(t, entries, 1)
	assert.Equal(t, "before-smoke-test", entries[0].Name)
	assert.Equal(t, snap.CatCustom, entries[0].Category)
	t.Logf("List: %d entries, first=%s category=%s", len(entries), entries[0].Name, entries[0].Category)

	// 4. Info — load meta from store.
	loadedMeta, err := store.LoadMeta(snapID)
	require.NoError(t, err)
	assert.Equal(t, snap.OpUpdate, loadedMeta.Operation)
	assert.Equal(t, "case", loadedMeta.EntityType)
	assert.Equal(t, snap.StatusAvailable, loadedMeta.Status)

	// 5. Delete the snapshot.
	require.NoError(t, store.Delete(snapID))
	require.NoError(t, manifest.Remove(snapID))
	assert.False(t, store.Exists(snapID))
	assert.Nil(t, manifest.Find(snapID))
	t.Logf("Snapshot %s deleted", snapID)
}

// ---------------------------------------------------------------------------
// Smoke 5: Double rollback protection on real server
// ---------------------------------------------------------------------------

func TestSmoke_DoubleRollbackBlocked(t *testing.T) {
	ctx := context.Background()
	sectionID := ensureSection(t)

	c := testCase(t, smokeCli, sectionID, "double-rollback")

	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := snap.NewStore()
	require.NoError(t, err)
	manifest, err := snap.LoadManifest(store)
	require.NoError(t, err)

	meta := snap.BuildMeta(snap.OpUpdate, "case", []int64{c.ID}, snap.Tier1, smokeCfg.ProjectID, smokeCfg.SuiteID, "", []string{"cases", "update"}, "")
	snapObj, err := snap.TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return smokeCli.GetCase(ctx, c.ID)
	})
	require.NoError(t, err)
	snapID := snapObj.Meta.ID

	// Mutate.
	newTitle := "[smoke] double-rollback MUTATED"
	_, err = smokeCli.UpdateCase(ctx, c.ID, &data.UpdateCaseRequest{Title: &newTitle})
	require.NoError(t, err)

	// First rollback — success.
	_, err = snap.Rollback(ctx, smokeCli, store, manifest, snapID)
	require.NoError(t, err)
	t.Logf("First rollback succeeded")

	// Second rollback — blocked.
	_, err = snap.Rollback(ctx, smokeCli, store, manifest, snapID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rolled_back")
	t.Logf("Second rollback correctly blocked: %v", err)
}

// ---------------------------------------------------------------------------
// Smoke 6: GC cleans orphans from real snap store
// ---------------------------------------------------------------------------

func TestSmoke_GCOrphans(t *testing.T) {
	ctx := context.Background()
	sectionID := ensureSection(t)

	c := testCase(t, smokeCli, sectionID, "gc-test")

	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := snap.NewStore()
	require.NoError(t, err)
	manifest, err := snap.LoadManifest(store)
	require.NoError(t, err)

	// Create a tracked snapshot.
	meta := snap.BuildMeta(snap.OpUpdate, "case", []int64{c.ID}, snap.Tier1, smokeCfg.ProjectID, smokeCfg.SuiteID, "", []string{"cases", "update"}, "")
	_, err = snap.TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return smokeCli.GetCase(ctx, c.ID)
	})
	require.NoError(t, err)

	// Create an orphan directly.
	orphanMeta := &snap.Meta{
		ID:        "cases/orphan_smoke",
		Category:  "cases",
		Operation: snap.OpUpdate,
		Status:    snap.StatusAvailable,
	}
	require.NoError(t, store.SaveMeta(orphanMeta))
	assert.True(t, store.Exists("cases/orphan_smoke"))

	// GC.
	cleaned, err := store.CleanOrphans(manifest.ManifestIDs())
	require.NoError(t, err)
	assert.Equal(t, 1, cleaned)
	assert.False(t, store.Exists("cases/orphan_smoke"))
	t.Logf("GC cleaned %d orphans", cleaned)
}
