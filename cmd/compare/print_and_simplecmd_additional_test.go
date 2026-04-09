package compare

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureCompareStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	fn()

	require.NoError(t, w.Close())
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	return buf.String()
}

func TestWave6G_PrintCompareResult_DefaultStructuredSaveError(t *testing.T) {
	homeFile := filepath.Join(t.TempDir(), "home-file")
	require.NoError(t, os.WriteFile(homeFile, []byte("not-a-dir"), 0o644))
	t.Setenv("HOME", homeFile)

	cmd := &cobra.Command{Use: "test"}

	err := PrintCompareResult(cmd, CompareResult{Resource: "cases"}, "P1", "P2", "json", "__DEFAULT__")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create exports directory")
}

func TestWave6G_PrintCompareResult_SaveToStructuredError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", true, "")

	err := PrintCompareResult(cmd, CompareResult{Resource: "cases"}, "P1", "P2", "json", "/nonexistent/path/result.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file write error")
}

func TestWave6G_PrintCompareResult_SaveToTablePath(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", true, "")

	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
	}
	path := filepath.Join(t.TempDir(), "result.txt")

	err := PrintCompareResult(cmd, result, "Project One", "Project Two", "table", path)
	require.NoError(t, err)

	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "Case A")
	assert.Contains(t, string(content), "Project One")
}

func TestWave6G_PrintCompareAllStageProgress_NilWriter(t *testing.T) {
	output := captureCompareStderr(t, func() {
		printCompareAllStageProgress(nil, "suites")
	})

	assert.Contains(t, output, "Compare all stages")
	assert.Contains(t, output, "suites")
	assert.Contains(t, output, "active")
}

func TestWave6G_SimpleCompareCmd_ParseCommonFlagsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, errors.New("projects unavailable")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() {
		SetGetClientForTests(nil)
	})

	cmd := newSimpleCompareCmd("dummy", "dummy", "dummy", "dummy", func(ctx context.Context, cli client.ClientInterface, pid int64) ([]ItemInfo, error) {
		t.Fatal("fetch should not be called")
		return nil, nil
	})
	addPersistentFlagsForTests(cmd)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pid1 not specified and interactive selection failed")
	assert.Contains(t, err.Error(), "projects unavailable")
}

func TestWave6G_SimpleCompareCmd_GetProjectNamesError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return nil, errors.New("project lookup failed")
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() {
		SetGetClientForTests(nil)
	})

	cmd := newSimpleCompareCmd("dummy", "dummy", "dummy", "dummy", func(ctx context.Context, cli client.ClientInterface, pid int64) ([]ItemInfo, error) {
		t.Fatal("fetch should not be called")
		return nil, nil
	})
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project lookup failed")
}

func TestWave6G_SimpleCompareCmd_PrintCompareResultError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Project"}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})
	t.Cleanup(func() {
		SetGetClientForTests(nil)
	})

	cmd := newSimpleCompareCmd("dummy", "dummy", "dummy", "dummy", func(ctx context.Context, cli client.ClientInterface, pid int64) ([]ItemInfo, error) {
		return []ItemInfo{{ID: pid, Name: "Item"}}, nil
	})
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=1", "--pid2=2", "--format=xml", "--save-to=" + filepath.Join(t.TempDir(), "out.xml"), "--quiet"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format: xml")
}