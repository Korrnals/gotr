package result

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
		{name: "add", build: func() *cobra.Command { return newAddCmd(testhelper.GetClientForTests) }, use: "add [test-id]", wantFlags: []string{"status-id", "comment", "version", "elapsed", "defects", "assigned-to", "dry-run"}},
		{name: "add-case", build: func() *cobra.Command { return newAddCaseCmd(testhelper.GetClientForTests) }, use: "add-case [run-id]", wantFlags: []string{"case-id", "status-id", "comment", "version", "elapsed", "defects", "assigned-to", "dry-run"}},
		{name: "add-bulk", build: func() *cobra.Command { return newAddBulkCmd(testhelper.GetClientForTests) }, use: "add-bulk [run-id]", wantFlags: []string{"results-file", "dry-run"}},
		{name: "get", build: func() *cobra.Command { return newGetCmd(testhelper.GetClientForTests) }, use: "get [test-id]", wantFlags: nil},
		{name: "get-case", build: func() *cobra.Command { return newGetCaseCmd(testhelper.GetClientForTests) }, use: "get-case [run-id] [case-id]", wantFlags: nil},
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
		{name: "add too many args", cmd: newAddCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
		{name: "add-case too many args", cmd: newAddCaseCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
		{name: "add-bulk missing args", cmd: newAddBulkCmd(testhelper.GetClientForTests), args: []string{}},
		{name: "add-bulk too many args", cmd: newAddBulkCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
		{name: "get too many args", cmd: newGetCmd(testhelper.GetClientForTests), args: []string{"1", "2"}},
		{name: "get-case too many args", cmd: newGetCaseCmd(testhelper.GetClientForTests), args: []string{"1", "2", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)
			err := tt.cmd.Execute()
			assert.Error(t, err)
		})
	}
}
