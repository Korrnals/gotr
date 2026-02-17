package bdds

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Login\n  Given user is on login page"), 0644)
	assert.NoError(t, err)

	mock := &client.MockClient{
		AddBDDFunc: func(caseID int64, content string) (*data.BDD, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Contains(t, content, "Feature: Login")
			return &data.BDD{ID: 1, CaseID: caseID, Content: content}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--file", bddFile})

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Test"), 0644)
	assert.NoError(t, err)

	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--file", bddFile, "--dry-run"})

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestAddCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddCmd_MissingContent(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не может быть пустым")
}

func TestAddCmd_FileNotFound(t *testing.T) {
	mock := &client.MockClient{}
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345", "--file", "/nonexistent/file.feature"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось прочитать файл")
}

func TestAddCmd_ClientError(t *testing.T) {
	tmpDir := t.TempDir()
	bddFile := filepath.Join(tmpDir, "scenario.feature")
	err := os.WriteFile(bddFile, []byte("Feature: Test"), 0644)
	assert.NoError(t, err)

	mock := &client.MockClient{
		AddBDDFunc: func(caseID int64, content string) (*data.BDD, error) {
			return nil, fmt.Errorf("кейс не найден")
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999", "--file", bddFile})

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "кейс не найден")
}
