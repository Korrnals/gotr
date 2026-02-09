// internal/models/data/priorities.go
package data

// Priority — приоритет тест-кейса
type Priority struct {
	ID        int64  `json:"id"`         // Уникальный ID приоритета
	Name      string `json:"name"`       // Название приоритета (например, "Critical", "High")
	ShortName string `json:"short_name"` // Короткое название
	Color     string `json:"color"`      // Цвет (HEX)
	IsDefault bool   `json:"is_default"` // Приоритет по умолчанию
	Priority  int    `json:"priority"`   // Числовое значение приоритета (чем меньше, тем выше)
}

// GetPrioritiesResponse — ответ на get_priorities
type GetPrioritiesResponse []Priority
