// Package progress provides async progress tracking functionality.
package progress

import (
	"context"
	"fmt"
	"time"
)

// AsyncProgress represents async progress tracking
type AsyncProgress struct {
	bar     *Bar
	updates chan int
	done    chan bool
	cancel  context.CancelFunc
}

// StartAsync starts async progress tracking
// Updates channel receives increment values
// Returns channel for sending updates
func StartAsync(total int64, description string, pm *Manager) chan int {
	if pm == nil {
		return make(chan int, 100) // No-op channel
	}

	updates := make(chan int, 100)
	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())

	bar := pm.NewBar(total, description)

	go func() {
		defer close(done)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		pending := 0
		for {
			select {
			case <-ctx.Done():
				// Flush pending updates
				if pending > 0 && bar != nil {
					bar.Add(pending)
				}
				return
			case n, ok := <-updates:
				if !ok {
					// Channel closed, flush and exit
					if pending > 0 && bar != nil {
						bar.Add(pending)
					}
					return
				}
				pending += n
			case <-ticker.C:
				if pending > 0 && bar != nil {
					bar.Add(pending)
					pending = 0
				}
			}
		}
	}()

	// Store for cleanup
	_ = &AsyncProgress{
		bar:     bar,
		updates: updates,
		done:    done,
		cancel:  cancel,
	}

	return updates
}

// NewMultiPhase creates a multi-phase progress tracker
// Returns a ProgressTracker that can switch between phases
func NewMultiPhase(pm *Manager) *ProgressTracker {
	if pm == nil {
		return &ProgressTracker{enabled: false}
	}
	return &ProgressTracker{
		pm:      pm,
		updates: make(chan ProgressUpdate, 100),
		enabled: true,
	}
}

// ProgressUpdate represents a progress update
type ProgressUpdate struct {
	Phase       string
	Increment   int
	Total       int64
	Description string
}

// ProgressTracker tracks multi-phase progress
type ProgressTracker struct {
	pm      *Manager
	bar     *Bar
	updates chan ProgressUpdate
	enabled bool
	cancel  context.CancelFunc
}

// Start begins tracking progress
func (pt *ProgressTracker) Start() {
	if !pt.enabled {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	pt.cancel = cancel

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		var pending int
		var currentPhase string

		for {
			select {
			case <-ctx.Done():
				if pending > 0 && pt.bar != nil {
					pt.bar.Add(pending)
				}
				return
			case update, ok := <-pt.updates:
				if !ok {
					if pending > 0 && pt.bar != nil {
						pt.bar.Add(pending)
					}
					return
				}

				// New phase - create new bar
				if update.Phase != currentPhase && update.Total > 0 {
					if pt.bar != nil {
						pt.bar.Finish()
					}
					pt.bar = pt.pm.NewBar(update.Total, update.Description)
					currentPhase = update.Phase
					pending = 0
				}

				pending += update.Increment

			case <-ticker.C:
				if pending > 0 && pt.bar != nil {
					pt.bar.Add(pending)
					pending = 0
				}
			}
		}
	}()
}

// Update sends a progress update
func (pt *ProgressTracker) Update(phase string, increment int, total int64, description string) {
	if !pt.enabled {
		return
	}
	select {
	case pt.updates <- ProgressUpdate{
		Phase:       phase,
		Increment:   increment,
		Total:       total,
		Description: description,
	}:
	default:
		// Channel full, skip update
	}
}

// Complete finishes progress tracking
func (pt *ProgressTracker) Complete() {
	if !pt.enabled {
		return
	}
	if pt.cancel != nil {
		pt.cancel()
	}
	close(pt.updates)
	if pt.bar != nil {
		pt.bar.Finish()
	}
}

// PhaseSpinner shows a spinner for a phase
func PhaseSpinner(pm *Manager, description string) func() {
	if pm == nil {
		return func() {}
	}

	spinner := pm.NewSpinner(description)
	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// Spinner updates automatically via mpb
			}
		}
	}()

	return func() {
		ticker.Stop()
		close(done)
		if spinner != nil {
			spinner.Finish()
		}
	}
}

// PrintOnNewLine ensures output starts on new line after progress
func PrintOnNewLine(format string, args ...interface{}) {
	fmt.Println()
	fmt.Printf(format, args...)
}
