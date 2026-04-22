// internal/service/migration/coverage.go
package migration

import (
	"github.com/Korrnals/gotr/internal/models/data"
)

// CoverageMissing describes one source case that has no matching row in
// target after the import, as determined by the same MatchKey logic used by
// FilterCases (scope = dst_section_id, name = compareField value).
type CoverageMissing struct {
	SrcID        int64
	Title        string
	SrcSectionID int64
	DstSectionID int64
	MatchValue   string
}

// CoverageReport summarizes post-import coverage for cases.
type CoverageReport struct {
	Source   int
	Target   int
	Matched  int
	Missing  []CoverageMissing
	Coverage float64 // Matched / Source, 0..1 (0 when Source == 0)
}

// VerifyCasesCoverage checks, for every source case, that at least one target
// case exists with the same (dst_section_id, compareField) MatchKey. Matching
// is multiset: one source case consumes one target case. Any source case that
// cannot be matched is appended to the returned CoverageReport.Missing list.
//
// The destination-section mapping used for scope resolution is taken from
// m.mapping (populated during ImportSections). When a source section id has
// no mapping and dstSuite is 0, the source section id is kept so target rows
// in the same source scope still match; otherwise the negative sentinel is
// used so unmapped scopes can never collide with real target ids.
func (m *Migration) VerifyCasesCoverage(source, target data.GetCasesResponse) CoverageReport {
	rep := CoverageReport{Source: len(source), Target: len(target)}

	bucket := NewBucket()
	for _, t := range target {
		val := fieldValue(t, m.compareField)
		if val == "" {
			continue
		}
		bucket.Add(caseMatchKey(t.SectionID, val), t.ID)
	}

	for _, c := range source {
		val := fieldValue(c, m.compareField)
		dstSectionID := m.resolveDstSectionIDForFilter(c.SectionID)
		if _, ok := bucket.ConsumeOne(caseMatchKey(dstSectionID, val)); ok {
			rep.Matched++
			continue
		}
		rep.Missing = append(rep.Missing, CoverageMissing{
			SrcID:        c.ID,
			Title:        c.Title,
			SrcSectionID: c.SectionID,
			DstSectionID: dstSectionID,
			MatchValue:   val,
		})
	}

	if rep.Source > 0 {
		rep.Coverage = float64(rep.Matched) / float64(rep.Source)
	}
	return rep
}
