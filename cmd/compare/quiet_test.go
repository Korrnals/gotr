package compare

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStandardIO(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutReader, stdoutWriter, err := os.Pipe()
	require.NoError(t, err)
	stderrReader, stderrWriter, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var copyWG sync.WaitGroup
	copyWG.Add(2)

	go func() {
		defer copyWG.Done()
		_, _ = io.Copy(&stdoutBuf, stdoutReader)
	}()
	go func() {
		defer copyWG.Done()
		_, _ = io.Copy(&stderrBuf, stderrReader)
	}()

	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	fn()

	require.NoError(t, stdoutWriter.Close())
	require.NoError(t, stderrWriter.Close())
	copyWG.Wait()

	return stdoutBuf.String(), stderrBuf.String()
}

func quietPlansMock() *client.MockClient {
	return &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test Project"}, nil
		},
		GetPlansFunc: func(ctx context.Context, projectID int64) (data.GetPlansResponse, error) {
			if projectID == 1 {
				return data.GetPlansResponse{{ID: 10, Name: "Plan A"}}, nil
			}
			return data.GetPlansResponse{{ID: 20, Name: "Plan B"}}, nil
		},
	}
}

func quietAllMock() *client.MockClient {
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

func TestPlansCmd_Quiet_JSON_PrintsOnlyResult(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return quietPlansMock()
	})

	cmd := newSimpleCompareCmd("plans", "plans", "test", "test", fetchPlanItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=json", "--quiet"})

	var cmdStdout bytes.Buffer
	var cmdStderr bytes.Buffer
	cmd.SetOut(&cmdStdout)
	cmd.SetErr(&cmdStderr)

	stdout, stderr := captureStandardIO(t, func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Empty(t, cmdStderr.String())
	assert.Contains(t, cmdStdout.String(), "\"resource\": \"plans\"")
	assert.Contains(t, cmdStdout.String(), "\"status\": \"complete\"")
	assert.NotContains(t, cmdStdout.String(), "STATS:")
	assert.NotContains(t, cmdStdout.String(), "Analysis completed")
	assert.NotContains(t, cmdStdout.String(), "Result saved to")
}

func TestPlansCmd_Quiet_SaveTo_SuppressesSuccessOutput(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return quietPlansMock()
	})

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "plans.json")

	cmd := newSimpleCompareCmd("plans", "plans", "test", "test", fetchPlanItems)
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=json", "--quiet", "--save-to=" + savePath})

	var cmdStdout bytes.Buffer
	var cmdStderr bytes.Buffer
	cmd.SetOut(&cmdStdout)
	cmd.SetErr(&cmdStderr)

	stdout, stderr := captureStandardIO(t, func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Empty(t, cmdStdout.String())
	assert.Empty(t, cmdStderr.String())

	content, err := os.ReadFile(savePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "\"resource\": \"plans\"")
}

func TestAllCmd_Quiet_SaveTo_SuppressesSuccessOutput(t *testing.T) {
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return quietAllMock()
	})

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "all.json")

	cmd := newAllCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=json", "--quiet", "--save-to=" + savePath})

	var cmdStdout bytes.Buffer
	var cmdStderr bytes.Buffer
	cmd.SetOut(&cmdStdout)
	cmd.SetErr(&cmdStderr)

	stdout, stderr := captureStandardIO(t, func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Empty(t, cmdStdout.String())
	assert.Empty(t, cmdStderr.String())

	content, err := os.ReadFile(savePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "\"meta\"")
	assert.Contains(t, string(content), "\"execution_status\": \"complete\"")
}