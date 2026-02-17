// Package compare tests - comprehensive test suite for fetcher functionality
package compare

import (
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для compareSuitesInternal ====================

func TestCompareSuitesInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 1 {
				return []data.Suite{
					{ID: 1, Name: "Suite A"},
					{ID: 2, Name: "Suite B"},
				}, nil
			}
			return []data.Suite{
				{ID: 1, Name: "Suite A"},
				{ID: 3, Name: "Suite C"},
			}, nil
		},
	}

	result, err := compareSuitesInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "suites", result.Resource)
	assert.Equal(t, int64(1), result.Project1ID)
	assert.Equal(t, int64(2), result.Project2ID)
	// Suite A is common, Suite B only in first, Suite C only in second
	assert.Equal(t, 1, len(result.OnlyInFirst))  // Suite B
	assert.Equal(t, 1, len(result.OnlyInSecond)) // Suite C
	assert.Equal(t, 1, len(result.Common))       // Suite A
}

func TestCompareSuitesInternal_FirstError(t *testing.T) {
	// Using GetSuitesParallelFunc to test fault-tolerant behavior
	mock := &client.MockClient{
		GetSuitesParallelFunc: func(projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error) {
			results := make(map[int64]data.GetSuitesResponse)
			// Return partial results (project 2 only) with error
			results[2] = []data.Suite{}
			return results, errors.New("API error for project 1")
		},
	}

	result, err := compareSuitesInternal(mock, 1, 2, nil)

	// Parallel implementation is fault-tolerant: returns partial results with warning
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "suites", result.Resource)
}

func TestCompareSuitesInternal_SecondError(t *testing.T) {
	// Using GetSuitesParallelFunc to test fault-tolerant behavior
	mock := &client.MockClient{
		GetSuitesParallelFunc: func(projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error) {
			results := make(map[int64]data.GetSuitesResponse)
			// Return partial results (project 1 only) with error
			results[1] = []data.Suite{}
			return results, errors.New("API error for project 2")
		},
	}

	result, err := compareSuitesInternal(mock, 1, 2, nil)

	// Parallel implementation is fault-tolerant: returns partial results with warning
	// Only returns error if BOTH projects fail
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "suites", result.Resource)
}

func TestCompareSuitesInternal_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
	}

	result, err := compareSuitesInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.OnlyInFirst))
	assert.Equal(t, 0, len(result.OnlyInSecond))
	assert.Equal(t, 0, len(result.Common))
}

// ==================== Тесты для fetchSuiteItems ====================

func TestFetchSuiteItems_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{
				{ID: 1, Name: "Suite A"},
				{ID: 2, Name: "Suite B"},
			}, nil
		},
	}

	items, err := fetchSuiteItems(mock, 1)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(items))
	assert.Equal(t, int64(1), items[0].ID)
	assert.Equal(t, "Suite A", items[0].Name)
}

func TestFetchSuiteItems_Error(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return nil, errors.New("API error")
		},
	}

	items, err := fetchSuiteItems(mock, 1)

	assert.Error(t, err)
	assert.Nil(t, items)
}

func TestFetchSuiteItems_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
	}

	items, err := fetchSuiteItems(mock, 1)

	assert.NoError(t, err)
	assert.Empty(t, items)
}

// ==================== Тесты для compareCasesInternal ====================

func TestCompareCasesInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if projectID == 1 {
				return []data.Case{
					{ID: 1, Title: "Case A"},
					{ID: 2, Title: "Case B"},
				}, nil
			}
			return []data.Case{
				{ID: 1, Title: "Case A"},
				{ID: 3, Title: "Case C"},
			}, nil
		},
	}

	result, err := compareCasesInternal(mock, 1, 2, "title", nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "cases", result.Resource)
	assert.Equal(t, 1, len(result.OnlyInFirst))
	assert.Equal(t, 1, len(result.OnlyInSecond))
	assert.Equal(t, 1, len(result.Common))
}

func TestCompareCasesInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareCasesInternal(mock, 1, 2, "title", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для fetchCaseItems ====================

func TestFetchCaseItems_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{
				{ID: 1, Title: "Case A"},
				{ID: 2, Title: "Case B"},
			}, nil
		},
	}

	items, err := fetchCaseItems(mock, 1, nil)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(items))
	assert.Equal(t, int64(1), items[0].ID)
	assert.Equal(t, "Case A", items[0].Name)
}

func TestFetchCaseItems_Error(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return nil, errors.New("API error")
		},
	}

	items, err := fetchCaseItems(mock, 1, nil)

	assert.Error(t, err)
	assert.Nil(t, items)
}

// ==================== Тесты для compareSectionsInternal ====================

func TestCompareSectionsInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{{ID: 1, Name: "Suite 1"}}, nil
		},
		GetSectionsFunc: func(projectID, suiteID int64) (data.GetSectionsResponse, error) {
			if projectID == 1 {
				return []data.Section{
					{ID: 1, Name: "Section A", SuiteID: 1},
				}, nil
			}
			return []data.Section{
				{ID: 1, Name: "Section A", SuiteID: 1},
				{ID: 2, Name: "Section B", SuiteID: 1},
			}, nil
		},
	}

	result, err := compareSectionsInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "sections", result.Resource)
}

func TestCompareSectionsInternal_SuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareSectionsInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для comparePlansInternal ====================

func TestComparePlansInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetPlansFunc: func(projectID int64) (data.GetPlansResponse, error) {
			if projectID == 1 {
				return []data.Plan{
					{ID: 1, Name: "Plan A"},
				}, nil
			}
			return []data.Plan{
				{ID: 1, Name: "Plan A"},
				{ID: 2, Name: "Plan B"},
			}, nil
		},
	}

	result, err := comparePlansInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "plans", result.Resource)
}

func TestComparePlansInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetPlansFunc: func(projectID int64) (data.GetPlansResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := comparePlansInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareRunsInternal ====================

func TestCompareRunsInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			if projectID == 1 {
				return []data.Run{
					{ID: 1, Name: "Run A"},
				}, nil
			}
			return []data.Run{
				{ID: 1, Name: "Run A"},
				{ID: 2, Name: "Run B"},
			}, nil
		},
	}

	result, err := compareRunsInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "runs", result.Resource)
}

func TestCompareRunsInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareRunsInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareMilestonesInternal ====================

func TestCompareMilestonesInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			if projectID == 1 {
				return []data.Milestone{
					{ID: 1, Name: "Milestone A"},
				}, nil
			}
			return []data.Milestone{
				{ID: 1, Name: "Milestone A"},
				{ID: 2, Name: "Milestone B"},
			}, nil
		},
	}

	result, err := compareMilestonesInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "milestones", result.Resource)
}

func TestCompareMilestonesInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareMilestonesInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareDatasetsInternal ====================

func TestCompareDatasetsInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			if projectID == 1 {
				return []data.Dataset{
					{ID: 1, Name: "Dataset A"},
				}, nil
			}
			return []data.Dataset{
				{ID: 1, Name: "Dataset A"},
				{ID: 2, Name: "Dataset B"},
			}, nil
		},
	}

	result, err := compareDatasetsInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "datasets", result.Resource)
}

func TestCompareDatasetsInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareDatasetsInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareGroupsInternal ====================

func TestCompareGroupsInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			if projectID == 1 {
				return []data.Group{
					{ID: 1, Name: "Group A"},
				}, nil
			}
			return []data.Group{
				{ID: 1, Name: "Group A"},
				{ID: 2, Name: "Group B"},
			}, nil
		},
	}

	result, err := compareGroupsInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "groups", result.Resource)
}

func TestCompareGroupsInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareGroupsInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareLabelsInternal ====================

func TestCompareLabelsInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			if projectID == 1 {
				return []data.Label{
					{ID: 1, Name: "Label A"},
				}, nil
			}
			return []data.Label{
				{ID: 1, Name: "Label A"},
				{ID: 2, Name: "Label B"},
			}, nil
		},
	}

	result, err := compareLabelsInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "labels", result.Resource)
}

func TestCompareLabelsInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareLabelsInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareTemplatesInternal ====================

func TestCompareTemplatesInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
			if projectID == 1 {
				return []data.Template{
					{ID: 1, Name: "Template A"},
				}, nil
			}
			return []data.Template{
				{ID: 1, Name: "Template A"},
				{ID: 2, Name: "Template B"},
			}, nil
		},
	}

	result, err := compareTemplatesInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "templates", result.Resource)
}

func TestCompareTemplatesInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareTemplatesInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareConfigurationsInternal ====================

func TestCompareConfigurationsInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			if projectID == 1 {
				return []data.ConfigGroup{
					{ID: 1, Name: "Group A"},
				}, nil
			}
			return []data.ConfigGroup{
				{ID: 1, Name: "Group A"},
				{ID: 2, Name: "Group B"},
			}, nil
		},
	}

	result, err := compareConfigurationsInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "configurations", result.Resource)
}

func TestCompareConfigurationsInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareConfigurationsInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareSharedStepsInternal ====================

func TestCompareSharedStepsInternal_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			if projectID == 1 {
				return []data.SharedStep{
					{ID: 1, Title: "Step A"},
				}, nil
			}
			return []data.SharedStep{
				{ID: 1, Title: "Step A"},
				{ID: 2, Title: "Step B"},
			}, nil
		},
	}

	result, err := compareSharedStepsInternal(mock, 1, 2, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "sharedsteps", result.Resource)
}

func TestCompareSharedStepsInternal_Error(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, errors.New("API error")
		},
	}

	result, err := compareSharedStepsInternal(mock, 1, 2, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== Тесты для compareItemInfos ====================

func TestCompareItemInfos(t *testing.T) {
	items1 := []ItemInfo{
		{ID: 1, Name: "Item A"},
		{ID: 2, Name: "Item B"},
	}
	items2 := []ItemInfo{
		{ID: 1, Name: "Item A"},
		{ID: 3, Name: "Item C"},
	}

	result := compareItemInfos("test", 1, 2, items1, items2)

	assert.Equal(t, "test", result.Resource)
	assert.Equal(t, int64(1), result.Project1ID)
	assert.Equal(t, int64(2), result.Project2ID)
	assert.Equal(t, 1, len(result.OnlyInFirst))
	assert.Equal(t, 1, len(result.OnlyInSecond))
	assert.Equal(t, 1, len(result.Common))
	assert.Equal(t, int64(2), result.OnlyInFirst[0].ID)
	assert.Equal(t, int64(3), result.OnlyInSecond[0].ID)
	assert.Equal(t, "Item A", result.Common[0].Name)
	assert.True(t, result.Common[0].IDsMatch)
}

func TestCompareItemInfos_DifferentIDs(t *testing.T) {
	items1 := []ItemInfo{
		{ID: 1, Name: "Item A"},
	}
	items2 := []ItemInfo{
		{ID: 2, Name: "Item A"},
	}

	result := compareItemInfos("test", 1, 2, items1, items2)

	assert.Equal(t, 0, len(result.OnlyInFirst))
	assert.Equal(t, 0, len(result.OnlyInSecond))
	assert.Equal(t, 1, len(result.Common))
	assert.False(t, result.Common[0].IDsMatch)
	assert.Equal(t, int64(1), result.Common[0].ID1)
	assert.Equal(t, int64(2), result.Common[0].ID2)
}

func TestCompareItemInfos_Empty(t *testing.T) {
	items1 := []ItemInfo{}
	items2 := []ItemInfo{}

	result := compareItemInfos("test", 1, 2, items1, items2)

	assert.Equal(t, 0, len(result.OnlyInFirst))
	assert.Equal(t, 0, len(result.OnlyInSecond))
	assert.Equal(t, 0, len(result.Common))
}

// ==================== Тесты для getFieldValue ====================

func TestGetFieldValue(t *testing.T) {
	c := data.Case{
		ID:             1,
		Title:          "Test Case",
		PriorityID:     2,
		TypeID:         3,
		MilestoneID:    4,
		Refs:           "REF-123",
		CustomPreconds: "Preconditions",
		CustomSteps:    "Steps",
		CustomExpected: "Expected",
	}

	assert.Equal(t, "Test Case", getFieldValue(c, "title"))
	assert.Equal(t, "2", getFieldValue(c, "priority_id"))
	assert.Equal(t, "3", getFieldValue(c, "type_id"))
	assert.Equal(t, "4", getFieldValue(c, "milestone_id"))
	assert.Equal(t, "REF-123", getFieldValue(c, "refs"))
	assert.Equal(t, "Preconditions", getFieldValue(c, "custom_preconds"))
	assert.Equal(t, "Steps", getFieldValue(c, "custom_steps"))
	assert.Equal(t, "Expected", getFieldValue(c, "custom_expected"))
	assert.Equal(t, "<unknown field>", getFieldValue(c, "unknown"))
}

// ==================== Тесты для getCaseKey ====================

func TestGetCaseKey(t *testing.T) {
	item := ItemInfo{ID: 1, Name: "Test Case"}

	// getCaseKey returns item.Name for all fields (simplified implementation)
	assert.Equal(t, "Test Case", getCaseKey(item, "title"))
	assert.Equal(t, "Test Case", getCaseKey(item, "priority_id"))
	assert.Equal(t, "Test Case", getCaseKey(item, "unknown"))
}
