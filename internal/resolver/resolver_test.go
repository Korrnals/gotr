// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

// fakeClient is a counting test double that lets us assert each ID is
// fetched at most once per session.
type fakeClient struct {
	projectCalls   int
	suiteCalls     int
	runCalls       int
	planCalls      int
	milestoneCalls int
	caseCalls      int
	sectionCalls   int
	userCalls      int
	failProject    bool
}

func (f *fakeClient) GetProject(_ context.Context, id int64) (*data.GetProjectResponse, error) {
	f.projectCalls++
	if f.failProject {
		return nil, errors.New("boom")
	}
	resp := data.GetProjectResponse(data.Project{ID: id, Name: "Project " + itoa(id)})
	return &resp, nil
}
func (f *fakeClient) GetSuite(_ context.Context, id int64) (*data.Suite, error) {
	f.suiteCalls++
	return &data.Suite{ID: id, Name: "Suite " + itoa(id)}, nil
}
func (f *fakeClient) GetRun(_ context.Context, id int64) (*data.Run, error) {
	f.runCalls++
	return &data.Run{ID: id, Name: "Run " + itoa(id)}, nil
}
func (f *fakeClient) GetPlan(_ context.Context, id int64) (*data.Plan, error) {
	f.planCalls++
	return &data.Plan{ID: id, Name: "Plan " + itoa(id)}, nil
}
func (f *fakeClient) GetMilestone(_ context.Context, id int64) (*data.Milestone, error) {
	f.milestoneCalls++
	return &data.Milestone{ID: id, Name: "MS " + itoa(id)}, nil
}
func (f *fakeClient) GetCase(_ context.Context, id int64) (*data.Case, error) {
	f.caseCalls++
	return &data.Case{ID: id, Title: "Case " + itoa(id)}, nil
}
func (f *fakeClient) GetSection(_ context.Context, id int64) (*data.Section, error) {
	f.sectionCalls++
	return &data.Section{ID: id, Name: "Section " + itoa(id)}, nil
}
func (f *fakeClient) GetUser(_ context.Context, id int64) (*data.User, error) {
	f.userCalls++
	return &data.User{ID: id, Email: "u@example.com"}, nil
}

func itoa(i int64) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = digits[i%10]
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}

func TestResolver_LazyAndCached(t *testing.T) {
	t.Parallel()
	f := &fakeClient{}
	r := New(f)

	if got := r.Project(context.Background(), 7); got != "Project 7" {
		t.Fatalf("project name = %q", got)
	}
	if got := r.Project(context.Background(), 7); got != "Project 7" {
		t.Fatalf("project name on second call = %q", got)
	}
	if f.projectCalls != 1 {
		t.Fatalf("expected 1 GetProject call, got %d", f.projectCalls)
	}

	_ = r.Suite(context.Background(), 1)
	_ = r.Suite(context.Background(), 1)
	_ = r.Suite(context.Background(), 2)
	if f.suiteCalls != 2 {
		t.Fatalf("expected 2 GetSuite calls, got %d", f.suiteCalls)
	}
}

func TestResolver_NilSafe(t *testing.T) {
	t.Parallel()
	var r *Resolver
	if got := r.Project(context.Background(), 1); got != "" {
		t.Fatalf("nil resolver should return empty, got %q", got)
	}
}

func TestResolver_NoClient(t *testing.T) {
	t.Parallel()
	r := New(nil)
	if got := r.Project(context.Background(), 1); got != "" {
		t.Fatalf("no-client resolver should return empty, got %q", got)
	}
}

func TestResolver_InvalidID(t *testing.T) {
	t.Parallel()
	f := &fakeClient{}
	r := New(f)
	if got := r.Project(context.Background(), 0); got != "" {
		t.Fatalf("zero id should return empty, got %q", got)
	}
	if got := r.Project(context.Background(), -1); got != "" {
		t.Fatalf("negative id should return empty, got %q", got)
	}
	if f.projectCalls != 0 {
		t.Fatalf("invalid IDs must not call API, got %d calls", f.projectCalls)
	}
}

func TestResolver_FailsCacheNegative(t *testing.T) {
	t.Parallel()
	f := &fakeClient{failProject: true}
	r := New(f)
	if got := r.Project(context.Background(), 5); got != "" {
		t.Fatalf("failing fetch should return empty, got %q", got)
	}
	if got := r.Project(context.Background(), 5); got != "" {
		t.Fatalf("repeat must hit cache, got %q", got)
	}
	if f.projectCalls != 1 {
		t.Fatalf("expected exactly 1 GetProject call on negative cache, got %d", f.projectCalls)
	}
}

func TestResolver_SetProjectSeed(t *testing.T) {
	t.Parallel()
	f := &fakeClient{}
	r := New(f)
	r.SetProject(42, "Pre-seeded")
	if got := r.Project(context.Background(), 42); got != "Pre-seeded" {
		t.Fatalf("seeded project = %q", got)
	}
	if f.projectCalls != 0 {
		t.Fatalf("seed must avoid API, got %d calls", f.projectCalls)
	}
}

func TestResolver_UserNameFallsBackToEmail(t *testing.T) {
	t.Parallel()
	f := &fakeClient{}
	r := New(f)
	if got := r.User(context.Background(), 1); got != "u@example.com" {
		t.Fatalf("user fallback to email failed, got %q", got)
	}
}

func TestResolver_ByKind(t *testing.T) {
	t.Parallel()
	f := &fakeClient{}
	r := New(f)
	cases := []struct {
		kind string
		id   int64
		want string
	}{
		{"case", 1, "Case 1"},
		{"cases", 1, "Case 1"},
		{"run", 2, "Run 2"},
		{"plan", 3, "Plan 3"},
		{"plan_entry", 3, "Plan 3"},
		{"milestone", 4, "MS 4"},
		{"section", 5, "Section 5"},
		{"suite", 6, "Suite 6"},
		{"user", 7, "u@example.com"},
		{"project", 8, "Project 8"},
		{"unknown", 9, ""},
	}
	for _, tc := range cases {
		if got := r.ByKind(context.Background(), tc.kind, tc.id); got != tc.want {
			t.Errorf("ByKind(%q,%d) = %q, want %q", tc.kind, tc.id, got, tc.want)
		}
	}
}
