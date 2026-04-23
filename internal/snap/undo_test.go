package snap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// CanUndo
// ---------------------------------------------------------------------------

func TestCanUndo_RolledBackDelete_WithNewID(t *testing.T) {
	meta := &Meta{
		Status:    StatusRolledBack,
		Operation: OpDelete,
		RollbackLog: []RollbackLogEntry{
			{Type: "case", ID: 42, NewID: 200, Status: RBRestored},
		},
	}
	assert.True(t, CanUndo(meta))
}

func TestCanUndo_RolledBackDelete_NoNewID(t *testing.T) {
	meta := &Meta{
		Status:    StatusRolledBack,
		Operation: OpDelete,
		RollbackLog: []RollbackLogEntry{
			{Type: "case", ID: 42, Status: RBRestored}, // old rollback without NewID
		},
	}
	assert.False(t, CanUndo(meta))
}

func TestCanUndo_RolledBackAdd(t *testing.T) {
	meta := &Meta{
		Status:    StatusRolledBack,
		Operation: OpAdd,
	}
	assert.False(t, CanUndo(meta))
}

func TestCanUndo_RolledBackUpdate(t *testing.T) {
	meta := &Meta{
		Status:    StatusRolledBack,
		Operation: OpUpdate,
	}
	assert.False(t, CanUndo(meta))
}

func TestCanUndo_Available(t *testing.T) {
	meta := &Meta{
		Status:    StatusAvailable,
		Operation: OpDelete,
	}
	assert.False(t, CanUndo(meta), "must be rolled_back")
}

func TestCanUndo_SyncOp(t *testing.T) {
	meta := &Meta{
		Status:    StatusRolledBack,
		Operation: OpSyncFull,
	}
	assert.False(t, CanUndo(meta))
}

// ---------------------------------------------------------------------------
// UndoRollback — case delete
// ---------------------------------------------------------------------------

func TestUndoRollback_CaseDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 42, Title: "Deleted Case", SectionID: 100}
	snapID := "cases/20260414T120000_delete_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpDelete,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusRolledBack,
		DataFile: "data.json", RollbackTier: Tier2, Timestamp: time.Now().UTC(),
		RollbackLog: []RollbackLogEntry{
			{Type: "case", ID: 42, NewID: 200, Status: RBRestored},
		},
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteCaseFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := UndoRollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.Undoable)
	assert.Equal(t, int64(200), deletedID)
	assert.Contains(t, result.DeletedIDs, int64(200))
	assert.Contains(t, result.Message, "200")

	// Status should be reset to available.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusAvailable, entry.Status)

	// RollbackLog should be cleared.
	updatedMeta, err := store.LoadMeta(snapID)
	require.NoError(t, err)
	assert.Empty(t, updatedMeta.RollbackLog)
}

func TestUndoRollback_CaseDelete_AlreadyGone(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260414T120000_delete_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpDelete,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusRolledBack,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
		RollbackLog: []RollbackLogEntry{
			{Type: "case", ID: 42, NewID: 200, Status: RBRestored},
		},
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		deleteCaseFunc: func(_ context.Context, _ int64) error {
			return fmt.Errorf("API returned 404: Not Found")
		},
	}

	result, err := UndoRollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "already gone")
}

// ---------------------------------------------------------------------------
// UndoRollback — section cascade
// ---------------------------------------------------------------------------

func TestUndoRollback_SectionCascade(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	cascade := CascadeData{
		Section: data.Section{ID: 10, Name: "Auth"},
		Cases:   []data.Case{{ID: 100, Title: "Login"}, {ID: 101, Title: "Logout"}},
	}
	snapID := "sections/20260414T120000_delete_10"
	meta := &Meta{
		ID: snapID, Category: Category("sections"), Operation: OpDelete,
		EntityType: "section", EntityIDs: []int64{10}, Status: StatusRolledBack,
		DataFile: "data.json", ProjectID: 1, Timestamp: time.Now().UTC(),
		RollbackLog: []RollbackLogEntry{
			{Type: "section", ID: 10, NewID: 50, Status: RBRestored},
			{Type: "case", ID: 100, NewID: 201, Status: RBRestored},
			{Type: "case", ID: 101, NewID: 202, Status: RBRestored},
		},
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", cascade)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	var deletedSectionID int64
	api := &mockCasesAPI{
		deleteSectionFunc: func(_ context.Context, id int64) error {
			deletedSectionID = id
			return nil
		},
	}

	result, err := UndoRollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(50), deletedSectionID)
	assert.Contains(t, result.DeletedIDs, int64(50))

	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusAvailable, entry.Status)
}

// ---------------------------------------------------------------------------
// UndoRollback — project delete
// ---------------------------------------------------------------------------

func TestUndoRollback_ProjectDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "projects/20260414T120000_delete_7"
	meta := &Meta{
		ID: snapID, Category: Category("projects"), Operation: OpDelete,
		EntityType: "project", EntityIDs: []int64{7}, Status: StatusRolledBack,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
		RollbackLog: []RollbackLogEntry{
			{Type: "project", ID: 7, NewID: 50, Status: RBRestored},
		},
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteProjectFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := UndoRollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(50), deletedID)

	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusAvailable, entry.Status)
}

// ---------------------------------------------------------------------------
// UndoRollback — error cases
// ---------------------------------------------------------------------------

func TestUndoRollback_NotRolledBack(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260414T120000_update_1"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	_, err = UndoRollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status is \"available\"")
}

func TestUndoRollback_AddNotUndoable(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260414T120000_add_bulk_0"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpAdd,
		EntityType: "case", Status: StatusRolledBack,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	_, err = UndoRollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo not available")
}

func TestUndoRollback_UpdateNotUndoable(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260414T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", Status: StatusRolledBack,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	_, err = UndoRollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo not available")
}

func TestUndoRollback_APIError(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260414T120000_delete_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpDelete,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusRolledBack,
		Timestamp: time.Now().UTC(),
		RollbackLog: []RollbackLogEntry{
			{Type: "case", ID: 42, NewID: 200, Status: RBRestored},
		},
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		deleteCaseFunc: func(_ context.Context, _ int64) error {
			return fmt.Errorf("API error 500: Internal Server Error")
		},
	}

	_, err = UndoRollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 500")

	// Status should NOT change on error.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusRolledBack, entry.Status)
}

// ---------------------------------------------------------------------------
// Rollback populates NewID (regression)
// ---------------------------------------------------------------------------

func TestRollback_CaseDelete_PopulatesNewID(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 99, Title: "Deleted", SectionID: 200, TypeID: 2}
	snapID := "cases/20260414T120000_delete_99"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpDelete,
		EntityType: "case", EntityIDs: []int64{99}, Status: StatusAvailable,
		DataFile: "data.json", RollbackTier: Tier2, Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		addCaseFunc: func(_ context.Context, _ int64, req *data.AddCaseRequest) (*data.Case, error) {
			return &data.Case{ID: 300, Title: req.Title}, nil
		},
	}

	_, err = Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)

	// Verify NewID is persisted in rollback log.
	updatedMeta, err := store.LoadMeta(snapID)
	require.NoError(t, err)
	require.Len(t, updatedMeta.RollbackLog, 1)
	assert.Equal(t, int64(300), updatedMeta.RollbackLog[0].NewID)
}
