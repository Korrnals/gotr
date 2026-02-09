package cases

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Dry Run Tests ====================

func TestUpdateCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--title=New Title", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_DryRun_NoFlags(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// ==================== Functional Tests with Mock ====================

func TestUpdateCmd_Success_Title(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Equal(t, "New Title", *req.Title)
			return &data.Case{ID: 12345, Title: "New Title"}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--title=New Title"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_Success_Priority(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Equal(t, int64(2), *req.PriorityID)
			return &data.Case{ID: 12345, Title: "Test", PriorityID: 2}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--priority-id=2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_Success_Type(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Equal(t, int64(3), *req.TypeID)
			return &data.Case{ID: 12345, Title: "Test", TypeID: 3}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--type-id=3"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_Success_Refs(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Equal(t, "JIRA-123, JIRA-456", *req.Refs)
			return &data.Case{ID: 12345, Title: "Test", Refs: "JIRA-123, JIRA-456"}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--refs=JIRA-123, JIRA-456"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestUpdateCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "--title=New"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case not found")
}

// ==================== Validation Tests ====================

func TestUpdateCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid", "--title=New"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
