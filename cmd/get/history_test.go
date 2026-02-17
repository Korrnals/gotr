package get

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для get case-history ====================

func TestCaseHistoryCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetHistoryForCaseFunc: func(caseID int64) (*data.GetHistoryForCaseResponse, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.GetHistoryForCaseResponse{
				History: []struct {
					ID        int64           "json:\"id\""
					TypeID    int64           "json:\"type_id\""
					CreatedOn int64           "json:\"created_on\""
					UserID    int64           "json:\"user_id\""
					Changes   []data.Change   "json:\"changes\""
				}{
					{
						ID:        1,
						TypeID:    1,
						CreatedOn: 1234567890,
						UserID:    5,
						Changes: []data.Change{
							{Field: "title", OldText: "Old Title", NewText: "New Title"},
						},
					},
				},
			}, nil
		},
	}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseHistoryCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID кейса")
}

func TestCaseHistoryCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCaseHistoryCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetHistoryForCaseFunc: func(caseID int64) (*data.GetHistoryForCaseResponse, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newCaseHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты для get sharedstep-history ====================

func TestSharedStepHistoryCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepHistoryFunc: func(stepID int64) (*data.GetSharedStepHistoryResponse, error) {
			assert.Equal(t, int64(56789), stepID)
			return &data.GetSharedStepHistoryResponse{
				History: []struct {
					ID                   int64          "json:\"id\""
					Timestamp            int64          "json:\"timestamp\""
					UserID               int64          "json:\"user_id\""
					CustomStepsSeparated []data.Step    "json:\"custom_steps_separated,omitempty\""
					Title                string         "json:\"title,omitempty\""
				}{
					{
						ID:        1,
						Timestamp: 1234567890,
						UserID:    5,
						Title:     "Updated Step Title",
					},
				},
			}, nil
		},
	}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"56789"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSharedStepHistoryCmd_InvalidStepID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID шага")
}

func TestSharedStepHistoryCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSharedStepHistoryCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSharedStepHistoryFunc: func(stepID int64) (*data.GetSharedStepHistoryResponse, error) {
			return nil, fmt.Errorf("shared step not found")
		},
	}

	cmd := newSharedStepHistoryCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCaseHistoryCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCaseHistoryCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestSharedStepHistoryCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newSharedStepHistoryCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}
