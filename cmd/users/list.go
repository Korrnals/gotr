// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package users

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

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
			ctx := cmd.Context()

			if len(args) == 0 {
				return listAllUsers(ctx, cmd, cli)
			}

			projectID, err := flags.ValidateRequiredID(args, 0, "project_id")
			if err != nil {
				return err
			}

			return listProjectUsers(ctx, cmd, cli, projectID)
		},
	}

	output.AddFlag(cmd)

	return cmd
}

type usersClient interface {
	GetUsers(ctx context.Context) (data.GetUsersResponse, error)
	GetUsersByProject(ctx context.Context, projectID int64) (data.GetUsersResponse, error)
}

func listAllUsers(ctx context.Context, cmd *cobra.Command, cli usersClient) error {
	pm := progress.NewManager()
	progress.Describe(pm.NewSpinner(""), "Загрузка пользователей...")

	users, err := cli.GetUsers(ctx)
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

	t := ui.NewTable(cmd)
	t.AppendHeader(table.Row{"ID", "NAME", "EMAIL", "IS_ADMIN", "ROLE_ID"})
	for _, u := range users {
		t.AppendRow(table.Row{u.ID, u.Name, u.Email, u.IsAdmin, u.RoleID})
	}
	ui.Table(cmd, t)
	return nil
}

func listProjectUsers(ctx context.Context, cmd *cobra.Command, cli usersClient, projectID int64) error {
	pm := progress.NewManager()
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка пользователей проекта %d...", projectID))

	users, err := cli.GetUsersByProject(ctx, projectID)
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

	t := ui.NewTable(cmd)
	t.AppendHeader(table.Row{"ID", "NAME", "EMAIL", "IS_ADMIN", "ROLE_ID"})
	for _, u := range users {
		t.AppendRow(table.Row{u.ID, u.Name, u.Email, u.IsAdmin, u.RoleID})
	}
	ui.Table(cmd, t)
	return nil
}

// Verify interface compliance
var _ usersClient = (client.ClientInterface)(nil)
