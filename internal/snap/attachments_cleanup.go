// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/snap/refs"
	"go.uber.org/zap"
)

// errGhostAttachment is returned by backupOneCleanup when the
// attachment disappeared between listing and download (404/400 race).
var errGhostAttachment = errors.New("attachment no longer exists on server")

// EntityTypeAttachments is the snap.Meta.EntityType used for bulk
// attachment-cleanup operations.
const EntityTypeAttachments = "attachments"

// MappingSchemaVersion is the schema_version emitted in attachments.json.
// Bump this constant whenever the on-disk shape changes in a way that
// requires migration logic on the rollback side.
const MappingSchemaVersion = 2

// cleanupFilesDir is the subdirectory under <snap>/ that stores
// attachment binaries (one file per backed-up attachment).
const cleanupFilesDir = "files"

// notRestorableTestReason is the static reason recorded for test-bound
// attachments. TestRail has no add_attachment_to_test endpoint, so
// these binaries can be downloaded for manual recovery but cannot be
// re-uploaded by the rollback workflow.
const notRestorableTestReason = "test-bound attachment cannot be re-uploaded via TestRail API"

// CleanupAttachmentsAPI is the API surface required to back up and
// restore bulk-deleted attachments. It is satisfied by *client.HTTPClient.
type CleanupAttachmentsAPI interface {
	DownloadAttachment(ctx context.Context, attachmentID int64) (io.ReadCloser, error)
	AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlan(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToPlanEntry(ctx context.Context, planID int64, entryID, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToResult(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error)
	AddAttachmentToRun(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error)
}

// Mapping is the on-disk shape of <snap>/attachments.json. It records one
// MappingEntry per attachment with cryptographic integrity (SHA-256),
// the original parent binding, and a slot for the new ID assigned
// during restore.
type Mapping struct {
	SchemaVersion int            `json:"schema_version"`
	SnapID        string         `json:"snap_id"`
	GeneratedAt   time.Time      `json:"generated_at"`
	Total         int            `json:"total"`
	Entries       []MappingEntry `json:"entries"`
}

// MappingEntry is the per-attachment record inside attachments.json.
type MappingEntry struct {
	OriginalID    int64  `json:"original_id"`
	NewID         int64  `json:"new_id,omitempty"`
	SHA256        string `json:"sha256"`
	Size          int64  `json:"size"`
	Name          string `json:"name"`
	EntityType    string `json:"entity_type"`
	EntityID      int64  `json:"entity_id"`
	ParentPlanID  int64  `json:"parent_plan_id,omitempty"`
	EntryID       string `json:"entry_id,omitempty"`
	File          string `json:"file"`
	Compressed    bool   `json:"compressed"`
	Restorable    bool   `json:"restorable"`
	NotRestorable string `json:"not_restorable_reason,omitempty"`
}

// BackupOptions parameterises BackupAttachmentsForCleanup. Zero
// values pick safe defaults: Concurrency=1, no compression.
type BackupOptions struct {
	// Compress enables gzip compression on every stored binary.
	Compress bool
	// Concurrency is the number of parallel downloads. Values <= 0
	// fall back to 1 (sequential).
	Concurrency int
}

// BackupResult summarizes a BackupAttachmentsForCleanup call.
type BackupResult struct {
	Saved        int
	Skipped      int // ghost attachments (404/400 race) skipped without error
	TotalBytes   int64
	GhostIDs     []int64
}

// CleanupRollbackResult holds the per-entry rollback outcome plus an
// old-id → new-id mapping populated when re-upload succeeded.
type CleanupRollbackResult struct {
	Restored int
	Skipped  int
	Failed   int
	Mapping  map[int64]int64
	Failures []CleanupRollbackFailure
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

// BackupAttachmentsForCleanup downloads every attachment binary,
// computes its SHA-256 inline (single pass via io.MultiWriter), and
// persists a versioned attachments.json incrementally. Returns the number
// of saved attachments and the total bytes written.
//
// Resume behavior: if <snap>/attachments.json already exists, its
// entries are loaded and any attachment whose ID is already present
// is skipped (no re-download). This makes interrupted cleanups
// resumable without re-transferring already-backed-up data.
//
// Ghost tolerance: if an attachment returns 400/404 during download
// (race between listing and backup), it is logged as a warning and
// counted in BackupResult.Skipped instead of aborting the run.
//
//nolint:gocyclo // Multi-stage backup pipeline (download → hash → write file → atomic mapping commit) is more readable kept as a single orchestrator.
func BackupAttachmentsForCleanup(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	store *Store,
	snapID string,
	atts []data.Attachment,
	opts BackupOptions,
) (BackupResult, error) {
	res := BackupResult{}
	if len(atts) == 0 {
		return res, nil
	}

	dir := filepath.Join(store.SnapDir(snapID), cleanupFilesDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return res, fmt.Errorf("create cleanup files dir: %w", err)
	}

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	// --- Resume: load existing mapping if present ---
	mapping := &Mapping{
		SchemaVersion: MappingSchemaVersion,
		SnapID:        snapID,
		GeneratedAt:   time.Now().UTC(),
		Total:         len(atts),
	}
	existingIDs := map[int64]struct{}{}
	if existing, err := LoadMapping(store, snapID); err == nil && existing != nil {
		mapping.Entries = append([]MappingEntry(nil), existing.Entries...)
		for _, e := range existing.Entries {
			existingIDs[e.OriginalID] = struct{}{}
		}
		res.Saved = len(existing.Entries)
		for _, e := range existing.Entries {
			res.TotalBytes += e.Size
		}
	}

	// Filter out already-backed-up attachments.
	var toDownload []data.Attachment
	for _, att := range atts {
		if _, ok := existingIDs[att.ID]; !ok {
			toDownload = append(toDownload, att)
		}
	}
	if len(toDownload) == 0 {
		return res, nil
	}

	var mappingMu sync.Mutex
	var saveErrMu sync.Mutex
	var saveErr error
	dirty := 0

	persistMapping := func(force bool) {
		mappingMu.Lock()
		defer mappingMu.Unlock()
		if !force && dirty < 16 {
			return
		}
		dirty = 0
		// Snapshot a stable copy: sort by original_id for determinism.
		entries := append([]MappingEntry(nil), mapping.Entries...)
		sort.Slice(entries, func(i, j int) bool { return entries[i].OriginalID < entries[j].OriginalID })
		snap := *mapping
		snap.Entries = entries
		if _, err := store.SaveData(snapID, "attachments.json", snap); err != nil {
			saveErrMu.Lock()
			if saveErr == nil {
				saveErr = err
			}
			saveErrMu.Unlock()
		}
	}

	// Periodic committer so a long-running backup keeps attachments.json
	// fresh even when individual download slots are slow.
	flushCtx, cancelFlush := context.WithCancel(ctx)
	flushDone := make(chan struct{})
	go func() {
		defer close(flushDone)
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-flushCtx.Done():
				return
			case <-t.C:
				persistMapping(true)
			}
		}
	}()

	results, _ := concurrent.ParallelMap(ctx, toDownload, concurrency, func(att data.Attachment, _ int) (MappingEntry, error) {
		if err := ctx.Err(); err != nil {
			return MappingEntry{}, err
		}
		entry, err := backupOneCleanup(ctx, api, dir, att, opts.Compress)
		if err != nil {
			return MappingEntry{}, err
		}
		mappingMu.Lock()
		mapping.Entries = append(mapping.Entries, *entry)
		dirty++
		shouldFlush := dirty >= 16
		mappingMu.Unlock()
		if shouldFlush {
			persistMapping(true)
		}
		if !entry.Restorable && entry.EntityType == "test" {
			log.Warn("attachment is test-bound and non-restorable on rollback",
				zap.Int64("attachment_id", entry.OriginalID),
				zap.String("name", entry.Name),
				zap.Int64("test_id", entry.EntityID))
		}
		return *entry, nil
	})

	cancelFlush()
	<-flushDone

	for _, r := range results {
		if r.Error != nil {
			if errors.Is(r.Error, errGhostAttachment) {
				res.Skipped++
				// Extract the attachment ID from the error message for logging.
				var ghostID int64
				_, _ = fmt.Sscanf(r.Error.Error(), "ghost attachment %d", &ghostID)
				if ghostID == 0 {
					// Fallback: try to parse from the wrapped error text.
					parts := strings.Split(r.Error.Error(), " ")
					for _, p := range parts {
						if id, err := fmt.Sscanf(p, "%d", &ghostID); err == nil && id == 1 && ghostID > 0 {
							break
						}
					}
				}
				if ghostID > 0 {
					res.GhostIDs = append(res.GhostIDs, ghostID)
				}
				log.Warn("attachment disappeared during backup (race); skipping",
					zap.Int64("attachment_id", ghostID),
					zap.Error(r.Error))
				continue
			}
			persistMapping(true)
			return res, fmt.Errorf("backup attachment: %w", r.Error)
		}
		res.Saved++
		res.TotalBytes += r.Data.Size
	}
	persistMapping(true)
	if saveErr != nil {
		return res, fmt.Errorf("persist attachments.json: %w", saveErr)
	}

	return res, nil
}

// backupOneCleanup downloads a single attachment, hashes it inline
// during streaming via io.MultiWriter, and returns a populated
// MappingEntry. The optional gzip wrapper sits between the hasher and
// the on-disk file so the SHA-256 reflects the post-compression bytes
// (i.e. exactly what is stored under <snap>/files/).
func backupOneCleanup(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	dir string,
	att data.Attachment,
	compress bool,
) (*MappingEntry, error) {
	body, err := api.DownloadAttachment(ctx, att.ID)
	if err != nil {
		if client.IsAttachmentNotFound(err) {
			return nil, fmt.Errorf("ghost attachment %d: %w", att.ID, errGhostAttachment)
		}
		return nil, fmt.Errorf("download %d: %w", att.ID, err)
	}
	defer body.Close()

	name := sanitizeAttachName(att.Name, att.ID)
	filename := name
	if compress {
		filename += ".gz"
	}
	outPath := filepath.Join(dir, filename)

	hasher := sha256.New()
	written, closeErr, copyErr := streamToFileWithHash(outPath, body, hasher, compress)
	if copyErr != nil {
		_ = os.Remove(outPath)
		return nil, fmt.Errorf("write %d: %w", att.ID, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(outPath)
		return nil, fmt.Errorf("close %d: %w", att.ID, closeErr)
	}

	entityType := att.InferredEntityType()
	entry := &MappingEntry{
		OriginalID: att.ID,
		Name:       att.Name,
		Size:       written,
		SHA256:     hex.EncodeToString(hasher.Sum(nil)),
		File:       filename,
		Compressed: compress,
		EntityType: entityType,
		Restorable: true,
	}
	switch entityType {
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
		entry.Restorable = false
		entry.NotRestorable = notRestorableTestReason
	}
	return entry, nil
}

// streamToFileWithHash writes body to outPath while feeding the same
// bytes into hasher via io.MultiWriter. When compress is true the
// gzip writer is layered between the hasher and the file so the hash
// matches the on-disk (compressed) bytes.
func streamToFileWithHash(outPath string, body io.Reader, hasher io.Writer, compress bool) (written int64, closeErr, copyErr error) {
	f, err := os.Create(outPath) //nolint:gosec // outPath rooted under the cleanup-snap dir.
	if err != nil {
		return 0, nil, err
	}
	mw := io.MultiWriter(f, hasher)
	if compress {
		gz := gzip.NewWriter(mw)
		written, copyErr = io.Copy(gz, body)
		if cerr := gz.Close(); closeErr == nil {
			closeErr = cerr
		}
	} else {
		written, copyErr = io.Copy(mw, body)
	}
	if cerr := f.Close(); closeErr == nil {
		closeErr = cerr
	}
	return written, closeErr, copyErr
}

// LoadMapping reads <snap>/attachments.json.
func LoadMapping(store *Store, snapID string) (*Mapping, error) {
	var m Mapping
	if err := store.LoadData(snapID, "attachments.json", &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// SaveMapping persists attachments.json atomically. Used after restore
// updates NewID on each entry.
func SaveMapping(store *Store, snapID string, m *Mapping) error {
	if _, err := store.SaveData(snapID, "attachments.json", m); err != nil {
		return err
	}
	return nil
}

// WriteReferencesSidecar persists <snap>/references.json containing
// the per-entity reference index. It tolerates an empty input by
// writing an empty array so the absence of the file unambiguously
// means "scan was skipped" rather than "scan found nothing".
func WriteReferencesSidecar(store *Store, snapID string, entries []refs.EntityRefs) error {
	if entries == nil {
		entries = []refs.EntityRefs{}
	}
	if _, err := store.SaveData(snapID, "references.json", entries); err != nil {
		return fmt.Errorf("save references.json: %w", err)
	}
	return nil
}

// LoadReferencesSidecar reads <snap>/references.json. Returns
// os.ErrNotExist when the snapshot was created with --skip-references.
func LoadReferencesSidecar(store *Store, snapID string) ([]refs.EntityRefs, error) {
	var on []refs.EntityRefs
	if err := store.LoadData(snapID, "references.json", &on); err != nil {
		return nil, err
	}
	return on, nil
}

// RestoreCleanupAttachments re-uploads every attachment recorded in
// the snapshot mapping. Successful re-uploads patch MappingEntry.NewID
// and the mapping is persisted again so a partial restore retains the
// assignments already made. When dryRun is true, no API calls are
// performed but the result still reports how many entries would be
// processed.
func RestoreCleanupAttachments(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	store *Store,
	snapID string,
	dryRun bool,
) (*CleanupRollbackResult, error) {
	mapping, err := LoadMapping(store, snapID)
	if err != nil {
		return nil, fmt.Errorf("load attachments.json: %w", err)
	}
	if mapping.SchemaVersion < MappingSchemaVersion {
		return nil, fmt.Errorf("attachments.json schema_version=%d unsupported (want >= %d)",
			mapping.SchemaVersion, MappingSchemaVersion)
	}

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
		newID, err := restoreOneCleanup(ctx, api, dir, entry)
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
			log.Warn("failed to persist updated attachments.json after restore",
				zap.String("snap_id", snapID),
				zap.Error(err))
		}
	}
	return res, nil
}

// restoreOneCleanup re-uploads a single mapping entry. Compressed
// binaries are decompressed to a temp file before upload.
func restoreOneCleanup(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	dir string,
	entry *MappingEntry,
) (int64, error) {
	srcPath := filepath.Join(dir, entry.File)

	uploadPath := srcPath
	if entry.Compressed {
		tmp, err := decompressToTemp(srcPath, entry.Name)
		if err != nil {
			return 0, fmt.Errorf("decompress: %w", err)
		}
		uploadPath = tmp
		defer os.Remove(tmp)
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
