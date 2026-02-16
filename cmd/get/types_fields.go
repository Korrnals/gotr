package get

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newCaseTypesCmd создаёт команду для получения типов кейсов
func newCaseTypesCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "case-types",
		Short: "Получить список типов кейсов",
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			types, err := cli.GetCaseTypes()
			if err != nil {
				return err
			}

			return handleOutput(command, types, start)
		},
	}
}

// newCaseFieldsCmd создаёт команду для получения полей кейсов
func newCaseFieldsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "case-fields",
		Short: "Получить список полей кейсов",
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			fields, err := cli.GetCaseFields()
			if err != nil {
				return err
			}

			return handleOutput(command, fields, start)
		},
	}
}

// caseTypesCmd — экспортированная команда для регистрации
var caseTypesCmd = newCaseTypesCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// caseFieldsCmd — экспортированная команда для регистрации
var caseFieldsCmd = newCaseFieldsCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
