package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/debug"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/config"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version is populated at build time via -ldflags.
	Version = "3.0.1" // default value for local development
	Commit  = "unknown"
	Date    = "unknown"
	userHomeDir = os.UserHomeDir
	// processExit is the exit function; overridable in tests.
	processExit = os.Exit
)

// rootCmd is the top-level command: gotr
var rootCmd = &cobra.Command{
	Use:   "gotr",
	Short: "CLI client for TestRail API",
	Long: `gotr is a convenient CLI for working with TestRail API v2.
Supports browsing available endpoints, executing requests, and more.`,
	// PersistentPreRunE initializes the HTTP client before every subcommand.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		debug.DebugPrint("{rootCmd} - Running command: %s", cmd.Use)
		debug.DebugPrint("{rootCmd} - Arguments: %v", args)
		quiet, _ := cmd.Flags().GetBool("quiet")
		ui.SetMessageQuiet(quiet)

		// Set up Viper: env vars, flags, config files
		viper.AutomaticEnv()

		// Read connection settings from Viper (config/env/flags)
		baseURL := viper.GetString("base_url")
		username := viper.GetString("username")
		insecure := viper.GetBool("insecure")
		debugMode := viper.GetBool("debug")

		// Support both password (Basic Auth) and api_key (API Key Auth).
		// password takes priority for backward compatibility with TESTRAIL_PASSWORD.
		apiKey := viper.GetString("password")
		if apiKey == "" {
			apiKey = viper.GetString("api_key")
		}

		debug.DebugPrint("{rootCmd} - PersistentPreRunE running for command: %s", cmd.Use)
		debug.DebugPrint("{rootCmd} - baseURL=%s, username=%s", baseURL, username)
		debug.DebugPrint("{rootCmd} - insecure=%v", insecure)

		// Ensure config is set and does not contain default placeholders
		if config.IsDefaultValue(baseURL, config.DefaultBaseURL) ||
			config.IsDefaultValue(username, config.DefaultUsername) ||
			config.IsDefaultValue(apiKey, config.DefaultAPIKey) {
			return fmt.Errorf("configuration not set or contains default values\n" +
				"Run 'gotr config init' to create configuration,\n" +
				"then edit the file ~/.gotr/config/default.yaml")
		}

		debug.DebugPrint("{rootCmd} - Connecting to %s as %s", baseURL, username)

		// Create the HTTP client with options
		opts := []client.ClientOption{}
		if insecure {
			opts = append(opts, client.WithSkipTlsVerify(true)) // TLS verification is enabled by default
		}

		httpClient, err := client.NewClient(baseURL, username, apiKey, debugMode, opts...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		debug.DebugPrint("{rootCmd} - Client created and stored in context")

		// Store the client in context so it is available in all subcommands
		ctx := context.WithValue(cmd.Context(), httpClientKey, httpClient)

		// Inject Prompter into context (TerminalPrompter or NonInteractivePrompter)
		nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
		var p interactive.Prompter
		if nonInteractive {
			p = interactive.NewNonInteractivePrompter()
		} else {
			p = interactive.NewTerminalPrompter()
		}
		ctx = interactive.WithPrompter(ctx, p)

		cmd.SetContext(ctx)

		return nil
	},
	// Run is intentionally omitted — cobra shows help by default.
}

// Execute is called from main.go with a cancelable context (supports signal.NotifyContext).
func Execute(ctx context.Context) {
	rootCmd.SilenceUsage = true  // do not print usage on error
	rootCmd.SilenceErrors = true // we handle error output ourselves
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "\nInterrupted.")
			os.Exit(130)
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// GetClient retrieves the HTTP client from the command context.
func GetClient(cmd *cobra.Command) client.ClientInterface {
	return GetClientFromCtx(cmd.Context())
}

// GetClientFromCtx retrieves the HTTP client from a context.
// It satisfies client.GetClientFunc and decouples from cobra.
// Terminates the process with a clear message if the client is missing
// (programming error — PersistentPreRunE must set the client first).
func GetClientFromCtx(ctx context.Context) client.ClientInterface {
	val := ctx.Value(httpClientKey)
	if val == nil {
		fmt.Fprintln(os.Stderr, "FATAL: HTTP client not initialized. Check --username, --api-key and --url")
		processExit(1)
		return nil // unreachable
	}
	cli, ok := val.(client.ClientInterface)
	if !ok {
		fmt.Fprintln(os.Stderr, "FATAL: stored value does not implement ClientInterface")
		processExit(1)
		return nil // unreachable
	}
	return cli
}

func initConfig() {
	// 1. Add standard config search paths
	home, err := userHomeDir()
	if err != nil {
		// Non-fatal: continue without the home directory path
		ui.Warningf(os.Stderr, "cannot get user home directory: %v", err)
	} else {
		configDir := filepath.Join(home, ".gotr", "config")
		viper.AddConfigPath(configDir) // ~/.gotr/config
	}

	// Also search current directory (useful for local testing)
	viper.AddConfigPath(".")

	// 2. Config file name without extension (viper tries .yaml, .json, etc.)
	viper.SetConfigName("default") // look for default.yaml in ~/.gotr/config/
	viper.SetConfigType("yaml")

	// 3. Bind environment variables automatically
	viper.SetEnvPrefix("testrail") // e.g. TESTRAIL_BASE_URL, TESTRAIL_USERNAME
	viper.AutomaticEnv()

	// 4. Read config file (missing file is not an error)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			ui.Warningf(os.Stderr, "Config file error: %v", err)
		}
	}
}
