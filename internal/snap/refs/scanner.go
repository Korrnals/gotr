// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package refs

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// ScanCase walks every markdown-bearing field of a Case and returns
// the aggregated EntityRefs (or nil when no references were found).
func ScanCase(c *data.Case) *EntityRefs {
	if c == nil {
		return nil
	}
	var refs []Reference
	refs = append(refs, ScanText(c.CustomPreconds, "custom_preconds")...)
	refs = append(refs, ScanText(c.CustomSteps, "custom_steps")...)
	refs = append(refs, ScanText(c.CustomExpected, "custom_expected")...)
	refs = append(refs, ScanText(c.CustomMission, "custom_mission")...)
	refs = append(refs, ScanText(c.CustomGoals, "custom_goals")...)
	refs = append(refs, ScanText(c.Refs, "refs")...)
	for i, step := range c.CustomStepsSeparated {
		refs = append(refs, ScanText(step.Content, fmt.Sprintf("custom_steps_separated[%d].content", i))...)
		refs = append(refs, ScanText(step.Expected, fmt.Sprintf("custom_steps_separated[%d].expected", i))...)
		refs = append(refs, ScanText(step.AdditionalInfo, fmt.Sprintf("custom_steps_separated[%d].additional_info", i))...)
	}
	if len(refs) == 0 {
		return nil
	}
	return &EntityRefs{EntityType: "case", EntityID: c.ID, Refs: refs}
}

// ScanResult walks the comment field of a Result.
//
// TestRail's API does not expose an update_result endpoint — results
// are immutable additions — so references found here are recorded for
// audit but cannot be rewritten on restore. The restore phase is
// expected to flag them in the report.
func ScanResult(r *data.Result) *EntityRefs {
	if r == nil {
		return nil
	}
	refs := ScanText(r.Comment, "comment")
	if len(refs) == 0 {
		return nil
	}
	return &EntityRefs{EntityType: "result", EntityID: r.ID, Refs: refs}
}

// ScanRun walks the description field of a Run. TestRail Runs do not
// carry a comment field; refs are attached to the cases under the run.
func ScanRun(r *data.Run) *EntityRefs {
	if r == nil {
		return nil
	}
	refs := ScanText(r.Description, "description")
	if len(refs) == 0 {
		return nil
	}
	return &EntityRefs{EntityType: "run", EntityID: r.ID, Refs: refs}
}

// ScanPlan walks the description field of a Plan.
func ScanPlan(p *data.Plan) *EntityRefs {
	if p == nil {
		return nil
	}
	refs := ScanText(p.Description, "description")
	if len(refs) == 0 {
		return nil
	}
	return &EntityRefs{EntityType: "plan", EntityID: p.ID, Refs: refs}
}

// ScanMilestone walks the description field of a Milestone.
func ScanMilestone(m *data.Milestone) *EntityRefs {
	if m == nil {
		return nil
	}
	refs := ScanText(m.Description, "description")
	if len(refs) == 0 {
		return nil
	}
	return &EntityRefs{EntityType: "milestone", EntityID: m.ID, Refs: refs}
}
