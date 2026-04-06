// cmd/completion.go

package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			_ = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
