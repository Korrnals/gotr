// Package crud provides generic executor functions for CRUD command patterns.
// It eliminates boilerplate JSON/flags parsing, API call wrapping, and output
// handling shared across add/update commands.
package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

// SnapHint provides snapshot metadata for crud.Execute auto-snapping.
// When provided, Execute will take a snapshot before mutation and finalize afterward.
type SnapHint struct {
	Op         snap.Operation
	EntityType string
	EntityIDs  []int64
	Tier       snap.Tier
	ProjectID  int64
	SuiteID    int64
	// FetchFn fetches the current entity state before mutation (for update ops).
	// Pass nil for add operations.
	FetchFn func(ctx context.Context) (interface{}, error)
}

// Execute handles the common JSON-or-flags → API call → output pattern
// used by add/update commands.
//
// If jsonData is non-empty, it is unmarshaled into Req.
// Otherwise, buildReq is called with validate=true to construct the request from flags.
// The API call is then made and the result is output via output.OutputResult.
//
// When snapHint is provided, a pre-mutation snapshot is taken automatically.
// For add operations (Tier2), the created entity ID is extracted via idExtractor.
func Execute[Req any, Resp any](
	cmd *cobra.Command,
	id int64,
	jsonData []byte,
	buildReq func(*cobra.Command, bool) (*Req, error),
	apiCall func(context.Context, int64, *Req) (Resp, error),
	failMsg string,
	snapHint ...SnapHint,
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

	// Snap: before mutation.
	var hook *snap.Hook
	if len(snapHint) > 0 {
		h := snapHint[0]
		hook = snap.NewHook(cmd)
		hook.Before(ctx, snap.BuildMeta(
			h.Op, h.EntityType, h.EntityIDs,
			h.Tier, h.ProjectID, h.SuiteID,
			snap.ResolveName(cmd), cmdArgs(cmd),
		), h.FetchFn)
	}

	result, err := apiCall(ctx, id, &req)
	if err != nil {
		return fmt.Errorf("%s: %w", failMsg, err)
	}

	// Snap: finalize add (extract ID from result if available).
	if hook != nil && len(snapHint) > 0 && snapHint[0].Op == snap.OpAdd {
		if rid := extractID(result); rid != 0 {
			hook.FinalizeAdd(rid)
		}
	}

	return output.OutputResult(cmd, result, "result")
}

// cmdArgs returns the CLI arguments for the current command (for snap metadata).
func cmdArgs(cmd *cobra.Command) []string {
	var args []string
	for p := cmd; p != nil; p = p.Parent() {
		if p.Name() != "" {
			args = append([]string{p.Name()}, args...)
		}
	}
	return args
}

// extractID uses reflection to find an ID int64 field in the response.
func extractID(v any) int64 {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return 0
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0
	}
	f := rv.FieldByName("ID")
	if !f.IsValid() || f.Kind() != reflect.Int64 {
		return 0
	}
	return f.Int()
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
