package get

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// newProjectsCmd создаёт команду для получения списка проектов
func newProjectsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "Получить все projects",
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			projects, err := cli.GetProjects(ctx)
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
		Use:   "project [project-id]",
		Short: "Получить один проект по ID проекта",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			var id int64
			var err error
			if len(args) == 0 {
				id, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			} else {
				id, err = flags.ParseID(args[0])
				if err != nil {
					return fmt.Errorf("invalid project_id: %w", err)
				}
				if id <= 0 {
					return fmt.Errorf("invalid project_id: %s", args[0])
				}
			}

			project, err := cli.GetProject(ctx, id)
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
