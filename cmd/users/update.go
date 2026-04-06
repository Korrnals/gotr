// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package users

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates the 'users update' command.
// Endpoint: POST /update_user/{user_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [user_id]",
		Short: "Обновить пользователя",
		Long: `Обновляет существующего пользователя в системе TestRail.

Требуются административные права для изменения пользователей.`,
		Example: `  # Обновить имя пользователя
  gotr users update 123 --name "New Name"

  # Сделать пользователя администратором
  gotr users update 123 --admin

  # Заблокировать пользователя
  gotr users update 123 --inactive`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var userID int64
			var err error
			if len(args) > 0 {
				userID, err = flags.ValidateRequiredID(args, 0, "user_id")
				if err != nil {
					return err
				}
			} else {
				if err := requireInteractiveUserArg(cmd.Context(), "gotr users update [user_id]"); err != nil {
					return err
				}
				userID, err = resolveUserIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			req := data.UpdateUserRequest{}
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = name
			}
			if cmd.Flags().Changed("email") {
				email, _ := cmd.Flags().GetString("email")
				req.Email = email
			}
			if cmd.Flags().Changed("role") {
				roleID, _ := cmd.Flags().GetInt64("role")
				req.RoleID = roleID
			}
			if cmd.Flags().Changed("admin") {
				isAdmin, _ := cmd.Flags().GetBool("admin")
				if isAdmin {
					req.IsAdmin = 1
				} else {
					req.IsAdmin = 0
				}
			}
			if cmd.Flags().Changed("inactive") {
				isActive, _ := cmd.Flags().GetBool("inactive")
				if isActive {
					req.IsActive = 0 // inactive = true means is_active = 0
				} else {
					req.IsActive = 1 // inactive = false means is_active = 1
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("users update")
				dr.PrintOperation(
					fmt.Sprintf("Update User %d", userID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/update_user/%d", userID),
					req,
				)
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()

			user, err := cli.UpdateUser(ctx, userID, req)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}

			_, err = output.Output(cmd, user, "users", "json")
			return err
		},
	}

	cmd.Flags().String("name", "", "Имя пользователя")
	cmd.Flags().String("email", "", "Email пользователя")
	cmd.Flags().Int64("role", 0, "ID роли пользователя")
	cmd.Flags().Bool("admin", false, "Сделать пользователя администратором")
	cmd.Flags().Bool("inactive", false, "Заблокировать пользователя")
	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без обновления пользователя")
	output.AddFlag(cmd)

	return cmd
}
