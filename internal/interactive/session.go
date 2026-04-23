package interactive

import (
	"context"
	"sync"
)

// WorkSession holds mutable state that is shared between commands
// during an interactive gotr work session. It allows parameter inheritance:
// compare → sync → snap without re-prompting for project IDs.
//
// WorkSession is concurrency-safe via an internal mutex.
type WorkSession struct {
	mu sync.Mutex

	ServerURL       string
	SrcProjectID    int64
	DstProjectID    int64
	SrcSuiteID      int64
	DstSuiteID      int64
}

// SetProjects stores the source and destination project IDs.
func (s *WorkSession) SetProjects(src, dst int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SrcProjectID = src
	s.DstProjectID = dst
}

// SetSuites stores the source and destination suite IDs.
func (s *WorkSession) SetSuites(src, dst int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SrcSuiteID = src
	s.DstSuiteID = dst
}

// Projects returns the stored project IDs (src, dst).
func (s *WorkSession) Projects() (src, dst int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.SrcProjectID, s.DstProjectID
}

// Suites returns the stored suite IDs (src, dst).
func (s *WorkSession) Suites() (src, dst int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.SrcSuiteID, s.DstSuiteID
}

type sessionContextKey struct{}

// WithSession returns a new context that carries the given WorkSession.
func WithSession(ctx context.Context, s *WorkSession) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// SessionFromContext returns the WorkSession from context, or nil if absent.
func SessionFromContext(ctx context.Context) *WorkSession {
	s, _ := ctx.Value(sessionContextKey{}).(*WorkSession)
	return s
}
