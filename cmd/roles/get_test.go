package roles

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

// ==================== Функциональные тесты с моком ====================

func TestGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetRoleFunc: func(roleID int64) (*data.Role, error) {
			assert.Equal(t, int64(1), roleID)
			return &data.Role{
				ID:   1,
				Name: "Administrator",
			}, nil
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGetCmd_WithOutputFile(t *testing.T) {
	mock := &client.MockClient{
		GetRoleFunc: func(roleID int64) (*data.Role, error) {
			return &data.Role{ID: 2, Name: "Tester"}, nil
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "role.json")

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"2", "-o", outputFile})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Tester")
}

func TestGetCmd_NotFound(t *testing.T) {
	mock := &client.MockClient{
		GetRoleFunc: func(roleID int64) (*data.Role, error) {
			return nil, fmt.Errorf("роль не найдена")
		},
	}

	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "роль не найдена")
}

// ==================== Тесты валидации ====================

func TestGetCmd_InvalidID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный role_id")
}

func TestGetCmd_ZeroID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный role_id")
}

func TestGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
