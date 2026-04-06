package plans

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandConstructors_MetadataAndFlags(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *cobra.Command
		use       string
		wantFlags []string
	}{
		{name: "close", build: func() *cobra.Command { return newCloseCmd(getClientForTests) }, use: "close [plan_id]", wantFlags: []string{"dry-run", "save"}},
		{name: "delete", build: func() *cobra.Command { return newDeleteCmd(getClientForTests) }, use: "delete [plan_id]", wantFlags: []string{"dry-run"}},
		{name: "get", build: func() *cobra.Command { return newGetCmd(getClientForTests) }, use: "get [plan_id]", wantFlags: []string{"save"}},
		{name: "list", build: func() *cobra.Command { return newListCmd(getClientForTests) }, use: "list [project_id]", wantFlags: []string{"save"}},
		{name: "update", build: func() *cobra.Command { return newUpdateCmd(getClientForTests) }, use: "update [plan_id]", wantFlags: []string{"dry-run", "save", "name", "description", "milestone-id"}},
		{name: "entry add", build: func() *cobra.Command { return newEntryAddCmd(getClientForTests) }, use: "add [plan_id]", wantFlags: []string{"dry-run", "save", "suite-id", "name", "config-ids"}},
		{name: "entry update", build: func() *cobra.Command { return newEntryUpdateCmd(getClientForTests) }, use: "update [plan_id] [entry_id]", wantFlags: []string{"dry-run", "save", "name"}},
		{name: "entry delete", build: func() *cobra.Command { return newEntryDeleteCmd(getClientForTests) }, use: "delete [plan_id] [entry_id]", wantFlags: []string{"dry-run"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.build()
			require.NotNil(t, cmd)
			assert.Equal(t, tt.use, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)

			for _, flagName := range tt.wantFlags {
				assert.NotNil(t, cmd.Flags().Lookup(flagName), "missing flag: %s", flagName)
			}
		})
	}
}

func TestNewEntryCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := newEntryCmd(getClientForTests)
	require.NotNil(t, cmd)
	assert.Equal(t, "entry", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.Nil(t, cmd.RunE)

	assert.NotNil(t, cmd.Commands())
	assert.Len(t, cmd.Commands(), 3)
	assert.NotNil(t, cmd.Commands()[0])
	assert.NotNil(t, cmd.Commands()[1])
	assert.NotNil(t, cmd.Commands()[2])

	names := []string{cmd.Commands()[0].Use, cmd.Commands()[1].Use, cmd.Commands()[2].Use}
	assert.Contains(t, names, "add [plan_id]")
	assert.Contains(t, names, "update [plan_id] [entry_id]")
	assert.Contains(t, names, "delete [plan_id] [entry_id]")
}

func TestCommandConstructors_ArgsValidation(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "close too many args", cmd: newCloseCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "delete too many args", cmd: newDeleteCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "get too many args", cmd: newGetCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "list too many args", cmd: newListCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "update too many args", cmd: newUpdateCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "entry add too many args", cmd: newEntryAddCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "entry update too many args", cmd: newEntryUpdateCmd(getClientForTests), args: []string{"1", "e", "extra"}},
		{name: "entry delete too many args", cmd: newEntryDeleteCmd(getClientForTests), args: []string{"1", "e", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestEntryCmd_EmptyEntryIDRejected(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "entry update empty entry id", cmd: newEntryUpdateCmd(getClientForTests), args: []string{"100", ""}},
		{name: "entry delete empty entry id", cmd: newEntryDeleteCmd(getClientForTests), args: []string{"100", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetContext(setupTestCmd(t, nil).Context())
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "entry_id is required")
		})
	}
}
