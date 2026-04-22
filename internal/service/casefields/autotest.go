package casefields

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/models/data"
)

const customAutotestOnSystemName = "custom_autotest_on"

// ResolveCustomAutotestOn returns the configured default value for custom_autotest_on
// when the field is required for the specified project. If the field is optional or
// not configured for the project, nil is returned.
func ResolveCustomAutotestOn(fields data.GetCaseFieldsResponse, projectID int64) (*int64, error) {
	field := findCustomAutotestOnField(fields)
	if field == nil {
		return nil, nil
	}

	configs := matchingConfigs(field.Configs, projectID)
	if len(configs) == 0 {
		return nil, nil
	}

	required := false
	defaultValue := ""
	for _, cfg := range configs {
		if cfg.Options.IsRequired {
			required = true
		}
		if defaultValue == "" && cfg.Options.DefaultValue != "" {
			defaultValue = cfg.Options.DefaultValue
		}
	}

	if defaultValue != "" {
		parsed, err := strconv.ParseInt(defaultValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid default_value for %s: %w", customAutotestOnSystemName, err)
		}
		return &parsed, nil
	}

	if required {
		return nil, fmt.Errorf("case field %s is required for project %d but has no default value", customAutotestOnSystemName, projectID)
	}

	return nil, nil
}

func findCustomAutotestOnField(fields data.GetCaseFieldsResponse) *struct {
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
} {
	for i := range fields {
		if fields[i].SystemName == customAutotestOnSystemName || fields[i].Name == customAutotestOnSystemName {
			return &fields[i]
		}
	}
	return nil
}

func matchingConfigs(configs []struct {
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
}, projectID int64) []struct {
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
} {
	projectSpecific := make([]struct {
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
	}, 0)
	global := make([]struct {
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
	}, 0)

	for _, cfg := range configs {
		if cfg.Context.IsGlobal {
			global = append(global, cfg)
			continue
		}
		for _, cfgProjectID := range cfg.Context.ProjectIDs {
			if cfgProjectID == projectID {
				projectSpecific = append(projectSpecific, cfg)
				break
			}
		}
	}

	if len(projectSpecific) > 0 {
		return projectSpecific
	}
	return global
}