// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package users

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/internal/output"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'users update'
// Эндпоинт: POST /update_user/{user_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	var (
		name     string
		email    string
		roleID   int64
		isAdmin  bool
		isActive bool
	)

	cmd := &cobra.Command{
		Use:   "update <user_id>",
		Short: "Обновить пользователя",
		Long: `Обновляет существующего пользователя в системе TestRail.

Требуются административные права для изменения пользователей.`,
		Example: `  # Обновить имя пользователя
  gotr users update 123 --name "New Name"

  # Сделать пользователя администратором
  gotr users update 123 --admin

  # Заблокировать пользователя
  gotr users update 123 --inactive`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || userID <= 0 {
				return fmt.Errorf("invalid user_id: %s", args[0])
			}

			cli := getClient(cmd)

			req := data.UpdateUserRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("email") {
				req.Email = email
			}
			if cmd.Flags().Changed("role") {
				req.RoleID = roleID
			}
			if cmd.Flags().Changed("admin") {
				if isAdmin {
					req.IsAdmin = 1
				} else {
					req.IsAdmin = 0
				}
			}
			if cmd.Flags().Changed("inactive") {
				if isActive {
					req.IsActive = 0  // inactive = true means is_active = 0
				} else {
					req.IsActive = 1  // inactive = false means is_active = 1
				}
			}

			user, err := cli.UpdateUser(userID, req)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}

			return output.Result(cmd, user)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Имя пользователя")
	cmd.Flags().StringVar(&email, "email", "", "Email пользователя")
	cmd.Flags().Int64Var(&roleID, "role", 0, "ID роли пользователя")
	cmd.Flags().BoolVar(&isAdmin, "admin", false, "Сделать пользователя администратором")
	cmd.Flags().BoolVar(&isActive, "inactive", false, "Заблокировать пользователя")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	return cmd
}
