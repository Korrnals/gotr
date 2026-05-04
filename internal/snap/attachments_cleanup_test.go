// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"context"
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

	snapID := "cleanup-attachments/test-snap-1"
	store := scratchStore(t, snapID)

	saved, total, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, false)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if saved != 3 || total <= 0 {
		t.Fatalf("unexpected saved=%d total=%d", saved, total)
	}

	// data.json must be present and parseable.
	var on CleanupAttachmentsData
	if err := store.LoadData(snapID, "data.json", &on); err != nil {
		t.Fatalf("load data.json: %v", err)
	}
	if len(on.Attachments) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(on.Attachments))
	}
	// Files must exist on disk.
	for _, e := range on.Attachments {
		p := filepath.Join(store.SnapDir(snapID), cleanupFilesDir, e.File)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing binary %s: %v", p, err)
		}
	}

	// Restore re-uploads each entry.
	res, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if res.Restored != 3 || res.Failed != 0 || res.Skipped != 0 {
		t.Fatalf("restore stats: %+v", res)
	}
	if len(res.Mapping) != 3 {
		t.Fatalf("expected 3 id mappings, got %d", len(res.Mapping))
	}
	if len(api.uploads) != 3 {
		t.Fatalf("expected 3 uploads, got %d", len(api.uploads))
	}
}

func TestRestoreCleanupAttachments_SkipsTestEntity(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(77, "test-payload")

	atts := []data.Attachment{{ID: 77, Name: "trace.txt", Size: 12, TestID: 42}}

	snapID := "cleanup-attachments/test-snap-2"
	store := scratchStore(t, snapID)

	if _, _, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, false); err != nil {
		t.Fatalf("backup: %v", err)
	}

	res, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if res.Skipped != 1 || res.Restored != 0 {
		t.Fatalf("expected 1 skipped, got %+v", res)
	}
	if len(res.Failures) != 1 || res.Failures[0].EntityType != "test" {
		t.Fatalf("expected test failure record, got %+v", res.Failures)
	}
}

func TestBackupAttachmentsForCleanup_WithCompression(t *testing.T) {
	api := newFakeCleanupAPI()
	api.seed(1, strings.Repeat("hello ", 200))

	atts := []data.Attachment{{ID: 1, Name: "big.txt", Size: 1200, CaseID: 1}}
	snapID := "cleanup-attachments/test-snap-3"
	store := scratchStore(t, snapID)

	saved, _, err := BackupAttachmentsForCleanup(context.Background(), api, store, snapID, atts, true)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if saved != 1 {
		t.Fatalf("saved=%d", saved)
	}

	var on CleanupAttachmentsData
	if err := store.LoadData(snapID, "data.json", &on); err != nil {
		t.Fatalf("load: %v", err)
	}
	if !on.Attachments[0].Compressed || !strings.HasSuffix(on.Attachments[0].File, ".gz") {
		t.Fatalf("expected compressed .gz, got %+v", on.Attachments[0])
	}

	// Roundtrip restore should still succeed (decompresses to a temp file).
	res, err := RestoreCleanupAttachments(context.Background(), api, store, snapID, false)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if res.Restored != 1 {
		t.Fatalf("restore: %+v", res)
	}
}
