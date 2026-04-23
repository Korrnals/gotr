package migration

import (
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestVerifyCasesCoverage_FullMatch(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})
	m.compareField = "Title"
	m.mapping.AddPair(10, 10, "existing")

	src := data.GetCasesResponse{
		{ID: 1, Title: "Login", SectionID: 10},
		{ID: 2, Title: "Logout", SectionID: 10},
	}
	// Source section 10 is assumed to be mapped identically (dst suite 0).
	tgt := data.GetCasesResponse{
		{ID: 100, Title: "Login", SectionID: 10},
		{ID: 101, Title: "Logout", SectionID: 10},
	}

	rep := m.VerifyCasesCoverage(src, tgt)

	assert.Equal(t, 2, rep.Matched)
	assert.Empty(t, rep.Missing)
	assert.Equal(t, 1.0, rep.Coverage)
}

func TestVerifyCasesCoverage_MissingInTarget(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})
	m.compareField = "Title"
	m.mapping.AddPair(10, 10, "existing")

	src := data.GetCasesResponse{
		{ID: 1, Title: "Login", SectionID: 10},
		{ID: 2, Title: "Orphan", SectionID: 10},
	}
	tgt := data.GetCasesResponse{
		{ID: 100, Title: "Login", SectionID: 10},
	}

	rep := m.VerifyCasesCoverage(src, tgt)

	assert.Equal(t, 1, rep.Matched)
	if assert.Len(t, rep.Missing, 1) {
		assert.Equal(t, int64(2), rep.Missing[0].SrcID)
		assert.Equal(t, "Orphan", rep.Missing[0].Title)
	}
	assert.InDelta(t, 0.5, rep.Coverage, 1e-9)
}

func TestVerifyCasesCoverage_MultisetRespected(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})
	m.compareField = "Title"
	m.mapping.AddPair(10, 10, "existing")

	// Two source cases with same title: each consumes one target row.
	src := data.GetCasesResponse{
		{ID: 1, Title: "Dup", SectionID: 10},
		{ID: 2, Title: "Dup", SectionID: 10},
	}
	tgt := data.GetCasesResponse{
		{ID: 100, Title: "Dup", SectionID: 10},
	}

	rep := m.VerifyCasesCoverage(src, tgt)

	assert.Equal(t, 1, rep.Matched)
	assert.Len(t, rep.Missing, 1)
}
