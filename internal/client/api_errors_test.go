package client

import (
	"errors"
	"testing"
)

func TestIsAPIMethodNotFound(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unrelated", errors.New("connection refused"), false},
		{"404 not found", errors.New("API returned 404 Not Found: Field :id is missing"), false},
		{"server 500", errors.New("API returned 500 Internal Server Error: oops"), false},
		{"unknown method", errors.New("API returned 404 File Not Found: Unknown method 'get_attachments_for_project'"), true},
		{"unknown method other case", errors.New("fetchAllPages get_x (offset=0): API returned 404 File Not Found: Unknown method 'foo'"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsAPIMethodNotFound(tc.err); got != tc.want {
				t.Fatalf("IsAPIMethodNotFound(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
