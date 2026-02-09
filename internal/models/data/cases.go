// models/data/cases.go
package data

import (
	"encoding/json"
)

// Case — основная структура одного кейса (используется в get_case и get_cases)
type Case struct {
	ID                   int64   `json:"id"`
	Title                string  `json:"title,omitempty"`
	SectionID            int64   `json:"section_id,omitempty"`
	TemplateID           int64   `json:"template_id,omitempty"`
	TypeID               int64   `json:"type_id,omitempty"`
	PriorityID           int64   `json:"priority_id,omitempty"`
	MilestoneID          int64   `json:"milestone_id,omitempty"`
	Refs                 string  `json:"refs,omitempty"`
	CreatedBy            int64   `json:"created_by,omitempty"`
	CreatedOn            int64   `json:"created_on,omitempty"`
	UpdatedBy            int64   `json:"updated_by,omitempty"`
	UpdatedOn            int64   `json:"updated_on,omitempty"`
	Estimate             string  `json:"estimate,omitempty"`
	EstimateForecast     string  `json:"estimate_forecast,omitempty"`
	SuiteID              int64   `json:"suite_id,omitempty"`
	DisplayOrder         int     `json:"display_order,omitempty"`
	IsDeleted            int     `json:"is_deleted,omitempty"` // 1 = deleted, 0 = not deleted
	CustomAutomationType int64   `json:"custom_automation_type,omitempty"`
	CustomPreconds       string  `json:"custom_preconds,omitempty"`
	CustomSteps          string  `json:"custom_steps,omitempty"`
	CustomExpected       string  `json:"custom_expected,omitempty"`
	CustomStepsSeparated []Step  `json:"custom_steps_separated,omitempty"` // Step из shared.go
	CustomMission        string  `json:"custom_mission,omitempty"`
	CustomGoals          string  `json:"custom_goals,omitempty"`
	Labels               []Label `json:"labels,omitempty"` // Label из shared.go
	// Для полностью кастомных полей (если TestRail вернёт неизвестное)
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// GetCasesResponse — ответ на get_cases (пагинированный список кейсов)
type GetCasesResponse []Case

// GetCaseResponse — ответ на get_case (один кейс напрямую)
type GetCaseResponse Case

// GetHistoryForCaseResponse — ответ на get_history_for_case
type GetHistoryForCaseResponse struct {
	Pagination
	History []struct {
		ID        int64    `json:"id"`
		TypeID    int64    `json:"type_id"`
		CreatedOn int64    `json:"created_on"`
		UserID    int64    `json:"user_id"`
		Changes   []Change `json:"changes"`
	} `json:"history"`
}

// Change — изменение в истории
type Change struct {
	TypeID   int64  `json:"type_id"`
	OldText  string `json:"old_text,omitempty"`
	NewText  string `json:"new_text,omitempty"`
	Field    string `json:"field,omitempty"`
	OldValue int64  `json:"old_value,omitempty"`
	NewValue int64  `json:"new_value,omitempty"`
}

// Request структуры //

// AddCaseRequest — запрос для add_case
type AddCaseRequest struct {
	Title                string `json:"title"`      // обязательно
	SectionID            int64  `json:"section_id"` // если нужно явно указывать
	TypeID               int64  `json:"type_id"`
	PriorityID           int64  `json:"priority_id"`
	Estimate             string `json:"estimate,omitempty"`
	CustomPreconds       string `json:"custom_preconds,omitempty"`
	CustomSteps          string `json:"custom_steps,omitempty"`           // Текстовый формат шагов (альтернатива CustomStepsSeparated)
	CustomExpected       string `json:"custom_expected,omitempty"`        // Ожидаемый результат (текстовый формат)
	CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"` // Структурированные шаги
	Refs                 string `json:"refs,omitempty"`
	MilestoneID          int64  `json:"milestone_id,omitempty"`
	TemplateID           int64  `json:"template_id,omitempty"`
}

// UpdateCaseRequest — запрос для update_case (частичные обновления)
// Используются указатели для различения "не задано" от "пустое значение"
type UpdateCaseRequest struct {
	Title                *string `json:"title,omitempty"`
	TypeID               *int64  `json:"type_id,omitempty"` // Для изменения типа кейса
	PriorityID           *int64  `json:"priority_id,omitempty"`
	Estimate             *string `json:"estimate,omitempty"`
	CustomPreconds       *string `json:"custom_preconds,omitempty"`
	CustomSteps          *string `json:"custom_steps,omitempty"`           // Текстовый формат шагов
	CustomExpected       *string `json:"custom_expected,omitempty"`        // Ожидаемый результат
	CustomStepsSeparated []Step  `json:"custom_steps_separated,omitempty"` // Step из shared.go
	Refs                 *string `json:"refs,omitempty"`
	MilestoneID          *int64  `json:"milestone_id,omitempty"`
	SuiteID              *int64  `json:"suite_id,omitempty"`    // Для перемещения между сьютами
	SectionID            *int64  `json:"section_id,omitempty"`  // Для перемещения между секциями
	TemplateID           *int64  `json:"template_id,omitempty"` // Для изменения шаблона
}

// CopyCasesRequest — запрос для copy_cases_to_section
type CopyCasesRequest struct {
	CaseIDs []int64 `json:"case_ids"` // IDs of cases to copy
}

// MoveCasesRequest — запрос для move_cases_to_section
type MoveCasesRequest struct {
	CaseIDs []int64 `json:"case_ids"`           // IDs of cases to move
	SuiteID int64   `json:"suite_id,omitempty"` // Target suite ID (for cross-suite moves)
}

// UpdateCasesRequest — запрос для bulk update_cases
type UpdateCasesRequest struct {
	CaseIDs    []int64 `json:"case_ids"`
	PriorityID int64   `json:"priority_id,omitempty"`
	Estimate   string  `json:"estimate,omitempty"`
}

// DeleteCasesRequest — запрос для delete_cases
type DeleteCasesRequest struct {
	CaseIDs []int64 `json:"case_ids"`
}

// GetCaseTypesResponse — ответ для get_case_types
type GetCaseTypesResponse []struct {
	ID        int64  `json:"id"`
	IsDefault bool   `json:"is_default"`
	Name      string `json:"name"`
}

// GetCaseFieldsResponse — ответ для get_case_fields
type GetCaseFieldsResponse []struct {
	Configs []struct {
		Context struct {
			IsGlobal   bool    `json:"is_global"`
			ProjectIDs []int64 `json:"project_ids,omitempty"`
		} `json:"context"`
		ID      string `json:"id"`
		Options struct {
			DefaultValue string `json:"default_value,omitempty"`
			Format       string `json:"format,omitempty"`
			IsRequired   bool   `json:"is_required"`
			Rows         string `json:"rows,omitempty"`
		} `json:"options"`
	} `json:"configs"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
	ID           int64  `json:"id"`
	Label        string `json:"label"`
	Name         string `json:"name"`
	SystemName   string `json:"system_name"`
	TypeID       int64  `json:"type_id"`
}

// AddCaseFieldRequest — запрос для add_case_field
type AddCaseFieldRequest struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Configs     []struct {
		Context struct {
			IsGlobal   bool    `json:"is_global"`
			ProjectIDs []int64 `json:"project_ids,omitempty"`
		} `json:"context"`
		Options struct {
			IsRequired bool   `json:"is_required"`
			Items      string `json:"items,omitempty"`
		} `json:"options"`
	} `json:"configs"`
	IncludeAll bool `json:"include_all"`
}

// AddCaseFieldResponse — ответ для add_case_field
type AddCaseFieldResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SystemName   string `json:"system_name"`
	EntityID     int64  `json:"entity_id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	TypeID       int64  `json:"type_id"`
	LocationID   int64  `json:"location_id"`
	DisplayOrder int    `json:"display_order"`
	Configs      string `json:"configs"`
	IsMulti      int    `json:"is_multi"`
	IsActive     int    `json:"is_active"`
	StatusID     int    `json:"status_id"`
	IsSystem     int    `json:"is_system"`
	IncludeAll   int    `json:"include_all"`
	TemplateIDs  []any  `json:"template_ids"`
}

// DiffCasesResponse — результат сравнения кейсов двух проектов
type DiffCasesResponse struct {
	OnlyInFirst  []Case `json:"only_in_first"`  // Есть только в pid1
	OnlyInSecond []Case `json:"only_in_second"` // Есть только в pid2
	DiffByField  []struct {
		CaseID int64 `json:"case_id"`
		First  Case  `json:"first"`
		Second Case  `json:"second"`
	} `json:"diff_by_field"` // Отличаются по полю
}
