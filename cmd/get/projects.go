package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newProjectsCmd создаёт команду для получения списка проектов
func newProjectsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "Получить все проекты",
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			projects, err := cli.GetProjects()
			if err != nil {
				return err
			}

			return handleOutput(command, projects, start)
		},
	}
}

// newProjectCmd создаёт команду для получения одного проекта
func newProjectCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "project <project-id>",
		Short: "Получить один проект по ID проекта",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			idStr := args[0]
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID проекта: %w", err)
			}

			project, err := cli.GetProject(id)
			if err != nil {
				return err
			}

			return handleOutput(command, project, start)
		},
	}
}

// projectsCmd — экспортированная команда для регистрации
var projectsCmd = newProjectsCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// projectCmd — экспортированная команда для регистрации
var projectCmd = newProjectCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
