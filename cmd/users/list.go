// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package users

import (
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'users list'
// Эндпоинты: GET /get_users, GET /get_users/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Получить список пользователей",
		Long: `Получить список пользователей.

Если project_id не указан — возвращает всех пользователей системы.
Если project_id указан — возвращает только пользователей проекта.`,
		Example: `  # Список всех пользователей
  gotr users list

  # Список пользователей проекта
  gotr users list 123

  # Вывод в JSON
  gotr users list -s output.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)

			if len(args) == 0 {
				return listAllUsers(cmd, cli)
			}

			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			return listProjectUsers(cmd, cli, projectID)
		},
	}

	output.AddFlag(cmd)

	return cmd
}

type usersClient interface {
	GetUsers() (data.GetUsersResponse, error)
	GetUsersByProject(projectID int64) (data.GetUsersResponse, error)
}

func listAllUsers(cmd *cobra.Command, cli usersClient) error {
	pm := progress.NewManager()
	progress.Describe(pm.NewSpinner(""), "Загрузка пользователей...")

	users, err := cli.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	saveFlag, _ := cmd.Flags().GetBool("save")
	if saveFlag {
		_, err := output.Output(cmd, users, "users", "json")
		return err
	}

	if len(users) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No users found")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tIS_ADMIN\tROLE_ID")
	for _, u := range users {
		fmt.Fprintf(w, "%d\t%s\t%s\t%v\t%d\n", u.ID, u.Name, u.Email, u.IsAdmin, u.RoleID)
	}
	return w.Flush()
}

func listProjectUsers(cmd *cobra.Command, cli usersClient, projectID int64) error {
	pm := progress.NewManager()
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка пользователей проекта %d...", projectID))

	users, err := cli.GetUsersByProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to list project users: %w", err)
	}

	saveFlag, _ := cmd.Flags().GetBool("save")
	if saveFlag {
		_, err := output.Output(cmd, users, "users", "json")
		return err
	}

	if len(users) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No users found for project %d\n", projectID)
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tIS_ADMIN\tROLE_ID")
	for _, u := range users {
		fmt.Fprintf(w, "%d\t%s\t%s\t%v\t%d\n", u.ID, u.Name, u.Email, u.IsAdmin, u.RoleID)
	}
	return w.Flush()
}

// Verify interface compliance
var _ usersClient = (client.ClientInterface)(nil)
