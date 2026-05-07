// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	snaplib "github.com/Korrnals/gotr/internal/snap"
)

// fakeAPI is an in-memory AttachmentsAPI for end-to-end cleanup tests.
type fakeAPI struct {
	projects map[int64]data.Project
	atts     map[int64][]data.Attachment // projectID -> attachments
	bodies   map[int64][]byte            // attachment id -> bytes

	deleted   sync.Map // attID -> struct{}
	deletes   atomic.Int64
	failOnDel map[int64]error
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		projects:  map[int64]data.Project{},
		atts:      map[int64][]data.Attachment{},
		bodies:    map[int64][]byte{},
		failOnDel: map[int64]error{},
	}
}

func (f *fakeAPI) GetProjects(_ context.Context) (data.GetProjectsResponse, error) {
	out := make(data.GetProjectsResponse, 0, len(f.projects))
	for _, p := range f.projects {
		out = append(out, p)
	}
	return out, nil
}

func (f *fakeAPI) GetAttachmentsForProject(_ context.Context, id int64) (data.GetAttachmentsResponse, error) {
	return data.GetAttachmentsResponse(f.atts[id]), nil
}

func (f *fakeAPI) DeleteAttachment(_ context.Context, id int64) error {
	if err, ok := f.failOnDel[id]; ok {
		return err
	}
	f.deleted.Store(id, struct{}{})
	f.deletes.Add(1)
	return nil
}

func (f *fakeAPI) DownloadAttachment(_ context.Context, id int64) (io.ReadCloser, error) {
	b, ok := f.bodies[id]
	if !ok {
		return nil, errors.New("body missing")
	}
	return io.NopCloser(strings.NewReader(string(b))), nil
}

func (f *fakeAPI) AddAttachmentToCase(context.Context, int64, string) (*data.AttachmentResponse, error) {
	return &data.AttachmentResponse{AttachmentID: 1}, nil
}
func (f *fakeAPI) AddAttachmentToPlan(context.Context, int64, string) (*data.AttachmentResponse, error) {
	return &data.AttachmentResponse{AttachmentID: 1}, nil
}
func (f *fakeAPI) AddAttachmentToPlanEntry(context.Context, int64, string, string) (*data.AttachmentResponse, error) {
	return &data.AttachmentResponse{AttachmentID: 1}, nil
}
func (f *fakeAPI) AddAttachmentToResult(context.Context, int64, string) (*data.AttachmentResponse, error) {
	return &data.AttachmentResponse{AttachmentID: 1}, nil
}
func (f *fakeAPI) AddAttachmentToRun(context.Context, int64, string) (*data.AttachmentResponse, error) {
	return &data.AttachmentResponse{AttachmentID: 1}, nil
}

// Reference-fetch stubs: the cleanup pipeline now invokes a reference
// scan after backup. Tests don't care about the bodies, so return
// empty entities to keep the scan a no-op.
func (f *fakeAPI) GetCase(context.Context, int64) (*data.Case, error) {
	return &data.Case{}, nil
}
func (f *fakeAPI) GetRun(context.Context, int64) (*data.Run, error) {
	return &data.Run{}, nil
}
func (f *fakeAPI) GetPlan(context.Context, int64) (*data.Plan, error) {
	return &data.Plan{}, nil
}
func (f *fakeAPI) GetMilestone(context.Context, int64) (*data.Milestone, error) {
	return &data.Milestone{}, nil
}

func TestAttachmentFilter_AllowedByAge(t *testing.T) {
	cutoff := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	f := AttachmentFilter{OlderThan: cutoff}

	old := data.Attachment{ID: 1, CreatedOn: cutoff.AddDate(0, -1, 0).Unix()}
	fresh := data.Attachment{ID: 2, CreatedOn: cutoff.AddDate(0, 1, 0).Unix()}
	missing := data.Attachment{ID: 3, CreatedOn: 0}

	if !f.Allowed(old) {
		t.Fatalf("old attachment should be allowed")
	}
	if f.Allowed(fresh) {
		t.Fatalf("fresh attachment must be filtered out")
	}
	if f.Allowed(missing) {
		t.Fatalf("zero CreatedOn must be filtered when age is set")
	}
}

func TestAttachmentFilter_AllowedByEntityType(t *testing.T) {
	f := AttachmentFilter{EntityTypes: map[string]struct{}{"result": {}}}

	res := data.Attachment{ID: 1, ResultID: 5}
	caseAtt := data.Attachment{ID: 2, CaseID: 9}

	if !f.Allowed(res) {
		t.Fatalf("result attachment must pass result-only filter")
	}
	if f.Allowed(caseAtt) {
		t.Fatalf("case attachment must be excluded by result-only filter")
	}
}

func TestExecute_DryRun_DoesNotSnapshotOrDelete(t *testing.T) {
	api := newFakeAPI()
	api.projects[1] = data.Project{ID: 1, Name: "P1"}
	api.atts[1] = []data.Attachment{
		{ID: 11, Name: "a.txt", Size: 1, ResultID: 100, CreatedOn: 1},
		{ID: 12, Name: "b.txt", Size: 2, ResultID: 100, CreatedOn: 2},
	}

	plan, err := BuildPlan(context.Background(), api, []int64{1}, AttachmentFilter{}, 1)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if plan.TotalCount != 2 {
		t.Fatalf("expected 2, got %d", plan.TotalCount)
	}

	store, err := snaplib.NewStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	manifest, err := snaplib.LoadManifest(store)
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}

	res, err := Execute(context.Background(), api, store, manifest, plan, ExecuteOptions{
		DryRun:      true,
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.SnapshotID != "" {
		t.Fatalf("dry-run must not create snapshot; got %s", res.SnapshotID)
	}
	if api.deletes.Load() != 0 {
		t.Fatalf("dry-run must not delete; got %d", api.deletes.Load())
	}
	if res.BackedUp != 2 {
		t.Fatalf("dry-run must report planned count: got %d", res.BackedUp)
	}
}

func TestExecute_FullRun_SnapshotsThenDeletes(t *testing.T) {
	api := newFakeAPI()
	api.projects[1] = data.Project{ID: 1, Name: "P1"}
	api.atts[1] = []data.Attachment{
		{ID: 11, Name: "a.txt", Size: 5, ResultID: 100, CreatedOn: 1},
		{ID: 12, Name: "b.txt", Size: 5, ResultID: 100, CreatedOn: 2},
	}
	api.bodies[11] = []byte("hello")
	api.bodies[12] = []byte("world")

	plan, err := BuildPlan(context.Background(), api, []int64{1}, AttachmentFilter{}, 1)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	store, err := snaplib.NewStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	manifest, err := snaplib.LoadManifest(store)
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}

	res, err := Execute(context.Background(), api, store, manifest, plan, ExecuteOptions{
		Concurrency:   2,
		SnapshotLabel: "test",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.SnapshotID == "" {
		t.Fatalf("expected snapshot ID")
	}
	if res.BackedUp != 2 {
		t.Fatalf("expected backed up = 2, got %d", res.BackedUp)
	}
	if res.Deleted != 2 || res.DeleteErrors != 0 {
		t.Fatalf("expected 2 deletes, got %+v", res)
	}
	if api.deletes.Load() != 2 {
		t.Fatalf("api deletes mismatch: %d", api.deletes.Load())
	}

	// Snapshot must be in manifest under cleanup-attachments category.
	if entry := manifest.Find(res.SnapshotID); entry == nil {
		t.Fatalf("snapshot not in manifest")
	}
}

func TestExecute_PartialDeleteFailures(t *testing.T) {
	api := newFakeAPI()
	api.projects[1] = data.Project{ID: 1, Name: "P1"}
	api.atts[1] = []data.Attachment{
		{ID: 11, Name: "a.txt", Size: 5, ResultID: 100, CreatedOn: 1},
		{ID: 12, Name: "b.txt", Size: 5, ResultID: 100, CreatedOn: 2},
	}
	api.bodies[11] = []byte("ok")
	api.bodies[12] = []byte("nope")
	api.failOnDel[12] = errors.New("403 forbidden")

	plan, err := BuildPlan(context.Background(), api, []int64{1}, AttachmentFilter{}, 1)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	store, _ := snaplib.NewStoreAt(t.TempDir())
	manifest, _ := snaplib.LoadManifest(store)

	res, err := Execute(context.Background(), api, store, manifest, plan, ExecuteOptions{Concurrency: 2})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Deleted != 1 || res.DeleteErrors != 1 {
		t.Fatalf("partial: expected 1/1, got %+v", res)
	}
	if len(res.Failures) != 1 || res.Failures[0].AttachmentID != 12 {
		t.Fatalf("failure record mismatch: %+v", res.Failures)
	}
}

func TestBuildPlan_AppliesLimit(t *testing.T) {
	api := newFakeAPI()
	api.projects[1] = data.Project{ID: 1, Name: "P1"}
	api.projects[2] = data.Project{ID: 2, Name: "P2"}
	api.atts[1] = []data.Attachment{
		{ID: 11, Size: 1, ResultID: 1, CreatedOn: 1},
		{ID: 12, Size: 1, ResultID: 1, CreatedOn: 2},
	}
	api.atts[2] = []data.Attachment{
		{ID: 21, Size: 1, ResultID: 2, CreatedOn: 1},
		{ID: 22, Size: 1, ResultID: 2, CreatedOn: 2},
	}

	plan, err := BuildPlan(context.Background(), api, []int64{1, 2}, AttachmentFilter{Limit: 3}, 2)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if plan.TotalCount != 3 {
		t.Fatalf("limit not applied: total=%d", plan.TotalCount)
	}
	if !plan.TruncatedByLimit {
		t.Fatalf("expected TruncatedByLimit=true")
	}
}
