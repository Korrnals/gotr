// Package compare tests - comprehensive test suite for "all" subcommand
package compare

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSaveAllResult_InvalidPathYAML(t *testing.T) {
	result := &allResult{}
	err := saveAllResult(result, "yaml", "/nonexistent/path/result.yaml")
	assert.Error(t, err)
}

// ==================== Тесты для allCmd с различными сценариями ====================

func TestAllCmd_WithErrors(t *testing.T) {
	// Мок который возвращает ошибки для некоторых ресурсов
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, errors.New("suites error")
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return nil, errors.New("sections error")
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err) // Команда не должна падать из-за ошибок отдельных ресурсов
}

func TestAllCmd_AllResourceErrorsStillCompletes(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesParallelFunc: func(ctx context.Context, projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error) {
			return nil, errors.New("parallel suites failed")
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return nil, errors.New("sharedsteps failed")
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, errors.New("runs failed")
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return nil, errors.New("plans failed")
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return nil, errors.New("milestones failed")
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return nil, errors.New("datasets failed")
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return nil, errors.New("groups failed")
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return nil, errors.New("labels failed")
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return nil, errors.New("templates failed")
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return nil, errors.New("configurations failed")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAllCmd_SaveYAML(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "result.yaml")

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=yaml", "--save-to=" + savePath})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	// Command should succeed - save flag triggers save to default location
	assert.NoError(t, err)
}

func TestAllCmd_SaveDefaultTable(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--save", "--quiet"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAllCmd_SaveToJsonByExtension(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "result.json")

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--save-to=" + savePath, "--quiet"})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, readErr := os.ReadFile(savePath)
	assert.NoError(t, readErr)
	assert.NotEmpty(t, content)
}

func TestAllCmd_NoClient(t *testing.T) {
	originalGetClient := getClient
	defer func() { getClient = originalGetClient }()

	SetGetClientForTests(nil)

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client not initialized")
}

func TestAllCmd_GetProjectNamesError(t *testing.T) {
	originalGetClient := getClient
	defer func() { getClient = originalGetClient }()

	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("projects unavailable")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get project 1")
}

func TestAllCmd_SaveTo_UnsupportedFormat_Error(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=xml", "--save-to=out.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format 'xml' for save")
}

func TestBuildAllMeta_Interrupted(t *testing.T) {
	res := &allResult{}
	fillInterruptedResults(res, 1, 2)
	errs := map[string]error{"execution": errors.New("interrupted by user")}

	meta := buildAllMeta(res, true, errs, 1500*time.Millisecond)

	assert.Equal(t, CompareStatusInterrupted, meta.ExecutionStatus)
	assert.True(t, meta.Interrupted)
	assert.Equal(t, 1, meta.ErrorCount)
	assert.Contains(t, meta.ErrorResources, "execution")
	assert.NotEmpty(t, meta.GeneratedAt)
}

func TestBuildAllMeta_Complete(t *testing.T) {
	res := &allResult{
		Cases:          &CompareResult{Status: CompareStatusComplete},
		Suites:         &CompareResult{Status: CompareStatusComplete},
		Sections:       &CompareResult{Status: CompareStatusComplete},
		SharedSteps:    &CompareResult{Status: CompareStatusComplete},
		Runs:           &CompareResult{Status: CompareStatusComplete},
		Plans:          &CompareResult{Status: CompareStatusComplete},
		Milestones:     &CompareResult{Status: CompareStatusComplete},
		Datasets:       &CompareResult{Status: CompareStatusComplete},
		Groups:         &CompareResult{Status: CompareStatusComplete},
		Labels:         &CompareResult{Status: CompareStatusComplete},
		Templates:      &CompareResult{Status: CompareStatusComplete},
		Configurations: &CompareResult{Status: CompareStatusComplete},
	}

	meta := buildAllMeta(res, false, map[string]error{}, 2*time.Second)

	assert.Equal(t, CompareStatusComplete, meta.ExecutionStatus)
	assert.False(t, meta.Interrupted)
	assert.Equal(t, 0, meta.ErrorCount)
	assert.Empty(t, meta.ErrorResources)
}

func TestFillResourcePartialResult_MarksFailedResourceAsPartial(t *testing.T) {
	res := &allResult{}

	fillResourcePartialResult(res, "datasets", 30, 34)

	if assert.NotNil(t, res.Datasets) {
		assert.Equal(t, CompareStatusPartial, res.Datasets.Status)
		assert.Equal(t, int64(30), res.Datasets.Project1ID)
		assert.Equal(t, int64(34), res.Datasets.Project2ID)
	}
}

func TestBuildAllMeta_WithResourceErrors_IsPartial(t *testing.T) {
	res := &allResult{
		Cases:          &CompareResult{Status: CompareStatusComplete},
		Suites:         &CompareResult{Status: CompareStatusComplete},
		Sections:       &CompareResult{Status: CompareStatusComplete},
		SharedSteps:    &CompareResult{Status: CompareStatusComplete},
		Runs:           &CompareResult{Status: CompareStatusComplete},
		Plans:          &CompareResult{Status: CompareStatusComplete},
		Milestones:     &CompareResult{Status: CompareStatusComplete},
		Datasets:       &CompareResult{Status: CompareStatusPartial},
		Groups:         &CompareResult{Status: CompareStatusPartial},
		Labels:         &CompareResult{Status: CompareStatusPartial},
		Templates:      &CompareResult{Status: CompareStatusComplete},
		Configurations: &CompareResult{Status: CompareStatusComplete},
	}

	errs := map[string]error{
		"datasets": errors.New("unsupported endpoint"),
		"groups":   errors.New("unsupported endpoint"),
		"labels":   errors.New("unsupported endpoint"),
	}

	meta := buildAllMeta(res, false, errs, 2*time.Second)

	assert.Equal(t, CompareStatusPartial, meta.ExecutionStatus)
	assert.Equal(t, 3, meta.ErrorCount)
	assert.Contains(t, meta.ErrorResources, "datasets")
	assert.Contains(t, meta.ErrorResources, "groups")
	assert.Contains(t, meta.ErrorResources, "labels")
}

func TestBuildAllMeta_UnsupportedEndpoints_NotCountedAsErrors(t *testing.T) {
	res := &allResult{
		Cases:          &CompareResult{Status: CompareStatusComplete},
		Suites:         &CompareResult{Status: CompareStatusComplete},
		Sections:       &CompareResult{Status: CompareStatusComplete},
		SharedSteps:    &CompareResult{Status: CompareStatusComplete},
		Runs:           &CompareResult{Status: CompareStatusComplete},
		Plans:          &CompareResult{Status: CompareStatusComplete},
		Milestones:     &CompareResult{Status: CompareStatusComplete},
		Datasets:       &CompareResult{Status: CompareStatusPartial},
		Groups:         &CompareResult{Status: CompareStatusPartial},
		Labels:         &CompareResult{Status: CompareStatusPartial},
		Templates:      &CompareResult{Status: CompareStatusComplete},
		Configurations: &CompareResult{Status: CompareStatusComplete},
	}

	errs := map[string]error{
		"datasets": errors.New("API returned 404 File Not Found: Unknown method 'get_datasets'"),
		"groups":   errors.New("API returned 404 File Not Found: Unknown method 'get_groups'"),
		"labels":   errors.New("API returned 404 File Not Found: Unknown method 'get_labels'"),
	}

	meta := buildAllMeta(res, false, errs, 2*time.Second)

	assert.Equal(t, CompareStatusPartial, meta.ExecutionStatus)
	assert.Equal(t, 0, meta.ErrorCount)
	assert.Empty(t, meta.ErrorResources)
	assert.Equal(t, 3, meta.UnsupportedCount)
	assert.Contains(t, meta.UnsupportedResources, "datasets")
	assert.Contains(t, meta.UnsupportedResources, "groups")
	assert.Contains(t, meta.UnsupportedResources, "labels")
}

func TestIsUnsupportedEndpointError(t *testing.T) {
	assert.True(t, isUnsupportedEndpointError(errors.New("API returned 404 File Not Found: Unknown method 'get_groups'")))
	assert.False(t, isUnsupportedEndpointError(errors.New("request timeout")))
}

func TestAddCommonFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "compare-all"}
	addCommonFlags(cmd)
	assert.NotNil(t, cmd)
}

func TestSaveAllSummaryToFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "summary.txt")

	err := saveAllSummaryToFile(
		&cobra.Command{},
		&allResult{},
		"Project A", 1,
		"Project B", 2,
		map[string]error{},
		tmpFile,
		1500*time.Millisecond,
	)
	assert.NoError(t, err)

	data, readErr := os.ReadFile(tmpFile)
	assert.NoError(t, readErr)
	assert.Contains(t, string(data), "RESOURCE SUMMARY")
}

func TestAppendResourceRow_Statuses(t *testing.T) {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"Resource", "Only in P1", "Only in P2", "Common", "Data status"})

	appendResourceRow(tw, "Cases", nil, errors.New("API returned 404 File Not Found: Unknown method 'get_cases'"))
	appendResourceRow(tw, "Runs", &CompareResult{Status: CompareStatusInterrupted}, nil)
	appendResourceRow(tw, "Plans", &CompareResult{Status: CompareStatusPartial}, errors.New("network"))
	appendResourceRow(tw, "Suites", &CompareResult{Status: CompareStatusComplete}, nil)
	appendResourceRow(tw, "Other", &CompareResult{Status: "mystery"}, nil)

	out := tw.Render()
	assert.Contains(t, out, "Cases")
	assert.Contains(t, out, "Runs")
	assert.Contains(t, out, "Plans")
	assert.Contains(t, out, "Suites")
	assert.Contains(t, out, "Other")
}

func TestFillResourcePartialResult_NoOpBranches(t *testing.T) {
	fillResourcePartialResult(nil, "cases", 1, 2)

	res := &allResult{}
	fillResourcePartialResult(res, "unknown", 1, 2)

	assert.Nil(t, res.Cases)
	assert.Nil(t, res.Suites)
}

func TestFillResourcePartialResult_AllResources(t *testing.T) {
	res := &allResult{}
	resources := []string{"cases", "suites", "sections", "shared_steps", "runs", "plans", "milestones", "datasets", "groups", "labels", "templates", "configurations"}

	for _, r := range resources {
		fillResourcePartialResult(res, r, 30, 34)
	}

	assert.NotNil(t, res.Cases)
	assert.NotNil(t, res.Suites)
	assert.NotNil(t, res.Sections)
	assert.NotNil(t, res.SharedSteps)
	assert.NotNil(t, res.Runs)
	assert.NotNil(t, res.Plans)
	assert.NotNil(t, res.Milestones)
	assert.NotNil(t, res.Datasets)
	assert.NotNil(t, res.Groups)
	assert.NotNil(t, res.Labels)
	assert.NotNil(t, res.Templates)
	assert.NotNil(t, res.Configurations)
}

func TestSaveAllSummaryToFile_DefaultPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cmd := &cobra.Command{Use: "compare-all"}
	cmd.Flags().Bool("quiet", false, "")

	err := saveAllSummaryToFile(
		cmd,
		&allResult{},
		"Project A", 1,
		"Project B", 2,
		map[string]error{},
		"__DEFAULT__",
		500*time.Millisecond,
	)
	assert.NoError(t, err)
}

func TestSaveAllSummaryToFile_WriteError(t *testing.T) {
	cmd := &cobra.Command{Use: "compare-all"}
	cmd.Flags().Bool("quiet", false, "")

	badPath := t.TempDir()
	err := saveAllSummaryToFile(
		cmd,
		&allResult{},
		"Project A", 1,
		"Project B", 2,
		map[string]error{},
		badPath,
		500*time.Millisecond,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file write error")
}

func TestSaveAllSummaryToFile_QuietMode(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "summary-quiet.txt")

	cmd := &cobra.Command{Use: "compare-all"}
	cmd.Flags().Bool("quiet", false, "")
	requireSetErr := cmd.Flags().Set("quiet", "true")
	assert.NoError(t, requireSetErr)

	err := saveAllSummaryToFile(
		cmd,
		&allResult{},
		"Project A", 1,
		"Project B", 2,
		map[string]error{},
		tmpFile,
		500*time.Millisecond,
	)
	assert.NoError(t, err)
}

func TestFillResourcePartialResult_DoesNotOverwriteExisting(t *testing.T) {
	res := &allResult{
		Cases: &CompareResult{Status: CompareStatusComplete},
	}

	fillResourcePartialResult(res, "cases", 1, 2)
	assert.Equal(t, CompareStatusComplete, res.Cases.Status)
}

func TestPrintAllSummaryTable_WithMixedErrors(t *testing.T) {
	result := &allResult{
		Cases: &CompareResult{Status: CompareStatusComplete, OnlyInFirst: []ItemInfo{{ID: 1, Name: "A"}}},
		Suites: &CompareResult{Status: CompareStatusPartial},
	}
	errs := map[string]error{
		"datasets": errors.New("API returned 404 File Not Found: Unknown method 'get_datasets'"),
		"runs":     errors.New("network timeout"),
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printAllSummaryTable("P1", 1, "P2", 2, result, errs, time.Second)

	_ = w.Close()
	os.Stdout = oldStdout
	_ = r.Close()
}

func TestAppendResourceRow_NilResultWithRegularError(t *testing.T) {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"Resource", "Only in P1", "Only in P2", "Common", "Data status"})

	appendResourceRow(tw, "Cases", nil, errors.New("network"))

	out := tw.Render()
	assert.Contains(t, out, "INTERRUPTED")
}

func TestAppendResourceRow_PartialWithUnsupportedError(t *testing.T) {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"Resource", "Only in P1", "Only in P2", "Common", "Data status"})

	appendResourceRow(
		tw,
		"Datasets",
		&CompareResult{Status: CompareStatusPartial},
		errors.New("API returned 404 File Not Found: Unknown method 'get_datasets'"),
	)

	out := tw.Render()
	assert.Contains(t, out, "UNSUPPORTED")
}

func TestCompareResultStatus(t *testing.T) {
	assert.Equal(t, CompareStatusInterrupted, compareResultStatus(nil))
	assert.Equal(t, CompareStatusInterrupted, compareResultStatus(&CompareResult{}))
	assert.Equal(t, CompareStatusComplete, compareResultStatus(&CompareResult{Status: CompareStatusComplete}))
}

func TestDeriveAllExecutionStatus(t *testing.T) {
	base := &allResult{
		Cases:          &CompareResult{Status: CompareStatusComplete},
		Suites:         &CompareResult{Status: CompareStatusComplete},
		Sections:       &CompareResult{Status: CompareStatusComplete},
		SharedSteps:    &CompareResult{Status: CompareStatusComplete},
		Runs:           &CompareResult{Status: CompareStatusComplete},
		Plans:          &CompareResult{Status: CompareStatusComplete},
		Milestones:     &CompareResult{Status: CompareStatusComplete},
		Datasets:       &CompareResult{Status: CompareStatusComplete},
		Groups:         &CompareResult{Status: CompareStatusComplete},
		Labels:         &CompareResult{Status: CompareStatusComplete},
		Templates:      &CompareResult{Status: CompareStatusComplete},
		Configurations: &CompareResult{Status: CompareStatusComplete},
	}

	assert.Equal(t, CompareStatusInterrupted, deriveAllExecutionStatus(base, true, nil))
	assert.Equal(t, CompareStatusPartial, deriveAllExecutionStatus(base, false, map[string]error{"cases": errors.New("boom")}))

	partial := *base
	partial.Runs = &CompareResult{Status: CompareStatusPartial}
	assert.Equal(t, CompareStatusPartial, deriveAllExecutionStatus(&partial, false, nil))

	interrupted := *base
	interrupted.Groups = &CompareResult{Status: CompareStatusInterrupted}
	assert.Equal(t, CompareStatusInterrupted, deriveAllExecutionStatus(&interrupted, false, nil))

	assert.Equal(t, CompareStatusComplete, deriveAllExecutionStatus(base, false, nil))
}

// ==================== Comprehensive tests for isContextCancellationError ====================

func TestIsContextCancellationError_NilError(t *testing.T) {
	assert.False(t, isContextCancellationError(nil))
}

func TestIsContextCancellationError_ContextCanceled(t *testing.T) {
	assert.True(t, isContextCancellationError(context.Canceled))
}

func TestIsContextCancellationError_ContextDeadlineExceeded(t *testing.T) {
	assert.True(t, isContextCancellationError(context.DeadlineExceeded))
}

func TestIsContextCancellationError_StringMatch_ContextCanceled(t *testing.T) {
	err := errors.New("operation failed: context canceled due to timeout")
	assert.True(t, isContextCancellationError(err))
}

func TestIsContextCancellationError_StringMatch_DeadlineExceeded(t *testing.T) {
	err := errors.New("request failed: deadline exceeded")
	assert.True(t, isContextCancellationError(err))
}

func TestIsContextCancellationError_CaseInsensitive(t *testing.T) {
	err := errors.New("CONTEXT CANCELED")
	assert.True(t, isContextCancellationError(err))

	err2 := errors.New("DeAdLiNe ExCeEdEd")
	assert.True(t, isContextCancellationError(err2))
}

func TestIsContextCancellationError_NotMatchingError(t *testing.T) {
	err := errors.New("network timeout")
	assert.False(t, isContextCancellationError(err))

	err2 := errors.New("connection refused")
	assert.False(t, isContextCancellationError(err2))
}

func TestSplitErrorsBySupport_SeparatesAndSorts(t *testing.T) {
	unsupported, regular := splitErrorsBySupport(map[string]error{
		"runs":     errors.New("read tcp timeout"),
		"datasets": errors.New("API returned 404 File Not Found: Unknown method 'get_datasets'"),
		"groups":   errors.New("API returned 404 File Not Found: Unknown method 'get_groups'"),
		"cases":    errors.New("upstream 503"),
	})

	assert.Equal(t, []string{"datasets", "groups"}, unsupported)
	assert.Equal(t, []string{"cases", "runs"}, regular)
}

func TestFillInterruptedResults_FillsAllMissingAndPreservesExisting(t *testing.T) {
	res := &allResult{
		Cases: &CompareResult{Status: CompareStatusComplete, Resource: "cases"},
	}

	fillInterruptedResults(res, 10, 20)

	assert.Equal(t, CompareStatusComplete, res.Cases.Status)
	require.NotNil(t, res.Suites)
	require.NotNil(t, res.Sections)
	require.NotNil(t, res.SharedSteps)
	require.NotNil(t, res.Runs)
	require.NotNil(t, res.Plans)
	require.NotNil(t, res.Milestones)
	require.NotNil(t, res.Datasets)
	require.NotNil(t, res.Groups)
	require.NotNil(t, res.Labels)
	require.NotNil(t, res.Templates)
	require.NotNil(t, res.Configurations)
	assert.Equal(t, CompareStatusInterrupted, res.Suites.Status)
	assert.Equal(t, int64(10), res.Suites.Project1ID)
	assert.Equal(t, int64(20), res.Suites.Project2ID)
}

func newAllCmdHappyPathMock() *client.MockClient {
	return &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return []data.Case{}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
		GetSharedStepsFunc: func(ctx context.Context, projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
		GetMilestonesFunc: func(ctx context.Context, projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
		GetDatasetsFunc: func(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
		GetGroupsFunc: func(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
		GetLabelsFunc: func(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
		GetTemplatesFunc: func(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
		GetConfigsFunc: func(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
}

func TestAllCmd_SaveDefaultStructuredFormats(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			mock := newAllCmdHappyPathMock()
			SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
				return mock
			})

			cmd := newAllCmd()
			addPersistentFlagsForTests(cmd)
			cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--save", "--format=" + format, "--quiet"})

			err := cmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestAllCmd_SaveToTableCustomPath(t *testing.T) {
	mock := newAllCmdHappyPathMock()
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	savePath := filepath.Join(t.TempDir(), "summary.txt")

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=table", "--save-to=" + savePath, "--quiet"})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, readErr := os.ReadFile(savePath)
	assert.NoError(t, readErr)
	assert.Contains(t, string(content), "RESOURCE SUMMARY")
}

func TestAllCmd_InterruptedFlowWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock := newAllCmdHappyPathMock()
	mock.GetSuitesParallelFunc = func(ctx context.Context, projectIDs []int64, workers int) (map[int64]data.GetSuitesResponse, error) {
		return nil, context.Canceled
	}
	mock.GetSuitesFunc = func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
		return nil, context.Canceled
	}

	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--quiet"})

	err := cmd.Execute()
	assert.NoError(t, err)
}


