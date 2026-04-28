package report

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShow_Print_MdFile(t *testing.T) {
	home := setHome(t)
	reportsDir := filepath.Join(home, ".gotr", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fp := filepath.Join(reportsDir, "r.md")
	body := "# header\nbody text"
	if err := os.WriteFile(fp, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newShowCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--print", "r.md"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out.String() != body {
		t.Errorf("got %q, want %q", out.String(), body)
	}
}

func TestShow_Print_Json(t *testing.T) {
	home := setHome(t)
	reportsDir := filepath.Join(home, ".gotr", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fp := filepath.Join(reportsDir, "r.json")
	if err := os.WriteFile(fp, []byte(`{"a":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newShowCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--print", "r.json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), `{"a":1}`) {
		t.Errorf("unexpected output: %q", out.String())
	}
}

func TestShow_Print_RejectsPDF(t *testing.T) {
	home := setHome(t)
	reportsDir := filepath.Join(home, ".gotr", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fp := filepath.Join(reportsDir, "r.pdf")
	if err := os.WriteFile(fp, []byte("%PDF-1.4\n..."), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newShowCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--print", "r.pdf"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for PDF + --print")
	}
	if !strings.Contains(err.Error(), "--print does not support PDF") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestShow_Print_NotFound(t *testing.T) {
	home := setHome(t)
	reportsDir := filepath.Join(home, ".gotr", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := newShowCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--print", "missing.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}
