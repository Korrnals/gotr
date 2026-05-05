package snap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"go.uber.org/zap"
)

// EntityTypeAttachments is the snap.Meta.EntityType used for bulk
// attachment-cleanup operations.
const EntityTypeAttachments = "attachments"

// CleanupAttachmentsAPI is the API surface required to back up and restore
// bulk-deleted attachments. It is satisfied by *client.HTTPClient.
type CleanupAttachmentsAPI interface {
	DownloadAttachment(ctx context.Context, attachmentID int64) (io.ReadCloser, error)
	AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlan(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanEntry(ctx context.Context, planID int64, entryID, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToResult(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToRun(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error)
}

// CleanupAttachmentEntry is a single attachment record stored in
// data.json of an attachments-cleanup snapshot. It carries everything
// needed to re-upload the file to its original parent on rollback.
type CleanupAttachmentEntry struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type,omitempty"`
	CreatedOn    int64  `json:"created_on"`
	EntityType   string `json:"entity_type"`            // "case"|"plan"|"plan_entry"|"result"|"run"|"test"
	EntityID     int64  `json:"entity_id"`              // parent entity ID (0 for plan_entry — see ParentPlanID/EntryID)
	ParentPlanID int64  `json:"parent_plan_id,omitempty"`
	EntryID      string `json:"entry_id,omitempty"`
	File         string `json:"file"`                   // relative path under <snap>/files/
	Compressed   bool   `json:"compressed"`
}

// CleanupAttachmentsData is the on-disk shape of data.json for an
// attachments-cleanup snapshot.
type CleanupAttachmentsData struct {
	Attachments []CleanupAttachmentEntry `json:"attachments"`
}

// cleanupFilesDir is the subdirectory that stores attachment binaries
// inside an attachments-cleanup snapshot.
const cleanupFilesDir = "files"

// BackupAttachmentsForCleanup is the v1-compatible entry point for
// downloading every attachment binary into the snapshot directory. It
// now delegates to BackupAttachmentsForCleanupV2 with sequential
// downloads (Concurrency=1) so legacy callers transparently gain
// SHA-256 integrity and mapping.json without any signature change.
//
// Returns the number of saved attachments and the total bytes written.
func BackupAttachmentsForCleanup(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	store *Store,
	snapID string,
	atts []data.Attachment,
	compress bool,
) (saved int, totalBytes int64, err error) {
	return BackupAttachmentsForCleanupV2(ctx, api, store, snapID, atts, BackupOptions{
		Compress:    compress,
		Concurrency: 1,
	})
}

// CleanupRollbackResult holds the per-entry rollback outcome plus an
// old-id → new-id mapping populated when re-upload succeeded.
type CleanupRollbackResult struct {
	Restored   int
	Skipped    int
	Failed     int
	Mapping    map[int64]int64
	Failures   []CleanupRollbackFailure
}

// CleanupRollbackFailure records a single failed restore.
type CleanupRollbackFailure struct {
	OriginalID int64
	EntityType string
	EntityID   int64
	Error      string
}

// ErrCleanupRollbackUnsupportedEntity is returned when an attachment
// entry references an entity kind without an Add* endpoint (e.g. test).
var ErrCleanupRollbackUnsupportedEntity = errors.New("attachment entity type does not support rollback re-upload")

// RestoreCleanupAttachments re-uploads every attachment recorded in the
// snapshot. It prefers the v2 mapping.json layout (which carries
// SHA-256 + Restorable flag) and falls back to the v1 data.json layout
// for snapshots taken before the v2 schema was introduced.
//
// When dryRun is true, no API calls are performed but the result still
// reports how many entries would be processed.
func RestoreCleanupAttachments(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	store *Store,
	snapID string,
	dryRun bool,
) (*CleanupRollbackResult, error) {
	mapping, err := LoadMapping(store, snapID)
	if err == nil && mapping.SchemaVersion >= MappingSchemaVersion {
		return restoreFromMapping(ctx, api, store, snapID, mapping, dryRun)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// A real I/O error reading mapping.json is not the same as
		// "missing"; surface it instead of silently falling back.
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			return nil, fmt.Errorf("load mapping.json: %w", err)
		}
	}
	log.Info("legacy snapshot detected; reference rewrite skipped",
		zap.String("snap_id", snapID))
	return restoreFromLegacyData(ctx, api, store, snapID, dryRun)
}

// restoreFromMapping is the v2 restore path. Each successful re-upload
// updates the corresponding MappingEntry.NewID and the mapping is
// persisted again so a partial restore retains the assignments
// already made.
//
//nolint:gocyclo // Restore loop with branching by Restorable + dryRun + per-entry fallback is more readable kept inline.
func restoreFromMapping(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	store *Store,
	snapID string,
	mapping *Mapping,
	dryRun bool,
) (*CleanupRollbackResult, error) {
	res := &CleanupRollbackResult{Mapping: map[int64]int64{}}
	if len(mapping.Entries) == 0 {
		return res, nil
	}
	dir := filepath.Join(store.SnapDir(snapID), cleanupFilesDir)
	for i := range mapping.Entries {
		entry := &mapping.Entries[i]
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if !entry.Restorable {
			res.Skipped++
			res.Failures = append(res.Failures, CleanupRollbackFailure{
				OriginalID: entry.OriginalID,
				EntityType: entry.EntityType,
				EntityID:   entry.EntityID,
				Error:      entry.NotRestorable,
			})
			continue
		}
		if dryRun {
			res.Restored++
			continue
		}
		legacy := CleanupAttachmentEntry{
			ID:           entry.OriginalID,
			Name:         entry.Name,
			EntityType:   entry.EntityType,
			EntityID:     entry.EntityID,
			ParentPlanID: entry.ParentPlanID,
			EntryID:      entry.EntryID,
			File:         entry.File,
			Compressed:   entry.Compressed,
		}
		newID, err := restoreOneCleanup(ctx, api, dir, legacy)
		if err != nil {
			res.Failed++
			res.Failures = append(res.Failures, CleanupRollbackFailure{
				OriginalID: entry.OriginalID,
				EntityType: entry.EntityType,
				EntityID:   entry.EntityID,
				Error:      err.Error(),
			})
			continue
		}
		entry.NewID = newID
		res.Mapping[entry.OriginalID] = newID
		res.Restored++
	}
	if !dryRun {
		// Persist updated NewIDs even on partial failure so a re-run
		// after the operator fixes a broken parent can pick up where
		// we left off.
		if err := SaveMapping(store, snapID, mapping); err != nil {
			log.Warn("failed to persist updated mapping.json after restore",
				zap.String("snap_id", snapID),
				zap.Error(err))
		}
	}
	return res, nil
}

// restoreFromLegacyData implements the v1 path used when only data.json
// exists (snapshots taken before the v2 schema rolled out).
func restoreFromLegacyData(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	store *Store,
	snapID string,
	dryRun bool,
) (*CleanupRollbackResult, error) {
	var dataFile CleanupAttachmentsData
	if err := store.LoadData(snapID, "data.json", &dataFile); err != nil {
		return nil, fmt.Errorf("load cleanup data.json: %w", err)
	}

	res := &CleanupRollbackResult{Mapping: map[int64]int64{}}
	if len(dataFile.Attachments) == 0 {
		return res, nil
	}

	dir := filepath.Join(store.SnapDir(snapID), cleanupFilesDir)
	for _, entry := range dataFile.Attachments {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if dryRun {
			res.Restored++
			continue
		}

		newID, err := restoreOneCleanup(ctx, api, dir, entry)
		if err != nil {
			if errors.Is(err, ErrCleanupRollbackUnsupportedEntity) {
				res.Skipped++
				res.Failures = append(res.Failures, CleanupRollbackFailure{
					OriginalID: entry.ID,
					EntityType: entry.EntityType,
					EntityID:   entry.EntityID,
					Error:      err.Error(),
				})
				continue
			}
			res.Failed++
			res.Failures = append(res.Failures, CleanupRollbackFailure{
				OriginalID: entry.ID,
				EntityType: entry.EntityType,
				EntityID:   entry.EntityID,
				Error:      err.Error(),
			})
			continue
		}
		res.Mapping[entry.ID] = newID
		res.Restored++
	}
	return res, nil
}

func restoreOneCleanup(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	dir string,
	entry CleanupAttachmentEntry,
) (int64, error) {
	srcPath := filepath.Join(dir, entry.File)

	uploadPath := srcPath
	tmpPath := ""
	if entry.Compressed {
		tmp, err := decompressToTemp(srcPath, entry.Name)
		if err != nil {
			return 0, fmt.Errorf("decompress: %w", err)
		}
		uploadPath = tmp
		tmpPath = tmp
		defer os.Remove(tmpPath)
	}

	switch entry.EntityType {
	case "case":
		resp, err := api.AddAttachmentToCase(ctx, entry.EntityID, uploadPath)
		if err != nil {
			return 0, err
		}
		return resp.AttachmentID, nil
	case "run":
		resp, err := api.AddAttachmentToRun(ctx, entry.EntityID, uploadPath)
		if err != nil {
			return 0, err
		}
		return resp.AttachmentID, nil
	case "plan":
		resp, err := api.AddAttachmentToPlan(ctx, entry.EntityID, uploadPath)
		if err != nil {
			return 0, err
		}
		return resp.AttachmentID, nil
	case "plan_entry":
		resp, err := api.AddAttachmentToPlanEntry(ctx, entry.ParentPlanID, entry.EntryID, uploadPath)
		if err != nil {
			return 0, err
		}
		return resp.AttachmentID, nil
	case "result":
		resp, err := api.AddAttachmentToResult(ctx, entry.EntityID, uploadPath)
		if err != nil {
			return 0, err
		}
		return resp.AttachmentID, nil
	case "test":
		// TestRail API has no add_attachment_to_test endpoint — test-bound
		// attachments live under their result and cannot be re-uploaded
		// directly without recreating the underlying result first.
		return 0, fmt.Errorf("%w: %q", ErrCleanupRollbackUnsupportedEntity, entry.EntityType)
	default:
		return 0, fmt.Errorf("%w: %q", ErrCleanupRollbackUnsupportedEntity, entry.EntityType)
	}
}
