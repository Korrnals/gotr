package reports

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
		{name: "list", build: func() *cobra.Command { return newListCmd(testhelper.GetClientForTests) }, use: "list [project_id]", wantFlags: []string{"save"}},
		{name: "run", build: func() *cobra.Command { return newRunCmd(testhelper.GetClientForTests) }, use: "run [template_id]", wantFlags: []string{"dry-run", "save"}},
		{name: "run-cross-project", build: func() *cobra.Command { return newRunCrossProjectCmd(testhelper.GetClientForTests) }, use: "run-cross-project [template_id]", wantFlags: []string{"dry-run", "save"}},
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
		{name: "list too many args", cmd: newListCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
		{name: "run too many args", cmd: newRunCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
		{name: "run-cross-project too many args", cmd: newRunCrossProjectCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
		})
	}
}
