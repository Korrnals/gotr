package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfigCmd_ValidArgsFunction(t *testing.T) {
	if assert.NotNil(t, configCmd.ValidArgsFunction) {
		items, directive := configCmd.ValidArgsFunction(configCmd, []string{}, "")
		assert.ElementsMatch(t, []string{"init", "path", "view", "edit"}, items)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

		items, directive = configCmd.ValidArgsFunction(configCmd, []string{"init"}, "")
		assert.Nil(t, items)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestListCmd_ValidArgsFunction(t *testing.T) {
	if assert.NotNil(t, listCmd.ValidArgsFunction) {
		items, directive := listCmd.ValidArgsFunction(listCmd, []string{}, "")
		assert.NotEmpty(t, items)
		assert.Contains(t, items, "cases")
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

		items, directive = listCmd.ValidArgsFunction(listCmd, []string{"cases"}, "")
		assert.Nil(t, items)
		assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
	}
}

func TestExportCmd_ValidArgsFunction(t *testing.T) {
	if assert.NotNil(t, exportCmd.ValidArgsFunction) {
		items, directive := exportCmd.ValidArgsFunction(exportCmd, []string{}, "")
		assert.NotEmpty(t, items)
		assert.Contains(t, items, "cases")
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

		items, directive = exportCmd.ValidArgsFunction(exportCmd, []string{"cases"}, "")
		assert.NotNil(t, items)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

		items, directive = exportCmd.ValidArgsFunction(exportCmd, []string{"cases", "get_cases"}, "")
		assert.Nil(t, items)
		assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
	}
}

func TestInitGlobalFlags_ExistOnRootCmd(t *testing.T) {
	persistent := rootCmd.PersistentFlags()
	for _, name := range []string{"url", "username", "api-key", "insecure", "config", "debug", "quiet", "non-interactive", "format"} {
		assert.NotNil(t, persistent.Lookup(name), "expected persistent flag %s to exist", name)
	}
}
