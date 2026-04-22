package casefields

import (
	"strconv"

	"github.com/Korrnals/gotr/internal/models/data"
)

// ResolveRequiredCustomFields returns a map of system_name → parsed default value
// for all case fields that are required for the given project AND have a non-empty
// default value configured. Fields listed in exclude are skipped (they are handled
// explicitly elsewhere, e.g. custom_autotest_on).
//
// The returned map is intended for use as AddCaseRequest.ExtraFields so that
// project-specific required fields are always present in create requests.
func ResolveRequiredCustomFields(
	fields data.GetCaseFieldsResponse,
	projectID int64,
	exclude []string,
) map[string]interface{} {
	excludeSet := make(map[string]struct{}, len(exclude))
	for _, name := range exclude {
		excludeSet[name] = struct{}{}
	}

	result := make(map[string]interface{})

	for _, field := range fields {
		if _, skip := excludeSet[field.SystemName]; skip {
			continue
		}

		configs := matchingConfigs(field.Configs, projectID)
		if len(configs) == 0 {
			continue
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

		if !required || defaultValue == "" {
			continue
		}

		// Try to parse as int64; fall back to string.
		if n, err := strconv.ParseInt(defaultValue, 10, 64); err == nil {
			result[field.SystemName] = n
		} else {
			result[field.SystemName] = defaultValue
		}
	}

	return result
}
