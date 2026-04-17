package cmd

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWorkHub_NonInteractive(t *testing.T) {
	cmd := &cobra.Command{Use: "work"}
	cmd.SetContext(context.Background())

	err := runWorkHub(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interactive mode")
}

func TestRunWorkHub_NoServer(t *testing.T) {
	viper.Set("base_url", "")
	defer viper.Set("base_url", "")

	p := interactive.NewMockPrompter()
	ctx := interactive.WithPrompter(context.Background(), p)

	cmd := &cobra.Command{Use: "work"}
	cmd.SetContext(ctx)

	err := runWorkHub(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no server configured")
}

func TestRunWorkHub_ServerConfirmReject(t *testing.T) {
	viper.Set("base_url", "https://test.testrail.io")
	defer viper.Set("base_url", "")

	p := interactive.NewMockPrompter().WithConfirmResponses(false)
	ctx := context.WithValue(context.Background(), serverURLKey, "https://test.testrail.io")
	ctx = interactive.WithPrompter(ctx, p)

	cmd := &cobra.Command{Use: "work"}
	cmd.SetContext(ctx)

	err := runWorkHub(cmd)
	assert.NoError(t, err, "rejecting server should return nil, not error")
}

func TestRunWorkHub_ServerConfirmAccept_ThenExit(t *testing.T) {
	viper.Set("base_url", "https://test.testrail.io")
	defer viper.Set("base_url", "")

	// Accept server confirmation, then Browse returns exit error.
	p := interactive.NewMockPrompter().
		WithConfirmResponses(true).
		WithSelectResponses(interactive.SelectResponse{Index: -1}) // Browse exhausts → error → exit
	ctx := context.WithValue(context.Background(), serverURLKey, "https://test.testrail.io")
	ctx = interactive.WithPrompter(ctx, p)

	cmd := &cobra.Command{Use: "work"}
	cmd.SetContext(ctx)

	err := runWorkHub(cmd)
	assert.NoError(t, err)
}

func TestRunWorkHub_SessionInjected(t *testing.T) {
	viper.Set("base_url", "https://session.testrail.io")
	defer viper.Set("base_url", "")

	// Accept server, then exit on menu.
	p := interactive.NewMockPrompter().
		WithConfirmResponses(true).
		WithSelectResponses(interactive.SelectResponse{Index: -1})
	ctx := context.WithValue(context.Background(), serverURLKey, "https://session.testrail.io")
	ctx = interactive.WithPrompter(ctx, p)

	cmd := &cobra.Command{Use: "work"}
	cmd.SetContext(ctx)

	// We can't inspect the session directly because ctx is modified inside runWorkHub.
	// But we verify no panic and the function runs cleanly.
	err := runWorkHub(cmd)
	assert.NoError(t, err)
}

func TestDispatchWorkGroup_Compare(t *testing.T) {
	ctx := context.Background()
	cmd := &cobra.Command{Use: "root"}
	// No subcommands registered, so dispatch should fail gracefully.
	err := dispatchWorkGroup(ctx, cmd, "compare")
	// compare hub requires a prompter in context — will return error.
	require.Error(t, err)
}

func TestDispatchWorkGroup_UnknownGroup(t *testing.T) {
	ctx := context.Background()
	cmd := &cobra.Command{Use: "root"}
	err := dispatchWorkGroup(ctx, cmd, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestWorkCmd_Properties(t *testing.T) {
	assert.Equal(t, "work", workCmd.Use)
	assert.NotEmpty(t, workCmd.Short)
	assert.NotNil(t, workCmd.RunE)
}

func TestWorkGroups_NotEmpty(t *testing.T) {
	assert.True(t, len(workGroups) >= 5, "should have at least 5 work groups")
	for _, g := range workGroups {
		assert.NotEmpty(t, g.label)
		assert.NotEmpty(t, g.key)
	}
}

func TestPrintWorkHeader_NoPanic(t *testing.T) {
	viper.Set("base_url", "https://test.testrail.io")
	defer viper.Set("base_url", "")
	assert.NotPanics(t, printWorkHeader)
}

func TestPrintWorkHeader_NoPanic_Empty(t *testing.T) {
	viper.Set("base_url", "")
	assert.NotPanics(t, printWorkHeader)
}
