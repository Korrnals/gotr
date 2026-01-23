// internal/models/data/sections.go
package data

// Section — структура секции TestRail
type Section struct {
	Depth        int64  `json:"depth"`
	Description  string `json:"description"`
	DisplayOrder int64  `json:"display_order"`
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ParentID     int64  `json:"parent_id"`
	SuiteID      int64  `json:"suite_id"`
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
