// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package users

import (
	"fmt"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'users add'
// Эндпоинт: POST /add_user
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	var (
		name     string
		email    string
		roleID   int64
		isAdmin  bool
		password string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Создать нового пользователя",
		Long: `Создает нового пользователя в системе TestRail.

Требуются административные права для создания пользователей.`,
		Example: `  # Создать обычного пользователя
  gotr users add --name "John Doe" --email "john@example.com"

  # Создать администратора
  gotr users add --name "Admin User" --email "admin@example.com" --admin --role 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)

			req := data.AddUserRequest{
				Name:     name,
				Email:    email,
				RoleID:   roleID,
				Password: password,
			}
			if isAdmin {
				req.IsAdmin = 1
			}

			user, err := cli.AddUser(req)
			if err != nil {
				return fmt.Errorf("failed to add user: %w", err)
			}

			_, err = save.Output(cmd, user, "users", "json")
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Имя пользователя (обязательно)")
	cmd.Flags().StringVar(&email, "email", "", "Email пользователя (обязательно)")
	cmd.Flags().Int64Var(&roleID, "role", 0, "ID роли пользователя")
	cmd.Flags().BoolVar(&isAdmin, "admin", false, "Сделать пользователя администратором")
	cmd.Flags().StringVar(&password, "password", "", "Пароль пользователя")
	save.AddFlag(cmd)

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}
