package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// workCmd is the interactive workspace entry point.
var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Interactive workspace — central entry point",
	Long: `Launches an interactive session where you can navigate between
all gotr commands without restarting the CLI.

Use this as the main entry point for interactive work with TestRail.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWorkHub(cmd)
	},
}

func init() {
	rootCmd.AddCommand(workCmd)
}

// workGroup is a top-level hub menu entry.
type workGroup struct {
	label string
	key   string
}

var workGroups = []workGroup{
	{"📊 Compare — compare resources between projects", "compare"},
	{"🔄 Sync — migrate data between projects", "sync"},
	{"📦 Snap — snapshots and rollback", "snap"},
	{"📋 Get — retrieve resource details", "get"},
	{"📝 Add — create resources", "add"},
	{"🗑  Delete — remove resources", "delete"},
	{"✏️  Update — modify resources", "update"},
	{"📤 Export — export data to files", "export"},
	{"⚙️  Config — view/edit configuration", "config"},
}

func runWorkHub(cmd *cobra.Command) error {
	ctx := cmd.Context()
	if !interactive.HasPrompterInContext(ctx) || interactive.IsNonInteractive(ctx) {
		return fmt.Errorf("gotr work requires interactive mode (remove --non-interactive)")
	}

	p := interactive.PrompterFromContext(ctx)

	// Server confirmation — first step before any work.
	baseURL := GetServerURLFromCtx(ctx)
	serverURL, err := interactive.SelectServer(ctx, p, baseURL)
	if err != nil {
		if interactive.IsGoBack(err) || interactive.IsExit(err) || interactive.IsInterrupt(err) {
			return nil
		}
		return fmt.Errorf("runWorkHub: %w", err)
	}

	printWorkHeader()

	// Create a work session and inject into context for parameter inheritance.
	session := &interactive.WorkSession{ServerURL: serverURL}
	ctx = interactive.WithSession(ctx, session)

	labels := make([]string, len(workGroups))
	for i, g := range workGroups {
		labels[i] = g.label
	}

	for {
		idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
			Prompt:    "What do you want to do?",
			Items:     labels,
			AllowBack: false,
		})
		if err != nil {
			if interactive.IsExit(err) || interactive.IsGoBack(err) {
				return nil
			}
			return fmt.Errorf("runWorkHub: %w", err)
		}

		selected := workGroups[idx]
		if err := dispatchWorkGroup(ctx, cmd, selected.key); err != nil {
			if interactive.IsGoBack(err) || interactive.IsExit(err) {
				continue
			}
			ui.Error(os.Stdout, fmt.Sprintf("%v", err))
		}
	}
}

func printWorkHeader() {
	baseURL := viper.GetString("base_url")
	if baseURL == "" {
		baseURL = "(not configured)"
	}
	fmt.Println()
	fmt.Println("╔══ gotr interactive workspace ══════════════════════╗")
	fmt.Printf("║  Server: %-42s║\n", baseURL)
	fmt.Println("╚═══════════════════════════════════════════════════╝")
	fmt.Println()
}

func dispatchWorkGroup(ctx context.Context, cmd *cobra.Command, key string) error {
	switch key {
	case "compare":
		return workCompareHub(ctx, cmd)
	case "sync":
		return workSyncHub(ctx, cmd)
	case "snap":
		return workSnapHub(ctx, cmd)
	default:
		return runSubcommandHub(ctx, cmd, key)
	}
}
