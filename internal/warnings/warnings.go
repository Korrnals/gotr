// Package warnings provides a centralized, suppressible warning sink for the
// gotr CLI.
//
// Warnings are classified by a stable Key (e.g. "tls_insecure", "flat_layout",
// "deprecation"). The user can silence specific warnings via the YAML config
// key `ui.suppress_warnings: [key1, key2]`. A per-invocation override,
// `--show-warnings`, force-enables all warnings regardless of configuration.
//
// Each warning key may also opt into a one-time "hint" shown the first time it
// is emitted in a given process; the hint explains how to silence the warning
// going forward.
package warnings

import (
	"fmt"
	"io"
	"sync"
)

// Key is a stable identifier for a class of warnings. Add new constants as
// needed; keys should be lowercase_underscore and documented in user-facing
// configuration docs.
type Key string

// Stable keys used across the codebase.
const (
	// KeyTLSInsecure covers the "TLS verification disabled" banner that is
	// emitted when the user enables --insecure / tls.insecure.
	KeyTLSInsecure Key = "tls_insecure"
	// KeyDeprecation covers any CLI/config surface that has been deprecated
	// and will be removed in a future release.
	KeyDeprecation Key = "deprecation"
	// KeyFlatLayout covers the one-time hint about the legacy flat
	// ~/.gotr/reports/ layout. Persistence of "already shown" belongs in
	// state.json (FlatLayoutWarned) rather than in-memory.
	KeyFlatLayout Key = "flat_layout"
)

var (
	mu         sync.Mutex
	suppressed = map[Key]bool{}
	showAll    bool
	shownHint  = map[Key]bool{}
)

// Init configures the registry. It is safe to call multiple times; later calls
// override earlier ones. An empty suppressList means "nothing suppressed".
func Init(suppressList []string, showAllFlag bool) {
	mu.Lock()
	defer mu.Unlock()
	suppressed = make(map[Key]bool, len(suppressList))
	for _, k := range suppressList {
		if k == "" {
			continue
		}
		suppressed[Key(k)] = true
	}
	showAll = showAllFlag
	shownHint = map[Key]bool{}
}

// Suppressed reports whether a given warning key is silenced by current
// configuration. Returns false when --show-warnings is set.
func Suppressed(k Key) bool {
	mu.Lock()
	defer mu.Unlock()
	if showAll {
		return false
	}
	return suppressed[k]
}

// Emitf writes a formatted warning to w unless the key is suppressed. On the
// first emission of a given key in this process, a one-time hint is appended
// explaining how to silence the warning via `ui.suppress_warnings`.
func Emitf(w io.Writer, k Key, format string, args ...any) {
	if Suppressed(k) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(w, "⚠️  %s\n", msg)

	mu.Lock()
	firstTime := !shownHint[k]
	shownHint[k] = true
	mu.Unlock()

	if firstTime {
		fmt.Fprintf(w, "   (tip: add '%s' to ui.suppress_warnings in ~/.gotr/config/default.yaml to silence this warning)\n", k)
	}
}

// Emit writes a plain warning message. See Emitf for formatting details.
func Emit(w io.Writer, k Key, msg string) {
	Emitf(w, k, "%s", msg)
}

// ResetForTest clears registry state. Call from tests only.
func ResetForTest() {
	mu.Lock()
	defer mu.Unlock()
	suppressed = map[Key]bool{}
	showAll = false
	shownHint = map[Key]bool{}
}
