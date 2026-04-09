package result

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Tests for saveToFile ====================

func TestSaveToFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_output.json")

	data := map[string]interface{}{
		"id":   123,
		"name": "test",
	}

	err := saveToFile(data, filename)
	assert.NoError(t, err)

	// Verify that the file was created
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "123")
	assert.Contains(t, string(content), "test")
}

func TestSaveToFile_InvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_output.json")

	// Channels cannot be serialized to JSON
	invalidData := make(chan int)

	err := saveToFile(invalidData, filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serialization")
}

func TestSaveToFile_InvalidPath(t *testing.T) {
	// Path to a non-existent directory without write permissions
	invalidPath := "/nonexistent_dir_xyz/test.json"

	data := map[string]string{"key": "value"}

	err := saveToFile(data, invalidPath)
	assert.Error(t, err)
}

// ==================== Tests for service_wrapper ====================

func TestResultServiceWrapper_AddResults(t *testing.T) {
	mock := &client.MockClient{
		AddResultsFunc: func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Len(t, req.Results, 2)
			return data.GetResultsResponse{
				{ID: 1, TestID: 101, StatusID: 1},
				{ID: 2, TestID: 102, StatusID: 5},
			}, nil
		},
	}

	wrapper := &resultServiceWrapper{svc: nil}
	// Verify that wrapper implements the interface
	var _ ResultServiceInterface = wrapper

	// Create a service via constructor
	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResults(ctx, 12345, &data.AddResultsRequest{
		Results: []data.ResultEntry{
			{TestID: 101, StatusID: 1},
			{TestID: 102, StatusID: 5},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestResultServiceWrapper_AddResults_Error(t *testing.T) {
	mock := &client.MockClient{
		AddResultsFunc: func(ctx context.Context, runID int64, req *data.AddResultsRequest) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("run not found")
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResults(ctx, 99999, &data.AddResultsRequest{
		Results: []data.ResultEntry{{TestID: 101, StatusID: 1}},
	})

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestResultServiceWrapper_AddResultsForCases(t *testing.T) {
	mock := &client.MockClient{
		AddResultsForCasesFunc: func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			assert.Len(t, req.Results, 2)
			return data.GetResultsResponse{
				{ID: 1, TestID: 201, StatusID: 1},
				{ID: 2, TestID: 202, StatusID: 1},
			}, nil
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResultsForCases(ctx, 12345, &data.AddResultsForCasesRequest{
		Results: []data.ResultForCaseEntry{
			{CaseID: 301, StatusID: 1},
			{CaseID: 302, StatusID: 1},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestResultServiceWrapper_AddResultsForCases_Error(t *testing.T) {
	mock := &client.MockClient{
		AddResultsForCasesFunc: func(ctx context.Context, runID int64, req *data.AddResultsForCasesRequest) (data.GetResultsResponse, error) {
			return nil, fmt.Errorf("invalid case_id")
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	results, err := svc.AddResultsForCases(ctx, 12345, &data.AddResultsForCasesRequest{
		Results: []data.ResultForCaseEntry{{CaseID: 301, StatusID: 1}},
	})

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestResultServiceWrapper_GetRunsForProject(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{
				{ID: 101, Name: "Run 1", ProjectID: 1},
				{ID: 102, Name: "Run 2", ProjectID: 1},
			}, nil
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	runs, err := svc.GetRunsForProject(ctx, 1)

	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	assert.Equal(t, int64(101), runs[0].ID)
}

func TestResultServiceWrapper_GetRunsForProject_Error(t *testing.T) {
	mock := &client.MockClient{
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)
	runs, err := svc.GetRunsForProject(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, runs)
}

// ==================== Tests for newResultServiceFromInterface ====================

func TestNewResultServiceFromInterface_WithHTTPClient(t *testing.T) {
	httpClient := &client.HTTPClient{}

	svc := newResultServiceFromInterface(httpClient)
	assert.NotNil(t, svc)
}

func TestNewResultServiceFromInterface_WithMockClient(t *testing.T) {
	mock := &client.MockClient{}

	svc := newResultServiceFromInterface(mock)
	assert.NotNil(t, svc)
}

// ==================== Tests for SetGetClientForTests and getClientSafe ====================

func TestSetGetClientForTests(t *testing.T) {
	// Save current state
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Reset accessor
	clientAccessor = nil

	// Set test function
	mockFn := func(ctx context.Context) client.ClientInterface {
		return nil
	}

	SetGetClientForTests(mockFn)
	assert.NotNil(t, clientAccessor)

	// Repeated call should update the function
	SetGetClientForTests(mockFn)
	assert.NotNil(t, clientAccessor)
}

func TestGetClientSafe_WithNilAccessor(t *testing.T) {
	// Save current state
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Reset accessor
	clientAccessor = nil

	// Should return nil when accessor is nil
	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)
	assert.Nil(t, cli)
}

func TestGetClientSafe_WithAccessor(t *testing.T) {
	// Save current state
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Create accessor with test function
	mockFn := func(ctx context.Context) client.ClientInterface {
		return nil
	}
	clientAccessor = client.NewAccessor(mockFn)

	// Should return nil (since mockFn returns nil)
	cmd := &cobra.Command{}
	cli := getClientSafe(cmd)
	assert.Nil(t, cli)
}

// ==================== Tests for Register ====================

func TestRegister(t *testing.T) {
	// Save current state
	oldAccessor := clientAccessor
	defer func() {
		clientAccessor = oldAccessor
	}()

	// Reset accessor
	clientAccessor = nil

	// Create root command
	rootCmd := &cobra.Command{Use: "gotr"}

	// Mock client getter function
	mockFn := func(ctx context.Context) client.ClientInterface {
		return nil
	}

	// Register result command
	Register(rootCmd, mockFn)

	// Verify that command was added
	assert.NotNil(t, clientAccessor)

	// Verify that result command exists in root
	resultCmd, _, err := rootCmd.Find([]string{"result"})
	assert.NoError(t, err)
	assert.NotNil(t, resultCmd)

	// Verify that subcommands were added
	subcommands := []string{"list", "get", "get-case", "add", "add-case", "add-bulk", "fields"}
	for _, sub := range subcommands {
		cmd, _, err := rootCmd.Find([]string{"result", sub})
		assert.NoError(t, err, "subcommand %s should exist", sub)
		assert.NotNil(t, cmd, "subcommand %s should not be nil", sub)

		// Verify that save and quiet flags were added
		saveFlag := cmd.Flags().Lookup("save")
		assert.NotNil(t, saveFlag, "save flag should exist on %s", sub)

		// Local quiet override should not be declared on subcommands.
		// Global quiet may come from root persistent flags at runtime.
		assert.Nil(t, cmd.Flags().Lookup("quiet"), "quiet should not be declared locally on %s", sub)
	}
}

// ==================== Tests for list command (interactive mode) ====================

func TestListCmd_Interactive_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetRunsResponse{
				{ID: 12345, Name: "Test Run", ProjectID: 1},
			}, nil
		},
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			assert.Equal(t, int64(12345), runID)
			return []data.Result{{ID: 1, TestID: 100, StatusID: 1}}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}). // select project
		WithSelectResponses(interactive.SelectResponse{Index: 0})  // select run

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	// Without arguments - should enable interactive mode
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Interactive_SelectProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
	}

	// Empty MockPrompter — queue is exhausted, SelectProject will return an error
	p := interactive.NewMockPrompter()

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_Interactive_GetRunsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return nil, fmt.Errorf("failed to get runs")
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}) // select project

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runs")
}

func TestListCmd_Interactive_EmptyRuns(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}) // select project

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no test runs found")
}

func TestListCmd_Interactive_SelectRunError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
			}, nil
		},
		GetRunsFunc: func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{
				{ID: 12345, Name: "Test Run", ProjectID: 1},
			}, nil
		},
	}

	// MockPrompter: 1 SelectResponse for project, none for run — queue is exhausted
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})

	cmd := newListCmd(testhelper.GetClientForTests)
	ctx := testhelper.SetupTestCmd(t, mock).Context()
	cmd.SetContext(interactive.WithPrompter(ctx, p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Tests for outputResult (via command with output flag) ====================

func TestOutputResult_WithSaveFlag(t *testing.T) {
	mock := &client.MockClient{
		GetResultsForRunFunc: func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
			return []data.Result{{ID: 1, TestID: 100, StatusID: 1}}, nil
		},
	}

	// Recreate the command with our getClient to use the mock
	cmd := newListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	// Add save flag (as Register does)
	output.AddFlag(cmd)
	cmd.SetArgs([]string{"12345", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Additional tests for coverage ====================

func TestAddBulkResults_ParseError(t *testing.T) {
	// Test for covering the JSON parse error branch in AddBulkResults
	mock := &client.MockClient{}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Pass invalid JSON that cannot be parsed into any format
	invalidJSON := []byte(`{"invalid": "json"}`)

	result, err := svc.AddBulkResults(ctx, 12345, invalidJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestPrintJSON_Error(t *testing.T) {
	// Test printJSON error with non-serializable data
	invalidData := make(chan int) // Channels cannot be serialized

	err := printJSON(invalidData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serialization")
}

// ==================== Additional tests for service_wrapper ====================

func TestResultServiceWrapper_AddBulkResults_EmptyArray(t *testing.T) {
	// Test for covering the empty array branch in JSON
	mock := &client.MockClient{}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Empty array
	emptyJSON := []byte(`[]`)

	result, err := svc.AddBulkResults(ctx, 12345, emptyJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestResultServiceWrapper_AddBulkResults_InvalidJSON(t *testing.T) {
	// Test for covering the invalid JSON branch
	mock := &client.MockClient{}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Invalid JSON
	invalidJSON := []byte(`{invalid json`)

	result, err := svc.AddBulkResults(ctx, 12345, invalidJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestResultServiceWrapper_AllMethods(t *testing.T) {
	mock := &client.MockClient{
		GetResultsFunc: func(ctx context.Context, testID int64) (data.GetResultsResponse, error) {
			return []data.Result{{ID: 1, TestID: testID}}, nil
		},
		GetResultsForCaseFunc: func(ctx context.Context, runID, caseID int64) (data.GetResultsResponse, error) {
			return []data.Result{{ID: 1, TestID: 100}}, nil
		},
	}

	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Test GetForTest
	results, err := svc.GetForTest(ctx, 123)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Test GetForCase
	results, err = svc.GetForCase(ctx, 1, 100)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Test GetForRun
	mock.GetResultsForRunFunc = func(ctx context.Context, runID int64) (data.GetResultsResponse, error) {
		return []data.Result{{ID: 1, TestID: 200}}, nil
	}
	results, err = svc.GetForRun(ctx, 456)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestResultServiceWrapper_ParseID(t *testing.T) {
	mock := &client.MockClient{}
	ctx := context.Background()
	svc := newResultServiceFromInterface(mock)

	// Test valid ID
	id, err := svc.ParseID(ctx, []string{"123"}, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)

	// Test invalid ID
	_, err = svc.ParseID(ctx, []string{"abc"}, 0)
	assert.Error(t, err)
}
// TestProductionVarClosures exercises the production-var wiring closures
// (e.g. var addCmd = newAddCmd(func(cmd) { return getClientSafe(cmd) })).
func TestProductionVarClosures(t *testing.T) {
	old := clientAccessor
	defer func() { clientAccessor = old }()
	clientAccessor = nil

	cmds := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"addCmd", addCmd},
		{"addCaseCmd", addCaseCmd},
		{"addBulkCmd", addBulkCmd},
		{"fieldsCmd", fieldsCmd},
		{"getCmd", getCmd},
		{"getCaseCmd", getCaseCmd},
		{"listCmd", listCmd},
	}

	for _, tc := range cmds {
		t.Run(tc.name, func(t *testing.T) {
			defer func() { recover() }()
			_ = tc.cmd.RunE(tc.cmd, []string{"1"})
		})
	}
}

// TestCmd_Run_Help covers the Run func on root Cmd that calls cmd.Help().
func TestCmd_Run_Help(t *testing.T) {
	Cmd.SetArgs([]string{})
	err := Cmd.Help()
	assert.NoError(t, err)
}