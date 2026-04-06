// Package ui — see display.go.
// File table.go — static output of final results (tables + JSON).
//
// Concept:
//
//	display.go — live progress (ANSI, refresh loop, Task reporter)
//	table.go   — static output: Table(), JSON(), NewTable()
//
// Usage:
//
//	// Create and populate a table:
//	t := ui.NewTable(cmd)
//	t.AppendHeader(table.Row{"ID", "Name"})
//	t.AppendRow(table.Row{1, "foo"})
//	ui.Table(cmd, t)    // render in --format (table/json/csv/md)
//
//	// Output an arbitrary value as JSON:
//	if err := ui.JSON(cmd, myStruct); err != nil { return err }
package ui

import (
	"encoding/json"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// OutputFormat represents allowed values for the --format flag.
type OutputFormat string

// Supported output formats for table rendering and command output selection.
const (
	FormatTable    OutputFormat = "table"
	FormatJSON     OutputFormat = "json"
	FormatCSV      OutputFormat = "csv"
	FormatMarkdown OutputFormat = "md"
	FormatHTML     OutputFormat = "html"
)

// NewTable creates a go-pretty table.Writer with base configuration:
//   - output goes to cmd.OutOrStdout()
//   - style: StyleRounded (rounded frame borders)
//
// Call Table(cmd, t) to render the table in the requested format.
func NewTable(cmd *cobra.Command) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	t.SetStyle(table.StyleRounded)
	return t
}

// Table renders a table in the format specified by the command's --format flag.
// If the flag is absent or has an unknown value, renders as a table (StyleRounded).
//
// Supported formats: table (default), csv, md, html.
// For JSON output, use ui.JSON(cmd, yourSlice).
func Table(cmd *cobra.Command, t table.Writer) {
	t.SetOutputMirror(cmd.OutOrStdout())

	format := getFormat(cmd)
	switch format {
	case FormatCSV:
		fmt.Fprintln(cmd.OutOrStdout(), t.RenderCSV())
	case FormatMarkdown:
		fmt.Fprintln(cmd.OutOrStdout(), t.RenderMarkdown())
	case FormatHTML:
		fmt.Fprintln(cmd.OutOrStdout(), t.RenderHTML())
	default:
		// table, json, or unknown format — render as table
		fmt.Fprintln(cmd.OutOrStdout(), t.Render())
	}
}

// JSON serializes v as indented JSON and prints it to cmd.OutOrStdout().
// Used for non-tabular data (objects, arrays, raw API responses).
func JSON(cmd *cobra.Command, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("ui.JSON: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

// IsJSON returns true if the command was invoked with --format json.
// Useful for conditional logic:
//
//	if ui.IsJSON(cmd) { /* raw output */ } else { /* table */ }
func IsJSON(cmd *cobra.Command) bool {
	return getFormat(cmd) == FormatJSON
}

// IsQuiet returns true if the --quiet flag is set.
func IsQuiet(cmd *cobra.Command) bool {
	q, _ := cmd.Flags().GetBool("quiet")
	return q
}

// getFormat reads --format from the command's flags.
// Looks first in local flags, then in inherited flags (parent PersistentFlags).
func getFormat(cmd *cobra.Command) OutputFormat {
	if f := cmd.Flags().Lookup("format"); f != nil {
		return OutputFormat(f.Value.String())
	}
	if f := cmd.InheritedFlags().Lookup("format"); f != nil {
		return OutputFormat(f.Value.String())
	}
	return FormatTable
}
