package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func withStdoutToDevNull(t *testing.T, fn func()) {
	t.Helper()
	old := os.Stdout
	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open devnull error: %v", err)
	}
	os.Stdout = nullFile

	fn()

	_ = nullFile.Close()
	os.Stdout = old
}

func TestCompletionCmd_Run_Shells(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			withStdoutToDevNull(t, func() {
				err := completionCmd.RunE(completionCmd, []string{shell})
				assert.NoError(t, err)
			})
		})
	}
}

func TestCompletionCmd_Metadata(t *testing.T) {
	assert.Equal(t, "completion [bash|zsh|fish|powershell]", completionCmd.Use)
	assert.Contains(t, completionCmd.ValidArgs, "bash")
	assert.Contains(t, completionCmd.ValidArgs, "zsh")
	assert.Contains(t, completionCmd.ValidArgs, "fish")
	assert.Contains(t, completionCmd.ValidArgs, "powershell")
	assert.NotNil(t, completionCmd.Args)
}
