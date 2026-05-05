package cleanup

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
)

// ScanStrategy selects how a project's attachments are enumerated.
//
// TestRail >= 7.5 (and TestRail Cloud) exposes the bulk
// get_attachments_for_project endpoint which returns every attachment
// in a single paginated stream. Older self-hosted servers (notably
// TestRail Server < 7.5) lack that endpoint and respond with HTTP 404
// "Unknown method 'get_attachments_for_project'". For those servers
// the cleanup must walk the entity graph
// (suites → cases, runs, plans → attachments) and dedup by ID.
type ScanStrategy string

const (
	// ScanStrategyAuto probes the project endpoint with a single call
	// against the first project in the cleanup set. On a clean 200 the
	// project scanner is used; on the canonical "Unknown method" 404
	// the entity scanner is used. Any other error aborts (no silent
	// fallback).
	ScanStrategyAuto ScanStrategy = "auto"
	// ScanStrategyProject forces use of get_attachments_for_project.
	// Fails on TestRail Server versions that do not expose the
	// endpoint; the error message is surfaced verbatim so the operator
	// can pick the right strategy.
	ScanStrategyProject ScanStrategy = "project"
	// ScanStrategyEntities forces the suites→cases / runs / plans
	// walk. Use this on TestRail Server < 7.5 or whenever the project
	// endpoint behaves unreliably.
	ScanStrategyEntities ScanStrategy = "entities"
)

// AttachmentScanner is the strategy contract used by BuildPlan to
// enumerate attachments belonging to a single project. Each scanner
// must return a deterministic, deduplicated slice.
type AttachmentScanner interface {
	// Name identifies the scanner in INFO logs and snapshot metadata.
	Name() string
	// Scan returns every attachment that the strategy can see for the
	// project. Filtering by age/parent kind is applied later by the
	// caller.
	Scan(ctx context.Context, projectID int64) ([]data.Attachment, error)
}

// ProjectAttachmentsAPI is the slice of the TestRail client surface
// required by the project-level scanner. Satisfied by *client.HTTPClient.
type ProjectAttachmentsAPI interface {
	GetAttachmentsForProject(ctx context.Context, projectID int64) (data.GetAttachmentsResponse, error)
}

// EntityAttachmentsAPI is the slice required by the entity-walk
// scanner. Satisfied by *client.HTTPClient.
type EntityAttachmentsAPI interface {
	GetSuites(ctx context.Context, projectID int64) (data.GetSuitesResponse, error)
	GetCases(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error)
	GetAttachmentsForCase(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error)
	GetRuns(ctx context.Context, projectID int64) (data.GetRunsResponse, error)
	GetAttachmentsForRun(ctx context.Context, runID int64) (data.GetAttachmentsResponse, error)
	GetPlans(ctx context.Context, projectID int64) (data.GetPlansResponse, error)
	GetPlan(ctx context.Context, planID int64) (*data.Plan, error)
	GetAttachmentsForPlan(ctx context.Context, planID int64) (data.GetAttachmentsResponse, error)
	GetTests(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error)
	GetAttachmentsForTest(ctx context.Context, testID int64) (data.GetAttachmentsResponse, error)
}

// ScannerAPI is the union required by ResolveScanner: both endpoints
// must be reachable, even if only one is ultimately used, because the
// auto-probe needs the project endpoint and the entity scanner needs
// the entity endpoints.
type ScannerAPI interface {
	ProjectAttachmentsAPI
	EntityAttachmentsAPI
}

// projectScanner wraps GetAttachmentsForProject. Single API call per
// project (paginated internally by the client).
type projectScanner struct {
	api ProjectAttachmentsAPI
}

// NewProjectScanner returns a scanner that uses the bulk
// get_attachments_for_project endpoint.
func NewProjectScanner(api ProjectAttachmentsAPI) AttachmentScanner {
	return &projectScanner{api: api}
}

func (s *projectScanner) Name() string { return "project" }

func (s *projectScanner) Scan(ctx context.Context, projectID int64) ([]data.Attachment, error) {
	atts, err := s.api.GetAttachmentsForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return atts, nil
}

// entityScanner walks suites → cases (and optionally runs/plans/tests),
// fetching attachments per entity in parallel and deduplicating by ID.
type entityScanner struct {
	api         EntityAttachmentsAPI
	walkCases   bool
	walkRuns    bool
	walkPlans   bool
	walkTests   bool
	concurrency int
}

// EntityScannerOptions narrows the entity walk. When all booleans are
// false the scanner walks every supported entity kind (cases, runs,
// plans, tests) — matching the default "scan everything" behavior
// callers expect when --entity-type is unset.
type EntityScannerOptions struct {
	WalkCases   bool
	WalkRuns    bool
	WalkPlans   bool
	WalkTests   bool
	Concurrency int
}

// NewEntityScanner returns a scanner that walks the entity graph.
// EntityTypes from the user-supplied AttachmentFilter should be
// translated into walkXxx booleans by the caller via
// EntityScannerOptionsFromTypes.
func NewEntityScanner(api EntityAttachmentsAPI, opts EntityScannerOptions) AttachmentScanner {
	// If no walk is requested, fall back to walking every entity kind:
	// the AttachmentFilter still narrows the result downstream.
	if !opts.WalkCases && !opts.WalkRuns && !opts.WalkPlans && !opts.WalkTests {
		opts.WalkCases = true
		opts.WalkRuns = true
		opts.WalkPlans = true
		opts.WalkTests = true
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 4
	}
	return &entityScanner{
		api:         api,
		walkCases:   opts.WalkCases,
		walkRuns:    opts.WalkRuns,
		walkPlans:   opts.WalkPlans,
		walkTests:   opts.WalkTests,
		concurrency: opts.Concurrency,
	}
}

// EntityScannerOptionsFromTypes translates AttachmentFilter.EntityTypes
// into the entity-walk plan. "case" implies the case walk;
// "result"/"test" imply the tests walk (TestRail exposes result-bound
// attachments via get_attachments_for_test, not under the parent case).
// "plan_entry" implies the plan walk because plan entries are reached
// through their parent plan.
func EntityScannerOptionsFromTypes(types map[string]struct{}, concurrency int) EntityScannerOptions {
	if len(types) == 0 {
		return EntityScannerOptions{WalkCases: true, WalkRuns: true, WalkPlans: true, WalkTests: true, Concurrency: concurrency}
	}
	opts := EntityScannerOptions{Concurrency: concurrency}
	for t := range types {
		switch t {
		case "case":
			opts.WalkCases = true
		case "result", "test":
			opts.WalkTests = true
		case "run":
			opts.WalkRuns = true
		case "plan", "plan_entry":
			opts.WalkPlans = true
		}
	}
	return opts
}

func (s *entityScanner) Name() string { return "entities" }

// stampParent fills in the parent-id field on attachments that the
// server returned without it. Some TestRail Server builds (notably the
// self-hosted < 7.5 fleet) omit run_id/plan_id/case_id in the
// per-entity attachment listings, breaking downstream
// InferredEntityType()-based filtering. We compensate by stamping the
// known parent id from the walk context. EntityType is left untouched
// when the server already populated it (TestRail >= 7.1 cloud format).
func stampParent(atts []data.Attachment, kind string, parentID int64, entryID string) {
	for i := range atts {
		switch kind {
		case "case":
			if atts[i].CaseID == 0 {
				atts[i].CaseID = parentID
			}
		case "run":
			if atts[i].RunID == 0 {
				atts[i].RunID = parentID
			}
		case "plan":
			if atts[i].PlanID == 0 {
				atts[i].PlanID = parentID
			}
		case "plan_entry":
			if atts[i].PlanID == 0 {
				atts[i].PlanID = parentID
			}
			if atts[i].EntryID == "" {
				atts[i].EntryID = entryID
			}
		case "test":
			if atts[i].TestID == 0 {
				atts[i].TestID = parentID
			}
		}
		if atts[i].EntityType == "" {
			// Prefer "result" when the payload has a result_id (TestRail
			// returns these under get_attachments_for_test) so the
			// AttachmentFilter "result" type matches.
			if atts[i].ResultID != 0 {
				atts[i].EntityType = "result"
			} else {
				atts[i].EntityType = kind
			}
		}
	}
}

//nolint:gocyclo // Sequential per-kind walks (cases / runs / plans) are clearer kept inline.
func (s *entityScanner) Scan(ctx context.Context, projectID int64) ([]data.Attachment, error) {
	seen := make(map[int64]struct{})
	var out []data.Attachment

	collect := func(items []data.Attachment) {
		for _, a := range items {
			if _, ok := seen[a.ID]; ok {
				continue
			}
			seen[a.ID] = struct{}{}
			out = append(out, a)
		}
	}

	if s.walkCases {
		suites, err := s.api.GetSuites(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("get_suites %d: %w", projectID, err)
		}
		for _, su := range suites {
			cases, err := s.api.GetCases(ctx, projectID, su.ID, 0)
			if err != nil {
				return nil, fmt.Errorf("get_cases p%d/s%d: %w", projectID, su.ID, err)
			}
			res, _ := concurrent.ParallelMap(ctx, cases, s.concurrency, func(c data.Case, _ int) ([]data.Attachment, error) {
				atts, err := s.api.GetAttachmentsForCase(ctx, c.ID)
				if err != nil {
					return nil, fmt.Errorf("get_attachments_for_case %d: %w", c.ID, err)
				}
				stampParent(atts, "case", c.ID, "")
				return atts, nil
			})
			for _, r := range res {
				if r.Error != nil {
					return nil, r.Error
				}
				collect(r.Data)
			}
		}
	}

	if s.walkRuns {
		runs, err := s.api.GetRuns(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("get_runs %d: %w", projectID, err)
		}
		res, _ := concurrent.ParallelMap(ctx, runs, s.concurrency, func(r data.Run, _ int) ([]data.Attachment, error) {
			atts, err := s.api.GetAttachmentsForRun(ctx, r.ID)
			if err != nil {
				return nil, fmt.Errorf("get_attachments_for_run %d: %w", r.ID, err)
			}
			stampParent(atts, "run", r.ID, "")
			return atts, nil
		})
		for _, r := range res {
			if r.Error != nil {
				return nil, r.Error
			}
			collect(r.Data)
		}
	}

	if s.walkPlans {
		plans, err := s.api.GetPlans(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("get_plans %d: %w", projectID, err)
		}
		res, _ := concurrent.ParallelMap(ctx, plans, s.concurrency, func(p data.Plan, _ int) ([]data.Attachment, error) {
			atts, err := s.api.GetAttachmentsForPlan(ctx, p.ID)
			if err != nil {
				return nil, fmt.Errorf("get_attachments_for_plan %d: %w", p.ID, err)
			}
			stampParent(atts, "plan", p.ID, "")
			return atts, nil
		})
		for _, r := range res {
			if r.Error != nil {
				return nil, r.Error
			}
			collect(r.Data)
		}
	}

	if s.walkTests {
		runIDs, err := s.collectRunIDs(ctx, projectID)
		if err != nil {
			return nil, err
		}
		// Enumerate tests per run, then attachments per test. Two
		// levels of parallelism; cap inner to avoid burst on large
		// runs by reusing the same concurrency budget.
		testsPerRun, _ := concurrent.ParallelMap(ctx, runIDs, s.concurrency, func(runID int64, _ int) ([]data.Test, error) {
			tests, err := s.api.GetTests(ctx, runID, nil)
			if err != nil {
				return nil, fmt.Errorf("get_tests run=%d: %w", runID, err)
			}
			return tests, nil
		})
		var allTests []data.Test
		for _, r := range testsPerRun {
			if r.Error != nil {
				return nil, r.Error
			}
			allTests = append(allTests, r.Data...)
		}
		attsPerTest, _ := concurrent.ParallelMap(ctx, allTests, s.concurrency, func(t data.Test, _ int) ([]data.Attachment, error) {
			atts, err := s.api.GetAttachmentsForTest(ctx, t.ID)
			if err != nil {
				return nil, fmt.Errorf("get_attachments_for_test %d: %w", t.ID, err)
			}
			stampParent(atts, "test", t.ID, "")
			return atts, nil
		})
		for _, r := range attsPerTest {
			if r.Error != nil {
				return nil, r.Error
			}
			collect(r.Data)
		}
	}

	return out, nil
}

// collectRunIDs returns the union of project-level run IDs (GetRuns)
// and plan-bound run IDs (GetPlans → GetPlan → Entries[].Runs[]). The
// tests walk uses this set to enumerate every test reachable in the
// project. Duplicates are removed.
func (s *entityScanner) collectRunIDs(ctx context.Context, projectID int64) ([]int64, error) {
	seen := make(map[int64]struct{})
	var ids []int64
	add := func(id int64) {
		if id == 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	runs, err := s.api.GetRuns(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get_runs %d: %w", projectID, err)
	}
	for _, r := range runs {
		add(r.ID)
	}

	plans, err := s.api.GetPlans(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get_plans %d: %w", projectID, err)
	}
	planDetails, _ := concurrent.ParallelMap(ctx, plans, s.concurrency, func(p data.Plan, _ int) (*data.Plan, error) {
		full, err := s.api.GetPlan(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("get_plan %d: %w", p.ID, err)
		}
		return full, nil
	})
	for _, r := range planDetails {
		if r.Error != nil {
			return nil, r.Error
		}
		if r.Data == nil {
			continue
		}
		for _, e := range r.Data.Entries {
			for _, run := range e.Runs {
				add(run.ID)
			}
		}
	}
	return ids, nil
}

// ResolveScanner picks the scanner implementation honoring strategy.
// On ScanStrategyAuto it issues a single probe call to
// GetAttachmentsForProject(probeProjectID); on the canonical "Unknown
// method" 404 (see client.IsAPIMethodNotFound) it falls back to the
// entity scanner and emits an INFO log via the optional logf hook.
// Any other probe error aborts: the cleanup workflow refuses to guess.
//
// probeProjectID must reference a project the API key can see. The
// caller is expected to pass the first project from the resolved set.
func ResolveScanner(
	ctx context.Context,
	api ScannerAPI,
	strategy ScanStrategy,
	entityOpts EntityScannerOptions,
	probeProjectID int64,
	logf func(format string, args ...any),
) (AttachmentScanner, error) {
	if strategy == "" {
		strategy = ScanStrategyAuto
	}
	switch strategy {
	case ScanStrategyProject:
		return NewProjectScanner(api), nil
	case ScanStrategyEntities:
		return NewEntityScanner(api, entityOpts), nil
	case ScanStrategyAuto:
		if probeProjectID == 0 {
			return nil, fmt.Errorf("auto scan strategy requires a project to probe (no projects resolved)")
		}
		_, err := api.GetAttachmentsForProject(ctx, probeProjectID)
		if err == nil {
			if logf != nil {
				logf("scan strategy: project (probe ok)")
			}
			return NewProjectScanner(api), nil
		}
		if client.IsAPIMethodNotFound(err) {
			if logf != nil {
				logf("scan strategy: entities (server lacks get_attachments_for_project — walking suites→cases/runs/plans)")
			}
			return NewEntityScanner(api, entityOpts), nil
		}
		return nil, fmt.Errorf("probe get_attachments_for_project on project %d: %w", probeProjectID, err)
	default:
		return nil, fmt.Errorf("unknown scan strategy %q", strategy)
	}
}
