// cmd/completion.go

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generates a shell completion script for the specified shell.

Examples:
	source <(gotr completion bash)                  # temporary for current session
	gotr completion bash > /usr/local/etc/bash_completion.d/gotr  # persistent (macOS/Linux)

Zsh:
	gotr completion zsh > "${fpath[1]}/_gotr"

Fish:
	gotr completion fish > ~/.config/fish/completions/gotr.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	// Disable PersistentPreRunE — client not needed.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent PersistentPreRunE (no-op)
	},
	PersistentPreRunE: nil,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
		if err != nil {
			return fmt.Errorf("failed to generate %s completion: %w", args[0], err)
		}
		return nil
	},
}
