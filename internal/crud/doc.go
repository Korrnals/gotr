// Package crud provides generic executor functions for CRUD command patterns.
//
// Execute[Req, Resp] handles the full lifecycle: JSON/flags parsing → API call → output.
// DryRun[Req] handles the preview path: JSON/flags parsing → dry-run print.
//
// Both share a single buildReq function per resource, eliminating boilerplate
// duplication between execute and dry-run code paths.
package crud
