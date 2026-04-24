// Package bundlecmd implements the top-level `gotr import` command and the
// `gotr export snap|report` / `gotr import snap|report` subcommands that
// package and restore portable snapshot and report bundles.
package bundlecmd

import (
	"os"

	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// Register attaches export and import subcommands onto the given roots.
//
//	exportParent → `gotr export` (already exists; we add snap/report as children)
//	root         → `gotr` (we add a new top-level `import` command)
//
// Version is the running gotr version (written into bundle manifests).
func Register(root, exportParent *cobra.Command, version string) {
	exportParent.AddCommand(newExportSnapCmd(version))
	exportParent.AddCommand(newExportReportCmd(version))
	exportParent.AddCommand(newExportOrganizeCmd())

	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import portable snapshot or report bundles",
		Long:  "Import gotr bundles produced by `gotr export snap` or `gotr export report`.",
	}
	importCmd.AddCommand(newImportSnapCmd())
	importCmd.AddCommand(newImportReportCmd())
	root.AddCommand(importCmd)
}

// warnf is a shared helper that writes a warning to stderr.
func warnf(format string, args ...any) {
	ui.Warningf(os.Stderr, format, args...)
}

// successf writes a success message to stdout.
func successf(format string, args ...any) {
	ui.Successf(os.Stdout, format, args...)
}

// infof writes an info message to stdout.
func infof(format string, args ...any) {
	ui.Infof(os.Stdout, format, args...)
}
