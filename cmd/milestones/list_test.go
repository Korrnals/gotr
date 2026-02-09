package milestones

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Функциональные тесты с моком ====================

func TestListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			assert.Equal(t, int64(1), projectID)
			return []data.Milestone{
				{ID: 100, Name: "Release 1.0"},
				{ID: 200, Name: "Release 2.0"},
			}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_Empty(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			return []data.Milestone{}, nil
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestListCmd_ClientError(t *testing.T) {
	mock := &client.MockClient{
		GetMilestonesFunc: func(projectID int64) ([]data.Milestone, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// ==================== Тесты валидации ====================

func TestListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
