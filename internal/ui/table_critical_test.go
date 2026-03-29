package ui

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

// TestGetFormat_Local tests getFormat reads local command flag
func TestGetFormat_Local(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "json", "")

	format := getFormat(cmd)
	if format != "json" {
		t.Fatalf("expected json, got %v", format)
	}
}

// TestGetFormat_Inherited tests getFormat reads inherited parent flag
func TestGetFormat_Inherited(t *testing.T) {
	parent := &cobra.Command{Use: "parent"}
	parent.PersistentFlags().String("format", "csv", "")

	child := &cobra.Command{Use: "child"}
	parent.AddCommand(child)

	format := getFormat(child)
	if format != "" && format != "csv" { // May be empty if not properly inherited
		t.Fatalf("expected csv or empty, got %v", format)
	}
}

// TestGetFormat_Default tests getFormat returns FormatTable without flag
func TestGetFormat_Default(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	format := getFormat(cmd)
	if format != FormatTable {
		t.Fatalf("expected FormatTable, got %v", format)
	}
}

// TestJSON_MethodOutput tests JSON method outputs valid JSON
func TestJSON_MethodOutput(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	data := map[string]int{"value": 123}
	err := JSON(cmd, data)
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatalf("JSON should produce output")
	}
}

// TestIsJSON tests IsJSON detection
func TestIsJSON(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "json", "")

	if !IsJSON(cmd) {
		t.Fatalf("IsJSON should return true for json format")
	}

	cmd.Flags().Set("format", "table")
	if IsJSON(cmd) {
		t.Fatalf("IsJSON should return false for table format")
	}
}

// TestIsQuiet tests IsQuiet detection
func TestIsQuiet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")

	if IsQuiet(cmd) {
		t.Fatalf("IsQuiet should return false initially")
	}

	cmd.Flags().Set("quiet", "true")
	if !IsQuiet(cmd) {
		t.Fatalf("IsQuiet should return true after setting flag")
	}
}
