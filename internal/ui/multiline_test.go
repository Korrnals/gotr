// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMultilineStatus_NonTTYNoOpsByDefault(t *testing.T) {
	var buf bytes.Buffer
	live := NewMultilineStatus(MultilineConfig{Writer: &buf})
	if live.Active() {
		t.Fatal("Active() = true on bytes.Buffer; want false")
	}
	live.Update(Snapshot{Title: "x"})
	if buf.Len() != 0 {
		t.Fatalf("non-tty writer received bytes: %q", buf.String())
	}
}

func TestMultilineStatus_QuietForcesInactive(t *testing.T) {
	var buf bytes.Buffer
	live := NewMultilineStatus(MultilineConfig{Writer: &buf, ForceTTY: true, Quiet: true})
	if live.Active() {
		t.Fatal("Active() = true with Quiet=true; want false")
	}
}

func TestRenderForTest_FormatsBlock(t *testing.T) {
	live := NewMultilineStatus(MultilineConfig{Writer: nil})
	live.Update(Snapshot{
		Title:       "Scanning attachments",
		Strategy:    "entity",
		ProjectIdx:  2,
		ProjectN:    5,
		ProjectName: "Demo project",
		Phase:       "cases",
		PhaseDone:   3,
		PhaseTotal:  10,
		Found:       7,
		Eligible:    5,
		Bytes:       1024 * 1024,
		Elapsed:     12 * time.Second,
		ETA:         18 * time.Second,
	})
	out := live.RenderForTest(live.snap)
	if !strings.Contains(out, "Scanning attachments") {
		t.Errorf("missing title: %q", out)
	}
	if !strings.Contains(out, "entity") {
		t.Errorf("missing strategy: %q", out)
	}
	if !strings.Contains(out, "2/5") {
		t.Errorf("missing project N/M: %q", out)
	}
	if !strings.Contains(out, "Demo project") {
		t.Errorf("missing project name: %q", out)
	}
	if !strings.Contains(out, "cases") {
		t.Errorf("missing phase name: %q", out)
	}
	if !strings.Contains(out, "3 / 10") {
		t.Errorf("missing phase progress: %q", out)
	}
	if !strings.Contains(out, "1.00 MiB") {
		t.Errorf("missing humanized bytes: %q", out)
	}
}

func TestHumanBytes_Boundaries(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.00 KiB (1024 B)"},
		{1024 * 1024, "1.00 MiB (1048576 B)"},
		{1024 * 1024 * 1024, "1.00 GiB (1073741824 B)"},
		{1024 * 1024 * 1024 * 1024, "1.00 TiB (1099511627776 B)"},
	}
	for _, tc := range cases {
		if got := HumanBytes(tc.in); got != tc.want {
			t.Errorf("HumanBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
