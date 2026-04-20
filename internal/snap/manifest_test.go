package snap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifest_AddAndFind(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)
	assert.Equal(t, 0, manifest.Len())

	meta := &Meta{
		ID:           "cases/snap_1",
		Name:         "my-snap",
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		RollbackTier: Tier1,
		Status:       StatusAvailable,
	}

	err = manifest.Add(meta)
	require.NoError(t, err)
	assert.Equal(t, 1, manifest.Len())

	// Find by ID.
	entry := manifest.Find("cases/snap_1")
	require.NotNil(t, entry)
	assert.Equal(t, "cases/snap_1", entry.ID)
	assert.Equal(t, "my-snap", entry.Name)
	assert.Equal(t, Category("cases"), entry.Category)

	// Find by name.
	entry = manifest.Find("my-snap")
	require.NotNil(t, entry)
	assert.Equal(t, "cases/snap_1", entry.ID)

	// Not found.
	assert.Nil(t, manifest.Find("nonexistent"))
}

func TestManifest_Remove(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta1 := &Meta{ID: "cases/snap_1", Category: Category("cases"), Operation: OpUpdate, EntityType: "case", Status: StatusAvailable}
	meta2 := &Meta{ID: "sections/snap_2", Category: Category("sections"), Operation: OpDelete, EntityType: "section", Status: StatusAvailable}

	require.NoError(t, manifest.Add(meta1))
	require.NoError(t, manifest.Add(meta2))
	assert.Equal(t, 2, manifest.Len())

	require.NoError(t, manifest.Remove("cases/snap_1"))
	assert.Equal(t, 1, manifest.Len())
	assert.Nil(t, manifest.Find("cases/snap_1"))
	assert.NotNil(t, manifest.Find("sections/snap_2"))
}

func TestManifest_UpdateStatus(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := &Meta{ID: "cases/snap_1", Category: Category("cases"), Operation: OpUpdate, EntityType: "case", Status: StatusAvailable}
	require.NoError(t, manifest.Add(meta))

	err = manifest.UpdateStatus("cases/snap_1", StatusRolledBack)
	require.NoError(t, err)

	entry := manifest.Find("cases/snap_1")
	require.NotNil(t, entry)
	assert.Equal(t, StatusRolledBack, entry.Status)

	// Non-existent ID.
	err = manifest.UpdateStatus("nope", StatusExpired)
	assert.Error(t, err)
}

func TestManifest_UpdateLabel(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := &Meta{ID: "cases/snap_1", Category: Category("cases"), Operation: OpUpdate, EntityType: "case", Status: StatusAvailable, Label: "old"}
	require.NoError(t, manifest.Add(meta))

	err = manifest.UpdateLabel("cases/snap_1", "pinned_old")
	require.NoError(t, err)

	entry := manifest.Find("cases/snap_1")
	require.NotNil(t, entry)
	assert.Equal(t, "pinned_old", entry.Label)

	manifest2, err := LoadManifest(store)
	require.NoError(t, err)
	entry2 := manifest2.Find("cases/snap_1")
	require.NotNil(t, entry2)
	assert.Equal(t, "pinned_old", entry2.Label)

	err = manifest.UpdateLabel("nope", "x")
	assert.Error(t, err)
}

func TestManifest_ListByEntity(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	require.NoError(t, manifest.Add(&Meta{ID: "cases/s1", Category: Category("cases"), EntityType: "case", Status: StatusAvailable}))
	require.NoError(t, manifest.Add(&Meta{ID: "sections/s2", Category: Category("sections"), EntityType: "section", Status: StatusAvailable}))
	require.NoError(t, manifest.Add(&Meta{ID: "cases/s3", Category: Category("cases"), EntityType: "case", Status: StatusAvailable}))

	cases := manifest.ListByEntity("case")
	assert.Len(t, cases, 2)

	sections := manifest.ListByEntity("section")
	assert.Len(t, sections, 1)
}

func TestManifest_ListByCategory(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	require.NoError(t, manifest.Add(&Meta{ID: "cases/s1", Category: Category("cases"), EntityType: "case", Status: StatusAvailable}))
	require.NoError(t, manifest.Add(&Meta{ID: "sync/s2", Category: CatSync, EntityType: "case", Operation: OpSyncFull, Status: StatusAvailable}))
	require.NoError(t, manifest.Add(&Meta{ID: "custom/my-snap", Category: CatCustom, Name: "my-snap", EntityType: "case", Status: StatusAvailable}))

	cases := manifest.ListByCategory(Category("cases"))
	assert.Len(t, cases, 1)

	syncs := manifest.ListByCategory(CatSync)
	assert.Len(t, syncs, 1)

	custom := manifest.ListByCategory(CatCustom)
	assert.Len(t, custom, 1)
}

func TestManifest_ManifestIDs(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	require.NoError(t, manifest.Add(&Meta{ID: "cases/s1", Category: Category("cases"), Status: StatusAvailable}))
	require.NoError(t, manifest.Add(&Meta{ID: "sync/s2", Category: CatSync, Status: StatusAvailable}))

	ids := manifest.ManifestIDs()
	assert.Len(t, ids, 2)
	_, ok1 := ids["cases/s1"]
	assert.True(t, ok1)
	_, ok2 := ids["sync/s2"]
	assert.True(t, ok2)
}

func TestManifest_Persistence(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := &Meta{ID: "runs/persist_1", Category: Category("runs"), Operation: OpDelete, EntityType: "run", Status: StatusAvailable}
	require.NoError(t, manifest.Add(meta))

	// Reload from disk.
	manifest2, err := LoadManifest(store)
	require.NoError(t, err)
	assert.Equal(t, 1, manifest2.Len())

	entry := manifest2.Find("runs/persist_1")
	require.NotNil(t, entry)
	assert.Equal(t, OpDelete, entry.Operation)
	assert.Equal(t, Category("runs"), entry.Category)
}

func TestManifest_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	// Remove manifest file explicitly to ensure clean state.
	os.Remove(dir + "/manifest.json")

	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)
	assert.Equal(t, 0, manifest.Len())
	assert.Nil(t, manifest.Latest())
}

func TestManifest_Add_CopiesProjectIDs(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := &Meta{
		ID:              "cases/proj_test",
		Category:        Category("cases"),
		Operation:       OpUpdate,
		EntityType:      "case",
		Status:          StatusAvailable,
		ProjectID:       42,
		SourceProjectID: 10,
	}

	require.NoError(t, manifest.Add(meta))

	entry := manifest.Find("cases/proj_test")
	require.NotNil(t, entry)
	assert.Equal(t, int64(42), entry.ProjectID, "ProjectID should be copied from Meta")
	assert.Equal(t, int64(10), entry.SourceProjectID, "SourceProjectID should be copied from Meta")
}

func TestManifest_LegacyEntry_ZeroProjectIDs(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	// Legacy meta without project IDs.
	meta := &Meta{
		ID:         "suites/legacy_1",
		Category:   Category("suites"),
		Operation:  OpUpdate,
		EntityType: "suite",
		Status:     StatusAvailable,
	}

	require.NoError(t, manifest.Add(meta))

	entry := manifest.Find("suites/legacy_1")
	require.NotNil(t, entry)
	assert.Equal(t, int64(0), entry.ProjectID, "legacy entry should have zero ProjectID")
	assert.Equal(t, int64(0), entry.SourceProjectID, "legacy entry should have zero SourceProjectID")
}

func TestManifest_ProjectIDs_PersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := &Meta{
		ID:              "cases/roundtrip_1",
		Category:        Category("cases"),
		Operation:       OpUpdate,
		EntityType:      "case",
		Status:          StatusAvailable,
		ProjectID:       100,
		SourceProjectID: 50,
	}
	require.NoError(t, manifest.Add(meta))

	// Reload from disk.
	manifest2, err := LoadManifest(store)
	require.NoError(t, err)

	entry := manifest2.Find("cases/roundtrip_1")
	require.NotNil(t, entry)
	assert.Equal(t, int64(100), entry.ProjectID, "ProjectID should survive round-trip")
	assert.Equal(t, int64(50), entry.SourceProjectID, "SourceProjectID should survive round-trip")
}

func TestManifest_OmitEmpty_LegacyRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	// Legacy meta without project IDs — they should not appear in JSON.
	meta := &Meta{
		ID:         "suites/omit_1",
		Category:   Category("suites"),
		Operation:  OpDelete,
		EntityType: "suite",
		Status:     StatusAvailable,
	}
	require.NoError(t, manifest.Add(meta))

	// Read raw JSON to verify omitempty.
	raw, err := os.ReadFile(store.BaseDir() + "/" + ManifestFile)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "project_id", "zero ProjectID should be omitted from JSON")
	assert.NotContains(t, string(raw), "source_project_id", "zero SourceProjectID should be omitted from JSON")

	// Reload and verify zero values.
	manifest2, err := LoadManifest(store)
	require.NoError(t, err)
	entry := manifest2.Find("suites/omit_1")
	require.NotNil(t, entry)
	assert.Equal(t, int64(0), entry.ProjectID)
	assert.Equal(t, int64(0), entry.SourceProjectID)
}
