package interactive

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsGoBack(t *testing.T) {
	assert.True(t, IsGoBack(ErrGoBack))
	assert.False(t, IsGoBack(ErrExit))
	assert.False(t, IsGoBack(nil))
}

func TestIsExit(t *testing.T) {
	assert.True(t, IsExit(ErrExit))
	assert.False(t, IsExit(ErrGoBack))
	assert.False(t, IsExit(nil))
}

func TestFindSubcommand_Found(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{Use: "child"}
	grandchild := &cobra.Command{Use: "grand"}
	child.AddCommand(grandchild)
	root.AddCommand(child)

	cmd, err := FindSubcommand(root, "child", "grand")
	require.NoError(t, err)
	assert.Equal(t, "grand", cmd.Name())
}

func TestFindSubcommand_NotFound(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	_, err := FindSubcommand(root, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find 'gotr nonexistent'")
}

func TestRunSubcommand_Success(t *testing.T) {
	called := false
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{
		Use: "child",
		RunE: func(cmd *cobra.Command, args []string) error {
			called = true
			return nil
		},
	}
	root.AddCommand(child)

	err := RunSubcommand(context.Background(), root, "child")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRunSubcommand_GoBackAbsorbed(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{
		Use: "child",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ErrGoBack
		},
	}
	root.AddCommand(child)

	err := RunSubcommand(context.Background(), root, "child")
	assert.NoError(t, err, "GoBack should be absorbed")
}

func TestRunSubcommand_ExitAbsorbed(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{
		Use: "child",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ErrExit
		},
	}
	root.AddCommand(child)

	err := RunSubcommand(context.Background(), root, "child")
	assert.NoError(t, err, "Exit should be absorbed")
}

func TestRunSubcommand_RealError(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{
		Use: "child",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("boom")
		},
	}
	root.AddCommand(child)

	err := RunSubcommand(context.Background(), root, "child")
	assert.EqualError(t, err, "boom")
}

func TestRunSubcommand_NotFound(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	err := RunSubcommand(context.Background(), root, "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find")
}
