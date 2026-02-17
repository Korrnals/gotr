// Package compare tests - comprehensive test suite for compare functionality
package compare

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для parseFlags ====================

func TestParseFlags_Success(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("pid1", "", "")
	cmd.Flags().String("pid2", "", "")
	cmd.Flags().String("field", "title", "")
	cmd.ParseFlags([]string{"--pid1=30", "--pid2=31", "--field=priority_id"})

	pid1, pid2, field, err := parseFlags(cmd)

	assert.NoError(t, err)
	assert.Equal(t, int64(30), pid1)
	assert.Equal(t, int64(31), pid2)
	assert.Equal(t, "priority_id", field)
}

func TestParseFlags_DefaultField(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("pid1", "", "")
	cmd.Flags().String("pid2", "", "")
	cmd.Flags().String("field", "title", "")
	cmd.ParseFlags([]string{"--pid1=30", "--pid2=31"})

	pid1, pid2, field, err := parseFlags(cmd)

	assert.NoError(t, err)
	assert.Equal(t, "title", field)
	assert.Equal(t, int64(30), pid1)
	assert.Equal(t, int64(31), pid2)
}

func TestParseFlags_InvalidPid1(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("pid1", "", "")
	cmd.Flags().String("pid2", "", "")
	cmd.Flags().String("field", "title", "")
	cmd.ParseFlags([]string{"--pid1=invalid", "--pid2=31"})

	_, _, _, err := parseFlags(cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pid1")
}

// ==================== Тесты для buildResourceDiff ====================

func TestBuildResourceDiff(t *testing.T) {
	first := []string{"A", "B", "C"}
	second := []string{"B", "C", "D"}

	diff := buildResourceDiff("test", first, second)

	assert.Equal(t, "test", diff.Resource)
	assert.Equal(t, 3, diff.TotalFirst)
	assert.Equal(t, 3, diff.TotalSecond)
	assert.Equal(t, 2, len(diff.Common))     // B, C
	assert.Equal(t, 1, len(diff.OnlyFirst))  // A
	assert.Equal(t, 1, len(diff.OnlySecond)) // D
}

func TestBuildResourceDiff_EmptyFirst(t *testing.T) {
	first := []string{}
	second := []string{"A", "B", "C"}

	diff := buildResourceDiff("test", first, second)

	assert.Equal(t, 0, diff.TotalFirst)
	assert.Equal(t, 3, diff.TotalSecond)
	assert.Equal(t, 0, len(diff.Common))
	assert.Equal(t, 0, len(diff.OnlyFirst))
	assert.Equal(t, 3, len(diff.OnlySecond))
}

func TestBuildResourceDiff_BothEmpty(t *testing.T) {
	first := []string{}
	second := []string{}

	diff := buildResourceDiff("test", first, second)

	assert.Equal(t, 0, diff.TotalFirst)
	assert.Equal(t, 0, diff.TotalSecond)
	assert.Equal(t, 0, len(diff.Common))
}

// ==================== Тесты для collectNames ====================

func TestCollectNames(t *testing.T) {
	names := collectNames(3, func(i int) string {
		return []string{"A", "", "B"}[i]
	})

	assert.Equal(t, 2, len(names))
	assert.Contains(t, names, "A")
	assert.Contains(t, names, "B")
}

func TestCollectNames_AllEmpty(t *testing.T) {
	names := collectNames(3, func(i int) string { return "" })
	assert.Equal(t, 0, len(names))
}

func TestCollectNames_ZeroSize(t *testing.T) {
	names := collectNames(0, func(i int) string { return "A" })
	assert.Nil(t, names)
}

// ==================== Тесты для GetProjectNames ====================

func TestGetProjectNames_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{
				ID:   projectID,
				Name: "Test Project " + string(rune('0'+projectID)),
			}, nil
		},
	}

	name1, name2, err := GetProjectNames(mock, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, "Test Project 1", name1)
	assert.Equal(t, "Test Project 2", name2)
}

func TestGetProjectName_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{
				ID:   projectID,
				Name: "Test Project",
			}, nil
		},
	}

	name, err := GetProjectName(mock, 1)
	assert.NoError(t, err)
	assert.Equal(t, "Test Project", name)
}

// ==================== Тесты для CompareResult ====================

func TestCompareResult_Struct(t *testing.T) {
	result := &CompareResult{
		Resource:     "cases",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Case 1"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Case 2"}},
		Common:       []CommonItemInfo{{Name: "Common", ID1: 3, ID2: 4, IDsMatch: false}},
	}

	assert.Equal(t, "cases", result.Resource)
	assert.Equal(t, int64(1), result.Project1ID)
	assert.Equal(t, int64(2), result.Project2ID)
	assert.Equal(t, 1, len(result.OnlyInFirst))
	assert.Equal(t, 1, len(result.OnlyInSecond))
	assert.Equal(t, 1, len(result.Common))
}

func TestItemInfo(t *testing.T) {
	item := ItemInfo{ID: 1, Name: "Test"}
	assert.Equal(t, int64(1), item.ID)
	assert.Equal(t, "Test", item.Name)
}

func TestCommonItemInfo(t *testing.T) {
	item := CommonItemInfo{Name: "Test", ID1: 1, ID2: 2, IDsMatch: false}
	assert.Equal(t, "Test", item.Name)
	assert.Equal(t, int64(1), item.ID1)
	assert.Equal(t, int64(2), item.ID2)
	assert.False(t, item.IDsMatch)
}

func TestIDMappingPair(t *testing.T) {
	pair := IDMappingPair{ID1: 1, ID2: 2, Name: "Test"}
	assert.Equal(t, int64(1), pair.ID1)
	assert.Equal(t, int64(2), pair.ID2)
	assert.Equal(t, "Test", pair.Name)
}

// ==================== Тесты для ResourceDiff ====================

func TestResourceDiff_Struct(t *testing.T) {
	diff := ResourceDiff{
		Resource:    "suites",
		TotalFirst:  5,
		TotalSecond: 7,
		Common:      []string{"A", "B"},
		OnlyFirst:   []string{"C"},
		OnlySecond:  []string{"D", "E"},
	}

	assert.Equal(t, "suites", diff.Resource)
	assert.Equal(t, 5, diff.TotalFirst)
	assert.Equal(t, 7, diff.TotalSecond)
	assert.Equal(t, 2, len(diff.Common))
	assert.Equal(t, 1, len(diff.OnlyFirst))
	assert.Equal(t, 2, len(diff.OnlySecond))
}
