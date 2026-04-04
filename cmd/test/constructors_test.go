package test

import (
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
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
		{name: "get", build: func() *cobra.Command { return newGetCmd(testhelper.GetClientForTests) }, use: "get [test-id]", wantFlags: []string{"save"}},
		{name: "list", build: func() *cobra.Command { return newListCmd(testhelper.GetClientForTests) }, use: "list [run-id]", wantFlags: []string{"save", "status-id", "assigned-to"}},
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
		{name: "get too many args", cmd: newGetCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
		{name: "list too many args", cmd: newListCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
		})
	}
}
