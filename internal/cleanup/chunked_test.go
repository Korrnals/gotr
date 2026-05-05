// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/cleanup/checkpoint"
	"github.com/Korrnals/gotr/internal/models/data"
)

// fakeProjectLister implements ProjectLister with an explicit list.
type fakeProjectLister struct {
	projects []data.Project
}

func (f *fakeProjectLister) GetProjects(_ context.Context) (data.GetProjectsResponse, error) {
	return data.GetProjectsResponse(f.projects), nil
}

// fakeChunkScanner is an AttachmentScanner mock with per-project
// behavior controllable by the test.
type fakeChunkScanner struct {
	mu      sync.Mutex
	calls   []int64
	delay   time.Duration            // applied before returning
	atts    map[int64][]data.Attachment
	errs    map[int64]error          // returned for that project ID
	timeout map[int64]time.Duration  // sleep for that long inside Scan, ignored if ctx cancels
}

func (f *fakeChunkScanner) Name() string { return "fake" }

func (f *fakeChunkScanner) Scan(ctx context.Context, projectID int64) ([]data.Attachment, error) {
	f.mu.Lock()
	f.calls = append(f.calls, projectID)
	f.mu.Unlock()

	if d, ok := f.timeout[projectID]; ok {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	} else if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if err, ok := f.errs[projectID]; ok {
		return nil, err
	}
	return f.atts[projectID], nil
}

func newFakeChunkSetup(projectCount int) (*fakeProjectLister, *fakeChunkScanner) {
	projects := make([]data.Project, 0, projectCount)
	atts := make(map[int64][]data.Attachment, projectCount)
	for i := 1; i <= projectCount; i++ {
		id := int64(i)
		projects = append(projects, data.Project{ID: id, Name: "p" + itoaSimple(i)})
		atts[id] = []data.Attachment{{ID: id*100 + 1, Size: 100, CreatedOn: 1000}}
	}
	return &fakeProjectLister{projects: projects}, &fakeChunkScanner{atts: atts}
}

func itoaSimple(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func newChunkStore(t *testing.T) *checkpoint.Store {
	t.Helper()
	s, err := checkpoint.NewStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	return s
}

func TestBuildPlanChunked_FlushesAtChunkBoundaries(t *testing.T) {
	lister, scanner := newFakeChunkSetup(7)
	store := newChunkStore(t)

	var chunkCalls atomic.Int32
	cfg := ChunkConfig{
		ChunkSize: 3,
		Store:     store,
		OnChunkComplete: func(_, _ int, _ *Plan) {
			chunkCalls.Add(1)
		},
	}
	plan, cp, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 2, cfg)
	if err != nil {
		t.Fatalf("BuildPlanChunked: %v", err)
	}
	if got := chunkCalls.Load(); got != 3 {
		t.Errorf("chunk callbacks = %d, want 3 (7 projects / chunk size 3)", got)
	}
	if plan.TotalCount != 7 {
		t.Errorf("plan.TotalCount = %d, want 7", plan.TotalCount)
	}
	for _, ps := range cp.Projects {
		if ps.State != checkpoint.StateDone {
			t.Errorf("project %d state = %q", ps.ID, ps.State)
		}
	}
	// Clean completion → checkpoint dir must be gone.
	if _, err := store.Load(cp.RunID); !errors.Is(err, checkpoint.ErrCheckpointNotFound) {
		t.Errorf("expected checkpoint auto-deleted, got %v", err)
	}
}

func TestBuildPlanChunked_TimeoutMarksOnlyAffectedProject(t *testing.T) {
	lister, scanner := newFakeChunkSetup(3)
	scanner.timeout = map[int64]time.Duration{2: 200 * time.Millisecond}
	store := newChunkStore(t)

	cfg := ChunkConfig{
		ChunkSize:             10, // single chunk
		ScanTimeoutPerProject: 30 * time.Millisecond,
		Store:                 store,
	}
	_, cp, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 3, cfg)
	if err != nil {
		t.Fatalf("BuildPlanChunked: %v", err)
	}
	got := map[int64]string{}
	for _, ps := range cp.Projects {
		got[ps.ID] = ps.State
	}
	if got[1] != checkpoint.StateDone || got[3] != checkpoint.StateDone {
		t.Errorf("expected projects 1 and 3 done, got %v", got)
	}
	if got[2] != checkpoint.StateTimeout {
		t.Errorf("project 2 state = %q, want timeout", got[2])
	}
	// Failure preserves checkpoint on disk.
	if _, err := store.Load(cp.RunID); err != nil {
		t.Errorf("checkpoint should be preserved on partial failure: %v", err)
	}
}

func TestBuildPlanChunked_FailedProjectPreservesCheckpoint(t *testing.T) {
	lister, scanner := newFakeChunkSetup(2)
	scanner.errs = map[int64]error{2: errors.New("boom")}
	store := newChunkStore(t)

	cfg := ChunkConfig{ChunkSize: 5, Store: store}
	_, cp, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 2, cfg)
	if err != nil {
		t.Fatalf("BuildPlanChunked: %v", err)
	}
	state := map[int64]string{}
	for _, ps := range cp.Projects {
		state[ps.ID] = ps.State
	}
	if state[1] != checkpoint.StateDone || state[2] != checkpoint.StateFailed {
		t.Fatalf("states = %v", state)
	}
	if _, err := store.Load(cp.RunID); err != nil {
		t.Errorf("checkpoint should be preserved on failure: %v", err)
	}
}

func TestBuildPlanChunked_ResumeSkipsDoneProjects(t *testing.T) {
	lister, scanner := newFakeChunkSetup(3)
	scanner.errs = map[int64]error{2: errors.New("boom")}
	store := newChunkStore(t)

	cfg := ChunkConfig{ChunkSize: 5, Store: store}
	_, cp, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 2, cfg)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	originalCalls := append([]int64(nil), scanner.calls...)
	if len(originalCalls) != 3 {
		t.Fatalf("first-run calls = %d, want 3", len(originalCalls))
	}

	// Flip project 2 to retry_pending and resume.
	loaded, err := store.Load(cp.RunID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for i := range loaded.Projects {
		if loaded.Projects[i].ID == 2 {
			loaded.Projects[i].State = checkpoint.StateRetryPending
		}
	}
	if err := store.Save(cp.RunID, loaded); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Second pass: clear errors so #2 succeeds.
	scanner.errs = nil
	scanner.calls = nil

	cfg2 := ChunkConfig{ChunkSize: 5, Store: store, Resume: true, RunID: cp.RunID}
	_, cp2, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 2, cfg2)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if len(scanner.calls) != 1 || scanner.calls[0] != 2 {
		t.Fatalf("resume rescanned %v, want only [2]", scanner.calls)
	}
	for _, ps := range cp2.Projects {
		if ps.State != checkpoint.StateDone {
			t.Errorf("project %d state = %q", ps.ID, ps.State)
		}
	}
	// Clean completion on resume → checkpoint deleted.
	if _, err := store.Load(cp.RunID); !errors.Is(err, checkpoint.ErrCheckpointNotFound) {
		t.Errorf("checkpoint should be deleted on clean resume: %v", err)
	}
}

func TestBuildPlanChunked_ResumeMismatchedFilterRejected(t *testing.T) {
	lister, scanner := newFakeChunkSetup(2)
	scanner.errs = map[int64]error{1: errors.New("boom")} // force preserve
	store := newChunkStore(t)

	original := ChunkConfig{
		ChunkSize: 5,
		Store:     store,
		FilterSnapshot: checkpoint.FilterSnapshot{
			OlderThanRaw: "90d",
			EntityTypes:  []string{"case"},
			ScanStrategy: "auto",
		},
	}
	_, cp, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 2, original)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}

	tampered := ChunkConfig{
		ChunkSize: 5,
		Store:     store,
		Resume:    true,
		RunID:     cp.RunID,
		FilterSnapshot: checkpoint.FilterSnapshot{
			OlderThanRaw: "30d", // changed
			EntityTypes:  []string{"case"},
			ScanStrategy: "auto",
		},
	}
	_, _, err = BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 2, tampered)
	if !errors.Is(err, ErrCheckpointMismatch) {
		t.Fatalf("err = %v, want ErrCheckpointMismatch", err)
	}
}

func TestBuildPlanChunked_StoreRequired(t *testing.T) {
	lister, scanner := newFakeChunkSetup(1)
	_, _, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{}, 1, ChunkConfig{})
	if err == nil {
		t.Fatalf("expected error on missing Store")
	}
}

func TestBuildPlanChunked_LimitAppliedAfterChunks(t *testing.T) {
	lister, scanner := newFakeChunkSetup(5)
	store := newChunkStore(t)

	cfg := ChunkConfig{ChunkSize: 2, Store: store}
	plan, _, err := BuildPlanChunked(context.Background(), lister, scanner, nil, AttachmentFilter{Limit: 3}, 2, cfg)
	if err != nil {
		t.Fatalf("BuildPlanChunked: %v", err)
	}
	if !plan.TruncatedByLimit {
		t.Errorf("expected TruncatedByLimit=true")
	}
	if plan.TotalCount != 3 {
		t.Errorf("plan.TotalCount = %d, want 3", plan.TotalCount)
	}
}
