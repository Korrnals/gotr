// Package interactive provides interactive mode (wizard) for add and update commands.
package interactive

import (
	"fmt"
)

// ProjectAnswers содержит ответы для создания/обновления проекта
type ProjectAnswers struct {
	Name             string
	Announcement     string
	ShowAnnouncement bool
	IsCompleted      bool
}

// SuiteAnswers содержит ответы для создания/обновления сьюта
type SuiteAnswers struct {
	Name        string
	Description string
	IsCompleted bool
}

// CaseAnswers содержит ответы для создания/обновления case
type CaseAnswers struct {
	Title      string
	SectionID  int64
	TypeID     int64
	PriorityID int64
	Refs       string
}

// RunAnswers содержит ответы для создания/обновления run
type RunAnswers struct {
	Name        string
	Description string
	SuiteID     int64
	MilestoneID int64
	AssignedTo  int64
	IncludeAll  bool
}

// AskProjectWithPrompter runs project wizard using unified prompter.
func AskProjectWithPrompter(p Prompter, isUpdate bool) (*ProjectAnswers, error) {
	answers := &ProjectAnswers{}

	name, err := p.Input("Project name:", "")
	if err != nil {
		return nil, fmt.Errorf("project name input failed: %w", err)
	}
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	announcement, err := p.MultilineInput("Announcement (optional):", "")
	if err != nil {
		return nil, fmt.Errorf("announcement input failed: %w", err)
	}

	showAnnouncement, err := p.Confirm("Show announcement?", true)
	if err != nil {
		return nil, fmt.Errorf("show announcement confirm failed: %w", err)
	}

	answers.Name = name
	answers.Announcement = announcement
	answers.ShowAnnouncement = showAnnouncement

	if isUpdate {
		isCompleted, err := p.Confirm("Mark as completed?", false)
		if err != nil {
			return nil, fmt.Errorf("is completed confirm failed: %w", err)
		}
		answers.IsCompleted = isCompleted
	}

	return answers, nil
}

// AskProject запускает wizard for project (compatibility wrapper).
func AskProject(isUpdate bool) (*ProjectAnswers, error) {
	answers, err := AskProjectWithPrompter(NewTerminalPrompter(), isUpdate)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

// AskSuiteWithPrompter runs suite wizard using unified prompter.
func AskSuiteWithPrompter(p Prompter, isUpdate bool) (*SuiteAnswers, error) {
	answers := &SuiteAnswers{}

	name, err := p.Input("Suite name:", "")
	if err != nil {
		return nil, fmt.Errorf("suite name input failed: %w", err)
	}
	if name == "" {
		return nil, fmt.Errorf("suite name is required")
	}

	description, err := p.MultilineInput("Description (optional):", "")
	if err != nil {
		return nil, fmt.Errorf("suite description input failed: %w", err)
	}

	answers.Name = name
	answers.Description = description

	if isUpdate {
		isCompleted, err := p.Confirm("Mark as completed?", false)
		if err != nil {
			return nil, fmt.Errorf("suite completion confirm failed: %w", err)
		}
		answers.IsCompleted = isCompleted
	}

	return answers, nil
}

// AskSuite запускает wizard для сьюта (compatibility wrapper).
func AskSuite(isUpdate bool) (*SuiteAnswers, error) {
	answers, err := AskSuiteWithPrompter(NewTerminalPrompter(), isUpdate)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

// AskCaseWithPrompter runs case wizard using unified prompter.
func AskCaseWithPrompter(p Prompter, isUpdate bool) (*CaseAnswers, error) {
	answers := &CaseAnswers{}

	title, err := p.Input("Title:", "")
	if err != nil {
		return nil, fmt.Errorf("case title input failed: %w", err)
	}
	if title == "" {
		return nil, fmt.Errorf("case title is required")
	}
	answers.Title = title

	// Тип case
	typeIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	typeOptions := []string{
		"1 - Acceptance",
		"2 - Accessibility",
		"3 - Automated",
		"4 - Compatibility",
		"5 - Destructive",
		"6 - Functional",
		"7 - Other",
		"8 - Performance",
		"9 - Regression",
		"10 - Security",
		"11 - Smoke & Sanity",
		"12 - Usability",
	}
	typeIdx, _, err := p.Select("Case type:", typeOptions)
	if err != nil {
		return nil, fmt.Errorf("case type selection failed: %w", err)
	}
	answers.TypeID = typeIDs[typeIdx]

	// Приоритет
	priorityIDs := []int64{1, 2, 3, 4}
	priorityOptions := []string{
		"1 - Critical",
		"2 - High",
		"3 - Medium",
		"4 - Low",
	}
	priorityIdx, _, err := p.Select("Priority:", priorityOptions)
	if err != nil {
		return nil, fmt.Errorf("priority selection failed: %w", err)
	}
	answers.PriorityID = priorityIDs[priorityIdx]

	// Refs
	refs, err := p.Input("References (comma-separated):", "")
	if err != nil {
		return nil, fmt.Errorf("references input failed: %w", err)
	}
	answers.Refs = refs

	return answers, nil
}

// AskCase запускает wizard для тест-case (compatibility wrapper).
func AskCase(isUpdate bool) (*CaseAnswers, error) {
	answers, err := AskCaseWithPrompter(NewTerminalPrompter(), isUpdate)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

// AskRunWithPrompter runs run wizard using unified prompter.
func AskRunWithPrompter(p Prompter, isUpdate bool) (*RunAnswers, error) {
	answers := &RunAnswers{}

	name, err := p.Input("Test run name:", "")
	if err != nil {
		return nil, fmt.Errorf("run name input failed: %w", err)
	}
	if name == "" {
		return nil, fmt.Errorf("run name is required")
	}
	answers.Name = name

	description, err := p.MultilineInput("Description (optional):", "")
	if err != nil {
		return nil, fmt.Errorf("run description input failed: %w", err)
	}
	answers.Description = description

	if !isUpdate {
		suiteIDRaw, err := p.Input("Suite ID:", "")
		if err != nil {
			return nil, fmt.Errorf("suite id input failed: %w", err)
		}
		if suiteIDRaw == "" {
			return nil, fmt.Errorf("suite id is required")
		}
		_, err = fmt.Sscanf(suiteIDRaw, "%d", &answers.SuiteID)
		if err != nil {
			return nil, fmt.Errorf("invalid suite id: %w", err)
		}
	}

	includeAll, err := p.Confirm("Include all cases from suite?", true)
	if err != nil {
		return nil, fmt.Errorf("include all confirm failed: %w", err)
	}
	answers.IncludeAll = includeAll

	return answers, nil
}

// AskRun запускает wizard для test run (compatibility wrapper).
func AskRun(isUpdate bool) (*RunAnswers, error) {
	answers, err := AskRunWithPrompter(NewTerminalPrompter(), isUpdate)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

// AskConfirmWithPrompter asks confirmation via unified prompter.
func AskConfirmWithPrompter(p Prompter, message string) (bool, error) {
	confirm, err := p.Confirm(message, true)
	if err != nil {
		return false, fmt.Errorf("confirmation failed: %w", err)
	}
	return confirm, nil
}

// AskConfirm запрашивает подтверждение перед выполнением действия (compatibility wrapper).
func AskConfirm(message string) (bool, error) {
	return AskConfirmWithPrompter(NewTerminalPrompter(), message)
}
