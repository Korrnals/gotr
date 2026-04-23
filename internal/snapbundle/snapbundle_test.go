package snapbundle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/bundle"
	"github.com/Korrnals/gotr/internal/snap"
)

// newStoreWithSnap creates an isolated snap store at t.TempDir() and seeds a
// fake snapshot at "sync/<name>" with meta.json + data.json that includes
// assignee_email so redaction can be verified.
func newStoreWithSnap(t *testing.T) (*snap.Store, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := snap.NewStoreAt(dir)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	snapID := "sync/20260101T000000_full_p1_to_p2"
	meta := map[string]any{
		"id":             snapID,
		"label":          "unit",
		"category":       "sync",
		"operation":      "sync_full",
		"entity_type":    "run",
		"entity_ids":     []int{1},
		"project_id":     1,
		"dst_project_id": 2,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"status":         "success",
		"data_file":      "data.json",
		"assignee_email": "alice@example.com",
	}
	data := map[string]any{
		"runs": []map[string]any{
			{"id": 1, "assignee": "alice", "assignee_email": "alice@example.com"},
		},
	}
	snapDir := filepath.Join(dir, "sync", "20260101T000000_full_p1_to_p2")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatalf("mkdir snap: %v", err)
	}
	writeJSON(t, filepath.Join(snapDir, "meta.json"), meta)
	writeJSON(t, filepath.Join(snapDir, "data.json"), data)
	return store, snapID
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestExportOne_And_ImportRoundtrip(t *testing.T) {
	store, snapID := newStoreWithSnap(t)
	dest := filepath.Join(t.TempDir(), "snap.tar.gz")

	res, err := ExportOne(store, snapID, dest, ExportOptions{GotrVersion: "test"})
	if err != nil {
		t.Fatalf("ExportOne: %v", err)
	}
	if res.ArchivePath != dest {
		t.Errorf("archive path = %s, want %s", res.ArchivePath, dest)
	}
	if len(res.Files) < 2 {
		t.Errorf("expected ≥2 files in manifest, got %d", len(res.Files))
	}

	// Import into a fresh store.
	freshDir := t.TempDir()
	fresh, err := snap.NewStoreAt(freshDir)
	if err != nil {
		t.Fatalf("fresh store: %v", err)
	}
	imp, err := Import(fresh, dest, ImportOptions{})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if imp.SnapID != snapID {
		t.Errorf("imported snap_id = %q, want %q", imp.SnapID, snapID)
	}
	if !fresh.Exists(snapID) {
		t.Errorf("expected snapshot %s in fresh store", snapID)
	}
}

func TestExportOne_Redact_StripsAssigneeEmail(t *testing.T) {
	store, snapID := newStoreWithSnap(t)
	dest := filepath.Join(t.TempDir(), "snap.tar.gz")

	res, err := ExportOne(store, snapID, dest, ExportOptions{GotrVersion: "test", Redact: true})
	if err != nil {
		t.Fatalf("ExportOne redact: %v", err)
	}
	found := false
	for _, f := range res.Redacted {
		if strings.Contains(f, "assignee_email") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected redacted list to include assignee_email, got %v", res.Redacted)
	}

	// Extract and confirm redacted value is in meta.json.
	tmp := t.TempDir()
	if _, err := bundle.ReadTarGz(dest, tmp); err != nil {
		t.Fatalf("ReadTarGz: %v", err)
	}
	metaPath := filepath.Join(tmp, "snaps", filepath.FromSlash(snapID), "meta.json")
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	if !strings.Contains(string(raw), "[redacted]") {
		t.Errorf("expected [redacted] marker in meta.json, got: %s", string(raw))
	}
	if strings.Contains(string(raw), "alice@example.com") {
		t.Errorf("assignee_email leaked in redacted export: %s", string(raw))
	}
}

func TestImport_RenameID(t *testing.T) {
	store, snapID := newStoreWithSnap(t)
	dest := filepath.Join(t.TempDir(), "snap.tar.gz")
	if _, err := ExportOne(store, snapID, dest, ExportOptions{GotrVersion: "test"}); err != nil {
		t.Fatalf("ExportOne: %v", err)
	}
	fresh, _ := snap.NewStoreAt(t.TempDir())
	newID := "sync/renamed_id"
	imp, err := Import(fresh, dest, ImportOptions{RenameID: newID})
	if err != nil {
		t.Fatalf("Import rename: %v", err)
	}
	if imp.SnapID != newID {
		t.Errorf("imported id = %q, want %q", imp.SnapID, newID)
	}
	if !fresh.Exists(newID) {
		t.Errorf("renamed snap missing on disk")
	}
	// meta.json must reflect the new id.
	metaRaw, err := os.ReadFile(filepath.Join(fresh.SnapDir(newID), "meta.json"))
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	var meta map[string]any
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		t.Fatalf("parse meta: %v", err)
	}
	if got := meta["id"]; got != newID {
		t.Errorf("meta.id = %v, want %s", got, newID)
	}
}

func TestImport_RefusesExisting_WithoutOverwrite(t *testing.T) {
	store, snapID := newStoreWithSnap(t)
	dest := filepath.Join(t.TempDir(), "snap.tar.gz")
	if _, err := ExportOne(store, snapID, dest, ExportOptions{GotrVersion: "test"}); err != nil {
		t.Fatalf("ExportOne: %v", err)
	}
	if _, err := Import(store, dest, ImportOptions{}); err == nil {
		t.Fatalf("expected Import to fail against existing snapshot, got nil error")
	}
}

func TestImport_OverwriteBacksUpToTrash(t *testing.T) {
	store, snapID := newStoreWithSnap(t)
	dest := filepath.Join(t.TempDir(), "snap.tar.gz")
	if _, err := ExportOne(store, snapID, dest, ExportOptions{GotrVersion: "test"}); err != nil {
		t.Fatalf("ExportOne: %v", err)
	}
	if _, err := Import(store, dest, ImportOptions{Overwrite: true}); err != nil {
		t.Fatalf("Import overwrite: %v", err)
	}
	trash := filepath.Join(store.BaseDir(), ".trash")
	ents, err := os.ReadDir(trash)
	if err != nil {
		t.Fatalf("read trash: %v", err)
	}
	if len(ents) == 0 {
		t.Errorf("expected backup in %s, found none", trash)
	}
}

func TestImport_DryRun_NoMutation(t *testing.T) {
	store, snapID := newStoreWithSnap(t)
	dest := filepath.Join(t.TempDir(), "snap.tar.gz")
	if _, err := ExportOne(store, snapID, dest, ExportOptions{GotrVersion: "test"}); err != nil {
		t.Fatalf("ExportOne: %v", err)
	}
	fresh, _ := snap.NewStoreAt(t.TempDir())
	res, err := Import(fresh, dest, ImportOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.SnapID != snapID {
		t.Errorf("dry-run snap_id = %s, want %s", res.SnapID, snapID)
	}
	if fresh.Exists(snapID) {
		t.Errorf("dry-run must not write snapshot to store")
	}
}

func TestImport_SchemaVersionMismatch_Rejected(t *testing.T) {
	// Build a handcrafted bad bundle: manifest schema_version = 999.
	tmp := t.TempDir()
	m := bundle.Manifest{SchemaVersion: 999, Kind: bundle.KindSnap, SnapID: "x/y"}
	mj, _ := m.MarshalJSON()
	archive := filepath.Join(tmp, "bad.tar.gz")
	entries := []bundle.Entry{
		{ArchivePath: bundle.ManifestName, Content: mj},
		{ArchivePath: bundle.ChecksumsName, Content: []byte("")},
	}
	if err := bundle.WriteTarGz(archive, entries); err != nil {
		t.Fatalf("WriteTarGz: %v", err)
	}
	store, _ := snap.NewStoreAt(t.TempDir())
	if _, err := Import(store, archive, ImportOptions{}); err == nil {
		t.Fatalf("expected schema version error, got nil")
	}
}
