// internal/client/paginator_test.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// ── decodeListResponse ────────────────────────────────────────────────────────

func TestDecodeListResponse_FlatArray(t *testing.T) {
	type item struct{ ID int }
	body := `[{"ID":1},{"ID":2},{"ID":3}]`

	items, pageLen, err := decodeListResponse[item]([]byte(body), "items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageLen != 3 {
		t.Errorf("pageLen = %d, want 3", pageLen)
	}
	if len(items) != 3 {
		t.Errorf("len(items) = %d, want 3", len(items))
	}
	if items[0].ID != 1 || items[2].ID != 3 {
		t.Errorf("unexpected items: %+v", items)
	}
}

func TestDecodeListResponse_PaginatedWrapper(t *testing.T) {
	type item struct{ ID int }
	body := `{"offset":0,"limit":250,"size":2,"_links":{"next":null,"prev":null},"items":[{"ID":10},{"ID":20}]}`

	items, pageLen, err := decodeListResponse[item]([]byte(body), "items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageLen != 2 {
		t.Errorf("pageLen = %d, want 2", pageLen)
	}
	if len(items) != 2 || items[0].ID != 10 || items[1].ID != 20 {
		t.Errorf("unexpected items: %+v", items)
	}
}

func TestDecodeListResponse_EmptyBody(t *testing.T) {
	type item struct{ ID int }
	items, pageLen, err := decodeListResponse[item]([]byte{}, "items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageLen != 0 || len(items) != 0 {
		t.Errorf("expected empty result, got %d items (pageLen=%d)", len(items), pageLen)
	}
}

func TestDecodeListResponse_MissingItemsField(t *testing.T) {
	type item struct{ ID int }
	// Field "runs" exists, but we request "plans" — should return (nil, 0, nil)
	body := `{"offset":0,"limit":250,"size":1,"runs":[{"ID":1}]}`

	items, pageLen, err := decodeListResponse[item]([]byte(body), "plans")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageLen != 0 || len(items) != 0 {
		t.Errorf("expected empty result for missing field, got %d items", len(items))
	}
}

func TestDecodeListResponse_InvalidJSON(t *testing.T) {
	type item struct{ ID int }
	body := `{invalid json`
	_, _, err := decodeListResponse[item]([]byte(body), "items")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestDecodeListResponse_UnexpectedFormat(t *testing.T) {
	type item struct{ ID int }
	body := `42` // neither { nor [
	_, _, err := decodeListResponse[item]([]byte(body), "items")
	if err == nil {
		t.Fatal("expected error for unexpected format, got nil")
	}
}

// ── fetchAllPages ──────────────────────────────────────────────────────────────

// testItem — test type for fetchAllPages
type testItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestFetchAllPages_FlatArray_BackwardCompat(t *testing.T) {
	body := `[{"id":1,"name":"a"},{"id":2,"name":"b"}]`

	c, srv := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	})
	defer srv.Close()

	ctx := context.Background()

	items, err := fetchAllPages[testItem](ctx, c, "get_items/1", nil, "items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
}

func TestFetchAllPages_FlatArray_AtOrAboveLimit_StopsAfterFirstRequest(t *testing.T) {
	requestCount := 0

	items := make([]testItem, paginationLimit)
	for i := range items {
		items[i] = testItem{ID: i + 1, Name: fmt.Sprintf("item-%d", i+1)}
	}
	bodyBytes, _ := json.Marshal(items)

	c, srv := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	})
	defer srv.Close()

	ctx := context.Background()
	got, err := fetchAllPages[testItem](ctx, c, "get_sections/1", map[string]string{"suite_id": "42"}, "sections")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != paginationLimit {
		t.Errorf("len(got) = %d, want %d", len(got), paginationLimit)
	}
	if requestCount != 1 {
		t.Errorf("requestCount = %d, want 1", requestCount)
	}
}

func TestFetchAllPages_SinglePage_PaginatedWrapper(t *testing.T) {
	// 3 elements < 250 → only one page
	items := []testItem{{ID: 1}, {ID: 2}, {ID: 3}}
	body, _ := json.Marshal(map[string]interface{}{
		"offset":  0,
		"limit":   250,
		"size":    3,
		"_links":  map[string]interface{}{"next": nil, "prev": nil},
		"results": items,
	})

	c, srv := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})
	defer srv.Close()

	ctx := context.Background()

	got, err := fetchAllPages[testItem](ctx, c, "get_items/1", nil, "results")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("len = %d, want 3", len(got))
	}
}

func TestFetchAllPages_MultiPage(t *testing.T) {
	// Page 1: 250 elements (offset=0) → triggers page 2
	// Page 2: 1 element (offset=250) → end
	requestCount := 0

	page1 := make([]testItem, 250)
	for i := range page1 {
		page1[i] = testItem{ID: i + 1}
	}
	page2 := []testItem{{ID: 251}}

	makeWrapper := func(items []testItem, field string) []byte {
		b, _ := json.Marshal(map[string]interface{}{
			"offset": 0,
			"limit":  250,
			"size":   len(items),
			"_links": map[string]interface{}{"next": nil, "prev": nil},
			field:    items,
		})
		return b
	}

	c, srv := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("offset") == "0" {
			w.Write(makeWrapper(page1, "runs"))
		} else {
			w.Write(makeWrapper(page2, "runs"))
		}
	})
	defer srv.Close()

	ctx := context.Background()

	got, err := fetchAllPages[testItem](ctx, c, "get_runs/1", nil, "runs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 251 {
		t.Errorf("len = %d, want 251", len(got))
	}
	if requestCount != 2 {
		t.Errorf("requestCount = %d, want 2 (two pages)", requestCount)
	}
}

func TestFetchAllPages_ServerError(t *testing.T) {
	c, srv := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"internal server error"}`)
	})
	defer srv.Close()

	ctx := context.Background()

	_, err := fetchAllPages[testItem](ctx, c, "get_items/1", nil, "items")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestFetchAllPages_BaseQueryPreserved(t *testing.T) {
	// baseQuery must be preserved across all pages
	receivedParams := map[string]string{}

	c, srv := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		receivedParams["suite_id"] = r.URL.Query().Get("suite_id")
		receivedParams["offset"] = r.URL.Query().Get("offset")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[]`) // flat empty — one request, done
	})
	defer srv.Close()

	baseQuery := map[string]string{"suite_id": "42"}
	ctx := context.Background()
	_, err := fetchAllPages[testItem](ctx, c, "get_sections/1", baseQuery, "sections")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedParams["suite_id"] != "42" {
		t.Errorf("suite_id = %q, want %q", receivedParams["suite_id"], "42")
	}
	if receivedParams["offset"] != "0" {
		t.Errorf("offset = %q, want %q", receivedParams["offset"], "0")
	}
}
