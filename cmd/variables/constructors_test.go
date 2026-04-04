package variables

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
		{name: "add", build: func() *cobra.Command { return newAddCmd(getClientForTests) }, use: "add [dataset_id]", wantFlags: []string{"dry-run", "save", "name"}},
		{name: "delete", build: func() *cobra.Command { return newDeleteCmd(getClientForTests) }, use: "delete [variable_id]", wantFlags: []string{"dry-run"}},
		{name: "list", build: func() *cobra.Command { return newListCmd(getClientForTests) }, use: "list [dataset_id]", wantFlags: []string{"save"}},
		{name: "update", build: func() *cobra.Command { return newUpdateCmd(getClientForTests) }, use: "update [variable_id]", wantFlags: []string{"dry-run", "save", "name"}},
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

func TestCommandConstructors_ArgsValidation(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "add too many args", cmd: newAddCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "delete too many args", cmd: newDeleteCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "list too many args", cmd: newListCmd(getClientForTests), args: []string{"1", "2"}},
		{name: "update too many args", cmd: newUpdateCmd(getClientForTests), args: []string{"1", "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
		})
	}
}
