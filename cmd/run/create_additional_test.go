package run

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestCreateCmd_TooManyArgs(t *testing.T) {
	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"30", "31", "--suite-id", "20069", "--name", "X"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCreateCmd_NoArgs_Interactive_SelectProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, assert.AnError
		},
	}

	cmd := newCreateCmd(testhelper.GetClientForTests)
	cmd.SetContext(interactive.WithPrompter(testhelper.SetupTestCmd(t, mock).Context(), interactive.NewMockPrompter()))
	cmd.SetArgs([]string{"--name", "X"})

	err := cmd.Execute()
	assert.Error(t, err)
}