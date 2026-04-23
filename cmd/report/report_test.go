package report

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	Register(root)

	reportCmd, _, err := root.Find([]string{"report"})
	require.NoError(t, err)
	assert.Equal(t, "report", reportCmd.Name())

	names := []string{}
	for _, c := range reportCmd.Commands() {
		names = append(names, c.Name())
	}
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "view")
}

func TestListAndView(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	reportsDir := filepath.Join(home, ".gotr", "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	reportFile := "migration-20260418T120000Z-cases_snap.md"
	reportPath := filepath.Join(reportsDir, reportFile)
	require.NoError(t, os.WriteFile(reportPath, []byte("# body"), 0o644))

	listCmd := newListCmd()
	listOut := &bytes.Buffer{}
	listCmd.SetOut(listOut)
	require.NoError(t, listCmd.Execute())
	assert.Contains(t, listOut.String(), reportFile)

	viewCmd := newViewCmd()
	viewOut := &bytes.Buffer{}
	viewCmd.SetOut(viewOut)
	viewCmd.SetArgs([]string{"latest"})
	require.NoError(t, viewCmd.Execute())
	assert.Contains(t, viewOut.String(), reportFile)
	assert.Contains(t, viewOut.String(), "# body")
}
