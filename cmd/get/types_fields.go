package get

import (
	"time"

	"github.com/spf13/cobra"
)

// caseTypesCmd — подкоманда для типов кейсов
var caseTypesCmd = &cobra.Command{
	Use:   "case-types",
	Short: "Получить список типов кейсов",
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		types, err := client.GetCaseTypes()
		if err != nil {
			return err
		}

		return handleOutput(command, types, start)
	},
}

// caseFieldsCmd — подкоманда для полей кейсов
var caseFieldsCmd = &cobra.Command{
	Use:   "case-fields",
	Short: "Получить список полей кейсов",
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		fields, err := client.GetCaseFields()
		if err != nil {
			return err
		}

		return handleOutput(command, fields, start)
	},
}
