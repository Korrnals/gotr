package snap

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/models/data"
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

// BackupAttachmentsForCleanup downloads each attachment binary into the
// snapshot directory and persists the metadata index to data.json. The
// caller is responsible for creating the snapshot Meta (with
// EntityType=EntityTypeAttachments and Operation=OpDelete) and for
// writing the manifest entry afterwards.
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
	if len(atts) == 0 {
		return 0, 0, nil
	}

	dir := filepath.Join(store.SnapDir(snapID), cleanupFilesDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, 0, fmt.Errorf("create cleanup files dir: %w", err)
	}

	entries := make([]CleanupAttachmentEntry, 0, len(atts))
	for _, att := range atts {
		if err := ctx.Err(); err != nil {
			return saved, totalBytes, err
		}
		entry, written, err := backupOneCleanup(ctx, api, dir, att, compress)
		if err != nil {
			return saved, totalBytes, fmt.Errorf("backup attachment %d (%s): %w", att.ID, att.Name, err)
		}
		entries = append(entries, *entry)
		saved++
		totalBytes += written
	}

	if _, err := store.SaveData(snapID, "data.json", CleanupAttachmentsData{Attachments: entries}); err != nil {
		return saved, totalBytes, fmt.Errorf("save cleanup data.json: %w", err)
	}
	return saved, totalBytes, nil
}

func backupOneCleanup(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	dir string,
	att data.Attachment,
	compress bool,
) (*CleanupAttachmentEntry, int64, error) {
	body, err := api.DownloadAttachment(ctx, att.ID)
	if err != nil {
		return nil, 0, err
	}
	defer body.Close()

	name := sanitizeAttachName(att.Name, att.ID)
	filename := name
	if compress {
		filename += ".gz"
	}

	outPath := filepath.Join(dir, filename)
	f, err := os.Create(outPath)
	if err != nil {
		return nil, 0, fmt.Errorf("create file: %w", err)
	}

	var w io.Writer = f
	var gw *gzip.Writer
	if compress {
		gw = gzip.NewWriter(f)
		w = gw
	}

	written, copyErr := io.Copy(w, body)
	var closeErr error
	if gw != nil {
		closeErr = gw.Close()
	}
	if cerr := f.Close(); closeErr == nil {
		closeErr = cerr
	}
	if copyErr != nil {
		_ = os.Remove(outPath)
		return nil, 0, fmt.Errorf("write: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(outPath)
		return nil, 0, fmt.Errorf("close: %w", closeErr)
	}

	info, statErr := os.Stat(outPath)
	if statErr != nil {
		return nil, 0, fmt.Errorf("stat: %w", statErr)
	}

	entry := &CleanupAttachmentEntry{
		ID:          att.ID,
		Name:        att.Name,
		Size:        att.Size,
		ContentType: att.ContentType,
		CreatedOn:   att.CreatedOn,
		EntityType:  att.InferredEntityType(),
		File:        filename,
		Compressed:  compress,
	}
	switch entry.EntityType {
	case "case":
		entry.EntityID = att.CaseID
	case "run":
		entry.EntityID = att.RunID
	case "plan":
		entry.EntityID = att.PlanID
	case "plan_entry":
		entry.ParentPlanID = att.PlanID
		entry.EntryID = att.EntryID
	case "result":
		entry.EntityID = att.ResultID
	case "test":
		entry.EntityID = att.TestID
	}
	_ = written // size is taken from final file (post-compression)
	return entry, info.Size(), nil
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
// snapshot's data.json to its original parent entity. When dryRun is
// true, no API calls are performed but the result still reports how
// many entries would be processed.
func RestoreCleanupAttachments(
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
