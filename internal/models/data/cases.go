package data

// Структура для `get_cases`
type GetCasesResponse struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Size   int `json:"size"`
	Links  struct {
		Next string `json:"next"`
		Prev any    `json:"prev"`
	} `json:"_links"`
	Cases []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"cases"`
}

// Структура для 'get_case'
type GetCaseResponse struct {
	ID                   int    `json:"id"`
	Title                string `json:"title"`
	SectionID            int    `json:"section_id"`
	TemplateID           int    `json:"template_id"`
	TypeID               int    `json:"type_id"`
	PriorityID           int    `json:"priority_id"`
	MilestoneID          any    `json:"milestone_id"`
	Refs                 any    `json:"refs"`
	CreatedBy            int    `json:"created_by"`
	CreatedOn            int    `json:"created_on"`
	UpdatedBy            int    `json:"updated_by"`
	UpdatedOn            int    `json:"updated_on"`
	Estimate             any    `json:"estimate"`
	EstimateForecast     string `json:"estimate_forecast"`
	SuiteID              int    `json:"suite_id"`
	DisplayOrder         int    `json:"display_order"`
	IsDeleted            int    `json:"is_deleted"`
	CustomAutomationType int    `json:"custom_automation_type"`
	CustomPreconds       any    `json:"custom_preconds"`
	CustomSteps          any    `json:"custom_steps"`
	CustomExpected       any    `json:"custom_expected"`
	CustomStepsSeparated any    `json:"custom_steps_separated"`
	CustomMission        any    `json:"custom_mission"`
	CustomGoals          any    `json:"custom_goals"`
	Labels               []struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		CreatedBy string `json:"created_by"`
		CreatedOn string `json:"created_on"`
	} `json:"labels"`
}

// Структура для 'get_history_for_case'
type GetHystoryForCaseResponse []struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Size   int `json:"size"`
	Links  struct {
		Next any `json:"next"`
		Prev any `json:"prev"`
	} `json:"_links"`
	History []struct {
		ID        int `json:"id"`
		TypeID    int `json:"type_id"`
		CreatedOn int `json:"created_on"`
		UserID    int `json:"user_id"`
		Changes   []struct {
			TypeID   int    `json:"type_id"`
			OldText  string `json:"old_text"`
			NewText  string `json:"new_text"`
			Field    string `json:"field"`
			OldValue int    `json:"old_value"`
			NewValue int    `json:"new_value"`
		} `json:"changes"`
	} `json:"history"`
}

// Структура для 'add_case'
type AddCaseRequest struct {
	CustomPreconds string `json:"custom_preconds"`
}

// Структура для 'update_case'
type UpdateCaseRequest struct {
	PriorityID int    `json:"priority_id"`
	Estimate   string `json:"estimate"`
}

// Структура для 'update_cases'
type UpdateCasesRequest struct {
	CaseIds    []int  `json:"case_ids"`
	PriorityID int    `json:"priority_id"`
	Estimate   string `json:"estimate"`
}

// Структура для 'update_cases'
type DeleteCasesRequest struct {
	CaseIds []int `json:"case_ids"`
}

// Структура для 'get_case_types'
type GetCaseTypesResponse []struct {
	ID        int    `json:"id"`
	IsDefault bool   `json:"is_default"`
	Name      string `json:"name"`
}

// Структура для 'get_case_fields'
type GetCaseFieldsResponse []struct {
	Configs []struct {
		Context struct {
			IsGlobal   bool `json:"is_global"`
			ProjectIds any  `json:"project_ids"`
		} `json:"context"`
		ID      string `json:"id"`
		Options struct {
			DefaultValue string `json:"default_value"`
			Format       string `json:"format"`
			IsRequired   bool   `json:"is_required"`
			Rows         string `json:"rows"`
		} `json:"options"`
	} `json:"configs"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
	ID           int    `json:"id"`
	Label        string `json:"label"`
	Name         string `json:"name"`
	SystemName   string `json:"system_name"`
	TypeID       int    `json:"type_id"`
}

// Структура для 'add_case_field' {request}
type AddCaseFieldRequest struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Configs     []struct {
		Context struct {
			IsGlobal   bool   `json:"is_global"`
			ProjectIds string `json:"project_ids"`
		} `json:"context"`
		Options struct {
			IsRequired bool   `json:"is_required"`
			Items      string `json:"items"`
		} `json:"options"`
	} `json:"configs"`
	IncludeAll bool `json:"include_all"`
}

// Структура для 'add_case_field' {response}
type AddCaseFieldResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	SystemName   string `json:"system_name"`
	EntityID     int    `json:"entity_id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	TypeID       int    `json:"type_id"`
	LocationID   int    `json:"location_id"`
	DisplayOrder int    `json:"display_order"`
	Configs      string `json:"configs"`
	IsMulti      int    `json:"is_multi"`
	IsActive     int    `json:"is_active"`
	StatusID     int    `json:"status_id"`
	IsSystem     int    `json:"is_system"`
	IncludeAll   int    `json:"include_all"`
	TemplateIds  []any  `json:"template_ids"`
}