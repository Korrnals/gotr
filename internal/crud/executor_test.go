package crud_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/crud"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

type testReq struct {
	Name string `json:"name"`
}

type testResp struct {
	ID int64 `json:"id"`
}

// --- DryRun tests ---

func TestDryRun_WithJSON(t *testing.T) {
	cmd := &cobra.Command{}
	jsonData, _ := json.Marshal(testReq{Name: "test"})
	dr := output.NewDryRunPrinter("test")

	err := crud.DryRun(cmd, dr, jsonData,
		func(_ *cobra.Command, _ bool) (*testReq, error) {
			t.Fatal("buildReq should not be called when JSON data is provided")
			return nil, nil
		},
		"Test Op", "POST", "/api/v2/test/1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDryRun_WithFlags(t *testing.T) {
	cmd := &cobra.Command{}

	called := false
	err := crud.DryRun(cmd, output.NewDryRunPrinter("test"), nil,
		func(_ *cobra.Command, validate bool) (*testReq, error) {
			called = true
			if validate {
				t.Fatal("DryRun should call buildReq with validate=false")
			}
			return &testReq{Name: "from-flags"}, nil
		},
		"Test Op", "POST", "/api/v2/test/1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("buildReq was not called")
	}
}

func TestDryRun_InvalidJSON(t *testing.T) {
	cmd := &cobra.Command{}

	err := crud.DryRun(cmd, output.NewDryRunPrinter("test"), []byte("{invalid"),
		func(_ *cobra.Command, _ bool) (*testReq, error) {
			return nil, nil
		},
		"Test Op", "POST", "/api/v2/test/1",
	)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- Execute tests (error paths only; happy path covered by cmd tests) ---

func TestExecute_BuildReqError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	expectedErr := fmt.Errorf("--name is required")
	err := crud.Execute(cmd, 0, nil,
		func(_ *cobra.Command, _ bool) (*testReq, error) {
			return nil, expectedErr
		},
		func(_ context.Context, _ int64, _ *testReq) (*testResp, error) {
			t.Fatal("apiCall should not be called on buildReq error")
			return nil, nil
		},
		"failed",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "--name is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute_APICallError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	output.AddFlag(cmd)

	jsonData, _ := json.Marshal(testReq{Name: "test"})
	err := crud.Execute(cmd, 1, jsonData,
		func(_ *cobra.Command, _ bool) (*testReq, error) {
			return nil, nil
		},
		func(_ context.Context, _ int64, _ *testReq) (*testResp, error) {
			return nil, fmt.Errorf("API error")
		},
		"failed to create",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "failed to create: API error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute_InvalidJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := crud.Execute(cmd, 0, []byte("{bad json"),
		func(_ *cobra.Command, _ bool) (*testReq, error) {
			return nil, nil
		},
		func(_ context.Context, _ int64, _ *testReq) (*testResp, error) {
			return nil, nil
		},
		"failed",
	)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExecute_ValidateCalledWithTrue(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := crud.Execute(cmd, 0, nil,
		func(_ *cobra.Command, validate bool) (*testReq, error) {
			if !validate {
				t.Fatal("Execute should call buildReq with validate=true")
			}
			return nil, fmt.Errorf("stop here")
		},
		func(_ context.Context, _ int64, _ *testReq) (*testResp, error) {
			return nil, nil
		},
		"failed",
	)
	if err == nil || err.Error() != "stop here" {
		t.Fatalf("unexpected error: %v", err)
	}
}
