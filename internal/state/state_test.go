package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Missing_ReturnsZero(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.FlatLayoutWarned {
		t.Errorf("expected zero FlatLayoutWarned, got true")
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := Save(&State{FlatLayoutWarned: true}); err != nil {
		t.Fatal(err)
	}
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(p) != filepath.Join(home, ".gotr") {
		t.Errorf("unexpected state path dir: %s", p)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("state file missing: %v", err)
	}

	s, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !s.FlatLayoutWarned {
		t.Error("expected FlatLayoutWarned=true after round trip")
	}
}

func TestLoad_Malformed_ReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".gotr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Error("expected error for malformed state.json")
	}
}

func TestLoad_Empty_OK(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".gotr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.FlatLayoutWarned {
		t.Error("expected zero state from empty file")
	}
}
