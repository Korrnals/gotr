package configurations

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationCommands_NoPrompter_RequirePositionalID(t *testing.T) {
	mock := &client.MockClient{}

	tests := []struct {
		name    string
		newCmd  func() *cobra.Command
		args    []string
		errPart string
	}{
		{
			name: "add-config",
			newCmd: func() *cobra.Command {
				return newAddConfigCmd(getClientForTests)
			},
			args:    []string{"--name", "cfg"},
			errPart: "group_id is required in non-interactive mode",
		},
		{
			name: "add-group",
			newCmd: func() *cobra.Command {
				return newAddGroupCmd(getClientForTests)
			},
			args:    []string{"--name", "grp"},
			errPart: "project_id is required in non-interactive mode",
		},
		{
			name: "delete-config",
			newCmd: func() *cobra.Command {
				return newDeleteConfigCmd(getClientForTests)
			},
			args:    []string{},
			errPart: "config_id is required in non-interactive mode",
		},
		{
			name: "delete-group",
			newCmd: func() *cobra.Command {
				return newDeleteGroupCmd(getClientForTests)
			},
			args:    []string{},
			errPart: "group_id is required in non-interactive mode",
		},
		{
			name: "list",
			newCmd: func() *cobra.Command {
				return newListCmd(getClientForTests)
			},
			args:    []string{},
			errPart: "project_id is required in non-interactive mode",
		},
		{
			name: "update-config",
			newCmd: func() *cobra.Command {
				return newUpdateConfigCmd(getClientForTests)
			},
			args:    []string{"--name", "cfg2"},
			errPart: "config_id is required in non-interactive mode",
		},
		{
			name: "update-group",
			newCmd: func() *cobra.Command {
				return newUpdateGroupCmd(getClientForTests)
			},
			args:    []string{"--name", "grp2"},
			errPart: "group_id is required in non-interactive mode",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.newCmd()
			cmd.SetContext(setupTestCmd(t, mock).Context())
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errPart)
		})
	}
}
