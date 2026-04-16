package interactive

import (
	"context"
	"testing"
)

func TestWorkSession_SetAndGetProjects(t *testing.T) {
	s := &WorkSession{}
	s.SetProjects(10, 20)

	src, dst := s.Projects()
	if src != 10 || dst != 20 {
		t.Errorf("Projects() = (%d, %d), want (10, 20)", src, dst)
	}
}

func TestWorkSession_SetAndGetSuites(t *testing.T) {
	s := &WorkSession{}
	s.SetSuites(5, 15)

	src, dst := s.Suites()
	if src != 5 || dst != 15 {
		t.Errorf("Suites() = (%d, %d), want (5, 15)", src, dst)
	}
}

func TestWithSession_RoundTrip(t *testing.T) {
	s := &WorkSession{ServerURL: "https://example.testrail.io"}
	s.SetProjects(1, 2)

	ctx := WithSession(context.Background(), s)
	got := SessionFromContext(ctx)
	if got == nil {
		t.Fatal("SessionFromContext returned nil")
	}
	if got.ServerURL != "https://example.testrail.io" {
		t.Errorf("ServerURL = %q, want %q", got.ServerURL, "https://example.testrail.io")
	}
	src, dst := got.Projects()
	if src != 1 || dst != 2 {
		t.Errorf("Projects() = (%d, %d), want (1, 2)", src, dst)
	}
}

func TestSessionFromContext_Missing(t *testing.T) {
	got := SessionFromContext(context.Background())
	if got != nil {
		t.Errorf("SessionFromContext on empty ctx = %v, want nil", got)
	}
}
