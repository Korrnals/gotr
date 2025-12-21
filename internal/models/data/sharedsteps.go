// models/data/shared_steps.go
package data

import "encoding/json"

// SharedStep — основная структура для одного shared step (get_shared_step, add_shared_step)
type SharedStep struct {
	ID                   int64  `json:"id"`
	Title                string `json:"title,omitempty"`
	ProjectID            int64  `json:"project_id,omitempty"`
	CreatedBy            int64  `json:"created_by,omitempty"`
	CreatedOn            int64  `json:"created_on,omitempty"`
	UpdatedBy            int64  `json:"updated_by,omitempty"`
	UpdatedOn            int64  `json:"updated_on,omitempty"`
	CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"`
	CaseIDs              []int64 `json:"case_ids,omitempty"` // Кейсы, где используется этот step
	// Для неизвестных расширений
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// GetSharedStepsResponse — ответ для get_shared_steps (пагинированный список)
type GetSharedStepsResponse struct {
	Pagination
	SharedSteps []SharedStep `json:"shared_steps"`
}

// GetSharedStepResponse — ответ для get_shared_step (один step)
type GetSharedStepResponse struct {
	SharedStep
}

// GetSharedStepHistoryResponse — ответ для get_shared_step_history
type GetSharedStepHistoryResponse struct {
	Pagination
	StepHistory []struct {
		ID                   int64  `json:"id"`
		Timestamp            int64  `json:"timestamp"`
		UserID               int64  `json:"user_id"`
		CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"`
		Title                string `json:"title,omitempty"`
	} `json:"step_history"`
}

// Request структуры

// AddSharedStepRequest — запрос для add_shared_step
type AddSharedStepRequest struct {
	Title                string `json:"title"` // Обязательное
	CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"`
	// ProjectID не нужен — выбирается автоматически по контексту
}

// UpdateSharedStepRequest — запрос для update_shared_step
type UpdateSharedStepRequest struct {
	Title                string `json:"title,omitempty"`
	CustomStepsSeparated []Step `json:"custom_steps_separated,omitempty"`
	// Другие поля, если TestRail поддерживает
}

// DeleteSharedStepRequest — запрос для delete_shared_step
type DeleteSharedStepRequest struct {
	KeepInCases int `json:"keep_in_cases"` // 1 — сохранить в кейсах, 0 — удалить
}