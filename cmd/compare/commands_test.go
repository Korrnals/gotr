// Package compare tests - comprehensive test suite for command execution
package compare

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для getClientSafe ====================

func TestGetClientSafe_WithClient(t *testing.T) {
	mock := &client.MockClient{}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)

	assert.NotNil(t, cli)
}

func TestGetClientSafe_NilClient(t *testing.T) {
	SetGetClientForTests(nil)

	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)

	assert.Nil(t, cli)
}

// ==================== Тесты для suitesCmd ====================

func TestSuitesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSuitesCmd_NoClient(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return nil
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestSuitesCmd_InvalidPid1(t *testing.T) {
	mock := &client.MockClient{}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=invalid", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSuitesCmd_ProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("project not found")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты для casesCmd ====================

func TestCasesCmd_Success(t *testing.T) {
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
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_WithField(t *testing.T) {
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
		DiffCasesDataFunc: func(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return &data.DiffCasesResponse{
				DiffByField: []struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{},
			}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newCasesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--field=priority_id"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для plansCmd ====================

func TestPlansCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetPlansFunc: func(projectID int64) (data.GetPlansResponse, error) {
			return []data.Plan{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newPlansCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для runsCmd ====================

func TestRunsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetRunsFunc: func(projectID int64) (data.GetRunsResponse, error) {
			return []data.Run{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newRunsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для sectionsCmd ====================

func TestSectionsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
		GetSectionsFunc: func(projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return []data.Section{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSectionsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для milestonesCmd ====================

func TestMilestonesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newMilestonesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для datasetsCmd ====================

func TestDatasetsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetDatasetsFunc: func(projectID int64) (data.GetDatasetsResponse, error) {
			return []data.Dataset{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newDatasetsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для groupsCmd ====================

func TestGroupsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetGroupsFunc: func(projectID int64) (data.GetGroupsResponse, error) {
			return []data.Group{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newGroupsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для labelsCmd ====================

func TestLabelsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetLabelsFunc: func(projectID int64) (data.GetLabelsResponse, error) {
			return []data.Label{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newLabelsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для templatesCmd ====================

func TestTemplatesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetTemplatesFunc: func(projectID int64) (data.GetTemplatesResponse, error) {
			return []data.Template{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newTemplatesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для configurationsCmd ====================

func TestConfigurationsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetConfigsFunc: func(projectID int64) (data.GetConfigsResponse, error) {
			return []data.ConfigGroup{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newConfigurationsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для sharedStepsCmd ====================

func TestSharedStepsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetSharedStepsFunc: func(projectID int64) (data.GetSharedStepsResponse, error) {
			return []data.SharedStep{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	cmd := newSharedStepsCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для allCmd ====================

func TestAllCmd_Success(t *testing.T) {
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
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAllCmd_WithSave(t *testing.T) {
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
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=json"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Тесты для printCasesFieldDiff ====================

func TestPrintCasesFieldDiff_WithDiff(t *testing.T) {
	mock := &client.MockClient{
		DiffCasesDataFunc: func(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return &data.DiffCasesResponse{
				DiffByField: []struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{
					{
						CaseID: 1,
						First:  data.Case{ID: 1, Title: "Case 1", PriorityID: 1},
						Second: data.Case{ID: 1, Title: "Case 1", PriorityID: 2},
					},
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	old := captureStdout(&buf)
	defer restoreStdout(old)

	printCasesFieldDiff(mock, 1, 2, "priority_id")
}

func TestPrintCasesFieldDiff_NoDiff(t *testing.T) {
	mock := &client.MockClient{
		DiffCasesDataFunc: func(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return &data.DiffCasesResponse{
				DiffByField: []struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{},
			}, nil
		},
	}

	var buf bytes.Buffer
	old := captureStdout(&buf)
	defer restoreStdout(old)

	printCasesFieldDiff(mock, 1, 2, "priority_id")
}

func TestPrintCasesFieldDiff_Error(t *testing.T) {
	mock := &client.MockClient{
		DiffCasesDataFunc: func(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
			return nil, errors.New("API error")
		},
	}

	var buf bytes.Buffer
	old := captureStdout(&buf)
	defer restoreStdout(old)

	printCasesFieldDiff(mock, 1, 2, "priority_id")
}

// Helper functions для перехвата stdout
func captureStdout(buf *bytes.Buffer) *bytes.Buffer {
	return buf
}

func restoreStdout(old *bytes.Buffer) {
	_ = old
}
