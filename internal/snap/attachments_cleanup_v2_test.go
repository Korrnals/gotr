// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

// TestBackupV2_PersistsMappingWithSHA256 asserts that the v2 backup
// emits mapping.json with one entry per attachment, each carrying the
// inline-computed SHA-256 of the on-disk file.
func TestBackupV2_PersistsMappingWithSHA256(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(11, "case-payload")
	api.seed(22, "result-payload")

	atts := []data.Attachment{
		{ID: 11, Name: "a.txt", CaseID: 5},
		{ID: 22, Name: "b.bin", ResultID: 9},
	}
	snapID := "cleanup-attachments/v2-mapping"
	store := scratchStore(t, snapID)

	saved, _, err := BackupAttachmentsForCleanupV2(context.Background(), api, store, snapID, atts, BackupOptions{Concurrency: 2})
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if saved != 2 {
		t.Fatalf("saved=%d", saved)
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
		// Re-hash the actual file and compare.
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
	// data.json must still exist for legacy back-compat.
	var legacy CleanupAttachmentsData
	if err := store.LoadData(snapID, "data.json", &legacy); err != nil {
		t.Fatalf("load legacy data.json: %v", err)
	}
	if len(legacy.Attachments) != 2 {
		t.Errorf("legacy entries=%d", len(legacy.Attachments))
	}
}

// TestBackupV2_TestBoundNonRestorable verifies that test-bound
// attachments are flagged Restorable=false with the documented reason
// and skipped by the restore phase.
func TestBackupV2_TestBoundNonRestorable(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(77, "trace")
	atts := []data.Attachment{{ID: 77, Name: "trace.txt", TestID: 42}}
	snapID := "cleanup-attachments/v2-testbound"
	store := scratchStore(t, snapID)

	if _, _, err := BackupAttachmentsForCleanupV2(context.Background(), api, store, snapID, atts, BackupOptions{}); err != nil {
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

	// Restore should mark it as Skipped, not failed.
	res, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if res.Skipped != 1 || res.Restored != 0 || res.Failed != 0 {
		t.Errorf("res=%+v", res)
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
	if _, _, err := BackupAttachmentsForCleanupV2(context.Background(), api, store, snapID, atts, BackupOptions{}); err != nil {
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

	// Tamper with a binary; verify must fail.
	first := filepath.Join(store.SnapDir(snapID), idx.Files[0].Path)
	if err := os.WriteFile(first, []byte("tampered"), 0o600); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if err := VerifyIntegrityIndex(store, snapID); err == nil {
		t.Fatalf("expected verify failure after tamper")
	}
}

// TestRestore_LegacyDataJSON_Fallback simulates a v1 snapshot by
// removing mapping.json and asserts the legacy code path still
// restores the entries.
func TestRestore_LegacyDataJSON_Fallback(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(1, "x")
	atts := []data.Attachment{{ID: 1, Name: "x.txt", CaseID: 5}}
	snapID := "cleanup-attachments/legacy-fallback"
	store := scratchStore(t, snapID)
	if _, _, err := BackupAttachmentsForCleanupV2(context.Background(), api, store, snapID, atts, BackupOptions{}); err != nil {
		t.Fatalf("backup: %v", err)
	}
	// Strip mapping.json to mimic a v1 snapshot.
	if err := os.Remove(filepath.Join(store.SnapDir(snapID), "mapping.json")); err != nil {
		t.Fatalf("rm mapping: %v", err)
	}

	res, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if res.Restored != 1 {
		t.Fatalf("legacy fallback failed: %+v", res)
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
	if _, _, err := BackupAttachmentsForCleanupV2(context.Background(), api, store, snapID, atts, BackupOptions{}); err != nil {
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
