// Package client provides a centralized HTTP client abstraction for all
// TestRail API v2 interactions within gotr.
//
// It implements automatic Basic Auth injection, intelligent rate limiting,
// and structured interfaces for every API resource (cases, suites, projects,
// runs, results, attachments, etc.). The package supports both sequential
// and parallel API operations with built-in progress monitoring, making it
// suitable for bulk data operations.
//
// The central type is [HTTPClient], which wraps HTTP transport, base URL
// handling, and authentication. All API operations are defined through the
// composite [ClientInterface], which aggregates resource-specific interfaces
// (ProjectsAPI, CasesAPI, RunsAPI, etc.) to support seamless mocking in tests.
//
// Commands interact through the [Accessor], which retrieves the client from
// context and decouples client access from Cobra dependencies.
//
// Client creation uses a functional options pattern via [ClientOption]:
//
//	cli, err := client.NewClient(baseURL, user, key, debug,
//	    client.WithSkipTlsVerify(true),
//	)
package client
