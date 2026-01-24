// cmd/completion.go (или в root.go)

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd — служебная команда для генерации автодополнения
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Генерирует скрипт автодополнения для указанной оболочки.

Примеры:
  source <(gotr completion bash)                  # временно для текущей сессии
  gotr completion bash > /usr/local/etc/bash_completion.d/gotr  # навсегда (macOS/Linux)

Zsh:
  gotr completion zsh > "${fpath[1]}/_gotr"

Fish:
  gotr completion fish > ~/.config/fish/completions/gotr.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	// ОТКЛЮЧАЕМ PersistentPreRunE — клиент не нужен!
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Ничего не делаем — переопределяем родительский
	},
	// Полностью отключаем всё, что делает rootCmd (клиент + вывод Viper)
	PersistentPreRunE: nil, // или PersistentPreRun: func() {}
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

func init() {
	rootCmd.AddCommand(completionCmd)
}
