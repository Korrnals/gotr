package interactive

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NavPrefix is the key prefix for cross-module navigation actions.
const NavPrefix = "nav:"

// CrossNavOptions returns the standard cross-navigation ActionOptions
// that can be appended to any post-action menu.
// Each key is prefixed with NavPrefix so HandleCrossNav can identify them.
func CrossNavOptions() []ActionOption {
	return []ActionOption{
		{Label: "📊 Compare: verify current state", Key: NavPrefix + "compare"},
		{Label: "🔄 Sync: migrate data", Key: NavPrefix + "sync"},
		{Label: "📦 Snap: manage snapshots", Key: NavPrefix + "snap"},
	}
}

// IsCrossNav returns true if the action key is a cross-navigation transition.
func IsCrossNav(key string) bool {
	return strings.HasPrefix(key, NavPrefix)
}

// HandleCrossNav executes a cross-module navigation transition.
// Returns true if the key was a nav: key and was handled, false otherwise.
//
// Server guard (B4): if a WorkSession exists, verifies that viper base_url
// still matches the session's ServerURL. Warns on mismatch but proceeds.
func HandleCrossNav(ctx context.Context, cmd *cobra.Command, key string) bool {
	if !IsCrossNav(key) {
		return false
	}

	// B4: server boundary check.
	if s := SessionFromContext(ctx); s != nil && s.ServerURL != "" {
		if current := viper.GetString("base_url"); current != "" && current != s.ServerURL {
			fmt.Fprintf(os.Stderr,
				"⚠ Server mismatch: session=%s, config=%s — using session server\n",
				s.ServerURL, current)
		}
	}

	target := strings.TrimPrefix(key, NavPrefix)
	root := cmd.Root()
	switch target {
	case "compare":
		_ = RunSubcommand(ctx, root, "compare", "all")
	case "sync":
		_ = RunSubcommand(ctx, root, "sync", "full")
	case "snap":
		_ = RunSubcommand(ctx, root, "snap", "list")
	}
	return true
}
