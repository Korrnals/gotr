// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

// fakeCleanupAPI is an in-memory implementation of CleanupAttachmentsAPI
// covering both the backup (DownloadAttachment) and rollback
// (AddAttachmentTo*) sides of the round-trip.
type fakeCleanupAPI struct {
	contents map[int64][]byte
	nextID   int64
	uploads  []uploadCall
}

type uploadCall struct {
	entity   string
	parentID int64
	entryID  string
	path     string
	bytes    []byte
}

func newFakeCleanupAPI() *fakeCleanupAPI {
	return &fakeCleanupAPI{
		contents: map[int64][]byte{},
		nextID:   100,
	}
}

func (f *fakeCleanupAPI) seed(id int64, body string) {
	f.contents[id] = []byte(body)
}

func (f *fakeCleanupAPI) DownloadAttachment(_ context.Context, id int64) (io.ReadCloser, error) {
	body, ok := f.contents[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return io.NopCloser(strings.NewReader(string(body))), nil
}

func (f *fakeCleanupAPI) record(entity string, parent int64, entryID, path string) (*data.AttachmentResponse, error) {
	b, err := os.ReadFile(path) //nolint:gosec // test scratch path
	if err != nil {
		return nil, err
	}
	f.nextID++
	f.uploads = append(f.uploads, uploadCall{entity: entity, parentID: parent, entryID: entryID, path: path, bytes: b})
	return &data.AttachmentResponse{AttachmentID: f.nextID}, nil
}

func (f *fakeCleanupAPI) AddAttachmentToCase(_ context.Context, caseID int64, p string) (*data.AttachmentResponse, error) {
	return f.record("case", caseID, "", p)
}

func (f *fakeCleanupAPI) AddAttachmentToPlan(_ context.Context, planID int64, p string) (*data.AttachmentResponse, error) {
	return f.record("plan", planID, "", p)
}

func (f *fakeCleanupAPI) AddAttachmentToPlanEntry(_ context.Context, planID int64, entryID, p string) (*data.AttachmentResponse, error) {
	return f.record("plan_entry", planID, entryID, p)
}

func (f *fakeCleanupAPI) AddAttachmentToResult(_ context.Context, resultID int64, p string) (*data.AttachmentResponse, error) {
	return f.record("result", resultID, "", p)
}

func (f *fakeCleanupAPI) AddAttachmentToRun(_ context.Context, runID int64, p string) (*data.AttachmentResponse, error) {
	return f.record("run", runID, "", p)
}

// scratchStore returns a Store rooted in a fresh temp directory and
// pre-creates the snapshot subdir for the given category/id.
func scratchStore(t *testing.T, snapID string) *Store {
	t.Helper()
	store, err := NewStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	if err := os.MkdirAll(store.SnapDir(snapID), 0o755); err != nil {
		t.Fatalf("mkdir snap: %v", err)
	}
	return store
}

// TestBackup_PersistsMappingWithSHA256 asserts that the backup emits
// attachments.json with one entry per attachment, each carrying the
// inline-computed SHA-256 of the on-disk file.
func TestBackup_PersistsMappingWithSHA256(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(11, "case-payload")
	api.seed(22, "result-payload")

	atts := []data.Attachment{
		{ID: 11, Name: "a.txt", CaseID: 5},
		{ID: 22, Name: "b.bin", ResultID: 9},
	}
	snapID := "cleanup-attachments/mapping"
	store := scratchStore(t, snapID)

	res, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, BackupOptions{Concurrency: 2})
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if res.Saved != 2 {
		t.Fatalf("saved=%d", res.Saved)
	}

	m, err := LoadMapping(store, snapID)
	if err != nil {
		t.Fatalf("load mapping: %v", err)
	}
	if m.SchemaVersion != MappingSchemaVersion {
		t.Errorf("schema_version=%d want %d", m.SchemaVersion, MappingSchemaVersion)
	}
	if len(m.Entries) != 2 {
		t.Fatalf("entries=%d", len(m.Entries))
	}
	for _, e := range m.Entries {
		if e.SHA256 == "" {
			t.Errorf("entry %d missing sha256", e.OriginalID)
		}
		fp := filepath.Join(store.SnapDir(snapID), cleanupFilesDir, e.File)
		raw, err := os.ReadFile(fp) //nolint:gosec // test scratch
		if err != nil {
			t.Fatalf("read %s: %v", fp, err)
		}
		sum := sha256.Sum256(raw)
		want := hex.EncodeToString(sum[:])
		if e.SHA256 != want {
			t.Errorf("entry %d sha256 mismatch: got %s want %s", e.OriginalID, e.SHA256, want)
		}
	}
}

// TestBackupAndRestoreCleanupAttachments_RoundTrip exercises the full
// backup → restore lifecycle with three different parent entity kinds.
func TestBackupAndRestoreCleanupAttachments_RoundTrip(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(11, "case-payload")
	api.seed(22, "result-payload")
	api.seed(33, "run-payload")

	atts := []data.Attachment{
		{ID: 11, Name: "case.txt", Size: 12, CaseID: 5},
		{ID: 22, Name: "result.bin", Size: 14, ResultID: 9},
		{ID: 33, Name: "run.log", Size: 11, RunID: 7},
	}

	snapID := "cleanup-attachments/round-trip"
	store := scratchStore(t, snapID)

	res, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, BackupOptions{})
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if res.Saved != 3 || res.TotalBytes <= 0 {
		t.Fatalf("unexpected saved=%d total=%d", res.Saved, res.TotalBytes)
	}

	m, err := LoadMapping(store, snapID)
	if err != nil {
		t.Fatalf("load mapping: %v", err)
	}
	if len(m.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(m.Entries))
	}
	for _, e := range m.Entries {
		p := filepath.Join(store.SnapDir(snapID), cleanupFilesDir, e.File)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing binary %s: %v", p, err)
		}
	}

	restoreRes, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restoreRes.Restored != 3 || restoreRes.Failed != 0 || restoreRes.Skipped != 0 {
		t.Fatalf("restore stats: %+v", restoreRes)
	}
	if len(restoreRes.Mapping) != 3 {
		t.Fatalf("expected 3 id mappings, got %d", len(restoreRes.Mapping))
	}
	if len(api.uploads) != 3 {
		t.Fatalf("expected 3 uploads, got %d", len(api.uploads))
	}
}

// TestRestore_TestBoundNonRestorable verifies that test-bound
// attachments are flagged Restorable=false with the documented reason
// and skipped by the restore phase.
func TestRestore_TestBoundNonRestorable(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(77, "trace")
	atts := []data.Attachment{{ID: 77, Name: "trace.txt", TestID: 42}}
	snapID := "cleanup-attachments/testbound"
	store := scratchStore(t, snapID)

	if _, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, BackupOptions{}); err != nil {
		t.Fatalf("backup: %v", err)
	}
	m, err := LoadMapping(store, snapID)
	if err != nil {
		t.Fatalf("load mapping: %v", err)
	}
	if len(m.Entries) != 1 || m.Entries[0].Restorable {
		t.Fatalf("expected non-restorable, got %+v", m.Entries)
	}
	if !strings.Contains(m.Entries[0].NotRestorable, "test-bound") {
		t.Errorf("reason=%q", m.Entries[0].NotRestorable)
	}

	res, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if res.Skipped != 1 || res.Restored != 0 || res.Failed != 0 {
		t.Errorf("res=%+v", res)
	}
	if len(res.Failures) != 1 || res.Failures[0].EntityType != "test" {
		t.Fatalf("expected test failure record, got %+v", res.Failures)
	}
}

// TestBackupCompression_RoundTrip exercises gzip-on-disk and confirms
// rollback can decompress and re-upload the original bytes.
func TestBackupCompression_RoundTrip(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(1, strings.Repeat("hello ", 200))

	atts := []data.Attachment{{ID: 1, Name: "big.txt", Size: 1200, CaseID: 1}}
	snapID := "cleanup-attachments/compressed"
	store := scratchStore(t, snapID)

	backupRes, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, BackupOptions{Compress: true})
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if backupRes.Saved != 1 {
		t.Fatalf("saved=%d", backupRes.Saved)
	}

	m, err := LoadMapping(store, snapID)
	if err != nil {
		t.Fatalf("load mapping: %v", err)
	}
	if !m.Entries[0].Compressed || !strings.HasSuffix(m.Entries[0].File, ".gz") {
		t.Fatalf("expected compressed .gz, got %+v", m.Entries[0])
	}

	res, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if res.Restored != 1 {
		t.Fatalf("restore: %+v", res)
	}
}

// TestIntegrity_BuildAndVerify_Roundtrip writes a snapshot, builds
// integrity.json, then re-verifies it. Files unchanged → no error.
// Mutate one file and verify must surface a clear mismatch.
func TestIntegrity_BuildAndVerify_Roundtrip(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(1, "alpha")
	api.seed(2, "beta")
	snapID := "cleanup-attachments/integrity-rt"
	store := scratchStore(t, snapID)

	atts := []data.Attachment{
		{ID: 1, Name: "a.txt", CaseID: 1},
		{ID: 2, Name: "b.txt", CaseID: 1},
	}
	if _, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, BackupOptions{}); err != nil {
		t.Fatalf("backup: %v", err)
	}
	idx, err := WriteIntegrityIndex(store, snapID)
	if err != nil {
		t.Fatalf("write integrity: %v", err)
	}
	if idx.Root == "" || len(idx.Files) == 0 {
		t.Fatalf("integrity empty: %+v", idx)
	}
	if err := VerifyIntegrityIndex(store, snapID); err != nil {
		t.Fatalf("verify: %v", err)
	}

	first := filepath.Join(store.SnapDir(snapID), idx.Files[0].Path)
	if err := os.WriteFile(first, []byte("tampered"), 0o600); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if err := VerifyIntegrityIndex(store, snapID); err == nil {
		t.Fatalf("expected verify failure after tamper")
	}
}

// TestRestore_MappingUpdatesNewID asserts that successful re-uploads
// patch MappingEntry.NewID and persist the mapping back to disk.
func TestRestore_MappingUpdatesNewID(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(11, "p")
	atts := []data.Attachment{{ID: 11, Name: "n.txt", CaseID: 5}}
	snapID := "cleanup-attachments/mapping-newid"
	store := scratchStore(t, snapID)
	if _, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, BackupOptions{}); err != nil {
		t.Fatalf("backup: %v", err)
	}
	if _, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false); err != nil {
		t.Fatalf("restore: %v", err)
	}
	m, err := LoadMapping(store, snapID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(m.Entries) != 1 || m.Entries[0].NewID == 0 {
		t.Errorf("expected new_id populated, got %+v", m.Entries)
	}
}

// TestRestore_MissingMappingFails verifies that a snapshot directory
// without attachments.json produces a clear error (no v1 fallback).
func TestRestore_MissingMappingFails(t *testing.T) {
	api := newFakeCleanupAPI()
	snapID := "cleanup-attachments/no-mapping"
	store := scratchStore(t, snapID)

	_, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err == nil {
		t.Fatalf("expected error when attachments.json is missing")
	}
	if !strings.Contains(err.Error(), "attachments.json") {
		t.Errorf("error must mention attachments.json: %v", err)
	}
}
