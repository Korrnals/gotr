// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/snap/refs"
	"go.uber.org/zap"
)

// ReferenceFetchAPI is the read slice required to walk markdown
// fields of every entity that owned a deleted attachment, so that
// references.json can be persisted at backup time.
//
// Result entities are intentionally excluded — TestRail has no
// update_result endpoint, so refs that point into result.comment
// can never be rewritten and we don't bother indexing them.
type ReferenceFetchAPI interface {
	GetCase(ctx context.Context, caseID int64) (*data.Case, error)
	GetRun(ctx context.Context, runID int64) (*data.Run, error)
	GetPlan(ctx context.Context, planID int64) (*data.Plan, error)
	GetMilestone(ctx context.Context, milestoneID int64) (*data.Milestone, error)
}

// ReferenceRewriteAPI is the write slice required to apply rewritten
// markdown fields back to TestRail during restore. It is a strict
// superset of ReferenceFetchAPI for ergonomics.
type ReferenceRewriteAPI interface {
	ReferenceFetchAPI
	UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	UpdateRun(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error)
	UpdatePlan(ctx context.Context, planID int64, req *data.UpdatePlanRequest) (*data.Plan, error)
	UpdateMilestone(ctx context.Context, milestoneID int64, req *data.UpdateMilestoneRequest) (*data.Milestone, error)
}

// ScanReferencesForAttachments walks every entity that owned one of
// the deleted attachments and returns the per-entity reference index.
// Entities are deduplicated by (entity_type, entity_id). API errors
// are logged and the affected entity is skipped — partial indexing is
// preferable to aborting the whole cleanup.
//
// Supported entity_type values: case, run, plan. plan_entry refs are
// resolved via the parent plan, milestones and tests are not walked
// here (milestones don't appear as attachment parents in practice; tests
// are non-restorable anyway).
func ScanReferencesForAttachments(
	ctx context.Context,
	api ReferenceFetchAPI,
	atts []data.Attachment,
) []refs.EntityRefs {
	type key struct {
		t  string
		id int64
	}
	seen := map[key]bool{}
	var targets []key
	for _, att := range atts {
		t := att.InferredEntityType()
		var id int64
		switch t {
		case "case":
			id = att.CaseID
		case "run":
			id = att.RunID
		case "plan", "plan_entry":
			t = "plan"
			id = att.PlanID
		default:
			continue
		}
		if id == 0 {
			continue
		}
		k := key{t: t, id: id}
		if seen[k] {
			continue
		}
		seen[k] = true
		targets = append(targets, k)
	}
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].t != targets[j].t {
			return targets[i].t < targets[j].t
		}
		return targets[i].id < targets[j].id
	})

	out := make([]refs.EntityRefs, 0, len(targets))
	for _, k := range targets {
		if err := ctx.Err(); err != nil {
			return out
		}
		var entry *refs.EntityRefs
		switch k.t {
		case "case":
			c, err := api.GetCase(ctx, k.id)
			if err != nil {
				log.Warn("ref-scan: get_case failed", zap.Int64("case_id", k.id), zap.Error(err))
				continue
			}
			entry = refs.ScanCase(c)
		case "run":
			r, err := api.GetRun(ctx, k.id)
			if err != nil {
				log.Warn("ref-scan: get_run failed", zap.Int64("run_id", k.id), zap.Error(err))
				continue
			}
			entry = refs.ScanRun(r)
		case "plan":
			p, err := api.GetPlan(ctx, k.id)
			if err != nil {
				log.Warn("ref-scan: get_plan failed", zap.Int64("plan_id", k.id), zap.Error(err))
				continue
			}
			entry = refs.ScanPlan(p)
		}
		if entry != nil && len(entry.Refs) > 0 {
			out = append(out, *entry)
		}
	}
	return out
}

// ReferenceRewriteResult summarises the per-entity rewrite phase.
type ReferenceRewriteResult struct {
	// EntitiesRewritten is the count of entities that received an
	// Update* call with at least one substituted reference.
	EntitiesRewritten int
	// EntitiesSkipped is entities where no numeric ID resolved (e.g.
	// only md5 refs, or every old ID was unmapped).
	EntitiesSkipped int
	// EntitiesFailed is entities where the API call returned an error.
	EntitiesFailed int
	// RefsRewritten is the total number of substituted URLs across all
	// entities.
	RefsRewritten int
	// RefsSkipped is the total number of refs left intact (md5 or
	// unmapped numeric IDs).
	RefsSkipped int
	// Failures records per-entity errors for the report.
	Failures []ReferenceRewriteFailure
}

// ReferenceRewriteFailure is a single per-entity failure record.
type ReferenceRewriteFailure struct {
	EntityType string
	EntityID   int64
	Error      string
}

// RewriteReferences applies idMap to every entity in entries and
// pushes the rewritten markdown fields back to TestRail. Entries
// whose entity type has no Update API (currently only "result") are
// counted under EntitiesSkipped.
//
// Best-effort: per-entity errors are recorded and execution
// continues. Returns nil error only on context cancellation.
//
//nolint:gocyclo // Per-entity dispatch (case/run/plan/milestone) is more readable as a single switch than fan-out helpers.
func RewriteReferences(
	ctx context.Context,
	api ReferenceRewriteAPI,
	entries []refs.EntityRefs,
	idMap map[int64]int64,
) (*ReferenceRewriteResult, error) {
	res := &ReferenceRewriteResult{}
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		switch e.EntityType {
		case "case":
			rewriteOneCase(ctx, api, &e, idMap, res)
		case "run":
			rewriteOneRun(ctx, api, &e, idMap, res)
		case "plan":
			rewriteOnePlan(ctx, api, &e, idMap, res)
		case "milestone":
			rewriteOneMilestone(ctx, api, &e, idMap, res)
		case "result":
			// TestRail has no update_result endpoint; comment refs
			// cannot be rewritten. Count under skipped.
			res.EntitiesSkipped++
			res.RefsSkipped += len(e.Refs)
			res.Failures = append(res.Failures, ReferenceRewriteFailure{
				EntityType: e.EntityType,
				EntityID:   e.EntityID,
				Error:      "result refs cannot be rewritten: TestRail has no update_result API",
			})
		default:
			res.EntitiesSkipped++
			res.Failures = append(res.Failures, ReferenceRewriteFailure{
				EntityType: e.EntityType,
				EntityID:   e.EntityID,
				Error:      fmt.Sprintf("unsupported entity_type: %s", e.EntityType),
			})
		}
	}
	return res, nil
}

// fieldSet collects unique top-level field paths covered by the
// reference index of one entity.
func fieldSet(refsList []refs.Reference) map[string]bool {
	out := map[string]bool{}
	for _, r := range refsList {
		// trim "[idx]..." → top-level json key for switching purposes.
		key := r.Field
		if dot := strings.IndexByte(key, '.'); dot != -1 {
			key = key[:dot]
		}
		if br := strings.IndexByte(key, '['); br != -1 {
			key = key[:br]
		}
		out[key] = true
	}
	return out
}

func rewriteOneCase(ctx context.Context, api ReferenceRewriteAPI, e *refs.EntityRefs, idMap map[int64]int64, res *ReferenceRewriteResult) {
	c, err := api.GetCase(ctx, e.EntityID)
	if err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "case", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	upd := &data.UpdateCaseRequest{}
	changed := 0
	skipped := 0
	fields := fieldSet(e.Refs)
	if fields["custom_preconds"] {
		if newText, n, sk := refs.Rewrite(c.CustomPreconds, idMap); n > 0 {
			upd.CustomPreconds = strPtr(newText)
			changed += n
		} else {
			skipped += sk
		}
	}
	if fields["custom_steps"] {
		if newText, n, sk := refs.Rewrite(c.CustomSteps, idMap); n > 0 {
			upd.CustomSteps = strPtr(newText)
			changed += n
		} else {
			skipped += sk
		}
	}
	if fields["custom_expected"] {
		if newText, n, sk := refs.Rewrite(c.CustomExpected, idMap); n > 0 {
			upd.CustomExpected = strPtr(newText)
			changed += n
		} else {
			skipped += sk
		}
	}
	if fields["refs"] {
		if newText, n, sk := refs.Rewrite(c.Refs, idMap); n > 0 {
			upd.Refs = strPtr(newText)
			changed += n
		} else {
			skipped += sk
		}
	}
	if fields["custom_steps_separated"] && len(c.CustomStepsSeparated) > 0 {
		newSteps := append([]data.Step(nil), c.CustomStepsSeparated...)
		stepChanged := false
		for i := range newSteps {
			if t, n, sk := refs.Rewrite(newSteps[i].Content, idMap); n > 0 {
				newSteps[i].Content = t
				changed += n
				stepChanged = true
			} else {
				skipped += sk
			}
			if t, n, sk := refs.Rewrite(newSteps[i].Expected, idMap); n > 0 {
				newSteps[i].Expected = t
				changed += n
				stepChanged = true
			} else {
				skipped += sk
			}
		}
		if stepChanged {
			upd.CustomStepsSeparated = newSteps
		}
	}
	if changed == 0 {
		res.EntitiesSkipped++
		res.RefsSkipped += skipped
		return
	}
	if _, err := api.UpdateCase(ctx, e.EntityID, upd); err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "case", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	res.EntitiesRewritten++
	res.RefsRewritten += changed
	res.RefsSkipped += skipped
}

func rewriteOneRun(ctx context.Context, api ReferenceRewriteAPI, e *refs.EntityRefs, idMap map[int64]int64, res *ReferenceRewriteResult) {
	r, err := api.GetRun(ctx, e.EntityID)
	if err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "run", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	newText, n, sk := refs.Rewrite(r.Description, idMap)
	if n == 0 {
		res.EntitiesSkipped++
		res.RefsSkipped += sk
		return
	}
	upd := &data.UpdateRunRequest{Description: strPtr(newText)}
	if _, err := api.UpdateRun(ctx, e.EntityID, upd); err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "run", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	res.EntitiesRewritten++
	res.RefsRewritten += n
	res.RefsSkipped += sk
}

func rewriteOnePlan(ctx context.Context, api ReferenceRewriteAPI, e *refs.EntityRefs, idMap map[int64]int64, res *ReferenceRewriteResult) {
	p, err := api.GetPlan(ctx, e.EntityID)
	if err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "plan", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	newText, n, sk := refs.Rewrite(p.Description, idMap)
	if n == 0 {
		res.EntitiesSkipped++
		res.RefsSkipped += sk
		return
	}
	upd := &data.UpdatePlanRequest{Description: newText}
	if _, err := api.UpdatePlan(ctx, e.EntityID, upd); err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "plan", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	res.EntitiesRewritten++
	res.RefsRewritten += n
	res.RefsSkipped += sk
}

func rewriteOneMilestone(ctx context.Context, api ReferenceRewriteAPI, e *refs.EntityRefs, idMap map[int64]int64, res *ReferenceRewriteResult) {
	m, err := api.GetMilestone(ctx, e.EntityID)
	if err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "milestone", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	newText, n, sk := refs.Rewrite(m.Description, idMap)
	if n == 0 {
		res.EntitiesSkipped++
		res.RefsSkipped += sk
		return
	}
	upd := &data.UpdateMilestoneRequest{Description: newText}
	if _, err := api.UpdateMilestone(ctx, e.EntityID, upd); err != nil {
		res.EntitiesFailed++
		res.Failures = append(res.Failures, ReferenceRewriteFailure{EntityType: "milestone", EntityID: e.EntityID, Error: err.Error()})
		return
	}
	res.EntitiesRewritten++
	res.RefsRewritten += n
	res.RefsSkipped += sk
}

func strPtr(s string) *string { return &s }
