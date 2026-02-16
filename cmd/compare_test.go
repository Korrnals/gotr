package cmd

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestBuildResourceDiff(t *testing.T) {
	first := []string{"A", "B", "C"}
	second := []string{"B", "C", "D"}

	diff := buildResourceDiff("test", first, second)

	assert.Equal(t, "test", diff.resource)
	assert.Equal(t, 3, diff.totalFirst)
	assert.Equal(t, 3, diff.totalSecond)
	assert.Equal(t, 2, len(diff.common))    // B, C
	assert.Equal(t, 1, len(diff.onlyFirst)) // A
	assert.Equal(t, 1, len(diff.onlySecond)) // D
}

func TestToNameSet(t *testing.T) {
	names := []string{"Test", "test", "  Test  ", "Another"}
	set := toNameSet(names)

	// Should normalize and deduplicate
	assert.Equal(t, 2, len(set))
	assert.Equal(t, "Test", set["test"]) // First occurrence preserved
	assert.Equal(t, "Another", set["another"])
}

func TestCollectNames(t *testing.T) {
	names := collectNames(3, func(i int) string {
		return []string{"A", "", "B"}[i]
	})

	assert.Equal(t, 2, len(names))
	assert.Contains(t, names, "A")
	assert.Contains(t, names, "B")
}

func TestFormatSectionName(t *testing.T) {
	suiteNames := map[int64]string{1: "Suite 1"}

	result := formatSectionName(1, "Section A", suiteNames)
	assert.Equal(t, "Suite 1 / Section A", result)

	result = formatSectionName(0, "Section B", suiteNames)
	assert.Equal(t, "suite:default / Section B", result)

	result = formatSectionName(999, "Section C", suiteNames)
	assert.Equal(t, "suite:999 / Section C", result)
}

func TestNormalizeName(t *testing.T) {
	assert.Equal(t, "test", normalizeName("Test"))
	assert.Equal(t, "test", normalizeName("  TEST  "))
	assert.Equal(t, "", normalizeName(""))
}

func TestNamesFromKeys(t *testing.T) {
	lookup := map[string]string{
		"key1": "Value 1",
		"key2": "Value 2",
	}
	keys := []string{"key2", "key1"}

	result := namesFromKeys(keys, lookup)
	assert.Equal(t, []string{"Value 1", "Value 2"}, result) // Sorted alphabetically
}

func TestFetchSuiteNames(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return []data.Suite{{ID: 1, Name: "Suite 1"}}, nil
			}
			return []data.Suite{{ID: 1, Name: "Suite 1"}, {ID: 2, Name: "Suite 2"}}, nil
		},
	}

	names, err := fetchSuiteNames(mock, 1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Suite 1"}, names)

	names, err = fetchSuiteNames(mock, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(names))
}

func TestFetchCaseNames(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{
				{ID: 1, Title: "Case 1"},
				{ID: 2, Title: "Case 2"},
			}, nil
		},
	}

	names, err := fetchCaseNames(mock, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(names))
	assert.Contains(t, names, "Case 1")
	assert.Contains(t, names, "Case 2")
}

func TestCompareNamedResource(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return []data.Suite{{ID: 1, Name: "Suite A"}}, nil
			}
			return []data.Suite{{ID: 1, Name: "Suite A"}, {ID: 2, Name: "Suite B"}}, nil
		},
	}

	err := compareNamedResource(mock, 1, 2, "suites", fetchSuiteNames)
	assert.NoError(t, err)
}

func TestCompareResourceFetcherError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return nil, assert.AnError
		},
	}

	err := compareNamedResource(mock, 1, 2, "suites", fetchSuiteNames)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suites")
}
