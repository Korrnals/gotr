package migration

import (
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

type filterTestEntity struct {
	Title string
	Name  string
}

func TestFieldValue(t *testing.T) {
	t.Run("direct field match", func(t *testing.T) {
		obj := filterTestEntity{Title: "Case A"}

		value := fieldValue(obj, "Title")

		assert.Equal(t, "Case A", value)
	})

	t.Run("case-insensitive match", func(t *testing.T) {
		obj := filterTestEntity{Name: "Smoke"}

		value := fieldValue(obj, "name")

		assert.Equal(t, "Smoke", value)
	})

	t.Run("pointer input", func(t *testing.T) {
		obj := &filterTestEntity{Title: "Pointer Case"}

		value := fieldValue(obj, "Title")

		assert.Equal(t, "Pointer Case", value)
	})

	t.Run("nil input", func(t *testing.T) {
		value := fieldValue(nil, "Title")

		assert.Equal(t, "", value)
	})

	t.Run("non-existent field", func(t *testing.T) {
		obj := filterTestEntity{Title: "Case B", Name: "Regression"}

		value := fieldValue(obj, "UnknownField")

		assert.Equal(t, "", value)
	})
}

func TestFilterSuites_TableDriven(t *testing.T) {
	testCases := []struct {
		name                 string
		source               data.GetSuitesResponse
		target               data.GetSuitesResponse
		wantFilteredNames    []string
		wantMappingSourceIDs []int64
		wantMappingTargetIDs []int64
	}{
		{
			name:                 "nil source and target",
			source:               nil,
			target:               nil,
			wantFilteredNames:    []string{},
			wantMappingSourceIDs: []int64{},
			wantMappingTargetIDs: []int64{},
		},
		{
			name:                 "empty source",
			source:               data.GetSuitesResponse{},
			target:               data.GetSuitesResponse{{ID: 20, Name: "Smoke"}},
			wantFilteredNames:    []string{},
			wantMappingSourceIDs: []int64{},
			wantMappingTargetIDs: []int64{},
		},
		{
			name: "exclude duplicate and include new",
			source: data.GetSuitesResponse{
				{ID: 1, Name: "Smoke"},
				{ID: 2, Name: "API"},
				{ID: 3, Name: ""},
			},
			target: data.GetSuitesResponse{
				{ID: 100, Name: "Smoke"},
				{ID: 200, Name: ""},
			},
			wantFilteredNames:    []string{"API", ""},
			wantMappingSourceIDs: []int64{1},
			wantMappingTargetIDs: []int64{100},
		},
		{
			name: "target with empty names is ignored for matching",
			source: data.GetSuitesResponse{
				{ID: 11, Name: ""},
				{ID: 12, Name: "Unique"},
			},
			target: data.GetSuitesResponse{
				{ID: 222, Name: ""},
			},
			wantFilteredNames:    []string{"", "Unique"},
			wantMappingSourceIDs: []int64{},
			wantMappingTargetIDs: []int64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := setupTestMigration(t, &MockClient{})

			filtered, err := m.FilterSuites(tc.source, tc.target)
			assert.NoError(t, err)

			filteredNames := make([]string, 0, len(filtered))
			for _, s := range filtered {
				filteredNames = append(filteredNames, s.Name)
			}
			assert.Equal(t, tc.wantFilteredNames, filteredNames)
			assert.Equal(t, len(tc.wantMappingSourceIDs), m.mapping.Count)

			for i := range tc.wantMappingSourceIDs {
				got, ok := m.mapping.GetTargetBySource(tc.wantMappingSourceIDs[i])
				assert.True(t, ok)
				assert.Equal(t, tc.wantMappingTargetIDs[i], got)
			}
		})
	}
}

func TestFilterSections_TableDriven(t *testing.T) {
	testCases := []struct {
		name                 string
		source               data.GetSectionsResponse
		target               data.GetSectionsResponse
		wantFilteredNames    []string
		wantMappingSourceIDs []int64
		wantMappingTargetIDs []int64
	}{
		{
			name:                 "nil source and target",
			source:               nil,
			target:               nil,
			wantFilteredNames:    []string{},
			wantMappingSourceIDs: []int64{},
			wantMappingTargetIDs: []int64{},
		},
		{
			name:                 "empty source",
			source:               data.GetSectionsResponse{},
			target:               data.GetSectionsResponse{{ID: 30, Name: "Backend"}},
			wantFilteredNames:    []string{},
			wantMappingSourceIDs: []int64{},
			wantMappingTargetIDs: []int64{},
		},
		{
			name: "exclude duplicate and include new",
			source: data.GetSectionsResponse{
				{ID: 4, Name: "Backend"},
				{ID: 5, Name: "Frontend"},
				{ID: 6, Name: ""},
			},
			target: data.GetSectionsResponse{
				{ID: 400, Name: "Backend"},
				{ID: 401, Name: ""},
			},
			wantFilteredNames:    []string{"Frontend", ""},
			wantMappingSourceIDs: []int64{4},
			wantMappingTargetIDs: []int64{400},
		},
		{
			name: "target with empty names is ignored for matching",
			source: data.GetSectionsResponse{
				{ID: 21, Name: ""},
				{ID: 22, Name: "Section A"},
			},
			target: data.GetSectionsResponse{
				{ID: 333, Name: ""},
			},
			wantFilteredNames:    []string{"", "Section A"},
			wantMappingSourceIDs: []int64{},
			wantMappingTargetIDs: []int64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := setupTestMigration(t, &MockClient{})

			filtered, err := m.FilterSections(tc.source, tc.target)
			assert.NoError(t, err)

			filteredNames := make([]string, 0, len(filtered))
			for _, s := range filtered {
				filteredNames = append(filteredNames, s.Name)
			}
			assert.Equal(t, tc.wantFilteredNames, filteredNames)
			assert.Equal(t, len(tc.wantMappingSourceIDs), m.mapping.Count)

			for i := range tc.wantMappingSourceIDs {
				got, ok := m.mapping.GetTargetBySource(tc.wantMappingSourceIDs[i])
				assert.True(t, ok)
				assert.Equal(t, tc.wantMappingTargetIDs[i], got)
			}
		})
	}
}

func TestFilterCases_TableDriven(t *testing.T) {
	testCases := []struct {
		name              string
		compareField      string
		source            data.GetCasesResponse
		target            data.GetCasesResponse
		wantFilteredTitles []string
	}{
		{
			name:               "empty source",
			compareField:       "Title",
			source:             data.GetCasesResponse{},
			target:             data.GetCasesResponse{{ID: 10, Title: "A"}},
			wantFilteredTitles: []string{},
		},
		{
			name:         "exclude duplicates by compare field",
			compareField: "Title",
			source: data.GetCasesResponse{
				{ID: 1, Title: "Duplicate"},
				{ID: 2, Title: "Unique"},
			},
			target: data.GetCasesResponse{
				{ID: 100, Title: "Duplicate"},
			},
			wantFilteredTitles: []string{"Unique"},
		},
		{
			name:         "missing compare field keeps source cases",
			compareField: "UnknownField",
			source: data.GetCasesResponse{
				{ID: 3, Title: "One"},
				{ID: 4, Title: "Two"},
			},
			target: data.GetCasesResponse{
				{ID: 200, Title: "Target"},
			},
			wantFilteredTitles: []string{"One", "Two"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := setupTestMigration(t, &MockClient{})
			m.compareField = tc.compareField

			filtered, err := m.FilterCases(tc.source, tc.target)
			assert.NoError(t, err)

			filteredTitles := make([]string, 0, len(filtered))
			for _, c := range filtered {
				filteredTitles = append(filteredTitles, c.Title)
			}
			assert.Equal(t, tc.wantFilteredTitles, filteredTitles)
		})
	}
}

func TestFilterSharedSteps_TableDriven(t *testing.T) {
	testCases := []struct {
		name                string
		compareField        string
		source              data.GetSharedStepsResponse
		target              data.GetSharedStepsResponse
		sourceCaseIDs       map[int64]struct{}
		wantFilteredTitles  []string
		wantMappingSourceID int64
		wantMappingTargetID int64
		wantMappingExists   bool
	}{
		{
			name:         "used by source suite is excluded from candidates",
			compareField: "Title",
			source: data.GetSharedStepsResponse{
				{ID: 1, Title: "Used", CaseIDs: []int64{10}},
				{ID: 2, Title: "Unused", CaseIDs: []int64{20}},
			},
			target: data.GetSharedStepsResponse{},
			sourceCaseIDs: map[int64]struct{}{
				10: {},
			},
			wantFilteredTitles: []string{"Unused"},
			wantMappingExists:  false,
		},
		{
			name:         "duplicate candidate is mapped as existing",
			compareField: "Title",
			source: data.GetSharedStepsResponse{
				{ID: 3, Title: "Duplicate", CaseIDs: []int64{99}},
				{ID: 4, Title: "New", CaseIDs: []int64{100}},
			},
			target: data.GetSharedStepsResponse{
				{ID: 300, Title: "Duplicate"},
			},
			sourceCaseIDs:       map[int64]struct{}{},
			wantFilteredTitles:  []string{"New"},
			wantMappingSourceID: 3,
			wantMappingTargetID: 300,
			wantMappingExists:   true,
		},
		{
			name:         "compare field fallback keeps candidates without field match",
			compareField: "UnknownField",
			source: data.GetSharedStepsResponse{
				{ID: 5, Title: "A", CaseIDs: []int64{}},
				{ID: 6, Title: "B", CaseIDs: []int64{}},
			},
			target: data.GetSharedStepsResponse{
				{ID: 600, Title: "A"},
			},
			sourceCaseIDs:      map[int64]struct{}{},
			wantFilteredTitles: []string{"A", "B"},
			wantMappingExists:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := setupTestMigration(t, &MockClient{})
			m.compareField = tc.compareField

			filtered, err := m.FilterSharedSteps(tc.source, tc.target, tc.sourceCaseIDs)
			assert.NoError(t, err)

			filteredTitles := make([]string, 0, len(filtered))
			for _, s := range filtered {
				filteredTitles = append(filteredTitles, s.Title)
			}
			assert.Equal(t, tc.wantFilteredTitles, filteredTitles)

			if tc.wantMappingExists {
				got, ok := m.mapping.GetTargetBySource(tc.wantMappingSourceID)
				assert.True(t, ok)
				assert.Equal(t, tc.wantMappingTargetID, got)
			} else {
				assert.Equal(t, 0, m.mapping.Count)
			}
		})
	}
}
