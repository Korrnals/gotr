package cleanup

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

// fakeScannerAPI is a permissive ScannerAPI implementation. Each
// per-method override lets tests dictate behavior for that endpoint
// alone (success, error, fixture data).
type fakeScannerAPI struct {
	getAttachmentsForProject func(int64) (data.GetAttachmentsResponse, error)
	getSuites                func(int64) (data.GetSuitesResponse, error)
	getCases                 func(projectID, suiteID int64) (data.GetCasesResponse, error)
	getAttachmentsForCase    func(int64) (data.GetAttachmentsResponse, error)
	getRuns                  func(int64) (data.GetRunsResponse, error)
	getAttachmentsForRun     func(int64) (data.GetAttachmentsResponse, error)
	getPlans                 func(int64) (data.GetPlansResponse, error)
	getAttachmentsForPlan    func(int64) (data.GetAttachmentsResponse, error)
}

func (f *fakeScannerAPI) GetAttachmentsForProject(_ context.Context, id int64) (data.GetAttachmentsResponse, error) {
	if f.getAttachmentsForProject != nil {
		return f.getAttachmentsForProject(id)
	}
	return nil, nil
}
func (f *fakeScannerAPI) GetSuites(_ context.Context, id int64) (data.GetSuitesResponse, error) {
	if f.getSuites != nil {
		return f.getSuites(id)
	}
	return nil, nil
}
func (f *fakeScannerAPI) GetCases(_ context.Context, p, s, _ int64) (data.GetCasesResponse, error) {
	if f.getCases != nil {
		return f.getCases(p, s)
	}
	return nil, nil
}
func (f *fakeScannerAPI) GetAttachmentsForCase(_ context.Context, id int64) (data.GetAttachmentsResponse, error) {
	if f.getAttachmentsForCase != nil {
		return f.getAttachmentsForCase(id)
	}
	return nil, nil
}
func (f *fakeScannerAPI) GetRuns(_ context.Context, id int64) (data.GetRunsResponse, error) {
	if f.getRuns != nil {
		return f.getRuns(id)
	}
	return nil, nil
}
func (f *fakeScannerAPI) GetAttachmentsForRun(_ context.Context, id int64) (data.GetAttachmentsResponse, error) {
	if f.getAttachmentsForRun != nil {
		return f.getAttachmentsForRun(id)
	}
	return nil, nil
}
func (f *fakeScannerAPI) GetPlans(_ context.Context, id int64) (data.GetPlansResponse, error) {
	if f.getPlans != nil {
		return f.getPlans(id)
	}
	return nil, nil
}
func (f *fakeScannerAPI) GetAttachmentsForPlan(_ context.Context, id int64) (data.GetAttachmentsResponse, error) {
	if f.getAttachmentsForPlan != nil {
		return f.getAttachmentsForPlan(id)
	}
	return nil, nil
}

func TestResolveScanner_AutoProbeOK_PicksProject(t *testing.T) {
	api := &fakeScannerAPI{
		getAttachmentsForProject: func(int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{}, nil
		},
	}
	var msg string
	logf := func(f string, a ...any) { msg = f }
	sc, err := ResolveScanner(context.Background(), api, ScanStrategyAuto, EntityScannerOptions{}, 1, logf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.Name() != "project" {
		t.Fatalf("got %q, want project", sc.Name())
	}
	if msg == "" {
		t.Fatalf("expected logf to be called")
	}
}

func TestResolveScanner_AutoProbeUnknownMethod_FallsBackToEntities(t *testing.T) {
	api := &fakeScannerAPI{
		getAttachmentsForProject: func(int64) (data.GetAttachmentsResponse, error) {
			return nil, errors.New("API returned 404 File Not Found: Unknown method 'get_attachments_for_project'")
		},
	}
	sc, err := ResolveScanner(context.Background(), api, ScanStrategyAuto, EntityScannerOptions{}, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.Name() != "entities" {
		t.Fatalf("got %q, want entities", sc.Name())
	}
}

func TestResolveScanner_AutoProbeOtherError_Aborts(t *testing.T) {
	api := &fakeScannerAPI{
		getAttachmentsForProject: func(int64) (data.GetAttachmentsResponse, error) {
			return nil, errors.New("API returned 500 Internal Server Error")
		},
	}
	if _, err := ResolveScanner(context.Background(), api, ScanStrategyAuto, EntityScannerOptions{}, 1, nil); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestResolveScanner_ExplicitOverridesProbe(t *testing.T) {
	probed := 0
	api := &fakeScannerAPI{
		getAttachmentsForProject: func(int64) (data.GetAttachmentsResponse, error) {
			probed++
			return data.GetAttachmentsResponse{}, nil
		},
	}
	sc, err := ResolveScanner(context.Background(), api, ScanStrategyEntities, EntityScannerOptions{WalkCases: true}, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.Name() != "entities" {
		t.Fatalf("got %q, want entities", sc.Name())
	}
	if probed != 0 {
		t.Fatalf("explicit strategy must skip probe, called %d times", probed)
	}

	sc, err = ResolveScanner(context.Background(), api, ScanStrategyProject, EntityScannerOptions{}, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.Name() != "project" {
		t.Fatalf("got %q, want project", sc.Name())
	}
}

func TestResolveScanner_AutoRequiresProbeProject(t *testing.T) {
	api := &fakeScannerAPI{}
	if _, err := ResolveScanner(context.Background(), api, ScanStrategyAuto, EntityScannerOptions{}, 0, nil); err == nil {
		t.Fatalf("expected error when probeProjectID == 0, got nil")
	}
}

func TestEntityScanner_WalksSuitesAndDeduplicates(t *testing.T) {
	api := &fakeScannerAPI{
		getSuites: func(int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 100}, {ID: 101}}, nil
		},
		getCases: func(_, suiteID int64) (data.GetCasesResponse, error) {
			switch suiteID {
			case 100:
				return data.GetCasesResponse{{ID: 1}, {ID: 2}}, nil
			case 101:
				return data.GetCasesResponse{{ID: 3}}, nil
			}
			return nil, nil
		},
		getAttachmentsForCase: func(caseID int64) (data.GetAttachmentsResponse, error) {
			// Same attachment id 999 appears in two cases; must be
			// emitted once.
			switch caseID {
			case 1:
				return data.GetAttachmentsResponse{{ID: 999, Size: 10}, {ID: 100, Size: 5}}, nil
			case 2:
				return data.GetAttachmentsResponse{{ID: 999, Size: 10}}, nil
			case 3:
				return data.GetAttachmentsResponse{{ID: 200, Size: 7}}, nil
			}
			return nil, nil
		},
	}
	sc := NewEntityScanner(api, EntityScannerOptions{WalkCases: true, Concurrency: 2})
	got, err := sc.Scan(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d attachments, want 3 (deduplicated): %+v", len(got), got)
	}
	ids := make([]int64, 0, len(got))
	for _, a := range got {
		ids = append(ids, a.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	want := []int64{100, 200, 999}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %v, want %v", ids, want)
		}
	}
}

func TestEntityScanner_PropagatesCaseError(t *testing.T) {
	api := &fakeScannerAPI{
		getSuites: func(int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 1}}, nil
		},
		getCases: func(int64, int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 11}}, nil
		},
		getAttachmentsForCase: func(int64) (data.GetAttachmentsResponse, error) {
			return nil, errors.New("boom")
		},
	}
	sc := NewEntityScanner(api, EntityScannerOptions{WalkCases: true})
	if _, err := sc.Scan(context.Background(), 1); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestEntityScannerOptionsFromTypes(t *testing.T) {
	cases := []struct {
		name              string
		types             map[string]struct{}
		wantCases, wantRuns, wantPlans bool
	}{
		{"empty -> all", nil, true, true, true},
		{"case", map[string]struct{}{"case": {}}, true, false, false},
		{"result implies case", map[string]struct{}{"result": {}}, true, false, false},
		{"test implies case", map[string]struct{}{"test": {}}, true, false, false},
		{"run only", map[string]struct{}{"run": {}}, false, true, false},
		{"plan_entry implies plan", map[string]struct{}{"plan_entry": {}}, false, false, true},
		{"mixed", map[string]struct{}{"case": {}, "run": {}}, true, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EntityScannerOptionsFromTypes(tc.types, 4)
			if got.WalkCases != tc.wantCases || got.WalkRuns != tc.wantRuns || got.WalkPlans != tc.wantPlans {
				t.Fatalf("got %+v, want cases=%v runs=%v plans=%v", got, tc.wantCases, tc.wantRuns, tc.wantPlans)
			}
		})
	}
}
