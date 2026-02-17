package cases

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Bulk Command Tests ====================

func TestNewBulkCmd(t *testing.T) {
	cmd := newBulkCmd(getClientForTests)

	// Verify command properties
	assert.Equal(t, "bulk", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Verify all subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 4)

	// Check subcommand names
	subNames := make([]string, 0, 4)
	for _, sub := range subcommands {
		subNames = append(subNames, sub.Name())
	}
	assert.Contains(t, subNames, "update")
	assert.Contains(t, subNames, "delete")
	assert.Contains(t, subNames, "copy")
	assert.Contains(t, subNames, "move")
}

// ==================== Bulk Update Tests ====================

func TestBulkUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--suite-id=100", "--priority-id=1", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateCasesFunc: func(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
			assert.Equal(t, int64(100), suiteID)
			assert.Equal(t, []int64{1, 2, 3}, req.CaseIDs)
			assert.Equal(t, int64(1), req.PriorityID)
			return &data.GetCasesResponse{}, nil
		},
	}

	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--suite-id=100", "--priority-id=1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkUpdateCmd_WithEstimate(t *testing.T) {
	mock := &client.MockClient{
		UpdateCasesFunc: func(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
			assert.Equal(t, int64(100), suiteID)
			assert.Equal(t, []int64{10, 20}, req.CaseIDs)
			assert.Equal(t, "1h 30m", req.Estimate)
			return &data.GetCasesResponse{}, nil
		},
	}

	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10,20", "--suite-id=100", "--estimate=1h 30m"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkUpdateCmd_NoCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--suite-id=100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case IDs")
}

func TestBulkUpdateCmd_MissingSuiteID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suite-id")
}

func TestBulkUpdateCmd_InvalidCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"abc,def", "--suite-id=100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid case IDs")
}

func TestBulkUpdateCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateCasesFunc: func(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--suite-id=100", "--priority-id=1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestBulkUpdateCmd_WithSave(t *testing.T) {
	mock := &client.MockClient{
		UpdateCasesFunc: func(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
			return &data.GetCasesResponse{
				{ID: 1, Title: "Test Case 1"},
				{ID: 2, Title: "Test Case 2"},
			}, nil
		},
	}

	cmd := newBulkUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2", "--suite-id=100", "--priority-id=1", "--save"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Bulk Delete Tests ====================

func TestBulkDeleteCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--suite-id=100", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkDeleteCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		DeleteCasesFunc: func(suiteID int64, req *data.DeleteCasesRequest) error {
			assert.Equal(t, int64(100), suiteID)
			assert.Equal(t, []int64{1, 2, 3}, req.CaseIDs)
			return nil
		},
	}

	cmd := newBulkDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--suite-id=100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkDeleteCmd_NoCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--suite-id=100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case IDs")
}

func TestBulkDeleteCmd_InvalidCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"abc,def", "--suite-id=100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid case IDs")
}

func TestBulkDeleteCmd_MissingSuiteID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suite-id")
}

func TestBulkDeleteCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeleteCasesFunc: func(suiteID int64, req *data.DeleteCasesRequest) error {
			return fmt.Errorf("cannot delete: cases have results")
		},
	}

	cmd := newBulkDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--suite-id=100"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Bulk Copy Tests ====================

func TestBulkCopyCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkCopyCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--section-id=50", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkCopyCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		CopyCasesToSectionFunc: func(sectionID int64, req *data.CopyCasesRequest) error {
			assert.Equal(t, int64(50), sectionID)
			assert.Equal(t, []int64{1, 2, 3}, req.CaseIDs)
			return nil
		},
	}

	cmd := newBulkCopyCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--section-id=50"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkCopyCmd_NoCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkCopyCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--section-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case IDs")
}

func TestBulkCopyCmd_InvalidCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkCopyCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"abc,def", "--section-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid case IDs")
}

func TestBulkCopyCmd_MissingSectionID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkCopyCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section-id")
}

func TestBulkCopyCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		CopyCasesToSectionFunc: func(sectionID int64, req *data.CopyCasesRequest) error {
			return fmt.Errorf("section not found")
		},
	}

	cmd := newBulkCopyCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--section-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section not found")
}

// ==================== Bulk Move Tests ====================

func TestBulkMoveCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkMoveCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--section-id=50", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkMoveCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		MoveCasesToSectionFunc: func(sectionID int64, req *data.MoveCasesRequest) error {
			assert.Equal(t, int64(50), sectionID)
			assert.Equal(t, []int64{1, 2, 3}, req.CaseIDs)
			return nil
		},
	}

	cmd := newBulkMoveCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--section-id=50"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestBulkMoveCmd_NoCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkMoveCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--section-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case IDs")
}

func TestBulkMoveCmd_InvalidCaseIDs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkMoveCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"abc,def", "--section-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid case IDs")
}

func TestBulkMoveCmd_MissingSectionID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkMoveCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section-id")
}

func TestBulkMoveCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		MoveCasesToSectionFunc: func(sectionID int64, req *data.MoveCasesRequest) error {
			return fmt.Errorf("cannot move: permission denied")
		},
	}

	cmd := newBulkMoveCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3", "--section-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

// ==================== Parse ID List Tests ====================

func TestParseIDList_BulkUpdate(t *testing.T) {
	ids := parseIDList([]string{"1,2,3", "4", "5,6"})
	assert.Equal(t, []int64{1, 2, 3, 4, 5, 6}, ids)
}

func TestParseIDList_BulkWithSpaces(t *testing.T) {
	ids := parseIDList([]string{"1, 2, 3"})
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

func TestParseIDList_BulkInvalidMixed(t *testing.T) {
	ids := parseIDList([]string{"1", "abc", "2", "-1", "0", "3"})
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

func TestParseIDList_EmptyInput(t *testing.T) {
	ids := parseIDList([]string{})
	assert.Empty(t, ids)
}

func TestParseIDList_AllInvalid(t *testing.T) {
	ids := parseIDList([]string{"abc", "def", "xyz"})
	assert.Empty(t, ids)
}

func TestParseIDList_EmptyParts(t *testing.T) {
	ids := parseIDList([]string{"1,,2", ",3,"})
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

// ==================== Output Result Tests ====================

func TestOutputResult_Stdout(t *testing.T) {
	// Create a command with the save flag properly defined
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", false, "")
	data := map[string]string{"key": "value"}

	err := outputResult(cmd, data)
	assert.NoError(t, err)
}

func TestOutputResult_WithSave(t *testing.T) {
	// Create a command with the save flag set to true
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", true, "")

	data := map[string]string{"key": "value", "test": "data"}
	err := outputResult(cmd, data)
	assert.NoError(t, err)
}
