package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/Korrnals/gotr/internal/models/config"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sensitiveConfigLine = regexp.MustCompile(`(?mi)^(\s*(api_key|password|token|authorization)\s*:\s*)([^\n\r]*)$`)

// configCmd is the parent "config" command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gotr configuration",
	Long:  `Commands for managing the gotr configuration file.`,
	// Disable PersistentPreRunE for the entire config branch.
	// This prevents client creation and mandatory flag checks.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent PersistentPreRunE (no-op)
	},
}

// configInitCmd creates a default configuration file.
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Long: `Creates a default configuration file at ~/.gotr/config/default.yaml.

Example:
	gotr config init

Default config path:
	$HOME/.gotr/config/default.yaml

Note:
	After creation, edit the file to fill in your TestRail credentials.`,

	// Disable PersistentPreRunE for the config branch.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent PersistentPreRunE (no-op)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return err
		}
		if err := cfg.WithDefaults().Create(); err != nil {
			return err
		}
		ui.Infof(os.Stdout, "Config file created: %s", cfg.Path)
		return nil
	},
}

// configPathCmd shows the current config file path.
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show the current config file path",
	Long:  `Prints the path to the currently used configuration file.`,

	// Disable PersistentPreRunE for the config branch.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent PersistentPreRunE (no-op)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return fmt.Errorf("failed to determine config path: %w", err)
		}

		used := viper.ConfigFileUsed()
		if used == "" {
			ui.Warningf(os.Stdout, "Config file not found.\nExpected location: %s", cfg.PathString())
			ui.Info(os.Stdout, "Create it with: gotr config init")
		} else {
			ui.Infof(os.Stdout, "Current config file: %s", used)
		}
		return nil
	},
}

// configViewCmd displays the current config file contents.
var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show the current config file contents",
	Long:  "Prints the configuration file contents in a readable format.",

	// Disable PersistentPreRunE for the config branch.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent PersistentPreRunE (no-op)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		used := viper.ConfigFileUsed()
		if used == "" {
			cfg, _ := config.Default()
			ui.Warningf(os.Stdout, "Config file not found: %s", cfg.PathString())
			ui.Info(os.Stdout, "Create it with: gotr config init")
			return nil
		}

		data, err := os.ReadFile(used)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		ui.Infof(os.Stdout, "Config file contents %s:\n\n%s", used, redactSensitiveConfig(string(data)))
		return nil
	},
}

func redactSensitiveConfig(content string) string {
	return sensitiveConfigLine.ReplaceAllString(content, `${1}"***"`)
}

// configEditCmd opens the config file in the default editor.
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config file in the default editor",
	Long: `Opens the current configuration file in the editor specified by the EDITOR environment variable.
If EDITOR is not set, falls back to vi/nano on Linux, notepad on Windows.

Examples:
	export EDITOR=code    # VS Code
	export EDITOR=nano
	gotr config edit      # opens in the specified editor`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Override parent PreRun (client not needed)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine config file path
		used := viper.ConfigFileUsed()
		if used == "" {
			cfg, err := config.Default()
			if err != nil {
				return fmt.Errorf("failed to determine config path: %w", err)
			}
			ui.Warningf(os.Stdout, "Config file not found: %s", cfg.PathString())
			ui.Info(os.Stdout, "Create it with: gotr config init")
			return nil
		}

		// Launch editor
		if err := ui.OpenEditor(used); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		ui.Infof(os.Stdout, "Config file opened in editor: %s", used)
		return nil
	},
}
