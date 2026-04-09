// internal/client/wave6g_target_coverage_test.go
// Wave 6G: targeted coverage for reachable zero-count blocks.
//
// Targets:
//   - paginator.go   : whitespace continue, items-unmarshal error, flat-unmarshal error, unexpected format
//   - cases.go       : decodeCasesResponseWithSize whitespace continue and all-whitespace body
//   - milestones.go  : nil-req guards in AddMilestone / UpdateMilestone
//   - request.go     : ReadResponse JSON-unmarshal fallback (non-JSON body)
//   - projects.go    : GetProjects / GetProject transport + decode errors
//   - extended.go    : AddGroup decode error
//   - concurrent.go  : GetCasesParallel / GetSuitesParallel pool.Wait error
package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
)

// ---------------------------------------------------------------------------
// paginator.go — decodeListResponse
// ---------------------------------------------------------------------------

// testDecodeItem is a minimal type for decodeListResponse generic parameter.
type testDecodeItem struct {
	ID int `json:"id"`
}

type readErrCloser struct{}

func (readErrCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("forced read failure")
}

func (readErrCloser) Close() error {
	return nil
}

// TestWave6G_PaginatorDecodeListResponse covers four branches:
//   - leading whitespace (space/tab) before '[' → continue loop then hit '['  (line 29.30,30.12)
//   - items-field value is a non-array (e.g. string) → items-unmarshal error  (line 42.54,44.5)
//   - flat-array body is malformed JSON                → flat-unmarshal error  (line 48.55,50.5)
//   - first non-whitespace byte is not '{' or '['      → unexpected format    (line 57.2,57.20)
func TestWave6G_PaginatorDecodeListResponse(t *testing.T) {
	t.Run("leading whitespace before array", func(t *testing.T) {
		// Body: " [{"id":1}]" — space/tab before '[' triggers the continue branch.
		bodies := []string{
			` [{"id":1}]`,
			"\t[{\"id\":2}]",
			"\n[{\"id\":3}]",
			"\r[{\"id\":4}]",
		}
		for _, b := range bodies {
			items, n, err := decodeListResponse[testDecodeItem]([]byte(b), "items")
			if err != nil {
				t.Errorf("body %q: unexpected error: %v", b, err)
			}
			if n != 1 {
				t.Errorf("body %q: got pageLen=%d, want 1", b, n)
			}
			_ = items
		}
	})

	t.Run("items field is not an array", func(t *testing.T) {
		// items value is a string, not an array → json.Unmarshal on raw fails
		body := []byte(`{"items":"this-is-a-string"}`)
		_, _, err := decodeListResponse[testDecodeItem](body, "items")
		if err == nil {
			t.Fatal("expected decode error for non-array items, got nil")
		}
		if !strings.Contains(err.Error(), "decode") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("flat array malformed JSON", func(t *testing.T) {
		// Flat array starting with '[' but containing bad JSON.
		body := []byte(`[{"id":1}, {bad json}]`)
		_, _, err := decodeListResponse[testDecodeItem](body, "items")
		if err == nil {
			t.Fatal("expected decode error for malformed flat array, got nil")
		}
		if !strings.Contains(err.Error(), "decode flat list") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("unexpected first byte", func(t *testing.T) {
		// Starts with 'x' → neither '{' nor '[' nor whitespace
		body := []byte(`x{"id":1}`)
		_, _, err := decodeListResponse[testDecodeItem](body, "items")
		if err == nil {
			t.Fatal("expected unexpected-format error, got nil")
		}
		if !strings.Contains(err.Error(), "unexpected response format") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("leading whitespace before object", func(t *testing.T) {
		// Leading space then '{' — tests the whitespace continue for object format too.
		body := []byte(` {"items":[{"id":5}]}`)
		items, n, err := decodeListResponse[testDecodeItem](body, "items")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 1 || len(items) != 1 || items[0].ID != 5 {
			t.Errorf("unexpected result: n=%d items=%v", n, items)
		}
	})
}

// ---------------------------------------------------------------------------
// cases.go — decodeCasesResponseWithSize
// ---------------------------------------------------------------------------

// TestWave6G_DecodeCasesResponseWithSize covers:
//   - leading whitespace before '{' → continue loop               (line 35.30,36.12)
//   - body consisting entirely of whitespace → final nil return   (line 60.2,60.21)
func TestWave6G_DecodeCasesResponseWithSize(t *testing.T) {
	t.Run("leading whitespace before paginated object", func(t *testing.T) {
		// Space before '{' triggers the whitespace continue.
		body := []byte(` {"cases":[{"id":10}],"size":1}`)
		cases, size, err := decodeCasesResponseWithSize(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 1 || size != 1 {
			t.Errorf("unexpected result: cases=%v size=%d", cases, size)
		}
	})

	t.Run("leading tab before flat array", func(t *testing.T) {
		body := []byte("\t[{\"id\":11}]")
		cases, _, err := decodeCasesResponseWithSize(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 1 {
			t.Errorf("expected 1 case, got %d", len(cases))
		}
	})

	t.Run("all-whitespace body returns zero values", func(t *testing.T) {
		// Body with only spaces exhausts the loop → hits the final return nil,-1,nil.
		body := []byte("   \t\n")
		cases, size, err := decodeCasesResponseWithSize(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cases != nil || size != -1 {
			t.Errorf("expected nil/-1, got cases=%v size=%d", cases, size)
		}
	})
}

// ---------------------------------------------------------------------------
// milestones.go — nil-req guards
// ---------------------------------------------------------------------------

// TestWave6G_Milestones_NilRequest covers the `if req == nil` guards in
// AddMilestone (line 45.16,47.3) and UpdateMilestone (line 93.16,95.3).
func TestWave6G_Milestones_NilRequest(t *testing.T) {
	// Handler would never be reached but mockClient requires a valid server.
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer s.Close()

	ctx := context.Background()

	t.Run("AddMilestone nil req", func(t *testing.T) {
		_, err := c.AddMilestone(ctx, 1, nil)
		if err == nil {
			t.Fatal("expected error for nil request, got nil")
		}
		if !strings.Contains(err.Error(), "request body is required") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("UpdateMilestone nil req", func(t *testing.T) {
		_, err := c.UpdateMilestone(ctx, 1, nil)
		if err == nil {
			t.Fatal("expected error for nil request, got nil")
		}
		if !strings.Contains(err.Error(), "request body is required") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// request.go — ReadResponse JSON unmarshal fallback
// ---------------------------------------------------------------------------

// TestWave6G_ReadResponse_NonJSONBody covers the fallback branch in ReadResponse
// where json.Unmarshal fails and bodyData is set to the raw string (line 29.16,31.3).
func TestWave6G_ReadResponse_NonJSONBody(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer s.Close()

	// Construct a synthetic *http.Response with a plain-text (non-JSON) body.
	body := io.NopCloser(strings.NewReader("this is not json"))
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       body,
	}

	rd, err := c.ReadResponse(context.Background(), resp, 0, "json")
	if err != nil {
		t.Fatalf("ReadResponse returned unexpected error: %v", err)
	}
	// When JSON unmarshal fails, Body is set to the raw string.
	bodyStr, ok := rd.Body.(string)
	if !ok {
		t.Fatalf("expected Body to be string after non-JSON fallback, got %T", rd.Body)
	}
	if bodyStr != "this is not json" {
		t.Errorf("unexpected body string: %q", bodyStr)
	}
}

// ---------------------------------------------------------------------------
// projects.go — GetProjects / GetProject transport + decode errors
// ---------------------------------------------------------------------------

// TestWave6G_Projects_TransportErrors covers the "request error" return branches
// for GetProjects (line 18.16,20.3) and GetProject (line 35.16,37.3).
func TestWave6G_Projects_TransportErrors(t *testing.T) {
	t.Run("GetProjects transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close() // close immediately → connection refused on first request

		_, err := c.GetProjects(context.Background())
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})

	t.Run("GetProject transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.GetProject(context.Background(), 1)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})
}

// TestWave6G_Projects_DecodeErrors covers the ReadJSONResponse decode-error
// branches for GetProjects (line 24.63,26.3) and GetProject (line 41.63,43.3).
func TestWave6G_Projects_DecodeErrors(t *testing.T) {
	t.Run("GetProjects decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{bad json"))
		})
		defer s.Close()

		_, err := c.GetProjects(context.Background())
		if err == nil {
			t.Fatal("expected decode error, got nil")
		}
		if !strings.Contains(err.Error(), "decode error") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("GetProject decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{bad json"))
		})
		defer s.Close()

		_, err := c.GetProject(context.Background(), 42)
		if err == nil {
			t.Fatal("expected decode error, got nil")
		}
		if !strings.Contains(err.Error(), "decode error") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// extended.go — AddGroup decode error
// ---------------------------------------------------------------------------

// TestWave6G_ExtendedAddGroup_DecodeError covers the decode-error return in
// AddGroup (extended.go:82.66,84.3) when the server returns 200 with bad JSON.
func TestWave6G_ExtendedAddGroup_DecodeError(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not valid json"))
	})
	defer s.Close()

	_, err := c.AddGroup(context.Background(), 1, "test-group", []int64{1, 2})
	if err == nil {
		t.Fatal("expected decode error for AddGroup, got nil")
	}
	if !strings.Contains(err.Error(), "decoding group") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// concurrent.go — pool.Wait() error paths
// ---------------------------------------------------------------------------

// TestWave6G_Concurrent_GetCasesParallel_PoolError covers the pool.Wait error
// return in GetCasesParallel (concurrent.go:47.24,49.3) by using a closed server
// so all tasks fail, causing pool.Wait() to return non-nil.
func TestWave6G_Concurrent_GetCasesParallel_PoolError(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
	s.Close() // requests will fail → errgroup captures the error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.GetCasesParallel(ctx, 1, []int64{101, 102}, 2, nil)
	if err == nil {
		t.Fatal("expected pool error from GetCasesParallel, got nil")
	}
}

// TestWave6G_Concurrent_GetSuitesParallel_PoolError covers the pool.Wait error
// return in GetSuitesParallel (concurrent.go:68.20,70.3).
func TestWave6G_Concurrent_GetSuitesParallel_PoolError(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
	s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.GetSuitesParallel(ctx, []int64{201, 202}, 2, nil)
	if err == nil {
		t.Fatal("expected pool error from GetSuitesParallel, got nil")
	}
}

// TestWave6G_Concurrent_GetCasesForSuitesParallel_AllFail covers the error path
// in GetCasesForSuitesParallel (concurrent.go:101.19,103.3 area) when all tasks fail
// and len(results)==0.
func TestWave6G_Concurrent_GetCasesForSuitesParallel_AllFail(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
	s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.GetCasesForSuitesParallel(ctx, 1, []int64{301, 302}, 2, nil)
	if err == nil {
		t.Fatal("expected error from GetCasesForSuitesParallel, got nil")
	}
}

// ---------------------------------------------------------------------------
// Additional: fetchAllPages decode error via decodeListResponse failure
// (covers paginator.go:93.20,95.4 — decErr != nil in fetchAllPages)
// ---------------------------------------------------------------------------

// TestWave6G_FetchAllPages_DecodeError causes fetchAllPages to receive a response
// that decodeListResponse rejects (unexpected format 'x…'), covering the decErr
// branch (paginator.go:93.20,95.4).
func TestWave6G_FetchAllPages_DecodeError(t *testing.T) {
	// Return a body that starts with 'x' — decodeListResponse returns unexpected format error.
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("xgarbage"))
	})
	defer s.Close()

	ctx := context.Background()
	// GetMilestones internally uses fetchAllPages → decErr != nil propagates up.
	_, err := c.GetMilestones(ctx, 1)
	if err == nil {
		t.Fatal("expected decode error from fetchAllPages, got nil")
	}
	if !strings.Contains(err.Error(), "decode") && !strings.Contains(err.Error(), "unexpected") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestWave6G_FetchAllPages_ReadBodyError(t *testing.T) {
	baseURL, err := url.Parse("http://example.test")
	if err != nil {
		t.Fatalf("failed to parse base URL: %v", err)
	}

	c := &HTTPClient{
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       readErrCloser{},
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}),
		},
		baseURL: baseURL,
	}

	_, err = fetchAllPages[testDecodeItem](context.Background(), c, "get_runs/1", nil, "runs")
	if err == nil {
		t.Fatal("expected read body error, got nil")
	}
	if !strings.Contains(err.Error(), "read body") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Additional: decodeListResponse wrapper unmarshal error
// (paginator.go:37.14,39.4 — wrapper json.Unmarshal fail) — already hit by
// returning '{' body with bad structure.  Re-verify via direct call.
// ---------------------------------------------------------------------------

// TestWave6G_Paginator_WrapperUnmarshalError covers the paginated-wrapper
// unmarshal error branch (paginator.go:37.14,39.4) when the outer JSON is
// malformed but still starts with '{'.
func TestWave6G_Paginator_WrapperUnmarshalError(t *testing.T) {
	body := []byte(`{invalid-json`)
	_, _, err := decodeListResponse[testDecodeItem](body, "items")
	if err == nil {
		t.Fatal("expected unmarshal error for malformed wrapper, got nil")
	}
	if !strings.Contains(err.Error(), "decode paginated wrapper") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// mock.go — cover monitor.Increment() in GetCasesParallel / GetSuitesParallel
// fallback paths. These are MockClient methods with non-nil monitor.
// ---------------------------------------------------------------------------

// mockProgressMonitor is a trivial ProgressMonitor for testing.
type mockProgressMonitor struct {
	count int
}

func (m *mockProgressMonitor) Increment()          { m.count++ }
func (m *mockProgressMonitor) IncrementBy(n int)  { m.count += n }

// TestWave6G_MockClient_MonitorIncrement exercises the mock GetCasesParallel
// default path (sequential loop) with a non-nil monitor, covering
// mock.go:350.21,352.4 (monitor.Increment) and mock.go:362/366 variants.
func TestWave6G_MockClient_MonitorIncrement(t *testing.T) {
	m := &MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{data.Case{ID: suiteID}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{data.Suite{ID: projectID}}, nil
		},
	}

	ctx := context.Background()

	t.Run("GetCasesParallel sequential with monitor", func(t *testing.T) {
		mon := &mockProgressMonitor{}
		results, err := m.GetCasesParallel(ctx, 1, []int64{10, 20}, 2, mon)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
		if mon.count != 2 {
			t.Errorf("expected 2 Increment calls, got %d", mon.count)
		}
	})

	t.Run("GetSuitesParallel with monitor", func(t *testing.T) {
		mon := &mockProgressMonitor{}
		results, err := m.GetSuitesParallel(ctx, []int64{30, 40}, 2, mon)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
		if mon.count != 2 {
			t.Errorf("expected 2 Increment calls, got %d", mon.count)
		}
	})
}

// ---------------------------------------------------------------------------
// mock.go line 375 — GetSuitesParallel error return in MockClient default loop
// ---------------------------------------------------------------------------

// TestWave6G_MockClient_GetSuitesParallel_Error covers the error-return path
// inside MockClient.GetSuitesParallel sequential fallback (mock.go:362.17,364.4).
func TestWave6G_MockClient_GetSuitesParallel_Error(t *testing.T) {
	m := &MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			if projectID == 999 {
				return nil, io.ErrUnexpectedEOF
			}
			return data.GetSuitesResponse{data.Suite{ID: projectID}}, nil
		},
	}

	ctx := context.Background()
	results, err := m.GetSuitesParallel(ctx, []int64{1, 999, 2}, 2, nil)
	if err == nil {
		t.Fatal("expected error from GetSuitesParallel, got nil")
	}
	// Partial results before the failure are returned.
	_ = results
}

// ---------------------------------------------------------------------------
// mock.go line 350 — GetCasesParallel error return in MockClient default loop
// ---------------------------------------------------------------------------

// TestWave6G_MockClient_GetCasesParallel_Error covers the error-return path
// inside MockClient.GetCasesParallel sequential fallback (mock.go:350.21,352.4).
func TestWave6G_MockClient_GetCasesParallel_Error(t *testing.T) {
	m := &MockClient{
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			if suiteID == 888 {
				return nil, io.ErrUnexpectedEOF
			}
			return data.GetCasesResponse{data.Case{ID: suiteID}}, nil
		},
	}

	ctx := context.Background()
	results, err := m.GetCasesParallel(ctx, 1, []int64{10, 888, 20}, 2, nil)
	if err == nil {
		t.Fatal("expected error from GetCasesParallel, got nil")
	}
	// Partial results may be present.
	_ = results
}

// ---------------------------------------------------------------------------
// Decode-error in UpdateProject (projects.go:64.63,66.3 area?)
// and AddProject decode is already covered in wave6c — skip duplicates.
// ---------------------------------------------------------------------------

// TestWave6G_Projects_UpdateProject_DecodeError covers the decode error in
// UpdateProject (the ReadJSONResponse path).
func TestWave6G_Projects_UpdateProject_DecodeError(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{bad"))
	})
	defer s.Close()

	_, err := c.UpdateProject(context.Background(), 1, &data.UpdateProjectRequest{})
	if err == nil {
		t.Fatal("expected decode error for UpdateProject, got nil")
	}
}

// TestWave6G_Projects_AddProject_DecodeError covers the decode error in
// AddProject (the ReadJSONResponse path after successful POST).
func TestWave6G_Projects_AddProject_DecodeError(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{bad"))
	})
	defer s.Close()

	_, err := c.AddProject(context.Background(), &data.AddProjectRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected decode error for AddProject, got nil")
	}
}

// ---------------------------------------------------------------------------
// Helper: verify MockClient.GetCasesForSuitesParallel with GetCasesParallelFunc
// covering mock.go:375.44,377.3
// ---------------------------------------------------------------------------

// TestWave6G_MockClient_GetCasesForSuitesParallel_WithFunc covers the
// GetCasesParallelFunc branch in MockClient.GetCasesForSuitesParallel
// and the GetCasesForSuitesParallelFunc branch.
func TestWave6G_MockClient_GetCasesForSuitesParallel_WithFunc(t *testing.T) {
	called := false
	m := &MockClient{
		GetCasesForSuitesParallelFunc: func(ctx context.Context, projectID int64, suiteIDs []int64, workers int) (data.GetCasesResponse, error) {
			called = true
			var res data.GetCasesResponse
			for _, sid := range suiteIDs {
				res = append(res, data.Case{ID: sid})
			}
			return res, nil
		},
	}

	ctx := context.Background()
	res, err := m.GetCasesForSuitesParallel(ctx, 1, []int64{1, 2}, 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("GetCasesForSuitesParallelFunc was not called")
	}
	if len(res) != 2 {
		t.Errorf("expected 2 cases (flattened), got %d", len(res))
	}
}

// TestWave6G_JSON_Usage ensures json import is used (avoids unused import).
func TestWave6G_JSON_Usage(t *testing.T) {
	// json.NewDecoder is used in tests via wave6c — this confirms package compiles.
	_ = json.RawMessage(`{}`)
}

// ---------------------------------------------------------------------------
// paginator.go — decodeListResponse all-whitespace body (line 57.2,57.20)
// ---------------------------------------------------------------------------

// TestWave6G_PaginatorDecodeListResponse_AllWhitespace covers the final
// `return nil, 0, nil` in decodeListResponse when body is only whitespace
// (the loop exhausts without hitting any non-whitespace byte) — line 57.2,57.20.
func TestWave6G_PaginatorDecodeListResponse_AllWhitespace(t *testing.T) {
	bodies := [][]byte{
		[]byte("   "),
		[]byte("\t\n\r "),
		[]byte(""),
	}
	for _, b := range bodies {
		items, n, err := decodeListResponse[testDecodeItem](b, "items")
		if err != nil {
			t.Errorf("body %q: unexpected error: %v", string(b), err)
		}
		if items != nil || n != 0 {
			t.Errorf("body %q: expected nil/0, got items=%v n=%d", string(b), items, n)
		}
	}
}

// ---------------------------------------------------------------------------
// concurrent.go — empty-input early-returns and workers=0 default
// ---------------------------------------------------------------------------

// TestWave6G_Concurrent_EmptyInputPaths covers:
//   - GetCasesParallel with empty suiteIDs      → line 47.24,49.3
//   - GetCasesParallel with workers=0           → line 51.18,53.3
//   - GetCasesParallel with non-nil monitor     → line 68.20,70.3
//   - GetCasesForSuitesParallel empty suiteIDs  → line 194.24,196.3
func TestWave6G_Concurrent_EmptyInputPaths(t *testing.T) {
	// Handler that returns OK with empty JSON array for any request.
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	})
	defer s.Close()

	ctx := context.Background()

	t.Run("GetCasesParallel empty suiteIDs", func(t *testing.T) {
		// covers concurrent.go:47.24,49.3 — early return for empty input
		result, err := c.GetCasesParallel(ctx, 1, []int64{}, 5, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("GetCasesParallel workers=0 uses default", func(t *testing.T) {
		// covers concurrent.go:51.18,53.3 — workers <= 0 sets defaultWorkers
		result, err := c.GetCasesParallel(ctx, 1, []int64{1}, 0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = result
	})

	t.Run("GetCasesParallel with monitor", func(t *testing.T) {
		// covers concurrent.go:68.20,70.3 — monitor != nil branch
		mon := &mockProgressMonitor{}
		result, err := c.GetCasesParallel(ctx, 1, []int64{1}, 1, mon)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = result
		// monitor.Increment is called via the pool's WithProgressMonitor
	})

	t.Run("GetCasesForSuitesParallel empty suiteIDs", func(t *testing.T) {
		// covers concurrent.go:194.24,196.3 — early return for empty input
		result, err := c.GetCasesForSuitesParallel(ctx, 1, []int64{}, 5, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})
}

// TestWave6G_Concurrent_GetSuitesParallel_EmptyAndDefault covers:
//   - GetSuitesParallel with empty projectIDs
//   - GetSuitesParallel with workers=0
func TestWave6G_Concurrent_GetSuitesParallel_EmptyAndDefault(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	})
	defer s.Close()

	ctx := context.Background()

	t.Run("GetSuitesParallel empty projectIDs", func(t *testing.T) {
		result, err := c.GetSuitesParallel(ctx, []int64{}, 5, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("GetSuitesParallel workers=0", func(t *testing.T) {
		result, err := c.GetSuitesParallel(ctx, []int64{1}, 0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = result
	})
}

// ---------------------------------------------------------------------------
// cases.go — GetCasesParallelCtx with nil config (line 564.19,566.3)
// ---------------------------------------------------------------------------

// TestWave6G_Cases_GetCasesParallelCtx_NilConfig covers the
// `config = concurrency.DefaultControllerConfig()` branch when config is nil
// (cases.go:564.19,566.3).
func TestWave6G_Cases_GetCasesParallelCtx_NilConfig(t *testing.T) {
	// Handler serving empty cases for any get_cases request.
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cases":[],"size":0}`))
	})
	defer s.Close()

	ctx := context.Background()

	// Passing nil config triggers the defaultControllerConfig branch.
	result, _, err := c.GetCasesParallelCtx(ctx, 1, []int64{1}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// ---------------------------------------------------------------------------
// Ensure concurrency import is used
// ---------------------------------------------------------------------------

func TestWave6G_ConcurrencyImportUsed(t *testing.T) {
	cfg := concurrency.DefaultControllerConfig()
	if cfg == nil {
		t.Fatal("expected non-nil default config")
	}
}
