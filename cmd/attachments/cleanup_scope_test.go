// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestEntityTypesCoversAll(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want bool
	}{
		{"nil treated as ALL", nil, true},
		{"empty treated as ALL", []string{}, true},
		{"all kinds explicit", append([]string(nil), validCleanupEntityTypes...), true},
		{"all kinds shuffled+upper", []string{"TEST", "result", "Plan_Entry", "plan", "run", "case"}, true},
		{"missing one", []string{"case", "run", "plan", "plan_entry", "result"}, false},
		{"single kind", []string{"result"}, false},
		{"two kinds", []string{"case", "run"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := entityTypesCoversAll(c.in); got != c.want {
				t.Fatalf("entityTypesCoversAll(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestPrintEntityScopeNotice_AllKindsWarning(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	printEntityScopeNotice(cmd, &cleanupOptions{
		EntityTypes: append([]string(nil), validCleanupEntityTypes...),
	})
	out := buf.String()
	if !strings.Contains(out, "⚠️") {
		t.Fatalf("expected warning glyph in:\n%s", out)
	}
	if !strings.Contains(out, "ALL attachment kinds") {
		t.Fatalf("expected ALL-scope wording in:\n%s", out)
	}
	if !strings.Contains(out, "--entity-type") {
		t.Fatalf("expected hint about --entity-type in:\n%s", out)
	}
}

func TestPrintEntityScopeNotice_NarrowScopeInfo(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	printEntityScopeNotice(cmd, &cleanupOptions{
		EntityTypes: []string{"result", "case"},
	})
	out := buf.String()
	if strings.Contains(out, "⚠️") {
		t.Fatalf("narrow scope must not warn:\n%s", out)
	}
	if !strings.Contains(out, "ℹ️") {
		t.Fatalf("expected info glyph:\n%s", out)
	}
	if !strings.Contains(out, "case, result") {
		t.Fatalf("expected sorted scope list 'case, result' in:\n%s", out)
	}
}
