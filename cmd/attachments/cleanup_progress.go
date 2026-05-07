// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Korrnals/gotr/internal/cleanup"
	"github.com/Korrnals/gotr/internal/ui"
)

// scanProgressAdapter fans cleanup.ScanProgress events into a
// ui.MultilineStatus snapshot (when Live is non-nil) and/or a logger
// (when Log is non-nil). Running totals (found / eligible / bytes) are
// accumulated across all projects.
//
// All methods are safe for concurrent use.
type scanProgressAdapter struct {
	mu         sync.Mutex
	startedAt  time.Time
	live       *ui.MultilineStatus
	log        io.Writer
	title      string
	strategy   string
	totalProj  int
	doneProj   int
	cur        ui.Snapshot
	totFound   int
	totElig    int
	totBytes   int64
	phaseTotal map[int64]map[cleanup.ScanPhase]int
	phaseDone  map[int64]map[cleanup.ScanPhase]int
}

func newScanProgressAdapter(title, strategy string, live *ui.MultilineStatus, log io.Writer) *scanProgressAdapter {
	return &scanProgressAdapter{
		startedAt:  time.Now(),
		live:       live,
		log:        log,
		title:      title,
		strategy:   strategy,
		phaseTotal: map[int64]map[cleanup.ScanPhase]int{},
		phaseDone:  map[int64]map[cleanup.ScanPhase]int{},
	}
}

func (a *scanProgressAdapter) snapshot() ui.Snapshot {
	s := a.cur
	s.Title = a.title
	s.Strategy = a.strategy
	s.Found = a.totFound
	s.Eligible = a.totElig
	s.Bytes = a.totBytes
	s.ProjectN = a.totalProj
	s.Elapsed = time.Since(a.startedAt)
	if a.doneProj > 0 && a.totalProj > 0 && a.doneProj < a.totalProj {
		perProj := s.Elapsed / time.Duration(a.doneProj)
		s.ETA = perProj * time.Duration(a.totalProj-a.doneProj)
	}
	return s
}

func (a *scanProgressAdapter) push() {
	if a.live == nil {
		return
	}
	a.live.Update(a.snapshot())
}

func (a *scanProgressAdapter) OnProjectStart(idx, total int, projectID int64, projectName string) {
	a.mu.Lock()
	a.totalProj = total
	a.cur.ProjectIdx = idx
	a.cur.ProjectName = projectName
	a.cur.Phase = ""
	a.cur.PhaseDone = 0
	a.cur.PhaseTotal = 0
	if _, ok := a.phaseTotal[projectID]; !ok {
		a.phaseTotal[projectID] = map[cleanup.ScanPhase]int{}
	}
	if _, ok := a.phaseDone[projectID]; !ok {
		a.phaseDone[projectID] = map[cleanup.ScanPhase]int{}
	}
	a.push()
	a.mu.Unlock()
}

func (a *scanProgressAdapter) OnPhase(projectID int64, phase cleanup.ScanPhase, totalUnits int) {
	a.mu.Lock()
	if _, ok := a.phaseTotal[projectID]; !ok {
		a.phaseTotal[projectID] = map[cleanup.ScanPhase]int{}
	}
	a.phaseTotal[projectID][phase] = totalUnits
	a.cur.Phase = string(phase)
	a.cur.PhaseTotal = totalUnits
	a.cur.PhaseDone = a.phaseDone[projectID][phase]
	a.push()
	a.mu.Unlock()
}

func (a *scanProgressAdapter) OnUnit(projectID int64, phase cleanup.ScanPhase, processed int) {
	a.mu.Lock()
	if _, ok := a.phaseDone[projectID]; !ok {
		a.phaseDone[projectID] = map[cleanup.ScanPhase]int{}
	}
	a.phaseDone[projectID][phase] = processed
	if a.cur.Phase == string(phase) {
		a.cur.PhaseDone = processed
	}
	a.push()
	a.mu.Unlock()
}

func (a *scanProgressAdapter) OnAttachmentsFound(_ int64, _, _ int, _ int64) {
	// Per-project totals are tallied in OnProjectDone for symmetry
	// with the chunked driver. OnAttachmentsFound is informational
	// and dropped by the cleanup adapter.
}

func (a *scanProgressAdapter) OnProjectDone(projectID int64, found, eligible int, bytes int64, elapsed time.Duration) {
	a.mu.Lock()
	a.doneProj++
	a.totFound += found
	a.totElig += eligible
	a.totBytes += bytes
	a.push()
	a.mu.Unlock()
	if a.log != nil {
		fmt.Fprintf(a.log, "INFO: project %d/%d done: id=%d found=%d eligible=%d bytes=%s elapsed=%s\n",
			a.doneProj, a.totalProj, projectID, found, eligible, ui.HumanBytes(bytes), elapsed.Round(time.Second))
	}
}

func (a *scanProgressAdapter) OnError(projectID int64, err error) {
	a.mu.Lock()
	a.doneProj++
	a.push()
	a.mu.Unlock()
	if a.log != nil {
		fmt.Fprintf(a.log, "WARN: project %d failed: %v\n", projectID, err)
	}
}
