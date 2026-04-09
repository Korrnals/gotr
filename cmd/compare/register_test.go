package compare

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_AttachesCommandWithFlagsAndSubcommands(t *testing.T) {
	origCmd := Cmd
	origGetClient := getClient
	t.Cleanup(func() {
		Cmd = origCmd
		getClient = origGetClient
	})

	root := &cobra.Command{Use: "root"}

	mockFn := func(cmd *cobra.Command) client.ClientInterface {
		_ = cmd
		return &client.MockClient{}
	}

	Register(root, mockFn)

	require.NotNil(t, Cmd)
	assert.Equal(t, "compare", Cmd.Use)
	if assert.NotNil(t, getClient) {
		got := getClient(&cobra.Command{})
		assert.IsType(t, &client.MockClient{}, got)
	}

	assert.NotNil(t, root.Commands())
	assert.Equal(t, Cmd, root.Commands()[0])

	for _, name := range []string{
		"pid1", "pid2", "save", "save-to", "rate-limit",
		"parallel-suites", "parallel-pages", "page-retries",
		"timeout", "retry-attempts", "retry-workers", "retry-delay",
	} {
		assert.NotNil(t, Cmd.PersistentFlags().Lookup(name), "missing persistent flag: %s", name)
	}

	subNames := map[string]bool{}
	for _, c := range Cmd.Commands() {
		subNames[c.Name()] = true
	}

	for _, expected := range []string{
		"cases", "suites", "sections", "sharedsteps", "runs", "plans",
		"milestones", "datasets", "groups", "labels", "templates",
		"configurations", "retry-failed-pages", "all",
	} {
		assert.True(t, subNames[expected], "missing subcommand: %s", expected)
	}
}

func TestSetGetClientForTests_OverridesGetter(t *testing.T) {
	orig := getClient
	t.Cleanup(func() { getClient = orig })

	expected := &client.MockClient{}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		_ = cmd
		return expected
	})

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	got := getClient(cmd)
	assert.Equal(t, expected, got)
}
