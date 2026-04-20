package sync

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/snap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// newTestCmd creates a cobra command with snap flags registered.
func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	snap.RegisterFlags(cmd)
	return cmd
}

// ==================== confirmSnapshot ====================

func TestConfirmSnapshot_ExplicitFlagTrue(t *testing.T) {
	cmd := newTestCmd()
	_ = cmd.Flags().Set("snapshot", "true")

	// Need Input mock for label prompt.
	p := interactive.NewMockPrompter().WithInputResponses("")
	ctx := interactive.WithPrompter(context.Background(), p)
	result := confirmSnapshot(ctx, cmd)
	assert.True(t, result.Create)
	assert.Empty(t, result.Label)
}

func TestConfirmSnapshot_ExplicitFlagFalse(t *testing.T) {
	cmd := newTestCmd()
	_ = cmd.Flags().Set("snapshot", "false")

	ctx := context.Background()
	result := confirmSnapshot(ctx, cmd)
	assert.False(t, result.Create)
}

func TestConfirmSnapshot_ConfigEnabled(t *testing.T) {
	cmd := newTestCmd()
	viper.Set("snap.enabled", true)
	defer viper.Reset()

	p := interactive.NewMockPrompter().WithInputResponses("")
	ctx := interactive.WithPrompter(context.Background(), p)
	result := confirmSnapshot(ctx, cmd)
	assert.True(t, result.Create)
}

func TestConfirmSnapshot_ConfigDisabled(t *testing.T) {
	cmd := newTestCmd()
	viper.Set("snap.enabled", false)
	defer viper.Reset()

	ctx := context.Background()
	result := confirmSnapshot(ctx, cmd)
	assert.False(t, result.Create)
}

func TestConfirmSnapshot_SmartPrompt_UserAccepts(t *testing.T) {
	cmd := newTestCmd()
	viper.Reset()

	p := interactive.NewMockPrompter().WithConfirmResponses(true).WithInputResponses("")
	ctx := interactive.WithPrompter(context.Background(), p)

	result := confirmSnapshot(ctx, cmd)
	assert.True(t, result.Create)
}

func TestConfirmSnapshot_SmartPrompt_UserDeclines(t *testing.T) {
	cmd := newTestCmd()
	viper.Reset()

	p := interactive.NewMockPrompter().WithConfirmResponses(false)
	ctx := interactive.WithPrompter(context.Background(), p)

	result := confirmSnapshot(ctx, cmd)
	assert.False(t, result.Create)
}

func TestConfirmSnapshot_NonInteractive_DefaultsTrue(t *testing.T) {
	cmd := newTestCmd()
	viper.Reset()

	p := interactive.NewNonInteractivePrompter()
	ctx := interactive.WithPrompter(context.Background(), p)

	result := confirmSnapshot(ctx, cmd)
	assert.True(t, result.Create, "Non-interactive mode should default to creating snapshot")
}

func TestConfirmSnapshot_LabelFromPrompt(t *testing.T) {
	cmd := newTestCmd()
	viper.Reset()

	p := interactive.NewMockPrompter().WithConfirmResponses(true).WithInputResponses("migration-v2")
	ctx := interactive.WithPrompter(context.Background(), p)

	result := confirmSnapshot(ctx, cmd)
	assert.True(t, result.Create)
	assert.Equal(t, "migration-v2", result.Label)
}

func TestConfirmSnapshot_LabelFromFlag(t *testing.T) {
	cmd := newTestCmd()
	_ = cmd.Flags().Set("snapshot", "true")
	_ = cmd.Flags().Set("snap-label", "my-label")

	// Input not called when label comes from flag.
	ctx := context.Background()
	result := confirmSnapshot(ctx, cmd)
	assert.True(t, result.Create)
	assert.Equal(t, "my-label", result.Label)
}

// ==================== syncPostAction ====================

func TestSyncPostAction_NilHook(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCmd()

	// Should not panic with nil hook.
	syncPostAction(ctx, cmd, nil, nil)
}

func TestSyncPostAction_DisabledHook(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCmd()
	hook := &snap.Hook{Enabled: false}

	// Should exit immediately for disabled hook.
	syncPostAction(ctx, cmd, hook, nil)
}

func TestSyncPostAction_Exit(t *testing.T) {
	cmd := newTestCmd()

	// Mock: select "Exit" (index 0).
	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0})
	ctx := interactive.WithPrompter(context.Background(), p)

	hook := &snap.Hook{
		Enabled: true,
		Snap: &snap.Snapshot{
			Meta: &snap.Meta{ID: "test-snap-123"},
		},
	}

	// Should not panic.
	syncPostAction(ctx, cmd, hook, nil)
}
