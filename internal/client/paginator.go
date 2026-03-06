package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// paginationLimit — стандартный размер страницы TestRail API.
const paginationLimit = 250

// decodeListResponse декодирует ответ list-endpoint'а TestRail API, который может быть:
//   - Paginated wrapper (TestRail 6.7+): {"offset":0,"limit":250,"size":N,"_links":{...},"<itemsField>":[...]}
//   - Flat array (старые версии TestRail Server):  [item1, item2, ...]
//
// Параметр itemsField — имя JSON-ключа для массива элементов в paginated-объекте
// (например, "runs", "plans", "sections", "milestones", "shared_steps", "tests", "results").
//
// Возвращает (items, pageLen, error), где pageLen — количество элементов на этой странице.
func decodeListResponse[T any](body []byte, itemsField string) (items []T, pageLen int, err error) {
	if len(body) == 0 {
		return nil, 0, nil
	}

	// Определяем формат по первому не-пробельному байту.
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
				// Ключ не найден — возможно другой формат; возвращаем пустой срез
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

// fetchAllPages загружает ВСЕ страницы из paginated list-endpoint'а TestRail API.
// Прозрачно обрабатывает оба формата ответа: paginated wrapper и flat array.
//
//   - c:          HTTP-клиент TestRail
//   - endpoint:   путь API, например "get_runs/30"
//   - baseQuery:  базовые query-параметры, к которым добавятся offset/limit; может быть nil
//   - itemsField: имя JSON-ключа элементов в paginated-ответе, например "runs", "plans"
func fetchAllPages[T any](c *HTTPClient, endpoint string, baseQuery map[string]string, itemsField string) ([]T, error) {
	var all []T
	offset := 0

	for {
		// Строим query с добавлением offset/limit к базовым параметрам
		query := make(map[string]string, len(baseQuery)+2)
		for k, v := range baseQuery {
			query[k] = v
		}
		query["offset"] = fmt.Sprintf("%d", offset)
		query["limit"] = fmt.Sprintf("%d", paginationLimit)

		resp, err := c.Get(endpoint, query)
		if err != nil {
			return nil, fmt.Errorf("fetchAllPages %s (offset=%d): %w", endpoint, offset, err)
		}

		// Явное закрытие в теле цикла — избегаем defer-накопление
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API вернул %s для %s: %s", resp.Status, endpoint, string(body))
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read body %s (offset=%d): %w", endpoint, offset, readErr)
		}

		page, pageLen, decErr := decodeListResponse[T](body, itemsField)
		if decErr != nil {
			return nil, fmt.Errorf("decode %s (offset=%d): %w", endpoint, offset, decErr)
		}

		all = append(all, page...)

		// Если получили меньше limit — больше страниц нет
		if pageLen < paginationLimit {
			break
		}

		offset += paginationLimit
	}

	return all, nil
}
