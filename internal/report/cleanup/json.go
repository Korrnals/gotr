// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"encoding/json"
	"fmt"
)

// RenderJSON marshals the report with stable, indented formatting.
func RenderJSON(r *Report) ([]byte, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("cleanup-report: marshal json: %w", err)
	}
	// Trailing newline so files diff cleanly with POSIX tools.
	return append(b, '\n'), nil
}
