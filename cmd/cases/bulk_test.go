package cases

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

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

func TestBulkCopyCmd_MissingSectionID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkCopyCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section-id")
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

func TestBulkMoveCmd_MissingSectionID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newBulkMoveCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1,2,3"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section-id")
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
