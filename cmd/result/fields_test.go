package result

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestFieldsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetResultFieldsFunc: func() (data.GetResultFieldsResponse, error) {
			return []data.ResultField{
				{ID: 1, Name: "Status", SystemName: "status_id", IsActive: true},
				{ID: 2, Name: "Comment", SystemName: "comment", IsActive: true},
				{ID: 3, Name: "Version", SystemName: "version", IsActive: true},
			}, nil
		},
	}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	save.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestFieldsCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetResultFieldsFunc: func() (data.GetResultFieldsResponse, error) {
			return []data.ResultField{}, nil
		},
	}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	save.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestFieldsCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetResultFieldsFunc: func() (data.GetResultFieldsResponse, error) {
			return nil, fmt.Errorf("failed to get result fields")
		},
	}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	save.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestFieldsCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	save.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
