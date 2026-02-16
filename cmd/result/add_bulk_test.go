package result

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddBulkCmd_Success(t *testing.T) {
	// Создаём временный JSON файл с результатами
	tmpDir := t.TempDir()
	resultsFile := filepath.Join(tmpDir, "results.json")
	jsonContent := `[
		{"test_id": 101, "status_id": 1, "comment": "Test 1 passed"},
		{"test_id": 102, "status_id": 5, "comment": "Test 2 failed"}
	]`
	if err := os.WriteFile(resultsFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &client.MockClient{
		AddResultsFunc: func(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Len(t, req.Results, 2)
			return []data.Result{
				{ID: 1, TestID: 101, StatusID: 1},
				{ID: 2, TestID: 102, StatusID: 5},
			}, nil
		},
	}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--results-file", resultsFile})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddBulkCmd_WithCases(t *testing.T) {
	// Создаём временный JSON файл с case-based результатами
	tmpDir := t.TempDir()
	resultsFile := filepath.Join(tmpDir, "results.json")
	jsonContent := `[
		{"case_id": 201, "status_id": 1, "comment": "Case 1 passed"},
		{"case_id": 202, "status_id": 1, "comment": "Case 2 passed"}
	]`
	if err := os.WriteFile(resultsFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &client.MockClient{
		AddResultsForCasesFunc: func(runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Len(t, req.Results, 2)
			return []data.Result{
				{ID: 1, TestID: 301, StatusID: 1},
				{ID: 2, TestID: 302, StatusID: 1},
			}, nil
		},
	}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--results-file", resultsFile})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddBulkCmd_MissingResultsFile(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddBulkCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--results-file", "/nonexistent/file.json"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file")
}

func TestAddBulkCmd_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	resultsFile := filepath.Join(tmpDir, "results.json")
	jsonContent := `[{"test_id": 101, "status_id": 1}]`
	if err := os.WriteFile(resultsFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &client.MockClient{}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--results-file", resultsFile, "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddBulkCmd_InvalidRunID(t *testing.T) {
	tmpDir := t.TempDir()
	resultsFile := filepath.Join(tmpDir, "results.json")
	jsonContent := `[{"test_id": 101, "status_id": 1}]`
	if err := os.WriteFile(resultsFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &client.MockClient{}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--results-file", resultsFile})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddBulkCmd_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	resultsFile := filepath.Join(tmpDir, "results.json")
	jsonContent := `invalid json content`
	if err := os.WriteFile(resultsFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &client.MockClient{}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--results-file", resultsFile})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddBulkCmd_APIError(t *testing.T) {
	tmpDir := t.TempDir()
	resultsFile := filepath.Join(tmpDir, "results.json")
	jsonContent := `[{"test_id": 101, "status_id": 1}]`
	if err := os.WriteFile(resultsFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &client.MockClient{
		AddResultsFunc: func(runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	cmd := newAddBulkCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--results-file", resultsFile})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
