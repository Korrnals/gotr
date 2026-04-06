package get

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newCaseTypesCmd creates the command for retrieving case types.
func newCaseTypesCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "case-types",
		Short: "Получить список типов кейсов",
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			types, err := cli.GetCaseTypes(ctx)
			if err != nil {
				return err
			}

			return handleOutput(command, types, start)
		},
	}
}

// newCaseFieldsCmd creates the command for retrieving case fields.
func newCaseFieldsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "case-fields",
		Short: "Получить список полей кейсов",
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			fields, err := cli.GetCaseFields(ctx)
			if err != nil {
				return err
			}

			return handleOutput(command, fields, start)
		},
	}
}

// caseTypesCmd is the exported command registered with the root.
var caseTypesCmd = newCaseTypesCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// caseFieldsCmd is the exported command registered with the root.
var caseFieldsCmd = newCaseFieldsCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
