// internal/migration/filter.go
package migration

import (
	"github.com/Korrnals/gotr/internal/models/data"
)

// FilterStats holds summary statistics from the last filter operation.
//
// Duplicates is preserved as a legacy alias for AlreadyInTarget so existing
// callers and tests keep compiling. New code should prefer AlreadyInTarget
// (items matched against a target row and therefore skipped) and
// SrcCollapsed (items whose MatchKey repeated within the source batch — a
// diagnostic signal that source contains duplicates which will land as new
// rows in target).
type FilterStats struct {
	Source          int // total items in source
	Target          int // total items in target
	AlreadyInTarget int // source items matched to a target row via MatchKey (skipped)
	Duplicates      int // alias of AlreadyInTarget (kept for backward compatibility)
	SrcCollapsed    int // source items sharing a MatchKey with another source item
	New             int // items ready for import
}

// LastFilterStats returns the statistics from the last Filter* call.
func (m *Migration) LastFilterStats() FilterStats {
	return m.lastFilterStats
}

// recordFilterStats stores stats and keeps Duplicates mirrored to
// AlreadyInTarget for legacy consumers.
func (m *Migration) recordFilterStats(source, target, alreadyInTarget, srcCollapsed, newCount int) {
	m.lastFilterStats = FilterStats{
		Source:          source,
		Target:          target,
		AlreadyInTarget: alreadyInTarget,
		Duplicates:      alreadyInTarget,
		SrcCollapsed:    srcCollapsed,
		New:             newCount,
	}
}

// countSrcCollapsed reports how many source entries share a MatchKey with a
// previously-seen source entry (i.e. |source| - |unique keys|).
func countSrcCollapsed(keys []MatchKey) int {
	seen := make(map[MatchKey]struct{}, len(keys))
	collapsed := 0
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			collapsed++
			continue
		}
		seen[k] = struct{}{}
	}
	return collapsed
}

// FilterSharedSteps filters shared steps by usage in suite and duplicates in target.
// Candidates are shared steps whose IDs appear in usedStepIDs (collected from case steps).
// Duplicates (matched against a target shared step by MatchKey) are added to
// the mapping with status "existing". New (non-duplicate) steps are returned
// for import.
func (m *Migration) FilterSharedSteps(source, target data.GetSharedStepsResponse, usedStepIDs map[int64]struct{}) (filtered data.GetSharedStepsResponse, err error) {
	m.logger.Info("Starting shared steps filtering by usage in suite")

	var candidates data.GetSharedStepsResponse
	for _, step := range source {
		if _, ok := usedStepIDs[step.ID]; ok {
			candidates = append(candidates, step)
		}
	}
	m.logger.Infow("Found candidates for transfer (used in suite)", "count", len(candidates))

	m.logger.Info("Checking for duplicates in target project")
	bucket := NewBucket()
	for _, t := range target {
		val := fieldValue(t, m.compareField)
		if val == "" {
			continue
		}
		bucket.Add(sharedStepMatchKey(val, t.CustomStepsSeparated), t.ID)
	}

	candidateKeys := make([]MatchKey, 0, len(candidates))
	for _, step := range candidates {
		val := fieldValue(step, m.compareField)
		key := sharedStepMatchKey(val, step.CustomStepsSeparated)
		candidateKeys = append(candidateKeys, key)
		if existingID, ok := bucket.ConsumeOne(key); ok {
			m.mapping.AddPair(step.ID, existingID, "existing")
			m.logger.Debugw("Duplicate shared step matched — added to mapping", "title", step.Title, "old_id", step.ID, "existing_id", existingID)
			continue
		}
		filtered = append(filtered, step)
	}

	alreadyInTarget := len(candidates) - len(filtered)
	srcCollapsed := countSrcCollapsed(candidateKeys)
	m.recordFilterStats(len(source), len(target), alreadyInTarget, srcCollapsed, len(filtered))
	m.logger.Infow("Ready to import new shared steps", "count", len(filtered), "already_in_target", alreadyInTarget, "src_collapsed", srcCollapsed)
	return filtered, nil
}

// FilterSuites filters suites by duplicate detection (by name) using a Bucket.
// Duplicates are added to the mapping with status "existing". New suites are
// returned for import.
func (m *Migration) FilterSuites(source, target data.GetSuitesResponse) (filtered data.GetSuitesResponse, err error) {
	m.logger.Info("Starting suites filtering by duplicates (by name)")

	bucket := NewBucket()
	for _, t := range target {
		if t.Name == "" {
			continue
		}
		bucket.Add(suiteMatchKey(t.Name), t.ID)
	}

	srcKeys := make([]MatchKey, 0, len(source))
	for _, s := range source {
		key := suiteMatchKey(s.Name)
		srcKeys = append(srcKeys, key)
		if existingID, ok := bucket.ConsumeOne(key); ok {
			m.mapping.AddPair(s.ID, existingID, "existing")
			m.logger.Debugw("Duplicate suite matched — added to mapping", "name", s.Name, "old_id", s.ID, "existing_id", existingID)
			continue
		}
		filtered = append(filtered, s)
	}

	alreadyInTarget := len(source) - len(filtered)
	srcCollapsed := countSrcCollapsed(srcKeys)
	m.recordFilterStats(len(source), len(target), alreadyInTarget, srcCollapsed, len(filtered))
	m.logger.Infow("Ready to import new suites", "count", len(filtered), "already_in_target", alreadyInTarget, "src_collapsed", srcCollapsed)
	return filtered, nil
}

// FilterCases filters cases using a (dst_section_id, compareField) MatchKey
// with multiset semantics. This prevents the historical "collapse" bug where
// N source cases sharing a title were all skipped because a single flat map
// entry covered them; now each target case can match exactly one source case.
func (m *Migration) FilterCases(source, target data.GetCasesResponse) (filtered data.GetCasesResponse, err error) {
	m.logger.Info("Starting cases filtering by (dst_section_id, compare_field)")

	bucket := NewBucket()
	for _, t := range target {
		val := fieldValue(t, m.compareField)
		if val == "" {
			continue
		}
		bucket.Add(caseMatchKey(t.SectionID, val), t.ID)
	}

	srcKeys := make([]MatchKey, 0, len(source))
	for _, c := range source {
		val := fieldValue(c, m.compareField)
		dstSectionID := m.resolveDstSectionIDForFilter(c.SectionID)
		key := caseMatchKey(dstSectionID, val)
		srcKeys = append(srcKeys, key)
		if _, ok := bucket.ConsumeOne(key); ok {
			m.logger.Debugw("Duplicate case matched — skipped", "title", c.Title, "src_section_id", c.SectionID, "dst_section_id", dstSectionID)
			continue
		}
		filtered = append(filtered, c)
	}

	alreadyInTarget := len(source) - len(filtered)
	srcCollapsed := countSrcCollapsed(srcKeys)
	m.recordFilterStats(len(source), len(target), alreadyInTarget, srcCollapsed, len(filtered))
	m.logger.Infow("Ready to import new cases", "count", len(filtered), "already_in_target", alreadyInTarget, "src_collapsed", srcCollapsed)
	return filtered, nil
}

// FilterSections filters sections by (dst_parent_id, name). At filter time
// the parent mapping may be partially populated (deep sections' parents get
// mapped during ImportSections, which runs after this call). When a parent
// cannot be resolved we fall back to a negative sentinel so unmapped source
// sections are treated as new and never incorrectly matched.
func (m *Migration) FilterSections(source, target data.GetSectionsResponse) (filtered data.GetSectionsResponse, err error) {
	m.logger.Info("Starting sections filtering by (dst_parent_id, name)")

	bucket := NewBucket()
	for _, t := range target {
		if t.Name == "" {
			continue
		}
		bucket.Add(sectionMatchKey(t.ParentID, t.Name), t.ID)
	}

	srcKeys := make([]MatchKey, 0, len(source))
	for _, s := range source {
		var dstParentID int64
		switch {
		case s.ParentID == 0:
			dstParentID = 0
		default:
			if mapped, ok := m.mapping.GetTargetBySource(s.ParentID); ok {
				dstParentID = mapped
			} else {
				dstParentID = -s.ParentID
			}
		}
		key := sectionMatchKey(dstParentID, s.Name)
		srcKeys = append(srcKeys, key)
		if existingID, ok := bucket.ConsumeOne(key); ok {
			m.mapping.AddPair(s.ID, existingID, "existing")
			m.logger.Debugw("Duplicate section matched — mapping added", "name", s.Name, "old_id", s.ID, "existing_id", existingID, "dst_parent_id", dstParentID)
			continue
		}
		filtered = append(filtered, s)
	}

	alreadyInTarget := len(source) - len(filtered)
	srcCollapsed := countSrcCollapsed(srcKeys)
	m.recordFilterStats(len(source), len(target), alreadyInTarget, srcCollapsed, len(filtered))
	m.logger.Infow("Ready to import new sections", "count", len(filtered), "already_in_target", alreadyInTarget, "src_collapsed", srcCollapsed)
	return filtered, nil
}
