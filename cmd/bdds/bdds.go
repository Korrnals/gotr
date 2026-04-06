// Package bdds implements CLI commands for managing TestRail BDD scenarios.
package bdds

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all BDD-related commands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	bddsCmd := &cobra.Command{
		Use:   "bdds",
		Short: "Управление BDD сценариями",
		Long: `Управление BDD (Behavior Driven Development) сценариями.

BDD сценарии описывают поведение системы в формате Given-When-Then
(Дано-Когда-Тогда) на языке Gherkin. Привязаны к тест-кейсам
и позволяют писать тесты на понятном бизнесу языке.

Доступные операции:
  • get — получить BDD для тест-кейса
  • add — добавить BDD к тест-кейсу`,
	}

	bddsCmd.AddCommand(newGetCmd(getClient))
	bddsCmd.AddCommand(newAddCmd(getClient))

	root.AddCommand(bddsCmd)
}
