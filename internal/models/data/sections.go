// internal/models/data/sections.go
package data

// Section — структура секции TestRail
type Section struct {
	Depth        int64  `json:"depth,omitempty"`         // Глубина вложенности
	Description  string `json:"description,omitempty"`   // Описание секции
	DisplayOrder int64  `json:"display_order,omitempty"` // Порядок отображения
	ID           int64  `json:"id"`                      // Уникальный ID (обязательное)
	Name         string `json:"name"`                    // Название секции (обязательное)
	ParentID     int64  `json:"parent_id,omitempty"`     // ID родительской секции (0 если нет)
	SuiteID      int64  `json:"suite_id,omitempty"`      // ID сьюты (может быть 0 для базовых)
}

// GetSectionsResponse — ответ get_sections (массив секций)
type GetSectionsResponse []Section

// GetSectionResponse — ответ get_section (одна секция)
type GetSectionResponse Section

// AddSectionRequest — запрос add_section (suite_id обязательно)
type AddSectionRequest struct {
	Name        string `json:"name"` // обязательно
	Description string `json:"description,omitempty"`
	SuiteID     int64  `json:"suite_id"` // обязательно
	ParentID    int64  `json:"parent_id,omitempty"`
}

// UpdateSectionRequest — запрос update_section (для изменения/перемещения)
type UpdateSectionRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    int64  `json:"parent_id,omitempty"`
}
