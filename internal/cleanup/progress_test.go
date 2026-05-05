// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
)

// recordedEvent captures one ScanProgress callback for assertions.
type recordedEvent struct {
	kind     string
	phase    ScanPhase
	processed int
	count    int
	bytes    int64
	err      error
}

type recorder struct {
	mu     sync.Mutex
	events []recordedEvent
}

func (r *recorder) OnProjectStart(int, int, int64, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "start"})
}
func (r *recorder) OnPhase(_ int64, p ScanPhase, total int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "phase", phase: p, count: total})
}
func (r *recorder) OnUnit(_ int64, p ScanPhase, processed int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "unit", phase: p, processed: processed})
}
func (r *recorder) OnAttachmentsFound(_ int64, found, eligible int, bytes int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "found", count: found, processed: eligible, bytes: bytes})
}
func (r *recorder) OnProjectDone(_ int64, found, _ int, _ int64, _ time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "done", count: found})
}
func (r *recorder) OnError(_ int64, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, recordedEvent{kind: "err", err: err})
}

func (r *recorder) phasesSeen() map[ScanPhase]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[ScanPhase]int{}
	for _, e := range r.events {
		if e.kind == "phase" {
			out[e.phase] = e.count
		}
	}
	return out
}

func TestNoProgress_ZeroValueIsSafe(t *testing.T) {
	// Smoke: each method must be callable without panic on the
	// singleton sink.
	NoProgress.OnProjectStart(0, 0, 0, "")
	NoProgress.OnPhase(0, PhaseProject, 0)
	NoProgress.OnUnit(0, PhaseProject, 0)
	NoProgress.OnAttachmentsFound(0, 0, 0, 0)
	NoProgress.OnProjectDone(0, 0, 0, 0, 0)
	NoProgress.OnError(0, errors.New("x"))
}

func TestProjectScanner_EmitsPhaseAndUnit(t *testing.T) {
	api := &fakeScannerAPI{
		getAttachmentsForProject: func(int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{{ID: 1, Size: 100}, {ID: 2, Size: 50}}, nil
		},
	}
	sc := NewProjectScanner(api)
	r := &recorder{}
	sc.(ScanProgressReceiver).SetProgress(r)
	if _, err := sc.Scan(context.Background(), 42); err != nil {
		t.Fatalf("scan: %v", err)
	}
	phases := r.phasesSeen()
	if phases[PhaseProject] != 1 {
		t.Fatalf("phase project total = %d, want 1", phases[PhaseProject])
	}
}

func TestEntityScanner_EmitsPhasesPerStage(t *testing.T) {
	api := &fakeScannerAPI{
		getSuites: func(int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 100}}, nil
		},
		getCases: func(_, _ int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{{ID: 1}, {ID: 2}}, nil
		},
		getAttachmentsForCase: func(int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{}, nil
		},
		getRuns: func(int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 11}, {ID: 12}, {ID: 13}}, nil
		},
		getAttachmentsForRun: func(int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{}, nil
		},
		getPlans: func(int64) (data.GetPlansResponse, error) {
			return data.GetPlansResponse{{ID: 21}}, nil
		},
		getAttachmentsForPlan: func(int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{}, nil
		},
		getTests: func(int64) ([]data.Test, error) {
			return []data.Test{{ID: 31}}, nil
		},
		getAttachmentsForTest: func(int64) (data.GetAttachmentsResponse, error) {
			return data.GetAttachmentsResponse{}, nil
		},
	}
	sc := NewEntityScanner(api, EntityScannerOptions{
		WalkCases: true, WalkRuns: true, WalkPlans: true, WalkTests: true, Concurrency: 2,
	})
	r := &recorder{}
	sc.(ScanProgressReceiver).SetProgress(r)
	if _, err := sc.Scan(context.Background(), 7); err != nil {
		t.Fatalf("scan: %v", err)
	}
	phases := r.phasesSeen()
	if phases[PhaseSuites] != 1 {
		t.Fatalf("suites total = %d, want 1", phases[PhaseSuites])
	}
	if phases[PhaseCases] != 2 {
		t.Fatalf("cases total = %d, want 2", phases[PhaseCases])
	}
	if phases[PhaseRuns] != 3 {
		t.Fatalf("runs total = %d, want 3", phases[PhaseRuns])
	}
	if phases[PhasePlans] != 1 {
		t.Fatalf("plans total = %d, want 1", phases[PhasePlans])
	}
	if phases[PhaseTests] != 3 {
		t.Fatalf("tests total = %d, want 3 (one per run)", phases[PhaseTests])
	}
}
