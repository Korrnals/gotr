// Package reporter centralizes stats reporting for gotr CLI output so
// format changes happen in one place, not scattered across commands.
//
// It uses go-pretty for rendering with ANSI-colored width-1 indicators
// instead of emoji (which have unpredictable widths across VS Code,
// iTerm, and GNOME Terminal). The builder pattern ([New], [Section],
// [Stat], [Print]) supports dynamic section and stat composition.
//
// Stat values are wrapped with colors (green/yellow/red/cyan) for
// visual emphasis. Commands pass emoji hints (✅, 🔗, ⚠️); reporter
// maps them to safe colored characters automatically.
package reporter
