// internal/models/data/statuses.go
package data

// Status — статус результата теста
type Status struct {
	ID          int64  `json:"id"`           // Уникальный ID статуса
	Name        string `json:"name"`         // Название статуса (например, "Passed", "Failed")
	Label       string `json:"label"`        // Метка для отображения
	ColorBright int64  `json:"color_bright"` // Яркий цвет
	ColorDark   int64  `json:"color_dark"`   // Тёмный цвет
	ColorMedium int64  `json:"color_medium"` // Средний цвет
	IsFinal     bool   `json:"is_final"`     // Финальный статус
	IsSystem    bool   `json:"is_system"`    // Системный статус
	IsUntested  bool   `json:"is_untested"`  // Статус "не протестировано"
}

// GetStatusesResponse — ответ на get_statuses
type GetStatusesResponse []Status
