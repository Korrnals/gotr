// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package users

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newAddCmd creates the 'users add' command.
// Endpoint: POST /add_user
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a new user",
		Long: `Creates a new user in the TestRail system.

Administrative privileges are required to create users.`,
		Example: `  # Create a regular user
  gotr users add --name "John Doe" --email "john@example.com"

  # Create an administrator
  gotr users add --name "Admin User" --email "admin@example.com" --admin --role 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			email, _ := cmd.Flags().GetString("email")
			roleID, _ := cmd.Flags().GetInt64("role")
			isAdmin, _ := cmd.Flags().GetBool("admin")
			password, _ := cmd.Flags().GetString("password")

			req := data.AddUserRequest{
				Name:     name,
				Email:    email,
				RoleID:   roleID,
				Password: password,
			}
			if isAdmin {
				req.IsAdmin = 1
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("users add")
				dr.PrintOperation(
					"Create User",
					"POST",
					"/index.php?/api/v2/add_user",
					req,
				)
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()

			user, err := cli.AddUser(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to add user: %w", err)
			}

			_, err = output.Output(cmd, user, "users", "json")
			return err
		},
	}

	cmd.Flags().String("name", "", "User name (required)")
	cmd.Flags().String("email", "", "User email (required)")
	cmd.Flags().Int64("role", 0, "User role ID")
	cmd.Flags().Bool("admin", false, "Make the user an administrator")
	cmd.Flags().String("password", "", "User password")
	cmd.Flags().Bool("dry-run", false, "Show what would be done without creating the user")
	output.AddFlag(cmd)

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}
