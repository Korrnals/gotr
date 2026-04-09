package output

import (
	"context"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TestOutputGetResult_QuietMode_Extended tests OutputGetResult skips saving in quiet mode
func TestOutputGetResult_QuietMode_Extended(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Set("quiet", "true")

	data := map[string]any{"test": "data"}
	err := OutputGetResult(cmd, data, time.Now())

	if err != nil {
		t.Fatalf("OutputGetResult quiet error: %v", err)
	}
}

// TestOutputGetResult_SaveFlagEnabled_Extended tests OutputGetResult with save flag
func TestOutputGetResult_SaveFlagEnabled_Extended(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Set("save", "true")

	data := map[string]any{"result": "success"}
	err := OutputGetResult(cmd, data, time.Now())

	if err != nil {
		t.Fatalf("OutputGetResult with save error: %v", err)
	}
}

// TestOutputGetResult_BodyOnly_Extended tests OutputGetResult with body-only flag
func TestOutputGetResult_BodyOnly_Extended(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Set("save", "true")
	cmd.Flags().Set("body-only", "true")

	data := map[string]any{"data": "only"}
	err := OutputGetResult(cmd, data, time.Now().Add(-time.Minute))

	if err != nil {
		t.Fatalf("OutputGetResult body-only error: %v", err)
	}
}

// TestOutputGetResult_JQEnabled_Extended tests OutputGetResult with jq enabled
func TestOutputGetResult_JQEnabled_Extended(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Set("jq", "true")

	viper.Set("jq_format", false)
	defer viper.Reset()

	data := map[string]any{"jq": "enabled"}
	err := OutputGetResult(cmd, data, time.Now())

	if err != nil {
		t.Fatalf("OutputGetResult jq error: %v", err)
	}
}

// TestPromptSavePathWithOptions_DefaultPath tests with default path preference
func TestPromptSavePathWithOptions_DefaultPath(t *testing.T) {
	mp := interactive.NewMockPrompter().WithConfirmResponses(true, true)

	path, err := PromptSavePathWithOptions(mp, "test", false)
	if err != nil {
		t.Fatalf("PromptSavePathWithOptions error: %v", err)
	}
	if path != defaultSavePathMarker {
		t.Fatalf("expected default marker, got %q", path)
	}
}

// TestPromptSavePathWithOptions_CustomPath tests with custom path input
func TestPromptSavePathWithOptions_CustomPath(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithConfirmResponses(true, false).
		WithInputResponses("output.json")

	path, err := PromptSavePathWithOptions(mp, "result", true)
	if err != nil {
		t.Fatalf("PromptSavePathWithOptions error: %v", err)
	}
	if path != "output.json" {
		t.Fatalf("expected 'output.json', got %q", path)
	}
}

// TestPromptSavePathWithOptions_NoSave tests when user chooses not to save
func TestPromptSavePathWithOptions_NoSave(t *testing.T) {
	mp := interactive.NewMockPrompter().WithConfirmResponses(false)

	path, err := PromptSavePathWithOptions(mp, "data", false)
	if err != nil {
		t.Fatalf("PromptSavePathWithOptions error: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path when declining save, got %q", path)
	}
}

// TestResolveSavePathFromFlags_SaveTo_Extended tests save-to flag resolution
func TestResolveSavePathFromFlags_SaveTo_Extended(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save-to", "", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Set("save-to", "/custom/path.json")

	path, explicit, err := ResolveSavePathFromFlags(cmd)
	if err != nil {
		t.Fatalf("ResolveSavePathFromFlags error: %v", err)
	}
	if !explicit {
		t.Fatalf("expected explicit=true")
	}
	if path != "/custom/path.json" {
		t.Fatalf("expected '/custom/path.json', got %q", path)
	}
}

// TestResolveSavePathFromFlags_Save_Extended tests save flag resolution
func TestResolveSavePathFromFlags_Save_Extended(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save-to", "", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Set("save", "true")

	path, explicit, err := ResolveSavePathFromFlags(cmd)
	if err != nil {
		t.Fatalf("ResolveSavePathFromFlags error: %v", err)
	}
	if !explicit {
		t.Fatalf("expected explicit=true")
	}
	if path != defaultSavePathMarker {
		t.Fatalf("expected default marker, got %q", path)
	}
}

// TestShouldPromptForInteractiveSave_True_Extended tests conditions for prompting
func TestShouldPromptForInteractiveSave_True_Extended(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	cmd.SetContext(ctx)

	if !ShouldPromptForInteractiveSave(cmd) {
		t.Fatalf("should prompt for interactive save")
	}
}

// TestShouldPromptForInteractiveSave_QuietMode_Extended tests no prompt in quiet mode
func TestShouldPromptForInteractiveSave_QuietMode_Extended(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.Flags().Set("quiet", "true")
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	cmd.SetContext(ctx)

	if ShouldPromptForInteractiveSave(cmd) {
		t.Fatalf("should not prompt in quiet mode")
	}
}

// TestShouldPromptForInteractiveSave_NonInteractive_Extended tests no prompt in non-interactive
func TestShouldPromptForInteractiveSave_NonInteractive_Extended(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.Flags().Set("non-interactive", "true")
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	cmd.SetContext(ctx)

	if ShouldPromptForInteractiveSave(cmd) {
		t.Fatalf("should not prompt in non-interactive mode")
	}
}

// TestIsSkippableInteractiveSavePromptError_Match tests error detection
func TestIsSkippableInteractiveSavePromptError_Match(t *testing.T) {
	tests := []struct {
		name  string
		msg   string
		match bool
	}{
		{"confirm", "mock confirm queue exhausted", true},
		{"input", "mock input queue exhausted", true},
		{"mixed case", "Mock Confirm Queue Exhausted", true},
		{"other", "some other error", false},
		{"nil", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.msg != "" {
				err = testError(tt.msg)
			}
			result := isSkippableInteractiveSavePromptError(err)
			if result != tt.match {
				t.Fatalf("expected %v, got %v for %q", tt.match, result, tt.msg)
			}
		})
	}
}

// Helper type for testing error matching
type testError string

func (e testError) Error() string {
	return string(e)
}

// TestOutputResult tests OutputResult wrapper function
func TestOutputResult(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("save", false, "")

	data := map[string]any{"status": "ok"}
	err := OutputResult(cmd, data, "test")

	if err != nil {
		t.Fatalf("OutputResult error: %v", err)
	}
}
