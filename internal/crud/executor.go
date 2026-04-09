// Package crud provides generic executor functions for CRUD command patterns.
// It eliminates boilerplate JSON/flags parsing, API call wrapping, and output
// handling shared across add/update commands.
package crud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// Execute handles the common JSON-or-flags → API call → output pattern
// used by add/update commands.
//
// If jsonData is non-empty, it is unmarshaled into Req.
// Otherwise, buildReq is called with validate=true to construct the request from flags.
// The API call is then made and the result is output via output.OutputResult.
func Execute[Req any, Resp any](
	cmd *cobra.Command,
	id int64,
	jsonData []byte,
	buildReq func(*cobra.Command, bool) (*Req, error),
	apiCall func(context.Context, int64, *Req) (Resp, error),
	failMsg string,
) error {
	ctx := cmd.Context()
	var req Req

	if len(jsonData) > 0 {
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		built, err := buildReq(cmd, true)
		if err != nil {
			return err
		}
		req = *built
	}

	result, err := apiCall(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("%s: %w", failMsg, err)
	}

	return output.OutputResult(cmd, result, "result")
}

// DryRun handles the common JSON-or-flags → dry-run print pattern
// used by add/update commands.
//
// If jsonData is non-empty, it is unmarshaled into Req for display.
// Otherwise, buildReq is called with validate=false to construct a preview from flags.
func DryRun[Req any](
	cmd *cobra.Command,
	dr *output.DryRunPrinter,
	jsonData []byte,
	buildReq func(*cobra.Command, bool) (*Req, error),
	label, method, apiPath string,
) error {
	var body interface{}

	if len(jsonData) > 0 {
		var req Req
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		body = req
	} else {
		built, _ := buildReq(cmd, false)
		body = built
	}

	dr.PrintOperation(label, method, apiPath, body)
	return nil
}
