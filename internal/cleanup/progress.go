// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import "time"

// ScanPhase identifies which entity-walk stage is currently running for
// a project. For ProjectScanner the phase is "project" (a single bulk
// call). For EntityScanner each strategy stage emits its own phase.
type ScanPhase string

// Recognized phase identifiers. Implementations are free to introduce
// additional strings, but UI consumers should treat unknown values as
// opaque labels.
const (
	PhaseProject ScanPhase = "project"
	PhaseSuites  ScanPhase = "suites"
	PhaseCases   ScanPhase = "cases"
	PhaseRuns    ScanPhase = "runs"
	PhasePlans   ScanPhase = "plans"
	PhaseTests   ScanPhase = "tests"
)

// ScanProgress receives lifecycle events from the scanner pipeline.
//
// All methods MUST be safe for concurrent use; implementations are
// expected to fan-in to a single rendering goroutine. Methods MUST
// return promptly (no blocking I/O); slow consumers drop events
// instead of stalling the scan.
type ScanProgress interface {
	OnProjectStart(idx, total int, projectID int64, projectName string)
	OnPhase(projectID int64, phase ScanPhase, totalUnits int)
	OnUnit(projectID int64, phase ScanPhase, processed int)
	OnAttachmentsFound(projectID int64, found, eligible int, bytes int64)
	OnProjectDone(projectID int64, found, eligible int, bytes int64, elapsed time.Duration)
	OnError(projectID int64, err error)
}

// nopProgress is the zero-cost default when no sink is supplied.
type nopProgress struct{}

func (nopProgress) OnProjectStart(int, int, int64, string)             {}
func (nopProgress) OnPhase(int64, ScanPhase, int)                      {}
func (nopProgress) OnUnit(int64, ScanPhase, int)                       {}
func (nopProgress) OnAttachmentsFound(int64, int, int, int64)          {}
func (nopProgress) OnProjectDone(int64, int, int, int64, time.Duration) {}
func (nopProgress) OnError(int64, error)                               {}

// NoProgress is a singleton no-op sink. Use it instead of nil so
// callers can blindly invoke methods without nil-checks.
var NoProgress ScanProgress = nopProgress{}

// ScanProgressReceiver is an optional capability: scanners that
// implement it accept a progress sink for the current scan call.
// Callers should use type-assertions to discover the capability:
//
//	if r, ok := scanner.(cleanup.ScanProgressReceiver); ok {
//	    r.SetProgress(sink)
//	}
type ScanProgressReceiver interface {
	SetProgress(p ScanProgress)
}
