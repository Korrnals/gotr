// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"context"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/snap/refs"
)

// fakeRewriteAPI is a focused test double for the rewrite-only slice
// of the rollback API. It records every Update* payload so tests can
// assert which fields were patched.
type fakeRewriteAPI struct {
	cases      map[int64]*data.Case
	runs       map[int64]*data.Run
	plans      map[int64]*data.Plan
	milestones map[int64]*data.Milestone

	caseUpdates      map[int64]*data.UpdateCaseRequest
	runUpdates       map[int64]*data.UpdateRunRequest
	planUpdates      map[int64]*data.UpdatePlanRequest
	milestoneUpdates map[int64]*data.UpdateMilestoneRequest
}

func newFakeRewriteAPI() *fakeRewriteAPI {
	return &fakeRewriteAPI{
		cases:            map[int64]*data.Case{},
		runs:             map[int64]*data.Run{},
		plans:            map[int64]*data.Plan{},
		milestones:       map[int64]*data.Milestone{},
		caseUpdates:      map[int64]*data.UpdateCaseRequest{},
		runUpdates:       map[int64]*data.UpdateRunRequest{},
		planUpdates:      map[int64]*data.UpdatePlanRequest{},
		milestoneUpdates: map[int64]*data.UpdateMilestoneRequest{},
	}
}

func (f *fakeRewriteAPI) GetCase(_ context.Context, id int64) (*data.Case, error) {
	return f.cases[id], nil
}
func (f *fakeRewriteAPI) GetRun(_ context.Context, id int64) (*data.Run, error) {
	return f.runs[id], nil
}
func (f *fakeRewriteAPI) GetPlan(_ context.Context, id int64) (*data.Plan, error) {
	return f.plans[id], nil
}
func (f *fakeRewriteAPI) GetMilestone(_ context.Context, id int64) (*data.Milestone, error) {
	return f.milestones[id], nil
}
func (f *fakeRewriteAPI) UpdateCase(_ context.Context, id int64, req *data.UpdateCaseRequest) (*data.Case, error) {
	f.caseUpdates[id] = req
	return f.cases[id], nil
}
func (f *fakeRewriteAPI) UpdateRun(_ context.Context, id int64, req *data.UpdateRunRequest) (*data.Run, error) {
	f.runUpdates[id] = req
	return f.runs[id], nil
}
func (f *fakeRewriteAPI) UpdatePlan(_ context.Context, id int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
	f.planUpdates[id] = req
	return f.plans[id], nil
}
func (f *fakeRewriteAPI) UpdateMilestone(_ context.Context, id int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error) {
	f.milestoneUpdates[id] = req
	return f.milestones[id], nil
}

// TestRewriteReferences_CaseFields verifies that every case markdown
// field touched by the index is rewritten and that a single Update
// call carries all patched fields.
func TestRewriteReferences_CaseFields(t *testing.T) {
	api := newFakeRewriteAPI()
	api.cases[5] = &data.Case{
		ID:             5,
		CustomPreconds: "see /index.php?/attachments/get/1",
		CustomSteps:    "do /index.php?/attachments/get/2",
		CustomExpected: "expect /index.php?/attachments/get/3",
		Refs:           "ticket /index.php?/attachments/get/4",
	}
	entries := []refs.EntityRefs{
		*refs.ScanCase(api.cases[5]),
	}
	idMap := map[int64]int64{1: 11, 2: 22, 3: 33, 4: 44}

	res, err := RewriteReferences(context.Background(), api, entries, idMap)
	if err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if res.EntitiesRewritten != 1 || res.RefsRewritten != 4 || res.RefsSkipped != 0 {
		t.Fatalf("res=%+v", res)
	}
	upd, ok := api.caseUpdates[5]
	if !ok {
		t.Fatalf("UpdateCase not called")
	}
	if upd.CustomPreconds == nil || !strings.Contains(*upd.CustomPreconds, "/11") {
		t.Errorf("preconds not rewritten: %+v", upd.CustomPreconds)
	}
	if upd.CustomSteps == nil || !strings.Contains(*upd.CustomSteps, "/22") {
		t.Errorf("steps not rewritten")
	}
	if upd.CustomExpected == nil || !strings.Contains(*upd.CustomExpected, "/33") {
		t.Errorf("expected not rewritten")
	}
	if upd.Refs == nil || !strings.Contains(*upd.Refs, "/44") {
		t.Errorf("refs not rewritten")
	}
}

// TestRewriteReferences_ResultIsSkipped asserts that result-comment
// references are reported as skipped (TestRail has no update_result).
func TestRewriteReferences_ResultIsSkipped(t *testing.T) {
	api := newFakeRewriteAPI()
	entries := []refs.EntityRefs{
		{
			EntityType: "result",
			EntityID:   42,
			Refs: []refs.Reference{
				{AttachmentID: 1, URL: "/index.php?/attachments/get/1", Field: "comment"},
			},
		},
	}
	res, err := RewriteReferences(context.Background(), api, entries, map[int64]int64{1: 99})
	if err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if res.EntitiesSkipped != 1 || res.RefsSkipped != 1 {
		t.Errorf("res=%+v", res)
	}
	if len(res.Failures) != 1 {
		t.Errorf("expected 1 failure, got %d", len(res.Failures))
	}
}

// TestRewriteReferences_UnmappedIDLeavesEntityUntouched asserts that an
// entity whose only refs are unmapped (e.g. md5) is reported under
// EntitiesSkipped without any Update call.
func TestRewriteReferences_UnmappedIDLeavesEntityUntouched(t *testing.T) {
	api := newFakeRewriteAPI()
	api.runs[7] = &data.Run{
		ID:          7,
		Description: "see /index.php?/attachments/get/abcdef0123456789abcdef0123456789",
	}
	entries := []refs.EntityRefs{*refs.ScanRun(api.runs[7])}
	res, err := RewriteReferences(context.Background(), api, entries, map[int64]int64{})
	if err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if res.EntitiesRewritten != 0 || res.EntitiesSkipped != 1 {
		t.Errorf("res=%+v", res)
	}
	if _, called := api.runUpdates[7]; called {
		t.Errorf("UpdateRun must not be called when no rewrites apply")
	}
}
