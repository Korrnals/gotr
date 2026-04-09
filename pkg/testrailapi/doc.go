// Package testrailapi provides a data-driven representation of the entire
// TestRail API v2, enabling CLI help text and endpoint discovery without
// hardcoded strings.
//
// Each resource (Cases, Projects, Runs, Suites, etc.) exposes its endpoints
// as an array of [APIPath] structs containing HTTP method, URI, description,
// and parameter metadata. [New] returns a fully populated [TestRailAPI] with
// all 25+ resource groups. Commands call resource.Paths() to list available
// endpoints or API.Paths() to dump the full endpoint catalog.
//
// This design keeps API metadata centralized and maintainable, decoupled
// from the HTTP client implementation in internal/client.
package testrailapi
