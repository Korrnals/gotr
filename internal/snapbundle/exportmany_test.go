package snapbundle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/bundle"
	"github.com/Korrnals/gotr/internal/snap"
)

// seedSnap creates a minimal valid snapshot at <storeDir>/<id> with
// meta.json + data.json. It returns the snap id verbatim.
func seedSnap(t *testing.T, storeDir, id, label string) string {
	t.Helper()
	meta := map[string]any{
		"id":             id,
		"label":          label,
		"category":       "sync",
		"operation":      "sync_full",
		"entity_type":    "run",
		"project_id":     1,
		"dst_project_id": 2,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"status":         "success",
		"data_file":      "data.json",
	}
	data := map[string]any{"runs": []map[string]any{{"id": 1, "label": label}}}
	dir := filepath.Join(storeDir, filepath.FromSlash(id))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	writeJSON(t, filepath.Join(dir, "meta.json"), meta)
	writeJSON(t, filepath.Join(dir, "data.json"), data)
	return id
}

func TestExportMany_RoundTrip(t *testing.T) {
	srcDir := t.TempDir()
	src, err := snap.NewStoreAt(srcDir)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	id1 := seedSnap(t, srcDir, "sync/20260101T000000_full_p1_to_p2", "alpha")
	id2 := seedSnap(t, srcDir, "sync/20260102T000000_full_p3_to_p4", "beta")
	id3 := seedSnap(t, srcDir, "sync/20260103T000000_full_p5_to_p6", "gamma")

	dest := filepath.Join(t.TempDir(), "bundle.tar.gz")
	res, err := ExportMany(src, []string{id1, id2, id3}, dest, ExportOptions{GotrVersion: "test"})
	if err != nil {
		t.Fatalf("ExportMany: %v", err)
	}
	if got := len(res.SnapIDs); got != 3 {
		t.Fatalf("SnapIDs len = %d, want 3", got)
	}

	// Verify manifest kind via peek.
	m, _, err := peekManifest(dest)
	if err != nil {
		t.Fatalf("peekManifest: %v", err)
	}
	if m.Kind != bundle.KindMigrationBundle {
		t.Errorf("kind = %q, want %q", m.Kind, bundle.KindMigrationBundle)
	}
	if len(m.SnapIDs) != 3 {
		t.Errorf("manifest SnapIDs len = %d, want 3", len(m.SnapIDs))
	}

	// Import into fresh store.
	dstDir := t.TempDir()
	dst, err := snap.NewStoreAt(dstDir)
	if err != nil {
		t.Fatalf("NewStoreAt dst: %v", err)
	}
	imp, err := Import(dst, dest, ImportOptions{})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(imp.SnapIDs) != 3 {
		t.Errorf("imp.SnapIDs len = %d, want 3", len(imp.SnapIDs))
	}
	for _, id := range []string{id1, id2, id3} {
		if !dst.Exists(id) {
			t.Errorf("missing %s in dst store", id)
		}
		// meta.json should be readable & match id.
		raw, err := os.ReadFile(filepath.Join(dst.SnapDir(id), "meta.json"))
		if err != nil {
			t.Errorf("read meta %s: %v", id, err)
			continue
		}
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Errorf("parse meta %s: %v", id, err)
			continue
		}
		if got["id"] != id {
			t.Errorf("meta.id = %v, want %s", got["id"], id)
		}
	}
}

func TestExportMany_RejectsRenameID(t *testing.T) {
	srcDir := t.TempDir()
	src, _ := snap.NewStoreAt(srcDir)
	id := seedSnap(t, srcDir, "sync/20260101T000000_full_p1_to_p2", "x")
	dest := filepath.Join(t.TempDir(), "b.tar.gz")
	if _, err := ExportMany(src, []string{id}, dest, ExportOptions{GotrVersion: "t"}); err != nil {
		t.Fatalf("ExportMany: %v", err)
	}
	dstDir := t.TempDir()
	dst, _ := snap.NewStoreAt(dstDir)
	if _, err := Import(dst, dest, ImportOptions{RenameID: "sync/other"}); err == nil {
		t.Fatalf("expected error on --rename-id with multi-snap bundle, got nil")
	}
}

func TestExportFull_PicksAllManifestEntries(t *testing.T) {
	srcDir := t.TempDir()
	src, _ := snap.NewStoreAt(srcDir)
	id1 := seedSnap(t, srcDir, "sync/20260101T000000_full_p1_to_p2", "a")
	id2 := seedSnap(t, srcDir, "sync/20260102T000000_full_p3_to_p4", "b")
	// Register both in the store manifest so ExportFull picks them up.
	sm, err := snap.LoadManifest(src)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	for _, id := range []string{id1, id2} {
		meta, err := src.LoadMeta(id)
		if err != nil {
			t.Fatalf("LoadMeta %s: %v", id, err)
		}
		if err := sm.Add(meta); err != nil {
			t.Fatalf("manifest.Add: %v", err)
		}
	}

	dest := filepath.Join(t.TempDir(), "full.tar.gz")
	res, err := ExportFull(src, dest, ExportOptions{GotrVersion: "t"})
	if err != nil {
		t.Fatalf("ExportFull: %v", err)
	}
	if len(res.SnapIDs) != 2 {
		t.Errorf("SnapIDs len = %d, want 2", len(res.SnapIDs))
	}
}
