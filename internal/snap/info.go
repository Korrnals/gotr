package snap

import (
	"os"

	"github.com/Korrnals/gotr/internal/ui"
)

// InfoBanner prints an informational message about snapshot state before mutation.
// Suppressed by --quiet (handled inside ui.Info/ui.Infof).
func InfoBanner(enabled bool) {
	if enabled {
		ui.Info(os.Stderr, "Snapshot: enabled. Data will be saved to ~/.gotr/snaps/ before mutation.")
		ui.Info(os.Stderr, "  Override: --snapshot=false | config: snap.enabled: false")
	} else {
		ui.Info(os.Stderr, "Snapshot: disabled for this operation.")
	}
}
