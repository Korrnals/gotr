// Package interactive provides interactive mode (wizard) for add and update commands.
package interactive

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
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

// CaseAnswers содержит ответы для создания/обновления кейса
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

// AskProject запускает wizard для проекта
func AskProject(isUpdate bool) (*ProjectAnswers, error) {
	answers := &ProjectAnswers{}

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Название проекта:",
			},
			Validate: survey.Required,
		},
		{
			Name: "announcement",
			Prompt: &survey.Multiline{
				Message: "Announcement (опционально):",
			},
		},
		{
			Name: "showAnnouncement",
			Prompt: &survey.Confirm{
				Message: "Показывать announcement?",
				Default: true,
			},
		},
	}

	if isUpdate {
		questions = append(questions, &survey.Question{
			Name: "isCompleted",
			Prompt: &survey.Confirm{
				Message: "Отметить как завершённый?",
				Default: false,
			},
		})
	}

	if err := survey.Ask(questions, answers); err != nil {
		return nil, err
	}

	return answers, nil
}

// AskSuite запускает wizard для сьюта
func AskSuite(isUpdate bool) (*SuiteAnswers, error) {
	answers := &SuiteAnswers{}

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Название сьюта:",
			},
			Validate: survey.Required,
		},
		{
			Name: "description",
			Prompt: &survey.Multiline{
				Message: "Описание (опционально):",
			},
		},
	}

	if isUpdate {
		questions = append(questions, &survey.Question{
			Name: "isCompleted",
			Prompt: &survey.Confirm{
				Message: "Отметить как завершённый?",
				Default: false,
			},
		})
	}

	if err := survey.Ask(questions, answers); err != nil {
		return nil, err
	}

	return answers, nil
}

// AskCase запускает wizard для тест-кейса
func AskCase(isUpdate bool) (*CaseAnswers, error) {
	answers := &CaseAnswers{}

	questions := []*survey.Question{
		{
			Name: "title",
			Prompt: &survey.Input{
				Message: "Заголовок (title):",
			},
			Validate: survey.Required,
		},
	}

	if !isUpdate {
		questions = append(questions, &survey.Question{
			Name: "sectionId",
			Prompt: &survey.Input{
				Message: "ID секции:",
				Default: "0",
			},
			Validate: survey.Required,
		})
	}

	// Тип кейса
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
	
	var typeStr string
	if err := survey.AskOne(&survey.Select{
		Message: "Тип кейса:",
		Options: typeOptions,
		Default: typeOptions[0],
	}, &typeStr); err != nil {
		return nil, err
	}
	// Парсим ID из строки "1 - Acceptance"
	fmt.Sscanf(typeStr, "%d", &answers.TypeID)

	// Приоритет
	priorityOptions := []string{
		"1 - Critical",
		"2 - High", 
		"3 - Medium",
		"4 - Low",
	}
	
	var priorityStr string
	if err := survey.AskOne(&survey.Select{
		Message: "Приоритет:",
		Options: priorityOptions,
		Default: priorityOptions[2],
	}, &priorityStr); err != nil {
		return nil, err
	}
	fmt.Sscanf(priorityStr, "%d", &answers.PriorityID)

	// Refs
	if err := survey.AskOne(&survey.Input{
		Message: "Ссылки/референсы (через запятую):",
	}, &answers.Refs); err != nil {
		return nil, err
	}

	return answers, nil
}

// AskRun запускает wizard для test run
func AskRun(isUpdate bool) (*RunAnswers, error) {
	answers := &RunAnswers{}

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Название test run:",
			},
			Validate: survey.Required,
		},
		{
			Name: "description",
			Prompt: &survey.Multiline{
				Message: "Описание (опционально):",
			},
		},
	}

	if !isUpdate {
		questions = append(questions, &survey.Question{
			Name: "suiteId",
			Prompt: &survey.Input{
				Message: "ID сьюта:",
			},
			Validate: survey.Required,
		})
	}

	if err := survey.Ask(questions, answers); err != nil {
		return nil, err
	}

	// Include all cases
	if err := survey.AskOne(&survey.Confirm{
		Message: "Включить все кейсы сьюта?",
		Default: true,
	}, &answers.IncludeAll); err != nil {
		return nil, err
	}

	return answers, nil
}

// AskConfirm запрашивает подтверждение перед выполнением действия
func AskConfirm(message string) (bool, error) {
	var confirm bool
	err := survey.AskOne(&survey.Confirm{
		Message: message,
		Default: true,
	}, &confirm)
	return confirm, err
}
