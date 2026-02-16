package plans

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Entry Add Tests ====================

func TestEntryAddCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--suite-id=50", "--name=Entry 1", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEntryAddCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		AddPlanEntryFunc: func(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, int64(50), req.SuiteID)
			assert.Equal(t, "Entry 1", req.Name)
			return &data.Plan{ID: 100}, nil
		},
	}

	cmd := newEntryAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--suite-id=50", "--name=Entry 1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEntryAddCmd_WithConfigIDs(t *testing.T) {
	mock := &client.MockClient{
		AddPlanEntryFunc: func(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, []int64{1, 2, 3}, req.ConfigIDs)
			return &data.Plan{ID: 100}, nil
		},
	}

	cmd := newEntryAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--suite-id=50", "--config-ids=1,2,3"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEntryAddCmd_MissingSuiteID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suite-id")
}

// ==================== Entry Update Tests ====================

func TestEntryUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "abc123", "--name=Updated", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEntryUpdateCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdatePlanEntryFunc: func(planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, "abc123", entryID)
			assert.Equal(t, "Updated Entry", req.Name)
			return &data.Plan{ID: 100}, nil
		},
	}

	cmd := newEntryUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "abc123", "--name=Updated Entry"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEntryUpdateCmd_MissingEntryID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Entry Delete Tests ====================

func TestEntryDeleteCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "abc123", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEntryDeleteCmd_Success(t *testing.T) {
	deleteCalled := false
	mock := &client.MockClient{
		DeletePlanEntryFunc: func(planID int64, entryID string) error {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, "abc123", entryID)
			deleteCalled = true
			return nil
		},
	}

	cmd := newEntryDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "abc123"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, deleteCalled)
}

func TestEntryDeleteCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		DeletePlanEntryFunc: func(planID int64, entryID string) error {
			return fmt.Errorf("entry not found")
		},
	}

	cmd := newEntryDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "xyz999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Parse Int List Tests ====================

func TestParseIntList_Valid(t *testing.T) {
	ids := parseIntList("1,2,3")
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

func TestParseIntList_WithSpaces(t *testing.T) {
	ids := parseIntList("1, 2, 3")
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

func TestParseIntList_InvalidMixed(t *testing.T) {
	ids := parseIntList("1,abc,2")
	assert.Equal(t, []int64{1, 2}, ids)
}

func TestParseIntList_Empty(t *testing.T) {
	ids := parseIntList("")
	assert.Empty(t, ids)
}

// ==================== Additional Edge Case Tests ====================

func TestEntryAddCmd_InvalidPlanID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--suite-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestEntryAddCmd_ZeroPlanID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0", "--suite-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestEntryAddCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		AddPlanEntryFunc: func(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
			return nil, fmt.Errorf("plan not found")
		},
	}

	cmd := newEntryAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "--suite-id=50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEntryUpdateCmd_InvalidPlanID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "abc123"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestEntryUpdateCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		UpdatePlanEntryFunc: func(planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
			return nil, fmt.Errorf("entry not found")
		},
	}

	cmd := newEntryUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100", "abc123", "--name=Updated"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEntryDeleteCmd_InvalidPlanID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newEntryDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "abc123"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestParseIntList_NegativeNumbers(t *testing.T) {
	// Negative numbers should be filtered out
	ids := parseIntList("1,-5,3")
	assert.Equal(t, []int64{1, 3}, ids)
}

func TestParseIntList_Zero(t *testing.T) {
	// Zero should be filtered out
	ids := parseIntList("1,0,3")
	assert.Equal(t, []int64{1, 3}, ids)
}
