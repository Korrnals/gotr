package snap

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeMeta is a small helper that materializes a snapshot directory with a
// valid meta.json under store.BaseDir() so RepairManifest sees it as
// "present on disk".
func writeMeta(t *testing.T, store *Store, meta *Meta) {
	t.Helper()
	require.NoError(t, store.SaveMeta(meta))
}

func TestRepairManifest_Empty(t *testing.T) {
	store, err := NewStoreAt(t.TempDir())
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	res, err := RepairManifest(store, manifest, false)
	require.NoError(t, err)

	assert.False(t, res.HasChanges())
	assert.Empty(t, res.Added)
	assert.Empty(t, res.Removed)
	assert.Empty(t, res.MetaErrors)
}

func TestRepairManifest_AddsMissingEntry(t *testing.T) {
	store, err := NewStoreAt(t.TempDir())
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	// Snapshot directory exists on disk with valid meta, but manifest is empty.
	meta := &Meta{
		ID:           "cases/20260101T000000_update_42",
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		EntityIDs:    []int64{42},
		RollbackTier: Tier1,
		Status:       StatusAvailable,
		Timestamp:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	writeMeta(t, store, meta)

	res, err := RepairManifest(store, manifest, false)
	require.NoError(t, err)

	require.Len(t, res.Added, 1)
	assert.Equal(t, "cases/20260101T000000_update_42", res.Added[0].SnapID)
	assert.Equal(t, "add", res.Added[0].Op)
	assert.True(t, res.HasChanges())

	// Manifest now has the entry persisted.
	reloaded, err := LoadManifest(store)
	require.NoError(t, err)
	require.Equal(t, 1, reloaded.Len())
	assert.NotNil(t, reloaded.Find("cases/20260101T000000_update_42"))
}

func TestRepairManifest_RemovesOrphanEntry(t *testing.T) {
	store, err := NewStoreAt(t.TempDir())
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	// Manifest has an entry that does NOT correspond to any directory on disk.
	require.NoError(t, manifest.Add(&Meta{
		ID:           "cases/ghost",
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		RollbackTier: Tier1,
		Status:       StatusAvailable,
	}))
	require.Equal(t, 1, manifest.Len())

	res, err := RepairManifest(store, manifest, false)
	require.NoError(t, err)

	require.Len(t, res.Removed, 1)
	assert.Equal(t, "cases/ghost", res.Removed[0].SnapID)
	assert.Equal(t, "remove", res.Removed[0].Op)

	reloaded, err := LoadManifest(store)
	require.NoError(t, err)
	assert.Equal(t, 0, reloaded.Len())
}

func TestRepairManifest_DryRunDoesNotPersist(t *testing.T) {
	store, err := NewStoreAt(t.TempDir())
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	// Mix: one missing-from-manifest dir + one orphan entry.
	writeMeta(t, store, &Meta{
		ID:           "cases/20260101T000000_update_42",
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		RollbackTier: Tier1,
		Status:       StatusAvailable,
	})
	require.NoError(t, manifest.Add(&Meta{
		ID:           "cases/ghost",
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		RollbackTier: Tier1,
		Status:       StatusAvailable,
	}))

	res, err := RepairManifest(store, manifest, true)
	require.NoError(t, err)

	assert.True(t, res.DryRun)
	assert.Len(t, res.Added, 1)
	assert.Len(t, res.Removed, 1)

	// Disk manifest must be untouched: still exactly the original ghost entry.
	reloaded, err := LoadManifest(store)
	require.NoError(t, err)
	require.Equal(t, 1, reloaded.Len())
	assert.NotNil(t, reloaded.Find("cases/ghost"))
	assert.Nil(t, reloaded.Find("cases/20260101T000000_update_42"))
}

func TestRepairManifest_UnreadableMetaIsReportedNotPruned(t *testing.T) {
	store, err := NewStoreAt(t.TempDir())
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	// Create a snapshot directory with an unparsable meta.json.
	badDir := filepath.Join(store.BaseDir(), "cases", "broken_snap")
	require.NoError(t, os.MkdirAll(badDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(badDir, "meta.json"), []byte("{ this is not json"), 0o644))

	res, err := RepairManifest(store, manifest, false)
	require.NoError(t, err)

	require.Len(t, res.MetaErrors, 1)
	assert.Equal(t, "cases/broken_snap", res.MetaErrors[0].SnapID)
	assert.Empty(t, res.Added)
	assert.Empty(t, res.Removed)

	// Manifest stays empty: we never auto-add or auto-remove a broken meta.
	reloaded, err := LoadManifest(store)
	require.NoError(t, err)
	assert.Equal(t, 0, reloaded.Len())
}

func TestRepairManifest_Idempotent(t *testing.T) {
	store, err := NewStoreAt(t.TempDir())
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	writeMeta(t, store, &Meta{
		ID:           "cases/20260101T000000_update_42",
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		RollbackTier: Tier1,
		Status:       StatusAvailable,
	})

	// First pass repairs.
	res1, err := RepairManifest(store, manifest, false)
	require.NoError(t, err)
	require.True(t, res1.HasChanges())

	// Second pass over the same store + a freshly-loaded manifest should be
	// a no-op.
	manifest2, err := LoadManifest(store)
	require.NoError(t, err)
	res2, err := RepairManifest(store, manifest2, false)
	require.NoError(t, err)
	assert.False(t, res2.HasChanges())
	assert.Empty(t, res2.Added)
	assert.Empty(t, res2.Removed)
}
