// internal/models/data/statuses.go
package data

// Status — статус теста (passed, failed, blocked и т.д.)
// Соответствует документации TestRail API v2
// https://support.testrail.com/hc/en-us/articles/7077868951828-Statuses
type Status struct {
	ID          int64  `json:"id"`              // The unique ID of the status
	Name        string `json:"name"`            // The name of the status (e.g., "passed", "failed")
	Label       string `json:"label"`           // The label of the status (human-readable)
	ColorBright int64  `json:"color_bright"`    // The bright color for the status (RGB)
	ColorMedium int64  `json:"color_medium"`    // The medium color for the status (RGB)
	ColorDark   int64  `json:"color_dark"`      // The dark color for the status (RGB)
	IsFinal     bool   `json:"is_final"`        // True if the status is a final status
	IsSystem    bool   `json:"is_system"`       // True if the status is a system status
	IsUntested  bool   `json:"is_untested"`     // True if the status is the "Untested" status
}

// GetStatusesResponse — ответ на get_statuses (массив статусов)
type GetStatusesResponse []Status

// Стандартные статусы TestRail (для справки)
// Фактические ID могут отличаться в зависимости от конфигурации TestRail
const (
	StatusUntested  = 3  // Untested (default)
	StatusPassed    = 1  // Passed
	StatusBlocked   = 2  // Blocked
	StatusRetest    = 4  // Retest
	StatusFailed    = 5  // Failed
	// Custom statuses могут иметь другие ID
)
