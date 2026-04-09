package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// paginationLimit is the default page size for TestRail API.
const paginationLimit = 250

// decodeListResponse decodes a TestRail list-endpoint response, which can be:
//   - Paginated wrapper (TestRail 6.7+): {"offset":0,"limit":250,"size":N,"_links":{...},"<itemsField>":[...]}
//   - Flat array (older TestRail Server):  [item1, item2, ...]
//
// itemsField is the JSON key for the items array in the paginated object
// (e.g. "runs", "plans", "sections", "milestones", "shared_steps", "tests", "results").
//
// Returns (items, pageLen, error), where pageLen is the number of items on this page.
func decodeListResponse[T any](body []byte, itemsField string) (items []T, pageLen int, err error) {
	if len(body) == 0 {
		return nil, 0, nil
	}

	// Detect format by first non-whitespace byte.
	for _, b := range body {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			// Paginated wrapper: {"runs":[...], "offset":0, "limit":250, "size":N, ...}
			var wrapper map[string]json.RawMessage
			if err := json.Unmarshal(body, &wrapper); err != nil {
				return nil, 0, fmt.Errorf("decode paginated wrapper: %w", err)
			}
			raw, ok := wrapper[itemsField]
			if !ok {
				// Key not found — possibly a different format; return empty slice
				return nil, 0, nil
			}
			if err := json.Unmarshal(raw, &items); err != nil {
				return nil, 0, fmt.Errorf("decode %q items: %w", itemsField, err)
			}
			return items, len(items), nil
		case '[':
			// Flat array: [item1, item2, ...]
			if err := json.Unmarshal(body, &items); err != nil {
				return nil, 0, fmt.Errorf("decode flat list: %w", err)
			}
			return items, len(items), nil
		default:
			return nil, 0, fmt.Errorf("unexpected response format (starts with %q)", string([]byte{b}))
		}
	}

	return nil, 0, nil
}

// fetchAllPages loads ALL pages from a paginated TestRail list-endpoint.
// Transparently handles both response formats: paginated wrapper and flat array.
//
//   - c:          TestRail HTTP client
//   - endpoint:   API path, e.g. "get_runs/30"
//   - baseQuery:  base query parameters; offset/limit will be appended; may be nil
//   - itemsField: JSON key for items in the paginated response, e.g. "runs", "plans"
func fetchAllPages[T any](ctx context.Context, c *HTTPClient, endpoint string, baseQuery map[string]string, itemsField string) ([]T, error) {
	var all []T
	offset := 0

	for {
		// Build query by appending offset/limit to base parameters
		query := make(map[string]string, len(baseQuery)+2)
		for k, v := range baseQuery {
			query[k] = v
		}
		query["offset"] = fmt.Sprintf("%d", offset)
		query["limit"] = fmt.Sprintf("%d", paginationLimit)

		resp, err := c.Get(ctx, endpoint, query)
		if err != nil {
			return nil, fmt.Errorf("fetchAllPages %s (offset=%d): %w", endpoint, offset, err)
		}

		// Explicit close inside loop body — avoids defer accumulation
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read body %s (offset=%d): %w", endpoint, offset, readErr)
		}

		page, pageLen, decErr := decodeListResponse[T](body, itemsField)
		if decErr != nil {
			return nil, fmt.Errorf("decode %s (offset=%d): %w", endpoint, offset, decErr)
		}

		all = append(all, page...)

		// Backward-compat mode: flat array responses are not paginated by TestRail.
		// If the payload starts with '[', stop after the first successful request,
		// even when len(page) >= paginationLimit.
		if isJSONArrayBody(body) {
			break
		}

		// Fewer items than limit means no more pages
		if pageLen < paginationLimit {
			break
		}

		offset += paginationLimit
	}

	return all, nil
}

func isJSONArrayBody(body []byte) bool {
	for _, b := range body {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '[':
			return true
		default:
			return false
		}
	}
	return false
}
