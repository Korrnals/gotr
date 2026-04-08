package users

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetByEmailCmd creates the 'users get-by-email' command.
// Endpoint: GET /get_user_by_email
func newGetByEmailCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-by-email [email]",
		Short: "Get user by email",
		Long: `Retrieves user information by their email address.

Displays full information: ID, name, email, activity status,
role, role ID, MFA status, and administrator flag.

Useful for finding a user when the email is known but not the ID.`,
		Example: `  # Get user by email
  gotr users get-by-email user@example.com

  # Save result to a file
  gotr users get-by-email user@example.com -o user.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var email string
			if len(args) > 0 {
				email = args[0]
			} else {
				if err := requireInteractiveUserArg(cmd.Context(), "gotr users get-by-email [email]"); err != nil {
					return err
				}
				var err error
				email, err = resolveEmailInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}
			if email == "" {
				return fmt.Errorf("email cannot be empty")
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetUserByEmail(ctx, email)
			if err != nil {
				return fmt.Errorf("failed to get user by email: %w", err)
			}

			_, err = output.Output(cmd, resp, "users", "json")
			return err
		},
	}

	output.AddFlag(cmd)

	return cmd
}
