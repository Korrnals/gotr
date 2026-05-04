// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"context"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// newCleanupCmdForTest builds a cobra command with the cleanup flag set
// (via newCleanupCmd's GetClient injection) so we can drive
// promptCleanupOptions in isolation. The fake client constructor is
// never called because the prompts run before any API access.
func newCleanupCmdForTest() *cobra.Command {
	return newCleanupCmd(func(*cobra.Command) client.ClientInterface { return nil })
}

func TestPromptCleanupOptions_NoOpWithoutPrompter(t *testing.T) {
	cmd := newCleanupCmdForTest()
	opts := &cleanupOptions{EntityTypes: []string{"result"}, Concurrency: 4}
	if err := promptCleanupOptions(context.Background(), cmd, opts); err != nil {
		t.Fatalf("expected no-op, got %v", err)
	}
	if opts.AllProjects || len(opts.ProjectIDs) != 0 || opts.OlderThan != 0 {
		t.Fatalf("opts should be untouched: %+v", opts)
	}
}

func TestPromptCleanupOptions_AllProjectsHappyPath(t *testing.T) {
	mock := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Value: "All visible projects"}).
		WithInputResponses(
			"result",     // entity types
			"3M",         // older-than
			"6",          // concurrency
			"7d",         // snapshot retention
		).
		WithConfirmResponses(
			true,  // take snapshot
			false, // dry-run
		)
	ctx := interactive.WithPrompter(context.Background(), mock)

	cmd := newCleanupCmdForTest()
	opts := &cleanupOptions{EntityTypes: []string{"result"}, Concurrency: 4}

	if err := promptCleanupOptions(ctx, cmd, opts); err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if !opts.AllProjects {
		t.Fatalf("AllProjects must be true")
	}
	if len(opts.ProjectIDs) != 0 {
		t.Fatalf("ProjectIDs must remain empty, got %v", opts.ProjectIDs)
	}
	if opts.OlderThan != 90*24*time.Hour {
		t.Fatalf("OlderThan should be 3M (~90d), got %v", opts.OlderThan)
	}
	if opts.Concurrency != 6 {
		t.Fatalf("Concurrency want 6, got %d", opts.Concurrency)
	}
	if opts.SkipSnapshot {
		t.Fatalf("SkipSnapshot must be false (snapshot accepted)")
	}
	if opts.SnapshotRetention != 7*24*time.Hour {
		t.Fatalf("SnapshotRetention want 7d, got %v", opts.SnapshotRetention)
	}
	if opts.DryRun {
		t.Fatalf("DryRun must be false")
	}
}

func TestPromptCleanupOptions_SpecificProjects(t *testing.T) {
	mock := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1, Value: "Specific project IDs"}).
		WithInputResponses(
			"42, 7,  9", // project IDs
			"result,run", // entity types
			"30d",        // older-than
			"",           // concurrency (keep default)
			"7d",         // retention
		).
		WithConfirmResponses(true, false)
	ctx := interactive.WithPrompter(context.Background(), mock)

	cmd := newCleanupCmdForTest()
	opts := &cleanupOptions{EntityTypes: []string{"result"}, Concurrency: 4}

	if err := promptCleanupOptions(ctx, cmd, opts); err != nil {
		t.Fatalf("prompt: %v", err)
	}
	want := []int64{7, 9, 42}
	if len(opts.ProjectIDs) != len(want) {
		t.Fatalf("project ids: %v", opts.ProjectIDs)
	}
	for i, v := range want {
		if opts.ProjectIDs[i] != v {
			t.Fatalf("project ids[%d]: got %d want %d", i, opts.ProjectIDs[i], v)
		}
	}
	if len(opts.EntityTypes) != 2 || opts.EntityTypes[0] != "result" || opts.EntityTypes[1] != "run" {
		t.Fatalf("entity types: %v", opts.EntityTypes)
	}
	if opts.Concurrency != 4 {
		t.Fatalf("concurrency: %d", opts.Concurrency)
	}
}

func TestParseEntityTypeList_Validation(t *testing.T) {
	if _, err := parseEntityTypeList("bogus"); err == nil {
		t.Fatalf("expected validation error for bogus type")
	}
	got, err := parseEntityTypeList("")
	if err != nil {
		t.Fatalf("empty input err: %v", err)
	}
	if len(got) != 1 || got[0] != "result" {
		t.Fatalf("empty input default: %v", got)
	}
}

func TestParseInt64List_DedupAndSort(t *testing.T) {
	got, err := parseInt64List("3, 1, 1, 2")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := []int64{1, 2, 3}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestConfirmCleanupExecution_ForceSkipsPrompt(t *testing.T) {
	ok, err := confirmCleanupExecution(context.Background(), &cleanupOptions{Force: true})
	if err != nil || !ok {
		t.Fatalf("force should auto-confirm: ok=%v err=%v", ok, err)
	}
}

func TestConfirmCleanupExecution_PrompterDecline(t *testing.T) {
	mock := interactive.NewMockPrompter().WithConfirmResponses(false)
	ctx := interactive.WithPrompter(context.Background(), mock)
	ok, err := confirmCleanupExecution(ctx, &cleanupOptions{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ok {
		t.Fatalf("decline must yield ok=false")
	}
}
