package output

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingStringValue struct{}

func (f *failingStringValue) String() string { return "" }

func (f *failingStringValue) Set(string) error { return nil }

func (f *failingStringValue) Type() string { return "failing-string" }

func TestResolveSavePathFromFlags_SaveTo(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save-to", "", "")
	cmd.Flags().Bool("save", false, "")
	cmd.SetArgs([]string{"--save-to=out.json"})
	cmd.Execute() //nolint: errcheck

	savePath, explicit, err := ResolveSavePathFromFlags(cmd)
	require.NoError(t, err)
	assert.True(t, explicit)
	assert.Equal(t, "out.json", savePath)
}

func TestResolveSavePathFromFlags_SaveDefault(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save-to", "", "")
	cmd.Flags().Bool("save", false, "")
	cmd.SetArgs([]string{"--save"})
	cmd.Execute() //nolint: errcheck

	savePath, explicit, err := ResolveSavePathFromFlags(cmd)
	require.NoError(t, err)
	assert.True(t, explicit)
	assert.Equal(t, "__DEFAULT__", savePath)
}

func TestResolveSavePathFromFlags_None(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save-to", "", "")
	cmd.Flags().Bool("save", false, "")

	savePath, explicit, err := ResolveSavePathFromFlags(cmd)
	require.NoError(t, err)
	assert.False(t, explicit)
	assert.Equal(t, "", savePath)
}

func TestResolveSavePathFromFlags_SaveToReadError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Var(&failingStringValue{}, "save-to", "")
	cmd.Flags().Bool("save", false, "")
	require.NoError(t, cmd.Flags().Set("save-to", "any"))

	savePath, explicit, err := ResolveSavePathFromFlags(cmd)
	require.Error(t, err)
	assert.False(t, explicit)
	assert.Equal(t, "", savePath)
	assert.Contains(t, err.Error(), "failed to read --save-to flag")
}

func TestResolveSavePathFromFlags_SaveChangedFalseStillExplicit(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("save-to", "", "")
	cmd.Flags().Bool("save", false, "")
	require.NoError(t, cmd.Flags().Set("save", "false"))

	savePath, explicit, err := ResolveSavePathFromFlags(cmd)
	require.NoError(t, err)
	assert.True(t, explicit)
	assert.Equal(t, "__DEFAULT__", savePath)
}

func TestPromptSavePath_DoNotSave(t *testing.T) {
	mp := interactive.NewMockPrompter().WithConfirmResponses(false)

	savePath, err := PromptSavePath(mp, "compare result")
	require.NoError(t, err)
	assert.Equal(t, "", savePath)
}

func TestPromptSavePath_DefaultSave(t *testing.T) {
	mp := interactive.NewMockPrompter().WithConfirmResponses(true, true)

	savePath, err := PromptSavePath(mp, "compare result")
	require.NoError(t, err)
	assert.Equal(t, "__DEFAULT__", savePath)
}

func TestPromptSavePath_CustomSave(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithConfirmResponses(true, false).
		WithInputResponses(" custom.json ")

	savePath, err := PromptSavePath(mp, "compare result")
	require.NoError(t, err)
	assert.Equal(t, "custom.json", savePath)
}

func TestPromptSavePath_EmptyCustomPath(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithConfirmResponses(true, false).
		WithInputResponses("   ")

	savePath, err := PromptSavePath(mp, "compare result")
	require.Error(t, err)
	assert.Equal(t, "", savePath)
	assert.Contains(t, err.Error(), "empty")
}

func TestPromptSavePath_NonInteractiveError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	p := interactive.PrompterFromContext(ctx)

	savePath, err := PromptSavePath(p, "compare result")
	require.Error(t, err)
	assert.Equal(t, "", savePath)
	assert.Contains(t, err.Error(), "non-interactive")
}

func TestPromptSavePathWithOptions_NoCustomPathFallsBackToDefault(t *testing.T) {
	mp := interactive.NewMockPrompter().WithConfirmResponses(true, false)

	savePath, err := PromptSavePathWithOptions(mp, "result", false)
	require.NoError(t, err)
	assert.Equal(t, "__DEFAULT__", savePath)
}

func TestPromptSavePathWithOptions_EmptySubjectCustomPath(t *testing.T) {
	mp := interactive.NewMockPrompter().
		WithConfirmResponses(true, false).
		WithInputResponses(" report.json ")

	savePath, err := PromptSavePathWithOptions(mp, "", true)
	require.NoError(t, err)
	assert.Equal(t, "report.json", savePath)
}

func TestShouldPromptForInteractiveSave_True(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()))

	assert.True(t, ShouldPromptForInteractiveSave(cmd))
}

func TestShouldPromptForInteractiveSave_FalseWhenNonInteractive(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("non-interactive", "true"))
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()))

	assert.False(t, ShouldPromptForInteractiveSave(cmd))
}

func TestShouldPromptForInteractiveSave_FalseWhenQuiet(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()))

	assert.False(t, ShouldPromptForInteractiveSave(cmd))
}

func TestShouldPromptForInteractiveSave_FalseForNilCommand(t *testing.T) {
	assert.False(t, ShouldPromptForInteractiveSave(nil))
}

func TestShouldPromptForInteractiveSave_FalseWithoutPrompter(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")

	assert.False(t, ShouldPromptForInteractiveSave(cmd))
}

func TestShouldPromptForInteractiveSave_TrueWhenFlagsMissingButPrompterPresent(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().AddFlag(&pflag.Flag{Name: "quiet", Value: &failingStringValue{}})
	cmd.Flags().AddFlag(&pflag.Flag{Name: "non-interactive", Value: &failingStringValue{}})
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()))

	assert.True(t, ShouldPromptForInteractiveSave(cmd))
}

func TestPromptSavePathWithOptions_FirstConfirmError(t *testing.T) {
	mp := interactive.NewMockPrompter()

	savePath, err := PromptSavePathWithOptions(mp, "result", true)
	require.Error(t, err)
	assert.Equal(t, "", savePath)
	assert.Contains(t, err.Error(), "interactive save selection failed")
}

func TestPromptSavePathWithOptions_SecondConfirmError(t *testing.T) {
	mp := interactive.NewMockPrompter().WithConfirmResponses(true)

	savePath, err := PromptSavePathWithOptions(mp, "result", true)
	require.Error(t, err)
	assert.Equal(t, "", savePath)
	assert.Contains(t, err.Error(), "interactive save path selection failed")
}

func TestPromptSavePathWithOptions_InputError(t *testing.T) {
	mp := interactive.NewMockPrompter().WithConfirmResponses(true, false)

	savePath, err := PromptSavePathWithOptions(mp, "result", true)
	require.Error(t, err)
	assert.Equal(t, "", savePath)
	assert.Contains(t, err.Error(), "interactive output path input failed")
}
