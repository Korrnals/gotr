# Stage 13 - API Compliance Matrix

Дата: 2026-03-27
Ветка: stage-13.0-final-refactoring
Шаг: Phase 2 → Step 4

## Scope

Аудит HTTP-слоя gotr: internal/client, pkg/testrailapi, internal/service.

---

## A. Transport & Auth

| Check | Status | Details |
| --- | --- | --- |
| Basic Auth на каждый запрос | PASS | authTransport.RoundTrip устанавливает SetBasicAuth. |
| Content-Type: application/json | PASS | authTransport выставляет заголовок если не задан. |
| User-Agent | PASS | gotr/2.7 User-Agent установлен в authTransport. |
| InsecureSkipVerify | CONDITIONAL | Отключается только явно через WithSkipTlsVerify(true). Приемлемо. |
| Context propagation | PASS | DoRequest использует http.NewRequestWithContext(ctx, ...). |
| Timeout по умолчанию | PASS | 30s по умолчанию, конфигурируется через WithTimeout. |
| MaxConnsPerHost | PASS | 0 (unlimited) — concurrency управляется на уровне parallel settings. |

---

## B. URL Construction

| Check | Status | Details |
| --- | --- | --- |
| API prefix константа | PASS | `const apiPrefix = "index.php?/api/v2/"` — единственная точка определения. |
| Прямые endpoint-строки вне client | PASS | 0 вхождений `index.php?/api` вне internal/client. |
| Query params через url.Values | PASS | DoRequest использует url.Values для query params. |
| URL-параметры в TestRail (& vs ?) | PASS | `fullURL += "&" + q.Encode()` — корректно для TestRail. |
| Base URL нормализация | PASS | NewClient разбирает только scheme+host, игнорирует лишние path-части. |

---

## C. Response Handling

| Check | Status | Details |
| --- | --- | --- |
| Body.Close в paginator | PASS | paginator.go:87 — явный resp.Body.Close() в теле цикла. |
| Body.Close в Get/Post обёртках | PASS | formatAPIError закрывает body для non-200. |
| Body.Close в ReadJSONResponse_OK | PASS | `defer resp.Body.Close()` перед decode. |
| **Body.Close в ReadJSONResponse_ERR** | **FAIL** | L54: `io.ReadAll(resp.Body)` без `defer resp.Body.Close()` в non-200 ветке. Connection leak при API ошибках. |
| decode error wrapping | PASS | `fmt.Errorf("decode error: %w", err)` — стандартный wrapping. |
| API error body included | PASS | formatAPIError читает + включает body в сообщение об ошибке. |

---

## D. Interface & Contract Coverage

| Check | Status | Details |
| --- | --- | --- |
| Compile-time interface check | PASS | `var _ ClientInterface = (*HTTPClient)(nil)` в interfaces.go:248. |
| Total HTTPClient methods | 139 | grep-счёт по internal/client/*.go. |
| Interface signatures | 144 | Включая overloads и parallel variants. |
| ClientInterface composite | PASS | Правильно объединяет 15 суб-интерфейсов через embedding. |
| Group/Role/Dataset/Variable/BDD/Label API | NOTE | Инлайнены в composite interface напрямую, без отдельных API-интерфейсных типов. Low priority. |
| MockClient coverage | PASS | mock.go реализует ClientInterface для тестирования. |

---

## E. Pagination

| Check | Status | Details |
| --- | --- | --- |
| Dual-mode (wrapper vs flat array) | PASS | decodeListResponse поддерживает paginated wrapper (TR 6.7+) и flat array (старые TR). |
| fetchAllPages loop | PASS | Корректный цикл с explicit body.Close на каждую страницу. |
| paginationLimit константа | PASS | 250 — стандартный лимит TestRail API. |

---

## F. Security

| Check | Status | Details |
| --- | --- | --- |
| TLS настройка | PASS | TLS включен по умолчанию, InsecureSkipVerify=false. |
| Credentials exposure в логах | PASS | Debug-логи не содержат credentials: только URL и endpoint. |
| Credentials в User-Agent | PASS | User-Agent содержит только gotr version. |

---

## Findings Summary

| ID | Severity | Location | Description |
| --- | --- | --- | --- |
| F5 | MEDIUM | internal/client/request.go:54 | Body leak в ReadJSONResponse: non-200 ветка вызывает io.ReadAll(resp.Body) без defer resp.Body.Close(). При API ошибках TCP соединение не возвращается в пул. |
| F6 | LOW | internal/client/interfaces.go | Group/Role/Dataset/Variable/BDD/Label методы инлайнены без отдельных интерфейсных типов. Нет отдельного GroupsAPI, RolesAPI, etc. для изолированного мокирования. |
| F7 | INFO | internal/client/ | 139 HTTPClient методов vs 144 interface signatures — расхождение ожидаемо (parallel variants, overloads). |

---

## Remediation

| ID | Action |
| --- | --- |
| R5 (MEDIUM) | Fix ReadJSONResponse: добавить `defer resp.Body.Close()` перед if-блоком non-200, убрать прямой io.ReadAll. |
| R6 (LOW) | Add GroupsAPI, RolesAPI, DatasetsAPI, VariablesAPI, BDDsAPI, LabelsAPI interfaces и встроить их в ClientInterface — для улучшенного мокирования по зонам ответственности. |

---

## Status

- API Transport/Auth: PASS (no action needed).
- URL construction: PASS (no action needed).
- Response handling: 1 MEDIUM bug identified (R5).
- Interface coverage: PASS with 1 LOW note (R6).
- Pagination: PASS.
- Security: PASS.
