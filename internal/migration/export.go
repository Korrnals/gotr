// internal/migration/export.go
package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gotr/internal/models/data"
)

// ExportSharedSteps — экспортирует shared steps (filtered или все source)
func (m *Migration) ExportSharedSteps(steps data.GetSharedStepsResponse, filtered bool, dir string) error {
	if len(steps) == 0 {
		m.logger.Info("Нет shared steps для экспорта")
		return nil
	}

	if dir == "" {
		dir = ".testrail"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileType := "all"
	if filtered {
		fileType = "filtered"
	}
	file := filepath.Join(dir, fmt.Sprintf("shared_steps_%s_%s.json", fileType, time.Now().Format("2006-01-02_15-04-05")))

	jsonData, err := json.MarshalIndent(steps, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(file, jsonData, 0644); err != nil {
		return err
	}

	m.logger.Info("Shared steps экспортированы", "file", file, "count", len(steps), "type", fileType)
	return nil
}

// ExportSuites — экспортирует suites
func (m *Migration) ExportSuites(suites data.GetSuitesResponse, filtered bool, dir string) error {
	if len(suites) == 0 {
		m.logger.Info("Нет suites для экспорта")
		return nil
	}

	if dir == "" {
		dir = ".testrail"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileType := "all"
	if filtered {
		fileType = "filtered"
	}
	file := filepath.Join(dir, fmt.Sprintf("suites_%s_%s.json", fileType, time.Now().Format("2006-01-02_15-04-05")))

	jsonData, err := json.MarshalIndent(suites, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(file, jsonData, 0644); err != nil {
		return err
	}

	m.logger.Info("Suites экспортированы", "file", file, "count", len(suites), "type", fileType)
	return nil
}

// ExportCases — экспортирует cases
func (m *Migration) ExportCases(cases data.GetCasesResponse, filtered bool, dir string) error {
	if len(cases) == 0 {
		m.logger.Info("Нет cases для экспорта")
		return nil
	}

	if dir == "" {
		dir = ".testrail"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileType := "all"
	if filtered {
		fileType = "filtered"
	}
	file := filepath.Join(dir, fmt.Sprintf("cases_%s_%s.json", fileType, time.Now().Format("2006-01-02_15-04-05")))

	jsonData, err := json.MarshalIndent(cases, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(file, jsonData, 0644); err != nil {
		return err
	}

	m.logger.Info("Cases экспортированы", "file", file, "count", len(cases), "type", fileType)
	return nil
}

// ExportSections — экспортирует sections
func (m *Migration) ExportSections(sections data.GetSectionsResponse, filtered bool, dir string) error {
	if len(sections) == 0 {
		m.logger.Info("Нет sections для экспорта")
		return nil
	}

	if dir == "" {
		dir = ".testrail"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileType := "all"
	if filtered {
		fileType = "filtered"
	}
	file := filepath.Join(dir, fmt.Sprintf("sections_%s_%s.json", fileType, time.Now().Format("2006-01-02_15-04-05")))

	jsonData, err := json.MarshalIndent(sections, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(file, jsonData, 0644); err != nil {
		return err
	}

	m.logger.Info("Sections экспортированы", "file", file, "count", len(sections), "type", fileType)
	return nil
}

// ExportMapping — использует метод Save из mapping.go
func (m *Migration) ExportMapping(dir string) error {
	return m.mapping.Save(dir)
}
