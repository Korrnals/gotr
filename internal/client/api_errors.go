package client

import "strings"

// IsAPIMethodNotFound reports whether err originates from a TestRail HTTP
// 404 response carrying the canonical "Unknown method" body. It is used
// by callers that want to gracefully degrade when an endpoint is not
// available on older TestRail Server versions (for example,
// get_attachments_for_project was added in TestRail 7.5).
//
// The matcher is intentionally textual because the existing client
// wraps every non-2xx response into fmt.Errorf("API returned %s: %s",
// resp.Status, body) without surfacing a typed error value. The check
// is conservative: it requires both the 404 status marker and the
// "Unknown method" phrase to avoid masking unrelated 404s.
func IsAPIMethodNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "404") && strings.Contains(s, "Unknown method")
}
