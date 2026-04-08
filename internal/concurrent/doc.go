// Package concurrent provides building blocks for efficient parallel API
// operations while respecting rate limits and handling transient failures.
//
// [WorkerPool] manages a bounded set of worker goroutines, submits tasks
// via Go closures, and aggregates errors through errgroup.Group with
// context-aware cancellation. [ParallelMap] and [ParallelForEach] provide
// higher-level helpers for applying a function to a slice concurrently.
//
// Together these primitives enable commands to safely fetch related
// resources (suites, cases, sections) in parallel while respecting
// TestRail API quotas.
package concurrent
