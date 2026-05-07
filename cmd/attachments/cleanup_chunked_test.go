// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/cleanup/checkpoint"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// TestCleanupCmd_ListCheckpoints_Empty verifies the no-checkpoints
// branch returns exit-code 0 and prints the friendly empty message.
func TestCleanupCmd_ListCheckpoints_Empty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	getClient := func(_ *cobra.Command) client.ClientInterface { return nil }
	cmd := newCleanupCmd(getClient)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--list-checkpoints"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "No cleanup checkpoints") {
		t.Fatalf("expected empty-message, got %q", out)
	}
}

// TestCleanupCmd_ListCheckpoints_PrintsTable seeds a checkpoint and
// asserts the table contains the run id and the per-state counters.
func TestCleanupCmd_ListCheckpoints_PrintsTable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := checkpoint.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	cp := &checkpoint.Checkpoint{
		RunID:     "20260505T120000-abcdef",
		StartedAt: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
		Projects: []checkpoint.ProjectStatus{
			{ID: 1, State: checkpoint.StateDone},
			{ID: 2, State: checkpoint.StateFailed},
			{ID: 3, State: checkpoint.StatePending},
		},
	}
	if err := store.Save(cp.RunID, cp); err != nil {
		t.Fatalf("Save: %v", err)
	}

	getClient := func(_ *cobra.Command) client.ClientInterface { return nil }
	cmd := newCleanupCmd(getClient)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--list-checkpoints"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "20260505T120000-abcdef") {
		t.Fatalf("output missing run id: %q", out)
	}
	if !strings.Contains(out, "RUN_ID") {
		t.Fatalf("output missing header: %q", out)
	}
	if !strings.Contains(out, "Resume an interrupted run") {
		t.Fatalf("output missing resume hint: %q", out)
	}
}

// TestCleanupCmd_ListCheckpoints_RejectsOtherFlags asserts the
// mutual-exclusion gate.
func TestCleanupCmd_ListCheckpoints_RejectsOtherFlags(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	getClient := func(_ *cobra.Command) client.ClientInterface { return nil }
	cmd := newCleanupCmd(getClient)
	cmd.SetArgs([]string{"--list-checkpoints", "--all-projects"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatalf("expected mutex error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %v, want mutex hint", err)
	}
}
