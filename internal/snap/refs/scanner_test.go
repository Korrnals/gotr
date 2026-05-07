// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package refs

import (
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestScanCase_AllFields(t *testing.T) {
	c := &data.Case{
		ID:             1,
		CustomPreconds: "see /index.php?/attachments/get/10",
		CustomSteps:    "step /index.php?/attachments/get/11",
		CustomExpected: "exp /index.php?/attachments/get/12",
		Refs:           "/index.php?/attachments/get/13",
		CustomStepsSeparated: []data.Step{
			{Content: "c /index.php?/attachments/get/20", Expected: "e /index.php?/attachments/get/21"},
			{Content: "no refs"},
		},
	}
	got := ScanCase(c)
	if got == nil {
		t.Fatal("nil")
	}
	if got.EntityType != "case" || got.EntityID != 1 {
		t.Fatalf("entity meta: %+v", got)
	}
	if len(got.Refs) != 6 {
		t.Fatalf("expected 6 refs, got %d: %+v", len(got.Refs), got.Refs)
	}
	want := []struct {
		ID    int64
		Field string
	}{
		{10, "custom_preconds"},
		{11, "custom_steps"},
		{12, "custom_expected"},
		{13, "refs"},
		{20, "custom_steps_separated[0].content"},
		{21, "custom_steps_separated[0].expected"},
	}
	for i, w := range want {
		if got.Refs[i].AttachmentID != w.ID || got.Refs[i].Field != w.Field {
			t.Errorf("[%d] got id=%d field=%s, want id=%d field=%s",
				i, got.Refs[i].AttachmentID, got.Refs[i].Field, w.ID, w.Field)
		}
	}
}

func TestScanCase_NoRefs(t *testing.T) {
	if got := ScanCase(&data.Case{ID: 5, CustomSteps: "plain"}); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
	if got := ScanCase(nil); got != nil {
		t.Errorf("expected nil for nil input")
	}
}

func TestScanResult(t *testing.T) {
	r := &data.Result{ID: 99, Comment: "see /index.php?/attachments/get/77"}
	got := ScanResult(r)
	if got == nil || got.EntityType != "result" || got.EntityID != 99 || len(got.Refs) != 1 {
		t.Fatalf("got %+v", got)
	}
	if got.Refs[0].Field != "comment" || got.Refs[0].AttachmentID != 77 {
		t.Errorf("ref: %+v", got.Refs[0])
	}
}

func TestScanRun_Plan_Milestone(t *testing.T) {
	if got := ScanRun(&data.Run{ID: 1, Description: "/index.php?/attachments/get/100"}); got == nil || got.Refs[0].AttachmentID != 100 {
		t.Errorf("run: %+v", got)
	}
	if got := ScanPlan(&data.Plan{ID: 2, Description: "/index.php?/attachments/get/200"}); got == nil || got.Refs[0].AttachmentID != 200 {
		t.Errorf("plan: %+v", got)
	}
	if got := ScanMilestone(&data.Milestone{ID: 3, Description: "/index.php?/attachments/get/300"}); got == nil || got.Refs[0].AttachmentID != 300 {
		t.Errorf("milestone: %+v", got)
	}
}

func TestScanCase_StableOrder(t *testing.T) {
	c := &data.Case{
		ID:          1,
		CustomSteps: "/index.php?/attachments/get/3 /index.php?/attachments/get/1",
	}
	got := ScanCase(c)
	if got == nil || len(got.Refs) != 2 {
		t.Fatalf("got %+v", got)
	}
	// document order, not numeric.
	if got.Refs[0].AttachmentID != 3 || got.Refs[1].AttachmentID != 1 {
		t.Errorf("expected document order [3,1], got %+v", got.Refs)
	}
}
