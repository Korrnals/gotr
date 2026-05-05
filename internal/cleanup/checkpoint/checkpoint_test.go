// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package checkpoint

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewRunID_Format(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 34, 56, 0, time.UTC)
	id := NewRunID(now)
	re := regexp.MustCompile(`^20260505T123456-[0-9a-f]{6}$`)
	if !re.MatchString(id) {
		t.Fatalf("unexpected run id %q", id)
	}
}

func TestNewRunID_Unique(t *testing.T) {
	now := time.Now()
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		id := NewRunID(now)
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate run id at iteration %d: %s", i, id)
		}
		seen[id] = struct{}{}
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	return s
}

func TestStore_SaveLoadRoundTrip(t *testing.T) {
	s := newTestStore(t)
	cp := &Checkpoint{
		RunID:     "run-1",
		StartedAt: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
		Filter: FilterSnapshot{
			OlderThanRaw: "90d",
			EntityTypes:  []string{"case", "result"},
			Concurrency:  4,
			ScanStrategy: "auto",
		},
		AllProjects: true,
		ChunkSize:   10,
		Projects: []ProjectStatus{
			{ID: 1, Name: "p1", State: StateDone, Found: 3, Eligible: 2, Bytes: 1024},
			{ID: 2, Name: "p2", State: StatePending},
		},
	}
	if err := s.Save("run-1", cp); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Load("run-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Schema != Schema {
		t.Errorf("Schema = %d, want %d", got.Schema, Schema)
	}
	if got.RunID != "run-1" {
		t.Errorf("RunID = %q", got.RunID)
	}
	if len(got.Projects) != 2 || got.Projects[0].State != StateDone {
		t.Errorf("Projects = %+v", got.Projects)
	}
	if got.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt not stamped")
	}
}

func TestStore_Load_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Load("missing")
	if !errors.Is(err, ErrCheckpointNotFound) {
		t.Fatalf("err = %v, want ErrCheckpointNotFound", err)
	}
}

func TestStore_Load_Malformed(t *testing.T) {
	s := newTestStore(t)
	dir := filepath.Join(s.Root(), "bad")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, FileCheckpoint), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := s.Load("bad")
	if !errors.Is(err, ErrCheckpointMalformed) {
		t.Fatalf("err = %v, want ErrCheckpointMalformed", err)
	}
}

// TestStore_AtomicWrite_LeftoverTmpIgnored simulates a crash that left
// a half-written .tmp-* file behind. Subsequent Save+Load must succeed
// and the orphan must not corrupt Load.
func TestStore_AtomicWrite_LeftoverTmpIgnored(t *testing.T) {
	s := newTestStore(t)
	dir := filepath.Join(s.Root(), "run-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".tmp-orphan"), []byte("partial-junk"), 0o644); err != nil {
		t.Fatalf("write orphan: %v", err)
	}
	cp := &Checkpoint{RunID: "run-1", StartedAt: time.Now().UTC(), Projects: []ProjectStatus{{ID: 1, State: StateDone}}}
	if err := s.Save("run-1", cp); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Load("run-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Projects) != 1 || got.Projects[0].State != StateDone {
		t.Fatalf("Projects = %+v", got.Projects)
	}
}

func TestStore_PartialPlanRoundTrip(t *testing.T) {
	s := newTestStore(t)
	type miniPlan struct {
		Total int    `json:"total"`
		Note  string `json:"note"`
	}
	if err := s.SavePartialPlan("run-1", miniPlan{Total: 7, Note: "ok"}); err != nil {
		t.Fatalf("SavePartialPlan: %v", err)
	}
	var got miniPlan
	if err := s.LoadPartialPlan("run-1", &got); err != nil {
		t.Fatalf("LoadPartialPlan: %v", err)
	}
	if got.Total != 7 || got.Note != "ok" {
		t.Fatalf("got %+v", got)
	}
	var dst miniPlan
	err := s.LoadPartialPlan("missing", &dst)
	if !errors.Is(err, ErrCheckpointNotFound) {
		t.Fatalf("err = %v, want ErrCheckpointNotFound", err)
	}
}

func TestStore_Delete(t *testing.T) {
	s := newTestStore(t)
	cp := &Checkpoint{RunID: "run-1", StartedAt: time.Now().UTC()}
	if err := s.Save("run-1", cp); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := s.Delete("run-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Load("run-1"); !errors.Is(err, ErrCheckpointNotFound) {
		t.Fatalf("Load after delete: %v", err)
	}
}

func TestStore_List_OrderingAndCounts(t *testing.T) {
	s := newTestStore(t)
	t0 := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	mk := func(id string, started time.Time, projects []ProjectStatus) {
		t.Helper()
		cp := &Checkpoint{RunID: id, StartedAt: started, Projects: projects}
		if err := s.Save(id, cp); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}
	mk("a", t0, []ProjectStatus{{ID: 1, State: StateDone}, {ID: 2, State: StatePending}})
	mk("b", t0.Add(2*time.Hour), []ProjectStatus{{ID: 3, State: StateFailed}, {ID: 4, State: StateTimeout}})
	mk("c", t0.Add(time.Hour), []ProjectStatus{{ID: 5, State: StateDone}})

	// junk dir should be ignored.
	if err := os.MkdirAll(filepath.Join(s.Root(), "junk"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	ids := make([]string, len(got))
	for i, c := range got {
		ids[i] = c.RunID
	}
	want := []string{"b", "c", "a"}
	if strings.Join(ids, ",") != strings.Join(want, ",") {
		t.Fatalf("order = %v, want %v", ids, want)
	}
	// Verify counts on "b": 1 failed + 1 timeout.
	if got[0].Failed != 1 || got[0].Timeout != 1 {
		t.Fatalf("counts on b: %+v", got[0])
	}
}
