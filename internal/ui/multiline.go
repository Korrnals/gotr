// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// Snapshot is the UI input value type built from cleanup.ScanProgress
// events. The cleanup command updates a single Snapshot under a mutex
// and the renderer last-write-wins reads it on each tick.
type Snapshot struct {
	Title       string
	Strategy    string
	ProjectIdx  int
	ProjectN    int
	ProjectName string
	Phase       string
	PhaseDone   int
	PhaseTotal  int
	Found       int
	Eligible    int
	Bytes       int64
	Elapsed     time.Duration
	ETA         time.Duration
}

// MultilineConfig configures MultilineStatus.
type MultilineConfig struct {
	Writer  io.Writer
	Quiet   bool
	Refresh time.Duration
	Lines   int
	NowFunc func() time.Time
	// ForceTTY bypasses the TTY detection (tests). When false the
	// renderer no-ops on non-tty writers.
	ForceTTY bool
}

// MultilineStatus is a 5-line live progress block rendered with ANSI
// cursor-up escapes. On non-TTY writers (or Quiet=true) it emits no
// output; the cleanup command falls back to per-project INFO logs.
type MultilineStatus struct {
	cfg    MultilineConfig
	mu     sync.Mutex
	snap   Snapshot
	dirty  bool
	rendered bool
}

// NewMultilineStatus constructs a renderer. Defaults: Refresh=100ms,
// Lines=5, Writer=os.Stderr, NowFunc=time.Now.
func NewMultilineStatus(cfg MultilineConfig) *MultilineStatus {
	if cfg.Refresh <= 0 {
		cfg.Refresh = 100 * time.Millisecond
	}
	if cfg.Lines <= 0 {
		cfg.Lines = 5
	}
	if cfg.Writer == nil {
		cfg.Writer = os.Stderr
	}
	if cfg.NowFunc == nil {
		cfg.NowFunc = time.Now
	}
	return &MultilineStatus{cfg: cfg}
}

// Update stores the latest snapshot. Non-blocking; last write wins.
func (m *MultilineStatus) Update(s Snapshot) {
	m.mu.Lock()
	m.snap = s
	m.dirty = true
	m.mu.Unlock()
}

// Active reports whether the renderer is going to emit live output.
// Useful for callers that want to switch to a logged-progress fallback.
func (m *MultilineStatus) Active() bool {
	if m.cfg.Quiet {
		return false
	}
	if m.cfg.ForceTTY {
		return true
	}
	return isWriterTTY(m.cfg.Writer)
}

// Start kicks off the renderer goroutine and returns a stop func that
// drains pending events and clears the rendered block. Calling stop
// twice is a no-op. ctx cancellation triggers graceful shutdown.
func (m *MultilineStatus) Start(ctx context.Context) func() {
	if !m.Active() {
		return func() {}
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(m.cfg.Refresh)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.render()
			}
		}
	}()
	stopped := false
	return func() {
		if stopped {
			return
		}
		stopped = true
		close(stop)
		<-done
		// Final clear so subsequent stdout output is not visually
		// interleaved with the spinner block.
		m.clear()
	}
}

// render draws the latest snapshot to the writer. Concurrent-safe:
// only one goroutine (the ticker) calls it.
func (m *MultilineStatus) render() {
	m.mu.Lock()
	snap := m.snap
	dirty := m.dirty
	rendered := m.rendered
	m.mu.Unlock()
	if !dirty && rendered {
		return
	}
	out := buildBlock(snap, m.cfg.Lines)
	w := m.cfg.Writer
	if rendered {
		// Move the cursor up Lines rows and clear each one before
		// re-drawing.
		fmt.Fprintf(w, "\033[%dA", m.cfg.Lines)
	}
	for _, line := range out {
		fmt.Fprintf(w, "\r\033[K%s\n", line)
	}
	m.mu.Lock()
	m.rendered = true
	m.dirty = false
	m.mu.Unlock()
}

// clear erases the rendered block (used on stop so subsequent output is
// not visually mixed with the spinner). It does not write a "final
// snapshot" — the cleanup command prints its own summary block.
func (m *MultilineStatus) clear() {
	m.mu.Lock()
	rendered := m.rendered
	m.mu.Unlock()
	if !rendered {
		return
	}
	w := m.cfg.Writer
	fmt.Fprintf(w, "\033[%dA", m.cfg.Lines)
	for i := 0; i < m.cfg.Lines; i++ {
		fmt.Fprintf(w, "\r\033[K\n")
	}
	fmt.Fprintf(w, "\033[%dA", m.cfg.Lines)
}

// RenderForTest builds the 5-line block as a single string (with
// trailing newlines) for use in tests. It does not write to the
// configured writer.
func (m *MultilineStatus) RenderForTest(s Snapshot) string {
	lines := buildBlock(s, m.cfg.Lines)
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return b.String()
}

// buildBlock returns the rendered lines for the current snapshot. Pure
// function; testable without ANSI sequences.
func buildBlock(s Snapshot, lines int) []string {
	out := make([]string, 0, lines)
	header := s.Title
	if s.Strategy != "" {
		header = fmt.Sprintf("%s — %s", s.Title, s.Strategy)
	}
	out = append(out, header)

	projLine := fmt.Sprintf("   project %d/%d", s.ProjectIdx, s.ProjectN)
	if s.ProjectName != "" {
		projLine = fmt.Sprintf("%s  →  %s", projLine, truncateRunes(s.ProjectName, 60))
	}
	out = append(out, projLine)

	bar := renderBar(s.PhaseDone, s.PhaseTotal, 10)
	phase := s.Phase
	if phase == "" {
		phase = "—"
	}
	out = append(out,
		fmt.Sprintf("   phase: %-10s %s %d / %d", phase, bar, s.PhaseDone, s.PhaseTotal),
		fmt.Sprintf("   found: %d  eligible: %d  size: %s", s.Found, s.Eligible, HumanBytes(s.Bytes)),
	)

	etaStr := "—"
	if s.ETA > 0 {
		etaStr = "~" + fmtDurationShort(s.ETA)
	}
	out = append(out, fmt.Sprintf("   ⏱ %s  ETA %s", fmtDurationShort(s.Elapsed), etaStr))

	for len(out) < lines {
		out = append(out, "")
	}
	return out[:lines]
}

func renderBar(done, total, width int) string {
	if width <= 0 {
		return ""
	}
	if total <= 0 {
		return strings.Repeat("░", width)
	}
	if done > total {
		done = total
	}
	filled := int(float64(width) * float64(done) / float64(total))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func fmtDurationShort(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	d = d.Round(time.Second)
	h := int(d / time.Hour)
	d -= time.Duration(h) * time.Hour
	m := int(d / time.Minute)
	d -= time.Duration(m) * time.Minute
	sec := int(d / time.Second)
	switch {
	case h > 0:
		return fmt.Sprintf("%dh%02dm%02ds", h, m, sec)
	case m > 0:
		return fmt.Sprintf("%dm%02ds", m, sec)
	default:
		return fmt.Sprintf("%ds", sec)
	}
}

// HumanBytes is a UI-shared helper for formatting byte counts used by
// progress snapshots and final summary blocks.
func HumanBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)
	switch {
	case n >= tb:
		return fmt.Sprintf("%.2f TiB (%d B)", float64(n)/float64(tb), n)
	case n >= gb:
		return fmt.Sprintf("%.2f GiB (%d B)", float64(n)/float64(gb), n)
	case n >= mb:
		return fmt.Sprintf("%.2f MiB (%d B)", float64(n)/float64(mb), n)
	case n >= kb:
		return fmt.Sprintf("%.2f KiB (%d B)", float64(n)/float64(kb), n)
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// truncateRunes shortens s to maxRunes runes, appending an ellipsis
// when truncation occurred.
func truncateRunes(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	if maxRunes <= 1 {
		return string(r[:maxRunes])
	}
	return string(r[:maxRunes-1]) + "…"
}

// isWriterTTY tells whether w is a *os.File pointing at a terminal.
// Anything else (bytes.Buffer, pipe-backed file) is treated as
// non-interactive and disables the live block.
func isWriterTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
