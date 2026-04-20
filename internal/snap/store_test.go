package snap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateID_Single(t *testing.T) {
	meta := &Meta{Operation: OpUpdate, EntityType: "case", EntityIDs: []int64{1234}}
	id := GenerateID(meta)
	assert.Contains(t, id, "_update_1234")
	assert.True(t, len(id) > len("cases/"), "ID should have category prefix")
	assert.Equal(t, Category("cases"), meta.Category)
}

func TestGenerateID_Bulk(t *testing.T) {
	meta := &Meta{Operation: OpDelete, EntityType: "case", EntityIDs: []int64{1, 2, 3}}
	id := GenerateID(meta)
	assert.Contains(t, id, "_delete_bulk_3")
	assert.Equal(t, Category("cases"), meta.Category)
}

func TestGenerateID_CustomName(t *testing.T) {
	meta := &Meta{Name: "before-migration", Operation: OpUpdate, EntityType: "case"}
	id := GenerateID(meta)
	assert.Equal(t, "custom/before-migration", id)
	assert.Equal(t, CatCustom, meta.Category)
}

func TestGenerateID_SyncOp(t *testing.T) {
	meta := &Meta{
		Operation:       OpSyncFull,
		EntityType:      "case",
		SourceProjectID: 10,
		ProjectID:       20,
	}
	id := GenerateID(meta)
	assert.Contains(t, id, "sync/")
	assert.Contains(t, id, "_full_p10_p20")
	assert.Equal(t, CatSync, meta.Category)
}

func TestResolveCategoryDir(t *testing.T) {
	tests := []struct {
		name     string
		meta     *Meta
		expected Category
	}{
		{"custom name", &Meta{Name: "my-snap", EntityType: "case"}, CatCustom},
		{"sync op", &Meta{Operation: OpSyncFull, EntityType: "case"}, CatSync},
		{"case entity", &Meta{Operation: OpUpdate, EntityType: "case"}, Category("cases")},
		{"section entity", &Meta{Operation: OpDelete, EntityType: "section"}, Category("sections")},
		{"run entity", &Meta{Operation: OpClose, EntityType: "run"}, Category("runs")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ResolveCategoryDir(tt.meta))
		})
	}
}

func TestStore_SaveAndLoadMeta(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	meta := &Meta{
		ID:           "cases/test_snap_1",
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		EntityIDs:    []int64{42},
		ProjectID:    30,
		RollbackTier: Tier1,
		Status:       StatusAvailable,
		DataFile:     "data.json",
	}

	err = store.SaveMeta(meta)
	require.NoError(t, err)

	loaded, err := store.LoadMeta("cases/test_snap_1")
	require.NoError(t, err)
	assert.Equal(t, meta.ID, loaded.ID)
	assert.Equal(t, meta.Category, loaded.Category)
	assert.Equal(t, meta.Operation, loaded.Operation)
	assert.Equal(t, meta.EntityType, loaded.EntityType)
	assert.Equal(t, meta.EntityIDs, loaded.EntityIDs)
	assert.Equal(t, meta.RollbackTier, loaded.RollbackTier)
	assert.Equal(t, meta.Status, loaded.Status)
}

func TestStore_SaveAndLoadData(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	data := map[string]interface{}{
		"id":    float64(42),
		"title": "Test case",
	}

	size, err := store.SaveData("cases/snap_data_test", "data.json", data)
	require.NoError(t, err)
	assert.Greater(t, size, int64(0))

	var loaded map[string]interface{}
	err = store.LoadData("cases/snap_data_test", "data.json", &loaded)
	require.NoError(t, err)
	assert.Equal(t, float64(42), loaded["id"])
	assert.Equal(t, "Test case", loaded["title"])
}

func TestStore_SaveData_CustomFilename(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	sections := []map[string]interface{}{{"id": float64(1), "name": "Section 1"}}
	cases := []map[string]interface{}{{"id": float64(10), "title": "Case 1"}}

	_, err = store.SaveData("sync/snap_sync_1", "sections.json", sections)
	require.NoError(t, err)
	_, err = store.SaveData("sync/snap_sync_1", "cases.json", cases)
	require.NoError(t, err)

	var loadedSections []map[string]interface{}
	err = store.LoadData("sync/snap_sync_1", "sections.json", &loadedSections)
	require.NoError(t, err)
	assert.Len(t, loadedSections, 1)

	var loadedCases []map[string]interface{}
	err = store.LoadData("sync/snap_sync_1", "cases.json", &loadedCases)
	require.NoError(t, err)
	assert.Len(t, loadedCases, 1)
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	// Create category-based snapshot directories.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "snap_a"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "snap_b"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sync", "snap_c"), 0o755))

	ids, err := store.List()
	require.NoError(t, err)
	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "cases/snap_a")
	assert.Contains(t, ids, "cases/snap_b")
	assert.Contains(t, ids, "sync/snap_c")
}

func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "to_delete"), 0o755))
	assert.True(t, store.Exists("cases/to_delete"))

	err = store.Delete("cases/to_delete")
	require.NoError(t, err)
	assert.False(t, store.Exists("cases/to_delete"))
}

func TestStore_Exists(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	assert.False(t, store.Exists("cases/nonexistent"))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "existing"), 0o755))
	assert.True(t, store.Exists("cases/existing"))
}

func TestStore_CollectOrphans(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	// Create 3 snapshot dirs, only 2 in manifest.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "snap_1"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "snap_2"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sync", "snap_3"), 0o755))

	manifestIDs := map[string]struct{}{
		"cases/snap_1": {},
		"sync/snap_3":  {},
	}

	orphans, err := store.CollectOrphans(manifestIDs)
	require.NoError(t, err)
	assert.Len(t, orphans, 1)
	assert.Contains(t, orphans, "cases/snap_2")
}

func TestStore_CleanOrphans(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "keep"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cases", "orphan"), 0o755))

	manifestIDs := map[string]struct{}{
		"cases/keep": {},
	}

	cleaned, err := store.CleanOrphans(manifestIDs)
	require.NoError(t, err)
	assert.Equal(t, 1, cleaned)
	assert.False(t, store.Exists("cases/orphan"))
	assert.True(t, store.Exists("cases/keep"))
}
