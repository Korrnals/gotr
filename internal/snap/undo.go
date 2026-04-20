package snap

import (
	"context"
	"fmt"
)

// UndoResult holds the outcome of an undo-rollback operation.
type UndoResult struct {
	SnapID     string
	Success    bool
	Message    string
	Undoable   bool    // whether undo was possible for this snapshot
	DeletedIDs []int64 // IDs of entities deleted during undo
}

// CanUndo reports whether a rolled-back snapshot can be undone.
// Only delete-operation rollbacks that produced re-created entities are undoable.
func CanUndo(meta *Meta) bool {
	if meta.Status != StatusRolledBack {
		return false
	}
	// Only delete operations create new entities that can be cleaned up.
	if meta.Operation != OpDelete {
		return false
	}
	// Must have at least one log entry with a new ID.
	for _, e := range meta.RollbackLog {
		if e.NewID > 0 && e.Status == RBRestored {
			return true
		}
	}
	return false
}

// UndoRollback reverses a previous rollback by cleaning up re-created entities.
//
// Supported flows:
//   - case delete rollback  → deletes re-created case (NewID)
//   - section delete cascade → deletes re-created section (NewID cascades children)
//   - project delete rollback → deletes re-created project (NewID)
//
// Unsupported (returns error):
//   - add rollbacks: deleted entity cannot be restored
//   - update rollbacks: post-mutation values are not preserved
//   - sync rollbacks: re-syncing is not automated
//
// On success the snapshot status is reset to "available".
func UndoRollback(ctx context.Context, api RollbackAPI, store *Store, manifest *Manifest, snapID string) (*UndoResult, error) {
	meta, err := store.LoadMeta(snapID)
	if err != nil {
		return nil, fmt.Errorf("load snapshot: %w", err)
	}

	if meta.Status != StatusRolledBack {
		return nil, fmt.Errorf("snapshot %q status is %q, expected %q", snapID, meta.Status, StatusRolledBack)
	}

	result := &UndoResult{SnapID: snapID}

	if !CanUndo(meta) {
		result.Undoable = false
		result.Message = undoNotAvailableReason(meta)
		return result, fmt.Errorf("undo not available: %s", result.Message)
	}

	result.Undoable = true

	switch meta.EntityType {
	case "case":
		err = undoCaseDelete(ctx, api, meta, result)
	case "section":
		err = undoSectionCascade(ctx, api, meta, result)
	case "project":
		err = undoProjectDelete(ctx, api, meta, result)
	default:
		return result, fmt.Errorf("undo not implemented for entity type %q", meta.EntityType)
	}

	if err != nil {
		return result, err
	}

	// Reset status to available; clear rollback log.
	meta.Status = StatusAvailable
	meta.RollbackLog = nil
	if err := store.SaveMeta(meta); err != nil {
		return result, fmt.Errorf("save meta: %w", err)
	}
	if err := manifest.UpdateStatus(snapID, StatusAvailable); err != nil {
		return result, fmt.Errorf("update manifest: %w", err)
	}

	result.Success = true
	return result, nil
}

// undoCaseDelete deletes the re-created case (NewID from rollback log).
func undoCaseDelete(ctx context.Context, api RollbackAPI, meta *Meta, result *UndoResult) error {
	for _, e := range meta.RollbackLog {
		if e.Type == "case" && e.NewID > 0 && e.Status == RBRestored {
			if err := api.DeleteCase(ctx, e.NewID); err != nil {
				if isGoneError(err) {
					result.Message = fmt.Sprintf("Case %d already gone — undo is a no-op", e.NewID)
					result.DeletedIDs = append(result.DeletedIDs, e.NewID)
					return nil
				}
				return fmt.Errorf("delete re-created case %d: %w", e.NewID, err)
			}
			result.DeletedIDs = append(result.DeletedIDs, e.NewID)
		}
	}
	if len(result.DeletedIDs) == 0 {
		return fmt.Errorf("no re-created case found in rollback log")
	}
	result.Message = fmt.Sprintf("Undo: deleted re-created case(s) %v, snapshot is available again", result.DeletedIDs)
	return nil
}

// undoSectionCascade deletes the re-created section (API cascades children).
func undoSectionCascade(ctx context.Context, api RollbackAPI, meta *Meta, result *UndoResult) error {
	for _, e := range meta.RollbackLog {
		if e.Type == "section" && e.NewID > 0 && e.Status == RBRestored {
			if err := api.DeleteSection(ctx, e.NewID); err != nil {
				if isGoneError(err) {
					result.Message = fmt.Sprintf("Section %d already gone — undo is a no-op", e.NewID)
					result.DeletedIDs = append(result.DeletedIDs, e.NewID)
					return nil
				}
				return fmt.Errorf("delete re-created section %d: %w", e.NewID, err)
			}
			result.DeletedIDs = append(result.DeletedIDs, e.NewID)
			result.Message = fmt.Sprintf("Undo: deleted re-created section %d (+ child cases), snapshot is available again", e.NewID)
			return nil
		}
	}
	return fmt.Errorf("no re-created section found in rollback log")
}

// undoProjectDelete deletes the re-created project.
func undoProjectDelete(ctx context.Context, api RollbackAPI, meta *Meta, result *UndoResult) error {
	for _, e := range meta.RollbackLog {
		if e.Type == "project" && e.NewID > 0 && e.Status == RBRestored {
			if err := api.DeleteProject(ctx, e.NewID); err != nil {
				if isGoneError(err) {
					result.Message = fmt.Sprintf("Project %d already gone — undo is a no-op", e.NewID)
					result.DeletedIDs = append(result.DeletedIDs, e.NewID)
					return nil
				}
				return fmt.Errorf("delete re-created project %d: %w", e.NewID, err)
			}
			result.DeletedIDs = append(result.DeletedIDs, e.NewID)
			result.Message = fmt.Sprintf("Undo: deleted re-created project %d, snapshot is available again", e.NewID)
			return nil
		}
	}
	return fmt.Errorf("no re-created project found in rollback log")
}

// undoNotAvailableReason returns a human-readable explanation.
func undoNotAvailableReason(meta *Meta) string {
	switch {
	case meta.Operation == OpAdd:
		return fmt.Sprintf("%s was deleted during rollback — entity data is not preserved for re-creation", meta.EntityType)
	case meta.Operation == OpUpdate:
		return fmt.Sprintf("%s fields were restored to pre-mutation values — post-mutation values are not preserved", meta.EntityType)
	case meta.IsSyncOp():
		return "sync entities were deleted during rollback — re-syncing requires a new sync operation"
	case meta.Operation == OpCopy:
		return fmt.Sprintf("%s was deleted during rollback — entity data is not preserved for re-creation", meta.EntityType)
	default:
		return fmt.Sprintf("undo is not supported for %s %s", meta.EntityType, meta.Operation)
	}
}
