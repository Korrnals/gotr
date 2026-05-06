// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

// Package resolver provides a lazy, in-memory cache that maps TestRail
// entity IDs to their human-readable names. It is used by report
// builders to enrich plain ID columns with object names (project,
// suite, run, plan, milestone, case, section, user) without changing
// underlying call signatures or duplicating client wiring.
//
// The cache is per-instance (per-CLI-invocation): values are fetched
// on first access and reused for the rest of the session. Lookups are
// best-effort — when the API call fails (network error, missing
// permissions, deleted entity) the resolver returns an empty string
// and never propagates the error to the caller. Callers MUST treat an
// empty string as "name unavailable" and fall back to displaying the
// raw ID alone.
package resolver

import (
	"context"
	"sync"

	"github.com/Korrnals/gotr/internal/models/data"
)

// Client is the minimal subset of the TestRail HTTP client surface
// needed to resolve names. All methods accept a context and an int64
// ID and return a struct with a Name (or Title/Email) field.
//
// The interface is intentionally narrow so test doubles and partial
// implementations (e.g. cleanup's existing ProjectLister) can be
// adapted without pulling in the full *client.HTTPClient.
type Client interface {
	GetProject(ctx context.Context, id int64) (*data.GetProjectResponse, error)
	GetSuite(ctx context.Context, id int64) (*data.Suite, error)
	GetRun(ctx context.Context, id int64) (*data.Run, error)
	GetPlan(ctx context.Context, id int64) (*data.Plan, error)
	GetMilestone(ctx context.Context, id int64) (*data.Milestone, error)
	GetCase(ctx context.Context, id int64) (*data.Case, error)
	GetSection(ctx context.Context, id int64) (*data.Section, error)
	GetUser(ctx context.Context, id int64) (*data.User, error)
}

// Resolver is a thread-safe, lazy ID→name resolver. The zero value is
// not usable; construct via New.
type Resolver struct {
	cli Client

	mu       sync.Mutex
	projects map[int64]string
	suites   map[int64]string
	runs     map[int64]string
	plans    map[int64]string
	miles    map[int64]string
	cases    map[int64]string
	sects    map[int64]string
	users    map[int64]string
}

// New builds a Resolver backed by cli. If cli is nil, every lookup
// short-circuits to an empty string — handy for unit tests of
// renderers that should not perform any I/O.
func New(cli Client) *Resolver {
	return &Resolver{
		cli:      cli,
		projects: make(map[int64]string),
		suites:   make(map[int64]string),
		runs:     make(map[int64]string),
		plans:    make(map[int64]string),
		miles:    make(map[int64]string),
		cases:    make(map[int64]string),
		sects:    make(map[int64]string),
		users:    make(map[int64]string),
	}
}

// Project returns the name of the project with id, fetching it on
// first access. Returns "" when id<=0, the resolver has no client, or
// the underlying GetProject call fails.
func (r *Resolver) Project(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.projects[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if p, err := r.cli.GetProject(ctx, id); err == nil && p != nil {
			name = p.Name
		}
	}
	r.mu.Lock()
	r.projects[id] = name
	r.mu.Unlock()
	return name
}

// Suite returns the name of the suite with id, fetching it on first
// access.
func (r *Resolver) Suite(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.suites[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if s, err := r.cli.GetSuite(ctx, id); err == nil && s != nil {
			name = s.Name
		}
	}
	r.mu.Lock()
	r.suites[id] = name
	r.mu.Unlock()
	return name
}

// Run returns the name of the run with id.
func (r *Resolver) Run(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.runs[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if v, err := r.cli.GetRun(ctx, id); err == nil && v != nil {
			name = v.Name
		}
	}
	r.mu.Lock()
	r.runs[id] = name
	r.mu.Unlock()
	return name
}

// Plan returns the name of the plan with id.
func (r *Resolver) Plan(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.plans[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if v, err := r.cli.GetPlan(ctx, id); err == nil && v != nil {
			name = v.Name
		}
	}
	r.mu.Lock()
	r.plans[id] = name
	r.mu.Unlock()
	return name
}

// Milestone returns the name of the milestone with id.
func (r *Resolver) Milestone(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.miles[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if v, err := r.cli.GetMilestone(ctx, id); err == nil && v != nil {
			name = v.Name
		}
	}
	r.mu.Lock()
	r.miles[id] = name
	r.mu.Unlock()
	return name
}

// Case returns the title of the case with id (TestRail exposes the
// human label as Title, not Name).
func (r *Resolver) Case(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.cases[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if v, err := r.cli.GetCase(ctx, id); err == nil && v != nil {
			name = v.Title
		}
	}
	r.mu.Lock()
	r.cases[id] = name
	r.mu.Unlock()
	return name
}

// Section returns the name of the section with id.
func (r *Resolver) Section(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.sects[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if v, err := r.cli.GetSection(ctx, id); err == nil && v != nil {
			name = v.Name
		}
	}
	r.mu.Lock()
	r.sects[id] = name
	r.mu.Unlock()
	return name
}

// User returns the display name of the user with id, falling back to
// the e-mail when Name is unset.
func (r *Resolver) User(ctx context.Context, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	r.mu.Lock()
	if v, ok := r.users[id]; ok {
		r.mu.Unlock()
		return v
	}
	r.mu.Unlock()
	name := ""
	if r.cli != nil {
		if u, err := r.cli.GetUser(ctx, id); err == nil && u != nil {
			if u.Name != "" {
				name = u.Name
			} else {
				name = u.Email
			}
		}
	}
	r.mu.Lock()
	r.users[id] = name
	r.mu.Unlock()
	return name
}

// SetProject pre-populates the project cache with a known name. Used
// by callers that already fetched a Project elsewhere (e.g. cleanup's
// project resolver) and want to seed the cache to avoid a round-trip.
func (r *Resolver) SetProject(id int64, name string) {
	if r == nil || id <= 0 {
		return
	}
	r.mu.Lock()
	r.projects[id] = name
	r.mu.Unlock()
}

// ByKind dispatches to the right typed lookup. kind is one of the
// strings emitted by data.Attachment.InferredEntityType() or the
// matching plural/synonym ("case","cases","run","runs","plan","plans",
// "plan_entry","milestone","milestones","section","sections","suite",
// "suites","user","users","project","projects"). Unknown kinds yield
// an empty string. plan_entry is treated as plan.
func (r *Resolver) ByKind(ctx context.Context, kind string, id int64) string {
	if r == nil || id <= 0 {
		return ""
	}
	switch kind {
	case "case", "cases":
		return r.Case(ctx, id)
	case "run", "runs":
		return r.Run(ctx, id)
	case "plan", "plans", "plan_entry":
		return r.Plan(ctx, id)
	case "milestone", "milestones":
		return r.Milestone(ctx, id)
	case "section", "sections":
		return r.Section(ctx, id)
	case "suite", "suites":
		return r.Suite(ctx, id)
	case "user", "users":
		return r.User(ctx, id)
	case "project", "projects":
		return r.Project(ctx, id)
	}
	return ""
}
