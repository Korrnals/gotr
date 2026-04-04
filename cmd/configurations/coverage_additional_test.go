package configurations

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationsCmds_TooManyArgs(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "add-config", cmd: newAddConfigCmd(getClientForTests), args: []string{"1", "2", "--name", "X"}},
		{name: "add-group", cmd: newAddGroupCmd(getClientForTests), args: []string{"1", "2", "--name", "X"}},
		{name: "delete-config", cmd: newDeleteConfigCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "delete-group", cmd: newDeleteGroupCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "list", cmd: newListCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "update-config", cmd: newUpdateConfigCmd(getClientForTests), args: []string{"1", "2", "--name", "X"}},
		{name: "update-group", cmd: newUpdateGroupCmd(getClientForTests), args: []string{"1", "2", "--name", "X"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetContext(setupTestCmd(t, &client.MockClient{}).Context())
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestListCmd_NoArgs_Interactive_SelectProjectError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, assert.AnError
		},
	}

	cmd := newListCmd(getClientForTests)
	p := interactive.NewMockPrompter()
	cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, mock).Context(), p))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestConfigurationsCmds_NoArgs_Interactive_ResolveError(t *testing.T) {
	getProjectsErrMock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, assert.AnError
		},
	}

	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "add-config", cmd: newAddConfigCmd(getClientForTests), args: []string{"--name", "X"}},
		{name: "add-group", cmd: newAddGroupCmd(getClientForTests), args: []string{"--name", "X"}},
		{name: "delete-config", cmd: newDeleteConfigCmd(getClientForTests), args: []string{}},
		{name: "delete-group", cmd: newDeleteGroupCmd(getClientForTests), args: []string{}},
		{name: "update-config", cmd: newUpdateConfigCmd(getClientForTests), args: []string{"--name", "X"}},
		{name: "update-group", cmd: newUpdateGroupCmd(getClientForTests), args: []string{"--name", "X"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
		})
	}
}