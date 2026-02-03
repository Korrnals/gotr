package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// projectsCmd — подкоманда для всех проектов
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Получить все проекты",
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		projects, err := client.GetProjects()
		if err != nil {
			return err
		}

		return handleOutput(command, projects, start)
	},
}

// projectCmd — подкоманда для одного проекта
var projectCmd = &cobra.Command{
	Use:   "project <project-id>",
	Short: "Получить один проект по ID проекта",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID проекта: %w", err)
		}

		project, err := client.GetProject(id)
		if err != nil {
			return err
		}

		return handleOutput(command, project, start)
	},
}
