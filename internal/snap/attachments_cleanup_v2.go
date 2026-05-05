// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"context"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/snap/refs"
	"go.uber.org/zap"
)

// MappingSchemaVersion is the schema_version emitted in mapping.json
// for v2 attachment-cleanup snapshots.
const MappingSchemaVersion = 2

// Mapping is the on-disk shape of <snap>/mapping.json. It records one
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

// MappingEntry is the per-attachment record inside mapping.json.
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

// notRestorableTestReason is the static reason recorded for test-bound
// attachments. TestRail has no add_attachment_to_test endpoint, so
// these binaries can be downloaded for manual recovery but cannot be
// re-uploaded by the rollback workflow.
const notRestorableTestReason = "test-bound attachment cannot be re-uploaded via TestRail API"

// BackupOptions parameterises BackupAttachmentsForCleanupV2. Zero
// values pick safe defaults: Concurrency=1, no compression.
type BackupOptions struct {
	// Compress enables gzip compression on every stored binary.
	Compress bool
	// Concurrency is the number of parallel downloads. Values <= 0
	// fall back to 1 (sequential) which preserves the v1 behaviour.
	Concurrency int
}

// BackupAttachmentsForCleanupV2 downloads every attachment binary,
// computes its SHA-256 inline (single pass, MultiWriter), and persists
// a versioned mapping.json incrementally. data.json is also written
// for legacy compatibility — old snapshots keep their existing
// rollback path while new ones gain integrity verification and
// reference rewriting.
//
//nolint:gocyclo // Multi-stage backup pipeline (download → hash → write file → atomic mapping commit) is more readable kept as a single orchestrator.
func BackupAttachmentsForCleanupV2(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	store *Store,
	snapID string,
	atts []data.Attachment,
	opts BackupOptions,
) (saved int, totalBytes int64, err error) {
	if len(atts) == 0 {
		return 0, 0, nil
	}

	dir := filepath.Join(store.SnapDir(snapID), cleanupFilesDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, 0, fmt.Errorf("create cleanup files dir: %w", err)
	}

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	mapping := &Mapping{
		SchemaVersion: MappingSchemaVersion,
		SnapID:        snapID,
		GeneratedAt:   time.Now().UTC(),
		Total:         len(atts),
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
		if _, err := store.SaveData(snapID, "mapping.json", snap); err != nil {
			saveErrMu.Lock()
			if saveErr == nil {
				saveErr = err
			}
			saveErrMu.Unlock()
		}
	}

	// Periodic committer so a long-running backup keeps mapping.json
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

	results, _ := concurrent.ParallelMap(ctx, atts, concurrency, func(att data.Attachment, _ int) (MappingEntry, error) {
		if err := ctx.Err(); err != nil {
			return MappingEntry{}, err
		}
		entry, err := backupOneCleanupV2(ctx, api, dir, att, opts.Compress)
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

	// Aggregate results, surfacing the first error (if any). Successful
	// entries already live in mapping.Entries via the worker callback;
	// we recompute totals from there to keep them authoritative.
	for _, r := range results {
		if r.Error != nil {
			persistMapping(true)
			return saved, totalBytes, fmt.Errorf("backup attachment: %w", r.Error)
		}
	}
	persistMapping(true)
	if saveErr != nil {
		return saved, totalBytes, fmt.Errorf("persist mapping.json: %w", saveErr)
	}

	saved = len(mapping.Entries)
	for _, e := range mapping.Entries {
		totalBytes += e.Size
	}

	// Write data.json (legacy v1 format) for backward-compat rollback.
	legacy := CleanupAttachmentsData{Attachments: make([]CleanupAttachmentEntry, 0, len(mapping.Entries))}
	for _, e := range mapping.Entries {
		legacy.Attachments = append(legacy.Attachments, CleanupAttachmentEntry{
			ID:           e.OriginalID,
			Name:         e.Name,
			Size:         e.Size,
			EntityType:   e.EntityType,
			EntityID:     e.EntityID,
			ParentPlanID: e.ParentPlanID,
			EntryID:      e.EntryID,
			File:         e.File,
			Compressed:   e.Compressed,
		})
	}
	sort.Slice(legacy.Attachments, func(i, j int) bool { return legacy.Attachments[i].ID < legacy.Attachments[j].ID })
	if _, err := store.SaveData(snapID, "data.json", legacy); err != nil {
		return saved, totalBytes, fmt.Errorf("save data.json: %w", err)
	}
	return saved, totalBytes, nil
}

// backupOneCleanupV2 downloads a single attachment, hashes it inline
// during streaming via io.MultiWriter, and returns a populated
// MappingEntry. The optional gzip wrapper sits between the hasher and
// the on-disk file so the SHA-256 reflects the post-compression bytes
// (i.e. exactly what is stored under <snap>/files/).
func backupOneCleanupV2(
	ctx context.Context,
	api CleanupAttachmentsAPI,
	dir string,
	att data.Attachment,
	compress bool,
) (*MappingEntry, error) {
	body, err := api.DownloadAttachment(ctx, att.ID)
	if err != nil {
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

// LoadMapping reads <snap>/mapping.json. Returns os.ErrNotExist when
// the snapshot was created before v2 (only data.json is present);
// callers detect this and fall back to legacy restore.
func LoadMapping(store *Store, snapID string) (*Mapping, error) {
	var m Mapping
	if err := store.LoadData(snapID, "mapping.json", &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// SaveMapping persists mapping.json atomically. Used after restore
// updates NewID on each entry.
func SaveMapping(store *Store, snapID string, m *Mapping) error {
	if _, err := store.SaveData(snapID, "mapping.json", m); err != nil {
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
// os.ErrNotExist when the snapshot was created without a reference
// scan (legacy v1, or v2 with --skip-references).
func LoadReferencesSidecar(store *Store, snapID string) ([]refs.EntityRefs, error) {
	var on []refs.EntityRefs
	if err := store.LoadData(snapID, "references.json", &on); err != nil {
		return nil, err
	}
	return on, nil
}
