// Package compare tests - comprehensive test suite for compare functionality
package compare

import (
	"context"
	"errors"
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

func TestParseFlags_InvalidPid2(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("pid1", "", "")
	cmd.Flags().String("pid2", "", "")
	cmd.Flags().String("field", "title", "")
	cmd.ParseFlags([]string{"--pid1=30", "--pid2=invalid"})

	_, _, _, err := parseFlags(cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pid2")
}

func TestParseFlags_EmptyFieldDefaultsToTitle(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("pid1", "", "")
	cmd.Flags().String("pid2", "", "")
	cmd.Flags().String("field", "", "")
	cmd.ParseFlags([]string{"--pid1=30", "--pid2=31"})

	pid1, pid2, field, err := parseFlags(cmd)

	assert.NoError(t, err)
	assert.Equal(t, int64(30), pid1)
	assert.Equal(t, int64(31), pid2)
	assert.Equal(t, "title", field)
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
	ctx := context.Background()
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{
				ID:   projectID,
				Name: "Test Project " + string(rune('0'+projectID)),
			}, nil
		},
	}

	name1, name2, err := GetProjectNames(ctx, mock, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, "Test Project 1", name1)
	assert.Equal(t, "Test Project 2", name2)
}

func TestGetProjectNames_FirstProjectError(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("project lookup failed")
		},
	}

	name1, name2, err := GetProjectNames(ctx, mock, 1, 2)
	assert.Error(t, err)
	assert.Empty(t, name1)
	assert.Empty(t, name2)
	assert.Contains(t, err.Error(), "failed to get project 1")
}

func TestGetProjectNames_SecondProjectError(t *testing.T) {
	ctx := context.Background()
	call := 0
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			call++
			if call == 1 {
				return &data.GetProjectResponse{ID: projectID, Name: "Project One"}, nil
			}
			return nil, errors.New("second lookup failed")
		},
	}

	name1, name2, err := GetProjectNames(ctx, mock, 1, 2)
	assert.Error(t, err)
	assert.Empty(t, name1)
	assert.Empty(t, name2)
	assert.Contains(t, err.Error(), "failed to get project 2")
}

func TestGetProjectName_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{
				ID:   projectID,
				Name: "Test Project",
			}, nil
		},
	}

	name, err := GetProjectName(context.Background(), mock, 1)
	assert.NoError(t, err)
	assert.Equal(t, "Test Project", name)
}

func TestGetProjectName_ProjectIsNilFallback(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, nil
		},
	}

	name, err := GetProjectName(context.Background(), mock, 42)
	assert.NoError(t, err)
	assert.Equal(t, "Project 42", name)
}

func TestGetProjectName_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("boom")
		},
	}

	name, err := GetProjectName(context.Background(), mock, 77)
	assert.Error(t, err)
	assert.Empty(t, name)
	assert.Contains(t, err.Error(), "failed to get project 77")
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
