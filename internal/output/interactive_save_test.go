package output

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
