// Package compare tests - comprehensive test suite for "all" subcommand
package compare

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для allResult структуры ====================

func TestAllResult_Struct(t *testing.T) {
	result := &allResult{
		Cases: &CompareResult{
			Resource:   "cases",
			Project1ID: 1,
			Project2ID: 2,
		},
		Suites: &CompareResult{
			Resource:   "suites",
			Project1ID: 1,
			Project2ID: 2,
		},
	}

	assert.NotNil(t, result.Cases)
	assert.NotNil(t, result.Suites)
	assert.Equal(t, "cases", result.Cases.Resource)
}

// ==================== Тесты для saveAllResult ====================

func TestSaveAllResult_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.json")

	result := &allResult{
		Cases: &CompareResult{
			Resource:   "cases",
			Project1ID: 1,
			Project2ID: 2,
		},
	}

	err := saveAllResult(result, "json", path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "cases")
}

func TestSaveAllResult_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.yaml")

	result := &allResult{}

	err := saveAllResult(result, "yaml", path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestSaveAllResult_InvalidFormat(t *testing.T) {
	result := &allResult{}
	err := saveAllResult(result, "invalid", "/tmp/test.txt")
	assert.Error(t, err)
}

func TestSaveAllResult_InvalidPath(t *testing.T) {
	result := &allResult{}
	err := saveAllResult(result, "json", "/nonexistent/path/result.json")
	assert.Error(t, err)
}

// ==================== Тесты для printResourceSummary ====================

func TestPrintResourceSummary_Empty(t *testing.T) {
	result := &CompareResult{
		Resource:     "datasets",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{},
		OnlyInSecond: []ItemInfo{},
		Common:       []CommonItemInfo{},
	}

	var buf bytes.Buffer
	old := captureStdout(&buf)
	defer restoreStdout(old)

	printResourceSummary("datasets", result)
}

// ==================== Тесты для allCmd с различными сценариями ====================

func TestAllCmd_WithErrors(t *testing.T) {
	// Мок который возвращает ошибки для некоторых ресурсов
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return nil, errors.New("suites error")
		},
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return nil, errors.New("sections error")
		},
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err) // Команда не должна падать из-за ошибок отдельных ресурсов
}

func TestAllCmd_SaveYAML(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=yaml", "--save"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	// Command should succeed - save flag triggers save to default location
	assert.NoError(t, err)
}
