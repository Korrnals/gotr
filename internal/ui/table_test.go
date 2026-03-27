package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newTableCmd() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(buf)
	cmd.Flags().String("format", "table", "")
	cmd.Flags().Bool("quiet", false, "")
	return cmd, buf
}

func TestTableHelpers(t *testing.T) {
	cmd, _ := newTableCmd()
	cmd.Flags().Set("format", "json")
	if !IsJSON(cmd) {
		t.Fatalf("expected IsJSON=true for format=json")
	}
	cmd.Flags().Set("quiet", "true")
	if !IsQuiet(cmd) {
		t.Fatalf("expected IsQuiet=true")
	}
}

func TestTableRenderFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "table", format: "table", want: "NAME"},
		{name: "csv", format: "csv", want: "Name"},
		{name: "md", format: "md", want: "|"},
		{name: "html", format: "html", want: "<table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, out := newTableCmd()
			cmd.Flags().Set("format", tt.format)

			tw := NewTable(cmd)
			tw.AppendHeader(table.Row{"Name"})
			tw.AppendRow(table.Row{"Value"})
			Table(cmd, tw)

			if got := out.String(); !strings.Contains(got, tt.want) {
				t.Fatalf("output for format=%s does not contain %q: %s", tt.format, tt.want, got)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	cmd, out := newTableCmd()
	err := JSON(cmd, map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "\"ok\": true") {
		t.Fatalf("unexpected JSON output: %s", got)
	}
}
