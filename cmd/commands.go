package cmd

import (
	"github.com/Korrnals/gotr/cmd/attachments"
	"github.com/Korrnals/gotr/cmd/bdds"
	"github.com/Korrnals/gotr/cmd/cases"
	"github.com/Korrnals/gotr/cmd/compare"
	"github.com/Korrnals/gotr/cmd/configurations"
	"github.com/Korrnals/gotr/cmd/datasets"
	"github.com/Korrnals/gotr/cmd/get"
	"github.com/Korrnals/gotr/cmd/groups"
	"github.com/Korrnals/gotr/cmd/labels"
	"github.com/Korrnals/gotr/cmd/milestones"
	"github.com/Korrnals/gotr/cmd/plans"
	"github.com/Korrnals/gotr/cmd/reports"
	"github.com/Korrnals/gotr/cmd/result"
	"github.com/Korrnals/gotr/cmd/roles"
	"github.com/Korrnals/gotr/cmd/run"
	"github.com/Korrnals/gotr/cmd/sync"
	"github.com/Korrnals/gotr/cmd/templates"
	"github.com/Korrnals/gotr/cmd/test"
	"github.com/Korrnals/gotr/cmd/tests"
	"github.com/Korrnals/gotr/cmd/users"
	"github.com/Korrnals/gotr/cmd/variables"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// init registers all commands at startup.
func init() {
	// Initialize config first
	initConfig()

	// Global flags for the root command
	initGlobalFlags()

	// Register top-level commands
	registerConfigCmd()
	registerListCmd()
	registerAddCmd()
	registerDeleteCmd()
	registerUpdateCmd()
	registerExportCmd()
	registerCompletionCmd()

	// Register subpackage commands (pass GetClient* accessor)
	attachments.Register(rootCmd, GetClientInterface)
	bdds.Register(rootCmd, GetClientInterface)
	cases.Register(rootCmd, GetClientInterface)
	compare.Register(rootCmd, GetClientInterface)
	configurations.Register(rootCmd, GetClientInterface)
	datasets.Register(rootCmd, GetClientInterface)
	get.Register(rootCmd, GetClient)
	groups.Register(rootCmd, GetClientInterface)
	labels.Register(rootCmd, GetClientInterface)
	milestones.Register(rootCmd, GetClientInterface)
	plans.Register(rootCmd, GetClientInterface)
	reports.Register(rootCmd, GetClientInterface)
	run.Register(rootCmd, GetClient)
	result.Register(rootCmd, GetClient)
	roles.Register(rootCmd, GetClientInterface)
	sync.Register(rootCmd, GetClient)
	test.Register(rootCmd, GetClient)
	templates.Register(rootCmd, GetClientInterface)
	tests.Register(rootCmd, GetClientInterface)
	users.Register(rootCmd, GetClientInterface)
	variables.Register(rootCmd, GetClientInterface)
}

// must panics if err is non-nil. Used for init-time bindings that
// indicate a programming error (e.g. binding a non-existent flag).
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// initGlobalFlags registers persistent flags shared by all subcommands.
func initGlobalFlags() {
	// Global flags — connection and basic settings only
	rootCmd.PersistentFlags().String("url", "", "TestRail base URL")
	rootCmd.PersistentFlags().StringP("username", "u", "", "TestRail user email")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "TestRail API key")
	rootCmd.PersistentFlags().Bool("insecure", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().BoolP("config", "c", false, "Create default config file")

	// Hidden debug flag
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")
	must(rootCmd.PersistentFlags().MarkHidden("debug"))

	// Quiet mode (suppress informational output for CI/CD)
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress progress, stats, and save messages")

	// Non-interactive mode (CI/CD, scripting)
	rootCmd.PersistentFlags().Bool("non-interactive", false, "Disable interactive prompts; fail if input required")

	// Global output format
	rootCmd.PersistentFlags().StringP("format", "f", "table", "Output format: table, json, csv, md, html")

	// Bind flags to Viper
	must(viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("url")))
	must(viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username")))
	must(viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key")))
	must(viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure")))
	must(viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")))
}

// ============================================
// Config
// ============================================

func registerConfigCmd() {
	rootCmd.AddCommand(configCmd)

	// Config subcommands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configEditCmd)

	// Shell completion for "gotr config "
	configCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"init", "path", "view", "edit"}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// ============================================
// List
// ============================================

func registerListCmd() {
	rootCmd.AddCommand(listCmd)

	// Flags specific to list
	listCmd.Flags().Bool("json", false, "Output as JSON")
	listCmd.Flags().Bool("short", false, "Short output (URI only)")

	// Shell completion for the first argument (resource name)
	listCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return ValidResources, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// ============================================
// Add
// ============================================

func registerAddCmd() {
	rootCmd.AddCommand(addCmd)
}

// ============================================
// Delete
// ============================================

func registerDeleteCmd() {
	rootCmd.AddCommand(deleteCmd)
}

// ============================================
// Update
// ============================================

func registerUpdateCmd() {
	rootCmd.AddCommand(updateCmd)
}

// ============================================
// Export
// ============================================

func registerExportCmd() {
	rootCmd.AddCommand(exportCmd)

	// Export-specific flags
	exportCmd.Flags().StringP("project-id", "p", "", "Project ID (for endpoints with {project_id})")
	exportCmd.Flags().StringP("suite-id", "s", "", "Suite ID (for get_cases)")
	exportCmd.Flags().String("section-id", "", "Section ID (for get_cases)")
	exportCmd.Flags().String("milestone-id", "", "Milestone ID (for get_runs)")

	// Save response to ~/.gotr/exports/
	exportCmd.Flags().Bool("save", false, "Save response to ~/.gotr/exports/export/")

	// Shell completion
	exportCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return ValidResources, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			endpoints, _ := getResourceEndpoints(args[0], "list")
			return endpoints, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// ============================================
// Completion
// ============================================

func registerCompletionCmd() {
	rootCmd.AddCommand(completionCmd)
}
